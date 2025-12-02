package api

import (
	"BackEnd/internal/domain"
	"BackEnd/internal/logic"
	"BackEnd/internal/svc"
	"BackEnd/pkg/httpx"
	"BackEnd/pkg/token"

	"github.com/gin-gonic/gin"
)

// User 用户相关 Handler
type User struct {
	svcCtx *svc.ServiceContext
	user   logic.UserLogic
}

// NewUser 创建用户 Handler
func NewUser(svcCtx *svc.ServiceContext, userLogic logic.UserLogic) *User {
	return &User{
		svcCtx: svcCtx,
		user:   userLogic,
	}
}

// InitRegister 注册路由
func (h *User) InitRegister(engine *gin.Engine) {
	// 公开路由（不需要认证）
	g0 := engine.Group("/v1/user")
	{
		g0.POST("/register", h.Register)
		g0.POST("/login", h.Login)
	}

	// 需要认证的路由
	g1 := engine.Group("/v1/user", h.svcCtx.Jwt.Handler)
	{
		// 原有路由
		g1.GET("/info", h.Info)
		g1.PUT("/profile", h.UpdateProfile)
		g1.POST("/password", h.ChangePassword)

		// 新增路由
		g1.GET("/:id", h.GetUserByID)
		g1.POST("", h.CreateUser)
		g1.PUT("", h.UpdateUser)
		g1.DELETE("/:id", h.DeleteUser)
		g1.GET("/list", h.ListUsers)
	}
}

// Register 用户注册
// @Summary      用户注册
// @Description  注册新用户
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        req  body      types.RegisterReq  true  "注册信息"
// @Success      200  {object}  object{code=int,message=string,data=types.RegisterResp}
// @Failure      400  {object}  object{code=int,message=string}
// @Router       /v1/user/register [post]
func (h *User) Register(ctx *gin.Context) {
	var req domain.RegisterReq

	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	err := h.user.Register(ctx.Request.Context(), req.Name, req.Password)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.SuccessWithMessage(ctx, "注册成功", domain.RegisterResp{Message: "注册成功"})
}

// Login 用户登录
// @Summary      用户登录
// @Description  用户登录获取 JWT Token
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        req  body      types.LoginReq  true  "登录信息"
// @Success      200  {object}  object{code=int,msg=string,data=types.LoginResp}
// @Failure      400  {object}  object{code=int,msg=string}
// @Failure      401  {object}  object{code=int,msg=string}
// @Router       /v1/user/login [post]
func (h *User) Login(ctx *gin.Context) {
	var req domain.LoginReq

	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	loginResp, err := h.user.Login(ctx.Request.Context(), req.Name, req.Password)
	if err != nil {
		httpx.Unauthorized(ctx, err.Error())
		return
	}

	httpx.Success(ctx, loginResp)
}

// Info 获取用户信息
// @Summary      获取用户信息
// @Description  获取当前登录用户的信息
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  object{code=int,message=string,data=types.UserInfoResp}
// @Failure      401  {object}  object{code=int,message=string}
// @Router       /v1/user/info [get]
func (h *User) Info(ctx *gin.Context) {
	userID, err := token.GetUserIDFromGin(ctx)
	if err != nil {
		httpx.Unauthorized(ctx, err.Error())
		return
	}

	userInfo, err := h.user.GetInfo(ctx.Request.Context(), userID)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.Success(ctx, userInfo)
}

// UpdateProfile 更新用户资料
// @Summary      更新用户资料
// @Description  更新当前登录用户的资料
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        req  body      types.UpdateProfileReq  true  "用户资料"
// @Success      200  {object}  object{code=int,message=string}
// @Failure      400  {object}  object{code=int,message=string}
// @Router       /v1/user/profile [put]
func (h *User) UpdateProfile(ctx *gin.Context) {
	userID, err := token.GetUserIDFromGin(ctx)
	if err != nil {
		httpx.Unauthorized(ctx, err.Error())
		return
	}

	var req domain.UpdateProfileReq

	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	err = h.user.UpdateProfile(ctx.Request.Context(), userID, req.Name)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.SuccessWithMessage(ctx, "更新成功", nil)
}

