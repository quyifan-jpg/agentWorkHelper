package api

import (
	"BackEnd/internal/middleware"
	"BackEnd/internal/svc"
	"fmt"

	"github.com/gin-gonic/gin"
)

// Handler 定义统一的 Handler 接口
type Handler interface {
	InitRegister(*gin.Engine)
}

// ApiHandler API 处理器管理器
type ApiHandler struct {
	srv    *gin.Engine
	addr   string
	svcCtx *svc.ServiceContext
}

// NewApiHandler 创建 API 处理器管理器
func NewApiHandler(svcCtx *svc.ServiceContext) *ApiHandler {
	h := &ApiHandler{
		srv:    gin.New(),
		addr:   "0.0.0.0:8889",
		svcCtx: svcCtx,
	}

	// 设置地址
	if svcCtx.Config.Host != "" && svcCtx.Config.Port > 0 {
		h.addr = svcCtx.Config.Host + ":" + fmt.Sprintf("%d", svcCtx.Config.Port)
	}

	// 添加 CORS 中间件（处理 Swagger 和 OPTIONS 预检请求）
	h.srv.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 添加中间件
	h.srv.Use(gin.Recovery())
	h.srv.Use(middleware.NewLog().Handler)

	// 注册所有 Handler
	handlers := initHandler(svcCtx)
	for _, handler := range handlers {
		handler.InitRegister(h.srv)
	}

	return h
}

// Run 启动服务
func (h *ApiHandler) Run() error {
	return h.srv.Run(h.addr)
}

// GetEngine 获取 Gin Engine（用于 Swagger 等）
func (h *ApiHandler) GetEngine() *gin.Engine {
	return h.srv
}

