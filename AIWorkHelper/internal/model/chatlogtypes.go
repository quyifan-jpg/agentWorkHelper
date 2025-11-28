/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package model 定义聊天相关的数据模型和类型
package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatType 聊天类型枚举
type ChatType int

const (
	GroupChatType  ChatType = iota + 1 // 群聊类型，值为1
	SingleChatType                     // 私聊类型，值为2
)

// ChatLog 聊天记录数据模型，对应MongoDB中的chat_log集合
type ChatLog struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"` // MongoDB文档ID

	ConversationId string   `bson:"conversationId"` // 会话ID，群聊为"all"，私聊为生成的唯一标识
	SendId         string   `bson:"sendId"`         // 发送者用户ID
	RecvId         string   `bson:"recvId"`         // 接收者用户ID，群聊时为空
	ChatType       ChatType `bson:"chatType"`       // 聊天类型：1=群聊，2=私聊
	MsgContent     string   `bson:"msgContent"`     // 消息内容
	SendTime       int64    `bson:"sendTime"`       // 发送时间戳

	UpdateAt int64 `bson:"updateAt,omitempty" json:"updateAt,omitempty"` // 更新时间戳
	CreateAt int64 `bson:"createAt,omitempty" json:"createAt,omitempty"` // 创建时间戳
}
