// Package router 提供路由器的配置选项
package router

import (
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

// Options 路由器的配置选项结构体
type Options struct {
	prompt       prompts.PromptTemplate // 提示词模板，用于生成路由决策的提示
	memory       schema.Memory          // 内存组件，用于存储对话历史
	callback     callbacks.Handler      // 回调处理器，用于监控路由过程
	emptyHandler Handler                // 默认处理器，当没有合适的处理器时使用
}

// Option 配置选项函数类型，用于设置Options的各个字段
type Option func(options *Options)

// executorDefaultOptions 创建默认的路由器配置选项
func executorDefaultOptions(handler []Handler) Options {
	return Options{
		prompt: createPrompt(handler), // 根据处理器列表创建路由提示词
		memory: memory.NewSimple(),    // 使用简单内存实现
	}
}

// WithMemory 设置内存组件的选项函数
func WithMemory(m schema.Memory) Option {
	return func(options *Options) {
		options.memory = m
	}
}

// WithEmptyHandler 设置默认处理器的选项函数
func WithEmptyHandler(emptyHandler Handler) Option {
	return func(options *Options) {
		options.emptyHandler = emptyHandler
	}
}

// Withcallback 设置回调处理器的选项函数
func Withcallback(callback callbacks.Handler) Option {
	return func(options *Options) {
		options.callback = callback
	}
}
