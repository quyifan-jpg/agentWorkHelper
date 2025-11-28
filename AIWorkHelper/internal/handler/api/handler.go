/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package api

import (
	"AIWorkHelper/internal/middleware"
	"gitee.com/dn-jinmin/tlog"
	"github.com/gin-gonic/gin"

	"AIWorkHelper/internal/handler"
	"AIWorkHelper/internal/svc"
	"AIWorkHelper/pkg/httpx"
)

type Handler interface {
	InitRegister(*gin.Engine)
}

type handle struct {
	srv  *gin.Engine
	addr string
}

func NewHandle(svc *svc.ServiceContext) *handle {
	h := &handle{
		srv:  gin.New(),
		addr: "0.0.0.0:8080",
	}
	if len(svc.Config.Addr) > 0 {
		h.addr = svc.Config.Addr
	}

	tlog.Init(
		tlog.WithLabel(svc.Config.Tlog.Label),
		tlog.WithMode(svc.Config.Tlog.Mode),
	)

	h.srv.Use(middleware.NewLog().Handler)
	httpx.SetErrorHandler(handler.ErrorHandler)

	handlers := initHandler(svc)
	for _, handler := range handlers {
		handler.InitRegister(h.srv)
	}

	return h
}

func (h *handle) Run() error {
	return h.srv.Run(h.addr)
}
