/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package token

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

var TestSecretKey = "aiworkhelper"

func TestGenToken(t *testing.T) {
	now := time.Now().Unix()
	t.Log(GetJwtToken(TestSecretKey, now, 60*60*60, "1"))
}

func TestVerifyJWTToken(t *testing.T) {

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTc2MTg0NjMsImlhdCI6MTcxNzQwMjQ2MywiaW1vb2MuY29tIjoiMSJ9.anVWrthElU1ZS34UcFpE380aSvp30KtWq1_CIl6YnKo"

	r := &http.Request{
		Header: make(http.Header),
	}
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))

	t.Log(VerifyJWTToken(TestSecretKey, r))
}
