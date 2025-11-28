/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package toolx

import (
	"AIWorkHelper/internal/svc"
	"AIWorkHelper/pkg/langchain/outputparserx"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/vectorstores/redisvector"
)

type KnowledgeUpdate struct {
	svc          *svc.ServiceContext
	Callback     callbacks.Handler
	outPutParser outputparserx.Structured
	store        *redisvector.Store
}

func NewKnowledgeUpdate(svc *svc.ServiceContext) *KnowledgeUpdate {
	return &KnowledgeUpdate{
		svc:      svc,
		Callback: svc.Callbacks,
		outPutParser: outputparserx.NewStructured([]outputparserx.ResponseSchema{
			{
				Name:        "path",
				Description: "the path to file",
			}, {
				Name:        "name",
				Description: "the name to file",
			}, {
				Name:        "time",
				Description: "file update time",
			},
		}),
	}
}

func (k *KnowledgeUpdate) Name() string {
	return "knowledge_update"
}

func (k *KnowledgeUpdate) Description() string {
	return `a knowledge base update interface.
use when you need to update knowledge base content.
your output should be in the following json format:
{"path": "file path", "name": "file name", "time": "update time"}`
}

func (k *KnowledgeUpdate) Call(ctx context.Context, input string) (string, error) {
	if err := k.svc.Auth(ctx); err != nil {
		return "", err
	}

	var data any
	data, err := k.outPutParser.Parse(input)
	if err != nil {
		// ```json str

		t := make(map[string]any)
		if err := json.Unmarshal([]byte(input), &t); err != nil {
			return "", err
		}
		data = t
	}

	file := data.(map[string]any)

	filePath := fmt.Sprintf("%v", file["path"])

	// 如果是相对路径,转换为绝对路径
	if !filepath.IsAbs(filePath) {
		workDir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("获取工作目录失败: %v", err)
		}
		filePath = filepath.Join(workDir, filePath)
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("文件不存在: %s", filePath)
	}

	// 使用新的PDF处理器
	pdfProcessor := NewPDFProcessor()
	chunkedDocuments, err := pdfProcessor.LoadAndSplitPDF(ctx, filePath, 500, 50)
	if err != nil {
		return "", fmt.Errorf("PDF处理失败: %v", err)
	}

	fmt.Printf("成功处理PDF文件: %s，生成 %d 个文档块\n", filePath, len(chunkedDocuments))

	if k.store == nil {
		k.store, err = getKnowledgeStore(ctx, k.svc)
		if err != nil {
			return "", err
		}
	}

	_, err = k.store.AddDocuments(ctx, chunkedDocuments)
	if err != nil {
		return "", err
	}

	return Success, nil
}

// getKnowledgeStore 获取知识库的向量存储
func getKnowledgeStore(ctx context.Context, svc *svc.ServiceContext) (*redisvector.Store, error) {
	embedder, err := embeddings.NewEmbedder(svc.LLMs)
	if err != nil {
		return nil, err
	}

	redisUrl := "redis://"
	if svc.Config.Redis.Password != "" {
		redisUrl += ":" + svc.Config.Redis.Password + "@"
	}
	redisUrl += svc.Config.Redis.Addr

	return redisvector.New(ctx, redisvector.WithEmbedder(embedder), redisvector.WithConnectionURL(redisUrl), redisvector.WithIndexName("knowledge", true))
}
