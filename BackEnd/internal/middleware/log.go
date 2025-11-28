package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Log 日志中间件
type Log struct{}

// NewLog 创建日志中间件实例
func NewLog() *Log {
	return &Log{}
}

// Handler 日志处理器
func (m *Log) Handler(ctx *gin.Context) {
	start := time.Now()
	path := ctx.Request.URL.Path
	raw := ctx.Request.URL.RawQuery

	// 处理请求
	ctx.Next()

	// 计算处理时间
	latency := time.Since(start)

	// 记录日志
	if raw != "" {
		path = path + "?" + raw
	}

	// 使用 fmt 输出日志
	logMsg := fmt.Sprintf("[%s] %s | %d | %s | %s | %s\n",
		time.Now().Format("2006/01/02 - 15:04:05"),
		ctx.ClientIP(),
		ctx.Writer.Status(),
		latency,
		ctx.Request.Method,
		path,
	)
	fmt.Fprint(gin.DefaultWriter, logMsg)
}

