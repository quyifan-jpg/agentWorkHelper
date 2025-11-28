package main

import (
	"BackEnd/internal/config"
	"BackEnd/internal/handler"
	"BackEnd/internal/model"
	"BackEnd/internal/svc"
	"flag"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var configFile = flag.String("f", "etc/backend.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	viper.SetConfigFile(*configFile)
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	if err := viper.Unmarshal(&c); err != nil {
		panic(fmt.Errorf("fatal error unmarshal config: %w", err))
	}

	svcCtx := svc.NewServiceContext(c)
	
	// Auto Migrate
	if err := svcCtx.DB.AutoMigrate(&model.User{}); err != nil {
		panic(err)
	}

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// User Routes
	userHandler := handler.NewUserHandler(svcCtx)
	v1 := r.Group("/v1")
	{
		v1.POST("/user/register", userHandler.Register)
		v1.POST("/user/login", userHandler.Login)
	}

	fmt.Printf("Starting server at %s...\n", c.Addr)
	if err := r.Run(c.Addr); err != nil {
		panic(err)
	}
}
