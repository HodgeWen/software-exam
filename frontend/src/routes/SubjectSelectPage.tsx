import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router";
import { fetchSubjects } from "../api/bank";
import { useAuthStore } from "../stores/auth";

// 科目选择页：刷题入口首页，选择科目后进入章节练习/随机练习/真题模拟
export function SubjectSelectPage() {
  const user = useAuthStore((s) => s.user);
  const clearAuth = useAuthStore((s) => s.clearAuth);
  const subjectsQuery = useQuery({ queryKey: ["subjects"], queryFn: fetchSubjects });

  return (
    <main className="mx-auto min-h-screen max-w-3xl p-6">
      <header className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">软考刷题</h1>
        <div className="flex items-center gap-3 text-sm">
          <span className="text-gray-600">{user?.username}</span>
          <button
            type="button"
            onClick={clearAuth}
            className="rounded bg-gray-200 px-3 py-1.5 hover:bg-gray-300"
          >
            退出登录
          </button>
        </div>
      </header>

      <h2 className="mt-8 mb-3 text-lg font-semibold">选择科目</h2>
      {subjectsQuery.isPending ? (
        <p className="text-sm text-gray-600">加载中…</p>
      ) : subjectsQuery.isError ? (
        <p role="alert" className="text-sm text-red-600">
          加载科目失败：{(subjectsQuery.error as Error).message}
        </p>
      ) : subjectsQuery.data.subjects.length === 0 ? (
        <p className="text-sm text-gray-600">暂无科目</p>
      ) : (
        <ul className="grid gap-3 sm:grid-cols-2">
          {subjectsQuery.data.subjects.map((s) => (
            <li key={s.id}>
              <Link
                to={`/subjects/${s.id}`}
                className="block rounded-lg border border-gray-200 p-4 hover:border-indigo-400 hover:bg-indigo-50"
              >
                <span className="font-medium">{s.name}</span>
                <span className="mt-1 block text-xs text-gray-500">{s.code}</span>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </main>
  );
}
