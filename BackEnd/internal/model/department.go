package model

import (
"time"

"gorm.io/gorm"
)

// Department 部门模型
type Department struct {
	ID        uint           `gorm:"primaryKey"`
	Name      string         `gorm:"type:varchar(100);not null;comment:部门名称"`
	LeaderID   uint           `gorm:"comment:部门负责人ID"` // 关联 User.ID
	ParentID   uint           `gorm:"default:0;comment:父部门ID"`
	ParentPath string         `gorm:"type:varchar(255);comment:父部门路径"`
	Level      int            `gorm:"default:1;comment:部门层级"`
	Leader     string         `gorm:"type:varchar(100);comment:部门负责人姓名"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relations
	Users []User `gorm:"many2many:department_users;"`
}

// DepartmentUser 部门-用户关联表 (多对多)
type DepartmentUser struct {
	DepartmentID uint `gorm:"primaryKey"`
	UserID       uint `gorm:"primaryKey"`
	CreatedAt    time.Time
}
