import type { Question } from "../api/bank";

interface AnswerSheetProps {
  questions: Question[];
  // 考试作答暂存：questionId -> 所选选项集合
  answers: Record<number, string[]>;
  currentIndex: number;
  onJump: (index: number) => void;
}

// 答题卡：展示全部题号与已答/未答状态，点击题号跳到对应题目
export function AnswerSheet({ questions, answers, currentIndex, onJump }: AnswerSheetProps) {
  const answeredCount = questions.filter((q) => (answers[q.id] ?? []).length > 0).length;

  return (
    <aside className="h-fit rounded-lg border border-gray-200 bg-white p-4 lg:w-64">
      <p className="text-sm font-semibold">答题卡</p>
      <div className="mt-3 grid grid-cols-6 gap-2 lg:grid-cols-5">
        {questions.map((q, i) => {
          const answered = (answers[q.id] ?? []).length > 0;
          return (
            <button
              key={q.id}
              type="button"
              aria-pressed={answered}
              aria-current={i === currentIndex}
              onClick={() => onJump(i)}
              className={`h-9 rounded border text-sm ${
                answered
                  ? "border-indigo-600 bg-indigo-600 text-white"
                  : "border-gray-300 text-gray-700 hover:border-indigo-400"
              } ${i === currentIndex ? "font-bold ring-2 ring-indigo-400" : ""}`}
            >
              {q.no}
            </button>
          );
        })}
      </div>
      <p className="mt-3 text-xs text-gray-500">
        已答 {answeredCount} / {questions.length} 题
      </p>
    </aside>
  );
}
