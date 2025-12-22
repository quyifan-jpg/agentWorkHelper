// Package router 提供路由决策的提示词模板和输出解析器
package router

import (
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/outputparser"
	"github.com/tmc/langchaingo/prompts"
)

const (
	_formatting   = "formatting"   // 格式化指令的模板变量名
	_destinations = "destinations" // 目标处理器列表的模板变量名
	_input        = "input"        // 用户输入的模板变量名
	_nextInput    = "next_inputs"  // 处理后输入的模板变量名

	// MULTI_PROMPT_ROUTER_TEMPLATE 多提示词路由器的模板，用于指导LLM选择合适的处理器
	MULTI_PROMPT_ROUTER_TEMPLATE = `Given a raw text input to a language model select the model prompt best suited for
the input. You will be given the names of the available prompts and a description of
what the prompt is best suited for. You may also revise the original input if you
think that revising it will ultimately lead to a better response from the language
model.

<< FORMATTING >>
Return a markdown code snippet with a JSON object formatted to look like:
{{.formatting}}

REMEMBER: "destination" MUST be one of the candidate prompt names specified below OR
it can be "DEFAULT" if the input is not well suited for any of the candidate prompts.
REMEMBER: "next_inputs" can just be the original input if you don't think any
modifications are needed.

<< CANDIDATE PROMPTS >>
{{.destinations}}

<< INPUT >>
{{.input}}

IMPORTANT: Return ONLY the JSON code snippet. Do not include any "Thought", reasoning, or extra text.
`
)

var (
	// _outputparser 结构化输出解析器，用于解析LLM返回的路由决策JSON
	_outputparser = outputparser.NewStructured([]outputparser.ResponseSchema{
		{
			Name:        _destinations,
			Description: `name of the question answering system to use or "DEFAULT"`,
		}, {
			Name:        _nextInput,
			Description: `a potentially modified version of the original input`,
		},
	})
)

// createPrompt 根据处理器列表创建路由决策的提示词模板
func createPrompt(handler []Handler) prompts.PromptTemplate {
	return prompts.PromptTemplate{
		Template:       MULTI_PROMPT_ROUTER_TEMPLATE,
		InputVariables: []string{_input},
		TemplateFormat: prompts.TemplateFormatGoTemplate,
		PartialVariables: map[string]any{
			_destinations: HandlerDestinations(handler),          // 生成处理器列表描述
			_formatting:   _outputparser.GetFormatInstructions(), // 生成JSON格式指令
		},
	}
}

// HandlerDestinations 将处理器列表转换为描述字符串，供LLM理解各处理器的用途
func HandlerDestinations(handler []Handler) string {
	var hs strings.Builder
	for _, h := range handler {
		hs.WriteString(fmt.Sprintf("- %s: %s\n", h.Name(), h.Description()))
	}

	return hs.String()
}
