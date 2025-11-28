# AI Chat API 测试指南

## 📋 测试环境说明
- **API 服务地址**: http://localhost:8888
- **AI模型**: 通义千问 (qwen3-max-preview)
- **API端点**: https://dashscope.aliyuncs.com/compatible-mode/v1
- **测试用户**: `root` / `123456`
- **数据库**: MongoDB (aiworkhelper数据库)
- **工具**: `curl` 命令行工具

## 🎯 测试目标
本指南将测试AI通用聊天功能的核心能力，包括：
- 用户认证和Token获取
- AI聊天API调用
- 通义千问模型响应验证
- 不同类型聊天请求的处理

---

## 🚀 第一步：用户登录获取Token

### 测试目的
获取有效的JWT Token，用于后续AI聊天API的认证。

### 1.1 登录获取Token

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
    "accessExpire": 1766910832
  },
  "msg": "success"
}
```

**验证要点**
- ✅ 返回状态码为 200
- ✅ 响应包含有效的 `token` 字段
- ✅ `accessExpire` 时间戳大于当前时间

---

## 🤖 第二步：测试AI基础聊天功能

### 测试目的
验证AI聊天API能够正确调用通义千问模型并返回合理的响应。

### 2.1 基础问候测试

**请求命令**
```bash
curl -X POST http://localhost:8888/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "prompts": "你好，你是谁",
    "chatType": 0,
    "relationId": 0
  }'
```

**成功响应示例**
```json
{
  "code": 200,
  "data": {
    "data": "你好！我是一个全能型人工智能助手，能够帮助你解答各种问题、提供信息查询、进行创作辅助、技术支持、语言翻译等服务。无论是学习、工作还是生活中的疑问，我都会尽力为你提供清晰、准确、实用的解决方案。随时告诉我你的需求，我会全力以赴为你服务！"
  },
  "msg": "success"
}
```

**验证要点**
- ✅ 返回状态码为 200
- ✅ 响应包含AI生成的文本内容
- ✅ 内容符合通义千问的回答风格
- ✅ 响应时间在合理范围内（通常 < 10秒）

### 2.2 技术问题测试

**请求命令**
```bash
curl -X POST http://localhost:8888/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "prompts": "请解释一下什么是RESTful API",
    "chatType": 0,
    "relationId": 0
  }'
```

**验证要点**
- ✅ AI能够理解技术术语
- ✅ 提供准确的技术解释
- ✅ 回答结构清晰、逻辑性强

### 2.3 创意写作测试

**请求命令**
```bash
curl -X POST http://localhost:8888/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "prompts": "写一首关于春天的短诗",
    "chatType": 0,
    "relationId": 0
  }'
```

**验证要点**
- ✅ AI能够进行创意写作
- ✅ 生成的内容符合要求格式
- ✅ 内容具有创意性和文学性

---

## 🔧 第三步：测试不同聊天类型

### 测试目的
验证不同 `chatType` 参数的处理能力。

### 3.1 测试 chatType = 1

**请求命令**
```bash
curl -X POST http://localhost:8888/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "prompts": "帮我分析一下Go语言的优势",
    "chatType": 1,
    "relationId": 123
  }'
```

### 3.2 测试 chatType = 2

**请求命令**
```bash
curl -X POST http://localhost:8888/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "prompts": "请推荐一些学习编程的资源",
    "chatType": 2,
    "relationId": 456
  }'
```

**验证要点**
- ✅ 不同 `chatType` 都能正常处理
- ✅ `relationId` 参数被正确传递
- ✅ 响应格式保持一致

---

## ⚠️ 第四步：错误场景测试

### 测试目的
验证API在异常情况下的错误处理能力。

### 4.1 无效Token测试

**请求命令**
```bash
curl -X POST http://localhost:8888/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer invalid_token" \
  -d '{
    "prompts": "测试无效token",
    "chatType": 0,
    "relationId": 0
  }'
```

**预期响应**
```json
{
  "code": 401,
  "msg": "Unauthorized"
}
```

### 4.2 缺少必要参数测试

**请求命令**
```bash
curl -X POST http://localhost:8888/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "chatType": 0,
    "relationId": 0
  }'
```

**验证要点**
- ✅ 返回适当的错误状态码
- ✅ 错误信息清晰明确
- ✅ 不会导致服务崩溃

### 4.3 空内容测试

**请求命令**
```bash
curl -X POST http://localhost:8888/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "prompts": "",
    "chatType": 0,
    "relationId": 0
  }'
```

---

## 📊 第五步：性能和稳定性测试

### 测试目的
验证AI聊天功能的性能表现和稳定性。

### 5.1 长文本处理测试

**请求命令**
```bash
curl -X POST http://localhost:8888/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "prompts": "请详细解释人工智能的发展历史，包括从图灵测试到现代深度学习的演进过程，以及各个重要里程碑事件的意义和影响",
    "chatType": 0,
    "relationId": 0
  }'
```

### 5.2 连续请求测试

连续发送多个请求，验证系统的并发处理能力：

```bash
# 请求1
curl -X POST http://localhost:8888/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{"prompts": "1+1等于几", "chatType": 0, "relationId": 0}'

# 请求2  
curl -X POST http://localhost:8888/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{"prompts": "今天天气怎么样", "chatType": 0, "relationId": 0}'
```

**验证要点**
- ✅ 响应时间稳定
- ✅ 内存使用正常
- ✅ 无内存泄漏
- ✅ 并发请求处理正确

---

## ✅ 测试检查清单

### 功能测试
- [ ] 用户登录获取Token成功
- [ ] 基础AI聊天功能正常
- [ ] 技术问题回答准确
- [ ] 创意写作能力正常
- [ ] 不同chatType处理正确

### 错误处理测试
- [ ] 无效Token正确拒绝
- [ ] 缺少参数错误处理
- [ ] 空内容请求处理

### 性能测试
- [ ] 长文本处理正常
- [ ] 连续请求稳定
- [ ] 响应时间合理
- [ ] 系统资源使用正常

### 通义千问特性验证
- [ ] 模型响应风格符合通义千问
- [ ] 中文处理能力优秀
- [ ] 专业知识回答准确
- [ ] 创意内容生成质量高

---

## 🔍 故障排除

### 常见问题

1. **连接超时**
   - 检查网络连接
   - 验证通义千问API密钥是否有效
   - 确认API配额是否充足

2. **认证失败**
   - 检查Token是否过期
   - 验证Authorization头格式
   - 确认用户权限

3. **响应异常**
   - 检查服务日志
   - 验证模型配置
   - 确认API端点正确

### 日志查看
```bash
# 查看服务运行日志
tail -f /path/to/service.log

# 检查错误日志
grep "ERROR" /path/to/service.log
```

---

## 📝 测试报告模板

### 测试环境
- 测试时间：[填写时间]
- 测试人员：[填写姓名]
- 服务版本：[填写版本]
- 模型版本：qwen3-max-preview

### 测试结果
- 功能测试：✅/❌
- 错误处理：✅/❌  
- 性能测试：✅/❌
- 整体评估：✅/❌

### 发现问题
[记录测试中发现的问题]

### 改进建议
[提出优化建议]
