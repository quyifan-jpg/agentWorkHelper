package httpx

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// BindAndValidate 绑定并验证请求参数
func BindAndValidate(ctx *gin.Context, obj interface{}) error {
	if err := ctx.ShouldBindJSON(obj); err != nil {
		return err
	}
	return validate.Struct(obj)
}

// BindQuery 绑定查询参数
func BindQuery(ctx *gin.Context, obj interface{}) error {
	return ctx.ShouldBindQuery(obj)
}

// BindURI 绑定 URI 参数
func BindURI(ctx *gin.Context, obj interface{}) error {
	return ctx.ShouldBindUri(obj)
}

