import type { ReactNode } from "react";
import { Navigate, useLocation } from "react-router";
import { selectIsAuthenticated, useAuthStore } from "../stores/auth";

// 受保护路由守卫：未登录跳登录页并记住来路，登录成功后回跳
export function RequireAuth({ children }: { children: ReactNode }) {
  const isAuthenticated = useAuthStore(selectIsAuthenticated);
  const location = useLocation();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location.pathname + location.search }} />;
  }
  return children;
}
