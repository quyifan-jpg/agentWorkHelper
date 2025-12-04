package model

import (
	"gorm.io/gorm"
)

// GroupMember 群聊成员数据模型
type GroupMember struct {
	gorm.Model
	GroupId string `gorm:"type:varchar(64);index;comment:群ID（conversationId）"` // 群ID
	UserId  uint   `gorm:"index;comment:用户ID"`                                // 用户ID
}

// TableName 指定表名
func (GroupMember) TableName() string {
	return "group_members"
}

