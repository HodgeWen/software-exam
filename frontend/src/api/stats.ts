import { api } from "./client";

// 刷题统计的前端契约，字段与 backend/internal/service/stats.go 的 StatsSummary 一一对应；
// accuracy 为 0~1 的四位小数比值，展示格式化交给前端

export interface ChapterStats {
  chapter_id: number;
  chapter_name: string;
  total: number;
  correct: number;
  accuracy: number;
}

export interface SubjectStats {
  subject_id: number;
  subject_name: string;
  total: number;
  correct: number;
  accuracy: number;
  chapters: ChapterStats[];
}

export interface StatsSummary {
  total: number;
  correct: number;
  accuracy: number;
  subjects: SubjectStats[];
}

export function fetchStats() {
  return api.get<StatsSummary>("/stats");
}
