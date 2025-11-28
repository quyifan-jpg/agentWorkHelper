/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package api

import (
	"github.com/gin-gonic/gin"

	"AIWorkHelper/internal/domain"
	"AIWorkHelper/internal/logic"
	"AIWorkHelper/internal/svc"
	"AIWorkHelper/pkg/httpx"
)

type User struct {
	svcCtx *svc.ServiceContext
	user   logic.User
}

func NewUser(svcCtx *svc.ServiceContext, user logic.User) *User {
	return &User{
		svcCtx: svcCtx,
		user:   user,
	}
}

func (h *User) InitRegister(engine *gin.Engine) {
	g0 := engine.Group("v1/user")
	g0.POST("/login", h.Login)

	g1 := engine.Group("v1/user", h.svcCtx.Jwt.Handler)
	g1.GET("/:id", h.Info)
	g1.POST("", h.Create)
	g1.PUT("", h.Edit)
	g1.DELETE("/:id", h.Delete)
	g1.GET("/list", h.List)
	g1.POST("/password", h.UpdatePassword)
}

// 用户登录
func (h *User) Login(ctx *gin.Context) {
	var req domain.LoginReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	res, err := h.user.Login(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.OkWithData(ctx, res)
	}
}

// 获取用户信息
func (h *User) Info(ctx *gin.Context) {
	var req domain.IdPathReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	res, err := h.user.Info(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.OkWithData(ctx, res)
	}
}

// 创建用户
func (h *User) Create(ctx *gin.Context) {
	var req domain.User
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	err := h.user.Create(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.Ok(ctx)
	}
}

// 编辑用户
func (h *User) Edit(ctx *gin.Context) {
	var req domain.User
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	err := h.user.Edit(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.Ok(ctx)
	}
}

// 删除用户
func (h *User) Delete(ctx *gin.Context) {
	var req domain.IdPathReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	err := h.user.Delete(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.Ok(ctx)
	}
}

// 用户列表
func (h *User) List(ctx *gin.Context) {
	var req domain.UserListReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	res, err := h.user.List(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.OkWithData(ctx, res)
	}
}

// 修改密码
func (h *User) UpdatePassword(ctx *gin.Context) {
	var req domain.UpdatePasswordReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	err := h.user.UpdatePassword(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.Ok(ctx)
	}
}
