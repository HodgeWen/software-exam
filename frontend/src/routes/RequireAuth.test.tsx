import { act, cleanup, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { MemoryRouter, Route, Routes, useLocation } from "react-router";
import { RequireAuth } from "./RequireAuth";
import { useAuthStore } from "../stores/auth";

function Protected() {
  return <p>受保护内容</p>;
}

// 复刻登录页对 location.state.from 的消费方式，验证守卫记住了来路
function LoginSpy() {
  const location = useLocation();
  const from = (location.state as { from?: string } | null)?.from ?? "";
  return <p>登录页，来路：{from}</p>;
}

function renderGuardAt(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route path="/login" element={<LoginSpy />} />
        <Route
          path="/"
          element={
            <RequireAuth>
              <Protected />
            </RequireAuth>
          }
        />
      </Routes>
    </MemoryRouter>,
  );
}

describe("RequireAuth 路由守卫", () => {
  beforeEach(() => {
    useAuthStore.setState({ token: null, user: null });
  });

  afterEach(cleanup);

  it("未登录访问受保护路由：跳登录页并记住来路", () => {
    renderGuardAt("/");
    expect(screen.queryByText("受保护内容")).toBeNull();
    expect(screen.getByText("登录页，来路：/")).toBeTruthy();
  });

  it("已登录：直接展示受保护内容", () => {
    useAuthStore.getState().setAuth("t", { id: 1, username: "u" });
    renderGuardAt("/");
    expect(screen.getByText("受保护内容")).toBeTruthy();
    expect(screen.queryByText(/登录页/)).toBeNull();
  });

  it("登录态失效（如 401 被清除）后立即跳回登录页", () => {
    useAuthStore.getState().setAuth("t", { id: 1, username: "u" });
    renderGuardAt("/");
    expect(screen.getByText("受保护内容")).toBeTruthy();

    act(() => useAuthStore.getState().clearAuth());

    expect(screen.queryByText("受保护内容")).toBeNull();
    expect(screen.getByText(/登录页/)).toBeTruthy();
  });
});
