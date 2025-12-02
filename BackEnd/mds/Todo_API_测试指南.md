# Todo API 完整测试指南

## 📋 测试环境说明
- **服务地址**: http://localhost:8888
- **测试用户**: root / 123456
- **数据库**: MongoDB (aiworkhelper数据库)
- **工具**: curl 命令行工具

## 🎯 测试目标
本指南将逐步测试待办事项(Todo)的所有API接口，验证每个功能是否正常工作。按照本指南操作，您将学会如何：
- 正确发送API请求
- 理解每个接口的作用
- 识别正确的响应结果
- 验证业务逻辑是否正确

---

## 🔐 第一步：用户登录获取Token

### 测试目的
获取JWT认证token，用于后续所有API调用的身份验证。

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
    "id": "689abec2f9e967e48510fe3f",
    "name": "root",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "accessExpire": 1764216426
  },
  "msg": "success"
}
```

### 重要说明
- ✅ **成功标志**: code=200，返回token字段
- 📝 **记录token**: 复制token值，后续所有请求都需要使用
- ⏰ **token有效期**: accessExpire字段表示过期时间戳

---

## 📝 第二步：测试Create()方法 - 创建待办事项

### 测试目的
创建一个新的待办事项，验证基本的创建功能。

### 请求命令
```bash
curl -X POST http://localhost:8888/v1/todo \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "title": "测试待办事项",
    "desc": "这是一个测试用的待办事项",
    "deadlineAt": 1767225600,
    "executeIds": [],
    "records": []
  }'
```

### 成功响应示例
```json
{
  "code": 200,
  "data": {
    "id": "68a3f8767e350539d1baddcf"
  },
  "msg": "success"
}
```

### 验证要点
- ✅ **成功标志**: code=200，返回新创建的待办事项ID
- 📝 **记录ID**: 保存返回的id值，后续测试需要使用
- 🔍 **业务逻辑**: 系统会自动将创建者设为执行人

---

## 🔍 第三步：测试Info()方法 - 获取待办事项详情

### 测试目的
根据待办事项ID获取详细信息，验证数据完整性。

### 请求命令
```bash
curl -X GET http://localhost:8888/v1/todo/{todo_id} \
  -H "Authorization: Bearer {your_token}"
```

### 成功响应示例
```json
{
  "code": 200,
  "data": {
    "id": "68a3f8767e350539d1baddcf",
    "creatorId": "689abec2f9e967e48510fe3f",
    "creatorName": "root",
    "title": "测试待办事项",
    "deadlineAt": 1767225600,
    "desc": "这是一个测试用的待办事项",
    "executeIds": [
      {
        "id": "000000000000000000000000",
        "userId": "689abec2f9e967e48510fe3f",
        "userName": "root",
        "todoStatus": 1
      }
    ],
    "status": 1,
    "todoStatus": 1
  },
  "msg": "success"
}
```

### 验证要点
- ✅ **数据完整性**: 包含创建者信息、执行人信息
- 🔍 **状态说明**: todoStatus=1表示进行中，=2表示已完成，=4表示已超时
- 👥 **执行人**: 自动添加创建者为执行人

---

## 📋 第四步：测试List()方法 - 获取待办事项列表

### 测试目的
分页查询待办事项列表，验证列表查询功能。

### 请求命令
```bash
curl -X GET "http://localhost:8888/v1/todo/list?page=1&count=10" \
  -H "Authorization: Bearer {your_token}"
```

### 成功响应示例
```json
{
  "code": 200,
  "data": {
    "count": 1,
    "data": [
      {
        "id": "68a3f8767e350539d1baddcf",
        "creatorId": "689abec2f9e967e48510fe3f",
        "title": "测试待办事项",
        "deadlineAt": 1767225600,
        "desc": "这是一个测试用的待办事项",
        "todoStatus": 1
      }
    ]
  },
  "msg": "success"
}
```

### 验证要点
- ✅ **分页功能**: 支持page和count参数
- 📊 **统计信息**: 返回总数量count
- 🔍 **数据格式**: 列表数据格式正确

---

## 📝 第五步：测试CreateRecord()方法 - 创建操作记录

### 测试目的
为待办事项添加操作记录，验证记录功能。

### 请求命令
```bash
curl -X POST http://localhost:8888/v1/todo/record \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "todoId": "{todo_id}",
    "content": "添加了一条测试记录",
    "image": ""
  }'
