/**
*/
// Package domain 定义WebSocket聊天相关的领域模型
package domain

// Message WebSocket聊天消息的领域模型
type Message struct {
	ConversationId string `json:"conversationId"` // 会话ID，群聊固定为"all"，私聊为双方用户ID生成的唯一标识

	RecvId string `json:"recvId"` // 接收者用户ID，群聊时为空
	SendId string `json:"sendId"` // 发送者用户ID，由服务器从JWT Token中提取

	ChatType    int    `json:"chatType"`    // 聊天类型：1=群聊，2=私聊
	Content     string `json:"content"`     // 消息内容文本
	ContentType int    `json:"contentType"` // 内容类型：1=文字，2=图片，3=表情包等
}
