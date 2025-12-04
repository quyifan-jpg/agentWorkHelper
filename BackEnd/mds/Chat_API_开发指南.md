# Chat API 开发指南

## 📋 概述

本文档基于 AIWorkHelper 项目的 Chat 功能实现，指导在 BackEnd 项目中实现 Chat API。Chat 功能包括：
- **私聊**：两个用户之间的点对点聊天
- **群聊**：群组内的广播聊天
- **聊天记录查询**：根据会话ID和时间范围查询历史消息
- **会话ID生成**：为私聊自动生成唯一会话ID

## 🎯 功能对比

| 功能 | AIWorkHelper | BackEnd |
|------|-------------|---------|
| 数据库 | MongoDB | MySQL + GORM |
| ID类型 | string (ObjectID) | uint (自增) |
| 会话ID生成 | SHA256哈希 | SHA256哈希（相同算法） |
| 群聊ID | "all" 或自定义 | "all" 或自定义 |

---

## 📦 第一步：数据模型设计

### 1.1 创建 ChatLog 模型

**文件**: `BackEnd/internal/model/chatlog.go`

```go
package model

import (
	"time"
	"gorm.io/gorm"
)

// ChatType 聊天类型枚举
type ChatType int

const (
	GroupChatType  ChatType = 1 // 群聊类型
	SingleChatType ChatType = 2 // 私聊类型
)

// ChatLog 聊天记录数据模型
type ChatLog struct {
	gorm.Model
	ConversationId string   `gorm:"type:varchar(64);index;comment:会话ID"` // 会话ID，群聊为"all"或群ID，私聊为生成的唯一标识
	SendId         uint     `gorm:"index;comment:发送者用户ID"`            // 发送者ID
	RecvId         uint     `gorm:"index;default:0;comment:接收者用户ID"`  // 接收者ID，群聊时为0
	ChatType       ChatType `gorm:"default:2;comment:聊天类型：1=群聊，2=私聊"` // 聊天类型
	MsgContent     string   `gorm:"type:text;comment:消息内容"`          // 消息内容
	SendTime       int64    `gorm:"index;comment:发送时间戳"`              // 发送时间戳
}

// TableName 指定表名
func (ChatLog) TableName() string {
	return "chat_logs"
}
```

### 1.2 数据库迁移

在 `BackEnd/cmd/api/main.go` 中添加：

```go
// 数据库迁移
if err := svcCtx.DB.AutoMigrate(
	&model.User{},
	&model.ChatLog{}, // 添加聊天记录表
); err != nil {
	panic(err)
}
```

---

## 📝 第二步：API 定义

### 2.1 创建 chat.api 文件

**文件**: `BackEnd/doc/chat.api`

```api
syntax = "v1"

import "base.api"

info (
	title:  "Chat API"
	author: "BackEnd"
)

type (
	// 发送消息请求
	SendMessageReq {
		RecvId         string `json:"recvId,omitempty"`         // 接收者ID（私聊必填，群聊为空）
		ChatType       int    `json:"chatType" binding:"required"` // 聊天类型：1=群聊，2=私聊
		ConversationId string `json:"conversationId,omitempty"`  // 会话ID（可选，私聊会自动生成）
		Content        string `json:"content" binding:"required"` // 消息内容
		ContentType    int    `json:"contentType,omitempty"`     // 内容类型：1=文本，2=图片等
	}

	// 发送消息响应
	SendMessageResp {
		ConversationId string `json:"conversationId"` // 会话ID
		SendId         string `json:"sendId"`         // 发送者ID
		RecvId         string `json:"recvId"`         // 接收者ID
		ChatType       int    `json:"chatType"`       // 聊天类型
		Content        string `json:"content"`        // 消息内容
		ContentType    int    `json:"contentType"`    // 内容类型
		SendTime       int64  `json:"sendTime"`      // 发送时间戳
	}

	// 查询聊天记录请求
	ChatListReq {
		ConversationId string `json:"conversationId" binding:"required"` // 会话ID
		StartTime      int64  `json:"startTime,omitempty"`               // 开始时间戳
		EndTime        int64  `json:"endTime,omitempty"`                 // 结束时间戳
		Page           int    `json:"page,omitempty"`                     // 页码
		Count          int    `json:"count,omitempty"`                    // 每页数量
	}

	// 聊天记录项
	ChatLogItem {
		Id             string `json:"id"`             // 记录ID
		ConversationId string `json:"conversationId"` // 会话ID
		SendId         string `json:"sendId"`         // 发送者ID
		RecvId         string `json:"recvId"`        // 接收者ID
		ChatType       int    `json:"chatType"`       // 聊天类型
		Content        string `json:"content"`        // 消息内容
		ContentType    int    `json:"contentType"`    // 内容类型
		SendTime       int64  `json:"sendTime"`      // 发送时间戳
	}

	// 查询聊天记录响应
	ChatListResp {
		Count int64         `json:"count"` // 总记录数
		List  []ChatLogItem `json:"list"`  // 聊天记录列表
	}
)

// 聊天服务 - 需要认证
@server (
	group:      v1/chat
	logic:      Chat
	middleware: Jwt
)
service Chat {
	@handler SendMessage
	post /message (SendMessageReq) returns (SendMessageResp)

	@handler ListMessages
	get /list (ChatListReq) returns (ChatListResp)
}
```

