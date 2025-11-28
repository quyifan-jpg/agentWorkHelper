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

type TodoHandle struct {
	*baseChat
}

func NewTodoHandle(svc *svc.ServiceContext) *TodoHandle {
	return &TodoHandle{
		baseChat: NewBaseChat(svc, []tools.Tool{
			toolx.NewUserList(svc),   // 用户列表查询工具，用于将用户名转换为用户ID
			toolx.NewTimeParser(svc), // 时间解析工具，用于将自然语言时间转换为Unix时间戳
			toolx.NewTodoAdd(svc),
			toolx.NewTodoFind(svc),
		}),
	}
}

func (t *TodoHandle) Name() string {
	return "todo"
}

func (t *TodoHandle) Description() string {
	return "suitable for todo processing, such as todo creation, query, modification, dele tion, etc"
}

//func (t *TodoHandle) Chains(input string) (string, error) {
//	return t.baseChat.Handle(input)
//}
