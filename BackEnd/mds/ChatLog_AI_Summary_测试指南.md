# AI群消息总结功能 完整测试指南

## 📋 测试环境说明
- **服务地址**: http://localhost:8888
- **测试用户**: root / 123456
- **数据库**: MongoDB (aiworkhelper数据库, chat_log集合)
- **工具**: curl 命令行工具
- **AI模型**: 阿里云DashScope

## 🎯 测试目标
本指南将逐步测试AI群消息总结功能的所有场景，验证功能是否正常工作。按照本指南操作,您将学会如何:
- 创建测试用户并发送群聊消息
- 使用AI总结群聊消息内容
- 理解总结结果的数据格式
- 配置时间范围筛选消息
- 识别总结的任务类型(待办任务、审批事项)
- 验证功能与原有功能的兼容性

---

## 🎬 准备工作：创建测试用户和群聊数据

在开始测试AI群消息总结功能之前,我们需要先准备测试环境:创建几个测试用户,并让他们发送一些群聊消息。这些消息将作为AI总结的原始数据。

### 步骤0.1: 以root用户登录

首先使用root账号登录获取管理员token:

```bash
curl -X POST http://localhost:8888/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "name": "root",
    "password": "123456"
  }'
```

**成功响应示例**:
```json
{
  "code": 200,
  "data": {
    "id": "68f492c0f989561ede9321c9",
    "name": "root",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "accessExpire": 1769656941
  },
  "msg": "success"
}
```

**⚠️ 重要**: 复制返回的token,保存为环境变量,后续步骤都会用到:
```bash
export ROOT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

### 步骤0.2: 创建测试用户

使用root账号创建几个测试用户,用于模拟真实的群聊场景。

#### 创建用户1: 张经理
```bash
curl -X POST http://localhost:8888/v1/user \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ROOT_TOKEN" \
  -d '{
    "name": "张经理",
    "password": "123456"
  }'
```

#### 创建用户2: 李员工
```bash
curl -X POST http://localhost:8888/v1/user \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ROOT_TOKEN" \
  -d '{
    "name": "李员工",
    "password": "123456"
  }'
```

#### 创建用户3: 王员工
```bash
curl -X POST http://localhost:8888/v1/user \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ROOT_TOKEN" \
  -d '{
    "name": "王员工",
    "password": "123456"
  }'
```

#### 创建用户4: 赵员工
```bash
curl -X POST http://localhost:8888/v1/user \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ROOT_TOKEN" \
  -d '{
    "name": "赵员工",
    "password": "123456"
  }'
```

**预期响应**:
```json
{
  "code": 200,
  "data": {},
  "msg": "success"
}
```

---

### 步骤0.3: 使用WebSocket发送群聊消息

群聊消息需要通过WebSocket连接发送。我们将使用一个简单的WebSocket客户端脚本来模拟多个用户发送群聊消息。

#### 创建WebSocket测试脚本

创建文件 `send_group_messages.html`:

```html
<!DOCTYPE html>
<html>
<head>
    <title>群聊消息测试工具</title>
    <meta charset="UTF-8">