### 2.2 生成类型定义

运行代码生成脚本：

```bash
cd BackEnd
./scripts/gen.sh
```

这将自动生成 `BackEnd/internal/domain/domain.go` 中的类型定义（如果使用 goctl-gin）或需要手动在 `BackEnd/internal/domain/` 中创建对应的类型文件。

---

## 🔧 第三步：Logic 层实现

### 3.1 创建 Chat Logic 接口

**文件**: `BackEnd/internal/logic/chat.go`

```go
package logic

import (
	"BackEnd/internal/domain"
	"BackEnd/internal/model"
	"BackEnd/internal/svc"
	"BackEnd/pkg/token"
	"BackEnd/pkg/util"
	"BackEnd/pkg/xerr"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"sort"
	"time"

	"github.com/rs/zerolog/log"
)

type Chat interface {
	SendMessage(ctx context.Context, req *domain.SendMessageReq) (*domain.SendMessageResp, error)
	ListMessages(ctx context.Context, req *domain.ChatListReq) (*domain.ChatListResp, error)
}

type chat struct {
	svcCtx *svc.ServiceContext
}

func NewChat(svcCtx *svc.ServiceContext) Chat {
	return &chat{
		svcCtx: svcCtx,
	}
}

// SendMessage 发送消息（私聊或群聊）
func (l *chat) SendMessage(ctx context.Context, req *domain.SendMessageReq) (*domain.SendMessageResp, error) {
	// 1. 获取当前用户ID
	userID, err := token.GetUserID(ctx)
	if err != nil {
		return nil, xerr.New(errors.New("user not authenticated"))
	}

	// 2. 验证请求参数
	if req.ChatType != int(model.GroupChatType) && req.ChatType != int(model.SingleChatType) {
		return nil, xerr.New(errors.New("invalid chat type"))
	}

	// 3. 处理会话ID
	conversationId := req.ConversationId
	if conversationId == "" {
		if req.ChatType == int(model.SingleChatType) {
			// 私聊：生成唯一会话ID
			if req.RecvId == "" {
				return nil, xerr.New(errors.New("recvId is required for single chat"))
			}
			recvID, err := util.StringToUint(req.RecvId)
			if err != nil {
				return nil, xerr.New(errors.New("invalid recvId"))
			}
			conversationId = GenerateConversationId(userID, recvID)
		} else {
			// 群聊：使用 "all" 作为默认会话ID
			conversationId = "all"
		}
	}

	// 4. 处理接收者ID
	var recvID uint
	if req.ChatType == int(model.SingleChatType) {
		recvID, err = util.StringToUint(req.RecvId)
		if err != nil {
			return nil, xerr.New(errors.New("invalid recvId"))
		}
	}

	// 5. 创建聊天记录
	chatLog := &model.ChatLog{
		ConversationId: conversationId,
		SendId:         userID,
		RecvId:         recvID,
		ChatType:       model.ChatType(req.ChatType),
		MsgContent:     req.Content,
		SendTime:       time.Now().Unix(),
	}

	if err := l.svcCtx.DB.WithContext(ctx).Create(chatLog).Error; err != nil {
		log.Error().Err(err).Msg("failed to create chat log")
		return nil, xerr.New(err)
	}

	// 6. 返回响应
	return &domain.SendMessageResp{
		ConversationId: conversationId,
		SendId:         util.UintToString(userID),
		RecvId:         util.UintToString(recvID),
		ChatType:       req.ChatType,
		Content:        req.Content,
		ContentType:    req.ContentType,
		SendTime:       chatLog.SendTime,
	}, nil
}

// ListMessages 查询聊天记录列表
func (l *chat) ListMessages(ctx context.Context, req *domain.ChatListReq) (*domain.ChatListResp, error) {
	// 1. 处理分页参数
	pagination := util.NormalizePagination(req.Page, req.Count)

	// 2. 构建查询
	db := l.svcCtx.DB.WithContext(ctx).Model(&model.ChatLog{}).
		Where("conversation_id = ?", req.ConversationId)

	// 3. 时间范围过滤
	if req.StartTime > 0 {
		db = db.Where("send_time >= ?", req.StartTime)
	}
	if req.EndTime > 0 {
		db = db.Where("send_time <= ?", req.EndTime)
	}

	// 4. 查询总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		log.Error().Err(err).Msg("failed to count chat logs")
		return nil, xerr.New(err)
	}

	// 5. 查询列表数据
	var chatLogs []model.ChatLog
	if err := db.Order("send_time desc").
		Offset(pagination.Offset).
		Limit(pagination.Count).
		Find(&chatLogs).Error; err != nil {
		log.Error().Err(err).Msg("failed to list chat logs")
		return nil, xerr.New(err)
	}

	// 6. 转换为响应格式
	list := make([]*domain.ChatLogItem, 0, len(chatLogs))
	for _, log := range chatLogs {
		list = append(list, &domain.ChatLogItem{
			Id:             util.UintToString(log.ID),
			ConversationId: log.ConversationId,
			SendId:         util.UintToString(log.SendId),
			RecvId:         util.UintToString(log.RecvId),
			ChatType:       int(log.ChatType),
			Content:        log.MsgContent,
			ContentType:    1, // 默认文本类型
			SendTime:       log.SendTime,
		})
	}

	return &domain.ChatListResp{
		Count: total,
		List:  list,
	}, nil
}

// GenerateConversationId 生成私聊的唯一会话ID
// 算法：对两个用户ID排序后拼接，计算SHA256哈希，取Base64编码的前22位
func GenerateConversationId(userId1, userId2 uint) string {
	// 转换为字符串并排序
	ids := []string{
		util.UintToString(userId1),
		util.UintToString(userId2),
	}
	sort.Strings(ids)

	// 拼接
	combined := ids[0] + ids[1]

	// 计算SHA256哈希
	hasher := sha256.New()
	hasher.Write([]byte(combined))
	hash := hasher.Sum(nil)

	// 返回Base64编码的前22位
	return base64.RawStdEncoding.EncodeToString(hash)[:22]
}
```

