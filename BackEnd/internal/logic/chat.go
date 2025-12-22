package logic

import (
	"BackEnd/internal/domain"
	"BackEnd/internal/logic/chatinternal"
	"BackEnd/internal/model"
	"BackEnd/internal/svc"
	"BackEnd/pkg/langchain"
	"BackEnd/pkg/langchain/router"
	"BackEnd/pkg/token"
	"BackEnd/pkg/util"
	"BackEnd/pkg/xerr"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/schema"
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
	// ListConversations 查询会话列表
	ListConversations(ctx context.Context, req *domain.ConversationListReq) (resp *domain.ConversationListResp, err error)
}

type chat struct {
	svcCtx   *svc.ServiceContext
	baseChat *chatinternal.BaseChat
	router   *router.Router // 智能路由器，用于选择合适的处理器
	memory   schema.Memory  // 多会话内存管理器，支持对话历史记忆
}

func NewChat(svcCtx *svc.ServiceContext) Chat {
	var baseChat *chatinternal.BaseChat
	var err error

	if svcCtx.LLMs != nil {
		baseChat = chatinternal.NewBaseChatFromLLM(svcCtx.LLMs)
	} else {
		baseChat, err = chatinternal.NewBaseChat(
			svcCtx.Config.AI.ApiKey,
			svcCtx.Config.AI.BaseURL,
			svcCtx.Config.AI.Model,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to initialize AI chat")
		}
	}

	var r *router.Router
	if baseChat != nil {
		todoHandle := chatinternal.NewTodoHandle(svcCtx)
		// Inject the logic implementation required by the tool
		deptLogic := NewDepartment(svcCtx)
		departmentHandle := chatinternal.NewDepartmentHandle(svcCtx, deptLogic)

		approvalLogic := NewApproval(svcCtx)
		approvalHandle := chatinternal.NewApprovalHandle(svcCtx, approvalLogic)

		// Prepare tool descriptions for the general chat handler so it knows about other tools
		otherHandlers := []router.Handler{
			todoHandle,
			departmentHandle,
			approvalHandle,
		}
		toolDescriptions := router.HandlerDestinations(otherHandlers)

		chatHandle := chatinternal.NewChatHandle(svcCtx, toolDescriptions)
		r = router.NewRouter(baseChat.GetLLM(), []router.Handler{
			todoHandle,
			departmentHandle,
			approvalHandle,
			chatHandle,
		}, router.WithEmptyHandler(chatHandle))
	}

	return &chat{
		svcCtx:   svcCtx,
		baseChat: baseChat,
		router:   r,
	}
}

// AIChat AI聊天接口
func (l *chat) AIChat(ctx context.Context, req *domain.ChatReq) (resp *domain.ChatResp, err error) {
	userID, err := token.GetUserID(ctx)
	if err != nil {
		return nil, xerr.New(err)
	}
	uidStr := util.UintToString(userID)

	ctx = context.WithValue(ctx, langchain.ChatID, uidStr)
	fmt.Printf("Token in AIChat: %s\n", token.GetTokenStr(ctx))

	return l.aiService(ctx, req)
}

// basicService 处理基础聊天服务请求
func (l *chat) aiService(ctx context.Context, req *domain.ChatReq) (resp *domain.ChatResp, err error) {
	if l.router == nil {
		return nil, errors.New("AI service not initialized")
	}
	v, err := chains.Call(ctx, l.router, map[string]any{
		langchain.Input: req.Prompts,
	}, chains.WithCallback(l.svcCtx.Callbacks))
	if err != nil {
		return nil, err
	}

	data := v[langchain.Output].(string)

	// AI : {"chatType": "", "data": ""}
	var res domain.ChatResp
	if err := json.Unmarshal([]byte(data), &res); err != nil {
		// If unmarshal fails, treat as plain text
		return &domain.ChatResp{
			ChatType: 0,
			Data:     data,
		}, nil
	}

	return &res, nil
}

func (l *chat) basicService(ctx context.Context, req *domain.ChatReq) (resp *domain.ChatResp, err error) {
	return nil, err
}

