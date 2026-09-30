import { useAuthStore } from "../stores/auth";

// 受保护首页占位：P8 交付科目选择页后替换
export function HomePage() {
  const user = useAuthStore((s) => s.user);
  const clearAuth = useAuthStore((s) => s.clearAuth);

  return (
    <main className="mx-auto flex min-h-screen max-w-3xl flex-col items-center justify-center gap-6 p-6">
      <h1 className="text-2xl font-bold">软考刷题</h1>
      <p className="text-gray-600">已登录：{user?.username ?? "未知用户"}，科目选择页即将上线</p>
      <button
        type="button"
        onClick={clearAuth}
        className="rounded bg-gray-200 px-4 py-2 text-sm hover:bg-gray-300"
      >
        退出登录
      </button>
    </main>
  );
}
