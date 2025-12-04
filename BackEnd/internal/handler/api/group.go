package api

import (
	"BackEnd/internal/domain"
	"BackEnd/internal/logic"
	"BackEnd/internal/svc"
	"BackEnd/pkg/httpx"
	"BackEnd/pkg/token"
	"BackEnd/pkg/util"

	"github.com/gin-gonic/gin"
)

type Group struct {
	svcCtx *svc.ServiceContext
	group  logic.Group
}

func NewGroup(svcCtx *svc.ServiceContext, group logic.Group) *Group {
	return &Group{
		svcCtx: svcCtx,
		group:  group,
	}
}

func (h *Group) InitRegister(engine *gin.Engine) {
	g := engine.Group("v1/group", h.svcCtx.Jwt.Handler)
	g.POST("/create", h.CreateGroup)
	g.GET("/:groupId/members", h.GetGroupMembers)
	g.POST("/members/add", h.AddMembers)
	g.DELETE("/:groupId/members/:userId", h.RemoveMember)
	g.GET("/:groupId/members/:userId/exists", h.IsMember)
	g.GET("/:groupId/count", h.GetMemberCount)
}

// CreateGroup 创建群聊
// @Summary 创建群聊
// @Description 创建群聊并添加成员（包括创建者）
// @Tags group
// @Accept json
// @Produce json
// @Param req body domain.CreateGroupReq true "创建群聊请求"
// @Success 200 {object} object{code=int,msg=string,data=domain.IdResp}
// @Router /v1/group/create [post]
func (h *Group) CreateGroup(ctx *gin.Context) {
	var req domain.CreateGroupReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	// 从JWT令牌中获取当前用户ID
	userID, err := token.GetUserID(ctx.Request.Context())
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	// 将创建者也添加到群成员中
	userIDStr := util.UintToString(userID)
	if err := h.group.AddMember(ctx.Request.Context(), req.GroupId, userIDStr); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	// 添加其他成员
	if err := h.group.AddMembers(ctx.Request.Context(), req.GroupId, req.MemberIds); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.Success(ctx, domain.IdResp{Id: req.GroupId})
}

// GetGroupMembers 获取群成员列表
// @Summary 获取群成员列表
// @Description 根据群ID获取群成员ID列表
// @Tags group
// @Accept json
// @Produce json
// @Param groupId path string true "群ID"
// @Success 200 {object} object{code=int,msg=string,data=[]string}
// @Router /v1/group/{groupId}/members [get]
func (h *Group) GetGroupMembers(ctx *gin.Context) {
	var req domain.GroupPathReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	memberIds, err := h.group.GetGroupMemberIds(ctx.Request.Context(), req.GroupId)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.Success(ctx, memberIds)
}

// AddMembers 添加群成员
// @Summary 添加群成员
// @Description 批量添加群成员
// @Tags group
// @Accept json
// @Produce json
// @Param req body domain.AddGroupMembersReq true "添加群成员请求"
// @Success 200 {object} object{code=int,msg=string}
// @Router /v1/group/members/add [post]
func (h *Group) AddMembers(ctx *gin.Context) {
	var req domain.AddGroupMembersReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	if err := h.group.AddMembers(ctx.Request.Context(), req.GroupId, req.MemberIds); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

// RemoveMember 移除群成员
// @Summary 移除群成员
// @Description 从群中移除指定成员
// @Tags group
// @Accept json
// @Produce json
// @Param groupId path string true "群ID"
// @Param userId path string true "用户ID"
// @Success 200 {object} object{code=int,msg=string}
// @Router /v1/group/{groupId}/members/{userId} [delete]
func (h *Group) RemoveMember(ctx *gin.Context) {
	var req domain.GroupPathReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	if err := h.group.RemoveMember(ctx.Request.Context(), req.GroupId, req.UserId); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

// IsMember 检查用户是否在群中
// @Summary 检查用户是否在群中
// @Description 检查指定用户是否是群成员
// @Tags group
// @Accept json
// @Produce json
// @Param groupId path string true "群ID"
// @Param userId path string true "用户ID"
// @Success 200 {object} object{code=int,msg=string,data=domain.IsMemberResp}
// @Router /v1/group/{groupId}/members/{userId}/exists [get]
func (h *Group) IsMember(ctx *gin.Context) {
	var req domain.GroupPathReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	isMember, err := h.group.IsMember(ctx.Request.Context(), req.GroupId, req.UserId)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.Success(ctx, domain.IsMemberResp{IsMember: isMember})
}

// GetMemberCount 获取群成员数量
// @Summary 获取群成员数量
// @Description 获取指定群的成员数量
// @Tags group
// @Accept json
// @Produce json
// @Param groupId path string true "群ID"
// @Success 200 {object} object{code=int,msg=string,data=domain.GroupMemberCountResp}
// @Router /v1/group/{groupId}/count [get]
func (h *Group) GetMemberCount(ctx *gin.Context) {
	var req domain.GroupPathReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	count, err := h.group.GetMemberCount(ctx.Request.Context(), req.GroupId)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.Success(ctx, domain.GroupMemberCountResp{Count: count})
}

