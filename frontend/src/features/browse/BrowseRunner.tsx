import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { addFavorite } from "../../api/favorites";
import type { RevealedQuestion } from "../../api/bank";
import { QuestionCard } from "../../components/QuestionCard";

interface BrowseRunnerProps {
  title: string;
  questions: RevealedQuestion[];
}

// 背题模式共用的逐题浏览流程：不做答、直接展示题目与答案解析；
// 可一键遮住答案自测，上一题/下一题翻页，进度与收藏态只存组件内
export function BrowseRunner({ title, questions }: BrowseRunnerProps) {
  const [index, setIndex] = useState(0);
  const [hidden, setHidden] = useState(false);

  const total = questions.length;
  const question = questions[index];

  const favoriteMutation = useMutation({
    mutationFn: (q: RevealedQuestion) => addFavorite(q.id),
  });
  const favorited =
    !!question && favoriteMutation.isSuccess && favoriteMutation.variables?.id === question.id;

  if (!question) return null;

  return (
    <section className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <p className="text-sm text-gray-600">
          {title} · 第 {index + 1} / {total} 题
        </p>
        <button
          type="button"
          onClick={() => setHidden((v) => !v)}
          className="rounded border border-indigo-300 px-3 py-1.5 text-sm text-indigo-600 hover:bg-indigo-50"
        >
          {hidden ? "显示答案" : "隐藏答案自测"}
        </button>
      </div>
      <QuestionCard
        question={question}
        selected={[]}
        result={null}
        onChange={() => {}}
        browse={{ answer: question.answer, analysis: question.analysis, hidden }}
      />
      <div className="flex flex-wrap items-center gap-3">
        <button
          type="button"
          disabled={index === 0}
          onClick={() => setIndex(index - 1)}
          className="rounded border border-gray-300 px-4 py-2 text-sm hover:bg-gray-50 disabled:opacity-50"
        >
          上一题
        </button>
        <button
          type="button"
          disabled={index + 1 >= total}
          onClick={() => setIndex(index + 1)}
          className="rounded bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
        >
          {index + 1 === total ? "已是最后一题" : "下一题"}
        </button>
        <button
          type="button"
          disabled={favoriteMutation.isPending || favorited}
          onClick={() => favoriteMutation.mutate(question)}
          className="rounded border border-indigo-300 px-4 py-2 text-sm text-indigo-600 hover:bg-indigo-50 disabled:opacity-50"
        >
          {favorited ? "已收藏" : favoriteMutation.isPending ? "收藏中…" : "收藏本题"}
        </button>
      </div>
      {favoriteMutation.isError && (
        <p role="alert" className="text-sm text-red-600">
          收藏失败：{(favoriteMutation.error as Error).message}
        </p>
      )}
    </section>
  );
}