// File 处理文件上传，将文件信息保存到记忆机制中（可选实现）
// 当前实现为占位符，后续可以集成 AI 记忆功能
func (l *chat) File(ctx context.Context, files []*domain.FileResp) (err error) {
	if l.baseChat == nil {
		return nil
	}

	// 构造文件信息
	fileData := make([]map[string]any, 0, len(files))
	for _, f := range files {
		fileData = append(fileData, map[string]any{
			"filename": f.Filename,
			"path":     f.File,
			"host":     f.Host,
		})
	}

	// 保存到 AI 记忆上下文
	// 这里简单地将文件列表作为 context 保存
	// 实际应用中可能需要提取文件内容
	err = l.baseChat.SaveContext(ctx, map[string]any{
		"uploaded_files": fileData,
	}, map[string]any{
		"output": "Files uploaded and context updated",
	})

	if err != nil {
		log.Error().Err(err).Msg("保存文件上下文失败")
		return err
	}

	log.Info().Int("count", len(files)).Msg("文件上下文已保存到 AI 记忆")
	return nil
}

func (l *chat) PrivateChat(ctx context.Context, req *domain.Message) error {
	// 调用通用的聊天日志保存方法
	return l.chatlog(ctx, req)
}

func (l *chat) GroupChat(ctx context.Context, req *domain.Message) (uids []string, err error) {
	if req.ConversationId == "" {
		req.ConversationId = "all"
	}
	req.RecvId = ""

	if err := l.chatlog(ctx, req); err != nil {
		return nil, err
	}

	return nil, nil
}

func (l *chat) chatlog(ctx context.Context, req *domain.Message) error {
	sendId := req.SendId

	// 确保会话存在
	var err error
	if req.ConversationId == "" {
		if req.ChatType == int(model.SingleChatType) && req.RecvId != "" {
			req.ConversationId, err = l.getOrCreatePrivateConversation(ctx, sendId, req.RecvId)
		} else if req.ChatType == int(model.GroupChatType) {
			if req.ConversationId == "" {
				return errors.New("群聊必须指定ConversationId")
			}
			req.ConversationId, err = l.getOrCreateGroupConversation(ctx, req.ConversationId, sendId)
		} else if req.ChatType == int(model.AIChatType) {
			req.ConversationId, err = l.getOrCreateAIConversation(ctx, sendId)
		}
		if err != nil {
			return err
		}
	} else {
		// 即使传了ID，也要确保数据库里有这个会话（针对群聊等）
		if req.ChatType == int(model.GroupChatType) {
			_, err = l.getOrCreateGroupConversation(ctx, req.ConversationId, sendId)
		} else if req.ChatType == int(model.AIChatType) {
			_, err = l.getOrCreateAIConversation(ctx, sendId)
		} else if req.ChatType == int(model.SingleChatType) && req.RecvId != "" {
			// 校验私聊ID是否正确
			expectedId := generateUniqueID(sendId, req.RecvId)
			if req.ConversationId != expectedId {
				// 如果前端传的不对，纠正它
				req.ConversationId = expectedId
			}
			_, err = l.getOrCreatePrivateConversation(ctx, sendId, req.RecvId)
		}
		if err != nil {
			return err
		}
	}

	chatlog := model.ChatLog{
		ConversationId: req.ConversationId,                // 会话ID
		SendId:         util.StringToUintSafe(sendId),     // 发送者ID
		RecvId:         util.StringToUintSafe(req.RecvId), // 接收者ID
		ChatType:       model.ChatType(req.ChatType),      // 聊天类型（1=群聊，2=私聊）
		MsgContent:     req.Content,                       // 消息内容
		SendTime:       time.Now().Unix(),                 // 发送时间戳
	}

	if err := l.svcCtx.DB.WithContext(ctx).Create(&chatlog).Error; err != nil {
		log.Error().Err(err).
			Str("conversation_id", chatlog.ConversationId).
			Msg("failed to create chat log record")
		return xerr.New(err)
	}

	// 更新会话的最后一条消息
	l.svcCtx.DB.WithContext(ctx).Model(&model.Conversation{}).
		Where("id = ?", req.ConversationId).
		Updates(map[string]interface{}{
			"last_message_id":   chatlog.ID,
			"last_message_time": chatlog.SendTime,
			"update_at":         time.Now().Unix(),
		})

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
		// 如果没有会话ID，尝试根据 targetUserId 和 chatType 生成
		if req.ChatType == int(model.SingleChatType) && req.TargetUserId != "" {
			// 私聊：生成唯一会话ID
			req.ConversationId = generateUniqueID(util.UintToString(userID), req.TargetUserId)
		} else if req.ChatType == int(model.AIChatType) {
			// AI聊：生成AI会话ID
			req.ConversationId = "ai_" + util.UintToString(userID)
		} else {
			return nil, xerr.New(errors.New("会话ID不能为空"))
		}
	}

	// 3. 构建查询条件
	query := l.svcCtx.DB.WithContext(ctx).Model(&model.ChatLog{})

	// 4. 会话ID过滤
	query = query.Where("conversation_id = ?", req.ConversationId)

	// 5. 权限检查：只能查询自己参与的消息
	// 私聊：发送者或接收者是自己
	// 群聊：所有消息都可以查询（因为群聊的 conversationId 是共享的）
	// AI聊：发送者或接收者是自己（AI ID为0，用户ID为当前用户）
	query = query.Where("(send_id = ? OR recv_id = ? OR chat_type = ? OR chat_type = ?)", userID, userID, model.GroupChatType, model.AIChatType)

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

