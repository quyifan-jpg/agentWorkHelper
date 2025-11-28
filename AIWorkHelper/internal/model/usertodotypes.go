/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package model

import (
	"AIWorkHelper/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TodoStatus 待办事项状态枚举
type TodoStatus int

const (
	TodoPending    TodoStatus = iota + 1 // 待处理/未开始
	TodoInProgress                       // 进行中
	TodoFinish                           // 已完成
	TodoCancel                           // 已取消
	TodoTimeout                          // 已超时
)

// UserTodo 用户待办关联数据模型
type UserTodo struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"` // 关联ID

	UserId     string     `bson:"userId,omitempty"`                             // 用户ID
	TodoId     string     `bson:"todoId,omitempty"`                             // 待办事项ID
	TodoStatus TodoStatus `bson:"todoStatus,omitempty"`                         // 待办状态
	UpdateAt   int64      `bson:"updateAt,omitempty" json:"updateAt,omitempty"` // 更新时间
	CreateAt   int64      `bson:"createAt,omitempty" json:"createAt,omitempty"` // 创建时间
}

// ToDomain 将数据模型转换为领域对象，需要传入用户名
func (m *UserTodo) ToDomain(username string) *domain.UserTodo {
	return &domain.UserTodo{
		ID:         m.ID.Hex(),
		UserId:     m.UserId,
		UserName:   username,
		TodoId:     m.TodoId,
		TodoStatus: int(m.TodoStatus),
	}
}
