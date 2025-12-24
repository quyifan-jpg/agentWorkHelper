package chatinternal

import (
	"BackEnd/internal/logic/chatinternal/toolx"
	"BackEnd/internal/svc"

	"github.com/tmc/langchaingo/tools"
)

type Knowledge struct {
	*AgentChat
}

func NewKnowledge(svc *svc.ServiceContext) *Knowledge {
	return &Knowledge{NewAgentChat(svc, []tools.Tool{
		toolx.NewKnowledgeUpdate(svc),
		toolx.NewKnowledgeRetrievalQA(svc),
	})}
}

func (t *Knowledge) Name() string {
	return "knowledge"
}

func (t *Knowledge) Description() string {
	return `This is the company's knowledge base.
Can answer employee consultation questions about company systems such as approval process, leave matters, attendance matters, employee manuals and other office content.
Can also be used for updating the knowledge base.`
}
