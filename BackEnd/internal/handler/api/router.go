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
		userLogic       = logic.NewUser(svc)
	)

	// new handlers
	var (
		todo       = NewTodo(svc, todoLogic)
		approval   = NewApproval(svc, approvalLogic)
		user       = NewUser(svc, userLogic)
		department = NewDepartment(svc, departmentLogic)
	)

	return []Handler{
		todo,
		approval,
		user,
		department,
	}
}
