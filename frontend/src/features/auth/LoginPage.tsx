import { useState } from "react";
import { useLocation, useNavigate } from "react-router";
import { useMutation } from "@tanstack/react-query";
import { login } from "../../api/auth";
import { useAuthStore } from "../../stores/auth";
import { AuthForm } from "./AuthForm";

export function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const setAuth = useAuthStore((s) => s.setAuth);
  const [error, setError] = useState<string | null>(null);

  const loginMutation = useMutation({
    mutationFn: ({ username, password }: { username: string; password: string }) =>
      login(username, password),
    onSuccess: ({ token, user }) => {
      setAuth(token, user);
      // 回跳守卫记录的来路，没有来路则进首页
      const from = (location.state as { from?: string } | null)?.from;
      navigate(from ?? "/", { replace: true });
    },
    onError: (err) => setError(err.message),
  });

  return (
    <AuthForm
      title="登录"
      submitLabel="登录"
      pending={loginMutation.isPending}
      error={error}
      onSubmit={(username, password) => {
        setError(null);
        loginMutation.mutate({ username, password });
      }}
      link={{ text: "还没有账号？", label: "去注册", to: "/register" }}
    />
  );
}
