package logic

import (
	"BackEnd/internal/model"
	"BackEnd/internal/svc"
	"BackEnd/pkg/jwt"
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// UserLogic 用户业务逻辑接口
type UserLogic interface {
	Register(ctx context.Context, username, password string) error
	Login(ctx context.Context, username, password string) (string, error)
	GetInfo(ctx context.Context, userID uint) (interface{}, error)
	UpdateProfile(ctx context.Context, userID uint, name string) error
}

type userLogic struct {
	svcCtx *svc.ServiceContext
}

// NewUserLogic 创建用户业务逻辑实例
func NewUserLogic(svcCtx *svc.ServiceContext) UserLogic {
	return &userLogic{
		svcCtx: svcCtx,
	}
}

// Register 用户注册
func (l *userLogic) Register(ctx context.Context, username, password string) error {
	// 检查用户是否已存在
	var count int64
	l.svcCtx.DB.WithContext(ctx).Model(&model.User{}).Where("name = ?", username).Count(&count)
	if count > 0 {
		return errors.New("username already exists")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 创建用户
	user := &model.User{
		Name:     username,
		Password: string(hashedPassword),
	}

	if err := l.svcCtx.DB.WithContext(ctx).Create(user).Error; err != nil {
		return err
	}

	return nil
}

// Login 用户登录
func (l *userLogic) Login(ctx context.Context, username, password string) (string, error) {
	var user model.User
	if err := l.svcCtx.DB.WithContext(ctx).Where("name = ?", username).First(&user).Error; err != nil {
		return "", errors.New("user not found")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid password")
	}

	// 生成 JWT token
	return jwt.GenerateToken(user.ID, l.svcCtx.Config.Auth.Secret, l.svcCtx.Config.Auth.Expire)
}

// GetInfo 获取用户信息
func (l *userLogic) GetInfo(ctx context.Context, userID uint) (interface{}, error) {
	var user model.User
	if err := l.svcCtx.DB.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}

	return map[string]interface{}{
		"id":       user.ID,
		"username": user.Name,
		"status":   user.Status,
		"isAdmin":  user.IsAdmin,
	}, nil
}

// UpdateProfile 更新用户资料
func (l *userLogic) UpdateProfile(ctx context.Context, userID uint, name string) error {
	if name == "" {
		return nil // 如果没有提供新名称，不更新
	}

	var user model.User
	if err := l.svcCtx.DB.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	user.Name = name
	if err := l.svcCtx.DB.WithContext(ctx).Save(&user).Error; err != nil {
		return err
	}

	return nil
}
