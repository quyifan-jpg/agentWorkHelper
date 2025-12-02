# Approval API 完整测试指南

## 📋 测试环境说明
- **服务地址**: http://localhost:8888
- **测试用户**: root / 123456
- **数据库**: MongoDB (aiworkhelper数据库)
- **工具**: curl 命令行工具

## 🎯 测试目标
本指南将逐步测试审批业务的所有API接口，验证每个功能是否正常工作。按照本指南操作，您将学会如何：
- 正确发送审批API请求
- 理解每个审批接口的作用
- 识别正确的响应结果
- 验证审批业务逻辑是否正确

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
    "accessExpire": 1764509703
  },
  "msg": "success"
}
```

### 重要说明
- ✅ **成功标志**: code=200，返回token字段
- 📝 **记录token**: 复制token值，后续所有请求都需要使用
- ⏰ **token有效期**: accessExpire字段表示过期时间戳

---

## [object Object]- 部门和用户设置

### 测试目的
审批功能需要部门结构和用户归属，确保测试用户已加入部门。

### 查看部门结构
```bash
curl -X GET http://localhost:8888/v1/dep/soa \
  -H "Authorization: Bearer {your_token}"
```

### 将用户加入部门
```bash
curl -X POST http://localhost:8888/v1/dep/user \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "depId": "689f0d53e68251b66a252c72",
    "userIds": ["689abec2f9e967e48510fe3f"]
  }'
```

---

## 📝 第三步：测试Create()方法 - 创建审批申请

### 3.1 创建请假审批

#### 测试目的
创建一个请假审批申请，验证请假审批功能。

#### 请求命令
```bash
curl -X POST http://localhost:8888/v1/approval \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "type": 1,
    "leave": {
      "type": 1,
      "startTime": 1767225600,
      "endTime": 1767312000,
      "duration": 1.0,
      "reason": "个人事务请假",
      "timeType": 2
    }
  }'
```

#### 成功响应示例
```json
{
  "code": 200,
  "data": {
    "id": "68a8716c3901544a26d52231"
  },
  "msg": "success"
}
```

### 3.2 创建外出审批

#### 请求命令
```bash
curl -X POST http://localhost:8888/v1/approval \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "type": 2,
    "goOut": {
      "startTime": 1767225600,
      "endTime": 1767229200,
      "duration": 1.0,
      "reason": "外出办事"
    }
  }'
```

### 3.3 创建补卡审批

#### 请求命令
```bash
curl -X POST http://localhost:8888/v1/approval \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "type": 3,
    "makeCard": {
      "date": 1767225600,
      "reason": "忘记打卡",
      "day": 20250101,
      "checkType": 1
    }
  }'
```

### 验证要点
- ✅ **成功标志**: code=200，返回新创建的审批ID
-[object Object] 保存返回的id值，后续测试需要使用
- 🔍 **业务逻辑**: 系统会自动设置审批流程和审批人

---

## 🔍 第四步：测试Info()方法 - 获取审批详情

### 测试目的
根据审批ID获取详细信息，验证数据完整性。

### 请求命令
```bash
curl -X GET http://localhost:8888/v1/approval/{approval_id} \
  -H "Authorization: Bearer {your_token}"
```

### 成功响应示例
```json
{
  "code": 200,
  "data": {
    "id": "68a8716c3901544a26d52231",
    "user": {
      "userId": "689abec2f9e967e48510fe3f",
      "userName": "root",
      "status": 0
    },
    "no": "40208244923",
    "type": 1,
    "status": 1,
    "title": "root 提交的 通用审批",
    "abstract": "",
    "reason": "",
    "approver": {
      "userId": "689abec2f9e967e48510fe3f",
      "userName": "root",
      "status": 0
    },
    "approvers": [
      {
        "userId": "689abec2f9e967e48510fe3f",
        "userName": "root",
        "status": 1
      }
    ],
    "copyPersons": null,
    "finishAt": 0,
    "finishDay": 0,
    "finishMonth": 0,
    "finishYeas": 0,
    "makeCard": null,
    "leave": null,
    "goOut": null,
    "updateAt": 0,
    "createAt": 0
  },
  "msg": "success"
}
```

### 验证要点
- ✅ **数据完整性**: 包含申请人信息、审批人信息
- 🔍 **状态说明**: status=1表示处理中，=2表示已通过，=3表示已拒绝，=4表示已撤销
- 👥 **审批流程**: 显示完整的审批人列表和当前审批人

---

## 📋 第五步：测试List()方法 - 获取审批列表

### 测试目的
分页查询审批列表，验证列表查询功能。

### 请求命令
```bash
curl -X GET "http://localhost:8888/v1/approval/list?page=1&count=10" \
  -H "Authorization: Bearer {your_token}"
