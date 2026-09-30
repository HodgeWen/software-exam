import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { api, ApiError, parseApiError } from "./client";
import { useAuthStore } from "../stores/auth";

// 构造带 JSON 体的 fetch 响应；body 为 undefined 时不带体（非 JSON 响应）
function jsonResponse(status: number, body: unknown): Response {
  return new Response(body === undefined ? null : JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

describe("parseApiError 错误体解析", () => {
  it('解析统一错误体 {"code","message"}', () => {
    const err = parseApiError(409, { code: "USERNAME_EXISTS", message: "用户名已存在" });
    expect(err).toBeInstanceOf(ApiError);
    expect(err.status).toBe(409);
    expect(err.code).toBe("USERNAME_EXISTS");
    expect(err.message).toBe("用户名已存在");
  });

  it("残缺或非对象错误体给兜底", () => {
    for (const body of [null, undefined, "oops", { foo: 1 }, { code: 1, message: "x" }]) {
      const err = parseApiError(500, body);
      expect(err.code).toBe("UNKNOWN");
      expect(err.message).toContain("500");
    }
  });
});

describe("api 客户端", () => {
  beforeEach(() => {
    useAuthStore.setState({ token: null, user: null });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("成功响应解析为 JSON", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse(200, { id: 1, username: "u" })));

    await expect(api.get<unknown>("/me")).resolves.toEqual({ id: 1, username: "u" });
  });

  it("登录后自动注入 Authorization: Bearer", async () => {
    useAuthStore.setState({ token: "t123", user: null });
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(200, {}));
    vi.stubGlobal("fetch", fetchMock);

    await api.get("/me");

    const init = fetchMock.mock.calls[0]?.[1] as RequestInit;
    expect(new Headers(init.headers).get("Authorization")).toBe("Bearer t123");
  });

  it("post 序列化请求体为 JSON", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(200, {}));
    vi.stubGlobal("fetch", fetchMock);

    await api.post("/auth/login", { username: "u", password: "p123456" });

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe("/api/v1/auth/login");
    expect(init.method).toBe("POST");
    expect(new Headers(init.headers).get("Content-Type")).toBe("application/json");
    expect(init.body).toBe(JSON.stringify({ username: "u", password: "p123456" }));
  });

  it("未带 token 的 401（登录凭据错误）按业务错误抛出，不动登录态", async () => {
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValue(
          jsonResponse(401, { code: "INVALID_CREDENTIALS", message: "用户名或密码错误" }),
        ),
    );

    const err: unknown = await api
      .post("/auth/login", { username: "u", password: "bad" })
      .catch((e) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect((err as ApiError).code).toBe("INVALID_CREDENTIALS");
    expect((err as ApiError).message).toBe("用户名或密码错误");
    expect(useAuthStore.getState().token).toBeNull();
  });

  it("携带 token 收到 401：清空登录态并抛错误", async () => {
    useAuthStore.setState({ token: "expired", user: { id: 1, username: "u" } });
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValue(jsonResponse(401, { code: "UNAUTHORIZED", message: "token 已失效" })),
    );

    const err: unknown = await api.get("/me").catch((e) => e);
    expect((err as ApiError).status).toBe(401);
    expect((err as ApiError).code).toBe("UNAUTHORIZED");
    expect(useAuthStore.getState().token).toBeNull();
    expect(useAuthStore.getState().user).toBeNull();
  });

  it("非 401 错误透传统一错误体", async () => {
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValue(jsonResponse(409, { code: "USERNAME_EXISTS", message: "用户名已存在" })),
    );

    const err: unknown = await api
      .post("/auth/register", { username: "u", password: "p123456" })
      .catch((e) => e);
    expect((err as ApiError).status).toBe(409);
    expect((err as ApiError).code).toBe("USERNAME_EXISTS");
  });

  it("非 JSON 错误响应给兜底错误", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse(502, undefined)));

    const err: unknown = await api.get("/me").catch((e) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect((err as ApiError).code).toBe("UNKNOWN");
  });
});
