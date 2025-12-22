package chatinternal

import (
	"BackEnd/internal/domain"
	"BackEnd/internal/logic/chatinternal/toolx"
	"BackEnd/internal/svc"
	"context"

	"github.com/tmc/langchaingo/tools"
)

// ApprovalLogic interface defines all approval operations required by tools
// This decouples chatinternal from internal/logic
type ApprovalLogic interface {
	Create(ctx context.Context, req *domain.Approval) (resp *domain.IdResp, err error)
	List(ctx context.Context, req *domain.ApprovalListReq) (resp *domain.ApprovalListResp, err error)
}

type ApprovalHandle struct {
	*AgentChat
}

// NewApprovalHandle now accepts the local interface
func NewApprovalHandle(svc *svc.ServiceContext, l ApprovalLogic) *ApprovalHandle {
	return &ApprovalHandle{
		AgentChat: NewAgentChat(svc, []tools.Tool{
			toolx.NewApprovalAdd(svc, l),
			toolx.NewApprovalFind(svc, l),
		}),
	}
}

func (t *ApprovalHandle) Name() string {
	return "approval"
}

func (t *ApprovalHandle) Description() string {
	return "suitable for approval processing, such as applying for leave, overtime, go out, or querying approval records."
}
