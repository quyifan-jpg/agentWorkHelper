/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package timeutil

import "time"

func Format(date int64) string {
	return time.Unix(date, 0).Format("2006-01-02")
}
