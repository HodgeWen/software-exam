import type { ReactNode } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router";
import { fetchSubjectDetail } from "../api/bank";

// 科目页：章节练习入口、随机练习入口与真题试卷列表（模拟考试入口）
export function SubjectPage() {
  const { subjectId } = useParams();
  const sid = Number(subjectId);
  const valid = Number.isInteger(sid) && sid > 0;

  const detailQuery = useQuery({
    queryKey: ["subject", sid],
    queryFn: () => fetchSubjectDetail(sid),
    enabled: valid,
  });

  let body: ReactNode;
  if (!valid || detailQuery.isPending) {
    body = <p className="text-sm text-gray-600">加载中…</p>;
  } else if (detailQuery.isError) {
    body = (
      <p role="alert" className="text-sm text-red-600">
        加载科目失败：{(detailQuery.error as Error).message}
      </p>
    );
  } else {
    const { subject, chapters, papers } = detailQuery.data;
    body = (
      <>
        <h1 className="text-xl font-bold">{subject.name}</h1>

        <section className="mt-6">
          <h2 className="mb-2 font-semibold">章节练习</h2>
          {chapters.length === 0 ? (
            <p className="text-sm text-gray-600">暂无章节</p>
          ) : (
            <ul className="flex flex-col gap-2">
              {chapters.map((c) => (
                <li
                  key={c.id}
                  className="flex items-center justify-between gap-3 rounded border border-gray-200 px-4 py-3 text-sm hover:border-indigo-400 hover:bg-indigo-50"
                >
                  <Link to={`/subjects/${sid}/chapters/${c.id}`} className="min-w-0 flex-1">
                    {c.name}
                  </Link>
                  <Link
                    to={`/subjects/${sid}/chapters/${c.id}/browse`}
                    className="shrink-0 text-gray-500 hover:text-indigo-600"
                  >
                    背题
                  </Link>
                  <Link
                    to={`/subjects/${sid}/chapters/${c.id}`}
                    className="shrink-0 text-indigo-600"
                  >
                    去练习
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </section>

        <section className="mt-6">
          <h2 className="mb-2 font-semibold">随机练习</h2>
          <Link
            to={`/subjects/${sid}/random`}
            className="flex items-center justify-between rounded border border-gray-200 px-4 py-3 text-sm hover:border-indigo-400 hover:bg-indigo-50"
          >
            <span>随机练习（打乱本科目全部题目）</span>
            <span className="text-indigo-600">开始</span>
          </Link>
        </section>

        <section className="mt-6">
          <h2 className="mb-2 font-semibold">真题模拟</h2>
          {papers.length === 0 ? (
            <p className="text-sm text-gray-600">暂无试卷</p>
          ) : (
            <ul className="flex flex-col gap-2">
              {papers.map((p) => (
                <li
                  key={p.id}
                  className="flex items-center justify-between gap-3 rounded border border-gray-200 px-4 py-3 text-sm hover:border-indigo-400 hover:bg-indigo-50"
                >
                  <Link to={`/papers/${p.id}/exam`} className="min-w-0 flex-1">
                    {p.name}
                    <span className="ml-2 text-xs text-gray-500">{p.duration_minutes} 分钟</span>
                  </Link>
                  <Link
                    to={`/papers/${p.id}/browse`}
                    className="shrink-0 text-gray-500 hover:text-indigo-600"
                  >
                    学习模式
                  </Link>
                  <Link to={`/papers/${p.id}/exam`} className="shrink-0 text-indigo-600">
                    开始考试
                  </Link>
                </li>
              ))}
            </ul>
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
      <div className="mt-4">{body}</div>
    </main>
  );
}
