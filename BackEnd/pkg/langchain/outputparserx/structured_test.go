package outputparserx

import "testing"

// Test_Structured_GetFormatInstructions 测试结构化输出解析器的格式指令生成功能
func Test_Structured_GetFormatInstructions(t *testing.T) {
	// 创建包含多种字段类型和嵌套结构的测试模式
	out := NewStructured([]ResponseSchema{
		{
			Name:        "title",
			Description: "this is title ",
		}, {
			Name:        "deadlineAt",
			Description: "todo deadline",
			Type:        "int64",
		}, {
			Name:        "executeIds",
			Description: "todo execute ids",
			Type:        "[]string",
		}, {
			Name:        "record",
			Description: "record todo handler",
			// 嵌套结构测试
			Schemas: []ResponseSchema{
				{
					Name:        "FinishAt",
					Description: "todo finish time",
					Type:        "int64",
				}, {
					Name:        "Content",
					Description: "record content",
				},
			},
		},
	})

	// 输出格式指令用于验证
	t.Log(out.GetFormatInstructions())
}
