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

// Approval 审批处理器结构体
type Approval struct {
	svcCtx   *svc.ServiceContext // 服务上下文，包含数据库连接等
	approval logic.Approval      // 审批业务逻辑接口
}

// NewApproval 创建审批处理器实例
func NewApproval(svcCtx *svc.ServiceContext, approval logic.Approval) *Approval {
	return &Approval{
		svcCtx:   svcCtx,
		approval: approval,
	}
}

// InitRegister 初始化审批相关的路由注册
func (h *Approval) InitRegister(engine *gin.Engine) {
	// 创建审批路由组，添加JWT中间件进行身份验证
	g := engine.Group("v1/approval", h.svcCtx.Jwt.Handler)
	g.GET("/:id", h.Info)        // 获取审批详情
	g.POST("", h.Create)         // 创建审批申请
	g.PUT("/dispose", h.Dispose) // 处理审批（通过/拒绝/撤销）
	g.GET("/list", h.List)       // 获取审批列表
}

// Info 获取审批详情接口
func (h *Approval) Info(ctx *gin.Context) {
	var req domain.IdPathReq
	// 绑定并验证请求参数
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	// 调用业务逻辑获取审批详情
	res, err := h.approval.Info(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.OkWithData(ctx, res)
	}
}

// Create 创建审批申请接口
func (h *Approval) Create(ctx *gin.Context) {
	var req domain.Approval
	// 绑定并验证请求参数
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	// 调用业务逻辑创建审批申请
	res, err := h.approval.Create(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.OkWithData(ctx, res)
	}
}

// Dispose 处理审批申请接口（通过/拒绝/撤销）
func (h *Approval) Dispose(ctx *gin.Context) {
	var req domain.DisposeReq
	// 绑定并验证请求参数
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	// 调用业务逻辑处理审批申请
	err := h.approval.Dispose(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.Ok(ctx)
	}
}

// List 获取审批列表接口
func (h *Approval) List(ctx *gin.Context) {
	var req domain.ApprovalListReq
	// 绑定并验证请求参数
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	// 调用业务逻辑获取审批列表
	res, err := h.approval.List(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.OkWithData(ctx, res)
	}
}