### 3.2 在 ServiceContext 中注册

**文件**: `BackEnd/internal/svc/servicecontext.go`

```go
type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
	Jwt    *middleware.Jwt
	User   logic.UserLogic
	Chat   logic.Chat // 添加 Chat Logic
	// ... 其他服务
}

func NewServiceContext(c config.Config) *ServiceContext {
	// ... 初始化代码
	return &ServiceContext{
		Config: c,
		DB:     db,
		Jwt:    jwtMiddleware,
		User:   logic.NewUserLogic(svcCtx),
		Chat:   logic.NewChat(svcCtx), // 添加 Chat Logic
		// ... 其他服务
	}
}
```

---

## 🌐 第四步：Handler 层实现

### 4.1 创建 Chat Handler

**文件**: `BackEnd/internal/handler/api/chat.go`

```go
package api

import (
	"BackEnd/internal/domain"
	"BackEnd/internal/logic"
	"BackEnd/internal/svc"
	"BackEnd/pkg/httpx"
	"github.com/gin-gonic/gin"
)

type Chat struct {
	svcCtx *svc.ServiceContext
	chat   logic.Chat
}

func NewChat(svcCtx *svc.ServiceContext) *Chat {
	return &Chat{
		svcCtx: svcCtx,
		chat:   svcCtx.Chat,
	}
}

func (h *Chat) InitRegister(engine *gin.Engine) {
	g := engine.Group("v1/chat", h.svcCtx.Jwt.Handler)
	g.POST("/message", h.SendMessage)
	g.GET("/list", h.ListMessages)
}

// SendMessage 发送消息
// @Summary 发送消息
// @Description 发送私聊或群聊消息
// @Tags chat
// @Accept json
// @Produce json
// @Param req body domain.SendMessageReq true "消息内容"
// @Success 200 {object} object{code=int,msg=string,data=domain.SendMessageResp}
// @Router /v1/chat/message [post]
func (h *Chat) SendMessage(ctx *gin.Context) {
	var req domain.SendMessageReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	resp, err := h.chat.SendMessage(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.Success(ctx, resp)
}

// ListMessages 查询聊天记录
// @Summary 查询聊天记录
// @Description 根据会话ID查询聊天记录列表
// @Tags chat
// @Accept json
// @Produce json
// @Param conversationId query string true "会话ID"
// @Param startTime query int false "开始时间戳"
// @Param endTime query int false "结束时间戳"
// @Param page query int false "页码"
// @Param count query int false "每页数量"
// @Success 200 {object} object{code=int,msg=string,data=domain.ChatListResp}
// @Router /v1/chat/list [get]
func (h *Chat) ListMessages(ctx *gin.Context) {
	var req domain.ChatListReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	resp, err := h.chat.ListMessages(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.Success(ctx, resp)
}
```

