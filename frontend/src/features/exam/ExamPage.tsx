import { useEffect, useMemo, useRef, useState } from "react";
import type { ReactNode } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router";
import { startExam, submitExam } from "../../api/exam";
import type { ExamStart, ExamSubmitResult } from "../../api/exam";
import { AnswerSheet } from "../../components/AnswerSheet";
import { QuestionCard } from "../../components/QuestionCard";
import { useExamStore } from "../../stores/exam";
import { useExamCountdown } from "./use-exam-countdown";

// 真题模拟考试页：开始考试拉整卷（无答案），倒计时作答 + 答题卡跳题，
// 手动/归零自动交卷后展示得分与逐题解析
export function ExamPage() {
  const { paperId } = useParams();
  const pid = Number(paperId);
  const valid = Number.isInteger(pid) && pid > 0;

  // 开始考试是创建考试记录的 POST：禁用重试与聚焦重取，避免重复生成记录
  const startQuery = useQuery({
    queryKey: ["exam", pid],
    queryFn: () => startExam(pid),
    enabled: valid,
    retry: false,
    refetchOnWindowFocus: false,
    staleTime: Infinity,
  });

  let body: ReactNode;
  if (!valid || startQuery.isPending) {
    body = <p className="text-sm text-gray-600">加载中…</p>;
  } else if (startQuery.isError) {
    body = (
      <p role="alert" className="text-sm text-red-600">
        开始考试失败：{(startQuery.error as Error).message}
      </p>
    );
  } else if (startQuery.data.questions.length === 0) {
    body = <p className="text-sm text-gray-600">试卷暂无题目</p>;
  } else {
    body = <ExamRunner start={startQuery.data} />;
  }

  return (
    <main className="mx-auto min-h-screen max-w-5xl p-6">
      <nav className="text-sm">
        <Link to="/" className="text-indigo-600 hover:underline">
          返回科目选择
        </Link>
      </nav>
      {body}
    </main>
  );
}

function ExamRunner({ start }: { start: ExamStart }) {
  const { exam, paper, questions } = start;
  const answers = useExamStore((s) => s.answers);
  const setAnswer = useExamStore((s) => s.setAnswer);
  const resetExam = useExamStore((s) => s.resetExam);
  const [index, setIndex] = useState(0);

  // 截止时间在拿到整卷时一次性定死，之后只随表走
  const deadline = useMemo(
    () => Date.now() + paper.duration_minutes * 60_000,
    [exam.id, paper.duration_minutes],
  );
  const remaining = useExamCountdown(deadline);

  const submitMutation = useMutation({
    mutationFn: () =>
      submitExam(
        exam.id,
        questions.map((q) => ({ question_id: q.id, selected: answers[q.id] ?? [] })),
      ),
    onSuccess: () => {
      // 只清作答暂存；整卷缓存保留——误退出后回来还能续答同一份卷（重复交卷由后端 409 拦截）
      resetExam();
    },
  });

  // 倒计时归零自动交卷；ref 只放行一次，手动交卷不经过这里
  const autoSubmittedRef = useRef(false);
  useEffect(() => {
    if (remaining > 0 || autoSubmittedRef.current || submitMutation.isPending) return;
    if (submitMutation.isSuccess) return;
    autoSubmittedRef.current = true;
    submitMutation.mutate();
  }, [remaining, submitMutation]);

  if (submitMutation.isSuccess) {
    return <ExamResult questions={questions} result={submitMutation.data} />;
  }

  const question = questions[index];
  const answeredCount = questions.filter((q) => (answers[q.id] ?? []).length > 0).length;
  const minutes = Math.floor(remaining / 60);
  const seconds = String(remaining % 60).padStart(2, "0");

  function confirmSubmit() {
    if (window.confirm(`已答 ${answeredCount} / ${questions.length} 题，确定交卷？`)) {
      submitMutation.mutate();
    }
  }

  return (
    <section className="mt-4 flex flex-col gap-4 lg:flex-row lg:items-start">
      <div className="flex flex-1 flex-col gap-4">
        <header className="flex flex-wrap items-center justify-between gap-2">
          <h1 className="text-xl font-bold">{paper.name}</h1>
          <p
            role="timer"
            className={`font-mono text-lg ${remaining <= 300 ? "text-red-600" : "text-gray-700"}`}
          >
            剩余 {minutes}:{seconds}
          </p>
        </header>
        <p className="text-sm text-gray-600">
          第 {index + 1} / {questions.length} 题
        </p>
        <QuestionCard
          question={question}
          selected={answers[question.id] ?? []}
          result={null}
          onChange={(next) => setAnswer(question.id, next)}
        />
        <div className="flex items-center justify-between">
          <button
            type="button"
            disabled={index === 0}
            onClick={() => setIndex(index - 1)}
            className="rounded border border-gray-300 px-4 py-2 text-sm hover:border-indigo-400 disabled:opacity-50"
          >
            上一题
          </button>
          <button
            type="button"
            disabled={index === questions.length - 1}
            onClick={() => setIndex(index + 1)}
            className="rounded border border-gray-300 px-4 py-2 text-sm hover:border-indigo-400 disabled:opacity-50"
          >
            下一题
          </button>
        </div>
        <div>
          <button
            type="button"
            disabled={submitMutation.isPending}
            onClick={confirmSubmit}
            className="rounded bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
          >
            {submitMutation.isPending ? "交卷中…" : "交卷"}
          </button>
          {submitMutation.isError && (
            <p role="alert" className="mt-2 text-sm text-red-600">
              交卷失败：{(submitMutation.error as Error).message}
            </p>
          )}
        </div>
      </div>
      <AnswerSheet questions={questions} answers={answers} currentIndex={index} onJump={setIndex} />
    </section>
  );
}

function ExamResult({
  questions,
  result,
}: {
  questions: ExamStart["questions"];
  result: ExamSubmitResult;
}) {
  const { exam, results } = result;
  const questionById = new Map(questions.map((q) => [q.id, q]));

  return (
    <section className="mt-4 flex flex-col gap-4">
      <div className="rounded-lg border border-gray-200 bg-white p-6 text-center">
        <h2 className="text-lg font-semibold">考试完成</h2>
        <p className="mt-2 text-sm text-gray-600">
          共 {exam.total_count} 题，答对 {exam.correct_count} 题，正确率{" "}
          {Math.round(exam.accuracy * 100)}%
        </p>
      </div>
      {results.map((r) => {
        const question = questionById.get(r.question_id);
        if (!question) return null;
        return (
          <QuestionCard
            key={r.question_id}
            question={question}
            selected={r.selected}
            result={{ correct: r.correct, answer: r.answer, analysis: r.analysis }}
            onChange={() => {}}
          />
        );
      })}
    </section>
  );
}
