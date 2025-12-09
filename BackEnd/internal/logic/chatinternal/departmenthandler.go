package chatinternal

import (
	"BackEnd/internal/logic/chatinternal/toolx"
	"BackEnd/internal/svc"

	"github.com/tmc/langchaingo/tools"
)

type DepartmentHandle struct {
	*AgentChat
}

func NewDepartmentHandle(svc *svc.ServiceContext, l toolx.DepartmentLogic) *DepartmentHandle {
	return &DepartmentHandle{
		AgentChat: NewAgentChat(svc, []tools.Tool{
			toolx.NewDepartmentList(svc),
			toolx.NewDepartmentUsers(svc, l),
		}),
	}
}

func (t *DepartmentHandle) Name() string {
	return "department"
}

func (t *DepartmentHandle) Description() string {
	return "suitable for department processing, such as department creation, query, modification, deletion, etc"
}
