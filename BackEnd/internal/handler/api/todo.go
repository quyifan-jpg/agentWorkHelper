package api

import (
	"BackEnd/internal/domain"
	"BackEnd/internal/logic"
	"BackEnd/internal/svc"
	"BackEnd/pkg/httpx"
	"BackEnd/pkg/token"

	"github.com/gin-gonic/gin"
)

type Todo struct {
	logic logic.TodoLogic
}

func NewTodo(svcCtx *svc.ServiceContext, l logic.TodoLogic) *Todo {
	return &Todo{
		logic: l,
	}
}

func (h *Todo) InitRegister(r *gin.Engine) {
	group := r.Group("/v1/todo")
	// group.Use(middleware.JwtAuth()) // TODO: Add auth middleware
	{
		group.POST("", h.Create)
		group.PUT("", h.Update)
		group.DELETE("/:id", h.Delete)
		group.GET("/:id", h.Get)
		group.GET("/list", h.List)
		group.POST("/finish", h.Finish)
		group.POST("/record", h.CreateRecord)
	}
}

func (h *Todo) Create(ctx *gin.Context) {
	var req domain.Todo
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	userID, err := token.GetUserIDFromGin(ctx)
	if err != nil {
		// For testing without auth middleware, we might need a fallback or just fail
		// httpx.Unauthorized(ctx, err.Error())
		// return
		// Temporary fallback for dev
		userID = 1
	}

	if err := h.logic.Create(ctx.Request.Context(), userID, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	httpx.SuccessWithMessage(ctx, "创建成功", nil)
}

func (h *Todo) Update(ctx *gin.Context) {
	var req domain.Todo
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	userID, err := token.GetUserIDFromGin(ctx)
	if err != nil {
		userID = 1 // Temporary fallback
	}

	if err := h.logic.Update(ctx.Request.Context(), userID, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	httpx.SuccessWithMessage(ctx, "更新成功", nil)
}

func (h *Todo) Delete(ctx *gin.Context) {
	idStr := ctx.Param("id")

	userID, err := token.GetUserIDFromGin(ctx)
	if err != nil {
		userID = 1 // Temporary fallback
	}

	if err := h.logic.Delete(ctx.Request.Context(), userID, idStr); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	httpx.SuccessWithMessage(ctx, "删除成功", nil)
}

func (h *Todo) Get(ctx *gin.Context) {
	idStr := ctx.Param("id")

	userID, err := token.GetUserIDFromGin(ctx)
	if err != nil {
		userID = 1 // Temporary fallback
	}

	resp, err := h.logic.Get(ctx.Request.Context(), userID, idStr)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	httpx.Success(ctx, resp)
}

func (h *Todo) List(ctx *gin.Context) {
	var req domain.TodoListReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	userID, err := token.GetUserIDFromGin(ctx)
	if err != nil {
		userID = 1 // Temporary fallback
	}

	resp, err := h.logic.List(ctx.Request.Context(), userID, &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	httpx.Success(ctx, resp)
}

func (h *Todo) Finish(ctx *gin.Context) {
	var req domain.FinishedTodoReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	userID, err := token.GetUserIDFromGin(ctx)
	if err != nil {
		userID = 1 // Temporary fallback
	}

	if err := h.logic.Finish(ctx.Request.Context(), userID, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	httpx.SuccessWithMessage(ctx, "操作成功", nil)
}

func (h *Todo) CreateRecord(ctx *gin.Context) {
	var req domain.TodoRecord
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	userID, err := token.GetUserIDFromGin(ctx)
	if err != nil {
		userID = 1 // Temporary fallback
	}

	if err := h.logic.CreateRecord(ctx.Request.Context(), userID, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	httpx.SuccessWithMessage(ctx, "记录成功", nil)
}
