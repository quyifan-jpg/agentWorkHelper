/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package main

import (
	"AIWorkHelper/internal/handler/ws"
	"flag"
	"fmt"
	"sync"

	"AIWorkHelper/internal/config"
	"AIWorkHelper/internal/handler/api"
	"AIWorkHelper/internal/svc"
	"AIWorkHelper/pkg/conf"
)

type Serve interface {
	Run() error
}

const (
	Api = "api"

	// add other module
)

var (
	configFile = flag.String("f", "./etc/api.yaml", "the config file")
	modeType   = flag.String("m", "api", "server run mod")

	sw sync.WaitGroup
)

func main() {
	flag.Parse()

	var cfg config.Config
	conf.MustLoad(*configFile, &cfg)

	// 调试：打印Langchain配置信息
	fmt.Printf("Langchain URL: %s\n", cfg.Langchain.Url)
	fmt.Printf("Langchain ApiKey: %s\n", cfg.Langchain.ApiKey)

	// 初始化唯一服务上下文
	svcContext, err := svc.NewServiceContext(cfg)
	if err != nil {
		panic(err)
	}

	sw.Add(1)
	// 运行http服务
	go func() {
		defer sw.Done()
		srv := api.NewHandle(svcContext)
		srv.Run()
	}()

	sw.Add(1)
	// 运行websocket服务
	go func() {
		defer sw.Done()
		srv := ws.NewWs(svcContext)
		srv.Run()
	}()

	sw.Wait()
}
