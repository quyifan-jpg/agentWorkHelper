package token

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
)

type contextKey string

const (
	userIDKey contextKey = "user_id"
)

// GetUserID 从上下文中获取用户ID
func GetUserID(ctx context.Context) (uint, error) {
	uid, ok := ctx.Value(userIDKey).(uint)
	if !ok {
		return 0, errors.New("user id not found in context")
	}
	return uid, nil
}

// SetUserID 设置用户ID到上下文
func SetUserID(ctx context.Context, userID uint) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserIDFromGin 从 Gin 上下文中获取用户ID
func GetUserIDFromGin(ctx *gin.Context) (uint, error) {
	return GetUserID(ctx.Request.Context())
}

