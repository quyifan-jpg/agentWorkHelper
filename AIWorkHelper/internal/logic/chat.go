/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package logic 提供聊天相关的业务逻辑处理
package logic

import (
	"AIWorkHelper/internal/domain"
	"AIWorkHelper/internal/logic/chatinternal"
	"AIWorkHelper/internal/model"
	"AIWorkHelper/internal/svc"
	"AIWorkHelper/pkg/langchain"
	"AIWorkHelper/pkg/langchain/memoryx"
	"AIWorkHelper/pkg/langchain/router"
	"AIWorkHelper/pkg/token"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"gitee.com/dn-jinmin/tlog"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/schema"
)

// Chat 聊天业务逻辑接口，定义私聊和群聊的处理方法
type Chat interface {
	PrivateChat(ctx context.Context, req *domain.Message) error                         // 处理私聊消息
	GroupChat(ctx context.Context, req *domain.Message) (uids []string, err error)      // 处理群聊消息
	AIChat(ctx context.Context, req *domain.ChatReq) (resp *domain.ChatResp, err error) // 处理AI聊天请求
	File(ctx context.Context, req []*domain.FileResp) (err error)                       // 处理文件消息
}

// chat 聊天业务逻辑实现结构体
type chat struct {
	svc    *svc.ServiceContext // 服务上下文，包含数据库连接等依赖
	router *router.Router      // 智能路由器，用于选择合适的处理器
	memory schema.Memory       // 多会话内存管理器，支持对话历史记忆
}

// NewChat 创建聊天业务逻辑实例
func NewChat(svc *svc.ServiceContext) Chat {
	// 初始化处理器列表，包含待办事项处理器
	handlers := []router.Handler{
		chatinternal.NewTodoHandle(svc),     // 待办事项处理器
		chatinternal.NewKnowledge(svc),      // 知识库处理器
		chatinternal.NewApprovalHandle(svc), // 审批处理器
		chatinternal.NewChatLogHandle(svc),  // 聊天处理器
	}
	// 创建多会话内存管理器，支持对话历史记忆
	memory := memoryx.NewMemoryx(func() schema.Memory {
		return memoryx.NewSummaryBuffer(svc.LLMs, 50, memoryx.WithCallback(svc.Callbacks),
			memoryx.WithOutParser(memoryOutput))
	})
	return &chat{
		memory: memory,
		svc:    svc,
		router: router.NewRouter(svc.LLMs, handlers,
			router.WithMemory(memory),          // 添加记忆组件支持
			router.Withcallback(svc.Callbacks), // 添加回调处理器
			router.WithEmptyHandler(chatinternal.NewDefaultHandler(svc)),
		), // 添加内存组件支持
	}
}

// PrivateChat 处理私聊消息，将消息保存到数据库
func (l *chat) PrivateChat(ctx context.Context, req *domain.Message) error {
	// 调用通用的聊天日志保存方法
	return l.chatlog(ctx, req)
}

// GroupChat 处理群聊消息，将消息保存到数据库
func (l *chat) GroupChat(ctx context.Context, req *domain.Message) (uids []string, err error) {
	// 保留前端传递的群聊conversationId，不再强制改为"all"
	// 这样每个群聊都有独立的conversationId，可以单独查询和总结
	// 注意：如果前端没有传conversationId，保持为空（会在chatlog中处理）

	// 保存群聊消息到数据库
	if err := l.chatlog(ctx, req); err != nil {
		return nil, err
	}

	// 返回空的用户ID列表（当前实现不需要返回特定用户列表）
	return nil, err
}

// chatlog 通用的聊天消息保存方法，将消息记录到数据库
func (l *chat) chatlog(ctx context.Context, req *domain.Message) error {
	sendId := req.SendId

	// 构建聊天日志数据模型
	chatlog := model.ChatLog{
		ConversationId: req.ConversationId,           // 会话ID
		SendId:         sendId,                       // 发送者ID
		RecvId:         req.RecvId,                   // 接收者ID
		ChatType:       model.ChatType(req.ChatType), // 聊天类型（1=群聊，2=私聊）
		MsgContent:     req.Content,                  // 消息内容
		SendTime:       time.Now().Unix(),            // 发送时间戳
	}

	// 如果没有指定会话ID，则为私聊生成唯一的会话ID
	if chatlog.ConversationId == "" {
		chatlog.ConversationId = GenerateUniqueID(sendId, req.RecvId)
	}

	// 将聊天记录保存到数据库
	return l.svc.ChatLogModel.Insert(ctx, &chatlog)
}

