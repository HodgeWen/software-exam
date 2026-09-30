import { api } from "./client";
import type { Paper, Question } from "./bank";

// 真题模拟考试的前端契约，字段与 backend/internal/handler/exam.go 的 JSON 视图一一对应；
// 考试记录只取前端用到的得分快照字段

export interface Exam {
  id: number;
  status: "in_progress" | "submitted";
  correct_count: number;
  total_count: number;
  accuracy: number;
}

export interface ExamStart {
  exam: Exam;
  paper: Paper;
  questions: Question[];
}

export interface ExamQuestionResult {
  question_id: number;
  no: number;
  selected: string[];
  correct: boolean;
  answer: string[];
  analysis: string;
}

export interface ExamSubmitResult {
  exam: Exam;
  results: ExamQuestionResult[];
}

export interface ExamAnswerInput {
  question_id: number;
  selected: string[];
}

// 开始考试：生成进行中的考试记录，返回整卷题目（不含答案与解析）与试卷时长
export function startExam(paperId: number) {
  return api.post<ExamStart>("/exams", { paper_id: paperId });
}

// 交卷：一次性提交全部作答，服务端统一评分并返回逐题对错与正确答案/解析
export function submitExam(examId: number, answers: ExamAnswerInput[]) {
  return api.post<ExamSubmitResult>(`/exams/${examId}/submit`, { answers });
}
