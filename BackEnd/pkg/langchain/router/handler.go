// Package router 提供智能路由功能，根据用户输入自动选择合适的处理器
package router

import "github.com/tmc/langchaingo/chains"

// Handler 路由处理器接口，定义了处理特定类型请求的处理器规范
type Handler interface {
	Name() string        // 返回处理器的名称，用于路由识别
	Description() string // 返回处理器的描述，用于AI理解处理器的用途
	Chains() chains.Chain // 返回处理器对应的LangChain链，用于实际处理请求
}
