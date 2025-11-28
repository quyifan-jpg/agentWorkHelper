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

// Department 部门处理器结构体
type Department struct {
	svcCtx     *svc.ServiceContext // 服务上下文
	department logic.Department    // 部门业务逻辑接口
}

// NewDepartment 创建部门处理器实例
func NewDepartment(svcCtx *svc.ServiceContext, department logic.Department) *Department {
	return &Department{
		svcCtx:     svcCtx,
		department: department,
	}
}

// InitRegister 注册部门相关路由
func (h *Department) InitRegister(engine *gin.Engine) {
	g := engine.Group("v1/dep", h.svcCtx.Jwt.Handler) // 创建部门路由组，添加JWT中间件
	g.GET("/soa", h.Soa)                              // 获取部门SOA信息
	g.GET("/:id", h.Info)                             // 获取部门详情
	g.POST("", h.Create)                              // 创建部门
	g.PUT("", h.Edit)                                 // 更新部门
	g.DELETE("/:id", h.Delete)                        // 删除部门
	g.POST("/user", h.SetDepartmentUsers)             // 设置部门用户
	g.POST("/user/add", h.AddDepartmentUser)          // 添加部门员工
	g.DELETE("/user/remove", h.RemoveDepartmentUser)  // 删除部门员工
	g.GET("/user/:id", h.DepartmentUserInfo)          // 获取用户部门信息
}

// Soa 获取部门SOA信息处理器
func (h *Department) Soa(ctx *gin.Context) {
	res, err := h.department.Soa(ctx.Request.Context())
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.OkWithData(ctx, res)
	}
}

// Info 获取部门详情处理器
func (h *Department) Info(ctx *gin.Context) {
	var req domain.IdPathReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	res, err := h.department.Info(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.OkWithData(ctx, res)
	}
}

// Create 创建部门处理器
func (h *Department) Create(ctx *gin.Context) {
	var req domain.Department
	if err := httpx.BindAndValidate(ctx, &req); err != nil { // 绑定并验证请求参数
		httpx.FailWithErr(ctx, err)
		return
	}

	err := h.department.Create(ctx.Request.Context(), &req) // 调用业务逻辑创建部门
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.Ok(ctx) // 返回成功响应
	}
}

// Edit 更新部门处理器
func (h *Department) Edit(ctx *gin.Context) {
	var req domain.Department
	if err := httpx.BindAndValidate(ctx, &req); err != nil { // 绑定并验证请求参数
		httpx.FailWithErr(ctx, err)
		return
	}

	err := h.department.Edit(ctx.Request.Context(), &req) // 调用业务逻辑更新部门
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.Ok(ctx) // 返回成功响应
	}
}

// Delete 删除部门处理器
func (h *Department) Delete(ctx *gin.Context) {
	var req domain.IdPathReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil { // 绑定并验证路径参数
		httpx.FailWithErr(ctx, err)
		return
	}

	err := h.department.Delete(ctx.Request.Context(), &req) // 调用业务逻辑删除部门
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.Ok(ctx) // 返回成功响应
	}
}

// SetDepartmentUsers 设置部门用户处理器
func (h *Department) SetDepartmentUsers(ctx *gin.Context) {
	var req domain.SetDepartmentUser
	if err := httpx.BindAndValidate(ctx, &req); err != nil { // 绑定并验证请求参数
		httpx.FailWithErr(ctx, err)
		return
	}

	err := h.department.SetDepartmentUsers(ctx.Request.Context(), &req) // 调用业务逻辑设置部门用户
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.Ok(ctx) // 返回成功响应
	}
}

// AddDepartmentUser 添加部门员工处理器
func (h *Department) AddDepartmentUser(ctx *gin.Context) {
	var req domain.AddDepartmentUser
	if err := httpx.BindAndValidate(ctx, &req); err != nil { // 绑定并验证请求参数
		httpx.FailWithErr(ctx, err)
		return
	}

	err := h.department.AddDepartmentUser(ctx.Request.Context(), &req) // 调用业务逻辑添加部门员工
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.Ok(ctx) // 返回成功响应
	}
}

// RemoveDepartmentUser 删除部门员工处理器
func (h *Department) RemoveDepartmentUser(ctx *gin.Context) {
	var req domain.RemoveDepartmentUser
	if err := httpx.BindAndValidate(ctx, &req); err != nil { // 绑定并验证请求参数
		httpx.FailWithErr(ctx, err)
		return
	}

	err := h.department.RemoveDepartmentUser(ctx.Request.Context(), &req) // 调用业务逻辑删除部门员工
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.Ok(ctx) // 返回成功响应
	}
}

// DepartmentUserInfo 获取用户部门信息处理器
func (h *Department) DepartmentUserInfo(ctx *gin.Context) {
	var req domain.IdPathReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil { // 绑定并验证路径参数
		httpx.FailWithErr(ctx, err)
		return
	}

	res, err := h.department.DepartmentUserInfo(ctx.Request.Context(), &req) // 调用业务逻辑获取用户部门信息
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.OkWithData(ctx, res) // 返回成功响应和数据
	}
}
