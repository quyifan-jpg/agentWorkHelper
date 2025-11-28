/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package memoryx 提供多会话内存管理功能，支持为不同聊天会话维护独立的对话记忆
package memoryx

import (
	"context"
	"github.com/tmc/langchaingo/schema"
	"AIWorkHelper/pkg/langchain"
	"sync"
)

// Memoryx 多会话内存管理器，为不同的聊天会话维护独立的对话记忆
type Memoryx struct {
	sync.Mutex                           // 互斥锁，保证并发安全
	getMemory     func() schema.Memory   // 内存创建函数，用于为新会话创建内存实例
	memorys       map[string]schema.Memory // 会话ID到内存实例的映射，存储各会话的对话记忆
	defaultMemory schema.Memory          // 默认内存实例，用于没有指定会话ID的情况
}

// NewMemoryx 创建新的多会话内存管理器实例
func NewMemoryx(handle func() schema.Memory) *Memoryx {
	return &Memoryx{
		getMemory:     handle,
		memorys:       make(map[string]schema.Memory),
		defaultMemory: handle(),
	}
}

// GetMemoryKey 获取当前会话的内存键名
func (s *Memoryx) GetMemoryKey(ctx context.Context) string {
	return s.memory(ctx).GetMemoryKey(ctx)
}

// MemoryVariables 获取当前会话内存动态加载的输入键列表
func (s *Memoryx) MemoryVariables(ctx context.Context) []string {
	return s.memory(ctx).MemoryVariables(ctx)
}

// LoadMemoryVariables 根据输入参数加载当前会话的内存变量，返回键值对映射
func (s *Memoryx) LoadMemoryVariables(ctx context.Context, inputs map[string]any) (map[string]any, error) {
	return s.memory(ctx).LoadMemoryVariables(ctx, inputs)
}

// SaveContext 将当前模型运行的上下文保存到对应会话的内存中
func (s *Memoryx) SaveContext(ctx context.Context, inputs map[string]any, outputs map[string]any) error {
	return s.memory(ctx).SaveContext(ctx, inputs, outputs)
}

// Clear 清空当前会话的内存内容
func (s *Memoryx) Clear(ctx context.Context) error {
	return s.memory(ctx).Clear(ctx)
}

// memory 根据上下文中的聊天ID获取对应的内存实例，如果不存在则创建新的
func (s *Memoryx) memory(ctx context.Context) schema.Memory {
	s.Lock()   // 加锁保证并发安全
	defer s.Unlock()

	var chatId string
	v := ctx.Value(langchain.ChatId) // 从上下文中获取聊天会话ID
	if v == nil {
		return s.defaultMemory // 如果没有会话ID，返回默认内存实例
	}

	chatId = v.(string)
	memory, ok := s.memorys[chatId] // 查找该会话ID对应的内存实例
	if !ok {
		memory = s.getMemory()        // 如果不存在，创建新的内存实例
		s.memorys[chatId] = memory    // 将新实例存储到映射中
	}

	return memory
}