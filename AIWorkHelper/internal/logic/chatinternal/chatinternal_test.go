/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package chatinternal

import (
	"AIWorkHelper/internal/config"
	"AIWorkHelper/internal/svc"
	"AIWorkHelper/pkg/conf"
	"fmt"
	"path/filepath"
)

var svcTest *svc.ServiceContext

func init() {
	var c config.Config
	conf.MustLoad(filepath.Join("../../../etc/api.yaml"), &c)

	fmt.Println(c)

	svc, err := svc.NewServiceContext(c)
	if err != nil {
		panic(err)
	}

	svcTest = svc
}
