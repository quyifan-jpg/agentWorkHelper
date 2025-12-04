package main

import (
	"flag"
	"fmt"
	"sync"

	"BackEnd/internal/config"
	"BackEnd/internal/handler/api"
	"BackEnd/internal/handler/ws"
	"BackEnd/internal/svc"
	"BackEnd/pkg/conf"
)

type Serve interface {
	Run() error
}

const (
	Api = "api"

	// add other module
)

var (
	configFile = flag.String("f", "./etc/local/api.yaml", "the config file")
	modeType   = flag.String("m", "api", "server run mod")
)

func main() {
	flag.Parse()

	var cfg config.Config
	conf.MustLoad(*configFile, &cfg)

	svcCtx := svc.NewServiceContext(cfg)

	var srv Serve
	switch *modeType {
	case Api:
		srv = api.NewApiHandler(svcCtx)
	// add other module case
	default:
		panic("请指定正确的服务")
	}

	// 使用 WaitGroup 管理多个服务
	var wg sync.WaitGroup

	// 启动 HTTP API 服务
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.Run(); err != nil {
			panic(fmt.Errorf("HTTP API server failed: %w", err))
		}
	}()

	// 启动 WebSocket 服务
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("WebSocket 服务启动失败: %v\n", r)
				panic(r) // 重新抛出 panic，让程序知道服务启动失败
			}
		}()
		wsServer := ws.NewWs(svcCtx)
		fmt.Println("正在启动 WebSocket 服务...")
		wsServer.Run()
	}()

	// 等待所有服务
	wg.Wait()
}
