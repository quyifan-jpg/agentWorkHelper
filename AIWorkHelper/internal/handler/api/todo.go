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

// Todo 待办事项HTTP处理器
type Todo struct {
	svcCtx *svc.ServiceContext // 服务上下文
	todo   logic.Todo          // 待办事项业务逻辑
}

// NewTodo 创建待办事项HTTP处理器实例
func NewTodo(svcCtx *svc.ServiceContext, todo logic.Todo) *Todo {
	return &Todo{
		svcCtx: svcCtx,
		todo:   todo,
	}
}

// InitRegister 注册待办事项相关的路由，所有路由都需要JWT认证
func (h *Todo) InitRegister(engine *gin.Engine) {
	g := engine.Group("v1/todo", h.svcCtx.Jwt.Handler) // 创建路由组并添加JWT中间件
	g.GET("/:id", h.Info)           // 获取待办详情
	g.POST("", h.Create)            // 创建待办
	g.PUT("", h.Edit)               // 编辑待办
	g.DELETE("/:id", h.Delete)      // 删除待办
	g.POST("/finish", h.Finish)     // 完成待办
	g.POST("/record", h.CreateRecord) // 创建操作记录
	g.GET("/list", h.List)          // 获取待办列表
}

// Info 获取待办事项详情的HTTP处理器
func (h *Todo) Info(ctx *gin.Context) {
	var req domain.IdPathReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil { // 绑定和验证请求参数
		httpx.FailWithErr(ctx, err)
		return
	}

	res, err := h.todo.Info(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.OkWithData(ctx, res)
	}
}

// Create 创建待办事项的HTTP处理器
func (h *Todo) Create(ctx *gin.Context) {
	var req domain.Todo
	if err := httpx.BindAndValidate(ctx, &req); err != nil { // 绑定和验证请求参数
		httpx.FailWithErr(ctx, err)
		return
	}

	res, err := h.todo.Create(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.OkWithData(ctx, res)
	}
}

// Edit 编辑待办事项的HTTP处理器
func (h *Todo) Edit(ctx *gin.Context) {
	var req domain.Todo
	if err := httpx.BindAndValidate(ctx, &req); err != nil { // 绑定和验证请求参数
		httpx.FailWithErr(ctx, err)
		return
	}

	err := h.todo.Edit(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.Ok(ctx)
	}
}

// Delete 删除待办事项的HTTP处理器
func (h *Todo) Delete(ctx *gin.Context) {
	var req domain.IdPathReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil { // 绑定和验证请求参数
		httpx.FailWithErr(ctx, err)
		return
	}

	err := h.todo.Delete(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.Ok(ctx)
	}
}

// Finish 完成待办事项的HTTP处理器
func (h *Todo) Finish(ctx *gin.Context) {
	var req domain.FinishedTodoReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil { // 绑定和验证请求参数
		httpx.FailWithErr(ctx, err)
		return
	}

	err := h.todo.Finish(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.Ok(ctx)
	}
}

// CreateRecord 创建用户操作记录的HTTP处理器
func (h *Todo) CreateRecord(ctx *gin.Context) {
	var req domain.TodoRecord
	if err := httpx.BindAndValidate(ctx, &req); err != nil { // 绑定和验证请求参数
		httpx.FailWithErr(ctx, err)
		return
	}

	err := h.todo.CreateRecord(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.Ok(ctx)
	}
}

// List 获取待办列表的HTTP处理器
func (h *Todo) List(ctx *gin.Context) {
	var req domain.TodoListReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil { // 绑定和验证请求参数
		httpx.FailWithErr(ctx, err)
		return
	}

	res, err := h.todo.List(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.OkWithData(ctx, res)
	}
}
