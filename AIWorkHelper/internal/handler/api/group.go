/**
 * @author: 公众号:IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package api 提供群聊管理的HTTP API处理器
package api

import (
	"AIWorkHelper/internal/domain"
	"AIWorkHelper/internal/logic"
	"AIWorkHelper/internal/svc"
	"AIWorkHelper/pkg/httpx"
	"AIWorkHelper/pkg/token"
	"errors"

	"github.com/gin-gonic/gin"
)

// Group 群聊管理API处理器
// 对应Java版本: GroupController类
type Group struct {
	svcCtx       *svc.ServiceContext
	groupService logic.GroupService
}

// NewGroup 创建群聊管理API处理器实例
func NewGroup(svcCtx *svc.ServiceContext) *Group {
	return &Group{
		svcCtx:       svcCtx,
		groupService: logic.NewGroupService(svcCtx),
	}
}

// InitRegister 注册路由
func (h *Group) InitRegister(engine *gin.Engine) {
	g := engine.Group("v1/group", h.svcCtx.Jwt.Handler)
	g.POST("/create", h.CreateGroup)                                // POST /v1/group/create
	g.GET("/:groupId/members", h.GetGroupMembers)                   // GET /v1/group/{groupId}/members
	g.POST("/members/add", h.AddMembers)                            // POST /v1/group/members/add
	g.DELETE("/:groupId/members/:userId", h.RemoveMember)           // DELETE /v1/group/{groupId}/members/{userId}
	g.GET("/:groupId/members/:userId/exists", h.IsMember)           // GET /v1/group/{groupId}/members/{userId}/exists
	g.GET("/:groupId/count", h.GetMemberCount)                      // GET /v1/group/{groupId}/count
}

// CreateGroup 创建群聊
// POST /v1/group/create
// 对应Java版本: createGroup方法
func (h *Group) CreateGroup(ctx *gin.Context) {
	var req domain.CreateGroupReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	// 从JWT令牌中获取当前用户ID
	userId := token.GetUId(ctx.Request.Context())
	if userId == "" {
		httpx.FailWithErr(ctx, errors.New("用户未登录"))
		return
	}

	// 将创建者也添加到群成员中
	if err := h.groupService.AddMember(ctx.Request.Context(), req.GroupId, userId); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	// 添加其他成员
	if err := h.groupService.AddMembers(ctx.Request.Context(), req.GroupId, req.MemberIds); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.Ok(ctx)
}

// GetGroupMembers 获取群成员列表
// GET /v1/group/{groupId}/members
// 对应Java版本: getGroupMembers方法
func (h *Group) GetGroupMembers(ctx *gin.Context) {
	var pathReq domain.GroupPathReq
	if err := ctx.ShouldBindUri(&pathReq); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	memberIds, err := h.groupService.GetGroupMemberIds(ctx.Request.Context(), pathReq.GroupId)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.OkWithData(ctx, memberIds)
}

// AddMembers 添加群成员
// POST /v1/group/members/add
// 对应Java版本: addMembers方法
func (h *Group) AddMembers(ctx *gin.Context) {
	var req domain.AddGroupMembersReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	if err := h.groupService.AddMembers(ctx.Request.Context(), req.GroupId, req.MemberIds); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.Ok(ctx)
}

// RemoveMember 移除群成员
// DELETE /v1/group/{groupId}/members/{userId}
// 对应Java版本: removeMember方法
func (h *Group) RemoveMember(ctx *gin.Context) {
	var pathReq domain.GroupPathReq
	if err := ctx.ShouldBindUri(&pathReq); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	if err := h.groupService.RemoveMember(ctx.Request.Context(), pathReq.GroupId, pathReq.UserId); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.Ok(ctx)
}

// IsMember 检查用户是否在群中
// GET /v1/group/{groupId}/members/{userId}/exists
// 对应Java版本: isMember方法
func (h *Group) IsMember(ctx *gin.Context) {
	var pathReq domain.GroupPathReq
	if err := ctx.ShouldBindUri(&pathReq); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	isMember, err := h.groupService.IsMember(ctx.Request.Context(), pathReq.GroupId, pathReq.UserId)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.OkWithData(ctx, isMember)
}

// GetMemberCount 获取群成员数量
// GET /v1/group/{groupId}/count
// 对应Java版本: getMemberCount方法
func (h *Group) GetMemberCount(ctx *gin.Context) {
	var pathReq domain.GroupPathReq
	if err := ctx.ShouldBindUri(&pathReq); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	count, err := h.groupService.GetMemberCount(ctx.Request.Context(), pathReq.GroupId)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.OkWithData(ctx, count)
}
