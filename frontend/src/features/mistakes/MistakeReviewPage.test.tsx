import { cleanup, render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { MistakeReviewPage } from "./MistakeReviewPage";
import { usePracticeStore } from "../../stores/practice";

const questions = [
  {
    id: 11,
    no: 1,
    type: "single",
    stem: "重刷题干",
    options: [
      { key: "A", text: "选项A" },
      { key: "B", text: "选项B" },
    ],
  },
];

let questionsBody: { questions: unknown[] };

function renderPage() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <MistakeReviewPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

beforeEach(() => {
  usePracticeStore.setState({ questions: [], index: 0, selected: [], results: {} });
  questionsBody = { questions };
  vi.stubGlobal(
    "fetch",
    vi.fn(
      async () =>
        new Response(JSON.stringify(questionsBody), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        }),
    ),
  );
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe("MistakeReviewPage 错题重刷", () => {
  it("以错题为题源进入练习流程，复用逐题作答交互", async () => {
    renderPage();

    expect(await screen.findByText("重刷题干")).toBeTruthy();
    expect(screen.getByText("错题重刷 · 第 1 / 1 题 · 已答 0 题")).toBeTruthy();
    expect(screen.getByRole("button", { name: "提交答案" })).toBeTruthy();
  });

  it("错题本为空时展示空态", async () => {
    questionsBody = { questions: [] };
    renderPage();

    expect(await screen.findByText("错题本为空，先去练习吧")).toBeTruthy();
  });
});
