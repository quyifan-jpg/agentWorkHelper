package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage 带消息的成功响应
func SuccessWithMessage(ctx *gin.Context, message string, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

// Fail 失败响应
func Fail(ctx *gin.Context, code int, message string) {
	ctx.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

// FailWithErr 使用错误对象的失败响应
func FailWithErr(ctx *gin.Context, err error) {
	Fail(ctx, http.StatusInternalServerError, err.Error())
}

// BadRequest 400 错误响应
func BadRequest(ctx *gin.Context, message string) {
	Fail(ctx, http.StatusBadRequest, message)
}

// Unauthorized 401 错误响应
func Unauthorized(ctx *gin.Context, message string) {
	Fail(ctx, http.StatusUnauthorized, message)
}

// NotFound 404 错误响应
func NotFound(ctx *gin.Context, message string) {
	Fail(ctx, http.StatusNotFound, message)
}

// InternalError 500 错误响应
func InternalError(ctx *gin.Context, message string) {
	Fail(ctx, http.StatusInternalServerError, message)
}

