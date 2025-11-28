/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package memoryx 提供内存组件的配置选项
package memoryx

import "github.com/tmc/langchaingo/callbacks"

// Options 内存组件的配置选项结构体
type Options struct {
	outParser        // 输出解析器，用于处理AI输出内容
	callback callbacks.Handler // 回调处理器，用于监控内存操作过程
}

// Option 配置选项函数类型，用于设置Options的各个字段
type Option func(options *Options)

// newOption 创建新的配置选项实例，应用所有传入的选项函数
func newOption(opts ...Option) *Options {
	opt := &Options{
		callback:  nil,
		outParser: nil,
	}

	for _, o := range opts {
		o(opt)
	}
	return opt
}

// WithCallback 设置回调处理器的选项函数
func WithCallback(handler callbacks.Handler) Option {
	return func(options *Options) {
		options.callback = handler
	}
}

// WithOutParser 设置输出解析器的选项函数
func WithOutParser(outParser outParser) Option {
	return func(options *Options) {
		options.outParser = outParser
	}
}
