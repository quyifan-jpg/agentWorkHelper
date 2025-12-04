# WebSocket 连接问题排查指南

## 🔍 常见连接失败原因

### 1. 服务未启动

**症状**：前端显示 "WebSocket 连接失败"

**检查方法**：

```bash
# 检查 WebSocket 服务是否在运行
lsof -i :9000
# 或
netstat -an | grep 9000
```

**解决方法**：

```bash
cd BackEnd
go run cmd/api/main.go -f etc/backend.yaml
```

应该看到：

```
启动 WebSocket 服务: ws://127.0.0.1:9000/ws
```

---

### 2. Token 无效或过期

**症状**：连接立即失败，控制台显示认证错误

**检查方法**：

- 查看后端日志：`WebSocket 认证失败`
- 检查前端 Token 是否有效

**解决方法**：

1. 重新登录获取新的 Token
2. 检查 Token 是否过期（默认 24 小时）
3. 确认 Token 格式正确

**测试 Token**：

```bash
# 使用 curl 测试 Token
curl -X POST http://localhost:8889/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"name": "root", "password": "123456"}'
```

---

### 3. 端口配置错误

**症状**：连接超时或连接被拒绝

**检查方法**：

1. 检查 `etc/backend.yaml` 中的 WebSocket 配置：

   ```yaml
   WS:
     Host: "127.0.0.1"
     Port: 9000
   ```

2. 检查前端环境变量 `VITE_WS_BASE_URL`：
   ```bash
   # 前端 .env 文件
   VITE_WS_BASE_URL=ws://127.0.0.1:9000
   ```

**解决方法**：

- 确保后端配置的端口与前端一致
- 确保端口未被占用

---

### 4. CORS 或跨域问题

**症状**：浏览器控制台显示 CORS 错误

**当前实现**：

- WebSocket 的 `CheckOrigin` 设置为 `true`（允许所有来源）
- 如果仍有问题，检查浏览器控制台

---

### 5. Token 传递方式问题

**前端传递方式**：

```typescript
// 通过 URL 参数传递（浏览器 WebSocket API 限制）
const wsUrl = `ws://127.0.0.1:9000/ws?token=${token}`;
```

**后端接收方式**：

1. 优先从请求头获取：`websocket: <token>`
2. 如果请求头没有，从 URL 参数获取：`?token=<token>`

**检查方法**：

- 查看浏览器 Network 标签，检查 WebSocket 连接的 URL
- 查看后端日志，确认 Token 来源

---

## 🐛 调试步骤

### 步骤 1：检查服务状态

```bash
# 检查 HTTP API 服务
curl http://localhost:8889/ping

# 检查 WebSocket 服务（需要 Token）
# 使用 wscat 测试
wscat -c "ws://127.0.0.1:9000/ws?token=<your_token>"
```

### 步骤 2：查看后端日志

启动服务后，查看日志输出：

```
启动 WebSocket 服务: ws://127.0.0.1:9000/ws
WebSocket 认证成功，开始升级连接 userID=1
WebSocket 连接升级成功 userID=1
开始处理 WebSocket 连接 userID=1
```

如果看到错误：

- `WebSocket 认证失败` → Token 问题
- `WebSocket 升级失败` → 连接问题
- `WebSocket 服务启动失败` → 端口被占用

### 步骤 3：检查前端连接

打开浏览器开发者工具：

1. **Console 标签**：查看连接日志

   ```
   正在连接WebSocket: ws://127.0.0.1:9000/ws?token=***
   WebSocket连接成功
   ```

2. **Network 标签**：查看 WebSocket 连接

   - 找到 `ws` 类型的连接
   - 检查状态码（101 = 成功，401 = 认证失败）

3. **检查错误消息**：
   - `WebSocket连接失败` → 服务未启动或网络问题
   - `Unauthorized` → Token 无效

---

## 🔧 常见修复方法

### 修复 1：Token 格式问题

**问题**：Token 可能包含特殊字符，需要 URL 编码

**解决**：前端已自动编码

```typescript
const wsUrl = `${this.url}?token=${encodeURIComponent(this.token)}`;
```

### 修复 2：端口冲突

**问题**：9000 端口被占用

**解决**：

```bash
# 查找占用端口的进程
lsof -i :9000
# 或修改配置文件中的端口
```

### 修复 3：服务启动顺序

**问题**：WebSocket 服务启动失败导致 HTTP 服务也失败

**解决**：检查 `main.go` 中的错误处理，确保一个服务失败不影响另一个

---

## 📊 日志级别

后端使用 `zerolog` 记录日志，日志级别：

- **Info**：正常连接、消息处理
- **Error**：认证失败、连接错误
- **Warn**：未知消息类型
- **Debug**：详细消息内容（需要设置日志级别）

---

## ✅ 验证连接成功

连接成功后，应该看到：

**后端日志**：

```
启动 WebSocket 服务: ws://127.0.0.1:9000/ws
WebSocket 认证成功，开始升级连接 userID=1
WebSocket 连接升级成功 userID=1
WebSocket 连接已建立 uid=1
开始处理 WebSocket 连接 userID=1
```

**前端控制台**：

```
正在连接WebSocket: ws://127.0.0.1:9000/ws?token=***
WebSocket连接成功
[WebSocket] 添加消息处理器，当前共1个
```

---

## 🚨 紧急排查清单

- [ ] WebSocket 服务是否启动？
- [ ] 端口 9000 是否可访问？
- [ ] Token 是否有效且未过期？
- [ ] 前端环境变量 `VITE_WS_BASE_URL` 是否正确？
- [ ] 浏览器控制台是否有错误信息？
- [ ] 后端日志是否有错误信息？
- [ ] 防火墙是否阻止了连接？

---

## 📝 测试命令

### 使用 wscat 测试

```bash
# 1. 安装 wscat
npm install -g wscat

# 2. 获取 Token（先登录）
TOKEN=$(curl -s -X POST http://localhost:8889/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"name": "root", "password": "123456"}' \
  | jq -r '.data.token')

# 3. 连接 WebSocket
wscat -c "ws://127.0.0.1:9000/ws?token=$TOKEN"

# 4. 发送测试消息
{"recvId": "2", "chatType": 2, "content": "测试", "contentType": 1}
```

---

## 💡 提示

1. **开发环境**：确保前后端服务都在运行
2. **生产环境**：检查防火墙和反向代理配置
3. **Token 管理**：Token 过期后需要重新登录
4. **连接重试**：前端会自动重试连接（最多 5 次）
