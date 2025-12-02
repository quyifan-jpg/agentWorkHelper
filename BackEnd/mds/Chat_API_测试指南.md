# Chat API 完整测试指南

## 📋 测试环境说明
- **API 服务地址**: http://localhost:8888
- **WebSocket 服务地址**: ws://localhost:9000/ws
- **测试用户**:
  - `root` / `123456`
  - `testuser1` / `123456`
  - `testuser2` / `123456`
- **数据库**: MongoDB (aiworkhelper数据库)
- **工具**:
  - `curl` 命令行工具 (用于HTTP API测试)
  - `wscat` WebSocket客户端 (用于WebSocket功能测试)

## 🔧 `wscat` 安装与使用

`wscat` 是一个基于Node.js的WebSocket客户端工具，非常适合用于命令行测试。

### 安装 `wscat`

确保您已安装 Node.js 和 npm，然后执行以下命令进行全局安装：
```bash
npm install -g wscat
```

### 连接命令

使用以下命令连接到WebSocket服务。请将 `{your_token}` 替换为实际的用户登录Token。

```bash
wscat -c ws://localhost:9000/ws -H "websocket:{your_token}"
```

- `-c`: 指定连接地址
- `-H`: 添加自定义请求头，我们的服务通过 `websocket` 头来传递Token进行认证

连接成功后，您将进入一个交互式终端，可以发送和接收WebSocket消息。

## 🎯 测试目标
本指南将逐步测试聊天(Chat)的所有核心功能，包括用户认证、私聊和群聊。按照本指南操作，您将学会如何：
- 使用 `wscat` 连接和测试WebSocket服务
- 理解聊天消息的JSON结构
- 验证私聊和群聊的业务逻辑是否正确

---


## 🚀 准备工作：创建测试用户

为了完整测试聊天功能，我们需要至少三个用户。系统默认只有一个 `root` 管理员。以下步骤将指导您创建 `testuser1` 和 `testuser2`。

### 0.1 获取管理员Token

首先，我们需要使用 `root` 账户登录，以获取创建新用户所需的管理员权限 Token。

**请求命令**
```bash
curl -X POST http://localhost:8888/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "name": "root",
    "password": "123456"
  }'
```

**成功响应示例 (记录 token)**
```json
{
  "code": 200,
  "data": {
    "id": "689abec2f9e967e48510fe3f",
    "name": "root",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "accessExpire": 1764767542
  },
  "msg": "success"
}
```

### 0.2 创建 `testuser1`

使用上一步获取的 `root` Token，发送以下请求来创建 `testuser1`。

**请求命令**
```bash
curl -X POST http://localhost:8888/v1/user \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {root_token}" \
  -d '{
    "name": "testuser1",
    "password": "123456",
    "status": 1
  }'
```

**成功响应示例**
```json
{
    "code": 200,
    "data": {},
    "msg": "success"
}
```

### 0.3 创建 `testuser2`

使用同一个 `root` Token，发送请求创建 `testuser2`。

**请求命令**
```bash
curl -X POST http://localhost:8888/v1/user \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {root_token}" \
  -d '{
    "name": "testuser2",
    "password": "123456",
    "status": 1
  }'
```

**成功响应示例**
```json
{
    "code": 200,
    "data": {},
    "msg": "success"
}
```

---
## 🔐 第一步：用户登录获取Token和ID

### 测试目的
分别为 `root`, `testuser1`, `testuser2` 三个用户登录，获取他们各自的JWT Token和用户ID。这些信息是后续所有WebSocket连接和API调用的基础。

### 操作说明
为每个用户执行 `curl` 登录命令，并**务必记录**下返回的 `id` 和 `token`。

### 1.1 登录 `root` 用户

**请求命令**
```bash
curl -X POST http://localhost:8888/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "name": "root",
    "password": "123456"
  }'
```

**成功响应示例 (记录 id 和 token)**
```json
{
  "code": 200,
  "data": {
    "id": "689abec2f9e967e48510fe3f",
    "name": "root",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "accessExpire": 1764767542
  },
  "msg": "success"
}
```

### 1.2 登录 `testuser1` 用户

**请求命令**
```bash
curl -X POST http://localhost:8888/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "name": "testuser1",
    "password": "123456"
  }'
```

**成功响应示例 (记录 id 和 token)**
```json
{
    "code": 200,
    "data": {
        "id": "68ac635879a48e9f5caf16b9",
        "name": "testuser1",
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "accessExpire": 1764774786
    },
    "msg": "success"
}
```

### 1.3 登录 `testuser2` 用户

**请求命令**
```bash
curl -X POST http://localhost:8888/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "name": "testuser2",
    "password": "123456"
  }'
```

**成功响应示例 (记录 id 和 token)**
```json
{
    "code": 200,
    "data": {
        "id": "68ac636779a48e9f5caf16ba",
        "name": "testuser2",
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "accessExpire": 1764774797
    },
    "msg": "success"
}
```

---

## 💬 第二步：测试私聊功能

### 测试目的
验证两个指定用户之间可以成功发送和接收私聊消息。

