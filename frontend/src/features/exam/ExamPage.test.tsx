import { cleanup, fireEvent, render, screen, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter, Route, Routes } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { ExamPage } from "./ExamPage";
import { useExamStore } from "../../stores/exam";

const questions = [
  {
    id: 11,
    no: 1,
    type: "single" as const,
    stem: "单选题干",
    options: [
      { key: "A", text: "单选项A" },
      { key: "B", text: "单选项B" },
    ],
  },
  {
    id: 12,
    no: 2,
    type: "multiple" as const,
    stem: "多选题干",
    options: [
      { key: "A", text: "多选项A" },
      { key: "B", text: "多选项B" },
      { key: "C", text: "多选项C" },
    ],
  },
];

const submitPayload = {
  exam: { id: 77, status: "submitted", correct_count: 1, total_count: 2, accuracy: 0.5 },
  results: [
    {
      question_id: 11,
      no: 1,
      selected: ["A"],
      correct: false,
      answer: ["B"],
      analysis: "单选解析：应选 B",
    },
    {
      question_id: 12,
      no: 2,
      selected: ["B", "C"],
      correct: true,
      answer: ["B", "C"],
      analysis: "多选解析：应选 BC",
    },
  ],
};

const submitBodies: unknown[] = [];
let fetchMock: ReturnType<typeof stubFetch>;

// 按开始考试/交卷两个端点分发响应，交卷请求体落档供断言
function stubFetch(durationMinutes: number) {
  return vi.fn(async (url: unknown, init?: RequestInit) => {
    if (url === "/api/v1/exams" && init?.method === "POST") {
      return new Response(
        JSON.stringify({
          exam: { id: 77, status: "in_progress", correct_count: 0, total_count: 0, accuracy: 0 },
          paper: {
            id: 3,
            subject_id: 1,
            code: "2025s",
            name: "2025 上半年真题",
            duration_minutes: durationMinutes,
          },
          questions,
        }),
        { status: 201, headers: { "Content-Type": "application/json" } },
      );
    }
    if (url === "/api/v1/exams/77/submit" && init?.method === "POST") {
      submitBodies.push(JSON.parse(String(init?.body)));
      return new Response(JSON.stringify(submitPayload), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    }
    throw new Error(`意外的请求：${String(url)}`);
  });
}

function renderPage() {
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter initialEntries={["/papers/3/exam"]}>
        <Routes>
          <Route path="/papers/:paperId/exam" element={<ExamPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

function clickOption(text: string) {
  fireEvent.click(screen.getByRole("button", { name: new RegExp(text) }));
}

beforeEach(() => {
  useExamStore.getState().resetExam();
  submitBodies.length = 0;
  fetchMock = stubFetch(30);
  vi.stubGlobal("fetch", fetchMock);
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  vi.useRealTimers();
});

describe("ExamPage 真题模拟考试", () => {
  it("拉整卷展示倒计时；答题卡跳题与作答暂存；取消确认不交卷，确认后展示得分与逐题解析", async () => {
    const confirmMock = vi.fn(() => true);
    vi.stubGlobal("confirm", confirmMock);
    renderPage();

    expect(await screen.findByRole("timer").then((el) => el.textContent)).toBe("剩余 30:00");
    expect(screen.getByText("2025 上半年真题")).toBeTruthy();
    expect(screen.getByText("第 1 / 2 题")).toBeTruthy();

    clickOption("单选项A");
    expect(screen.getByRole("button", { name: "1" }).getAttribute("aria-pressed")).toBe("true");
    expect(screen.getByText("已答 1 / 2 题")).toBeTruthy();

    // 答题卡跳到第 2 题，再跳回来：作答暂存不丢失
    fireEvent.click(screen.getByRole("button", { name: "2" }));
    expect(screen.getByText("第 2 / 2 题")).toBeTruthy();
    expect(screen.getByText("多选题干")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "1" }));
    expect(screen.getByRole("button", { name: /单选项A/ }).getAttribute("aria-pressed")).toBe(
      "true",
    );

    // 第一次确认取消：不发起交卷
    confirmMock.mockReturnValueOnce(false);
    fireEvent.click(screen.getByRole("button", { name: "交卷" }));
    expect(fetchMock).toHaveBeenCalledTimes(1);

    fireEvent.click(screen.getByRole("button", { name: "交卷" }));
    expect(await screen.findByText("考试完成")).toBeTruthy();
    expect(screen.getByText("共 2 题，答对 1 题，正确率 50%")).toBeTruthy();
    expect(screen.getByText("回答错误")).toBeTruthy();
    expect(screen.getByText("正确答案：B")).toBeTruthy();
    expect(screen.getByText("解析：单选解析：应选 B")).toBeTruthy();
    expect(screen.getByText("回答正确")).toBeTruthy();
    expect(screen.getByText("正确答案：B、C")).toBeTruthy();
    expect(screen.queryByText("交卷")).toBeNull();

    expect(confirmMock).toHaveBeenCalledTimes(2);
    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(submitBodies).toEqual([
      {
        answers: [
          { question_id: 11, selected: ["A"] },
          { question_id: 12, selected: [] },
        ],
      },
    ]);
  });

  it("倒计时归零时自动交卷，不弹确认且只交一次", async () => {
    vi.useFakeTimers();
    fetchMock = stubFetch(1);
    vi.stubGlobal("fetch", fetchMock);
    const confirmMock = vi.fn();
    vi.stubGlobal("confirm", confirmMock);
    renderPage();

    await act(async () => {
      await vi.advanceTimersByTimeAsync(0);
    });
    expect(screen.getByRole("timer").textContent).toBe("剩余 1:00");

    await act(async () => {
      await vi.advanceTimersByTimeAsync(61_000);
    });
    // 归零交卷的 mutation 结果经 react-query 的定时器批次通知，再推进一次让状态落地
    await act(async () => {
      await vi.advanceTimersByTimeAsync(5_000);
    });
    expect(screen.getByText("考试完成")).toBeTruthy();
    expect(confirmMock).not.toHaveBeenCalled();

    await act(async () => {
      await vi.advanceTimersByTimeAsync(120_000);
    });
    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(submitBodies).toEqual([
      {
        answers: [
          { question_id: 11, selected: [] },
          { question_id: 12, selected: [] },
        ],
      },
    ]);
  });
});