// ListConversations 查询会话列表
func (l *chat) ListConversations(ctx context.Context, req *domain.ConversationListReq) (resp *domain.ConversationListResp, err error) {
	userID, err := token.GetUserID(ctx)
	if err != nil {
		return nil, xerr.New(err)
	}
	uidStr := util.UintToString(userID)

	// 1. 查询用户参与的所有会话ID
	var participants []model.Participant
	if err := l.svcCtx.DB.WithContext(ctx).
		Where("user_id = ?", uidStr).
		Find(&participants).Error; err != nil {
		return nil, xerr.New(err)
	}

	if len(participants) == 0 {
		return &domain.ConversationListResp{
			List:  []*domain.Conversation{},
			Total: 0,
		}, nil
	}

	conversationIds := make([]string, len(participants))
	for i, p := range participants {
		conversationIds[i] = p.ConversationId
	}

	// 2. 查询会话详情
	var conversations []model.Conversation
	query := l.svcCtx.DB.WithContext(ctx).Model(&model.Conversation{}).
		Where("id IN ?", conversationIds)

	// 分页
	pagination := util.NormalizePagination(req.Page, req.Count)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, xerr.New(err)
	}

	if err := query.Order("last_message_time DESC").
		Offset(pagination.Offset).
		Limit(pagination.Count).
		Find(&conversations).Error; err != nil {
		return nil, xerr.New(err)
	}

	// 3. 转换为响应格式
	list := make([]*domain.Conversation, 0, len(conversations))
	for _, c := range conversations {
		// 对于私聊，名称显示为对方名称，头像显示为对方头像
		name := c.Name
		avatar := ""
		if c.Type == int(model.SingleChatType) {
			// 查找对方ID
			var otherParticipant model.Participant
			l.svcCtx.DB.WithContext(ctx).
				Where("conversation_id = ? AND user_id != ?", c.Id, uidStr).
				First(&otherParticipant)

			if otherParticipant.UserId != "" {
				// 查询对方用户信息
				var otherUser model.User
				if err := l.svcCtx.DB.WithContext(ctx).
					Where("id = ?", otherParticipant.UserId).
					First(&otherUser).Error; err == nil {
					name = otherUser.Name
					// avatar = otherUser.Avatar // 假设User表有Avatar字段
				}
			}
		}

		// 获取最后一条消息内容
		var lastMsg model.ChatLog
		if c.LastMessageId > 0 {
			l.svcCtx.DB.WithContext(ctx).First(&lastMsg, c.LastMessageId)
		}

		// 获取成员ID列表
		var memberIds []string
		var members []model.Participant
		if err := l.svcCtx.DB.WithContext(ctx).
			Where("conversation_id = ?", c.Id).
			Find(&members).Error; err == nil {
			for _, m := range members {
				memberIds = append(memberIds, m.UserId)
			}
		}

		list = append(list, &domain.Conversation{
			Id:              c.Id,
			Type:            c.Type,
			Name:            name,
			LastMessage:     lastMsg.MsgContent,
			LastMessageTime: c.LastMessageTime,
			UnreadCount:     0, // TODO: 实现未读计数
			Avatar:          avatar,
			MemberIds:       memberIds,
		})
	}

	return &domain.ConversationListResp{
		List:  list,
		Total: total,
	}, nil
}

