/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package model

import (
	"AIWorkHelper/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Todo 待办事项数据模型
type Todo struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"` // 待办事项ID

	CreatorId  string               `bson:"creatorId"`  // 创建人ID
	Title      string               `bson:"title"`      // 待办标题
	DeadlineAt int64                `bson:"deadlineAt"` // 截止时间
	Desc       string               `bson:"desc"`       // 待办描述
	Records    []*TodoRecord        `bson:"records"`    // 操作记录列表
	Executes   []*UserTodo          `bson:"executes"`   // 执行人列表
	TodoStatus `bson:"todo_status"` // 待办状态

	UpdateAt int64 `bson:"updateAt,omitempty" json:"updateAt,omitempty"` // 更新时间
	CreateAt int64 `bson:"createAt,omitempty" json:"createAt,omitempty"` // 创建时间
}

// ToDomainTodo 将数据模型转换为领域对象
func (m *Todo) ToDomainTodo() *domain.Todo {
	return &domain.Todo{
		ID:         m.ID.Hex(),
		CreatorId:  m.CreatorId,
		Title:      m.Title,
		DeadlineAt: m.DeadlineAt,
		Desc:       m.Desc,
		ExecuteIds: nil,
		Status:     int(m.TodoStatus),
		TodoStatus: int(m.TodoStatus),
	}
}
