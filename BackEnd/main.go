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

	svc, err := svc.NewServiceContext(cfg)
	if err != nil {
		panic(err)
	}

	var srv Serve
	switch *modeType {
	case Api:
		srv = api.NewHandle(svc)
	// add other module case
	default:
		panic("请指定正确的服务")
	}

	srv.Run()
}
