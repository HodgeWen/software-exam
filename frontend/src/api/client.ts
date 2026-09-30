import { useAuthStore } from "../stores/auth";

// 统一错误体 {"code","message"} 对应的异常；组件只 catch ApiError 展示 message
export class ApiError extends Error {
  constructor(
    readonly status: number,
    readonly code: string,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

interface ErrorBody {
  code?: unknown;
  message?: unknown;
}

// 解析后端统一错误体；残缺/非 JSON 时给出兜底，避免把 undefined 抛给页面
export function parseApiError(status: number, body: unknown): ApiError {
  if (
    typeof body === "object" &&
    body !== null &&
    typeof (body as ErrorBody).code === "string" &&
    typeof (body as ErrorBody).message === "string"
  ) {
    const { code, message } = body as ErrorBody as { code: string; message: string };
    return new ApiError(status, code, message);
  }
  return new ApiError(status, "UNKNOWN", `请求失败（${status}）`);
}

const baseUrl = "/api/v1";

// 全站唯一 fetch 出口：自动注入 Authorization、解析统一错误体、401 统一处理
async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const { token, clearAuth } = useAuthStore.getState();
  const headers = new Headers(init?.headers);
  if (token) headers.set("Authorization", `Bearer ${token}`);
  if (init?.body !== undefined) headers.set("Content-Type", "application/json");

  const res = await fetch(`${baseUrl}${path}`, { ...init, headers });

  if (res.ok) {
    return res.status === 204 ? (undefined as T) : ((await res.json()) as T);
  }

  // 携带 token 仍 401 说明登录态失效：清空后由路由守卫跳登录页。
  // 未携带 token 的 401（如登录接口凭据错误）按普通业务错误抛给表单展示。
  if (res.status === 401 && token) {
    clearAuth();
  }
  throw parseApiError(res.status, await res.json().catch(() => null));
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: "POST",
      body: body === undefined ? undefined : JSON.stringify(body),
    }),
  put: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: "PUT",
      body: body === undefined ? undefined : JSON.stringify(body),
    }),
  delete: <T>(path: string) => request<T>(path, { method: "DELETE" }),
};
