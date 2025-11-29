package api

import (
	"BackEnd/internal/logic"
	"BackEnd/internal/svc"
)

func initHandler(svc *svc.ServiceContext) []Handler {
	// new logics
	var (
		userLogic = logic.NewUser(svc)
	)

	// new handlers
	var (
		user = NewUser(svc, userLogic)
	)

	return []Handler{
		user,
	}
}
