# 第3阶段结业作业

# Huayi IM - 基于 Go 的即时消息服务端

本项目是基于 Go 实现的，支持 RESTful API、Websocket 的即时通讯 App 的服务端。

## 作业目标

实现一个可独立运行的即时通信（IM）服务端，单进程同时支持 HTTP、WebSocket 协议，具备用户_一对一聊天_ 和_基于 topic_ 的群聊能力。核心链路必须使用 Go 并发特性（goroutine、channel、select-case、sync、timer、context 等）。

以下是本产品中的一些概念：

###  好友 

为简化概念，本产品_不包含_“好友”的概念，只要知道对方的用户名，任意两两用户之间都可以互相接发消息。

### 一对一单聊

一对一单聊是指两个 Users 之间互相接发消息。

_客户端发送消息通过 HTTP 协议，接收消息通过 Websocket。_

### 群聊

本产品没有“群组”的概念，群聊基于 topic。用户可以通过 HTTP 接口显式地加入 topic 群聊，也可以隐式地加入群聊，隐式加入群聊的场景有：

*   用户没有加入过某 topic，但其知道某 topic 的值，他就可以发送关于这个 topic 的消息。一旦他发送过关于这个 topic 的消息，即事实上加入了这个 topic。 
    

*   其他用户（UserB）在参与某 topic 时，提及了该用户（UserA），则该用户（UserA）事实上自动加入了这个 topic。
    

_客户端发送消息通过 HTTP 协议，接收消息通过 Websocket。_

### 离线消息

当用户建立 Websocket 连接时，如果存在与其相关的离线消息，服务端应当立即下发给他。离线消息最大保存时间为 10 分钟。

示例：UserA 给 UserB 发送若干消息，此时 UserB 不在线。当 UserB 上线后，会收到 10min 内的所有留言消息。

### 心跳

:::
Websocket 在网络协议层面已经约定了心跳协议，此处为提高业务复杂性，因此要求再实现一道业务层面的心跳协议。
:::

#### 上行心跳 Ping

在客户端与服务端的 Websocket 连接持续期间内，客户端会周期性地（每 **5s**）向服务端发送心跳消息进行保活。++当心跳间断超过 1min，服务端则认为客户端已下线，将用户标记为 online=false。++

**服务端收到消息后会进行一次应答（Ack）**。

除心跳外，客户端对服务端的所有请求、上行消息，都有类似于心跳的保活能力。也就是说即使服务端已经丢失客户端心跳一段时间了，但只要还能收到客户端的其他请求或消息，也认为客户端目前是活跃的。

#### 下行心跳 Pong

在客户端与服务端的 Websocket 连接持续期间内，服务端也会周期性地（**每 20s**）向客户端发送心跳消息，主动探测客户端是否活跃。客户端收到服务端主动探测时，会进行一次应答。++服务端连续 3 次（大约 1min）没有收到客户端应答，则认为客户端已下线，将用户标记为 online=false。++

# 协议与接口

:::
本文中接口协议和消息格式仅供参考，开发时请以 openapi.yaml 为准。

**openapi.yaml 地址：**++**https://github.com/dspo/go-homework-s3**++
:::

## HTTP

### Any /api/healthz - 健康检查

### POST /api/lgin - 登录

### POST /api/logout - 登出

### GET /api/users?online=true - 查询用户

### GET /api/topics - 查询 topics 列表

### POST /api/topics - 显式创建 topic

### DELETE /api/topics/{topic} - 删除 topic

### POST /api/topics/{topic}/actions/join - 显式加入某 topic

### POST /api/topics/{topic}/actions/quit - 显式退出某 topic

### POST /api/messages - 发送消息

## WebSocket

### 握手 GET /api/ws

### 客户端心跳格式

```json
{
  "message-id": 10,
  "from": "Harry",
  "message-type": "ping"
}
```
```json
{
  "message-type": "ack",
  "ack-id": 10
}
```

### 服务端心跳格式

```json
{
  "message-type": "pong",
  "to": "Harry",
  "message-id": 11
}
```
```json
{
  "message-type": "ack",
  "ack-id": 11
}
```

### 上行消息格式

```go
{
  "from": "Harry",
  "to": ["Ron"],
  "topic": "", # 为空表示单聊；否则表示某个话题下的群聊
  "message-type": "message"
  "content-type": "text/plain"
  "content": "Are you OK ?"
}
```

### 下行消息格式

```go
{
  "from": "Harry",
  "to": ["Ron"],
  "topic": "", # 为空表示单聊；否则表示某个话题下的群聊
  "message-type": "message"
  "content-type": "text/plain"
  "content": "Are you OK ?"
}
```

## 功能与技术要求

*   功能完整性：
    

完整实现文本所约定的接口和功能，并能正确工作。要正确支持单聊、群聊、心跳、在线下线管理。

*   技术运用：
    

