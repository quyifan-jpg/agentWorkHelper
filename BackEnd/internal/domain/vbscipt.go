/**
 */
package domain

import "time"

type AiChatType int

const (
	DefaultHandler = iota // 默认处理器类型
	TodoFind              // 待办查询类型
	TodoAdd               // 待办添加类型

	ApprovalFind // 审批查询类型

	ChatLog // 聊天日志类型
)

// ChatFile 聊天文件信息结构
type ChatFile struct {
	Path string    // 文件路径
	Name string    // 文件名称
	Time time.Time // 文件时间
}