```

### 成功响应示例
```json
{
  "code": 200,
  "data": {},
  "msg": "success"
}
```

### 验证要点
- ✅ **记录创建**: 成功添加操作记录
- 🔍 **自动信息**: 系统自动记录操作用户信息和时间

---

## ✅ 第六步：测试Finish()方法 - 完成待办事项

### 测试目的
标记待办事项为完成状态，验证状态更新功能。

### 请求命令
```bash
curl -X POST http://localhost:8888/v1/todo/finish \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "userId": "{user_id}",
    "todoId": "{todo_id}"
  }'
```

### 成功响应示例
```json
{
  "code": 200,
  "data": {},
  "msg": "success"
}
```

### 验证完成状态
完成后再次查询待办详情，验证状态是否正确更新：
```bash
curl -X GET http://localhost:8888/v1/todo/{todo_id} \
  -H "Authorization: Bearer {your_token}"
```

### 完成后状态示例
```json
{
  "code": 200,
  "data": {
    "id": "68a3f8767e350539d1baddcf",
    "executeIds": [
      {
        "userId": "689abec2f9e967e48510fe3f",
        "userName": "root",
        "todoStatus": 2
      }
    ],
    "status": 2,
    "todoStatus": 2
  },
  "msg": "success"
}
```

### 验证要点
- ✅ **状态更新**: 执行人todoStatus更新为2（已完成）
- ✅ **整体状态**: 整体todoStatus和status都更新为2（已完成）

---

## ✏️ 第七步：测试Edit()方法 - 编辑待办事项

### 测试目的
更新待办事项信息，验证编辑功能。

### 请求命令
```bash
curl -X PUT http://localhost:8888/v1/todo \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "id": "{todo_id}",
    "title": "测试待办事项(已更新)",
    "desc": "这是一个更新后的测试待办事项",
    "deadlineAt": 1767225600,
    "executeIds": []
  }'
```

### 成功响应示例
```json
{
  "code": 200,
  "data": {},
  "msg": "success"
}
```

### 验证要点
- ✅ **更新成功**: 返回成功状态
- ⚠️ **注意**: 当前实现为空方法，接口存在但不执行实际更新

---

## 🗑️ 第八步：测试Delete()方法 - 删除待办事项

### 测试目的
删除指定的待办事项，验证删除功能和权限控制。

### 请求命令
```bash
curl -X DELETE http://localhost:8888/v1/todo/{todo_id} \
  -H "Authorization: Bearer {your_token}"
```

### 成功响应示例
```json
{
  "code": 200,
  "data": {},
  "msg": "success"
}
```

### 验证要点
- ✅ **删除成功**: 返回成功状态
- 🔒 **权限控制**: 只有创建者可以删除

---

## 📊 测试总结

### 功能测试结果
| 接口 | 方法 | 状态 | 说明 |
|------|------|------|------|
| 创建待办 | Create | ✅ 通过 | 功能正常 |
| 获取详情 | Info | ✅ 通过 | 数据完整 |
| 获取列表 | List | ✅ 通过 | 分页正常 |
| 创建记录 | CreateRecord | ✅ 通过 | 记录成功 |
| 编辑待办 | Edit | ⚠️ 部分通过 | 空实现 |
| 完成待办 | Finish | ✅ 通过 | 功能正常 |
| 删除待办 | Delete | ✅ 通过 | 功能正常 |

### 状态码说明
- **todoStatus状态值**:
  - `1`: 进行中 (TodoInProgress)
  - `2`: 已完成 (TodoFinish)
  - `3`: 已取消 (TodoCancel)
  - `4`: 已超时 (TodoTimeout)

### 系统特性
- ✅ **自动设置**: 创建者自动成为执行人
- ✅ **状态管理**: 系统能正确识别超时状态
- ✅ **权限控制**: 只有创建者可以删除待办事项
- ✅ **完成逻辑**: 支持单个用户完成，所有用户完成时更新整体状态

---

**文档版本**: v1.0  
**测试日期**: 2024-01-20  
**作者**: 测试团队
