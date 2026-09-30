import { beforeEach, describe, expect, it } from "vitest";
import { usePracticeStore } from "./practice";
import type { Question } from "../api/bank";

const q = (id: number): Question => ({
  id,
  no: id,
  type: "single",
  stem: `题干${id}`,
  options: [{ key: "A", text: "选项A" }],
});

beforeEach(() => {
  usePracticeStore.getState().start([q(1), q(2)]);
});

describe("practice store 会话状态", () => {
  it("start 装入新题目序列并重置进度与结果", () => {
    usePracticeStore.getState().setSelected(["A"]);
    usePracticeStore.getState().recordResult(1, { correct: true, answer: ["A"], analysis: "x" });

    usePracticeStore.getState().start([q(9)]);

    const s = usePracticeStore.getState();
    expect(s.questions.map((x) => x.id)).toEqual([9]);
    expect(s.index).toBe(0);
    expect(s.selected).toEqual([]);
    expect(s.results).toEqual({});
  });

  it("recordResult 按题目 id 记录判分结果", () => {
    usePracticeStore.getState().recordResult(2, { correct: false, answer: ["A"], analysis: "y" });

    expect(usePracticeStore.getState().results[2]).toEqual({
      correct: false,
      answer: ["A"],
      analysis: "y",
    });
    expect(usePracticeStore.getState().results[1]).toBeUndefined();
  });

  it("next 推进并清空所选，末题后停在完成位不越界", () => {
    usePracticeStore.getState().setSelected(["A"]);

    usePracticeStore.getState().next();
    expect(usePracticeStore.getState().index).toBe(1);
    expect(usePracticeStore.getState().selected).toEqual([]);

    usePracticeStore.getState().next();
    expect(usePracticeStore.getState().index).toBe(2);

    usePracticeStore.getState().next();
    expect(usePracticeStore.getState().index).toBe(2);
  });
});
