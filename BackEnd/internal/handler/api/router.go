package api

import (
	"BackEnd/internal/logic"
	"BackEnd/internal/svc"
)

// initHandler 初始化所有 Handler
func initHandler(svcCtx *svc.ServiceContext) []Handler {
	// 初始化业务逻辑层
	var (
		userLogic = logic.NewUserLogic(svcCtx)
	)

	// 初始化 Handler
	var (
		user = NewUser(svcCtx, userLogic)
	)

	return []Handler{
		user,
	}
}

