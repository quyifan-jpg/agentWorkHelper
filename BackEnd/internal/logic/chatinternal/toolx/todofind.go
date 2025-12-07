package toolx

import (
	"BackEnd/internal/domain"
	"BackEnd/internal/svc"
	"BackEnd/pkg/curl"
	"BackEnd/pkg/httpx"
	"BackEnd/pkg/langchain/outputparserx"
	"BackEnd/pkg/token"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/langchaingo/callbacks"
)

// TodoFind 待办事项查询工具，实现AI代理的待办查找功能
type TodoFind struct {
	svc          *svc.ServiceContext      // 服务上下文
	callback     callbacks.Handler        // 回调处理器，用于记录执行日志
	outputparser outputparserx.Structured // 结构化输出解析器，解析AI输出为查询条件
}

// NewTodoFind 创建待办事项查询工具实例
func NewTodoFind(svc *svc.ServiceContext) *TodoFind {
	return &TodoFind{
		svc:      svc,
		callback: svc.Callbacks,
		// 配置结构化输出解析器，定义查询条件的字段格式
		outputparser: outputparserx.NewStructured([]outputparserx.ResponseSchema{
			{
				Name:        "id",
				Description: "todo id",
				Type:        "string",
			},
			{
				Name:        "startTime",
				Description: "start time, data application time stamp, such as 1720921573. none is empty",
				Type:        "int64",
			},
			{
				Name:        "endTime",
				Description: "end time, data application time stamp, such as 1720921573. none is empty",
				Type:        "int64",
			},
			{
				Name:        "userId",
				Description: "user id",
				Type:        "string",
			},
		}),
	}
}

// Name 返回工具名称，用于AI代理识别
func (t *TodoFind) Name() string {
	return "todo_find"
}

// Description 返回工具描述和使用说明，包含输出格式指令
func (t *TodoFind) Description() string {
	return `
	a todo find interface.
	use when you need to find, query, search or list todos.
	use when user asks: "我的待办", "查询待办", "有哪些待办", "待办事项", "find my todos", etc.
	If user doesn't provide specific conditions (id, startTime, endTime), query all todos by leaving those fields empty.
	If the condition is null, return {}
	keep Chinese output.` + t.outputparser.GetFormatInstructions()
}

// Call 执行待办事项查询操作
func (t *TodoFind) Call(ctx context.Context, input string) (string, error) {
	// 记录工具调用日志
	if t.callback != nil {
		t.callback.HandleText(ctx, "todo add start : "+input)
	}

	// 解析AI输入为查询条件
	out, err := t.outputparser.Parse(input)
	if err != nil {
		return "", err
	}

	// 构建查询参数
	data := out.(map[string]any)
	if data == nil {
		data = make(map[string]any)
	}
	uid, _ := token.GetUserID(ctx)
	data["userId"] = uid              // 设置当前用户ID
	data["count"] = 10                // 设置查询数量限制
	conversionTime("startTime", data) // 转换开始时间格式
	conversionTime("endTime", data)   // 转换结束时间格式

	// 确保Host包含协议
	host := t.svc.Config.Host
	if host == "0.0.0.0" {
		host = "127.0.0.1"
	}
	if !strings.HasPrefix(host, "http") {
		host = "http://" + host
	}
	// Append port if not present
	if !strings.Contains(host, ":") || (strings.Count(host, ":") == 1 && strings.HasPrefix(host, "http")) {
		host = fmt.Sprintf("%s:%d", host, t.svc.Config.Port)
	}

	// 调用API查询待办事项
	res, err := curl.GetRequest(token.GetTokenStr(ctx), host+"/v1/todo/list", data)
	if err != nil {
		return "", err
	}

	if t.callback != nil {
		t.callback.HandleText(ctx, "todo find end data : "+string(res))
	}

	// 解析API响应并格式化输出（对标Java版本的handleFindTodo方法）
	return t.formatTodoList(res)
}

// conversionTime 转换时间字段格式，将float64转换为int64
func conversionTime(filed string, data map[string]any) {
	if v, ok := data[filed]; ok {
		tmp := v.(float64)
		data[filed] = int64(tmp)
	}
}

// formatTodoList 格式化待办列表输出
// 对标Java版本的handleFindTodo方法（TodoAIHandler.java:263-317）
func (t *TodoFind) formatTodoList(res []byte) (string, error) {
	// 解析HTTP响应
	var apiResponse httpx.Response
	if err := json.Unmarshal(res, &apiResponse); err != nil {
		return "", err
	}

	// 检查响应状态码
	if apiResponse.Code != 200 {
		return "", errors.New(apiResponse.Msg)
	}

	// 解析待办列表数据
	var listResp domain.TodoListResp
	dataBytes, err := json.Marshal(apiResponse.Data)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(dataBytes, &listResp); err != nil {
		return "", err
	}

	// 如果没有待办事项
	if listResp.List == nil || len(listResp.List) == 0 {
		return "您当前没有待办事项。", nil
	}

	// 格式化输出待办列表（对标Java版本第296-311行）
	var result strings.Builder
	result.WriteString("您的待办事项:\n\n")

	for i, todo := range listResp.List {
		// 格式化序号和标题
		result.WriteString(fmt.Sprintf("%d. %s\n", i+1, todo.Title))

		// 格式化状态
		result.WriteString(fmt.Sprintf("   状态: %s\n", getTodoStatusName(todo.TodoStatus)))

		// 格式化截止时间
		result.WriteString(fmt.Sprintf("   截止时间: %s\n", formatTodoTimestamp(todo.DeadlineAt)))

		// 如果有描述则显示
		if todo.Desc != "" {
			result.WriteString(fmt.Sprintf("   描述: %s\n", todo.Desc))
		}

		result.WriteString("\n")
	}

	return result.String(), nil
}

// getTodoStatusName 获取待办状态名称（对标Java版本TodoStatus枚举）
// Java版本: TodoStatus.java
// PENDING(1, "待处理"), IN_PROGRESS(2, "进行中"), FINISHED(3, "已完成"),
// CANCELLED(4, "已取消"), TIMEOUT(5, "已超时")
func getTodoStatusName(status int) string {
	switch status {
	case 1:
		return "待处理"
	case 2:
		return "进行中"
	case 3:
		return "已完成"
	case 4:
		return "已取消"
	case 5:
		return "已超时"
	default:
		return "未知状态"
	}
}

// formatTodoTimestamp 格式化待办时间戳（对标Java版本第349-358行）
func formatTodoTimestamp(timestamp int64) string {
	if timestamp == 0 {
		return "未设置"
	}
	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02 15:04")
}
