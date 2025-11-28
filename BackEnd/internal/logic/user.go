package logic

import (
	"BackEnd/internal/model"
	"BackEnd/internal/svc"
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type UserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserLogic {
	return &UserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserLogic) Register(username, password string) error {
	// Check if user exists
	var count int64
	l.svcCtx.DB.Model(&model.User{}).Where("name = ?", username).Count(&count)
	if count > 0 {
		return errors.New("username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Create user
	user := &model.User{
		Name:     username,
		Password: string(hashedPassword),
	}

	if err := l.svcCtx.DB.Create(user).Error; err != nil {
		return err
	}

	return nil
}

func (l *UserLogic) Login(username, password string) (string, error) {
	var user model.User
	if err := l.svcCtx.DB.Where("name = ?", username).First(&user).Error; err != nil {
		return "", errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid password")
	}

	// TODO: Generate JWT token
	return "dummy-token", nil
}
