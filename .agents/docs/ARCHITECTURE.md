# 架构

## 业务架构

软考在线刷题网站，面向备考全国计算机技术与软件专业技术资格（水平）考试（软考）初级/中级/高级的考生。

核心域：

- 题库：科目 → 章节练习题 / 真题试卷 → 题目（单选/多选），含答案与解析
- 刷题练习：按科目/章节顺序或随机练习，逐题提交即时判分与解析
- 模拟考试：整卷计时作答、答题卡跳题、交卷统一评分
- 错题本：答错自动入库，支持错题重刷与移除
- 收藏：题目收藏与取消
- 统计：答题量、正确率、按科目/章节分布
- 账号：注册/登录（JWT），刷题进度与错题存服务端

主要业务流程：登录 → 选科目 → 选模式（章节练习 / 随机练习 / 真题模拟 / 错题重刷）→ 作答 → 判分与解析 → 错题与统计自动沉淀。

首批题库：软件设计师（中级）示例题；数据模型按多科目可扩展设计。

## 技术架构

前后端分离（与 PROJECT.md 一致）：

- `frontend/`：React SPA，Vite+ 统一工具链开发与构建，产物为纯静态文件
- `backend/`：Go REST API（Gin），前缀 `/api/v1`；除注册/登录等公开接口外走 JWT Bearer 鉴权
- 数据：SQLite 单文件（GORM + 纯 Go 驱动，无 CGO），AutoMigrate 建表 + 种子脚本导入题库
- 开发期前端由 Vite 代理 `/api` 到后端；生产部署形态见「未决」

### 技术栈

前端（`frontend/`）：

| 层 | 选型 | 备注 |
| --- | --- | --- |
| 语言 / runtime | TypeScript 7.0（原生编译器 tsgo）/ Node.js 24 | 类型检查经 vite-plus 集成的 tsgolint |
| 框架 | React 19.3 + react-router 8.4 | SPA |
| 状态 / 数据请求 | zustand 5（客户端状态）+ @tanstack/react-query 5（服务端状态） | |
| 样式 | Tailwind CSS 4.3 | 手写轻组件，不引重组件库 |
| 构建 / 包管理 | vite-plus 1.0（`vp` CLI：dev/check/test/build，内含 Vite 8.3 + Rolldown、Oxlint、Oxfmt、Vitest 5）/ npm | 2026-09 发布的统一工具链 |
| 测试 | Vitest 5（`vp test`） | |

后端（`backend/`）：

| 层 | 选型 | 备注 |
| --- | --- | --- |
| 语言 / runtime | Go 1.27 | |
| 框架 | Gin 1.12 | REST JSON |
| 数据 | SQLite（glebarez/sqlite 纯 Go 驱动）+ GORM 1.31 | 单文件 `*.db`，AutoMigrate + 种子 |
| 鉴权 | golang-jwt/jwt v5.3 | Bearer Token |
| 测试 | go test + httptest（内存 SQLite） | |

## 未决

- 生产部署形态：前端静态托管 + 独立 API，还是 Go 二进制直接托管前端 dist
- 真实题库数据的来源与批量导入格式
