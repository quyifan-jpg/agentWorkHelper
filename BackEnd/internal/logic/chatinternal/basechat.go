package chatinternal

import (
	"context"
	"strings"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

type BaseChat struct {
	llm          llms.Model
	memory       schema.Memory
	chain        chains.Chain
	manualMemory bool
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

func NewBaseChatFromLLMWithPrompt(llm llms.Model, prompt prompts.PromptTemplate) *BaseChat {
	mem := memory.NewConversationBuffer()
	// Create an LLMChain with the custom prompt.
	// Note regarding memory: LLMChain usually doesn't automatically load/save context unless configured.
	// We will handle memory manually in the Chat method if manualMemory is true.
	chain := chains.NewLLMChain(llm, prompt)
	chain.Memory = mem

	return &BaseChat{
		llm:          llm,
		memory:       mem,
		chain:        chain,
		manualMemory: true,
	}
}

func (b *BaseChat) Chat(ctx context.Context, input string) (string, error) {
	inputs := map[string]any{
		"input": input,
	}

	if b.manualMemory {
		// Load history manually
		history, err := b.memory.LoadMemoryVariables(ctx, inputs)
		if err != nil {
			return "", err
		}

		// Ensure history key exists even if map is empty
		historyContent, ok := history["history"]
		if !ok {
			inputs["history"] = ""
		} else {
			inputs["history"] = historyContent
		}

		for k, v := range history {
			inputs[k] = v
		}
	}

	res, err := chains.Call(ctx, b.chain, inputs)
	if err != nil {
		return "", err
	}

	text := res["text"].(string)

	if b.manualMemory {
		// Save history manually
		if err := b.memory.SaveContext(ctx, map[string]any{"input": input}, map[string]any{"text": text}); err != nil {
			return "", err
		}
	}

	return text, nil
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