### 测试流程
1.  **准备两个终端**：一个代表 `testuser1`，另一个代表 `testuser2`。
2.  **分别建立连接**：在两个终端中，使用各自的Token通过 `wscat` 连接到WebSocket服务。
3.  **发送消息**：`testuser1` 向 `testuser2` 发送一条消息。
4.  **验证接收**：`testuser2` 的终端应能收到该消息。
5.  **回复消息**：`testuser2` 向 `testuser1` 回复一条消息。
6.  **验证回复**：`testuser1` 的终端应能收到回复。

### 2.1 建立WebSocket连接

**终端 1: `testuser1` 连接**

(请将 `{testuser1_token}` 替换为实际的Token)
```bash
wscat -c ws://localhost:9000/ws -H "websocket:{testuser1_token}"
```

**终端 2: `testuser2` 连接**

(请将 `{testuser2_token}` 替换为实际的Token)
```bash
wscat -c ws://localhost:9000/ws -H "websocket:{testuser2_token}"
```

### 2.2 `testuser1` 发送消息

在 **终端 1** (`testuser1`) 的 `wscat` 交互界面中，输入以下JSON内容并回车。这会向 `testuser2` (ID: `68ac636779a48e9f5caf16ba`) 发送一条消息。

```json
{
  "recvId": "68ac636779a48e9f5caf16ba",
  "chatType": 1,
  "contentType": 1,
  "content": "你好，testuser2！"
}
```

### 2.3 `testuser2` 验证接收

在 **终端 2** (`testuser2`) 的 `wscat` 交互界面中，应立即收到以下消息：

```json
{
    "conversationId": "...", // 后端生成的唯一会话ID
    "recvId": "68ac636779a48e9f5caf16ba",
    "sendId": "68ac635879a48e9f5caf16b9", // 确认是testuser1发来的
    "chatType": 1,
    "content": "你好，testuser2！",
    "contentType": 1
}
```

### 2.4 `testuser2` 回复消息

在 **终端 2** (`testuser2`) 中，输入以下JSON内容并回车，向 `testuser1` (ID: `68ac635879a48e9f5caf16b9`) 回复消息。

```json
{
  "recvId": "68ac635879a48e9f5caf16b9",
  "chatType": 1,
  "contentType": 1,
  "content": "你好，testuser1，消息已收到！"
}
```

### 2.5 `testuser1` 验证回复

在 **终端 1** (`testuser1`) 中，应能收到 `testuser2` 的回复：

```json
{
    "conversationId": "...", // 与上一条消息的会话ID相同
    "recvId": "68ac635879a48e9f5caf16b9",
    "sendId": "68ac636779a48e9f5caf16ba", // 确认是testuser2发来的
    "chatType": 1,
    "content": "你好，testuser1，消息已收到！",
    "contentType": 1
}
```

### 验证要点
- ✅ **双向通信**: 消息可以成功地在两个用户之间来回传递。
- ✅ **发送者ID**: 接收到的消息中 `sendId` 字段正确标识了发送方。
- ✅ **会话ID**: 同一对用户之间的私聊，`conversationId` 应该保持一致。

---

## 📢 第三步：测试群聊功能

### 测试目的
验证群聊消息可以被所有在线用户接收（广播模式）。

### 测试流程
1.  **准备三个终端**：分别代表 `root`, `testuser1`, `testuser2`。
2.  **全部建立连接**：在三个终端中，使用各自的Token连接到WebSocket服务。
3.  **发送群聊消息**：`root` 用户发送一条群聊消息。
4.  **验证接收**：`testuser1` 和 `testuser2` 的终端都应能收到该消息。

### 3.1 建立WebSocket连接

**终端 1: `root` 连接**
```bash
wscat -c ws://localhost:9000/ws -H "websocket:{root_token}"
```

**终端 2: `testuser1` 连接**
```bash
wscat -c ws://localhost:9000/ws -H "websocket:{testuser1_token}"
```

**终端 3: `testuser2` 连接**
```bash
wscat -c ws://localhost:9000/ws -H "websocket:{testuser2_token}"
```

### 3.2 `root` 发送群聊消息

在 **终端 1** (`root`) 的 `wscat` 交互界面中，输入以下JSON内容并回车。`chatType: 2` 表示这是一条群聊消息。

```json
{
  "chatType": 2,
  "contentType": 1,
  "content": "大家好，这是一条群聊测试消息！"
}
```

### 3.3 `testuser1` 和 `testuser2` 验证接收

在 **终端 2** (`testuser1`) 和 **终端 3** (`testuser2`) 的 `wscat` 交互界面中，都应立即收到以下消息：

```json
{
    "conversationId": "all", // 群聊的会话ID固定为 'all'
    "recvId": "",
    "sendId": "689abec2f9e967e48510fe3f", // 确认是root用户发来的
    "chatType": 2,
    "content": "大家好，这是一条群聊测试消息！",
    "contentType": 1
}
```

### 验证要点
- ✅ **广播功能**: 消息被成功广播给了除发送者外的所有在线用户。
- ✅ **发送者ID**: 接收到的消息中 `sendId` 字段正确标识了发送方 `root`。
- ✅ **会话ID**: 群聊的 `conversationId` 固定为 `all`。




