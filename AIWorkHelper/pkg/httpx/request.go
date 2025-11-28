/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package httpx

import "github.com/gin-gonic/gin"

func BindAndValidate(ctx *gin.Context, v any) error {
	if err := ctx.ShouldBind(v); err != nil {
		return err
	}

	if err := ctx.ShouldBindUri(v); err != nil {
		return err
	}

	return nil
}
