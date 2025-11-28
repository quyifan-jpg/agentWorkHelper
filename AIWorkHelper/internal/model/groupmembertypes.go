/**
 * @author: 公众号:IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package model 定义群聊成员相关的数据模型和类型
package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GroupMember 群聊成员数据模型，对应MongoDB中的group_member集合
// 用于管理群聊的成员关系，支持多个群聊的成员管理
type GroupMember struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"` // MongoDB文档ID

	GroupId string `bson:"groupId"` // 群ID (conversationId)
	UserId  string `bson:"userId"`  // 用户ID

	UpdateAt int64 `bson:"updateAt,omitempty" json:"updateAt,omitempty"` // 更新时间戳
	CreateAt int64 `bson:"createAt,omitempty" json:"createAt,omitempty"` // 创建时间戳
}
