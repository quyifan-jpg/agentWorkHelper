/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package handler 提供 HTTP 请求处理相关功能
package handler

import (
	"gitee.com/dn-jinmin/tlog"
	"github.com/gin-gonic/gin"
)

// cause 接口定义了获取根本错误原因的方法，用于错误链追踪
type cause interface {
	Cause() error
}

// ErrorHandler 全局错误处理函数，实现 httpx.SetErrorHandler 所需的错误处理接口
// 该函数会被 httpx 包在调用 FailWithErr 时自动调用，用于统一处理应用中的错误
// 参数: ctx - Gin 上下文, err - 需要处理的错误
// 返回: HTTP状态码和处理后的错误对象
func ErrorHandler(ctx *gin.Context, err error) (int, error) {
	var e error
	if ce, ok := err.(cause); ok { // 如果错误实现了 cause 接口，获取根本原因
		e = ce.Cause()
		tlog.ErrorCtx(ctx.Request.Context(), "err", err.Error())
	} else { // 否则直接使用原始错误
		e = err
		tlog.ErrorCtx(ctx.Request.Context(), "err", err.Error())
	}
	return 500, e // 统一返回 500 状态码和错误对象
}
