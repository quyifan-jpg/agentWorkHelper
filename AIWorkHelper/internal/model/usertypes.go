/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	// TODO: Fill your own fields

	Name     string `bson:"name" json:"name"`
	Password string `bson:"Password" json:"Password"`
	Status   int    `bson:"status" json:"status"`
	IsAdmin  bool   `bson:"isAdmin" json:"isAdmin"` // 是否为管理员
	UpdateAt int64  `bson:"updateAt" json:"updateAt"`
	CreateAt int64  `bson:"createAt" json:"createAt"`
}
