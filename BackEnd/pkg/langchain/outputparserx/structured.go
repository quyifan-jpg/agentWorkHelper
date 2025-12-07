package outputparserx

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// ParseError 输出解析器返回的错误类型
type ParseError struct {
	Text   string // 解析失败的文本内容
	Reason string // 失败原因描述
}

// Error 实现error接口，返回格式化的错误信息
func (e ParseError) Error() string {
	return fmt.Sprintf("parse text %s. %s", e.Text, e.Reason)
}

const (
	// _structuredFormatInstructionTemplate is a template for the format
	// instructions of the structured output parser.
	_structuredFormatInstructionTemplate = "The output should be a markdown code snippet formatted in the following schema: \n```json\n%s\n```" // nolint

	// _structuredLineTemplate is a single line of the json schema in the
	// format instruction of the structured output parser. The fist verb is
	// the name, the second verb is the type and the third is a description of
	// what the field should contain.
	_structuredLineTemplate = "\"%s\": %s // %s\n"
)

// ResponseSchema 结构化输出解析器的响应模式定义
// 用于描述LLM应该如何格式化其响应
type ResponseSchema struct {
	Name        string           // 字段名称，作为解析输出映射的键
	Description string           // 字段描述，说明该值应包含的内容
	Type        string           // 字段类型，如string、int64、[]string等
	Require     bool             // 是否为必填字段
	Schemas     []ResponseSchema // 嵌套的子模式，用于复杂对象结构
}

// Structured 结构化输出解析器，将LLM输出解析为键值对
// 通过响应模式列表定义LLM输出应包含的字段名称和描述
type Structured struct {
	ResponseSchemas []ResponseSchema // 响应模式列表，定义输出结构
}

// NewStructured 从响应模式列表创建新的结构化输出解析器
func NewStructured(schema []ResponseSchema) Structured {
	return Structured{
		ResponseSchemas: schema,
	}
}

// 静态断言确保Structured实现了OutputParser接口
var _ schema.OutputParser[any] = Structured{}

// parse 将LLM输出解析为映射，如果输出不包含响应模式中指定的必填字段则返回错误
func (p Structured) parse(text string) (map[string]any, error) {
	var jsonString string

	// 尝试提取markdown代码块中的JSON
	withoutJSONStart := strings.Split(text, "```json")
	if len(withoutJSONStart) > 1 {
		// 找到了```json标记，提取JSON内容
		withoutJSONEnd := strings.Split(withoutJSONStart[1], "```")
		if len(withoutJSONEnd) < 1 {
			return nil, ParseError{Text: text, Reason: "no ``` at end of output"}
		}
		jsonString = strings.TrimSpace(withoutJSONEnd[0])
	} else {
		// 没有找到markdown代码块，直接使用原始文本作为JSON字符串
		jsonString = strings.TrimSpace(text)
	}

	// 解析JSON为映射
	var parsed map[string]any
	err := json.Unmarshal([]byte(jsonString), &parsed)
	if err != nil {
		return nil, ParseError{Text: text, Reason: fmt.Sprintf("invalid JSON: %v", err)}
	}

	// 验证解析的映射包含响应模式中指定的所有必填字段
	missingKeys := make([]string, 0)
	for _, rs := range p.ResponseSchemas {
		if _, ok := parsed[rs.Name]; !ok && rs.Require {
			missingKeys = append(missingKeys, rs.Name)
		}
	}

	// 如果有缺失的必填字段，返回错误
	if len(missingKeys) > 0 {
		return nil, ParseError{
			Text:   text,
			Reason: fmt.Sprintf("output is missing the following fields %v", missingKeys),
		}
	}

	return parsed, nil
}

// Parse 解析文本并返回结果
func (p Structured) Parse(text string) (any, error) {
	return p.parse(text)
}

// ParseWithPrompt 与Parse功能相同，忽略提示值参数
func (p Structured) ParseWithPrompt(text string, _ llms.PromptValue) (any, error) {
	return p.parse(text)
}

// GetFormatInstructions 返回说明LLM应如何格式化响应的字符串
func (p Structured) GetFormatInstructions() string {
	// 注释掉的旧实现保留作为参考
	//jsonLines := ""
	//for _, rs := range p.ResponseSchemas {
	//	if len(rs.Type) == 0 {
	//		rs.Type = "string"
	//	}
	//
	//	jsonLines += "\t" + fmt.Sprintf(
	//		_structuredLineTemplate,
	//		rs.Name,
	//		rs.Type,
	//		//"string", /* type of the filed*/
	//		rs.Description,
	//	)
	//}

	// 使用新的JSON格式化方法生成格式指令
	return fmt.Sprintf(_structuredFormatInstructionTemplate, p.jsonMarshal(p.ResponseSchemas, 0))
}

// jsonMarshal 递归生成JSON模式字符串，支持嵌套结构
func (p Structured) jsonMarshal(schemas []ResponseSchema, level int) string {
	level++

	// 计算缩进空白
	endBlank := ""
	fieldBlank := "\t"

	for i := 0; i < level; i++ {
		fieldBlank += "\t"
		endBlank += "\t"
	}

	// 构建JSON对象
	jsonLines := "{"
	for _, rs := range schemas {
		// 处理嵌套模式
		if len(rs.Schemas) > 0 {
			rs.Type = p.jsonMarshal(rs.Schemas, level)
		}

		// 设置字段缩进
		blank := fieldBlank
		if len(jsonLines) == 1 {
			blank = "\t"
		}

		// 设置默认类型
		if len(rs.Type) == 0 {
			rs.Type = "string"
		}
		// 添加字段定义
		jsonLines += blank + fmt.Sprintf(
			_structuredLineTemplate,
			rs.Name,
			rs.Type,
			rs.Description,
		)
	}

	return jsonLines + endBlank + "}"
}

// Type 返回输出解析器的类型标识
func (p Structured) Type() string {
	return "structured_parser"
}
