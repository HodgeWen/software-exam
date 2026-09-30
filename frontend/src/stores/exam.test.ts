import { beforeEach, describe, expect, it } from "vitest";
import { useExamStore } from "./exam";

beforeEach(() => {
  useExamStore.getState().resetExam();
});

describe("exam store 作答暂存", () => {
  it("setAnswer 按题目 id 记录所选，交卷前不因切题丢失", () => {
    useExamStore.getState().setAnswer(11, ["A"]);
    useExamStore.getState().setAnswer(12, ["B", "C"]);

    const answers = useExamStore.getState().answers;
    expect(answers[11]).toEqual(["A"]);
    expect(answers[12]).toEqual(["B", "C"]);
  });

  it("setAnswer 重答同一题时整体替换而不是叠加", () => {
    useExamStore.getState().setAnswer(11, ["A"]);
    useExamStore.getState().setAnswer(11, ["B"]);

    expect(useExamStore.getState().answers[11]).toEqual(["B"]);
  });

  it("resetExam 清空全部作答", () => {
    useExamStore.getState().setAnswer(11, ["A"]);

    useExamStore.getState().resetExam();

    expect(useExamStore.getState().answers).toEqual({});
  });
});
