/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package callbackx 提供Token计数回调处理器
package callbackx

import (
	"context"
	"gitee.com/dn-jinmin/tlog"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
)

// TitTokenHandle Token计数处理器，用于统计和记录LLM调用消耗的Token数量
type TitTokenHandle struct {
	callbacks.SimpleHandler // 嵌入简单回调处理器，提供默认实现
	logger                  tlog.Logger // 日志记录器，用于记录Token使用情况
}

// NewTitTokenHandle 创建新的Token计数处理器实例
func NewTitTokenHandle(logger tlog.Logger) *TitTokenHandle {
	return &TitTokenHandle{
		SimpleHandler: callbacks.SimpleHandler{},
		logger:        logger,
	}
}

// HandleLLMGenerateContentEnd 处理LLM内容生成结束事件，统计并记录消耗的Token数量
func (l *TitTokenHandle) HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse) {
	var count int
	// 遍历所有选择项，累计Token消耗数量
	for i, _ := range res.Choices {
		if v, ok := res.Choices[i].GenerationInfo["TotalTokens"]; ok {
			count += v.(int)
		}
	}

	// 如果没有Token消耗记录，直接返回
	if count == 0 {
		return
	}

	// 记录Token消耗数量到日志
	l.logger.InfofCtx(ctx, "TitTokenHandle:llm_generate_content_end", "count %d", count)
}
