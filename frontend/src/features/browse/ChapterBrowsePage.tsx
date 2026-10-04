import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router";
import { fetchChapterQuestionsRevealed } from "../../api/bank";
import { BrowseRunner } from "./BrowseRunner";

// 章节背题页：拉取带答案与解析的整章题目，逐题浏览不做答
export function ChapterBrowsePage() {
  const { subjectId, chapterId } = useParams();
  const sid = Number(subjectId);
  const cid = Number(chapterId);
  const valid = Number.isInteger(sid) && sid > 0 && Number.isInteger(cid) && cid > 0;

  const questionsQuery = useQuery({
    queryKey: ["chapter-questions-revealed", sid, cid],
    queryFn: () => fetchChapterQuestionsRevealed(sid, cid),
    enabled: valid,
  });

  return (
    <main className="mx-auto min-h-screen max-w-3xl p-6">
      <nav className="text-sm">
        <Link to={`/subjects/${sid}`} className="text-indigo-600 hover:underline">
          返回科目页
        </Link>
      </nav>
      <h1 className="mt-2 text-xl font-bold">背题模式</h1>
      <div className="mt-4">
        {!valid || questionsQuery.isPending ? (
          <p className="text-sm text-gray-600">加载中…</p>
        ) : questionsQuery.isError ? (
          <p role="alert" className="text-sm text-red-600">
            加载题目失败：{(questionsQuery.error as Error).message}
          </p>
        ) : questionsQuery.data.questions.length === 0 ? (
          <p className="text-sm text-gray-600">本章暂无题目</p>
        ) : (
          <BrowseRunner title="章节背题" questions={questionsQuery.data.questions} />
        )}
      </div>
    </main>
  );
}