// GenerateUniqueID 根据传递的两个字符串 ID 生成唯一的 ID
func GenerateUniqueID(id1, id2 string) string {
	// 将两个 ID 放入切片中
	ids := []string{id1, id2}

	// 对 IDs 切片进行排序
	sort.Strings(ids)

	// 将排序后的 ID 组合起来
	combined := ids[0] + ids[1]

	// 创建 SHA-256 哈希对象
	hasher := sha256.New()

	// 写入合并后的字符串
	hasher.Write([]byte(combined))

	// 计算哈希值
	hash := hasher.Sum(nil)

	// 返回哈希值的十六进制字符串表示
	return base64.RawStdEncoding.EncodeToString(hash)[:22] // 可以选择更短的长度
}

// AIChat 处理 AI 聊天请求，根据请求类型选择不同的服务处理方法
func (l *chat) AIChat(ctx context.Context, req *domain.ChatReq) (resp *domain.ChatResp, err error) {
	fmt.Println(" ----------------------------- ", token.GetUId(ctx))

	uid := token.GetUId(ctx)
	ctx = context.WithValue(ctx, langchain.ChatId, uid)

	if req.ChatType > 0 {
		return l.basicService(ctx, req)
	}
	return l.aiService(ctx, req)
}

// basicService 处理基础聊天服务请求
func (l *chat) aiService(ctx context.Context, req *domain.ChatReq) (resp *domain.ChatResp, err error) {
	// 将chatlog相关参数通过context传递,避免影响memory的保存逻辑
	ctx = context.WithValue(ctx, "relationId", req.RelationId)
	ctx = context.WithValue(ctx, "startTime", req.StartTime)
	ctx = context.WithValue(ctx, "endTime", req.EndTime)

	v, err := chains.Call(ctx, l.router, map[string]any{
		langchain.Input: req.Prompts,
	}, chains.WithCallback(l.svc.Callbacks))
	if err != nil {
		// 特殊处理：如果是agent输出解析错误，尝试从错误消息中提取有用信息
		if strings.Contains(err.Error(), "unable to parse agent output") {
			// 提取错误消息中的实际内容，通常在冒号后面
			parts := strings.Split(err.Error(), ": ")
			if len(parts) > 1 {
				content := parts[len(parts)-1]
				// 返回提取到的内容
				return &domain.ChatResp{
					ChatType: domain.DefaultHandler,
					Data:     content,
				}, nil
			}
		}
		return nil, err
	}

	// 检查输出类型并处理非字符串输出
	var data string
	if _, ok := v[langchain.OutPut].(string); !ok {
		// 如果输出不是字符串，直接返回原始数据
		return &domain.ChatResp{
			ChatType: domain.DefaultHandler,
			Data:     v,
		}, nil
	}
	data = v[langchain.OutPut].(string)

	// 尝试解析AI输出为结构化响应格式: {"chatType": "", data: ""}
	var res domain.ChatResp
	if err := json.Unmarshal([]byte(data), &res); err != nil {
		// 解析失败时返回默认处理器类型和原始数据
		return &domain.ChatResp{
			ChatType: domain.DefaultHandler,
			Data:     data,
		}, nil
	}

	return &res, nil
}

func (l *chat) basicService(ctx context.Context, req *domain.ChatReq) (resp *domain.ChatResp, err error) {
	return nil, err
}

func memoryOutput(ctx context.Context, v string) string {
	var res domain.ChatResp
	if err := json.Unmarshal([]byte(v), &res); err != nil {
		tlog.ErrorfCtx(ctx, "memoryOutput", "v %s, err %s", v, err.Error())
		return v
	}

	tlog.InfoCtx(ctx, "memoryOutput", v)

	switch res.Data.(type) {
	case string:
		return res.Data.(string)
	default:
		return v
	}
}

// File 将上传的文件信息保存到记忆机制中，使AI能够记住用户上传的文件
func (l *chat) File(ctx context.Context, files []*domain.FileResp) (err error) {
	uid := token.GetUId(ctx)                            // 获取当前用户ID
	ctx = context.WithValue(ctx, langchain.ChatId, uid) // 将用户ID添加到上下文中
	fmt.Println("File --- ", uid)
	data := make([]*domain.ChatFile, 0, len(files))
	for _, file := range files {
		data = append(data, &domain.ChatFile{ // 构建文件信息结构
			Path: file.File,
			Name: file.Filename,
			Time: time.Now(),
		})
	}

	b, err := json.Marshal(data) // 将文件信息序列化为JSON
	if err != nil {
		return err
	}

	err = l.memory.SaveContext(ctx, map[string]any{ // 将文件信息保存到记忆机制中
		langchain.OutPut: string(b),
	}, map[string]any{
		langchain.OutPut: "uploaded files", // AI的回复内容
	})

	// 调试功能：打印当前记忆内容
	memoryContent, err := l.memory.LoadMemoryVariables(ctx, map[string]any{})
	if err != nil {
		return err
	}
	fmt.Println(memoryContent)

	return
}
