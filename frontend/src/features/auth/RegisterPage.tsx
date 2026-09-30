import { useState } from "react";
import { useNavigate } from "react-router";
import { useMutation } from "@tanstack/react-query";
import { login, register } from "../../api/auth";
import { useAuthStore } from "../../stores/auth";
import { AuthForm } from "./AuthForm";

export function RegisterPage() {
  const navigate = useNavigate();
  const setAuth = useAuthStore((s) => s.setAuth);
  const [error, setError] = useState<string | null>(null);

  // 注册成功不返回 token，直接用同一凭据登录，省一次手动输入
  const registerMutation = useMutation({
    mutationFn: async ({ username, password }: { username: string; password: string }) => {
      await register(username, password);
      return login(username, password);
    },
    onSuccess: ({ token, user }) => {
      setAuth(token, user);
      navigate("/", { replace: true });
    },
    onError: (err) => setError(err.message),
  });

  return (
    <AuthForm
      title="注册"
      submitLabel="注册并登录"
      pending={registerMutation.isPending}
      error={error}
      onSubmit={(username, password) => {
        setError(null);
        registerMutation.mutate({ username, password });
      }}
      link={{ text: "已有账号？", label: "去登录", to: "/login" }}
      passwordAutoComplete="new-password"
    />
  );
}
