package main

import (
	"flag"

	"BackEnd/internal/config"
	"BackEnd/internal/handler/api"
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

	srv.Run()
}