</head>
<body>
    <h2>群聊消息测试工具</h2>
    <div>
        <label>用户Token: </label>
        <input type="text" id="token" size="80" placeholder="粘贴登录后的token">
        <button onclick="connect()">连接</button>
        <button onclick="disconnect()">断开</button>
    </div>
    <div>
        <label>消息内容: </label>
        <input type="text" id="message" size="80" placeholder="输入消息内容">
        <button onclick="sendMessage()">发送群消息</button>
    </div>
    <div>
        <h3>连接状态: <span id="status">未连接</span></h3>
        <h3>消息日志:</h3>
        <pre id="log" style="border:1px solid #ccc; padding:10px; height:300px; overflow-y:auto;"></pre>
    </div>

    <script>
        let ws = null;

        function connect() {
            const token = document.getElementById('token').value;
            if (!token) {
                alert('请先输入token');
                return;
            }

            ws = new WebSocket('ws://localhost:9000/ws');

            ws.onopen = function() {
                document.getElementById('status').textContent = '已连接';
                log('WebSocket连接成功');

                // 发送认证消息(带token的header)
                ws.send(JSON.stringify({
                    type: 'auth',
                    token: token
                }));
            };

            ws.onmessage = function(event) {
                log('收到消息: ' + event.data);
            };

            ws.onerror = function(error) {
                log('错误: ' + JSON.stringify(error));
            };

            ws.onclose = function() {
                document.getElementById('status').textContent = '已断开';
                log('WebSocket连接关闭');
            };
        }

        function disconnect() {
            if (ws) {
                ws.close();
                ws = null;
            }
        }

        function sendMessage() {
            if (!ws || ws.readyState !== WebSocket.OPEN) {
                alert('请先连接WebSocket');
                return;
            }

            const message = document.getElementById('message').value;
            if (!message) {
                alert('请输入消息内容');
                return;
            }

            // 发送群聊消息
            const msg = {
                chatType: 1,        // 1=群聊
                content: message,
                contentType: 1      // 1=文本消息
            };

            ws.send(JSON.stringify(msg));
            log('发送群消息: ' + message);
            document.getElementById('message').value = '';
        }

        function log(message) {
            const logDiv = document.getElementById('log');
            const time = new Date().toLocaleTimeString();
            logDiv.textContent += `[${time}] ${message}\n`;
            logDiv.scrollTop = logDiv.scrollHeight;
        }
    </script>
</body>
</html>
```

#### 使用测试脚本发送群聊消息

1. **在浏览器中打开** `send_group_messages.html`
2. **分别用不同用户登录并发送消息**:

**张经理的消息**:
- 先用张经理账号登录获取token,粘贴到工具中
- 连接WebSocket
- 发送: `大家好，下周一我们需要完成新功能的开发，李员工负责前端，王员工负责后端`

**李员工的消息**:
- 用李员工账号登录获取token,粘贴到工具中
- 连接WebSocket
- 发送: `收到！前端界面设计我这边周三前完成`
- 发送: `经理，我下周二需要请假一天，去医院复查`
- 发送: `谢谢王员工，应该没问题，我会尽快完成的`

**王员工的消息**:
- 用王员工账号登录获取token
- 连接WebSocket
- 发送: `好的，后端接口开发我预计周四完成`
- 发送: `李员工，你那边如果来不及，我可以协助前端开发`

**张经理的回复**:
- 发送: `很好，赵员工你负责测试工作`
- 发送: `可以，注意身体，工作进度调整一下就好`

**赵员工的消息**:
- 用赵员工账号登录获取token
- 连接WebSocket
- 发送: `明白，我会在周五进行全面测试`

**root用户的消息**:
- 用root账号token连接WebSocket
- 发送: `大家好，我们需要讨论一下下周的项目进度`
- 发送: `我这边的开发任务基本完成了，还需要进行测试`
- 发送: `测试这边我来负责，预计需要2天时间`
- 发送: `好的，那我们周三开会讨论测试结果`
- 发送: `我需要请假一天，周四有事情要处理`
- 发送: `没问题，记得提交请假申请`
- 发送: `已经提交了，谢谢提醒`

---

### 步骤0.4: 验证群聊数据已保存

发送完群聊消息后,数据会自动保存到MongoDB的`chat_log`集合中。可以通过MongoDB客户端验证:

```bash
# 使用mongo shell查询(如果有MongoDB命令行工具)
mongo aiworkhelper --eval 'db.chat_log.find({conversationId:"all"}).count()'

# 或使用mongosh
mongosh aiworkhelper --eval 'db.chat_log.find({conversationId:"all"}).count()'
```

**预期结果**: 应该返回大于0的数字,表示群聊消息已成功保存。

---

### 📝 准备工作总结

现在你已经完成了测试准备工作:
- ✅ 创建了5个测试用户(root + 4个员工)
- ✅ 发送了约15条群聊消息,包含:
  - 任务分配(张经理分配开发任务)
  - 请假申请(李员工和root的请假)
  - 工作进度汇报
  - 会议安排
- ✅ 数据已保存到chat_log集合

接下来就可以开始测试AI群消息总结功能了!

---

## 🔐 第一步：用户登录获取Token

### 测试目的
获取JWT认证token,用于后续所有API调用的身份验证。

### 请求命令
```bash
curl -X POST http://localhost:8888/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "name": "root",
    "password": "123456"
  }'
