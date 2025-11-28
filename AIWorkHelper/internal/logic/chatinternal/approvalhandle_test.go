/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package chatinternal

import (
	"AIWorkHelper/pkg/langchain"
	"AIWorkHelper/pkg/token"
	"context"
	"testing"

	"github.com/tmc/langchaingo/chains"
)

func TestApproval(t *testing.T) {
	// 注意这里的token要换成root用户登陆的token
	tokenStr := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhaXdvcmtoZWxwZXIiOiI2OGY0OTJjMGY5ODk1NjFlZGU5MzIxYzkiLCJleHAiOjE3Njk0OTg4MjgsImlhdCI6MTc2MDg1ODgyOH0.lUPaCOMPQNPKDT2c0GGA0mlYn1q2aCnEYDvkp9roLds"

	ctx := context.Background()
	ctx = context.WithValue(ctx, token.Authorization, tokenStr)

	chat := NewApprovalHandle(svcTest)
	res, err := chat.baseChat.transform(ctx, map[string]any{
		langchain.Input: "提交一个明天上午请假审批",
		"history":       "",
	}, chains.WithCallback(svcTest.Callbacks))

	if err != nil {
		t.Fatal(err)
	}

	t.Log(res)
}
