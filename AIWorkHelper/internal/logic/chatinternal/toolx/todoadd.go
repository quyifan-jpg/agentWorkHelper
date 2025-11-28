/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package toolx

import (
	"AIWorkHelper/internal/domain"
	"AIWorkHelper/internal/svc"
	"AIWorkHelper/pkg/curl"
	"AIWorkHelper/pkg/langchain/outputparserx"
	"AIWorkHelper/pkg/token"
	"context"
	"encoding/json"
	"github.com/tmc/langchaingo/callbacks"
)

// TodoAdd 待办事项添加工具，实现AI代理的待办创建功能
type TodoAdd struct {
	svc          *svc.ServiceContext      // 服务上下文
	callback     callbacks.Handler        // 回调处理器，用于记录执行日志
	outputparser outputparserx.Structured // 结构化输出解析器，解析AI输出为结构化数据
}

// NewTodoAdd 创建待办事项添加工具实例
func NewTodoAdd(svc *svc.ServiceContext) *TodoAdd {
	return &TodoAdd{
		svc:      svc,
		callback: svc.Callbacks,
		// 配置结构化输出解析器，定义待办事项的字段格式
		outputparser: outputparserx.NewStructured([]outputparserx.ResponseSchema{
			{
				Name:        "title",
				Description: "todo title",
			}, {
				Name:        "deadlineAt",
				Description: "the deadline Unix timestamp (in seconds). You MUST use the time_parser tool first to convert the user's time expression to a timestamp, then use the timestamp value returned by time_parser tool here.",
				Type:        "int64",
			}, {
				Name:        "desc",
				Description: "todo description",
			}, {
				Name:        "executeIds",
				Description: "list of participating users in the backlog. the data type is a set of string ids. none is empty",
				Type:        "[]string",
			},
		}),
	}
}

// Name 返回工具名称，用于AI代理识别
func (t *TodoAdd) Name() string {
	return "todo_add"
}

// Description 返回工具描述和使用说明，包含输出格式指令
func (t *TodoAdd) Description() string {
	template := `
	a todo add interface.
	use when you need to create a todo.
	IMPORTANT: if user mentions a person's name (like "王员工"), you MUST first use the user_list tool to query and get the user's ID, then use that ID in executeIds field.
	keep Chinese output.
` + t.outputparser.GetFormatInstructions()

	return template
}

// Call 执行待办事项创建操作
func (t *TodoAdd) Call(ctx context.Context, input string) (string, error) {
	// 记录工具调用日志
	if t.callback != nil {
		t.callback.HandleText(ctx, "todo add start : "+input)
	}

	// 解析AI输入为结构化数据
	data, err := t.outputparser.Parse(input)
	if err != nil {
		return "", err
	}

	// 调用API创建待办事项
	res, err := curl.PostRequest(token.GetTokenStr(ctx), t.svc.Config.Host+"/v1/todo", data)

	// 解析API响应获取创建的待办ID
	var idResp domain.IdRespInfo
	if err := json.Unmarshal(res, &idResp); err != nil {
		return "", err
	}

	// 返回成功消息和创建的待办ID
	return Success + "\ncreated todo id : " + idResp.Data.Id, nil
}
