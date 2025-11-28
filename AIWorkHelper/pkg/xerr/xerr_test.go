/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package xerr

import (
	"errors"
	"testing"
)

func TestWithMessage(t *testing.T) {
	err := WithMessage(errors.New("测试"), "测试输出")
	t.Log(err)

	type cause interface {
		Cause() error
	}

	e, ok := err.(cause)
	if ok {
		t.Log(e.Cause().Error())
	}
}
