import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router";
import { fetchMistakeQuestions } from "../../api/mistakes";
import { PracticeRunner } from "../practice/PracticeRunner";

// 错题重刷页：以当前用户错题为题源，复用练习流程逐题作答即时判分；
// 不缓存旧批次——列表可能已因移除/新增错题变化，每次进入重新取题
export function MistakeReviewPage() {
  const questionsQuery = useQuery({
    queryKey: ["mistake-questions"],
    queryFn: fetchMistakeQuestions,
    gcTime: 0,
  });

  return (
    <main className="min-h-screen bg-gray-50">
      <div className="mx-auto max-w-3xl px-4 py-8">
        <nav className="text-sm">
          <Link to="/mistakes" className="text-indigo-600 hover:underline">
            ← 返回错题本
          </Link>
        </nav>
        <h1 className="mt-3 text-2xl font-bold text-gray-900">错题重刷</h1>
        <p className="mt-1 text-sm text-gray-500">以当前错题为题源，逐题作答即时判分</p>
        <div className="mt-5">
          {questionsQuery.isPending ? (
            <p className="text-sm text-gray-600">加载中…</p>
          ) : questionsQuery.isError ? (
            <p
              role="alert"
              className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-600"
            >
              加载错题失败：{(questionsQuery.error as Error).message}
            </p>
          ) : questionsQuery.data.questions.length === 0 ? (
            <p className="rounded-xl border border-dashed border-gray-300 bg-white px-6 py-12 text-center text-sm text-gray-500">
              错题本为空，先去练习吧
            </p>
          ) : (
            <PracticeRunner title="错题重刷" questions={questionsQuery.data.questions} />
          )}
        </div>
      </div>
    </main>
  );
}
