/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package memoryx 提供带Token限制的摘要缓冲区内存实现
package memoryx

import (
	"context"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
)

// outParser 输出解析器函数类型，用于处理AI输出内容
type outParser func(ctx context.Context, input string) string

// SummaryBuffer 摘要缓冲区内存，当对话超过Token限制时自动生成摘要
type SummaryBuffer struct {
	outParser                         // 输出解析器，用于处理AI输出
	*memory.ConversationBuffer        // 嵌入对话缓冲区，提供基础对话存储功能
	chain         chains.Chain        // LLM链，用于生成对话摘要
	callback      callbacks.Handler   // 回调处理器，用于监控操作过程
	MaxTokenLimit int                 // 最大Token限制，超过此限制时触发摘要生成
	buffer        llms.ChatMessage    // 摘要缓冲区，存储生成的摘要内容
}

// NewSummaryBuffer 创建新的摘要缓冲区内存实例
func NewSummaryBuffer(llms llms.Model, maxTokenLimit int, opts ...Option) *SummaryBuffer {
	opt := newOption(opts...)

	return &SummaryBuffer{
		ConversationBuffer: memory.NewConversationBuffer(),
		chain:              chains.NewLLMChain(llms, createSummaryPrompt(), chains.WithCallback(opt.callback)),
		callback:           opt.callback,
		MaxTokenLimit:      maxTokenLimit,
		buffer:             nil,
		outParser:          opt.outParser,
	}
}

// GetMemoryKey 获取内存键名
func (s *SummaryBuffer) GetMemoryKey(ctx context.Context) string {
	return s.ConversationBuffer.GetMemoryKey(ctx)
}

// MemoryVariables 获取内存动态加载的输入键列表
func (s *SummaryBuffer) MemoryVariables(ctx context.Context) []string {
	return s.ConversationBuffer.MemoryVariables(ctx)
}

// LoadMemoryVariables 加载内存变量，将摘要和对话历史合并返回
func (s *SummaryBuffer) LoadMemoryVariables(ctx context.Context, inputs map[string]any) (map[string]any, error) {
	var (
		res []llms.ChatMessage
		err error
	)
	// 如果存在摘要缓冲区，先添加摘要内容
	if s.buffer != nil {
		res = append(res, s.buffer)
	}

	// 获取对话历史消息
	message, err := s.ChatHistory.Messages(ctx)
	if err != nil {
		return nil, err
	}
	res = append(res, message...)

	// 将所有消息转换为字符串格式
	bufferString, err := llms.GetBufferString(res, s.HumanPrefix, s.AIPrefix)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		s.MemoryKey: bufferString,
	}, nil
}

// SaveContext 保存对话上下文到内存，当Token超限时自动生成摘要
func (s *SummaryBuffer) SaveContext(ctx context.Context, inputs map[string]any, outputs map[string]any) error {
	// 保存用户输入消息
	userInputValue, err := memory.GetInputValue(inputs, s.InputKey)
	if err != nil {
		return err
	}
	err = s.ChatHistory.AddUserMessage(ctx, userInputValue)
	if err != nil {
		return err
	}

	// 保存AI输出消息
	aiOutPutValue, err := memory.GetInputValue(outputs, s.OutputKey)
	if err != nil {
		return err
	}
	// 如果配置了输出解析器，先处理AI输出
	if s.outParser != nil {
		aiOutPutValue = s.outParser(ctx, aiOutPutValue)
	}

	err = s.ChatHistory.AddAIMessage(ctx, aiOutPutValue)
	if err != nil {
		return err
	}

	// 检查Token使用量是否超过限制
	messages, err := s.ChatHistory.Messages(ctx)
	if err != nil {
		return err
	}
	bufferString, err := llms.GetBufferString(messages, s.ConversationBuffer.HumanPrefix, s.ConversationBuffer.AIPrefix)
	if err != nil {
		return err
	}

	// 如果没有超过Token限制，直接返回
	if llms.CountTokens("", bufferString) <= s.MaxTokenLimit {
		return nil
	}

	// 超过限制时生成摘要
	var newLines string
	if s.buffer != nil {
		newLines = s.buffer.GetContent()
	}

	// 使用LLM生成新的摘要
	newSummary, err := chains.Predict(ctx, s.chain, map[string]any{
		"summary":   bufferString,
		"new_lines": newLines,
	})
	if err != nil {
		return err
	}

	// 将摘要保存到缓冲区，清空对话历史
	s.buffer = &llms.SystemChatMessage{
		Content: newSummary,
	}

	return s.ChatHistory.SetMessages(ctx, nil)
}

// Clear 清空内存内容，包括摘要缓冲区和对话历史
func (s *SummaryBuffer) Clear(ctx context.Context) error {
	s.buffer = nil // 清空摘要缓冲区
	return s.ConversationBuffer.Clear(ctx)
}
