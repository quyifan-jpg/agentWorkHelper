/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package token 提供 JWT Token 解析和验证功能
package token

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
)

// JWT 标准声明字段常量
const (
	jwtAudience   = "aud"           // 受众
	jwtExpire     = "exp"           // 过期时间
	jwtId         = "jti"           // JWT ID
	jwtIssueAt    = "iat"           // 签发时间
	jwtIssuer     = "iss"           // 签发者
	jwtNotBefore  = "nbf"           // 生效时间
	jwtSubject    = "sub"           // 主题
	Authorization = "Authorization" // HTTP 头字段名
)

// Token 解析相关错误定义
var (
	ErrTokenNotFound = errors.New("token不存在")
	ErrTokenInvalid  = errors.New("token is invalid")
	ErrClaimsInvalid = errors.New("invalid token claims")
)

// Parse JWT Token 解析器
type Parse struct {
	AccessSecret string // JWT 签名密钥
}

// NewTokenParse 创建 Token 解析器实例
func NewTokenParse(secret string) *Parse {
	return &Parse{AccessSecret: secret}
}

// Parse 从 HTTP 请求中解析 JWT Token
func (p *Parse) Parse(r *http.Request) (jwt.MapClaims, string, error) {
	// 从请求头提取 Token
	tokenStr := p.extractTokenFromHeader(r)
	if len(tokenStr) == 0 {
		return nil, tokenStr, ErrTokenNotFound
	}
	return p.ParseToken(tokenStr)
}

// ParseToken 解析 JWT Token 字符串
func (p *Parse) ParseToken(tokenStr string) (jwt.MapClaims, string, error) {
	// 解析 Token
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(p.AccessSecret), nil
	})
	if err != nil {
		return nil, tokenStr, err
	}

	// 验证 Token 有效性
	if !token.Valid {
		return nil, tokenStr, ErrTokenInvalid
	}

	// 提取声明信息
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, tokenStr, ErrClaimsInvalid
	}

	// 验证声明有效性
	if err = claims.Valid(); err != nil {
		return nil, tokenStr, err
	}
	return claims, tokenStr, nil
}

// ParseWithContext 解析 Token 并将用户信息注入到请求上下文
func (p *Parse) ParseWithContext(r *http.Request) (*http.Request, error) {
	// 解析 Token
	claims, tokenStr, err := p.Parse(r)
	if err != nil {
		return r, err
	}

	// 将自定义声明注入上下文
	ctx := r.Context()
	for k, v := range claims {
		switch k {
		case jwtAudience, jwtExpire, jwtId, jwtIssueAt, jwtIssuer, jwtNotBefore, jwtSubject:
			// 忽略 JWT 标准声明
		default:
			ctx = context.WithValue(ctx, k, v)
		}
	}

	// 将原始 Token 字符串也保存到上下文
	ctx = context.WithValue(ctx, Authorization, tokenStr)

	return r.WithContext(ctx), nil
}

// extractTokenFromHeader 从 HTTP 请求头中提取 Token
func (p *Parse) extractTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Token 格式应为 "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return authHeader
	}

	return parts[1]
}

// GetTokenStr 从上下文中获取原始 Token 字符串
func GetTokenStr(ctx context.Context) string {
	tokenStr, ok := ctx.Value(Authorization).(string)
	if !ok {
		return ""
	}
	return tokenStr
}
