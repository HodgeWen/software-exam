import { api } from "./client";
import type { Question } from "./bank";

// 错题本的前端契约，字段与 backend/internal/handler/mistake.go 的 JSON 视图一一对应；
// 题目视图不含正确答案与解析——重刷判分统一走 submitAnswer

export interface MistakeItem {
  question: Question;
  wrong_count: number;
  last_wrong_at: string;
}

export function fetchMistakes(page: number, pageSize: number, subjectId?: number) {
  const params = new URLSearchParams({ page: String(page), page_size: String(pageSize) });
  if (subjectId) params.set("subject_id", String(subjectId));
  return api.get<{ total: number; list: MistakeItem[] }>(`/mistakes?${params}`);
}

// 错题重刷题源：当前用户全部错题的题目序列
export function fetchMistakeQuestions() {
  return api.get<{ questions: Question[] }>("/mistakes/questions");
}

export function removeMistake(questionId: number) {
  return api.delete<void>(`/mistakes/${questionId}`);
}
