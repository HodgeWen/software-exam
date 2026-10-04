import type { AnswerResult, Question } from "../api/bank";
import { OptionGroup } from "./OptionGroup";

interface QuestionCardProps {
  question: Question;
  selected: string[];
  // 非 null 表示该题已判分：选项锁定并展示正确答案与解析
  result: AnswerResult | null;
  onChange: (next: string[]) => void;
  // 背题模式：传入即不做答、选项锁定并直接亮出正确答案与解析；
  // hidden 为 true 时遮住答案区（自测），选项锁定且不高亮
  browse?: { answer: string[]; analysis: string; hidden: boolean };
}

// 题目卡片：题型标签 + 题干 + 选项组 +（判分后的）对错、正确答案与解析展示；
// 练习与错题重刷共用，考试页作答态也可复用（不传 result）；browse 供背题模式复用
export function QuestionCard({ question, selected, result, onChange, browse }: QuestionCardProps) {
  const multiple = question.type === "multiple";

  if (browse) {
    return (
      <article className="rounded-lg border border-gray-200 bg-white p-5">
        <span className="rounded bg-gray-100 px-2 py-0.5 text-xs text-gray-600">
          {multiple ? "多选" : "单选"}
        </span>
        <p className="mt-2 whitespace-pre-wrap font-medium">{question.stem}</p>
        <div className="mt-4">
          <OptionGroup
            options={question.options}
            multiple={multiple}
            selected={[]}
            disabled
            correctAnswer={browse.hidden ? undefined : browse.answer}
            onChange={onChange}
          />
        </div>
        {!browse.hidden && (
          <div className="mt-4 rounded border border-indigo-200 bg-indigo-50 p-3 text-sm">
            <p className="font-medium text-indigo-700">正确答案：{browse.answer.join("、")}</p>
            <p className="mt-1 whitespace-pre-wrap text-gray-700">解析：{browse.analysis}</p>
          </div>
        )}
      </article>
    );
  }

  return (
    <article className="rounded-lg border border-gray-200 bg-white p-5">
      <span className="rounded bg-gray-100 px-2 py-0.5 text-xs text-gray-600">
        {multiple ? "多选" : "单选"}
      </span>
      <p className="mt-2 whitespace-pre-wrap font-medium">{question.stem}</p>
      <div className="mt-4">
        <OptionGroup
          options={question.options}
          multiple={multiple}
          selected={selected}
          disabled={result !== null}
          correctAnswer={result?.answer}
          onChange={onChange}
        />
      </div>
      {result && (
        <div
          className={`mt-4 rounded border p-3 text-sm ${
            result.correct ? "border-green-300 bg-green-50" : "border-red-300 bg-red-50"
          }`}
        >
          <p className={`font-medium ${result.correct ? "text-green-700" : "text-red-700"}`}>
            {result.correct ? "回答正确" : "回答错误"}
          </p>
          <p className="mt-1">正确答案：{result.answer.join("、")}</p>
          <p className="mt-1 whitespace-pre-wrap text-gray-700">解析：{result.analysis}</p>
        </div>
      )}
    </article>
  );
}
