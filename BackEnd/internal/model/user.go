package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey"`
	Name      string         `gorm:"type:varchar(100);uniqueIndex;not null"`
	Password  string         `gorm:"type:varchar(255);not null"`
	Status    int            `gorm:"default:0"` // 0: normal, 1: disabled
	IsAdmin   bool           `gorm:"default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
