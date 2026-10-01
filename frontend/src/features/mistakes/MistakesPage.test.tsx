import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { MistakesPage } from "./MistakesPage";

const subjectsPayload = {
  subjects: [{ id: 1, code: "RJ-SJ", name: "软件设计师" }],
};

const mistakeRow = {
  question: {
    id: 11,
    no: 1,
    type: "single",
    stem: "错题题干",
    options: [
      { key: "A", text: "选项A" },
      { key: "B", text: "选项B" },
    ],
  },
  subject_name: "软件设计师",
  chapter_name: "计算机系统基础知识",
  wrong_count: 2,
  last_wrong_at: "2026-09-01T10:00:00Z",
};

let mistakesBody: { total: number; list: unknown[] };
let mistakeQueries: string[] = [];
let deleteURLs: string[] = [];

function jsonResponse(payload: unknown) {
  return new Response(JSON.stringify(payload), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
}

// 按请求路径分发：科目列表与错题列表可重复拉取，DELETE 移除后列表清空
function stubFetch() {
  return vi.fn(async (url: unknown, init?: RequestInit) => {
    const [path, query] = String(url).split("?");
    if (path === "/api/v1/subjects") return jsonResponse(subjectsPayload);
    if (path === "/api/v1/mistakes") {
      mistakeQueries.push(query ?? "");
      return jsonResponse(mistakesBody);
    }
    if (init?.method === "DELETE") {
      deleteURLs.push(path);
      mistakesBody = { total: 0, list: [] };
      return new Response(null, { status: 204 });
    }
    throw new Error(`未预期的请求: ${String(url)}`);
  });
}

function renderPage() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <MistakesPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

beforeEach(() => {
  mistakesBody = { total: 1, list: [mistakeRow] };
  mistakeQueries = [];
  deleteURLs = [];
  vi.stubGlobal("fetch", stubFetch());
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe("MistakesPage 错题本列表", () => {
  it("展示错题行（题型/科目/知识点/错误次数/题干）与分页信息", async () => {
    renderPage();

    expect(await screen.findByText("错题题干")).toBeTruthy();
    expect(screen.getByText("单选")).toBeTruthy();
    expect(screen.getByText("计算机系统基础知识")).toBeTruthy();
    expect(screen.getByText("错 2 次")).toBeTruthy();
    expect(screen.getByText("第 1 / 1 页 · 共 1 条")).toBeTruthy();
    expect((screen.getByRole("button", { name: "上一页" }) as HTMLButtonElement).disabled).toBe(
      true,
    );
  });

  it("按科目过滤：点击标签后请求携带 subject_id 并回到第一页", async () => {
    renderPage();
    await screen.findByText("错题题干");

    fireEvent.click(screen.getByRole("button", { name: "软件设计师" }));

    await waitFor(() => expect(mistakeQueries.at(-1)).toContain("subject_id=1"));
    expect(mistakeQueries.at(-1)).toContain("page=1");
  });

  it("手动移除：DELETE 对应题目并刷新列表到空态", async () => {
    renderPage();
    fireEvent.click(await screen.findByRole("button", { name: "移除" }));

    expect(await screen.findByText("暂无错题，练习中答错的题会自动进入错题本")).toBeTruthy();
    expect(deleteURLs).toEqual(["/api/v1/mistakes/11"]);
  });
});
