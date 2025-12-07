package chatinternal

import (
	"BackEnd/internal/svc"
	"BackEnd/pkg/langchain"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/tools"
)

type AgentChat struct {
	agentsChain chains.Chain
}

func NewAgentChat(svc *svc.ServiceContext, tools []tools.Tool) *AgentChat {
	return &AgentChat{
		agentsChain: agents.NewExecutor(agents.NewOneShotAgent(svc.LLMs, tools, agents.WithPromptPrefix(_defaultMrklPrefix))),
	}
}

func (t *AgentChat) Chains() chains.Chain {
	return chains.NewTransform(t.transform, nil, nil)
}

func (t *AgentChat) transform(ctx context.Context, inputs map[string]any,
	opts ...chains.ChainCallOption) (map[string]any,
	error) {

	for s, a := range inputs {
		if _, ok := a.(string); !ok {
			delete(inputs, s)
		}
	}
	inputs["today"] = time.Now().Format("2006-01-02")
	inputs["history"] = ""

	outPut, err := t.agentsChain.Call(ctx, inputs, opts...)
	if err != nil {
		fmt.Printf("AgentChat execution error: %v\n", err)
		return nil, err
	}
	fmt.Printf("AgentChat output keys: %v\n", outPut)
	v, ok := outPut["output"]
	if !ok {
		return outPut, nil
	}

	text := v.(string)

	withoutJSONStart := strings.Split(text, "```json")
	if !(len(withoutJSONStart) > 1) {
		return map[string]any{
			langchain.Output: v,
		}, err
	}

	withoutJSONEnd := strings.Split(withoutJSONStart[1], "```")
	if len(withoutJSONEnd) < 1 {
		return map[string]any{
			langchain.Output: v,
		}, err
	}

	return map[string]any{
		langchain.Output: withoutJSONEnd[0],
	}, nil
}
