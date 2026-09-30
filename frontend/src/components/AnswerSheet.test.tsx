import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { AnswerSheet } from "./AnswerSheet";
import type { Question } from "../api/bank";

const questions: Question[] = [
  { id: 11, no: 1, type: "single", stem: "题干1", options: [{ key: "A", text: "选项A" }] },
  { id: 12, no: 2, type: "single", stem: "题干2", options: [{ key: "A", text: "选项A" }] },
  { id: 13, no: 3, type: "single", stem: "题干3", options: [{ key: "A", text: "选项A" }] },
];

function renderSheet(answers: Record<number, string[]>, currentIndex = 0) {
  const onJump = vi.fn();
  render(
    <AnswerSheet
      questions={questions}
      answers={answers}
      currentIndex={currentIndex}
      onJump={onJump}
    />,
  );
  return onJump;
}

afterEach(cleanup);

describe("AnswerSheet 答题卡", () => {
  it("展示全部题号，已答/未答状态与已答统计", () => {
    renderSheet({ 11: ["A"], 13: ["A"] });

    expect(screen.getByRole("button", { name: "1" }).getAttribute("aria-pressed")).toBe("true");
    expect(screen.getByRole("button", { name: "2" }).getAttribute("aria-pressed")).toBe("false");
    expect(screen.getByRole("button", { name: "3" }).getAttribute("aria-pressed")).toBe("true");
    expect(screen.getByText("已答 2 / 3 题")).toBeTruthy();
  });

  it("当前题高亮标记", () => {
    renderSheet({}, 1);

    expect(screen.getByRole("button", { name: "2" }).getAttribute("aria-current")).toBe("true");
    expect(screen.getByRole("button", { name: "1" }).getAttribute("aria-current")).toBe("false");
  });

  it("点击题号回调对应下标实现跳题", () => {
    const onJump = renderSheet({}, 0);

    fireEvent.click(screen.getByRole("button", { name: "3" }));
    expect(onJump).toHaveBeenCalledWith(2);

    fireEvent.click(screen.getByRole("button", { name: "1" }));
    expect(onJump).toHaveBeenCalledWith(0);
  });
});
