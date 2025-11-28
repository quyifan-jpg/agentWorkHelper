/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package token 提供 JWT Token 生成和用户信息提取功能
package token

import (
	"context"
	"github.com/golang-jwt/jwt"
)

// Identify JWT Token 中用户标识的键名
const Identify = "aiworkhelper"

// GetJwtToken 生成 JWT Token
// secretKey: 签名密钥
// iat: 签发时间戳
// seconds: 有效期（秒）
// uid: 用户ID
func GetJwtToken(secretKey string, iat, seconds int64, uid string) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds // 过期时间
	claims["iat"] = iat           // 签发时间
	claims[Identify] = uid        // 用户标识

	// 创建 Token 并签名
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}

// GetUId 从上下文中获取用户ID
func GetUId(ctx context.Context) string {
	var uid string
	if jsonUid, ok := ctx.Value(Identify).(string); ok {
		uid = jsonUid
	}
	return uid
}
