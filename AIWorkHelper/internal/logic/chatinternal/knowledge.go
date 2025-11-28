/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package chatinternal

import (
	"AIWorkHelper/internal/logic/chatinternal/toolx"
	"AIWorkHelper/internal/svc"
	"github.com/tmc/langchaingo/tools"
)

type Knowledge struct {
	*baseChat
}

func NewKnowledge(svc *svc.ServiceContext) *Knowledge {
	return &Knowledge{NewBaseChat(svc, []tools.Tool{
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
