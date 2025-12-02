package logic

import (
	"BackEnd/internal/domain"
	"BackEnd/internal/model"
	"BackEnd/internal/svc"
	"BackEnd/pkg/jwt"
	"context"
	"errors"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// UserLogic 用户业务逻辑接口
type UserLogic interface {
	Register(ctx context.Context, username, password string) error
	Login(ctx context.Context, username, password string) (*domain.LoginResp, error)
	GetInfo(ctx context.Context, userID uint) (interface{}, error)
	GetInfoByID(ctx context.Context, userID string) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID uint, name string) error
	ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error
	List(ctx context.Context, req *domain.UserListReq) (*domain.UserListResp, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, userID string) error
}

type userLogic struct {
	svcCtx *svc.ServiceContext
}

// NewUserLogic 创建用户业务逻辑实例
func NewUser(svcCtx *svc.ServiceContext) UserLogic {
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
func (l *userLogic) Login(ctx context.Context, username, password string) (*domain.LoginResp, error) {
	var user model.User
	if err := l.svcCtx.DB.WithContext(ctx).Where("name = ?", username).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}

	// 生成 JWT token
	token, err := jwt.GenerateToken(user.ID, l.svcCtx.Config.Auth.Secret, l.svcCtx.Config.Auth.Expire)
	if err != nil {
		return nil, err
	}

	// 返回完整登录信息
	now := time.Now().Unix()
	return &domain.LoginResp{
		Status:       1,
		Id:           strconv.Itoa(int(user.ID)),
		Name:         user.Name,
		AccessToken:  token,
		AccessExpire: now + l.svcCtx.Config.Auth.Expire,
		RefreshAfter: now + l.svcCtx.Config.Auth.Expire/2,
	}, nil
}

// GetInfo 获取用户信息
func (l *userLogic) GetInfo(ctx context.Context, userID uint) (interface{}, error) {
	var user model.User
	if err := l.svcCtx.DB.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}

	return map[string]interface{}{
		"id":      user.ID,
		"name":    user.Name,
		"status":  user.Status,
		"isAdmin": user.IsAdmin,
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

// ChangePassword 修改密码
func (l *userLogic) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	// 查找用户
	var user model.User
	if err := l.svcCtx.DB.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("old password is incorrect")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to encrypt new password")
	}

	// 更新密码
	user.Password = string(hashedPassword)
	if err := l.svcCtx.DB.WithContext(ctx).Save(&user).Error; err != nil {
		return errors.New("failed to update password")
	}

	return nil
}

// GetInfoByID 根据ID获取用户信息
func (l *userLogic) GetInfoByID(ctx context.Context, userID string) (*domain.User, error) {
	var user model.User
	if err := l.svcCtx.DB.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}

	return &domain.User{
		Id:     strconv.Itoa(int(user.ID)),
		Name:   user.Name,
		Status: user.Status,
	}, nil
}

// List 用户列表（分页 + 搜索）
func (l *userLogic) List(ctx context.Context, req *domain.UserListReq) (*domain.UserListResp, error) {
	// 设置默认分页参数
	page := req.Page
	if page <= 0 {
		page = 1
	}
	count := req.Count
	if count <= 0 {
		count = 10
	}

	query := l.svcCtx.DB.WithContext(ctx).Model(&model.User{})

	// 按用户名模糊搜索
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}

	// 按ID列表查询
	if len(req.Ids) > 0 {
		// 转换 string ID 到 uint
		var uintIds []uint
		for _, idStr := range req.Ids {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				uintIds = append(uintIds, uint(id))
			}
		}
		if len(uintIds) > 0 {
			query = query.Where("id IN ?", uintIds)
		}
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	var users []model.User
	offset := (page - 1) * count
	if err := query.Offset(offset).Limit(count).Find(&users).Error; err != nil {
		return nil, err
	}

	// 转换为响应格式
	var userList []*domain.User
	for _, u := range users {
		userList = append(userList, &domain.User{
			Id:     strconv.Itoa(int(u.ID)),
			Name:   u.Name,
			Status: u.Status,
		})
	}

	return &domain.UserListResp{
		Count: total,
		List:  userList,
	}, nil
}

// Create 创建用户
func (l *userLogic) Create(ctx context.Context, user *domain.User) error {
	// 检查用户名是否已存在
	var count int64
	l.svcCtx.DB.WithContext(ctx).Model(&model.User{}).Where("name = ?", user.Name).Count(&count)
	if count > 0 {
		return errors.New("user already exists")
	}

	// 设置默认密码
	password := user.Password
	if password == "" {
		password = "123456" // 默认密码
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 创建用户
	newUser := &model.User{
		Name:     user.Name,
		Password: string(hashedPassword),
		Status:   user.Status,
	}

	if err := l.svcCtx.DB.WithContext(ctx).Create(newUser).Error; err != nil {
		return err
	}

	return nil
}

// Update 更新用户
func (l *userLogic) Update(ctx context.Context, user *domain.User) error {
	var existingUser model.User
	if err := l.svcCtx.DB.WithContext(ctx).Where("id = ?", user.Id).First(&existingUser).Error; err != nil {
		return errors.New("user not found")
	}

	// 更新字段
	if user.Name != "" {
		existingUser.Name = user.Name
	}
	existingUser.Status = user.Status

	// 如果提供了新密码，则更新密码
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		existingUser.Password = string(hashedPassword)
	}

	if err := l.svcCtx.DB.WithContext(ctx).Save(&existingUser).Error; err != nil {
		return err
	}

	return nil
}

// Delete 删除用户（软删除）
func (l *userLogic) Delete(ctx context.Context, userID string) error {
	// 使用 GORM 软删除
	if err := l.svcCtx.DB.WithContext(ctx).Delete(&model.User{}, userID).Error; err != nil {
		return err
	}
	return nil
}
