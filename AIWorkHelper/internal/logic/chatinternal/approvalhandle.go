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

type ApprovalHandle struct {
	*baseChat
}

func NewApprovalHandle(svc *svc.ServiceContext) *ApprovalHandle {
	return &ApprovalHandle{
		baseChat: NewBaseChat(svc, []tools.Tool{
			toolx.NewTimeParser(svc), // 时间解析工具，用于将自然语言时间转换为Unix时间戳
			toolx.NewApprovalAdd(svc),
			toolx.NewApprovalFind(svc),
		}),
	}
}
func (t *ApprovalHandle) Name() string {
	return "approval"
}

func (t *ApprovalHandle) Description() string {
	return "This is about approval matters. Such as sick leave, personal leave, going out, etc.\n\n"
}
