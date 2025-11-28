/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
/**
 * @author: dn-jinmin/dn-jinmin
 * @doc:
 */

package chatinternal

import (
	"AIWorkHelper/pkg/token"
	"context"
	"github.com/tmc/langchaingo/chains"
	"testing"
)

func TestChatLogHandle_transform(t *testing.T) {
	chat := NewChatLogHandle(svcTest)
	tokenStr := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzI2MTE2NTksImlhdCI6MTcyMzk3MTY1OSwiaW1vb2MuY29tIjoiNjZhZjUxYjUxNGZkMzVlMjQwYzlhYjkyIn0.3Q7BQzegPz-oNebf_K3z61yBuoJqqowmr2KgWFEvOaQ"
	ctx := context.Background()
	ctx = context.WithValue(ctx, token.Authorization, tokenStr)

	res, err := chat.transform(ctx, map[string]any{
		"relationId": "PU0zXtG2ePJzJpkfbE5gm/",
		"startTime":  1724084938,
		"endTime":    1724085498,
	}, chains.WithCallback(svcTest.Callbacks))
	if err != nil {
		t.Fatal(err)
	}

	t.Log(res)
}
