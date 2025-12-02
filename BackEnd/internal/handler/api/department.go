package api

import (
	"BackEnd/internal/domain"
	"BackEnd/internal/logic"
	"BackEnd/internal/svc"
	"BackEnd/pkg/httpx"

	"github.com/gin-gonic/gin"
)

type Department struct {
	logic logic.DepartmentLogic
}

func NewDepartment(svcCtx *svc.ServiceContext, l logic.DepartmentLogic) *Department {
	return &Department{
		logic: l,
	}
}

func (h *Department) InitRegister(r *gin.Engine) {
	group := r.Group("/v1/department")
	// group.Use(middleware.JwtAuth()) // TODO: Add auth middleware
	{
		group.POST("", h.Create)
		group.PUT("", h.Update)
		group.DELETE("/:id", h.Delete)
		group.GET("/:id", h.Get)
		group.GET("/list", h.List)
		group.POST("/user", h.AddUser)
		group.DELETE("/user", h.RemoveUser)
	}
}

func (h *Department) Create(ctx *gin.Context) {
	var req domain.CreateDepartmentReq
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

func (h *Department) Update(ctx *gin.Context) {
	var req domain.UpdateDepartmentReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	if err := h.logic.Update(ctx.Request.Context(), &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	httpx.SuccessWithMessage(ctx, "更新成功", nil)
}

func (h *Department) Delete(ctx *gin.Context) {
	idStr := ctx.Param("id")

	if err := h.logic.Delete(ctx.Request.Context(), idStr); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	httpx.SuccessWithMessage(ctx, "删除成功", nil)
}

func (h *Department) Get(ctx *gin.Context) {
	idStr := ctx.Param("id")

	resp, err := h.logic.Get(ctx.Request.Context(), idStr)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	httpx.Success(ctx, resp)
}

func (h *Department) List(ctx *gin.Context) {
	var req domain.DepartmentListReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	resp, err := h.logic.List(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	httpx.Success(ctx, resp)
}

func (h *Department) AddUser(ctx *gin.Context) {
	var req domain.AddDepartmentUserReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	if err := h.logic.AddUser(ctx.Request.Context(), &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	httpx.SuccessWithMessage(ctx, "添加成功", nil)
}

func (h *Department) RemoveUser(ctx *gin.Context) {
	deptIDStr := ctx.Query("departmentId")
	userIDStr := ctx.Query("userId")

	if deptIDStr == "" || userIDStr == "" {
		httpx.BadRequest(ctx, "invalid params")
		return
	}

	if err := h.logic.RemoveUser(ctx.Request.Context(), deptIDStr, userIDStr); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	httpx.SuccessWithMessage(ctx, "移除成功", nil)
}
