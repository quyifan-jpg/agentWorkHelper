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
	// AIChat AI聊天接口（与 AIWorkHelper 保持一致）
	AIChat(ctx context.Context, req *domain.ChatReq) (resp *domain.ChatResp, err error)
	// PrivateChat 处理私聊消息（通过 WebSocket 调用）
	PrivateChat(ctx context.Context, req *domain.Message) error
	// GroupChat 处理群聊消息（通过 WebSocket 调用）
	GroupChat(ctx context.Context, req *domain.Message) (uids []string, err error)
	// File 处理文件上传（可选实现）
	File(ctx context.Context, files []*domain.FileResp) (err error)
	// ListMessages 查询历史消息列表
	ListMessages(ctx context.Context, req *domain.ChatMessageListReq) (resp *domain.ChatMessageListResp, err error)
}

type chat struct {
	svcCtx *svc.ServiceContext
}

func NewChat(svcCtx *svc.ServiceContext) Chat {
	return &chat{
		svcCtx: svcCtx,
	}
}

// AIChat AI聊天接口（暂时返回简单响应，后续可集成 AI 功能）
func (l *chat) AIChat(ctx context.Context, req *domain.ChatReq) (resp *domain.ChatResp, err error) {
	// TODO: 实现 AI 聊天逻辑
	// 当前返回简单响应，后续可以集成 LangChain 等 AI 框架
	return &domain.ChatResp{
		ChatType: 0,
		Data:     "AI 功能暂未实现，请稍后",
	}, nil
}

// File 处理文件上传，将文件信息保存到记忆机制中（可选实现）
// 当前实现为占位符，后续可以集成 AI 记忆功能
func (l *chat) File(ctx context.Context, files []*domain.FileResp) (err error) {
	// TODO: 实现文件信息保存到 AI 记忆机制
	// 当前实现为占位符，后续可以：
	// 1. 将文件信息保存到数据库
	// 2. 集成 LangChain 等 AI 框架的记忆机制
	// 3. 将文件内容提取并保存到向量数据库
	
	// 当前仅记录日志
	log.Info().
		Int("file_count", len(files)).
		Msg("文件上传成功，文件信息已记录")
	
	return nil
}

func (l *chat) PrivateChat(ctx context.Context, req *domain.Message) error {
	// 调用通用的聊天日志保存方法
	return l.chatlog(ctx, req)
}

func (l *chat) GroupChat(ctx context.Context, req *domain.Message) (uids []string, err error) {
	// 群聊：设置会话ID为 "all"（如果没有指定）
	if req.ConversationId == "" {
		req.ConversationId = "all"
	}
	// 群聊时 RecvId 为空
	req.RecvId = ""

	if err := l.chatlog(ctx, req); err != nil {
		return nil, err
	}

	return nil, nil
}

func (l *chat) chatlog(ctx context.Context, req *domain.Message) error {
	sendId := req.SendId
	chatlog := model.ChatLog{
		ConversationId: req.ConversationId,           // 会话ID
		SendId:         util.StringToUintSafe(sendId),                       // 发送者ID
		RecvId:         util.StringToUintSafe(req.RecvId),                   // 接收者ID
		ChatType:       model.ChatType(req.ChatType), // 聊天类型（1=群聊，2=私聊）
		MsgContent:     req.Content,                  // 消息内容
		SendTime:       time.Now().Unix(),            // 发送时间戳
	}
	// 如果没有指定会话ID，则生成
	if chatlog.ConversationId == "" {
		if req.ChatType == 2 && req.RecvId != "" {
			// 私聊：生成唯一会话ID
			chatlog.ConversationId = generateUniqueID(sendId, req.RecvId)
		} else if req.ChatType == 1 {
			// 群聊：使用 "all"
			chatlog.ConversationId = "all"
		}
	}
	if err := l.svcCtx.DB.WithContext(ctx).Create(&chatlog).Error; err != nil {
        log.Error().Err(err).
            Str("conversation_id", chatlog.ConversationId).
            Msg("failed to create chat log record") 
        return xerr.New(err)
    } 
    return nil
}


func generateUniqueID(id1, id2 string) string {
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

// ListMessages 查询历史消息列表
func (l *chat) ListMessages(ctx context.Context, req *domain.ChatMessageListReq) (resp *domain.ChatMessageListResp, err error) {
	// 1. 获取当前用户ID
	userID, err := token.GetUserID(ctx)
	if err != nil {
		return nil, xerr.New(err)
	}

	// 2. 验证必填参数
	if req.ConversationId == "" {
		return nil, xerr.New(errors.New("会话ID不能为空"))
	}

	// 3. 构建查询条件
	query := l.svcCtx.DB.WithContext(ctx).Model(&model.ChatLog{})

	// 4. 会话ID过滤
	query = query.Where("conversation_id = ?", req.ConversationId)

	// 5. 权限检查：只能查询自己参与的消息
	// 私聊：发送者或接收者是自己
	// 群聊：所有消息都可以查询（因为群聊的 conversationId 是共享的）
	query = query.Where("(send_id = ? OR recv_id = ? OR chat_type = ?)", userID, userID, model.GroupChatType)

	// 6. 时间范围过滤
	if req.StartTime > 0 {
		query = query.Where("send_time >= ?", req.StartTime)
	}
	if req.EndTime > 0 {
		query = query.Where("send_time <= ?", req.EndTime)
	}

	// 7. 分页处理
	pagination := util.NormalizePagination(req.Page, req.Count)

	// 8. 查询总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		log.Error().Err(err).Msg("查询消息总数失败")
		return nil, xerr.New(err)
	}

	// 9. 查询消息列表（按时间倒序，最新的在前）
	var chatLogs []model.ChatLog
	if err := query.Order("send_time DESC").
		Offset(pagination.Offset).
		Limit(pagination.Count).
		Find(&chatLogs).Error; err != nil {
		log.Error().Err(err).Msg("查询消息列表失败")
		return nil, xerr.New(err)
	}

	// 10. 获取所有发送者ID，批量查询用户信息
	sendIds := make([]uint, 0)
	sendIdMap := make(map[uint]bool)
	for _, log := range chatLogs {
		if !sendIdMap[log.SendId] {
			sendIds = append(sendIds, log.SendId)
			sendIdMap[log.SendId] = true
		}
	}

	// 11. 批量查询用户信息
	userMap := make(map[uint]string) // 用户ID -> 用户名
	if len(sendIds) > 0 {
		var users []model.User
		if err := l.svcCtx.DB.WithContext(ctx).
			Where("id IN ?", sendIds).
			Find(&users).Error; err == nil {
			for _, user := range users {
				userMap[user.ID] = user.Name
			}
		}
	}

	// 12. 转换为响应格式
	messages := make([]*domain.ChatMessage, 0, len(chatLogs))
	for _, log := range chatLogs {
		sendName := userMap[log.SendId]
		if sendName == "" {
			sendName = "未知用户"
		}

		messages = append(messages, &domain.ChatMessage{
			Id:          log.ID,
			SendId:      util.UintToString(log.SendId),
			SendName:    sendName,
			Content:     log.MsgContent,
			ContentType: 1, // 默认文字类型，后续可以扩展
			SendTime:    log.SendTime,
			ChatType:    int(log.ChatType),
		})
	}

	// 13. 返回结果（按时间正序，最旧的在前，方便前端显示）
	// 反转列表，让最旧的消息在前
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return &domain.ChatMessageListResp{
		List:  messages,
		Total: total,
		Page:  pagination.Page,
		Count: pagination.Count,
	}, nil
}
