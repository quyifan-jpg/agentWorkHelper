package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// cause 接口定义了获取根本错误原因的方法
type cause interface {
	Cause() error
}

// ErrorHandler 全局错误处理函数
func ErrorHandler(ctx *gin.Context, err error) (int, error) {
	var e error
	if ce, ok := err.(cause); ok {
		e = ce.Cause()
	} else {
		e = err
	}

	// 根据错误类型返回不同的状态码
	if e.Error() == "user not found" || e.Error() == "invalid password" {
		return http.StatusUnauthorized, e
	}
	if e.Error() == "username already exists" {
		return http.StatusBadRequest, e
	}

	return http.StatusInternalServerError, e
}

