import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { act, renderHook } from "@testing-library/react";
import { useExamCountdown } from "./use-exam-countdown";

beforeEach(() => {
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
});

describe("useExamCountdown 考试倒计时", () => {
  it("每秒递减剩余时间", () => {
    const deadline = Date.now() + 10 * 60_000;
    const { result } = renderHook(() => useExamCountdown(deadline));

    expect(result.current).toBe(600);

    act(() => {
      vi.advanceTimersByTime(90_000);
    });
    expect(result.current).toBe(510);
  });

  it("归零后停在下界不再继续递减", () => {
    const deadline = Date.now() + 30_000;
    const { result } = renderHook(() => useExamCountdown(deadline));

    act(() => {
      vi.advanceTimersByTime(120_000);
    });
    expect(result.current).toBe(0);

    act(() => {
      vi.advanceTimersByTime(60_000);
    });
    expect(result.current).toBe(0);
  });

  it("截止时间变更时按新截止时间重新计表", () => {
    const first = renderHook(({ deadline }) => useExamCountdown(deadline), {
      initialProps: { deadline: Date.now() + 60_000 },
    });

    act(() => {
      vi.advanceTimersByTime(30_000);
    });
    expect(first.result.current).toBe(30);

    act(() => {
      first.rerender({ deadline: Date.now() + 300_000 });
    });
    expect(first.result.current).toBe(300);
  });
});
