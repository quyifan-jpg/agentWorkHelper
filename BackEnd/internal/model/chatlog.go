package model

import (
	"gorm.io/gorm"
)

// ChatType 聊天类型枚举
type ChatType int

const (
	GroupChatType  ChatType = 1 // 群聊类型
	SingleChatType ChatType = 2 // 私聊类型
)

// ChatLog 聊天记录数据模型
type ChatLog struct {
	gorm.Model
	ConversationId string   `gorm:"type:varchar(64);index;comment:会话ID"` // 会话ID，群聊为"all"或群ID，私聊为生成的唯一标识
	SendId         uint     `gorm:"index;comment:发送者用户ID"`            // 发送者ID
	RecvId         uint     `gorm:"index;default:0;comment:接收者用户ID"`  // 接收者ID，群聊时为0
	ChatType       ChatType `gorm:"default:2;comment:聊天类型：1=群聊，2=私聊"` // 聊天类型
	MsgContent     string   `gorm:"type:text;comment:消息内容"`          // 消息内容
	SendTime       int64    `gorm:"index;comment:发送时间戳"`              // 发送时间戳
}

// TableName 指定表名
func (ChatLog) TableName() string {
	return "chat_logs"
}

