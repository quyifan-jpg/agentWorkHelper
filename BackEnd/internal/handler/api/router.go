package api

import (
	"BackEnd/internal/logic"
	"BackEnd/internal/svc"
)

func initHandler(svc *svc.ServiceContext) []Handler {
	// new logics
	var (
		todoLogic       = logic.NewTodo(svc)
		approvalLogic   = logic.NewApproval(svc)
		userLogic       = logic.NewUser(svc)
		departmentLogic = logic.NewDepartment(svc)
	)

	// new handlers
	var (
		department = NewDepartment(svc, departmentLogic)
		todo       = NewTodo(svc, todoLogic)
		approval   = NewApproval(svc, approvalLogic)
		user       = NewUser(svc, userLogic)
	)

	return []Handler{
		department,
		todo,
		approval,
		user,
	}
}
