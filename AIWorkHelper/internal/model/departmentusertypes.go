/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DepartmentUser 部门用户关联数据模型
type DepartmentUser struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"` // 关联ID

	DepId  string `bson:"depId,omitempty"`  // 部门ID
	UserId string `bson:"userId,omitempty"` // 用户ID

	UpdateAt int64 `bson:"updateAt,omitempty" json:"updateAt,omitempty"` // 更新时间
	CreateAt int64 `bson:"createAt,omitempty" json:"createAt,omitempty"` // 创建时间
}