### 4.2 在 Router 中注册

**文件**: `BackEnd/internal/handler/api/router.go`

```go
func (h *ApiHandler) InitRegister(engine *gin.Engine) {
	// ... 其他路由注册
	NewChat(h.svcCtx).InitRegister(engine)
}
```

---

## 🔑 关键实现细节

### 5.1 会话ID生成算法

私聊的会话ID生成算法与 AIWorkHelper 保持一致：

```go
// 1. 将两个用户ID转换为字符串并排序
ids := []string{userId1, userId2}
sort.Strings(ids)

// 2. 拼接后计算SHA256哈希
combined := ids[0] + ids[1]
hash := sha256.Sum256([]byte(combined))

// 3. Base64编码并取前22位
conversationId := base64.RawStdEncoding.EncodeToString(hash[:])[:22]
```

**特点**：
- 同一对用户的会话ID始终相同
- 无论谁先发送消息，生成的ID都一致
- 22位长度足够唯一且不会太长

### 5.2 群聊会话ID

- 默认使用 `"all"` 作为群聊会话ID
- 也可以支持自定义群ID（通过 `conversationId` 字段传入）

### 5.3 ID 类型转换

- API 层：使用 `string` 类型（前端友好）
- Logic 层：转换为 `uint` 进行数据库操作
- 使用 `util.StringToUint` 和 `util.UintToString` 工具函数

---

## 🧪 第五步：测试

### 5.1 创建测试脚本

**文件**: `BackEnd/scripts/test_chat.sh`

