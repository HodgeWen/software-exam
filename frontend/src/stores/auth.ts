import { create } from "zustand";
import { persist } from "zustand/middleware";

export interface AuthUser {
  id: number;
  username: string;
}

interface AuthState {
  token: string | null;
  user: AuthUser | null;
  setAuth: (token: string, user: AuthUser) => void;
  clearAuth: () => void;
}

// 刷新页面不丢登录态：token/用户信息持久化到 localStorage
export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      setAuth: (token, user) => set({ token, user }),
      clearAuth: () => set({ token: null, user: null }),
    }),
    { name: "auth" },
  ),
);

// 登录态派生：守卫与组件统一用这个选择器，避免各写一遍判空
export const selectIsAuthenticated = (s: AuthState): boolean => s.token !== null;
