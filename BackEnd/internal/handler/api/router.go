package api

import (
"BackEnd/internal/logic"
"BackEnd/internal/svc"
)

func initHandler(svc *svc.ServiceContext) []Handler {
	// new logics
	var (
userLogic = logic.NewUser(svc)
deptLogic = logic.NewDepartment(svc)
todoLogic = logic.NewTodo(svc)
)

	// new handlers
	var (
user = NewUser(svc, userLogic)
dept = NewDepartment(svc, deptLogic)
todo = NewTodo(svc, todoLogic)
)

	return []Handler{
		user,
		dept,
		todo,
	}
}
