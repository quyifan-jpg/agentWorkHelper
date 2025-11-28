/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package model

import "AIWorkHelper/internal/domain"

// TodoRecord 待办事项操作记录数据模型
type TodoRecord struct {
	UserId   string `json:"userId,omitempty"`   // 操作用户ID
	UserName string `json:"userName,omitempty"` // 操作用户名
	Content  string `json:"content,omitempty"`  // 操作内容
	Image    string `json:"image,omitempty"`    // 操作相关图片
	CreateAt int64  `json:"createAt,omitempty"` // 操作时间
}

// ToDomainTodoRecord 将数据模型转换为领域对象
func (m *TodoRecord) ToDomainTodoRecord() *domain.TodoRecord {
	return &domain.TodoRecord{
		UserId:   m.UserId,
		UserName: m.UserName,
		Content:  m.Content,
		Image:    m.Image,
		CreateAt: m.CreateAt,
	}
}

// ToDomainTodoRecords 将待办事项的操作记录列表转换为领域对象列表
func (m *Todo) ToDomainTodoRecords() []*domain.TodoRecord {
	res := make([]*domain.TodoRecord, 0, len(m.Records))
	for _, record := range m.Records {
		res = append(res, record.ToDomainTodoRecord())
	}
	return res
}
