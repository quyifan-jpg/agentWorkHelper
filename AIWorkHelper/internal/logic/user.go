/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package logic 提供业务逻辑层实现
package logic

import (
	"AIWorkHelper/internal/model"
	"AIWorkHelper/pkg/encrypt"
	"AIWorkHelper/pkg/token"
	"AIWorkHelper/pkg/xerr"

	"context"
	"errors"
	"strings"
	"time"

	"AIWorkHelper/internal/domain"
	"AIWorkHelper/internal/svc"
)

// User 用户业务逻辑接口
type User interface {
	// Login 用户登录验证
	Login(ctx context.Context, req *domain.LoginReq) (resp *domain.LoginResp, err error)
	// Info 根据ID获取用户信息
	Info(ctx context.Context, req *domain.IdPathReq) (resp *domain.User, err error)
	// Create 创建新用户
	Create(ctx context.Context, req *domain.User) (err error)
	// Edit 更新用户信息
	Edit(ctx context.Context, req *domain.User) (err error)
	// Delete 删除指定用户
	Delete(ctx context.Context, req *domain.IdPathReq) (err error)
	// List 分页查询用户列表
	List(ctx context.Context, req *domain.UserListReq) (resp *domain.UserListResp, err error)
	// UpdatePassword 更新用户密码
	UpdatePassword(ctx context.Context, req *domain.UpdatePasswordReq) (err error)
}

// user 用户业务逻辑实现
type user struct {
	svcCtx *svc.ServiceContext // 服务上下文
}

// NewUser 创建用户业务逻辑实例
func NewUser(svcCtx *svc.ServiceContext) User {
	return &user{
		svcCtx: svcCtx,
	}
}

// Login 用户登录验证
// 验证用户名和密码，成功后生成 JWT Token 并返回用户信息
func (l *user) Login(ctx context.Context, req *domain.LoginReq) (resp *domain.LoginResp, err error) {
	// 根据用户名查找用户
	user, err := l.svcCtx.UserModel.FindByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	// 验证密码是否正确
	if !encrypt.ValidatePasswordHash(req.Password, user.Password) {
		return nil, errors.New("密码错误")
	}

	// 生成 JWT Token
	now := time.Now().Unix()
	token, err := token.GetJwtToken(l.svcCtx.Config.Jwt.Secret, now, l.svcCtx.Config.Jwt.Expire, user.ID.Hex())
	if err != nil {
		return nil, xerr.WithMessagef(err, "Login|GetJwtToken err with name: %s", req.Name)
	}

	// 返回登录成功响应
	return &domain.LoginResp{
		Id:           user.ID.Hex(),                    // 用户ID
		Name:         user.Name,                        // 用户名
		AccessToken:  token,                            // JWT Token
		AccessExpire: l.svcCtx.Config.Jwt.Expire + now, // Token 过期时间
	}, nil
}

// 根据ID获取用户
func (l *user) Info(ctx context.Context, req *domain.IdPathReq) (resp *domain.User, err error) {
	return
}

// Create 创建新用户
// 检查用户名是否已存在，设置默认密码，加密后保存到数据库
func (l *user) Create(ctx context.Context, req *domain.User) (err error) {
	// 检查用户名是否已存在
	u, err := l.svcCtx.UserModel.FindByName(ctx, req.Name)
	if err != nil && !strings.Contains(err.Error(), model.ErrNotUser.Error()) {
		return xerr.WithMessagef(err, "user model find by user err req.name %s", req.Name)
	}
	if u != nil {
		return errors.New("已存在该用户")
	}

	// 设置用户密码，默认为 "123456"
	password := "123456"
	if len(req.Password) > 0 {
		password = req.Password
	}

	// 加密密码
	encryptPass, err := encrypt.GenPasswordHash([]byte(password))
	if err != nil {
		return xerr.WithMessagef(err, "encrypt.GenPasswordHash req.name %s", password)
	}

	// 创建用户记录
	// 如果前端传递了status，使用传递的值；否则默认为1（启用）
	status := 1
	if req.Status != 0 {
		status = req.Status
	}

	return l.svcCtx.UserModel.Insert(ctx, &model.User{
		Name:     req.Name,            // 用户名
		Password: string(encryptPass), // 加密后的密码
		Status:   status,              // 用户状态：0=禁用，1=启用
	})
}

// Edit 更新用户信息
// 支持更新用户名、密码（可选）、状态等信息
func (l *user) Edit(ctx context.Context, req *domain.User) (err error) {
	// 根据ID查找用户是否存在
	existingUser, err := l.svcCtx.UserModel.FindOne(ctx, req.Id)
	if err != nil {
		return xerr.WithMessagef(err, "Edit|FindOne err with id: %s", req.Id)
	}

	// 如果用户名有变化，检查新用户名是否已被其他用户使用
	if req.Name != "" && req.Name != existingUser.Name {
		u, err := l.svcCtx.UserModel.FindByName(ctx, req.Name)
		if err != nil && !strings.Contains(err.Error(), model.ErrNotUser.Error()) {
			return xerr.WithMessagef(err, "Edit|FindByName err with name: %s", req.Name)
		}
		if u != nil && u.ID.Hex() != req.Id {
			return errors.New("用户名已被占用")
		}
		existingUser.Name = req.Name
	}

	// 如果提供了新密码，则更新密码
	if req.Password != "" {
		encryptPass, err := encrypt.GenPasswordHash([]byte(req.Password))
		if err != nil {
			return xerr.WithMessagef(err, "Edit|GenPasswordHash err")
		}
		existingUser.Password = string(encryptPass)
	}

	// 更新状态
	existingUser.Status = req.Status
	existingUser.UpdateAt = time.Now().Unix()

	// 保存到数据库
	return l.svcCtx.UserModel.Update(ctx, existingUser)
}

// Delete 删除指定用户
func (l *user) Delete(ctx context.Context, req *domain.IdPathReq) (err error) {
	return l.svcCtx.UserModel.Delete(ctx, req.Id)
}

// 分页查询用户
func (l *user) List(ctx context.Context, req *domain.UserListReq) (resp *domain.UserListResp, err error) {
	// 调用 Model 层查询用户列表
	users, count, err := l.svcCtx.UserModel.List(ctx, req)
	if err != nil {
		return nil, xerr.WithMessagef(err, "UserModel.List failed with req: %+v", req)
	}

	// 将 Model 层的 User 转换为 Domain 层的 User
	userList := make([]*domain.User, 0, len(users))
	for _, u := range users {
		userList = append(userList, &domain.User{
			Id:     u.ID.Hex(),
			Name:   u.Name,
			Status: u.Status,
			// 注意: 不返回密码字段,保证安全性
		})
	}

	// 返回用户列表响应
	return &domain.UserListResp{
		Count: count,
		List:  userList,
	}, nil
}

// 更新用户密码
func (l *user) UpdatePassword(ctx context.Context, req *domain.UpdatePasswordReq) (err error) {
	return
}
