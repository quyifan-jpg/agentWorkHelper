package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构（适配前端格式）
type Response struct {
	Code int         `json:"code"`    // 200 表示成功，其他表示失败
	Msg  string      `json:"msg"`     // 响应消息
	Data interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "success",
		Data: data,
	})
}

// SuccessWithMessage 带消息的成功响应
func SuccessWithMessage(ctx *gin.Context, message string, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  message,
		Data: data,
	})
}

// Fail 失败响应
func Fail(ctx *gin.Context, code int, message string) {
	ctx.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  message,
	})
}

// FailWithErr 使用错误对象的失败响应
func FailWithErr(ctx *gin.Context, err error) {
	Fail(ctx, http.StatusInternalServerError, err.Error())
}

// BadRequest 400 错误响应
func BadRequest(ctx *gin.Context, message string) {
	Fail(ctx, 400, message)
}

// Unauthorized 401 错误响应
func Unauthorized(ctx *gin.Context, message string) {
	Fail(ctx, 401, message)
}

// NotFound 404 错误响应
func NotFound(ctx *gin.Context, message string) {
	Fail(ctx, http.StatusNotFound, message)
}

// InternalError 500 错误响应
func InternalError(ctx *gin.Context, message string) {
	Fail(ctx, http.StatusInternalServerError, message)
}

