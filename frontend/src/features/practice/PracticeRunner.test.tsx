import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { PracticeRunner } from "./PracticeRunner";
import type { Question } from "../../api/bank";
import { usePracticeStore } from "../../stores/practice";

const questions: Question[] = [
  {
    id: 11,
    no: 1,
    type: "single",
    stem: "单选题干",
    options: [
      { key: "A", text: "单选项A" },
      { key: "B", text: "单选项B" },
    ],
  },
  {
    id: 12,
    no: 2,
    type: "multiple",
    stem: "多选题干",
    options: [
      { key: "A", text: "多选项A" },
      { key: "B", text: "多选项B" },
      { key: "C", text: "多选项C" },
      { key: "D", text: "多选项D" },
    ],
  },
];

// 按题目 id 返回判分结果：单选答错、多选答对，模拟后端逐题判分
function stubSubmitFetch() {
  return vi.fn(async (_url: unknown, init?: RequestInit) => {
    const body = JSON.parse(String(init?.body)) as { question_id: number };
    const payload =
      body.question_id === 11
        ? { correct: false, answer: ["B"], analysis: "单选解析：应选 B" }
        : { correct: true, answer: ["C", "D"], analysis: "多选解析：应选 CD" };
    return new Response(JSON.stringify(payload), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    });
  });
}

let fetchMock: ReturnType<typeof stubSubmitFetch>;

function renderRunner() {
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <PracticeRunner title="章节练习" questions={questions} />
    </QueryClientProvider>,
  );
}

function clickOption(text: string) {
  fireEvent.click(screen.getByRole("button", { name: new RegExp(text) }));
}

beforeEach(() => {
  usePracticeStore.setState({ questions: [], index: 0, selected: [], results: {} });
  fetchMock = stubSubmitFetch();
  vi.stubGlobal("fetch", fetchMock);
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe("PracticeRunner 练习流程", () => {
  it("未作答时不能提交；单选提交后立即展示判分结果与解析，选项锁定", async () => {
    renderRunner();

    expect(screen.getByText("章节练习 · 第 1 / 2 题 · 已答 0 题")).toBeTruthy();
    const submit = screen.getByRole("button", { name: "提交答案" }) as HTMLButtonElement;
    expect(submit.disabled).toBe(true);

    clickOption("单选项A");
    expect((screen.getByRole("button", { name: "提交答案" }) as HTMLButtonElement).disabled).toBe(
      false,
    );
    fireEvent.click(screen.getByRole("button", { name: "提交答案" }));

    expect(await screen.findByText("回答错误")).toBeTruthy();
    expect(screen.getByText("正确答案：B")).toBeTruthy();
    expect(screen.getByText("解析：单选解析：应选 B")).toBeTruthy();
    expect(screen.getByText("章节练习 · 第 1 / 2 题 · 已答 1 题")).toBeTruthy();

    // 已判分：选项与提交入口锁定，不再发起新的判分请求
    clickOption("单选项B");
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("多选可勾选多项；判分展示为答对；进度推进到完成页并统计正确率", async () => {
    renderRunner();

    clickOption("单选项A");
    fireEvent.click(screen.getByRole("button", { name: "提交答案" }));
    await screen.findByText("回答错误");

    fireEvent.click(screen.getByRole("button", { name: "下一题" }));
    expect(screen.getByText("章节练习 · 第 2 / 2 题 · 已答 1 题")).toBeTruthy();
    expect(screen.getByText("多选题干")).toBeTruthy();

    clickOption("多选项C");
    clickOption("多选项D");
    expect(screen.getByRole("button", { name: /多选项C/ }).getAttribute("aria-pressed")).toBe(
      "true",
    );
    expect(screen.getByRole("button", { name: /多选项D/ }).getAttribute("aria-pressed")).toBe(
      "true",
    );

    fireEvent.click(screen.getByRole("button", { name: "提交答案" }));
    expect(await screen.findByText("回答正确")).toBeTruthy();
    expect(screen.getByText("正确答案：C、D")).toBeTruthy();

    fireEvent.click(screen.getByRole("button", { name: "完成练习" }));
    expect(screen.getByText("练习完成")).toBeTruthy();
    expect(screen.getByText("共 2 题，答对 1 题，正确率 50%")).toBeTruthy();
  });

  it("完成后可重新练习：进度归零回到第一题", async () => {
    renderRunner();

    clickOption("单选项A");
    fireEvent.click(screen.getByRole("button", { name: "提交答案" }));
    await screen.findByText("回答错误");
    fireEvent.click(screen.getByRole("button", { name: "下一题" }));
    clickOption("多选项A");
    fireEvent.click(screen.getByRole("button", { name: "提交答案" }));
    await screen.findByText("回答正确");
    fireEvent.click(screen.getByRole("button", { name: "完成练习" }));

    fireEvent.click(screen.getByRole("button", { name: "重新练习" }));
    expect(screen.getByText("章节练习 · 第 1 / 2 题 · 已答 0 题")).toBeTruthy();
    expect(screen.queryByText("回答正确")).toBeNull();
    expect(screen.queryByText("练习完成")).toBeNull();
  });

  it("收藏本题：POST /favorites，成功后按钮变为已收藏并禁用", async () => {
    const calls: Array<[string, RequestInit | undefined]> = [];
    vi.stubGlobal(
      "fetch",
      vi.fn(async (url: string, init?: RequestInit) => {
        calls.push([url, init]);
        if (String(url).endsWith("/api/v1/favorites")) return new Response(null, { status: 204 });
        return new Response(JSON.stringify({ correct: false, answer: ["B"], analysis: "解析" }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      }),
    );
    renderRunner();

    fireEvent.click(screen.getByRole("button", { name: "收藏本题" }));
    expect(await screen.findByText("已收藏")).toBeTruthy();
    expect((screen.getByRole("button", { name: "已收藏" }) as HTMLButtonElement).disabled).toBe(
      true,
    );

    const favorite = calls.find(([u]) => String(u).endsWith("/api/v1/favorites"));
    expect(favorite).toBeTruthy();
    expect(favorite?.[1]?.method).toBe("POST");
    expect(favorite?.[1]?.body).toBe(JSON.stringify({ question_id: 11 }));
  });
});
