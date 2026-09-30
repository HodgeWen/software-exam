# 开发规范

## 命名

- 前端：组件文件 `PascalCase.tsx`；hooks 文件 `use-xxx.ts` 导出 `useXxx`；其余 ts 文件 `camelCase.ts`；目录 `kebab-case`
- 后端：Go 官方惯例；文件名小写单词；数据库表/字段 `snake_case`（GORM 默认映射）

## 目录与代码结构

- 前端：功能页放 `src/features/<domain>/`，域内组件/hooks/类型就近放；`src/routes/` 只放路由装配、登录守卫与科目选择页；跨域复用组件进 `src/components/`；所有后端调用集中在 `src/api/`
- 后端：`cmd/server/` 只放入口装配；`internal/handler`（参数校验/响应）→ `internal/service`（业务）→ `internal/repository`（GORM 访问）；handler 不越过 service 直接触达 repository

## 代码风格

- 前端格式化与 lint 由 vite-plus 内置 Oxfmt / Oxlint 承担（`vp check`），不引入 Prettier / ESLint
- 后端 `gofmt` + `go vet`，不引入其它格式化工具
- 注释语言：中文；只写解释「为什么」的注释

## 测试

- 后端：`go test ./...`；评分、错题判定、JWT 等核心逻辑必须有单测；handler 用 httptest + 内存 SQLite 写集成测试
- 前端：Vitest（`vp test`）；工具函数与核心交互逻辑必须写，纯展示组件不强制

## 接口

- REST，前缀 `/api/v1`，请求/响应均 JSON
- 错误响应统一 `{"code": "<业务错误码>", "message": "<人读信息>"}`，配恰当 HTTP 状态码
- 列表接口分页：入参 `page` / `page_size`，响应含 `total`

## 数据与存储

- SQLite 单文件（git 忽略 `*.db`）；GORM AutoMigrate 建表
- 种子题库走可重复执行的幂等导入

## 日志

- 后端：标准库 `slog`，结构化 JSON
- 前端：仅开发期 console，生产不留

## 明确禁止

- cooking `spec.md` 缺少可被 `spec-files.mjs parse` 通过的「影响文件」章节
- 前端组件内直接调用 `fetch`（必须走 `src/api/`）
- 提交 `node_modules/`、`dist/`、`*.db`、`.env*`
- 后端 handler 层写 SQL 或业务逻辑
