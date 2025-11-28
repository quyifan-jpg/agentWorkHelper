/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package middleware 提供 HTTP 请求中间件
package middleware

import (
	"fmt"
	"gitee.com/dn-jinmin/tlog"
	"github.com/gin-gonic/gin"
	"time"
)

// Log 日志中间件结构体，用于记录 HTTP 请求的链路追踪日志
type Log struct{}

// NewLog 创建日志中间件实例
func NewLog() *Log {
	return &Log{}
}

// Handler 日志中间件处理函数，为每个 HTTP 请求生成链路追踪日志
func (w *Log) Handler(ctx *gin.Context) {
	startTime := time.Now()                                                   // 记录请求开始时间
	url := fmt.Sprintf("%s:%s", ctx.Request.URL.Path, ctx.Request.Method)    // 构造请求标识：路径:方法

	ctx.Request = ctx.Request.WithContext(tlog.TraceStart(ctx.Request.Context())) // 启动链路追踪，生成 trace ID
	defer func() {
		tlog.InfoCtx(ctx.Request.Context(), url, "time", tlog.RTField(startTime, time.Now())) // 记录请求完成日志和响应时间
	}()

	ctx.Next() // 继续执行后续中间件和处理函数
}

