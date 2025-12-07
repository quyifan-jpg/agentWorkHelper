package toolx

import (
	"BackEnd/internal/domain"
	"BackEnd/pkg/httpx"
	"errors"

	"encoding/json"
)

var (
	// Success 工具执行成功的标准消息
	Success = `executes successfully. `

	// SuccessWithData 带数据的成功消息模板，指导AI保持原始输出格式
	SuccessWithData = `
executes successfully.
After you have determined the final answer, you should not make changes to the content, do not summarize the content, do not output your thoughts, and only keep the output of the original results.
Keep the output in json format as follows.\n
`
)

// ResParser 解析API响应并格式化为AI可理解的输出
func ResParser(v []byte, chatType domain.AiChatType, err error) (string, error) {
	if err != nil {
		return "", err
	}

	// 解析HTTP响应
	var res httpx.Response
	if err := json.Unmarshal(v, &res); err != nil {
		return "", err
	}

	// 检查响应状态码
	if res.Code != 200 {
		return "", errors.New(res.Msg)
	}

	// 根据聊天类型处理不同的响应格式
	switch chatType {
	case domain.TodoAdd:
		// 待办创建只返回成功消息
		return Success, err
	}

	// 构建标准聊天响应格式
	data := domain.ChatResp{
		ChatType: int(chatType),
		Data:     res.Data,
	}

	// 序列化响应数据
	d, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	// 返回带数据的成功消息
	return SuccessWithData + string(d) + "\n\n\n", nil

}
