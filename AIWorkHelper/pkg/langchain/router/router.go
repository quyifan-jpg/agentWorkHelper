/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package router 提供智能路由器实现，根据用户输入自动选择合适的处理器
package router

import (
	"context"
	"errors"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/outputparser"
	"github.com/tmc/langchaingo/schema"
)

const Empty = "DEFAULT" // 默认处理器标识，当没有合适的处理器时使用

var ErrNotHandles = errors.New("不存在合适的handler") // 没有找到合适处理器的错误

// Router 智能路由器，使用LLM分析用户输入并选择最合适的处理器
type Router struct {
	handlers     map[string]Handler        // 处理器映射表，键为处理器名称
	chain        chains.Chain              // LLM链，用于分析输入并做出路由决策
	callbacks    callbacks.Handler         // 回调处理器，用于监控路由过程
	memory       schema.Memory             // 内存组件，用于存储对话历史
	outputparser outputparser.Structured   // 输出解析器，用于解析LLM的路由决策结果
	emptyHandle  Handler                   // 默认处理器，当没有合适处理器时使用
}

// NewRouter 创建新的智能路由器实例
func NewRouter(llm llms.Model, handler []Handler, opts ...Option) *Router {
	opt := executorDefaultOptions(handler)
	for _, o := range opts {
		o(&opt)
	}

	// 构建处理器映射表
	hs := make(map[string]Handler, len(handler))
	for _, h := range handler {
		hs[h.Name()] = h
	}

	return &Router{
		handlers:     hs,
		chain:        chains.NewLLMChain(llm, opt.prompt),
		callbacks:    opt.callback,
		memory:       opt.memory,
		emptyHandle:  opt.emptyHandler,
		outputparser: _outputparser,
	}
}

// Call 执行路由调用，分析输入并选择合适的处理器进行处理
func (r *Router) Call(ctx context.Context, inputs map[string]any, options ...chains.ChainCallOption) (map[string]any, error) {
	// 触发链开始回调
	if r.callbacks != nil {
		r.callbacks.HandleChainStart(ctx, inputs)
	}

	// 如果没有注册任何处理器，使用默认处理器或返回错误
	if len(r.handlers) == 0 {
		if r.emptyHandle != nil {
			return chains.Call(ctx, r.emptyHandle.Chains(), inputs)
		} else {
			return nil, ErrNotHandles
		}
	}

	// 使用LLM分析输入并做出路由决策
	result, err := chains.Call(ctx, r.chain, inputs, options...)
	if err != nil {
		return nil, err
	}

	// 提取LLM的文本输出
	text, ok := result["text"]
	if !ok {
		return nil, chains.ErrNotFound
	}

	// 解析LLM输出，获取路由决策结果
	out, err := r.outputparser.Parse(text.(string))
	if err != nil {
		return nil, err
	}

	// 触发链结束回调
	if r.callbacks != nil {
		r.callbacks.HandleChainEnd(ctx, map[string]any{
			"out": out,
		})
	}

	// 根据路由决策选择对应的处理器
	data := out.(map[string]string)
	next, ok := data[_destinations]
	if !ok || next == Empty || r.handlers[next] == nil {
		// 如果没有找到合适的处理器，使用默认处理器
		if r.emptyHandle != nil {
			return chains.Call(ctx, r.emptyHandle.Chains(), inputs)
		} else {
			return nil, ErrNotHandles
		}
	}

	// 调用选定的处理器
	return chains.Call(ctx, r.handlers[next].Chains(), inputs)
}

// GetMemory 获取路由器的内存组件
func (r *Router) GetMemory() schema.Memory {
	return r.memory
}

// GetInputKeys 获取输入键列表（当前返回nil，表示接受任意输入）
func (r *Router) GetInputKeys() []string {
	return nil
}

// GetOutputKeys 获取输出键列表（当前返回nil，表示输出格式由具体处理器决定）
func (r *Router) GetOutputKeys() []string {
	return nil
}
