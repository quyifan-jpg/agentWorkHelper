/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package svc 提供服务上下文和依赖注入
package svc

import (
	"AIWorkHelper/internal/config"
	"AIWorkHelper/internal/middleware"
	"AIWorkHelper/internal/model"
	"AIWorkHelper/pkg/langchain/callbackx"
	"AIWorkHelper/pkg/mongoutil"
	"AIWorkHelper/pkg/token"
	"context"
	"errors"

	"gitee.com/dn-jinmin/tlog"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms/openai"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrAuth = errors.New("不具有权限")

// ServiceContext 服务上下文，管理所有依赖组件
type ServiceContext struct {
	Config config.Config // 应用配置

	// 数据库和模型
	Mongo               *mongo.Database           // MongoDB 数据库实例
	UserModel           model.UserModel           // 用户模型
	DepartmentModel     model.DepartmentModel     // 部门模型
	DepartmentUserModel model.DepartmentUserModel //部门用户模型
	TodoModel           model.TodoModel           // 待办模型
	UserTodoModel       model.UserTodoModel       // 用户待办模型
	TodoRecordModel     model.TodoRecordModel     // 待办记录模型
	ApprovalModel       model.ApprovalModel       // 审批模型
	ChatLogModel        model.ChatLogModel        // 聊天模型
	GroupMemberModel    model.GroupMemberModel    // 群成员模型

	// 中间件
	Jwt *middleware.Jwt // JWT 认证中间件

	LLMs      *openai.LLM
	Callbacks callbacks.Handler

	Auth func(ctx context.Context) error
}

// NewServiceContext 创建服务上下文实例
func NewServiceContext(c config.Config) (*ServiceContext, error) {
	// 初始化 MongoDB 连接
	mongoDb, err := mongoutil.MongodbDatabase(&mongoutil.MongodbConfig{
		User:     c.Mongo.User,
		Password: c.Mongo.Password,
		Hosts:    c.Mongo.Hosts,
		Port:     c.Mongo.Port,
		Database: c.Mongo.Database,
	})
	if err != nil {
		return nil, err
	}

	logger := tlog.NewLogger()
	callbacks := callbacks.CombiningHandler{
		Callbacks: []callbacks.Handler{
			callbackx.NewLogHandler(logger),
			callbackx.NewTitTokenHandle(logger),
		}}

	opts := []openai.Option{
		openai.WithBaseURL(c.Langchain.Url),
		openai.WithToken(c.Langchain.ApiKey),
		openai.WithCallback(callbacks),
		openai.WithEmbeddingModel("text-embedding-3-small"), // 设置默认嵌入模型
		openai.WithModel("gpt-4o"),                          // 设置默认模型
	}

	llm, err := openai.New(opts...)
	if err != nil {
		return nil, err
	}

	userModel := model.NewUserModel(mongoDb)
	// 构建服务上下文
	svc := &ServiceContext{
		Config:              c,
		Mongo:               mongoDb,
		UserModel:           model.NewUserModel(mongoDb),           // 初始化用户模型
		DepartmentModel:     model.NewDepartmentModel(mongoDb),     // 初始化部门模型
		DepartmentUserModel: model.NewDepartmentUserModel(mongoDb), // 初始化部门用户模型
		TodoModel:           model.NewTodoModel(mongoDb),           // 初始化待办模型
		UserTodoModel:       model.NewUserTodoModel(mongoDb),       // 初始化用户待办模型
		TodoRecordModel:     model.NewTodoRecordModel(mongoDb),     // 初始化待办记录模型
		ApprovalModel:       model.NewApprovalModel(mongoDb),       // 初始化审批模型
		ChatLogModel:        model.NewChatLogModel(mongoDb),        // 初始化聊天模型
		GroupMemberModel:    model.NewGroupMemberModel(mongoDb),    // 初始化群成员模型
		Jwt:                 middleware.NewJwt(c.Jwt.Secret),       // 初始化 JWT 中间件

		LLMs:      llm,
		Callbacks: callbacks,
		Auth: func(ctx context.Context) error {
			uid := token.GetUId(ctx)
			if uid == "" {
				return ErrAuth
			}

			user, err := userModel.FindOne(ctx, uid)
			if err != nil {
				return err
			}

			if !user.IsAdmin {
				return ErrAuth
			}

			return nil
		},
	}

	// 初始化系统用户数据
	if err := initUser(svc); err != nil {
		return nil, err
	}

	return svc, nil
}

// initUser 初始化系统管理员用户
func initUser(svc *ServiceContext) error {
	ctx := context.Background()

	// 检查是否已存在管理员用户
	systemUser, err := svc.UserModel.FindAdminUser(ctx)
	if err != nil && err != model.ErrNotUser {
		return err
	}

	// 如果管理员用户已存在，直接返回
	if systemUser != nil {
		return nil
	}

	// 创建默认管理员用户
	return svc.UserModel.Insert(ctx, &model.User{
		Name:     "root",                                                         // 默认管理员用户名
		Password: "$2a$10$/UfHc5FZSS.gj7C7uWIOWeTao//mq.OMdmgSpW09AbCopkWPwl59e", // 加密后的密码，原密码123456
		Status:   0,                                                              // 用户状态
		IsAdmin:  true,                                                           // 管理员标识
	})
}
