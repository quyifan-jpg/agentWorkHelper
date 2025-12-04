package api

import (
	"github.com/gin-gonic/gin"

	"BackEnd/internal/domain"
	"BackEnd/internal/logic"
	"BackEnd/internal/svc"
	"BackEnd/pkg/httpx"
)

type Chat struct {
	svcCtx *svc.ServiceContext
	chat   logic.Chat
}

func NewChat(svcCtx *svc.ServiceContext, chat logic.Chat) *Chat {
	return &Chat{
		svcCtx: svcCtx,
		chat:   chat,
	}
}

func (h *Chat) InitRegister(engine *gin.Engine) {
	g := engine.Group("v1/chat", h.svcCtx.Jwt.Handler)
	g.POST("", h.Chat)                           // POST /v1/chat - AI聊天接口
	g.GET("/messages", h.ListMessages)           // GET /v1/chat/messages - 查询历史消息
	g.GET("/conversations", h.ListConversations) // GET /v1/chat/conversations - 查询会话列表
}

// Chat AI聊天接口
// @Summary AI聊天
// @Description AI智能对话接口，支持待办查询、审批查询、群消息总结等功能
// @Tags chat
// @Accept json
// @Produce json
// @Param req body domain.ChatReq true "AI聊天请求"
// @Success 200 {object} object{code=int,msg=string,data=domain.ChatResp}
// @Router /v1/chat [post]
func (h *Chat) Chat(ctx *gin.Context) {
	var req domain.ChatReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	res, err := h.chat.AIChat(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.Success(ctx, res)
}

// ListMessages 查询历史消息列表
// @Summary 查询历史消息
// @Description 根据会话ID查询历史消息列表，支持分页和时间范围过滤
// @Tags chat
// @Accept json
// @Produce json
// @Param conversationId query string true "会话ID"
// @Param page query int false "页码，默认1"
// @Param count query int false "每页数量，默认20"
// @Param startTime query int64 false "开始时间戳（可选）"
// @Param endTime query int64 false "结束时间戳（可选）"
// @Success 200 {object} object{code=int,msg=string,data=domain.ChatMessageListResp}
// @Router /v1/chat/messages [get]
func (h *Chat) ListMessages(ctx *gin.Context) {
	var req domain.ChatMessageListReq
	// var req domain.ChatMessageListReq
	// GET请求使用Query参数绑定
	if err := ctx.ShouldBindQuery(&req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	resp, err := h.chat.ListMessages(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.Success(ctx, resp)
}

// ListConversations 查询会话列表
// @Summary 查询会话列表
// @Description 查询当前用户的会话列表
// @Tags chat
// @Accept json
// @Produce json
// @Param page query int false "页码，默认1"
// @Param count query int false "每页数量，默认20"
// @Success 200 {object} object{code=int,msg=string,data=domain.ConversationListResp}
// @Router /v1/chat/conversations [get]
func (h *Chat) ListConversations(ctx *gin.Context) {
	var req domain.ConversationListReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	resp, err := h.chat.ListConversations(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.Success(ctx, resp)
}
