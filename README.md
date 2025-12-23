# 第3阶段结业作业（编辑中）

# Huayi IM - 基于 Go 的即时消息服务端

本项目是基于 Go 实现的，支持 RESTful API、Websocket 的即时通讯 App 的服务端。

## 作业目标

实现一个可独立运行的即时通信（IM）服务端，单进程同时支持 HTTP、WebSocket 协议，具备用户一对一聊天和基于 topic 的群聊能力。核心链路必须大量且显式使用 Go 并发特性（goroutine、channel、select-case、sync、timer、context 等）。

## 产品功能

本产品的主要功能是为用户提供一对一单聊和基于话题的群聊功能。

以下是本产品中的一些概念：

###  好友 

为简化概念，本产品不包含“好友”的概念，只要知道对方的用户名，任意两两用户之间都可以互相接发消息。

### 一对一单聊

一对一单聊是指两个 Users 之间互相接发消息。

客户端发送消息通过 HTTP 协议，接收消息通过 Websocket。

### 群聊

本产品没有“群组”的概念，群聊基于 topic。用户可以通过 HTTP 接口显式地加入 topic 群聊，也可以隐式地加入群聊，隐式加入群聊的场景有：

*   用户没有加入过某 topic，但其知道某 topic 的值，他就可以发送关于这个 topic 的消息。一旦他发送过关于这个 topic 的消息，即事实上加入了这个 topic。 
    

*   其他用户（UserB）在参与某 topic 时，提及了该用户（UserA），则该用户（UserA）事实上自动加入了这个 topic。
    

客户端发送消息通过 HTTP 协议，接收消息通过 Websocket。

### 离线消息

当用户登录时（建立 Websocket 连接时），如果存在与其相关的离线消息，服务端应当立即下发给他。离线消息最大保存时间为 10 分钟。

示例：UserA 给 UserB 发送若干消息，此时 UserB 不在线。当 UserB 上线后，会收到 10min 内的所有留言消息。

### 心跳

:::
注意：Websocket 在网络协议层面已经约定了心跳协议，此处为提高业务复杂性，因此要求再实现一道业务层面的心跳协议。
:::

#### 上行心跳

客户端登录后，会与服务端建立 Websocket 连接。在连接期间，会持续性地**每 5s** 向服务端发送心跳消息。++当心跳间断超过 1min，服务端则认为客户端已下线，将用户标记为 online=false。++

**服务端收到消息后会进行一次应答**。

除 Websocket 中自动心跳外，客户端对服务端的所有请求、上行消息，都有类似于心跳的保活能力。也就是说即使服务端已经丢失客户端心跳一段时间了，但是还能收到客户端的其他请求或消息，也认为客户端目前是在线的。

#### 下行心跳

客户端与服务端建立 Websocket 连接后，在连接期间，服务端也会持续性地**每 20s** 向客户端发送心跳消息，主动探测客户端是否活跃。客户端收到服务端主动探测时，会进行一次应答。++服务端连续 3 次没有收到客户端应答，则认为客户端已下线，将用户标记为 online=false。++

# 协议与接口

## HTTP

### Any /healthz - 健康检查

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

**具体消息格式以 openapi.yaml 为准**

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

可以使用第三方库开发 Websocket 部分。

可以使用数据库进行持久化，也可以不使用。验收时只会在应用程序的一个生命周期中进行测试，因此不要求数据持久化。

*   性能要求
    
    *   以 1c 1G 机器为基准，至少要支持 100 Users 在线互聊和群聊。
        

*   禁止
    
    禁止单线程事件循环实现消息分发；fan-out 必须依赖并发安全数据结构和 goroutine 派发。
    
    社区中可能存在成熟的消息治理库，不得直接使用这样的库。核心的消息转发功能应当自行实现。
    
    写缓冲必须是有界 channel，满时 select 分支处理背压；不得用无界队列。
    
*   讨论与辅助
    

鼓励学员间积极讨论，但不得交换源码。可以使用 AI 辅助编程，但不得使用 AI solo 整个项目。

*   进阶
    

限流/熔断/重试/延迟/重传等复杂能力可以不做。

*   加分项
    
    *   充分的 e2e 测试（包含 Websocket）。
        
    *   压力测试或性能测试。
        

## 交付内容

*   `README.md`：架构、协议、配置项、并发设计、运行方法。
    
*   源码。
    
*   测试：至少提供 HTTP 接口的集成测试，固定提供 Websocket 集成测试。请提供一键脚本（如 `make e2e`）。