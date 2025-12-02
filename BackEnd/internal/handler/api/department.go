package api

import (
	"BackEnd/internal/domain"
	"BackEnd/internal/logic"
	"BackEnd/internal/svc"
	"BackEnd/pkg/httpx"

	"github.com/gin-gonic/gin"
)

type Department struct {
	logic  logic.DepartmentLogic
	svcCtx *svc.ServiceContext
}

func NewDepartment(svcCtx *svc.ServiceContext, l logic.DepartmentLogic) *Department {
	return &Department{
		logic:  l,
		svcCtx: svcCtx,
	}
}

func (h *Department) InitRegister(r *gin.Engine) {
	group := r.Group("/v1/dep")
	group.Use(h.svcCtx.Jwt.Handler)
	{
		group.GET("/soa", h.Soa)
		group.GET("/:id", h.Info)
		group.POST("", h.Create)
		group.PUT("", h.Edit)
		group.DELETE("/:id", h.Delete)
		group.POST("/user", h.SetDepartmentUsers)
		group.POST("/user/add", h.AddDepartmentUser)
		group.DELETE("/user/remove", h.RemoveDepartmentUser)
		group.GET("/user/:id", h.DepartmentUserInfo)
	}
}

func (h *Department) Soa(ctx *gin.Context) {
	var req domain.DepartmentListReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	resp, err := h.logic.Soa(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	httpx.Success(ctx, resp)
}

func (h *Department) Info(ctx *gin.Context) {
	var req domain.IdPathReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	resp, err := h.logic.Info(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	httpx.Success(ctx, resp)
}

func (h *Department) Create(ctx *gin.Context) {
	var req domain.Department
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	if err := h.logic.Create(ctx.Request.Context(), &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	httpx.SuccessWithMessage(ctx, "创建成功", nil)
}

func (h *Department) Edit(ctx *gin.Context) {
	var req domain.Department
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	if err := h.logic.Edit(ctx.Request.Context(), &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	httpx.SuccessWithMessage(ctx, "更新成功", nil)
}

func (h *Department) Delete(ctx *gin.Context) {
	var req domain.IdPathReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	if err := h.logic.Delete(ctx.Request.Context(), &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	httpx.SuccessWithMessage(ctx, "删除成功", nil)
}

func (h *Department) SetDepartmentUsers(ctx *gin.Context) {
	var req domain.SetDepartmentUser
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	if err := h.logic.SetDepartmentUsers(ctx.Request.Context(), &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	httpx.SuccessWithMessage(ctx, "设置成功", nil)
}

func (h *Department) AddDepartmentUser(ctx *gin.Context) {
	var req domain.AddDepartmentUser
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	if err := h.logic.AddDepartmentUser(ctx.Request.Context(), &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	httpx.SuccessWithMessage(ctx, "添加成功", nil)
}

func (h *Department) RemoveDepartmentUser(ctx *gin.Context) {
	var req domain.RemoveDepartmentUser
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	if err := h.logic.RemoveDepartmentUser(ctx.Request.Context(), &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	httpx.SuccessWithMessage(ctx, "删除成功", nil)
}

func (h *Department) DepartmentUserInfo(ctx *gin.Context) {
	var req domain.IdPathReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	resp, err := h.logic.DepartmentUserInfo(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	httpx.Success(ctx, resp)
}
