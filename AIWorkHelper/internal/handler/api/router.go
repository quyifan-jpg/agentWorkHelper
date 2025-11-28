/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package api

import (
	"AIWorkHelper/internal/logic"
	"AIWorkHelper/internal/svc"
)

func initHandler(svc *svc.ServiceContext) []Handler {
	// new logics
	var (
		chatLogic       = logic.NewChat(svc)
		userLogic       = logic.NewUser(svc)
		departmentLogic = logic.NewDepartment(svc)
		todoLogic       = logic.NewTodo(svc)
		approvalLogic   = logic.NewApproval(svc)
	)

	// new handlers
	var (
		todo       = NewTodo(svc, todoLogic)
		approval   = NewApproval(svc, approvalLogic)
		chat       = NewChat(svc, chatLogic)
		upload     = NewUpload(svc, chatLogic)
		user       = NewUser(svc, userLogic)
		department = NewDepartment(svc, departmentLogic)
		group      = NewGroup(svc) // 群聊管理处理器
	)

	return []Handler{
		todo,
		approval,
		chat,
		upload,
		user,
		department,
		group, // 注册群聊API路由
	}
}
