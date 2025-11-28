/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package toolx

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gen2brain/go-fitz"
	"github.com/tmc/langchaingo/schema"
)

// PDFProcessor 提供改进的PDF文本提取功能
type PDFProcessor struct{}

// NewPDFProcessor 创建新的PDF处理器
func NewPDFProcessor() *PDFProcessor {
	return &PDFProcessor{}
}

// ExtractTextFromPDF 从PDF文件中提取文本，使用go-fitz库（基于MuPDF）解决中文乱码问题
func (p *PDFProcessor) ExtractTextFromPDF(filePath string) (string, error) {
	// 验证文件存在
	if _, err := os.Stat(filePath); err != nil {
		return "", fmt.Errorf("文件不存在: %v", err)
	}

	// 打开PDF文档
	doc, err := fitz.New(filePath)
	if err != nil {
		return "", fmt.Errorf("无法打开PDF文件: %v", err)
	}
	defer doc.Close()

	// 提取所有页面的文本
	var buf bytes.Buffer
	totalPages := doc.NumPage()

	for n := 0; n < totalPages; n++ {
		text, err := doc.Text(n)
		if err != nil {
			// 记录警告但继续处理其他页面
			fmt.Printf("警告: 无法提取第 %d 页: %v\n", n+1, err)
			continue
		}
		buf.WriteString(text)
		buf.WriteString("\n")
	}

	pdfText := buf.String()
	if len(strings.TrimSpace(pdfText)) == 0 {
		return "", fmt.Errorf("PDF文件中没有提取到有效文本内容")
	}

	// 清理文本
	cleanedText := p.cleanText(pdfText)

	fmt.Printf("PDF文本提取成功，总页数: %d，总长度: %d 字符，预览: %.200s\n",
		totalPages, len(cleanedText), cleanedText)

	return cleanedText, nil
}

// cleanText 清理提取的文本，去除不必要的空白和格式化
func (p *PDFProcessor) cleanText(text string) string {
	// 去除多余的空白字符
	text = strings.TrimSpace(text)

	// 将多个连续的空行合并为单个空行
	lines := strings.Split(text, "\n")
	var cleanedLines []string
	prevEmpty := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if !prevEmpty {
				cleanedLines = append(cleanedLines, "")
				prevEmpty = true
			}
		} else {
			cleanedLines = append(cleanedLines, line)
			prevEmpty = false
		}
	}

	return strings.Join(cleanedLines, "\n")
}

// LoadAndSplitPDF 加载PDF并分割成文档块，与LangChain兼容
func (p *PDFProcessor) LoadAndSplitPDF(_ context.Context, filePath string, chunkSize int, chunkOverlap int) ([]schema.Document, error) {
	// 提取PDF文本
	text, err := p.ExtractTextFromPDF(filePath)
	if err != nil {
		return nil, err
	}

	// 将文本分割成块
	chunks := p.splitText(text, chunkSize, chunkOverlap)

	// 转换为LangChain文档格式
	documents := make([]schema.Document, len(chunks))
	for i, chunk := range chunks {
		documents[i] = schema.Document{
			PageContent: chunk,
			Metadata: map[string]any{
				"source": filePath,
				"page":   i + 1,
			},
		}
	}

	fmt.Printf("PDF分割完成，共生成 %d 个文档块\n", len(documents))
	return documents, nil
}

// splitText 将文本分割成指定大小的块
func (p *PDFProcessor) splitText(text string, chunkSize int, chunkOverlap int) []string {
	if chunkSize <= 0 {
		chunkSize = 1000 // 默认块大小
	}
	if chunkOverlap < 0 {
		chunkOverlap = 0
	}

	var chunks []string
	textRunes := []rune(text) // 使用rune来正确处理中文字符
	textLen := len(textRunes)

	start := 0
	for start < textLen {
		end := start + chunkSize
		if end > textLen {
			end = textLen
		}

		chunk := string(textRunes[start:end])
		chunk = strings.TrimSpace(chunk)

		if len(chunk) > 0 {
			chunks = append(chunks, chunk)
		}

		// 计算下一个块的起始位置
		if end >= textLen {
			break
		}
		start = end - chunkOverlap
		if start < 0 {
			start = 0
		}
	}

	return chunks
}