// ChangePassword 修改密码
// @Summary      修改密码
// @Description  修改当前登录用户的密码
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        req  body      types.UpdatePasswordReq  true  "密码信息"
// @Success      200  {object}  object{code=int,msg=string}
// @Failure      400  {object}  object{code=int,msg=string}
// @Failure      401  {object}  object{code=int,msg=string}
// @Router       /v1/user/password [post]
func (h *User) ChangePassword(ctx *gin.Context) {
	// 从 JWT token 中获取当前用户 ID
	userID, err := token.GetUserIDFromGin(ctx)
	if err != nil {
		httpx.Unauthorized(ctx, err.Error())
		return
	}
	// 绑定并验证请求参数
	var req domain.UpdatePasswordReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}
	// 调用业务逻辑层修改密码
	err = h.user.ChangePassword(ctx.Request.Context(), userID, req.OldPwd, req.NewPwd)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	httpx.SuccessWithMessage(ctx, "密码修改成功", nil)
}

// GetUserByID 获取指定用户信息
// @Summary      获取用户信息
// @Description  根据ID获取用户信息
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "用户ID"
// @Success      200  {object}  object{code=int,msg=string,data=types.UserInfoResp}
// @Failure      400  {object}  object{code=int,msg=string}
// @Router       /v1/user/{id} [get]
func (h *User) GetUserByID(ctx *gin.Context) {
	var req domain.IdPathReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	userInfo, err := h.user.GetInfoByID(ctx.Request.Context(), req.Id)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.Success(ctx, userInfo)
}

// CreateUser 创建用户
// @Summary      创建用户
// @Description  管理员创建新用户
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        req  body      types.User  true  "用户信息"
// @Success      200  {object}  object{code=int,msg=string}
// @Failure      400  {object}  object{code=int,msg=string}
// @Router       /v1/user [post]
func (h *User) CreateUser(ctx *gin.Context) {
	var req domain.User
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	// TODO: 检查当前用户是否为管理员
	// currentUserID, _ := token.GetUserIDFromGin(ctx)
	// ...

	err := h.user.Create(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.SuccessWithMessage(ctx, "创建成功", nil)
}

// UpdateUser 更新用户
// @Summary      更新用户
// @Description  管理员更新用户信息
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        req  body      types.User  true  "用户信息"
// @Success      200  {object}  object{code=int,msg=string}
// @Failure      400  {object}  object{code=int,msg=string}
// @Router       /v1/user [put]
func (h *User) UpdateUser(ctx *gin.Context) {
	var req domain.User
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	err := h.user.Update(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.SuccessWithMessage(ctx, "更新成功", nil)
}

// DeleteUser 删除用户
// @Summary      删除用户
// @Description  管理员删除用户
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "用户ID"
// @Success      200  {object}  object{code=int,msg=string}
// @Failure      400  {object}  object{code=int,msg=string}
// @Router       /v1/user/{id} [delete]
func (h *User) DeleteUser(ctx *gin.Context) {
	var req domain.IdPathReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	err := h.user.Delete(ctx.Request.Context(), req.Id)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.SuccessWithMessage(ctx, "删除成功", nil)
}

// ListUsers 用户列表
// @Summary      用户列表
// @Description  分页查询用户列表
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page   query     int     false  "页码"
// @Param        count  query     int     false  "每页数量"
// @Param        name   query     string  false  "用户名搜索"
// @Success      200  {object}  object{code=int,msg=string,data=types.UserListResp}
// @Failure      400  {object}  object{code=int,msg=string}
// @Router       /v1/user/list [get]
func (h *User) ListUsers(ctx *gin.Context) {
	var req domain.UserListReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	resp, err := h.user.List(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.Success(ctx, resp)
}
