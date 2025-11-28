/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package chatinternal

import (
	"AIWorkHelper/internal/domain"
	"AIWorkHelper/internal/svc"
	"AIWorkHelper/pkg/langchain"
	"fmt"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/prompts"
)

// DefaultHandler 默认聊天处理器，用于处理通用AI对话
type DefaultHandler struct {
	svc *svc.ServiceContext // 服务上下文
	c   chains.Chain        // LangChain链
}

// NewDefaultHandler 创建默认聊天处理器
func NewDefaultHandler(svc *svc.ServiceContext) *DefaultHandler {
	// 定义AI助手的基础提示词
	template := "you are an all-round assistant, please help me answer this question: \n\n<< input >>\n{{.input}}"

	// 构建提示词模板
	prompt := prompts.PromptTemplate{
		Template:       BASE_PROMPAT_TEMPLATE + template + "\n\n" + OUT_PROMPT_TEMPLATE,
		InputVariables: []string{langchain.Input},
		TemplateFormat: prompts.TemplateFormatGoTemplate,
		PartialVariables: map[string]any{
			"chatType": fmt.Sprintf("%d", domain.DefaultHandler),
			"data":     "solution",
		},
	}

	return &DefaultHandler{
		svc: svc,
		c:   chains.NewLLMChain(svc.LLMs, prompt, chains.WithCallback(svc.Callbacks)),
	}
}

// Name 返回处理器名称
func (d *DefaultHandler) Name() string {
	return "default"
}

// Description 返回处理器描述
func (d *DefaultHandler) Description() string {
	return "suitable for answering multiple questions"
}

// Chains 返回LangChain链实例
func (d *DefaultHandler) Chains() chains.Chain {
	return d.c
}