```

### 成功响应示例
```json
{
  "code": 200,
  "data": {
    "id": "68f492c0f989561ede9321c9",
    "name": "root",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "accessExpire": 1769656941
  },
  "msg": "success"
}
```

### 重要说明
- ✅ **成功标志**: code=200,返回token字段
- 📝 **记录token**: 复制token值,后续所有请求都需要使用
- ⏰ **token有效期**: accessExpire字段表示过期时间戳

---

## 📊 第二步：测试基本群消息总结 - 默认时间范围

### 测试目的
使用默认时间范围(最近24小时)总结群聊消息,验证基本总结功能。

### 请求命令
```bash
curl -X POST http://localhost:8888/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "prompts": "请帮我总结一下最近的群聊消息",
    "relationId": "all"
  }'
```

### 成功响应示例
```json
{
  "code": 200,
  "data": {
    "chatType": 4,
    "data": [
      {
        "Type": 1,
        "Title": "新功能开发任务分配",
        "Content": "张经理分配下周一完成新功能开发任务：李员工负责前端,预计周三前完成界面设计；王员工负责后端,预计周四完成接口开发；赵员工负责测试,计划周五进行全面测试。"
      },
      {
        "Type": 2,
        "Title": "李员工请假申请",
        "Content": "李员工申请下周二请假一天去医院复查,张经理已批准,并提醒调整工作进度。王员工主动提出可协助前端开发,李员工表示感谢并确认能按时完成。"
      },
      {
        "Type": 1,
        "Title": "项目测试与会议安排",
        "Content": "root负责项目测试工作,预计需2天时间,并安排周三开会讨论测试结果。"
      },
      {
        "Type": 2,
        "Title": "root请假申请",
        "Content": "root申请周四请假一天处理私事,已提交请假申请并获得确认。"
      }
    ]
  },
  "msg": "success"
}
```

### 验证要点
- ✅ **成功标志**: code=200, chatType=4 (表示聊天日志总结类型)
- 📊 **数据结构**: 返回数组,每个元素包含Type、Title、Content三个字段
- 🔍 **总结质量**: AI能正确识别任务分配、请假申请等事项
- 👥 **人员信息**: 总结中包含完整的人员信息

---

## 🕐 第三步：测试指定时间范围的消息总结

### 测试目的
使用自定义时间范围筛选消息并总结,验证时间过滤功能。

### 获取时间戳
```bash
# 获取当前时间戳(秒)
date +%s

# 获取3天前的时间戳
date -v-3d +%s

# 或使用在线工具: https://tool.lu/timestamp/
```

### 请求命令
```bash
curl -X POST http://localhost:8888/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "prompts": "请总结一下这段时间的群聊内容",
    "relationId": "all",
    "startTime": 1729468800,
    "endTime": 1729641600
  }'
```

### 成功响应示例
```json
{
  "code": 200,
  "data": {
    "chatType": 4,
    "data": [
      {
        "Type": 1,
        "Title": "指定时间段内的任务",
        "Content": "这是该时间段内讨论的任务内容..."
      }
    ]
  },
  "msg": "success"
}
```

### 验证要点
- ✅ **时间过滤**: 只返回指定时间范围内的消息总结
- 📅 **参数说明**: startTime和endTime为Unix时间戳(秒)
- 🔍 **默认行为**: 不提供时间参数时,默认总结最近24小时消息

---

## 🔍 第四步：测试不同提问方式的总结

### 测试目的
验证AI能理解不同的提问方式,都能正确调用总结功能。

### 测试用例1: 直接请求总结
```bash
curl -X POST http://localhost:8888/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "prompts": "总结群聊",
    "relationId": "all"
  }'
