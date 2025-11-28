package middleware

import (
	"BackEnd/pkg/jwt"
	"BackEnd/pkg/token"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Jwt JWT 认证中间件
type Jwt struct {
	secret string
}

// NewJwt 创建 JWT 中间件实例
func NewJwt(secret string) *Jwt {
	return &Jwt{
		secret: secret,
	}
}

// Handler JWT 认证处理器
func (m *Jwt) Handler(ctx *gin.Context) {
	// 从 Header 中获取 Token
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Authorization header is required",
		})
		ctx.Abort()
		return
	}

	// 检查 Bearer 前缀
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Invalid authorization header format",
		})
		ctx.Abort()
		return
	}

	tokenString := parts[1]

	// 解析 Token
	userID, err := jwt.ParseToken(tokenString, m.secret)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Invalid or expired token",
		})
		ctx.Abort()
		return
	}

	// 将用户ID注入到请求上下文中
	ctx.Request = ctx.Request.WithContext(token.SetUserID(ctx.Request.Context(), userID))
	ctx.Next()
}

