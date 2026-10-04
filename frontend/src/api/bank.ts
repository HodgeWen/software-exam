import { api } from "./client";

// 题库取题与单题判分的前端契约，字段与 backend/internal/handler 的 JSON 视图一一对应；
// 题目视图不含正确答案与解析——判分结果统一由 submitAnswer 返回

export interface Subject {
  id: number;
  code: string;
  name: string;
}

export interface Chapter {
  id: number;
  subject_id: number;
  code: string;
  name: string;
  sort: number;
}

export interface Paper {
  id: number;
  subject_id: number;
  code: string;
  name: string;
  duration_minutes: number;
}

export type QuestionType = "single" | "multiple";

export interface QuestionOption {
  key: string;
  text: string;
}

export interface Question {
  id: number;
  no: number;
  type: QuestionType;
  stem: string;
  options: QuestionOption[];
}

export interface AnswerResult {
  correct: boolean;
  answer: string[];
  analysis: string;
}

export function fetchSubjects() {
  return api.get<{ subjects: Subject[] }>("/subjects");
}

export function fetchSubjectDetail(subjectId: number) {
  return api.get<{ subject: Subject; chapters: Chapter[]; papers: Paper[] }>(
    `/subjects/${subjectId}`,
  );
}

// 顺序取题：返回指定科目+章节的全部题目（按题目稳定顺序）
export function fetchChapterQuestions(subjectId: number, chapterId: number) {
  return api.get<{ questions: Question[] }>(
    `/subjects/${subjectId}/chapters/${chapterId}/questions`,
  );
}

// 随机取题：按科目返回打乱后的题目序列
export function fetchRandomQuestions(subjectId: number) {
  return api.get<{ questions: Question[] }>(`/subjects/${subjectId}/questions/random`);
}

// 单题提交：服务端判分并即时返回对/错、正确答案与解析
export function submitAnswer(questionId: number, selected: string[]) {
  return api.post<AnswerResult>("/answers", { question_id: questionId, selected });
}

// 背题模式题目视图：在 Question 基础上附带正确答案与解析（reveal=1 接口专用）
export interface RevealedQuestion extends Question {
  answer: string[];
  analysis: string;
}

// 背题取题：按章节顺序返回带正确答案与解析的题目
export function fetchChapterQuestionsRevealed(subjectId: number, chapterId: number) {
  return api.get<{ questions: RevealedQuestion[] }>(
    `/subjects/${subjectId}/chapters/${chapterId}/questions?reveal=1`,
  );
}

// 背题取题：整卷按卷内题号返回带正确答案与解析的题目
export function fetchPaperQuestionsRevealed(paperId: number) {
  return api.get<{ paper: Paper; questions: RevealedQuestion[] }>(`/papers/${paperId}?reveal=1`);
}
