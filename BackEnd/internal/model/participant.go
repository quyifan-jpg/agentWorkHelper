package model

// Participant 会话参与者表
type Participant struct {
	Id             uint   `gorm:"primaryKey;autoIncrement;comment:ID"`
	ConversationId string `gorm:"type:varchar(64);not null;index:idx_conversation_user;comment:会话ID"`
	UserId         string `gorm:"type:varchar(64);not null;index:idx_conversation_user;comment:用户ID"`
	Role           int    `gorm:"type:tinyint;default:0;comment:角色:0=普通成员,1=管理员,2=群主"`
	JoinTime       int64  `gorm:"autoCreateTime;comment:加入时间"`
}

func (Participant) TableName() string {
	return "participants"
}
