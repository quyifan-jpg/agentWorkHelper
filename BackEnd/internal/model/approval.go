package model

import (
	"time"

	"gorm.io/gorm"
)

type Approval struct {
	gorm.Model
	// 基础字段
	No       string         `gorm:"type:varchar(32);uniqueIndex"` // 审批单号
	Title    string         `gorm:"type:varchar(128)"`
	Reason   string         `gorm:"type:varchar(255)"`
	Type     ApprovalType   // 1:请假, 2:外出, 3:补卡
	Status   ApprovalStatus // 0:待审批, 1:通过, 2:驳回, 3:撤销, 4:草稿
	Abstract string         `gorm:"type:varchar(255)"` // 摘要

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
	FinishAt    *time.Time
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

	Status ApprovalStatus `gorm:"default:0"`         // 0:待审批, 1:已通过, 2:已驳回
	Reason string         `gorm:"type:varchar(255)"` // 审批意见
}

// 下面是 JSON 结构体定义，不需要 gorm.Model
type MakeCard struct {
	Date      int64         `json:"date,omitempty"`          //补卡时间
	Reason    string        `json:"reason,omitempty"`        //补卡理由
	Day       int64         `json:"day,omitempty"`           //补卡日期(20221011)
	CheckType WorkCheckType `json:"workCheckType,omitempty"` //补卡类型
}

type Leave struct {
	Type      LeaveType      `json:"type,omitempty"`      //请假类型
	StartTime int64          `json:"startTime,omitempty"` //开始时间
	EndTime   int64          `json:"endTime,omitempty"`   //结束时间
	Duration  float32        `json:"duration,omitempty"`  //时长
	Reason    string         `json:"reason,omitempty"`    //请假原由
	TimeType  TimeFormatType `json:"timeType,omitempty"`  //请假类型  1=小时 2=天
}

type GoOut struct {
	StartTime int64   `json:"startTime,omitempty"` //开始时间
	EndTime   int64   `json:"endTime,omitempty"`   //结束时间
	Duration  float32 `json:"duration,omitempty"`  //时长
	Reason    string  `json:"reason,omitempty"`    //请假原由
}

// Enums and Constants

type ApprovalType int

const (
	LeaveApproval    ApprovalType = 1 // 请假
	GoOutApproval    ApprovalType = 2 // 外出
	MakeCardApproval ApprovalType = 3 // 补卡
)

func (t ApprovalType) ToString() string {
	switch t {
	case LeaveApproval:
		return "请假"
	case GoOutApproval:
		return "外出"
	case MakeCardApproval:
		return "补卡"
	}
	return "未知"
}

type ApprovalStatus int

const (
	Processed ApprovalStatus = 0 // 待处理
	Pass      ApprovalStatus = 1 // 通过
	Refuse    ApprovalStatus = 2 // 驳回
	Cancel    ApprovalStatus = 3 // 撤销
	Draft     ApprovalStatus = 4 // 草稿
)

type LeaveType int

const (
	Matter        LeaveType = 1 //事假
	Rest          LeaveType = 2 //调休
	Fall          LeaveType = 3 //病假
	Annual        LeaveType = 4 //年假
	Maternity     LeaveType = 5 //产假
	Paternity     LeaveType = 6 //陪产假
	Marriage      LeaveType = 7 //婚假
	Funeral       LeaveType = 8 //丧假
	Breastfeeding LeaveType = 9 //哺乳假
)

func (t LeaveType) ToString() string {
	switch t {
	case Matter:
		return "事假"
	case Rest:
		return "调休"
	case Fall:
		return "病假"
	case Annual:
		return "年假"
	case Maternity:
		return "产假"
	case Paternity:
		return "陪产假"
	case Marriage:
		return "婚假"
	case Funeral:
		return "丧假"
	case Breastfeeding:
		return "哺乳假"
	}
	return ""
}

// WorkCheckType 打卡类型
// 1. 上班卡, 2. 下班卡
type WorkCheckType int

const (
	OnWorkCheck  WorkCheckType = 1 // 上班
	OffWorkCheck WorkCheckType = 2 // 下班
)

// 1. 小时， 2. 天，3. 半天，4. 上半天， 5. 下半天
type TimeFormatType int

const (
	HourTimeFormatType TimeFormatType = iota + 1
	DayTimeFormatType
)