```

### 测试用例2: 询问式提问
```bash
curl -X POST http://localhost:8888/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "prompts": "群里最近都聊了什么?",
    "relationId": "all"
  }'
```

### 测试用例3: 具体内容查询
```bash
curl -X POST http://localhost:8888/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "prompts": "帮我看看群聊中有哪些待办任务",
    "relationId": "all"
  }'
```

### 验证要点
- ✅ **智能路由**: AI能识别不同提问方式,正确路由到chat_log处理器
- 🤖 **自然语言理解**: 支持多种表达方式
- 📊 **结果一致性**: 不同提问方式返回相同格式的总结

---

## 📋 第五步：验证总结结果的数据类型

### 测试目的
理解总结结果中不同Type值的含义,验证分类准确性。

### Type类型说明
| Type值 | 类型名称 | 说明 | 示例 |
|--------|---------|------|------|
| 1 | 待办任务 | 需要执行的工作任务、会议安排等 | "完成新功能开发"、"周三开会" |
| 2 | 审批事项 | 请假申请、报销申请等需要审批的事项 | "李员工请假申请"、"root请假" |

### 验证要点
- ✅ **任务分类**: AI能准确区分待办任务和审批事项
- 📝 **内容完整**: 总结包含关键信息:人员、时间、事项内容
- 🎯 **标题概括**: Title字段准确概括事项核心
- 📄 **详细描述**: Content字段包含完整上下文信息

---

## 🔄 第六步：测试功能兼容性 - 验证原有功能正常

### 测试目的
确保AI群消息总结功能不影响原有的todo、approval、knowledge等功能。

### 测试用例1: Todo查询功能
```bash
curl -X POST http://localhost:8888/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "prompts": "查询我的待办事项"
  }'
```

### 预期响应
```json
{
  "code": 200,
  "data": {
    "data": "```json\n{\"chatType\":1,\"data\":{\"count\":0,\"data\":null}}\n```"
  },
  "msg": "success"
}
```

### 测试用例2: 普通AI对话
```bash
curl -X POST http://localhost:8888/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "prompts": "你好,今天天气怎么样?"
  }'
```

### 预期响应
```json
{
  "code": 200,
  "data": {
    "data": "我无法获取实时天气信息,建议您查看当地天气预报应用..."
  },
  "msg": "success"
}
```

### 验证要点
- ✅ **功能隔离**: 不同功能相互独立,不互相影响
- 🔀 **智能路由**: AI能正确识别用户意图,路由到对应处理器
- ✨ **新功能集成**: 新功能无缝集成到现有系统中

---

## ⚠️ 第七步：测试异常场景

### 场景1: 缺少relationId参数
```bash
curl -X POST http://localhost:8888/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "prompts": "请帮我总结一下最近的群聊消息"
  }'
```

### 预期响应
```json
{
  "code": 500,
  "data": {},
  "msg": "请确定需要总结的会话对象"
}
```

### 场景2: 时间范围内无消息
```bash
curl -X POST http://localhost:8888/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "prompts": "总结群聊消息",
    "relationId": "all",
    "startTime": 946684800,
    "endTime": 946771200
  }'
