import { useState } from "react";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { OptionGroup } from "./OptionGroup";
import type { QuestionOption } from "../api/bank";

const options: QuestionOption[] = [
  { key: "A", text: "选项A" },
  { key: "B", text: "选项B" },
  { key: "C", text: "选项C" },
];

// 选项按钮的无障碍名是「标号 + 选项文本」（如 "A 选项A"），用子串匹配点选
function clickOption(text: string) {
  fireEvent.click(screen.getByRole("button", { name: new RegExp(text) }));
}

// 用受控父组件挂载：onChange 后回填 selected，模拟真实页面里的选项组
function setup(initial: string[], props: Partial<Parameters<typeof OptionGroup>[0]> = {}) {
  const onChange = vi.fn();
  function Harness() {
    const [selected, setSelected] = useState(initial);
    return (
      <OptionGroup
        options={options}
        multiple={false}
        {...props}
        selected={selected}
        onChange={(next) => {
          onChange(next);
          setSelected(next);
        }}
      />
    );
  }
  render(<Harness />);
  return onChange;
}

afterEach(cleanup);

describe("OptionGroup 选项组交互", () => {
  it("单选：点击选项替换所选", () => {
    const onChange = setup([]);

    clickOption("选项B");
    expect(onChange).toHaveBeenCalledWith(["B"]);

    clickOption("选项A");
    expect(onChange).toHaveBeenCalledWith(["A"]);
  });

  it("多选：点击勾选/取消，可勾选多项", () => {
    const onChange = setup(["B"], { multiple: true });

    clickOption("选项A");
    expect(onChange).toHaveBeenCalledWith(["A", "B"]);

    clickOption("选项B");
    expect(onChange).toHaveBeenCalledWith(["A"]);

    clickOption("选项C");
    expect(onChange).toHaveBeenCalledWith(["A", "C"]);
  });

  it("disabled（已判分）：点击不再触发 onChange", () => {
    const onChange = setup(["A"], { disabled: true });

    clickOption("选项B");
    expect(onChange).not.toHaveBeenCalled();
  });

  it("判分后按正确答案着色：正确项绿色、错选项红色、其余弱化", () => {
    setup(["B"], { disabled: true, correctAnswer: ["A"] });

    expect(screen.getByRole("button", { name: /选项A/ }).className).toContain("border-green-600");
    expect(screen.getByRole("button", { name: /选项B/ }).className).toContain("border-red-600");
    expect(screen.getByRole("button", { name: /选项C/ }).className).toContain("border-gray-200");
  });

  it("选中态经 aria-pressed 暴露", () => {
    setup(["A"]);

    expect(screen.getByRole("button", { name: /选项A/ }).getAttribute("aria-pressed")).toBe("true");
    expect(screen.getByRole("button", { name: /选项B/ }).getAttribute("aria-pressed")).toBe(
      "false",
    );
  });
});
