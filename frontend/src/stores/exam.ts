import { create } from "zustand";

// 考试作答暂存：questionId -> 所选选项集合。考试期间只在客户端维护，
// 切题/答题卡跳题不丢失，交卷时一次性取出提交；开始新考试时整体重置
interface ExamState {
  answers: Record<number, string[]>;
  setAnswer: (questionId: number, selected: string[]) => void;
  resetExam: () => void;
}

export const useExamStore = create<ExamState>()((set) => ({
  answers: {},
  setAnswer: (questionId, selected) =>
    set((s) => ({ answers: { ...s.answers, [questionId]: selected } })),
  resetExam: () => set({ answers: {} }),
}));
