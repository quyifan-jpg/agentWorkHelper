package model

import (
	"time"

	"gorm.io/gorm"
)

type Approval struct {
	gorm.Model
	// 基础字段
	No       string `gorm:"type:varchar(32);uniqueIndex"` // 审批单号
	Title    string `gorm:"type:varchar(128)"`
	Reason   string `gorm:"type:varchar(255)"`
	Type     int    // 1:请假, 2:外出, 3:补卡
	Status   int    // 0:待审批, 1:通过, 2:驳回, 3:撤销, 4:草稿
	Abstract string `gorm:"type:varchar(255)"` // 摘要

	// 申请人关联
	UserID uint
	User   User `gorm:"foreignKey:UserID"`

	// 审批流程关联
	Approvers []Approver `gorm:"foreignKey:ApprovalID"`

	// 业务详情数据 (JSON存储)
	// GORM v2 支持 serializer:json，会自动将结构体序列化为 JSON 字符串存入数据库
	MakeCard *MakeCard `gorm:"serializer:json"`
	Leave    *Leave    `gorm:"serializer:json"`
	GoOut    *GoOut    `gorm:"serializer:json"`

	// 时间字段
	FinishAt    time.Time
	FinishDay   int64
	FinishMonth int64
	FinishYeas  int64
}

// Approver 审批流程节点表
type Approver struct {
	gorm.Model
	ApprovalID uint // 关联 Approval
	UserID     uint // 审批人 ID
	User       User `gorm:"foreignKey:UserID"` // 关联 User 表获取名字

	Status int    // 0:待审批, 1:已通过, 2:已驳回
	Reason string `gorm:"type:varchar(255)"` // 审批意见
}

// 下面是 JSON 结构体定义，不需要 gorm.Model
type MakeCard struct {
	Date      int64  `json:"date,omitempty"`          //补卡时间
	Reason    string `json:"reason,omitempty"`        //补卡理由
	Day       int64  `json:"day,omitempty"`           //补卡日期(20221011)
	CheckType int    `json:"workCheckType,omitempty"` //补卡类型
}

type Leave struct {
	Type      int     `json:"type,omitempty"`      //请假类型
	StartTime int64   `json:"startTime,omitempty"` //开始时间
	EndTime   int64   `json:"endTime,omitempty"`   //结束时间
	Duration  float32 `json:"duration,omitempty"`  //时长
	Reason    string  `json:"reason,omitempty"`    //请假原由
	TimeType  int     `json:"timeType,omitempty"`  //请假类型  1=小时 2=天
}

type GoOut struct {
	StartTime int64   `json:"startTime,omitempty"` //开始时间
	EndTime   int64   `json:"endTime,omitempty"`   //结束时间
	Duration  float32 `json:"duration,omitempty"`  //时长
	Reason    string  `json:"reason,omitempty"`    //请假原由
}
