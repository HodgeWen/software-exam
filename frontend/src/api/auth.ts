import { api } from "./client";
import type { AuthUser } from "../stores/auth";

interface LoginResponse {
  token: string;
  user: AuthUser;
}

export function login(username: string, password: string) {
  return api.post<LoginResponse>("/auth/login", { username, password });
}

// 注册成功返回用户信息但不发 token，仍需走一次登录
export function register(username: string, password: string) {
  return api.post<AuthUser>("/auth/register", { username, password });
}

export function fetchMe() {
  return api.get<AuthUser>("/me");
}
