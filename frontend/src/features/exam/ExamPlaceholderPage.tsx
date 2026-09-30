import { Link } from "react-router";

// 真题模拟考试路由占位页：科目页的试卷入口先接到这里，P9 交付整卷模拟（倒计时+答题卡+交卷评分）后替换
export function ExamPlaceholderPage() {
  return (
    <main className="mx-auto min-h-screen max-w-3xl p-6">
      <nav className="text-sm">
        <Link to="/" className="text-indigo-600 hover:underline">
          返回科目选择
        </Link>
      </nav>
      <h1 className="mt-2 text-xl font-bold">真题模拟考试</h1>
      <p className="mt-4 text-sm text-gray-600">整卷模拟功能将在后续阶段开放。</p>
    </main>
  );
}
