import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { FavoritesPage } from "./FavoritesPage";

const favoriteRow = {
  question: {
    id: 21,
    no: 3,
    type: "multiple",
    stem: "收藏题干",
    options: [
      { key: "A", text: "选项A" },
      { key: "B", text: "选项B" },
    ],
  },
  created_at: "2026-09-02T08:00:00Z",
};

let favoritesBody: { total: number; list: unknown[] };
let deleteURLs: string[] = [];

function stubFetch() {
  return vi.fn(async (url: unknown, init?: RequestInit) => {
    const path = String(url).split("?")[0];
    if (path === "/api/v1/favorites" && init?.method !== "DELETE") {
      return new Response(JSON.stringify(favoritesBody), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    }
    if (init?.method === "DELETE") {
      deleteURLs.push(path);
      favoritesBody = { total: 0, list: [] };
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
        <FavoritesPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

beforeEach(() => {
  favoritesBody = { total: 1, list: [favoriteRow] };
  deleteURLs = [];
  vi.stubGlobal("fetch", stubFetch());
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe("FavoritesPage 收藏列表", () => {
  it("展示收藏行（题型/题干）与分页信息", async () => {
    renderPage();

    expect(await screen.findByText("收藏题干")).toBeTruthy();
    expect(screen.getByText("多选")).toBeTruthy();
    expect(screen.getByText("第 1 / 1 页 · 共 1 条")).toBeTruthy();
  });

  it("取消收藏：DELETE 对应题目并刷新列表到空态", async () => {
    renderPage();
    fireEvent.click(await screen.findByRole("button", { name: "取消收藏" }));

    expect(await screen.findByText("暂无收藏，练习时点击「收藏本题」即可收藏")).toBeTruthy();
    expect(deleteURLs).toEqual(["/api/v1/favorites/21"]);
  });
});