// getOrCreatePrivateConversation 获取或创建私聊会话
func (l *chat) getOrCreatePrivateConversation(ctx context.Context, uid1, uid2 string) (string, error) {
	conversationId := generateUniqueID(uid1, uid2)

	// 检查会话是否存在
	var count int64
	if err := l.svcCtx.DB.WithContext(ctx).Model(&model.Conversation{}).
		Where("id = ?", conversationId).
		Count(&count).Error; err != nil {
		return "", err
	}

	if count == 0 {
		// 创建会话
		conversation := model.Conversation{
			Id:        conversationId,
			Type:      int(model.SingleChatType),
			Name:      "私聊", // 私聊名称通常在展示时动态生成
			CreatorId: uid1,
			CreateAt:  time.Now().Unix(),
			UpdateAt:  time.Now().Unix(),
		}
		if err := l.svcCtx.DB.WithContext(ctx).Create(&conversation).Error; err != nil {
			return "", err
		}

		// 创建参与者
		participants := []model.Participant{
			{ConversationId: conversationId, UserId: uid1, Role: 0, JoinTime: time.Now().Unix()},
			{ConversationId: conversationId, UserId: uid2, Role: 0, JoinTime: time.Now().Unix()},
		}
		if err := l.svcCtx.DB.WithContext(ctx).Create(&participants).Error; err != nil {
			return "", err
		}
	}

	return conversationId, nil
}

// getOrCreateGroupConversation 获取或创建群聊会话
func (l *chat) getOrCreateGroupConversation(ctx context.Context, groupId string, creatorId string) (string, error) {
	// 检查会话是否存在
	var count int64
	if err := l.svcCtx.DB.WithContext(ctx).Model(&model.Conversation{}).
		Where("id = ?", groupId).
		Count(&count).Error; err != nil {
		return "", err
	}

	if count == 0 {
		// 创建会话
		conversation := model.Conversation{
			Id:        groupId,
			Type:      int(model.GroupChatType),
			Name:      "群聊", // 初始名称，后续可以修改
			CreatorId: creatorId,
			CreateAt:  time.Now().Unix(),
			UpdateAt:  time.Now().Unix(),
		}
		if err := l.svcCtx.DB.WithContext(ctx).Create(&conversation).Error; err != nil {
			return "", err
		}

		// 注意：群聊参与者的添加通常在创建群聊或邀请成员时处理，这里仅确保会话存在
		// 如果是首次创建，至少添加创建者
		if creatorId != "" {
			participant := model.Participant{
				ConversationId: groupId,
				UserId:         creatorId,
				Role:           2, // 群主
				JoinTime:       time.Now().Unix(),
			}
			if err := l.svcCtx.DB.WithContext(ctx).Create(&participant).Error; err != nil {
				return "", err
			}
		}
	}

	return groupId, nil
}

// getOrCreateAIConversation 获取或创建AI会话
func (l *chat) getOrCreateAIConversation(ctx context.Context, userId string) (string, error) {
	conversationId := "ai_" + userId

	var count int64
	if err := l.svcCtx.DB.WithContext(ctx).Model(&model.Conversation{}).
		Where("id = ?", conversationId).
		Count(&count).Error; err != nil {
		return "", err
	}

	if count == 0 {
		conversation := model.Conversation{
			Id:        conversationId,
			Type:      int(model.AIChatType),
			Name:      "AI助手",
			CreatorId: userId,
			CreateAt:  time.Now().Unix(),
			UpdateAt:  time.Now().Unix(),
		}
		if err := l.svcCtx.DB.WithContext(ctx).Create(&conversation).Error; err != nil {
			return "", err
		}

		participant := model.Participant{
			ConversationId: conversationId,
			UserId:         userId,
			Role:           0,
			JoinTime:       time.Now().Unix(),
		}
		if err := l.svcCtx.DB.WithContext(ctx).Create(&participant).Error; err != nil {
			return "", err
		}
	}

	return conversationId, nil
}
