package chatinternal

import (
	"context"
	"strings"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/schema"
)

type BaseChat struct {
	llm    llms.Model
	memory schema.Memory
	chain  chains.Chain
}

type baseChat struct {
	agentsChain chains.Chain
}

func NewBaseChat(apiKey, baseURL, modelName string) (*BaseChat, error) {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if modelName == "" {
		modelName = "gpt-3.5-turbo"
	}

	llm, err := openai.New(
		openai.WithToken(apiKey),
		openai.WithBaseURL(baseURL),
		openai.WithModel(modelName),
	)
	if err != nil {
		return nil, err
	}

	mem := memory.NewConversationBuffer()

	// Create a simple conversation chain
	chain := chains.NewConversation(llm, mem)

	return &BaseChat{
		llm:    llm,
		memory: mem,
		chain:  chain,
	}, nil
}

func NewBaseChatFromLLM(llm llms.Model) *BaseChat {
	mem := memory.NewConversationBuffer()
	chain := chains.NewConversation(llm, mem)

	return &BaseChat{
		llm:    llm,
		memory: mem,
		chain:  chain,
	}
}

func (b *BaseChat) Chat(ctx context.Context, input string) (string, error) {
	res, err := chains.Call(ctx, b.chain, map[string]any{
		"input": input,
	})
	if err != nil {
		return "", err
	}

	return res["text"].(string), nil
}

// SaveContext saves context (like file content) to memory
func (b *BaseChat) SaveContext(ctx context.Context, inputs map[string]any, outputs map[string]any) error {
	return b.memory.SaveContext(ctx, inputs, outputs)
}

// ClearMemory clears the conversation history
func (b *BaseChat) ClearMemory(ctx context.Context) error {
	return b.memory.Clear(ctx)
}

// ParseJSONOutput helper to parse JSON from AI response (if needed)
func (b *BaseChat) ParseJSONOutput(text string) string {
	withoutJSONStart := strings.Split(text, "```json")
	if len(withoutJSONStart) > 1 {
		withoutJSONEnd := strings.Split(withoutJSONStart[1], "```")
		if len(withoutJSONEnd) > 0 {
			return strings.TrimSpace(withoutJSONEnd[0])
		}
	}
	return text
}

// GetLLM returns the underlying LLM model
func (b *BaseChat) GetLLM() llms.Model {
	return b.llm
}