```

### 预期行为
- 返回空数组或提示该时间段无消息

### 验证要点
- ✅ **参数校验**: 必需参数缺失时返回明确错误信息
- 🔍 **数据处理**: 无数据时优雅处理,不报错
- 📝 **错误提示**: 错误信息清晰,便于定位问题

---

## 🛠️ 第八步：技术实现验证

### 测试目的
验证技术实现的正确性,包括AI路由、参数传递、数据处理等。

### 验证点1: AI智能路由
观察后台日志,确认AI正确识别chat_log处理器:
```json
{
  "out": {
    "destinations": "chat_log",
    "next_inputs": "请帮我总结一下最近的群聊消息"
  }
}
```

### 验证点2: 参数传递机制
- ✅ **Context传递**: relationId、startTime、endTime通过context传递
- ✅ **Memory兼容**: 只有input参数传入router,避免memory冲突

### 验证点3: 数据查询
确认从chat_log集合正确查询数据:
- 查询条件: conversationId="all"
- 时间过滤: sendTime在startTime和endTime之间
- 排序方式: 按sendTime升序

### 验证点4: AI总结质量
- ✅ **上下文理解**: AI能理解对话上下文
- ✅ **人员关系**: 能区分上下级关系
- ✅ **事项分类**: 准确分类为任务或审批
- ✅ **中文输出**: 保持中文输出,格式规范

---

## 📊 测试总结

### 功能测试结果
| 测试场景 | 状态 | 说明 |
|---------|------|------|
| 基本总结功能 | ✅ 通过 | AI能正确总结群聊消息 |
| 时间范围过滤 | ✅ 通过 | 支持自定义时间范围 |
| 多种提问方式 | ✅ 通过 | 智能识别用户意图 |
| 结果分类准确性 | ✅ 通过 | 准确区分任务和审批 |
| 功能兼容性 | ✅ 通过 | 不影响原有功能 |
| 异常处理 | ✅ 通过 | 错误提示清晰 |
| 技术实现 | ✅ 通过 | 架构设计合理 |

### ChatType类型说明
- **chatType值**:
  - `0`: 默认处理器 (DefaultHandler)
  - `1`: Todo查询 (TodoFind)
  - `2`: Todo添加 (TodoAdd)
  - `3`: 审批查询 (ApprovalFind)
  - `4`: **聊天日志总结 (ChatLog)** ← 新增

### 总结数据Type说明
- **Type类型**:
  - `1`: 待办任务 (task to be done)
  - `2`: 审批事项 (approval)

### 系统特性
- ✅ **智能路由**: 基于LangChain的智能路由系统,自动选择合适的处理器
- ✅ **参数隔离**: 通过Context传递额外参数,避免与Memory机制冲突
- ✅ **时间灵活**: 支持自定义时间范围,默认最近24小时
- ✅ **AI增强**: 使用阿里云DashScope进行智能总结,提取关键信息
- ✅ **数据完整**: 总结包含人员信息、时间信息、事项内容等完整上下文
- ✅ **向后兼容**: 完全兼容原有功能,无缝集成

### 使用建议
1. **日常使用**: 每天下班前总结当天群聊内容,了解工作进展
2. **周期回顾**: 指定时间范围总结一周或一月的重要事项
3. **任务跟踪**: 从总结中提取待办任务,及时跟进
4. **审批管理**: 识别审批事项,避免遗漏

### 注意事项
- ⚠️ **必需参数**: relationId参数必须提供(群聊为"all")
- ⚠️ **时间格式**: startTime和endTime使用Unix时间戳(秒级)
- ⚠️ **数据依赖**: 需要chat_log集合中有历史消息数据
- ⚠️ **Token消耗**: AI总结会消耗LLM tokens,注意成本控制

---

## 🔧 故障排查指南

### 问题1: 返回"请确定需要总结的会话对象"
**原因**: 缺少relationId参数
**解决**: 在请求中添加`"relationId": "all"`

### 问题2: 总结结果为空
**原因**:
- 指定时间范围内无消息
- chat_log集合中无数据

**解决**:
- 检查时间范围是否正确
- 确认有用户发送过群聊消息

### 问题3: AI路由错误
**原因**: 提问方式不明确
**解决**: 使用明确的关键词,如"总结"、"群聊消息"等

### 问题4: 功能冲突
**原因**: 修改代码后可能影响其他功能
**解决**: 参考本指南第六步测试原有功能是否正常

---

## 📚 相关文档
- [Todo API测试指南](Todo_API_测试指南.md)
- [Approval API测试指南](Approval_API_测试指南.md)
- [AI Chat API测试指南](AI_Chat_API_测试指南.md)
- [Department API测试指南](Department_API_完整测试指南.md)

---

**文档版本**: v1.0
**最后更新**: 2025-10-21
**作者**: AI辅助团队
**功能状态**: ✅ 测试通过,生产就绪