开发者必须积极运用第 3  阶段培训的 Go 并发编程相关技术，如 Context、channel、sync 包等。但开发者可以自由决定在何时 、使用哪些技术。

可以使用第三方库实现 Websocket 协议部分。

可以使用数据库进行持久化，也可以不使用。验收时只会在应用程序的一个生命周期中进行测试，因此不要求数据持久化。

*   性能要求
    
    *   以 1C 1G 机器为基准，至少要支持 100 Users 在线互聊和群聊。
        

*   禁止
    
    禁止单线程事件循环实现消息分发；fan-out 必须依赖并发安全数据结构和 goroutine 派发。
    
    不得直接使用社区中成熟的消息治理库，核心的消息治理功能应当自行实现。
    
    开发时应当了解内存是有限的，不可以无限增长。
    
*   讨论与辅助
    

鼓励学员间积极讨论，但不得交换源码。可以使用 AI 辅助编程，但不得使用 AI solo 整个项目。

*   进阶
    

不必实现限流/熔断/重试/延迟/重传等复杂能力。能成功实现视为加分项。

*   加分项
    
    *   充分的 e2e 测试（包含 Websocket）。
        
    *   压力测试或性能测试。
        

## 交付内容

*   `README.md`：架构、协议、配置项、并发设计、运行方法。
    
*   源码。
    
*   测试：至少提供 HTTP 接口的集成测试，鼓励提供 Websocket 集成测试。
    
*   请提供尽可能简单的脚本或命令让验收者方便验收，如 `make e2e`。

---

## 参考实现（服务端）

* HTTP 框架：`gin`
* 配置：`viper`（读取 `config.yaml` 与环境变量）
* 依赖注入/生命周期：`fx`
* WebSocket：`nhooyr.io/websocket`
* 持久化：无（纯内存，进程重启即丢）
* 认证：登录返回 `sid`；HTTP 使用 `Authorization: Bearer <sid>`；WebSocket 使用 `/api/ws?sid=...`

## 前端客户端（im-app）使用指南

本仓库包含一个纯前端的聊天客户端 `im-app/`（React + TSX + TailwindCSS + shadcn/ui），用于**连接和调试**本 IM 服务端。

### 1) 安装前置环境

* 安装 Node.js（建议 LTS 版本）
* 安装 `pnpm`（如果未安装）：

```bash
npm i -g pnpm
```

### 2) 安装前端依赖

在仓库根目录执行：

```bash
cd im-app
pnpm install
```

### 3) 启动服务端 + 启动前端开发服务

先启动服务端（仓库根目录）：

```bash
make run
```

再启动前端（`im-app/` 目录）：

```bash
pnpm dev
```

浏览器打开终端输出的地址（默认 `http://localhost:5173`）。

### 4) 多开客户端模拟多用户互聊

你可以多开多个前端实例（不同端口）来模拟多用户在线：

```bash
pnpm dev --port 5174
pnpm dev --port 5175
```

分别在不同端口页面用不同 username 登录，即可互相聊天。

### 5) 用 im-app 调试服务端（推荐检查点）

* **代理与跨域**：开发模式下，前端会把 `/api/*`（包含 WebSocket）代理到 `http://localhost:8080`（见 `im-app/vite.config.ts`），因此不需要服务端额外配置 CORS。
* **鉴权是否生效**：登录成功后，服务端会返回 `sid`；前端会把 `sid` 存在浏览器 `localStorage` 并在后续请求中带上 `Authorization: Bearer <sid>`。
* **WebSocket 是否在线**：打开浏览器 DevTools → Network → WS，确认 `GET /api/ws?sid=...` 为 `101 Switching Protocols`，并且连接持续存在。
* **心跳是否正常**：在 WS 的 Frames 中能看到客户端周期性 `ping`，服务端回 `ack`；以及服务端周期性 `pong`，客户端回 `ack`。
* **离线消息**：关闭/刷新某个用户页面使其离线，另一用户给他发消息；再重新打开该用户页面，连接建立后应立即收到 10 分钟内的离线消息。

### 运行

```bash
make run
```

默认监听 `:8080`，可通过 `config.yaml` 或环境变量覆盖（示例见 `config.example.yaml`）。

### 测试

```bash
make e2e
```

### 并发设计概览

* **消息分发**：使用带缓冲队列 + worker pool 的 `Dispatcher` 并发投递到各连接的发送队列；慢消费者会被断开以避免内存无限增长。
* **连接读写**：每个 WebSocket 连接独立的读 goroutine + 写 goroutine + 心跳 goroutine（服务端 `pong`）+ 空闲检测 goroutine。
* **在线状态**：基于连接存在与活跃时间；任何 HTTP 请求/WS 上行消息都会 `Touch` 活跃时间。
* **离线消息**：按用户保存 10 分钟 TTL 的离线队列；建立 WS 连接时立即下发并清空；后台定时清理过期消息。
