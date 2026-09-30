import { api } from "./client";
import type { Question } from "./bank";

// 收藏的前端契约，字段与 backend/internal/handler/favorite.go 的 JSON 视图一一对应

export interface FavoriteItem {
  question: Question;
  created_at: string;
}

export function fetchFavorites(page: number, pageSize: number) {
  const params = new URLSearchParams({ page: String(page), page_size: String(pageSize) });
  return api.get<{ total: number; list: FavoriteItem[] }>(`/favorites?${params}`);
}

export function addFavorite(questionId: number) {
  return api.post<void>("/favorites", { question_id: questionId });
}

export function removeFavorite(questionId: number) {
  return api.delete<void>(`/favorites/${questionId}`);
}
