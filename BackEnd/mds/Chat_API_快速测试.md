# Chat API 快速测试指南

## 🚀 快速开始

### 1. 启动服务

确保 BackEnd 服务正在运行：

```bash
cd BackEnd
go run cmd/api/main.go -f etc/backend.yaml
```

服务应该运行在 `http://localhost:8889`

### 2. 运行测试脚本

```bash
cd BackEnd
bash scripts/test_chat.sh
```

## 📋 测试内容

测试脚本会自动执行以下步骤：

1. **登录用户 1 (root)** - 获取 Token 和用户 ID
2. **登录用户 2 (testuser1)** - 如果不存在会自动创建
3. **发送私聊消息** - 用户 1 向用户 2 发送私聊消息
4. **回复私聊消息** - 用户 2 回复用户 1
5. **查询私聊记录** - 查询会话历史（如果已实现）
6. **发送群聊消息** - 用户 1 发送群聊消息
7. **查询群聊记录** - 查询群聊历史（如果已实现）

## ✅ 预期结果

### 私聊测试

- ✅ 消息发送成功
- ✅ 会话 ID 自动生成（22 位字符串）
- ✅ 回复消息使用相同的会话 ID
- ✅ 会话 ID 一致性验证通过

### 群聊测试

- ✅ 群聊消息发送成功
- ✅ 会话 ID 为 `"all"`
- ✅ 消息保存到数据库

## 🔍 手动测试

如果脚本测试有问题，可以手动测试：

### 1. 登录获取 Token

```bash
curl -X POST http://localhost:8889/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "name": "root",
    "password": "123456"
  }'
```

记录返回的 `token` 和 `id`

### 2. 发送私聊消息

```bash
curl -X POST http://localhost:8889/v1/chat/message \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "recvId": "{target_user_id}",
    "chatType": 2,
    "content": "你好，这是一条私聊消息",
    "contentType": 1
  }'
```

### 3. 发送群聊消息

```bash
curl -X POST http://localhost:8889/v1/chat/message \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "chatType": 1,
    "content": "大家好，这是一条群聊消息",
    "contentType": 1
  }'
```

## 🐛 常见问题

### 1. 服务未启动

```
curl: (7) Failed to connect to localhost port 8889
```

**解决**：确保服务正在运行

### 2. Token 过期

```
{"code":401,"msg":"unauthorized"}
```

**解决**：重新登录获取新的 Token

### 3. 用户不存在

```
{"code":400,"msg":"user not found"}
```

**解决**：使用 root 用户创建测试用户，或脚本会自动创建

### 4. 数据库连接失败

检查 `etc/backend.yaml` 中的数据库配置

## 📊 验证数据库

可以连接数据库查看聊天记录：

```sql
-- 查看所有聊天记录
SELECT * FROM chat_logs ORDER BY send_time DESC LIMIT 10;

-- 查看私聊记录
SELECT * FROM chat_logs WHERE chat_type = 2 ORDER BY send_time DESC;

-- 查看群聊记录
SELECT * FROM chat_logs WHERE chat_type = 1 ORDER BY send_time DESC;

-- 查看特定会话的记录
SELECT * FROM chat_logs WHERE conversation_id = '{conversation_id}' ORDER BY send_time ASC;
```

## 🎯 下一步

测试通过后，可以：

1. 实现 `ListMessages` 查询功能
2. 添加 WebSocket 实时推送
3. 实现消息已读状态
4. 添加文件上传功能
