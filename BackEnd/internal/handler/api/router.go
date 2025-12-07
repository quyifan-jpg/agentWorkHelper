package api

import (
	"BackEnd/internal/logic"
	"BackEnd/internal/svc"
)

func initHandler(svc *svc.ServiceContext) []Handler {
	// new logics
	var (
		userLogic       = logic.NewUser(svc)
		departmentLogic = logic.NewDepartment(svc)
		todoLogic       = logic.NewTodo(svc)
		approvalLogic   = logic.NewApproval(svc)
		chatLogic       = logic.NewChat(svc)
		groupLogic      = logic.NewGroup(svc)
	)

	// new handlers
	var (
		chat       = NewChat(svc, chatLogic)
		upload     = NewUpload(svc, chatLogic)
		group      = NewGroup(svc, groupLogic)
		user       = NewUser(svc, userLogic)
		department = NewDepartment(svc, departmentLogic)
		todo       = NewTodo(svc, todoLogic)
		approval   = NewApproval(svc, approvalLogic)
	)

	return []Handler{
		chat,
		upload,
		group,
		user,
		department,
		todo,
		approval,
	}
}
