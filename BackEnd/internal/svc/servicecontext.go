package svc

import (
	"BackEnd/internal/config"
	"BackEnd/internal/middleware"
	"BackEnd/internal/model"
	"fmt"

	// "log"

	"github.com/tmc/langchaingo/callbacks"

	"github.com/tmc/langchaingo/llms/openai"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config    config.Config
	DB        *gorm.DB
	Jwt       *middleware.Jwt
	Callbacks callbacks.Handler
	LLMs      *openai.LLM
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 连接数据库
	db, err := gorm.Open(mysql.Open(c.MySQL.DSN), &gorm.Config{})
	if err != nil {
		// 如果连接失败，尝试创建数据库
		dsnWithoutDB := "root:root@tcp(127.0.0.1:3306)/?charset=utf8mb4&parseTime=True&loc=Local"
		tempDB, err := gorm.Open(mysql.Open(dsnWithoutDB), &gorm.Config{})
		if err != nil {
			panic(err)
		}
		if err := tempDB.Exec("CREATE DATABASE IF NOT EXISTS aiworkhelper").Error; err != nil {
			panic(err)
		}
		// 重试连接
		db, err = gorm.Open(mysql.Open(c.MySQL.DSN), &gorm.Config{})
		if err != nil {
			panic(err)
		}
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(
		&model.User{},
		&model.Department{},
		&model.DepartmentUser{},
		&model.Todo{},
		&model.TodoRecord{},
		&model.UserTodo{},
		&model.Approval{},
		&model.Approver{},
		&model.ChatLog{},      // 聊天记录表
		&model.GroupMember{},  // 群聊成员表
		&model.Conversation{}, // 会话表
		&model.Participant{},  // 参与者表
	); err != nil {
		panic(err)
	}

	// 初始化 LLM
	if c.AI.BaseURL == "" {
		c.AI.BaseURL = "https://api.openai.com/v1"
	}
	if c.AI.Model == "" {
		c.AI.Model = "gpt-3.5-turbo"
	}

	llm, err := openai.New(
		openai.WithToken(c.AI.ApiKey),
		openai.WithBaseURL(c.AI.BaseURL),
		openai.WithModel(c.AI.Model),
	)
	if err != nil {
		panic(err)
		// log.Error().Err(err).Msg("Failed to initialize LLM")
		// 暂时不panic，允许LLM初始化失败（例如没有配置key）
	}

	return &ServiceContext{
		Config: c,
		DB:     db,
		Jwt:    middleware.NewJwt(c.Auth.Secret),
		LLMs:   llm,
	}
}

// GetBaseURL returns the base URL for internal API calls, handling the 0.0.0.0 case
func (s *ServiceContext) GetBaseURL() string {
	host := "127.0.0.1" // Default fallback
	if s.Config.Host != "0.0.0.0" && s.Config.Host != "" {
		host = s.Config.Host
	}
	return fmt.Sprintf("http://%s:%d", host, s.Config.Port)
}
