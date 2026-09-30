import { useEffect } from "react";
import { useMutation } from "@tanstack/react-query";
import { submitAnswer } from "../../api/bank";
import type { Question } from "../../api/bank";
import { QuestionCard } from "../../components/QuestionCard";
import { usePracticeStore } from "../../stores/practice";

interface PracticeRunnerProps {
  title: string;
  questions: Question[];
}

// 章节练习与随机练习共用的逐题作答流程：作答 → 提交判分 → 展示对错与解析 → 下一题。
// 会话状态（进度/所选/结果）放 zustand，判分经 react-query 走 src/api
export function PracticeRunner({ title, questions }: PracticeRunnerProps) {
  const sessionQuestions = usePracticeStore((s) => s.questions);
  const index = usePracticeStore((s) => s.index);
  const selected = usePracticeStore((s) => s.selected);
  const results = usePracticeStore((s) => s.results);
  const start = usePracticeStore((s) => s.start);
  const setSelected = usePracticeStore((s) => s.setSelected);
  const recordResult = usePracticeStore((s) => s.recordResult);
  const next = usePracticeStore((s) => s.next);

  useEffect(() => {
    start(questions);
  }, [questions, start]);

  const total = questions.length;
  const answered = Object.keys(results).length;
  const finished = index >= total;
  const question = questions[index];
  const result = question ? (results[question.id] ?? null) : null;

  const submitMutation = useMutation({
    mutationFn: (q: Question) => submitAnswer(q.id, selected),
    onSuccess: (r) => {
      if (question) recordResult(question.id, r);
    },
  });

  // 会话尚未与本批题目同步时先不渲染，避免闪现上一次练习的内容
  if (sessionQuestions !== questions) return null;

  if (finished) {
    const correctCount = Object.values(results).filter((r) => r.correct).length;
    return (
      <section className="rounded-lg border border-gray-200 bg-white p-6 text-center">
        <h2 className="text-lg font-semibold">练习完成</h2>
        <p className="mt-2 text-sm text-gray-600">
          共 {total} 题，答对 {correctCount} 题，正确率{" "}
          {total > 0 ? Math.round((correctCount / total) * 100) : 0}%
        </p>
        <button
          type="button"
          onClick={() => start(questions)}
          className="mt-4 rounded bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700"
        >
          重新练习
        </button>
      </section>
    );
  }

  return (
    <section className="flex flex-col gap-4">
      <p className="text-sm text-gray-600">
        {title} · 第 {index + 1} / {total} 题 · 已答 {answered} 题
      </p>
      <QuestionCard
        question={question}
        selected={selected}
        result={result}
        onChange={setSelected}
      />
      <div>
        {result ? (
          <button
            type="button"
            onClick={next}
            className="rounded bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700"
          >
            {index + 1 === total ? "完成练习" : "下一题"}
          </button>
        ) : (
          <button
            type="button"
            disabled={selected.length === 0 || submitMutation.isPending}
            onClick={() => submitMutation.mutate(question)}
            className="rounded bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
          >
            {submitMutation.isPending ? "判分中…" : "提交答案"}
          </button>
        )}
        {submitMutation.isError && (
          <p role="alert" className="mt-2 text-sm text-red-600">
            {(submitMutation.error as Error).message}
          </p>
        )}
      </div>
    </section>
  );
}
