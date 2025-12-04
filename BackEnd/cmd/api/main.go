package main

import (
	_ "BackEnd/docs" // Import generated docs
	"BackEnd/internal/config"
	"BackEnd/internal/handler/api"
	"BackEnd/internal/handler/ws"
	"BackEnd/internal/model"
	"BackEnd/internal/svc"
	"context"
	"flag"
	"fmt"
	"sync"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/crypto/bcrypt"

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

	// 初始化 root 用户
	if err := initRootUser(svcCtx); err != nil {
		panic(fmt.Errorf("failed to init root user: %w", err))
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

	// 使用 WaitGroup 管理多个服务
	var wg sync.WaitGroup

	// 启动 HTTP API 服务
	wg.Add(1)
	go func() {
		defer wg.Done()
		addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
		fmt.Printf("Starting HTTP API server at %s...\n", addr)
		fmt.Printf("Swagger UI: http://%s/swagger/index.html\n", addr)
		if err := apiHandler.Run(); err != nil {
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

// initRootUser 初始化 root 管理员用户
func initRootUser(svcCtx *svc.ServiceContext) error {
	ctx := context.Background()

	// 检查是否已存在 root 用户
	var count int64
	svcCtx.DB.WithContext(ctx).Model(&model.User{}).Where("name = ?", "root").Count(&count)
	if count > 0 {
		fmt.Println("Root user already exists, skipping initialization")
		return nil
	}

	// 加密密码 "123456"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 创建 root 用户
	rootUser := &model.User{
		Name:     "root",
		Password: string(hashedPassword),
		Status:   0,      // 0: normal
		IsAdmin:  true,   // 设置为管理员
	}

	if err := svcCtx.DB.WithContext(ctx).Create(rootUser).Error; err != nil {
		return fmt.Errorf("failed to create root user: %w", err)
	}

	fmt.Println("Root user initialized successfully (username: root, password: 123456)")
	return nil
}
