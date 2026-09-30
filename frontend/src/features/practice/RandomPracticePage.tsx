import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router";
import { fetchRandomQuestions } from "../../api/bank";
import { PracticeRunner } from "./PracticeRunner";

// 随机练习页：消费随机取题接口（服务端打乱），复用章节练习的组件与交互；
// 不缓存旧批次，每次进入重新取一批
export function RandomPracticePage() {
  const { subjectId } = useParams();
  const sid = Number(subjectId);
  const valid = Number.isInteger(sid) && sid > 0;

  const questionsQuery = useQuery({
    queryKey: ["random-questions", sid],
    queryFn: () => fetchRandomQuestions(sid),
    enabled: valid,
    gcTime: 0,
  });

  return (
    <main className="mx-auto min-h-screen max-w-3xl p-6">
      <nav className="text-sm">
        <Link to={`/subjects/${sid}`} className="text-indigo-600 hover:underline">
          返回科目页
        </Link>
      </nav>
      <h1 className="mt-2 text-xl font-bold">随机练习</h1>
      <div className="mt-4">
        {!valid || questionsQuery.isPending ? (
          <p className="text-sm text-gray-600">加载中…</p>
        ) : questionsQuery.isError ? (
          <p role="alert" className="text-sm text-red-600">
            加载题目失败：{(questionsQuery.error as Error).message}
          </p>
        ) : questionsQuery.data.questions.length === 0 ? (
          <p className="text-sm text-gray-600">该科目暂无题目</p>
        ) : (
          <PracticeRunner title="随机练习" questions={questionsQuery.data.questions} />
        )}
      </div>
    </main>
  );
}
