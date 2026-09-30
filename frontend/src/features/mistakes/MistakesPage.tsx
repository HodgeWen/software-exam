import { useState } from "react";
import type { ReactNode } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "react-router";
import { fetchSubjects } from "../../api/bank";
import { fetchMistakes, removeMistake } from "../../api/mistakes";
import { Pagination } from "../../components/Pagination";

const PAGE_SIZE = 10;

// 错题本页：按科目过滤 + 分页的错题列表，可手动移除；重刷是独立路由
export function MistakesPage() {
  const [subjectId, setSubjectId] = useState(0);
  const [page, setPage] = useState(1);
  const queryClient = useQueryClient();

  const subjectsQuery = useQuery({ queryKey: ["subjects"], queryFn: fetchSubjects });
  const mistakesQuery = useQuery({
    queryKey: ["mistakes", subjectId, page],
    queryFn: () => fetchMistakes(page, PAGE_SIZE, subjectId || undefined),
  });

  const removeMutation = useMutation({
    mutationFn: removeMistake,
    onSuccess: () => {
      // 移除的是本页最后一条时先回退一页，避免翻页请求落在空页上
      if ((mistakesQuery.data?.list.length ?? 0) <= 1 && page > 1) {
        setPage((p) => p - 1);
      }
      void queryClient.invalidateQueries({ queryKey: ["mistakes"] });
    },
  });

  let body: ReactNode;
  if (mistakesQuery.isPending) {
    body = <p className="text-sm text-gray-600">加载中…</p>;
  } else if (mistakesQuery.isError) {
    body = (
      <p role="alert" className="text-sm text-red-600">
        加载错题失败：{(mistakesQuery.error as Error).message}
      </p>
    );
  } else if (mistakesQuery.data.list.length === 0) {
    body = <p className="text-sm text-gray-600">暂无错题，练习中答错的题会自动进入错题本</p>;
  } else {
    body = (
      <>
        <ul className="flex flex-col gap-3">
          {mistakesQuery.data.list.map((item) => (
            <li
              key={item.question.id}
              className="flex items-start justify-between gap-4 rounded border border-gray-200 bg-white p-4"
            >
              <div className="min-w-0">
                <p className="flex items-center gap-2 text-xs text-gray-500">
                  <span>{item.question.type === "multiple" ? "多选" : "单选"}</span>
                  <span>错 {item.wrong_count} 次</span>
                  <span>最近答错 {new Date(item.last_wrong_at).toLocaleString()}</span>
                </p>
                <p className="mt-1 line-clamp-2 text-sm font-medium">{item.question.stem}</p>
              </div>
              <button
                type="button"
                disabled={removeMutation.isPending && removeMutation.variables === item.question.id}
                onClick={() => removeMutation.mutate(item.question.id)}
                className="shrink-0 rounded border border-red-300 px-3 py-1.5 text-sm text-red-600 hover:bg-red-50 disabled:opacity-50"
              >
                移除
              </button>
            </li>
          ))}
        </ul>
        <Pagination
          page={page}
          pageSize={PAGE_SIZE}
          total={mistakesQuery.data.total}
          onChange={setPage}
        />
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
      <header className="mt-2 flex flex-wrap items-center justify-between gap-2">
        <h1 className="text-xl font-bold">错题本</h1>
        <div className="flex items-center gap-3 text-sm">
          <select
            value={subjectId}
            aria-label="按科目过滤"
            onChange={(e) => {
              setSubjectId(Number(e.target.value));
              setPage(1);
            }}
            className="rounded border border-gray-300 px-2 py-1.5"
          >
            <option value={0}>全部科目</option>
            {(subjectsQuery.data?.subjects ?? []).map((s) => (
              <option key={s.id} value={s.id}>
                {s.name}
              </option>
            ))}
          </select>
          <Link
            to="/mistakes/review"
            className="rounded bg-indigo-600 px-3 py-1.5 font-medium text-white hover:bg-indigo-700"
          >
            错题重刷
          </Link>
        </div>
      </header>
      <div className="mt-4">{body}</div>
    </main>
  );
}
