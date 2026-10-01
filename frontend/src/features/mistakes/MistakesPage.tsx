import { useState } from "react";
import type { ReactNode } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "react-router";
import { fetchSubjects } from "../../api/bank";
import { fetchMistakes, removeMistake } from "../../api/mistakes";
import type { MistakeItem } from "../../api/mistakes";
import { Pagination } from "../../components/Pagination";

const PAGE_SIZE = 10;

// 答错次数分级：越高越红，用于卡片左缘与次数徽标着色
function severityOf(count: number) {
  if (count >= 3) return "border-l-red-400";
  if (count === 2) return "border-l-amber-400";
  return "border-l-slate-300";
}

function countBadgeOf(count: number) {
  if (count >= 3) return "bg-red-50 text-red-600";
  if (count === 2) return "bg-amber-50 text-amber-600";
  return "bg-slate-100 text-slate-500";
}

function formatWhen(iso: string) {
  const d = new Date(iso);
  const hh = String(d.getHours()).padStart(2, "0");
  const mm = String(d.getMinutes()).padStart(2, "0");
  return `${d.getMonth() + 1}月${d.getDate()}日 ${hh}:${mm}`;
}

// 错题卡片：左侧按严重度着色，标注题型/科目/知识点归属，可移除
function MistakeCard({
  item,
  removing,
  onRemove,
}: {
  item: MistakeItem;
  removing: boolean;
  onRemove: () => void;
}) {
  const multiple = item.question.type === "multiple";
  return (
    <li
      className={`rounded-xl border border-l-4 border-gray-200 bg-white p-4 shadow-sm transition hover:shadow-md ${severityOf(
        item.wrong_count,
      )}`}
    >
      <div className="flex items-center gap-2 text-xs">
        <span
          className={`rounded px-1.5 py-0.5 font-medium ${
            multiple ? "bg-violet-50 text-violet-600" : "bg-indigo-50 text-indigo-600"
          }`}
        >
          {multiple ? "多选" : "单选"}
        </span>
        {item.subject_name && (
          <span className="rounded bg-sky-50 px-1.5 py-0.5 text-sky-600">{item.subject_name}</span>
        )}
        <span className="rounded bg-emerald-50 px-1.5 py-0.5 text-emerald-600">
          {item.chapter_name || "真题卷"}
        </span>
        <span className="ml-auto flex items-center gap-2">
          <span
            className={`rounded-full px-2 py-0.5 font-medium ${countBadgeOf(item.wrong_count)}`}
          >
            错 {item.wrong_count} 次
          </span>
          <button
            type="button"
            disabled={removing}
            onClick={onRemove}
            className="rounded px-2 py-0.5 text-gray-400 transition hover:bg-red-50 hover:text-red-500 disabled:opacity-50"
          >
            移除
          </button>
        </span>
      </div>
      <p className="mt-2 line-clamp-2 text-sm font-medium text-gray-900">{item.question.stem}</p>
      <p className="mt-2 text-xs text-gray-400">最近答错 {formatWhen(item.last_wrong_at)}</p>
    </li>
  );
}

// 错题本页：统计头 + 科目筛选标签 + 按严重度着色的错题列表，可手动移除；重刷是独立路由
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
    body = (
      <ul className="flex flex-col gap-3" aria-hidden>
        {[0, 1, 2].map((i) => (
          <li key={i} className="h-24 animate-pulse rounded-xl border border-gray-200 bg-white" />
        ))}
      </ul>
    );
  } else if (mistakesQuery.isError) {
    body = (
      <p
        role="alert"
        className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-600"
      >
        加载错题失败：{(mistakesQuery.error as Error).message}
      </p>
    );
  } else if (mistakesQuery.data.list.length === 0) {
    body = (
      <div className="flex flex-col items-center gap-3 rounded-xl border border-dashed border-gray-300 bg-white px-6 py-14 text-center">
        <span className="text-4xl">🗂️</span>
        <p className="text-sm text-gray-500">暂无错题，练习中答错的题会自动进入错题本</p>
        <Link
          to="/"
          className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-indigo-700"
        >
          去刷题
        </Link>
      </div>
    );
  } else {
    body = (
      <>
        <ul className="flex flex-col gap-3">
          {mistakesQuery.data.list.map((item) => (
            <MistakeCard
              key={item.question.id}
              item={item}
              removing={removeMutation.isPending && removeMutation.variables === item.question.id}
              onRemove={() => removeMutation.mutate(item.question.id)}
            />
          ))}
        </ul>
        <div className="mt-5">
          <Pagination
            page={page}
            pageSize={PAGE_SIZE}
            total={mistakesQuery.data.total}
            onChange={setPage}
          />
        </div>
      </>
    );
  }

  const subjects = subjectsQuery.data?.subjects ?? [];

  return (
    <main className="min-h-screen bg-gray-50">
      <div className="mx-auto max-w-4xl px-4 py-8">
        <nav className="text-sm">
          <Link to="/" className="text-indigo-600 hover:underline">
            ← 返回科目选择
          </Link>
        </nav>

        <header className="mt-3 flex flex-wrap items-end justify-between gap-3">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">错题本</h1>
            <p className="mt-1 text-sm text-gray-500">攻克薄弱知识点，直到错题清零</p>
          </div>
          <Link
            to="/mistakes/review"
            className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-indigo-700"
          >
            错题重刷
          </Link>
        </header>

        <section className="mt-5 grid gap-3 sm:grid-cols-[1fr_auto]">
          <div className="flex items-center gap-4 rounded-xl border border-gray-200 bg-white px-5 py-4 shadow-sm">
            <span className="text-3xl font-bold text-indigo-600">
              {mistakesQuery.data?.total ?? "–"}
            </span>
            <div className="text-sm">
              <p className="font-medium text-gray-900">
                道错题{subjectId ? " · 当前科目" : " · 全部科目"}
              </p>
              <p className="mt-0.5 text-xs text-gray-500">重刷通过后可手动移除，答错会自动累计</p>
            </div>
          </div>
          <div className="hidden items-center rounded-xl bg-gradient-to-br from-indigo-600 to-violet-600 px-5 py-4 text-sm text-white shadow-sm sm:flex">
            高频错题更醒目，优先消灭
          </div>
        </section>

        <div role="group" aria-label="按科目过滤" className="mt-5 flex flex-wrap gap-2 text-sm">
          <button
            type="button"
            onClick={() => {
              setSubjectId(0);
              setPage(1);
            }}
            className={`rounded-full border px-3 py-1.5 transition ${
              subjectId === 0
                ? "border-indigo-600 bg-indigo-600 text-white"
                : "border-gray-300 bg-white text-gray-600 hover:border-indigo-400"
            }`}
          >
            全部科目
          </button>
          {subjects.map((s) => (
            <button
              key={s.id}
              type="button"
              onClick={() => {
                setSubjectId(s.id);
                setPage(1);
              }}
              className={`rounded-full border px-3 py-1.5 transition ${
                subjectId === s.id
                  ? "border-indigo-600 bg-indigo-600 text-white"
                  : "border-gray-300 bg-white text-gray-600 hover:border-indigo-400"
              }`}
            >
              {s.name}
            </button>
          ))}
        </div>

        <div className="mt-5">{body}</div>
      </div>
    </main>
  );
}