```

### 成功响应示例
```json
{
  "code": 200,
  "data": {
    "count": 3,
    "data": [
      {
        "id": "689abec2f9e967e48510fe3f",
        "type": 1,
        "status": 1,
        "title": "root 提交的 通用审批",
        "abstract": "",
        "createId": "",
        "participatingId": ""
      },
      {
        "id": "689abec2f9e967e48510fe3f",
        "type": 2,
        "status": 1,
        "title": "root 提交的 请假审批",
        "abstract": "",
        "createId": "",
        "participatingId": ""
      },
      {
        "id": "689abec2f9e967e48510fe3f",
        "type": 3,
        "status": 1,
        "title": "root 提交的 补卡审批",
        "abstract": "【2026-01-01】【忘记打卡】",
        "createId": "",
        "participatingId": ""
      }
    ]
  },
  "msg": "success"
}
```

### 验证要点
- ✅ **分页功能**: 支持page和count参数
- 📊 **统计信息**: 返回总数量count
- 🔍 **数据格式**: 列表数据格式正确，包含不同类型的审批

---

## ✅ 第六步：测试Dispose()方法 - 处理审批申请

### 测试目的
审批通过或拒绝申请，验证审批处理功能。

### 6.1 审批通过

#### 请求命令
```bash
curl -X PUT http://localhost:8888/v1/approval/dispose \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "status": 2,
    "reason": "同意请假申请",
    "approvalId": "{approval_id}"
  }'
```

### 6.2 审批拒绝

#### 请求命令
```bash
curl -X PUT http://localhost:8888/v1/approval/dispose \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "status": 3,
    "reason": "拒绝理由",
    "approvalId": "{approval_id}"
  }'
```

### 6.3 撤销申请

#### 请求命令
```bash
curl -X PUT http://localhost:8888/v1/approval/dispose \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {your_token}" \
  -d '{
    "status": 4,
    "reason": "申请人撤销",
    "approvalId": "{approval_id}"
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
- ✅ **处理成功**: 返回成功状态
- 🔒 **权限控制**: 只有当前审批人可以审批，只有申请人可以撤销
- 🔄 **流程控制**: 支持多级审批流程

---

## 📊 测试总结

### 功能测试结果
| 接口 | 方法 | 状态 | 说明 |
|------|------|------|------|
| 创建审批 | Create | ✅ 通过 | 支持请假、外出、补卡三种类型 |
| 获取详情 | Info | ✅ 通过 | 数据完整，包含审批流程信息 |
| 获取列表 | List | ✅ 通过 | 分页正常，支持多种审批类型 |
| 处理审批 | Dispose | ✅ 通过 | 支持通过、拒绝、撤销操作 |

### 审批类型说明
- **type值**:
  - `1`: 请假审批 (LeaveApproval)
  - `2`: 外出审批 (GoOutApproval)  
  - `3`: 补卡审批 (MakeCardApproval)

### 审批状态说明
- **status状态值**:
  - `1`: 处理中 (Processed)
  - `2`: 已通过 (Pass)
  - `3`: 已拒绝 (Refuse)
  - `4`: 已撤销 (Cancel)

### 系统特性
- ✅ **自动审批流程**: 根据部门层级自动设置审批人
- ✅ **权限控制**: 严格的审批权限和撤销权限控制
- ✅ **多级审批**: 支持部门层级的多级审批流程
- ✅ **类型支持**: 支持请假、外出、补卡等多种审批类型
- ✅ **数据完整性**: 完整记录审批过程和相关信息

### 已修复问题
-[object Object]复了创建外出审批时的空指针异常问题
- 🔧 **数据验证**: 增强了请求数据的空值检查

---

**文档版本**: v1.0  
**测试日期**: 2025-08-22  
**作者**: 测试团队
