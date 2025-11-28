/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package chatinternal

import (
	"AIWorkHelper/internal/svc"
	"AIWorkHelper/pkg/langchain"
	"context"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/tools"
	"strings"
)

type baseChat struct {
	agentsChain chains.Chain
}

func NewBaseChat(svc *svc.ServiceContext, tools []tools.Tool) *baseChat {
	return &baseChat{
		agentsChain: agents.NewExecutor(agents.NewOneShotAgent(svc.LLMs, tools, agents.WithPromptPrefix(_defaultMrklPrefix))),
	}
}

func (t *baseChat) Chains() chains.Chain {
	return chains.NewTransform(t.transform, nil, nil)
}

func (t *baseChat) transform(ctx context.Context, inputs map[string]any,
	opts ...chains.ChainCallOption) (map[string]any,
	error) {

	for s, a := range inputs {
		if _, ok := a.(string); !ok {
			delete(inputs, s)
		}
	}

	outPut, err := t.agentsChain.Call(ctx, inputs, opts...)
	if err != nil {
		return nil, err
	}
	v, ok := outPut["output"]
	if !ok {
		return outPut, nil
	}

	text := v.(string)

	withoutJSONStart := strings.Split(text, "```json")
	if !(len(withoutJSONStart) > 1) {
		return map[string]any{
			langchain.OutPut: v,
		}, err
	}

	withoutJSONEnd := strings.Split(withoutJSONStart[1], "```")
	if len(withoutJSONEnd) < 1 {
		return map[string]any{
			langchain.OutPut: v,
		}, err
	}

	return map[string]any{
		langchain.OutPut: withoutJSONEnd[0],
	}, nil
}
