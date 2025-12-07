package chatinternal

import (
	"BackEnd/internal/svc"

	"github.com/tmc/langchaingo/chains"
)

type ChatHandle struct {
	*BaseChat
}

func NewChatHandle(svc *svc.ServiceContext) *ChatHandle {
	var baseChat *BaseChat
	if svc.LLMs != nil {
		baseChat = NewBaseChatFromLLM(svc.LLMs)
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
	return "suitable for general chat, answering questions, writing, coding, translation, etc. If the user's intent is not todo, use this handler."
}

func (t *ChatHandle) Chains() chains.Chain {
	return t.chain
}
