import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router";
import { fetchPaperQuestionsRevealed } from "../../api/bank";
import { BrowseRunner } from "./BrowseRunner";

// 试卷学习页：整卷带答案与解析逐题浏览，供不做卷先学真题用
export function PaperBrowsePage() {
  const { paperId } = useParams();
  const pid = Number(paperId);
  const valid = Number.isInteger(pid) && pid > 0;

  const paperQuery = useQuery({
    queryKey: ["paper-revealed", pid],
    queryFn: () => fetchPaperQuestionsRevealed(pid),
    enabled: valid,
  });

  return (
    <main className="mx-auto min-h-screen max-w-3xl p-6">
      <nav className="text-sm">
        <Link to="/" className="text-indigo-600 hover:underline">
          返回科目选择
        </Link>
      </nav>
      <h1 className="mt-2 text-xl font-bold">
        {paperQuery.data ? paperQuery.data.paper.name : "真题学习"}
      </h1>
      <div className="mt-4">
        {!valid || paperQuery.isPending ? (
          <p className="text-sm text-gray-600">加载中…</p>
        ) : paperQuery.isError ? (
          <p role="alert" className="text-sm text-red-600">
            加载试卷失败：{(paperQuery.error as Error).message}
          </p>
        ) : paperQuery.data.questions.length === 0 ? (
          <p className="text-sm text-gray-600">本卷暂无题目</p>
        ) : (
          <BrowseRunner title="真题学习" questions={paperQuery.data.questions} />
        )}
      </div>
    </main>
  );
}
