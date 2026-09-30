import { cleanup, render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { StatsPage } from "./StatsPage";

const statsPayload = {
  total: 12,
  correct: 9,
  accuracy: 0.75,
  subjects: [
    {
      subject_id: 1,
      subject_name: "软件设计师",
      total: 12,
      correct: 9,
      accuracy: 0.75,
      chapters: [
        {
          chapter_id: 2,
          chapter_name: "计算机组成与体系结构",
          total: 4,
          correct: 3,
          accuracy: 0.75,
        },
        { chapter_id: 3, chapter_name: "程序设计语言", total: 8, correct: 6, accuracy: 0.75 },
      ],
    },
  ],
};

beforeEach(() => {
  vi.stubGlobal(
    "fetch",
    vi.fn(
      async () =>
        new Response(JSON.stringify(statsPayload), {
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

function renderPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <StatsPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("StatsPage 刷题统计", () => {
  it("展示总量、答对数与总正确率概览", async () => {
    renderPage();

    expect(await screen.findByText("总答题量")).toBeTruthy();
    expect(screen.getByText("12")).toBeTruthy();
    expect(screen.getByText("9")).toBeTruthy();
    expect(screen.getByText("75%")).toBeTruthy();
  });

  it("展示按科目与按章节的分布", async () => {
    renderPage();

    expect(await screen.findByText("软件设计师")).toBeTruthy();
    expect(screen.getByText("共答 12 题 · 答对 9 题 · 正确率 75%")).toBeTruthy();
    expect(screen.getByText("4 题 · 答对 3 · 正确率 75%")).toBeTruthy();
    expect(screen.getByText("8 题 · 答对 6 · 正确率 75%")).toBeTruthy();
  });
});
