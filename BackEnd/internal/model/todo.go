package model

import (
"time"

"gorm.io/gorm"
)

// Todo 待办事项
type Todo struct {
	ID          uint           `gorm:"primaryKey"`
	CreatorID   uint           `gorm:"index;comment:创建人ID"`
	Title       string         `gorm:"type:varchar(255);not null;comment:标题"`
	Desc        string         `gorm:"type:text;comment:描述"`
	DeadlineAt  time.Time      `gorm:"comment:截止时间"`
	Status      int            `gorm:"default:0;comment:状态"` // 业务状态
	TodoStatus  int            `gorm:"default:0;comment:待办状态"` // 完成状态
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relations
	Creator     User           `gorm:"foreignKey:CreatorID"`
	Executors   []User         `gorm:"many2many:user_todos;"`
	Records     []TodoRecord   `gorm:"foreignKey:TodoID"`
}

// TodoRecord 待办事项操作记录
type TodoRecord struct {
	ID        uint      `gorm:"primaryKey"`
	TodoID    uint      `gorm:"index;not null"`
	UserID    uint      `gorm:"index;not null"`
	Content   string    `gorm:"type:text"`
	Image     string    `gorm:"type:varchar(255)"`
	CreatedAt time.Time

	// Relations
	User      User      `gorm:"foreignKey:UserID"`
}

// UserTodo 用户-待办关联表 (多对多)
type UserTodo struct {
	TodoID     uint `gorm:"primaryKey"`
	UserID     uint `gorm:"primaryKey"`
	TodoStatus int  `gorm:"default:0"` // 个人的完成状态
	CreatedAt  time.Time
}
