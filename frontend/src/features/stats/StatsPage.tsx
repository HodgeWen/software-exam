import type { ReactNode } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router";
import { fetchStats } from "../../api/stats";
import type { SubjectStats } from "../../api/stats";

// 统计页：总答题量/总正确率概览 + 按科目、章节的作答分布
export function StatsPage() {
  const statsQuery = useQuery({ queryKey: ["stats"], queryFn: fetchStats });

  let body: ReactNode;
  if (statsQuery.isPending) {
    body = <p className="text-sm text-gray-600">加载中…</p>;
  } else if (statsQuery.isError) {
    body = (
      <p role="alert" className="text-sm text-red-600">
        加载统计失败：{(statsQuery.error as Error).message}
      </p>
    );
  } else {
    const { total, correct, accuracy, subjects } = statsQuery.data;
    body = (
      <>
        <div className="grid gap-3 sm:grid-cols-3">
          <div className="rounded-lg border border-gray-200 bg-white p-4 text-center">
            <p className="text-xs text-gray-500">总答题量</p>
            <p className="mt-1 text-2xl font-bold">{total}</p>
          </div>
          <div className="rounded-lg border border-gray-200 bg-white p-4 text-center">
            <p className="text-xs text-gray-500">答对题数</p>
            <p className="mt-1 text-2xl font-bold">{correct}</p>
          </div>
          <div className="rounded-lg border border-gray-200 bg-white p-4 text-center">
            <p className="text-xs text-gray-500">总正确率</p>
            <p className="mt-1 text-2xl font-bold">{Math.round(accuracy * 100)}%</p>
          </div>
        </div>

        <section className="mt-6">
          <h2 className="mb-2 font-semibold">科目分布</h2>
          {subjects.length === 0 ? (
            <p className="text-sm text-gray-600">暂无作答记录，先去刷题吧</p>
          ) : (
            subjects.map((s) => <SubjectSection key={s.subject_id} subject={s} />)
          )}
        </section>
      </>
    );
  }

  return (
    <main className="mx-auto min-h-screen max-w-3xl p-6">
      <nav className="text-sm">
        <Link to="/" className="text-indigo-600 hover:underline">
          返回科目选择
        </Link>
      </nav>
      <h1 className="mt-2 text-xl font-bold">刷题统计</h1>
      <div className="mt-4">{body}</div>
    </main>
  );
}

function SubjectSection({ subject }: { subject: SubjectStats }) {
  return (
    <div className="mb-4 rounded-lg border border-gray-200 bg-white p-4">
      <h3 className="font-medium">{subject.subject_name}</h3>
      <p className="mt-1 text-sm text-gray-600">
        共答 {subject.total} 题 · 答对 {subject.correct} 题 · 正确率{" "}
        {Math.round(subject.accuracy * 100)}%
      </p>
      {subject.chapters.length > 0 && (
        <ul className="mt-3 flex flex-col gap-1.5">
          {subject.chapters.map((c) => (
            <li key={c.chapter_id} className="flex items-center justify-between gap-4 text-sm">
              <span className="min-w-0 truncate">{c.chapter_name}</span>
              <span className="shrink-0 text-gray-600">
                {c.total} 题 · 答对 {c.correct} · 正确率 {Math.round(c.accuracy * 100)}%
              </span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
