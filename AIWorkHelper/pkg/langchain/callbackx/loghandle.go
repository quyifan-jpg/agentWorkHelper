/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package callbackx 提供LangChain回调处理器的实现，用于记录和监控AI对话过程
package callbackx

import (
	"context"
	"encoding/json"
	"gitee.com/dn-jinmin/tlog"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// LogHandle 日志处理器，实现LangChain的回调接口，用于记录AI对话的各个阶段
type LogHandle struct {
	tlog.Logger // 嵌入日志记录器，提供日志记录功能
}

// NewLogHandler 创建新的日志处理器实例
func NewLogHandler(logger tlog.Logger) *LogHandle {
	return &LogHandle{Logger: logger}
}

// HandleText 处理文本输出事件，记录生成的文本内容
func (l *LogHandle) HandleText(ctx context.Context, text string) {
	l.InfoCtx(ctx, "text", text)
}

// HandleLLMStart 处理LLM开始执行事件，记录输入的提示词
func (l *LogHandle) HandleLLMStart(ctx context.Context, prompts []string) {
	l.InfoCtx(ctx, "llm_start", prompts)
}

// HandleLLMGenerateContentStart 处理LLM开始生成内容事件，记录输入的消息内容
func (l *LogHandle) HandleLLMGenerateContentStart(ctx context.Context, ms []llms.MessageContent) {
	l.InfoCtx(ctx, "llm_generate_content_start", l.mustJsonMarshal(ms))
}

// HandleLLMGenerateContentEnd 处理LLM完成内容生成事件，记录生成的响应内容
func (l *LogHandle) HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse) {
	l.InfoCtx(ctx, "llm_generate_content_end", l.mustJsonMarshal(res))
}

// HandleLLMError 处理LLM执行错误事件，记录错误信息
func (l *LogHandle) HandleLLMError(ctx context.Context, err error) {
	l.ErrorCtx(ctx, "llm_error", err.Error())
}

// HandleChainStart 处理链式调用开始事件，记录输入参数
func (l *LogHandle) HandleChainStart(ctx context.Context, inputs map[string]any) {
	l.InfoCtx(ctx, "chain_start", l.mustJsonMarshal(inputs))
}

// HandleChainEnd 处理链式调用结束事件，记录输出结果
func (l *LogHandle) HandleChainEnd(ctx context.Context, outputs map[string]any) {
	l.InfoCtx(ctx, "chain_end", l.mustJsonMarshal(outputs))
}

// HandleChainError 处理链式调用错误事件，记录错误信息
func (l *LogHandle) HandleChainError(ctx context.Context, err error) {
	l.ErrorCtx(ctx, "chain_error", err.Error())
}

// HandleToolStart 处理工具调用开始事件，记录输入内容
func (l *LogHandle) HandleToolStart(ctx context.Context, input string) {
	l.InfoCtx(ctx, "tool_start", input)
}

// HandleToolEnd 处理工具调用结束事件，记录输出内容
func (l *LogHandle) HandleToolEnd(ctx context.Context, output string) {
	l.InfoCtx(ctx, "tool_end", output)
}

// HandleToolError 处理工具调用错误事件，记录错误信息
func (l *LogHandle) HandleToolError(ctx context.Context, err error) {
	l.ErrorCtx(ctx, "tool_error", err.Error())
}

// HandleAgentAction 处理智能体动作事件，记录智能体执行的动作
func (l *LogHandle) HandleAgentAction(ctx context.Context, action schema.AgentAction) {
	l.InfoCtx(ctx, "agent_action", l.mustJsonMarshal(action))
}

// HandleAgentFinish 处理智能体完成事件，记录智能体的最终结果
func (l *LogHandle) HandleAgentFinish(ctx context.Context, finish schema.AgentFinish) {
	l.InfoCtx(ctx, "agent_finish", l.mustJsonMarshal(finish))
}

// HandleRetrieverStart 处理检索器开始事件，记录查询内容
func (l *LogHandle) HandleRetrieverStart(ctx context.Context, query string) {
	l.InfoCtx(ctx, "retriever_start", query)
}

// HandleRetrieverEnd 处理检索器结束事件，记录查询结果和检索到的文档
func (l *LogHandle) HandleRetrieverEnd(ctx context.Context, query string, documents []schema.Document) {
	l.InfofCtx(ctx, "retriever_end", "query %s, documents %s", query, l.mustJsonMarshal(documents))
}

// HandleStreamingFunc 处理流式输出事件，记录流式数据块（当前已注释，避免日志过多）
func (l *LogHandle) HandleStreamingFunc(ctx context.Context, chunk []byte) {
	//l.InfoCtx(ctx, "streaming_func", string(chunk))
}

// mustJsonMarshal 将任意对象序列化为JSON字符串，失败时返回空字符串
func (l *LogHandle) mustJsonMarshal(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
