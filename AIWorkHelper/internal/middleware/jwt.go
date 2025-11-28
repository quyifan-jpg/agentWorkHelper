/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package middleware 提供 HTTP 中间件功能
package middleware

import (
	"AIWorkHelper/pkg/httpx"
	"AIWorkHelper/pkg/token"
	"github.com/gin-gonic/gin"
)

// Jwt JWT 认证中间件
type Jwt struct {
	tokenParser *token.Parse // Token 解析器
}

// NewJwt 创建 JWT 中间件实例
func NewJwt(secret string) *Jwt {
	return &Jwt{
		tokenParser: token.NewTokenParse(secret),
	}
}

// Handler JWT 认证处理器
func (m *Jwt) Handler(ctx *gin.Context) {
	// 解析 Token 并注入用户信息到请求上下文
	r, err := m.tokenParser.ParseWithContext(ctx.Request)
	if err != nil {
		// Token 解析失败，返回错误并终止请求
		httpx.FailWithErr(ctx, err)
		ctx.Abort()
		return
	}

	// 更新请求上下文，继续处理
	ctx.Request = r
	ctx.Next()
}
