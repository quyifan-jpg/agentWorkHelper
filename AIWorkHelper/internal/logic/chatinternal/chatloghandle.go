/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package chatinternal

import (
	"AIWorkHelper/internal/domain"
	"AIWorkHelper/internal/model"
	"AIWorkHelper/internal/svc"
	"AIWorkHelper/pkg/langchain"
	"AIWorkHelper/pkg/langchain/outputparserx"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/prompts"
	"strings"
	"time"
)

// Task 聊天日志总结任务结构体
type Task struct {
	Type    int    // 任务类型: 1=待办任务, 2=审批事项
	Title   string // 任务标题
	Content string // 任务详细内容
}

// ChatLogHandle 聊天日志处理器，用于总结和分析群聊消息内容
type ChatLogHandle struct {
	svc    *svc.ServiceContext      // 服务上下文，包含数据库连接、LLM客户端等依赖
	chains chains.Chain             // LangChain链，用于调用LLM进行消息总结
	out    outputparserx.Structured // 结构化输出解析器
}

// NewChatLogHandle 创建聊天日志处理器实例
func NewChatLogHandle(svc *svc.ServiceContext) *ChatLogHandle {
	return &ChatLogHandle{
		svc:    svc,
		chains: chains.NewLLMChain(svc.LLMs, prompts.NewPromptTemplate(_defaultChatLogPrompts, []string{"input"})), // 使用预定义的提示词模板创建LLM链
		out:    outputparserx.Structured{},
	}
}

// Name 返回处理器名称，用于路由器识别
func (c *ChatLogHandle) Name() string {
	return "chat_log"
}

// Description 返回处理器描述，路由器根据此描述判断是否使用该处理器
func (c *ChatLogHandle) Description() string {
	return "used to summarize and analyze the content of a chat session"

}

// Chains 返回处理链，使用transform函数处理请求
func (c *ChatLogHandle) Chains() chains.Chain {
	return chains.NewTransform(c.transform, nil, nil)
}

// transform 核心转换函数，负责查询聊天记录并调用LLM进行总结
func (c *ChatLogHandle) transform(ctx context.Context, inputs map[string]any,
	opts ...chains.ChainCallOption) (map[string]any, error) {
	// 从context中提取relationId参数（会话ID，"all"表示群聊）
	var cid string
	if id := ctx.Value("relationId"); id != nil {
		cid, _ = id.(string)
	}
	if cid == "" {
		return nil, errors.New("请确定需要总结的会话对象")
	}

	// 从context中提取时间范围参数（Unix时间戳，单位：秒）
	var startTime, endTime int64
	if st := ctx.Value("startTime"); st != nil {
		if v, ok := st.(int64); ok {
			startTime = v
		}
	}
	if et := ctx.Value("endTime"); et != nil {
		if v, ok := et.(int64); ok {
			endTime = v
		}
	}

	// 如果没有指定时间范围，使用默认值（最近24小时）
	if startTime == 0 && endTime == 0 {
		currentTime := time.Now().Unix()
		startTime = currentTime - 24*3600
		endTime = currentTime
	}

	// 清理inputs，删除所有非字符串类型的键，避免memory组件处理时出错
	for s, a := range inputs {
		if _, ok := a.(string); !ok {
			delete(inputs, s)
		}
	}

	// 查询聊天记录并格式化为文本
	msgs, err := c.chatLog(ctx, cid, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// 使用LLM进行总结，创建新的map只包含input键，避免影响memory
	res, err := chains.Call(ctx, c.chains, map[string]any{
		"input": msgs,
	}, opts...)
	if err != nil {
		return nil, err
	}

	// 从LLM响应中提取文本内容
	text, ok := res[langchain.OutPut].(string)
	if !ok {
		return nil, chains.ErrInvalidOutputValues
	}

	// 解析LLM返回的JSON结果为Task数组
	var data []*Task
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		return nil, err
	}

	// 构建最终响应，包含chatType和data
	b, err := json.Marshal(domain.ChatResp{
		ChatType: domain.ChatLog,
		Data:     data,
	})
	if err != nil {
		return nil, err
	}

	// 返回标准格式的输出给路由器
	return map[string]any{
		langchain.OutPut: string(b),
	}, nil
}

// chatLog 查询聊天记录并格式化为文本，供LLM分析
func (c *ChatLogHandle) chatLog(ctx context.Context, cid string, startTime, endTime int64) (string, error) {
	// 从数据库查询指定时间范围内的聊天记录
	list, err := c.svc.ChatLogModel.ListBySendTime(ctx, cid, startTime, endTime)
	if err != nil {
		return "", err
	}

	// 聊天记录格式：用户名(用户ID): 消息内容
	chatStr := "%s(%s): %s\n"

	var (
		res    strings.Builder                // 用于拼接最终的聊天记录文本
		record = make(map[string]*model.User) // 缓存用户信息，避免重复查询数据库
	)

	// 遍历聊天记录，查询发送者信息并格式化
	for i := range list {
		var u *model.User
		// 先从缓存中查找用户信息
		if v, ok := record[list[i].SendId]; ok {
			u = v
		} else {
			// 缓存中不存在，从数据库查询
			t, err := c.svc.UserModel.FindOne(ctx, list[i].SendId)
			if err != nil {
				return "", err
			}

			u = t
			record[list[i].SendId] = t // 缓存用户信息
		}

		// 格式化聊天记录：用户名(用户ID): 消息内容
		res.Write([]byte(fmt.Sprintf(chatStr, u.Name, u.ID.Hex(), list[i].MsgContent)))
	}

	return res.String(), nil
}
