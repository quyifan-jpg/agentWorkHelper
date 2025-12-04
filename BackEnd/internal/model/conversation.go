package model

// Conversation 会话表
type Conversation struct {
	Id              string `gorm:"primaryKey;type:varchar(64);comment:会话ID"`
	Type            int    `gorm:"type:tinyint;not null;comment:会话类型:1=群聊,2=私聊,3=AI"`
	Name            string `gorm:"type:varchar(100);comment:会话名称"`
	LastMessageId   uint   `gorm:"comment:最后一条消息ID"`
	LastMessageTime int64  `gorm:"comment:最后一条消息时间"`
	CreatorId       string `gorm:"type:varchar(64);comment:创建者ID"`
	CreateAt        int64  `gorm:"autoCreateTime;comment:创建时间"`
	UpdateAt        int64  `gorm:"autoUpdateTime;comment:更新时间"`
}

func (Conversation) TableName() string {
	return "conversations"
}
