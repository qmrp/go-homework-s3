# Huayi IM Web Client (React)

一个纯前端聊天客户端（React + TSX），用于连接本仓库的 Go 服务端：

- 登录/登出（sid 会话）
- 单聊/群聊（topic）
- WebSocket 心跳：客户端每 5s `ping`；收到服务端 `pong` 立即 `ack`
- 非激活会话也能收消息；新用户/topic 自动出现在侧边栏；按最近消息排序
- 样式仅使用 TailwindCSS + shadcn/ui
- 使用 `sid`（登录响应体）作为 `Authorization: Bearer <sid>`，并用 WS `?sid=` 建立连接（避免 Cookie 跨端口共享导致多开串号）

## 运行

1) 启动 Go 服务端（仓库根目录）：

```bash
make run
```

2) 启动前端（本目录）：

```bash
pnpm dev
```

打开 `http://localhost:5173`。

前端开发服务器已在 `im-app/vite.config.ts` 配置好代理：`/api` 会转发到 `http://localhost:8080`（包含 WebSocket），因此无需 CORS。

## 多开客户端

```bash
pnpm dev --port 5174
pnpm dev --port 5175
```
