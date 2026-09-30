import { useState } from "react";
import type { ReactNode } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "react-router";
import { fetchFavorites, removeFavorite } from "../../api/favorites";
import { Pagination } from "../../components/Pagination";

const PAGE_SIZE = 10;

// 收藏页：收藏题目分页列表，可取消收藏
export function FavoritesPage() {
  const [page, setPage] = useState(1);
  const queryClient = useQueryClient();

  const favoritesQuery = useQuery({
    queryKey: ["favorites", page],
    queryFn: () => fetchFavorites(page, PAGE_SIZE),
  });

  const removeMutation = useMutation({
    mutationFn: removeFavorite,
    onSuccess: () => {
      // 取消的是本页最后一条时先回退一页，避免翻页请求落在空页上
      if ((favoritesQuery.data?.list.length ?? 0) <= 1 && page > 1) {
        setPage((p) => p - 1);
      }
      void queryClient.invalidateQueries({ queryKey: ["favorites"] });
    },
  });

  let body: ReactNode;
  if (favoritesQuery.isPending) {
    body = <p className="text-sm text-gray-600">加载中…</p>;
  } else if (favoritesQuery.isError) {
    body = (
      <p role="alert" className="text-sm text-red-600">
        加载收藏失败：{(favoritesQuery.error as Error).message}
      </p>
    );
  } else if (favoritesQuery.data.list.length === 0) {
    body = <p className="text-sm text-gray-600">暂无收藏，练习时点击「收藏本题」即可收藏</p>;
  } else {
    body = (
      <>
        <ul className="flex flex-col gap-3">
          {favoritesQuery.data.list.map((item) => (
            <li
              key={item.question.id}
              className="flex items-start justify-between gap-4 rounded border border-gray-200 bg-white p-4"
            >
              <div className="min-w-0">
                <p className="flex items-center gap-2 text-xs text-gray-500">
                  <span>{item.question.type === "multiple" ? "多选" : "单选"}</span>
                  <span>收藏于 {new Date(item.created_at).toLocaleString()}</span>
                </p>
                <p className="mt-1 line-clamp-2 text-sm font-medium">{item.question.stem}</p>
              </div>
              <button
                type="button"
                disabled={removeMutation.isPending && removeMutation.variables === item.question.id}
                onClick={() => removeMutation.mutate(item.question.id)}
                className="shrink-0 rounded border border-red-300 px-3 py-1.5 text-sm text-red-600 hover:bg-red-50 disabled:opacity-50"
              >
                取消收藏
              </button>
            </li>
          ))}
        </ul>
        <Pagination
          page={page}
          pageSize={PAGE_SIZE}
          total={favoritesQuery.data.total}
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
      <h1 className="mt-2 text-xl font-bold">我的收藏</h1>
      <div className="mt-4">{body}</div>
    </main>
  );
}
