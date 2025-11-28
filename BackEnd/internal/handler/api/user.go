package api

import (
	"BackEnd/internal/logic"
	"BackEnd/internal/svc"
	"BackEnd/internal/types"
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
		g1.GET("/info", h.Info)
		g1.PUT("/profile", h.UpdateProfile)
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
	var req types.RegisterReq

	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	err := h.user.Register(ctx.Request.Context(), req.Username, req.Password)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	httpx.SuccessWithMessage(ctx, "注册成功", types.RegisterResp{Message: "注册成功"})
}

// Login 用户登录
// @Summary      用户登录
// @Description  用户登录获取 JWT Token
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        req  body      types.LoginReq  true  "登录信息"
// @Success      200  {object}  object{code=int,message=string,data=types.LoginResp}
// @Failure      400  {object}  object{code=int,message=string}
// @Failure      401  {object}  object{code=int,message=string}
// @Router       /v1/user/login [post]
func (h *User) Login(ctx *gin.Context) {
	var req types.LoginReq

	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.BadRequest(ctx, err.Error())
		return
	}

	token, err := h.user.Login(ctx.Request.Context(), req.Username, req.Password)
	if err != nil {
		httpx.Unauthorized(ctx, err.Error())
		return
	}

	httpx.Success(ctx, types.LoginResp{Token: token})
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

	var req types.UpdateProfileReq

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

