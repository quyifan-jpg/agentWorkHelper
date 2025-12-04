package api

import (
	"BackEnd/internal/logic"
	"BackEnd/internal/svc"
)

func initHandler(svc *svc.ServiceContext) []Handler {
	// new logics
	var (
		departmentLogic = logic.NewDepartment(svc)
		todoLogic       = logic.NewTodo(svc)
		approvalLogic   = logic.NewApproval(svc)
		chatLogic       = logic.NewChat(svc)
		groupLogic      = logic.NewGroup(svc)
		userLogic       = logic.NewUser(svc)
	)

	// new handlers
	var (
		todo       = NewTodo(svc, todoLogic)
		approval   = NewApproval(svc, approvalLogic)
		chat       = NewChat(svc, chatLogic)
		upload     = NewUpload(svc, chatLogic)
		group      = NewGroup(svc, groupLogic)
		user       = NewUser(svc, userLogic)
		department = NewDepartment(svc, departmentLogic)
	)

	return []Handler{
		todo,
		approval,
		chat,
		upload,
		group,
		user,
		department,
	}
}
