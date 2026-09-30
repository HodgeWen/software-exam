---
name: react
description: React 19 开发实践。按项目 React minor（优先 19.0 / 19.1 / 19.2）选用该版本 API。在编辑 .jsx / .tsx、React 组件、Hooks 或处理 react / react-dom 时使用。
---

# React 19

默认函数组件 + Hooks。路由、数据请求、CSS、状态库、构建工具跟项目走，不要在本技能里指定 Next / Remix / Vite。

禁止凭训练数据写 API。先定 minor，再读对应 reference。

## 定版本

1. 读已安装的 `react` 版本（`node_modules/react/package.json` 的 `version`，或 lockfile）。不要只看 `package.json` 的 `^` 范围。`react` 与 `react-dom` 必须同一 minor。
2. 取 **minor**（`19.2.8` → `19.2`），打开下表文件，按该文件写代码。

| minor | 文件 |
| ----- | ---- |
| 19.0 | [references/19.0.md](references/19.0.md) |
| 19.1 | [references/19.1.md](references/19.1.md) |
| 19.2 | [references/19.2.md](references/19.2.md) |

未列出的更新 minor：以已覆盖的最高 minor 为底，再查官方该 minor 的 blog / changelog。更旧：不要使用本技能里的新 API。React 18 不要套本技能的 19 API。

## 不要用 Effect 同步 React 内部状态

Effect 只用来同步**外部系统**（DOM 插件、WebSocket、非 React widget）。能在渲染里算的，不要放进 `useEffect` + `setState`。

| 想做的事 | 写法 |
| -------- | ---- |
| 从 props/state 派生 | 渲染时直接算。贵再 `useMemo`（项目已开 React Compiler 则不必手写） |
| 用户点了按钮 / 提交了表单 | 事件处理函数或 Action，不要 Effect |
| 重置整棵子树状态 | 给组件换稳定业务 `key`，不要 `useEffect(() => setX(init), [id])` |
| 列表选中项 | 存 `selectedId`，渲染时 `items.find`；不要另存一份对象再 Effect 对齐 |
| 通知父组件 | 在事件里调 callback，不要渲染后 Effect 里 `onChange(value)` |
| 拉数据 | 跟项目已有框架 / 库；没有再考虑 Effect。不要每个组件手搓一套 fetch+loading |

不要为「少写依赖」去关 lint。该进依赖的就进；19.2 的 `useEffectEvent` 只用于「Effect 里触发的事件」，见 19.2 档。

## 状态与数据流

- 能算就不存。`fullName = first + ' ' + last`，不要再来一份 state。
- 表单：优先非受控 + `<form action={...}>`。受控只在必须逐键响应时用。
- 变更服务端数据用 **Action**（`startTransition` 包 async，或 `useActionState` / form `action`），不要手写 `isPending` + `try/finally` 样板。
- `use()` 读 Promise 或 Context。Promise 必须来自框架 / `cache` / 父组件已缓存的资源，**禁止**在 Client 组件 render 里 `use(fetch(...))`。
- 函数组件 `ref` 当普通 prop 接。不要再包 `forwardRef`。
- Context 直接当 provider：`<ThemeContext value={v}>`，不要 `<ThemeContext.Provider>`。
- 跨层状态：Context 或项目已有 store。不要为透传再造一层。

## 组件

- 单文件不要涨到巨型（约 500 行）；页面编排路由/入口，UI 块和 `useXxx` 拆出去。
- 列表 `key` 用稳定 id，不要 index（除非静态且永不重排）。
- 传给列表项的 props 尽量稳定：算好 `active` 再传，不要每个 item 都收会变的 `activeId`。
- 文档元数据：组件里直接写 `<title>` / `<meta>` / `<link>`，React 会提升到 `head`。项目已有 metadata 库则跟库走。
- `"use server"` 只标记 **Server Action**，不是 Server Component。没有框架的 RSC 支持就不要写。
