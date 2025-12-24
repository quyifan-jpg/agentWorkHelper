package toolx

import (
	"BackEnd/internal/svc"
	"context"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/redisvector"
)

type KnowledgeRetrievalQA struct {
	svc      *svc.ServiceContext
	Callback callbacks.Handler
	store    *redisvector.Store
	qa       chains.Chain
}

func NewKnowledgeRetrievalQA(svc *svc.ServiceContext) *KnowledgeRetrievalQA {
	return &KnowledgeRetrievalQA{svc: svc}
}

func (k *KnowledgeRetrievalQA) Name() string {
	return "knowledge_retrieval_qa"
}

func (k *KnowledgeRetrievalQA) Description() string {
	return `a knowledge retrieval interface.
use it when you need to inquire about work-related policies, such as employee manuals, attendance rules, etc.
keep Chinese output.`
}

func (k *KnowledgeRetrievalQA) Call(ctx context.Context, input string) (string, error) {
	var err error
	if k.qa == nil {
		k.store, err = getKnowledgeStore(ctx, k.svc)
		if err != nil {
			return "", err
		}

		k.qa = chains.NewRetrievalQAFromLLM(k.svc.LLMs, vectorstores.ToRetriever(k.store, 1))
	}

	res, err := chains.Predict(ctx, k.qa, map[string]any{
		"query": input,
	})
	if err != nil {
		return "", err
	}

	return `The following are the consultation results. When outputting, please output the results directly, do not make summaries, keep them in Chinese, and only do the original output:
\n` + res, nil
}
