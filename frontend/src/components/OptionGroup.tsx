import type { QuestionOption } from "../api/bank";

interface OptionGroupProps {
  options: QuestionOption[];
  multiple: boolean;
  selected: string[];
  disabled?: boolean;
  // 判分后传入正确选项：正确项标绿、错选项标红；不传按普通作答态渲染
  correctAnswer?: string[];
  onChange: (next: string[]) => void;
}

// 单选/多选选项组：受控组件，点击时算好下一份所选集合交给 onChange——
// 单选替换所选，多选勾选/取消（可勾选多项），选择逻辑只在这里维护一份
export function OptionGroup({
  options,
  multiple,
  selected,
  disabled = false,
  correctAnswer,
  onChange,
}: OptionGroupProps) {
  const judged = correctAnswer !== undefined;

  function handleClick(key: string) {
    if (disabled) return;
    if (!multiple) {
      onChange([key]);
      return;
    }
    const next = selected.includes(key) ? selected.filter((k) => k !== key) : [...selected, key];
    next.sort();
    onChange(next);
  }

  function optionClass(key: string): string {
    const base = "flex w-full items-start gap-3 rounded border px-3 py-2 text-left text-sm";
    if (correctAnswer?.includes(key)) return `${base} border-green-600 bg-green-50`;
    if (judged && selected.includes(key)) return `${base} border-red-600 bg-red-50`;
    if (judged) return `${base} border-gray-200 text-gray-500`;
    if (selected.includes(key)) return `${base} border-indigo-600 bg-indigo-50`;
    return `${base} border-gray-300 hover:bg-gray-50`;
  }

  return (
    <div className="flex flex-col gap-2">
      {options.map((o) => (
        <button
          key={o.key}
          type="button"
          aria-pressed={selected.includes(o.key)}
          disabled={disabled}
          onClick={() => handleClick(o.key)}
          className={optionClass(o.key)}
        >
          <span className="mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full border border-current text-xs">
            {o.key}
          </span>
          <span className="whitespace-pre-wrap">{o.text}</span>
        </button>
      ))}
    </div>
  );
}
