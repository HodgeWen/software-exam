import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite-plus";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    // 开发期把 /api 代理到本地 Go 后端（backend 默认 :8080），前端代码全部走相对路径
    proxy: {
      "/api": "http://localhost:8080",
    },
  },
  test: {
    // 默认 node；需要 DOM 的测试文件用 // @vitest-environment happy-dom 头注解单独声明
    environment: "node",
    // 持久化 transform 结果到文件系统，跨进程复用，加快冷启动
    fsModuleCache: true,
  },
});
