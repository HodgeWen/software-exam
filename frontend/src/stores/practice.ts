import { create } from "zustand";
import type { AnswerResult, Question } from "../api/bank";

// 练习会话客户端状态：章节练习与随机练习共用一套逐题作答流程，
// 题目序列由页面按路由拉取后经 start 装入，进度/所选/判分结果都只在客户端维护
interface PracticeState {
  questions: Question[];
  index: number;
  selected: string[]; // 当前题所选选项
  results: Record<number, AnswerResult>; // questionId -> 判分结果
  start: (questions: Question[]) => void;
  setSelected: (next: string[]) => void;
  recordResult: (questionId: number, result: AnswerResult) => void;
  next: () => void;
}

export const usePracticeStore = create<PracticeState>()((set) => ({
  questions: [],
  index: 0,
  selected: [],
  results: {},
  start: (questions) => set({ questions, index: 0, selected: [], results: {} }),
  setSelected: (next) => set({ selected: next }),
  recordResult: (questionId, result) =>
    set((s) => ({ results: { ...s.results, [questionId]: result } })),
  // 末题之后停在「已完成」位置（index === questions.length），不越界
  next: () => set((s) => ({ index: Math.min(s.index + 1, s.questions.length), selected: [] })),
}));
