import { useState } from "react";
import type { FormEvent } from "react";
import { Link } from "react-router";

interface AuthLink {
  text: string;
  label: string;
  to: string;
}

interface AuthFormProps {
  title: string;
  submitLabel: string;
  pending: boolean;
  error: string | null;
  onSubmit: (username: string, password: string) => void;
  link: AuthLink;
  passwordAutoComplete?: "current-password" | "new-password";
}

// 登录/注册共用同一组字段（与后端 credentials 校验一致），共用一个表单壳
export function AuthForm({
  title,
  submitLabel,
  pending,
  error,
  onSubmit,
  link,
  passwordAutoComplete = "current-password",
}: AuthFormProps) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    onSubmit(username, password);
  }

  return (
    <main className="flex min-h-screen items-center justify-center bg-gray-50 p-6">
      <div className="w-full max-w-sm rounded-lg bg-white p-8 shadow">
        <h1 className="mb-6 text-center text-xl font-bold">{title}</h1>
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <label className="flex flex-col gap-1 text-sm">
            用户名
            <input
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              minLength={2}
              maxLength={64}
              autoComplete="username"
              className="rounded border border-gray-300 px-3 py-2"
            />
          </label>
          <label className="flex flex-col gap-1 text-sm">
            密码
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={6}
              maxLength={72}
              autoComplete={passwordAutoComplete}
              className="rounded border border-gray-300 px-3 py-2"
            />
          </label>
          {/* 错误槽位常驻固定高度，有无错误都占同一块空间，避免接口错误挤开下方按钮 */}
          <p role="alert" className="min-h-5 text-sm text-red-600">
            {error}
          </p>
          <button
            type="submit"
            disabled={pending}
            className="rounded bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
          >
            {submitLabel}
          </button>
        </form>
        <p className="mt-4 text-center text-sm text-gray-600">
          {link.text}
          <Link to={link.to} className="ml-1 text-indigo-600 hover:underline">
            {link.label}
          </Link>
        </p>
      </div>
    </main>
  );
}
