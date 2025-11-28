package main

import (
	_ "BackEnd/docs" // Import generated docs
	"BackEnd/internal/config"
	"BackEnd/internal/handler/api"
	"BackEnd/internal/model"
	"BackEnd/internal/svc"
	"flag"
	"fmt"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var configFile = flag.String("f", "etc/backend.yaml", "the config file")

// @title           AIWorkHelper API
// @version         1.0
// @description     This is the backend API for AIWorkHelper.
// @host            localhost:8889
// @BasePath        /v1
func main() {
	flag.Parse()

	// 加载配置
	var c config.Config
	viper.SetConfigFile(*configFile)
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	if err := viper.Unmarshal(&c); err != nil {
		panic(fmt.Errorf("fatal error unmarshal config: %w", err))
	}

	// 初始化服务上下文
	svcCtx := svc.NewServiceContext(c)

	// 数据库迁移
	if err := svcCtx.DB.AutoMigrate(&model.User{}); err != nil {
		panic(err)
	}

	// 创建 API Handler
	apiHandler := api.NewApiHandler(svcCtx)

	// 添加 Swagger 文档
	engine := apiHandler.GetEngine()
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查
	engine.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// 启动服务
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	fmt.Printf("Starting server at %s...\n", addr)
	fmt.Printf("Swagger UI: http://%s/swagger/index.html\n", addr)
	if err := apiHandler.Run(); err != nil {
		panic(err)
	}
}
