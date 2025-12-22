package chatinternal

import (
	"BackEnd/internal/svc"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/prompts"
)

type ChatHandle struct {
	*BaseChat
}

func NewChatHandle(svc *svc.ServiceContext, toolDescriptions string) *ChatHandle {
	var baseChat *BaseChat
	if svc.LLMs != nil {
		// Use custom prompt if tool descriptions are provided
		if toolDescriptions != "" {
			prompt := prompts.NewPromptTemplate(
				ChatWithToolsTemplate,
				[]string{"history", "input"},
			)
			// Inject tool descriptions into the template partials or directly check if template supports it.
			// Ideally, we should use PartialVariables for tool_descriptions.
			// Re-creating prompt with partials:
			prompt = prompts.NewPromptTemplate(
				ChatWithToolsTemplate,
				[]string{"history", "input"},
			)
			prompt.PartialVariables = map[string]any{
				"tool_descriptions": toolDescriptions,
			}
			// When using manual memory with Chat method in basechat.go,
			// the "history" variable is loaded from memory and added to inputs.
			// However, langchaingo's LLMChain validation checks if all InputVariables are present.
			// Since "history" is supplied by memory (manually or automatically), it might be missing from initial inputs.
			// But here we are manually handling memory in BaseChat.Chat() which ADDS "history" to inputs map.

			// Wait, the error "missing key in input values: history" comes from chains.Call -> LLMChain.Call.
			// In BaseChat.Chat:
			// 1. We create inputs map with "input".
			// 2. We load memory variables. Default memory key is "history".
			// 3. We add memory variables to inputs map.
			// 4. We call chain.Call(ctx, inputs).
			// So "history" SHOULD be there if memory loads correctly.

			// Check memory initialization in NewBaseChatFromLLMWithPrompt.
			// mem := memory.NewConversationBuffer() -> Default InputKey is empty, OutputKey is empty.
			// MemoryKey is "history".

			baseChat = NewBaseChatFromLLMWithPrompt(svc.LLMs, prompt)
		} else {
			baseChat = NewBaseChatFromLLM(svc.LLMs)
		}
	} else {
		// Fallback if LLM not in context (shouldn't happen if initialized correctly)
		var err error
		baseChat, err = NewBaseChat(svc.Config.AI.ApiKey, svc.Config.AI.BaseURL, svc.Config.AI.Model)
		if err != nil {
			return nil
		}
	}
	return &ChatHandle{
		BaseChat: baseChat,
	}
}

func (t *ChatHandle) Name() string {
	return "chat"
}

func (t *ChatHandle) Description() string {
	return "suitable for general chat, answering questions about system capabilities (e.g. 'what tools do you have?'), writing, coding, translation, etc. Use this handler if the request does not clearly fit other specialized handlers."
}

func (t *ChatHandle) Chains() chains.Chain {
	return t.chain
}
