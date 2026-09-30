import { beforeEach, describe, expect, it } from "vitest";
import { selectIsAuthenticated, useAuthStore } from "./auth";

describe("auth store token 存取", () => {
  beforeEach(() => {
    localStorage.clear();
    useAuthStore.setState({ token: null, user: null });
  });

  it("初始未登录", () => {
    expect(useAuthStore.getState().token).toBeNull();
    expect(useAuthStore.getState().user).toBeNull();
    expect(selectIsAuthenticated(useAuthStore.getState())).toBe(false);
  });

  it("setAuth 写入 token 与用户信息并持久化", () => {
    useAuthStore.getState().setAuth("jwt-token", { id: 7, username: "tom" });

    const state = useAuthStore.getState();
    expect(state.token).toBe("jwt-token");
    expect(state.user).toEqual({ id: 7, username: "tom" });
    expect(selectIsAuthenticated(state)).toBe(true);
    // 持久化到 localStorage，刷新页面后由 persist 中间件恢复登录态
    expect(localStorage.getItem("auth")).toContain("jwt-token");
  });

  it("clearAuth 清空登录态并同步持久化", () => {
    useAuthStore.getState().setAuth("jwt-token", { id: 7, username: "tom" });
    useAuthStore.getState().clearAuth();

    const state = useAuthStore.getState();
    expect(state.token).toBeNull();
    expect(state.user).toBeNull();
    expect(selectIsAuthenticated(state)).toBe(false);
    const persisted = JSON.parse(localStorage.getItem("auth") ?? "{}") as {
      state?: { token?: string | null };
    };
    expect(persisted.state?.token).toBeNull();
  });
});
