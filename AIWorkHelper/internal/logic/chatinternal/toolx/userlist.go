/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package toolx

import (
	"AIWorkHelper/internal/domain"
	"AIWorkHelper/internal/svc"
	"AIWorkHelper/pkg/token"
	"context"
	"encoding/json"
	"fmt"
	"github.com/tmc/langchaingo/callbacks"
)

// UserList 用户列表查询工具，实现AI代理的用户信息查询功能
type UserList struct {
	svc      *svc.ServiceContext // 服务上下文
	callback callbacks.Handler   // 回调处理器，用于记录执行日志
}

// NewUserList 创建用户列表查询工具实例
func NewUserList(svc *svc.ServiceContext) *UserList {
	return &UserList{
		svc:      svc,
		callback: svc.Callbacks,
	}
}

// Name 返回工具名称，用于AI代理识别
func (u *UserList) Name() string {
	return "user_list"
}

// Description 返回工具描述和使用说明
func (u *UserList) Description() string {
	return `
	a user list query interface.
	use when you need to find user information by name or get all users.
	this tool helps you convert user names to user IDs.
	input: JSON string with optional "name" field to filter by user name. example: {"name": "王员工"}
	if input is empty or {}, returns all active users.
	output: JSON array of users with fields: id (user ID string), name (user name), status (1=active, 0=disabled)
	keep Chinese output.
`
}

// Call 执行用户列表查询操作
func (u *UserList) Call(ctx context.Context, input string) (string, error) {
	// 记录工具调用日志
	if u.callback != nil {
		u.callback.HandleText(ctx, "user list query start : "+input)
	}

	// 解析输入参数
	var req domain.UserListReq
	if input != "" && input != "{}" {
		if err := json.Unmarshal([]byte(input), &req); err != nil {
			return "", fmt.Errorf("invalid input format: %w", err)
		}
	}

	// 设置默认分页参数
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Count == 0 {
		req.Count = 100 // 查询前100个用户
	}

	// 查询用户列表
	users, _, err := u.svc.UserModel.List(ctx, &req)
	if err != nil {
		return "", fmt.Errorf("failed to query user list: %w", err)
	}

	// 构建返回结果
	type UserInfo struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Status int    `json:"status"` // 1=启用（在线），0=禁用（离线）
	}

	result := make([]UserInfo, 0, len(users))
	currentUserId := token.GetUId(ctx)

	for _, user := range users {
		// 跳过当前用户自己
		if user.ID.Hex() == currentUserId {
			continue
		}

		result = append(result, UserInfo{
			ID:     user.ID.Hex(),
			Name:   user.Name,
			Status: user.Status,
		})
	}

	// 将结果序列化为JSON
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to serialize result: %w", err)
	}

	return Success + "\nuser list:\n" + string(jsonResult), nil
}
