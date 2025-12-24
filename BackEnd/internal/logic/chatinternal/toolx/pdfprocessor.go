package toolx

import (
	"context"
	"fmt"
	"os"

	"github.com/dslipak/pdf"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

// PDFProcessor 处理PDF文件的结构体
type PDFProcessor struct{}

// NewPDFProcessor 创建一个新的PDF处理器
func NewPDFProcessor() *PDFProcessor {
	return &PDFProcessor{}
}

// LoadAndSplitPDF 加载PDF文件并将其拆分为文档块
func (p *PDFProcessor) LoadAndSplitPDF(ctx context.Context, filePath string, chunkSize, chunkOverlap int) ([]schema.Document, error) {
	// 1. 读取PDF内容
	content, err := p.readPDF(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取PDF文件失败: %v", err)
	}

	// 2. 创建文本拆分器
	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(chunkSize),
		textsplitter.WithChunkOverlap(chunkOverlap),
	)

	// 3. 拆分文本
	chunks, err := splitter.SplitText(content)
	if err != nil {
		return nil, fmt.Errorf("拆分文本失败: %v", err)
	}

	// 4. 转换为Document对象
	documents := make([]schema.Document, 0, len(chunks))
	for _, chunk := range chunks {
		doc := schema.Document{
			PageContent: chunk,
			Metadata: map[string]any{
				"source": filePath,
			},
		}
		documents = append(documents, doc)
	}

	return documents, nil
}

// readPDF 读取PDF文件内容
func (p *PDFProcessor) readPDF(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	r, err := pdf.NewReader(f, 0)
	if err != nil {
		return "", err // 可能需要特定的错误处理
	}

	totalPage := r.NumPage()
	var content string

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		text, err := p.GetPlainText(nil)
		if err != nil {
			// 记录错误但继续处理其他页面
			fmt.Printf("警告: 读取第 %d 页失败: %v\n", pageIndex, err)
			continue
		}
		content += text + "\n"
	}

	return content, nil
}