```bash
#!/bin/bash

BASE_URL="http://localhost:8889"
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}=== Chat API 测试 ===${NC}\n"

# 1. 登录获取 Token
echo -e "${YELLOW}步骤 1: 登录用户1${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/v1/user/login" \
  -H "Content-Type: application/json" \
  -d '{"name": "root", "password": "123456"}')

TOKEN1=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token' 2>/dev/null)
USER_ID1=$(echo "$LOGIN_RESPONSE" | jq -r '.data.id' 2>/dev/null)

if [ -z "$TOKEN1" ] || [ "$TOKEN1" == "null" ]; then
    echo -e "${RED}❌ 登录失败${NC}"
    exit 1
fi
echo -e "${GREEN}✅ 登录成功，用户ID: ${USER_ID1}${NC}\n"

# 2. 登录用户2
echo -e "${YELLOW}步骤 2: 登录用户2${NC}"
LOGIN_RESPONSE2=$(curl -s -X POST "${BASE_URL}/v1/user/login" \
  -H "Content-Type: application/json" \
  -d '{"name": "testuser1", "password": "123456"}')

TOKEN2=$(echo "$LOGIN_RESPONSE2" | jq -r '.data.token' 2>/dev/null)
USER_ID2=$(echo "$LOGIN_RESPONSE2" | jq -r '.data.id' 2>/dev/null)

if [ -z "$TOKEN2" ] || [ "$TOKEN2" == "null" ]; then
    echo -e "${RED}❌ 登录失败${NC}"
    exit 1
fi
echo -e "${GREEN}✅ 登录成功，用户ID: ${USER_ID2}${NC}\n"

# 3. 用户1发送私聊消息
echo -e "${YELLOW}步骤 3: 用户1发送私聊消息${NC}"
SEND_RESPONSE=$(curl -s -X POST "${BASE_URL}/v1/chat/message" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN1}" \
  -d "{
    \"recvId\": \"${USER_ID2}\",
    \"chatType\": 2,
    \"content\": \"你好，这是一条私聊消息\"
  }")

CONVERSATION_ID=$(echo "$SEND_RESPONSE" | jq -r '.data.conversationId' 2>/dev/null)
echo -e "${GREEN}✅ 发送成功，会话ID: ${CONVERSATION_ID}${NC}\n"

# 4. 查询聊天记录
echo -e "${YELLOW}步骤 4: 查询聊天记录${NC}"
LIST_RESPONSE=$(curl -s -X GET "${BASE_URL}/v1/chat/list?conversationId=${CONVERSATION_ID}" \
  -H "Authorization: Bearer ${TOKEN1}")

echo "$LIST_RESPONSE" | jq '.'
echo -e "${GREEN}✅ 查询成功${NC}\n"

# 5. 发送群聊消息
echo -e "${YELLOW}步骤 5: 发送群聊消息${NC}"
GROUP_RESPONSE=$(curl -s -X POST "${BASE_URL}/v1/chat/message" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN1}" \
  -d '{
    "chatType": 1,
    "content": "大家好，这是一条群聊消息"
  }')

echo "$GROUP_RESPONSE" | jq '.'
echo -e "${GREEN}✅ 群聊消息发送成功${NC}\n"

echo -e "${GREEN}=== 测试完成 ===${NC}"
```

### 5.2 运行测试

```bash
chmod +x BackEnd/scripts/test_chat.sh
./BackEnd/scripts/test_chat.sh
```

---

## 📊 数据库表结构

### chat_logs 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint (PK) | 主键，自增 |
| conversation_id | varchar(64) | 会话ID，索引 |
| send_id | uint | 发送者ID，索引 |
| recv_id | uint | 接收者ID，索引 |
| chat_type | int | 聊天类型：1=群聊，2=私聊 |
| msg_content | text | 消息内容 |
| send_time | bigint | 发送时间戳，索引 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |
| deleted_at | datetime | 软删除时间 |

---

## 🔄 与 AIWorkHelper 的差异

| 项目 | AIWorkHelper | BackEnd |
|------|-------------|---------|
| **数据库** | MongoDB | MySQL |
| **ID类型** | string (ObjectID) | uint (自增) |
| **查询方式** | MongoDB查询 | GORM查询 |
| **会话ID生成** | 相同算法 | 相同算法 |
| **群聊ID** | "all" 或自定义 | "all" 或自定义 |

---

## ✅ 完成 checklist

- [ ] 创建 `ChatLog` 数据模型
- [ ] 数据库迁移（AutoMigrate）
- [ ] 创建 `chat.api` 文件
- [ ] 运行 `gen.sh` 生成类型
- [ ] 实现 `Chat` Logic 接口
- [ ] 实现 `Chat` Handler
- [ ] 在 ServiceContext 中注册
- [ ] 在 Router 中注册路由
- [ ] 运行测试脚本验证功能
- [ ] 生成 Swagger 文档

---

## 📚 参考资源

- AIWorkHelper Chat 实现：`AIWorkHelper/internal/logic/chat.go`
- AIWorkHelper Chat 模型：`AIWorkHelper/internal/model/chatlogtypes.go`
- AIWorkHelper Chat API：`AIWorkHelper/doc/chat.api`

---

## 🎯 下一步

完成基础 Chat API 后，可以考虑实现：
1. **WebSocket 实时聊天**：使用 WebSocket 实现实时消息推送
2. **文件上传**：支持图片、文件等多媒体消息
3. **消息已读状态**：标记消息已读/未读
4. **消息撤回**：支持消息撤回功能

