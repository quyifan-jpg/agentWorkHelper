/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package memoryx 提供对话摘要的提示词模板
package memoryx

import "github.com/tmc/langchaingo/prompts"

const (
	// _DEFAULT_SUMMARIZER_TEMPLATE 默认的对话摘要提示词模板，用于将长对话压缩成简洁的摘要
	_DEFAULT_SUMMARIZER_TEMPLATE = `Progressively summarize the lines of conversation provided, adding onto the previous summary returning a new summary.

EXAMPLE
Current summary:
The human asks what the AI thinks of artificial intelligence. The AI thinks artificial intelligence is a force for good.

New lines of conversation:
Human: Why do you think artificial intelligence is a force for good?
AI: Because artificial intelligence will help humans reach their full potential.

New summary:
The human asks what the AI thinks of artificial intelligence. The AI thinks artificial intelligence is a force for good because it will help humans reach their full potential.
END OF EXAMPLE

Current summary:
{{.summary}}

New lines of conversation:
{{.new_lines}}

New summary:`
)

// createSummaryPrompt 创建对话摘要的提示词模板，包含当前摘要和新对话内容两个变量
func createSummaryPrompt() prompts.PromptTemplate {
	return prompts.NewPromptTemplate(_DEFAULT_SUMMARIZER_TEMPLATE, []string{
		"summary", "new_lines",
	})
}
