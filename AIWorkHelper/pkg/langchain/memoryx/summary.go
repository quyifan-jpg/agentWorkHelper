/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package memoryx 提供对话摘要内存实现
package memoryx

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
)

// Summary 对话摘要内存，将长对话自动压缩成摘要以节省Token消耗
type Summary struct {
	*memory.ConversationBuffer        // 嵌入对话缓冲区，提供基础的对话存储功能
	callback callbacks.Handler       // 回调处理器，用于监控摘要生成过程
	chain    chains.Chain            // LLM链，用于生成对话摘要
}

// NewSummary 创建新的对话摘要内存实例
func NewSummary(llm llms.Model, opts ...Option) *Summary {
	opt := newOption(opts...)
	return &Summary{
		callback:           opt.callback,
		ConversationBuffer: memory.NewConversationBuffer(),
		chain:              chains.NewLLMChain(llm, createSummaryPrompt(), chains.WithCallback(opt.callback)),
	}
}

// GetMemoryKey 获取内存键名
func (s *Summary) GetMemoryKey(ctx context.Context) string {
	return s.ConversationBuffer.GetMemoryKey(ctx)
}

// MemoryVariables 获取内存动态加载的输入键列表
func (s *Summary) MemoryVariables(ctx context.Context) []string {
	return s.ConversationBuffer.MemoryVariables(ctx)
}

// LoadMemoryVariables 根据输入参数加载内存变量，返回键值对映射
func (s *Summary) LoadMemoryVariables(ctx context.Context, inputs map[string]any) (map[string]any, error) {
	return s.ConversationBuffer.LoadMemoryVariables(ctx, inputs)
}

// SaveContext 将当前模型运行的上下文保存到内存，自动生成对话摘要
func (s *Summary) SaveContext(ctx context.Context, inputs map[string]any, outputs map[string]any) error {
	// 获取当前的内存内容（摘要）
	message, err := s.LoadMemoryVariables(ctx, inputs)
	if err != nil {
		return err
	}
	summary := message[s.MemoryKey]

	// 构建新的对话内容
	userInputValue, err := memory.GetInputValue(inputs, s.InputKey)
	if err != nil {
		return err
	}
	userOutPutValue, err := memory.GetInputValue(outputs, s.OutputKey)
	if err != nil {
		return err
	}
	newLines := fmt.Sprintf("Homan: %s\nAi: %s", userInputValue, userOutPutValue)

	// 使用LLM生成新的摘要
	newSummary, err := chains.Predict(ctx, s.chain, map[string]any{
		"summary":   summary,
		"new_lines": newLines,
	})
	if err != nil {
		return err
	}

	// 将新摘要保存为系统消息
	return s.ChatHistory.SetMessages(ctx, []llms.ChatMessage{
		llms.SystemChatMessage{Content: newSummary},
	})
}

// Clear 清空内存内容
func (s *Summary) Clear(ctx context.Context) error {
	return s.ConversationBuffer.Clear(ctx)
}
