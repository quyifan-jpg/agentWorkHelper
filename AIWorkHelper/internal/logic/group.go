/**
 * @author: 公众号:IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package logic 提供群聊管理相关的业务逻辑处理
package logic

import (
	"AIWorkHelper/internal/model"
	"AIWorkHelper/internal/svc"
	"context"

	"gitee.com/dn-jinmin/tlog"
)

// GroupService 群聊管理服务接口，定义群聊管理的核心方法
// 对应Java版本: GroupService接口
type GroupService interface {
	GetGroupMemberIds(ctx context.Context, groupId string) ([]string, error)     // 获取群成员ID列表
	AddMember(ctx context.Context, groupId, userId string) error                 // 添加单个群成员
	RemoveMember(ctx context.Context, groupId, userId string) error              // 移除群成员
	IsMember(ctx context.Context, groupId, userId string) (bool, error)          // 检查是否是群成员
	GetMemberCount(ctx context.Context, groupId string) (int64, error)           // 获取群成员数量
	AddMembers(ctx context.Context, groupId string, userIds []string) error      // 批量添加群成员
}

// groupService 群聊管理服务实现
// 对应Java版本: GroupServiceImpl类
type groupService struct {
	svc *svc.ServiceContext
}

// NewGroupService 创建群聊管理服务实例
func NewGroupService(svc *svc.ServiceContext) GroupService {
	return &groupService{
		svc: svc,
	}
}

// GetGroupMemberIds 获取群成员ID列表
// 对应Java版本: getGroupMemberIds方法
func (s *groupService) GetGroupMemberIds(ctx context.Context, groupId string) ([]string, error) {
	tlog.DebugfCtx(ctx, "GetGroupMemberIds", "groupId=%s", groupId)

	members, err := s.svc.GroupMemberModel.FindByGroupId(ctx, groupId)
	if err != nil {
		tlog.ErrorfCtx(ctx, "GetGroupMemberIds", "查询群成员失败 groupId=%s err=%v", groupId, err)
		return nil, err
	}

	// 提取用户ID列表
	userIds := make([]string, 0, len(members))
	for _, member := range members {
		userIds = append(userIds, member.UserId)
	}

	return userIds, nil
}

// AddMember 添加单个群成员
// 对应Java版本: addMember方法
func (s *groupService) AddMember(ctx context.Context, groupId, userId string) error {
	tlog.InfofCtx(ctx, "AddMember", "groupId=%s userId=%s", groupId, userId)

	// 检查是否已存在
	exists, err := s.svc.GroupMemberModel.ExistsByGroupIdAndUserId(ctx, groupId, userId)
	if err != nil {
		tlog.ErrorfCtx(ctx, "AddMember", "检查群成员失败 groupId=%s userId=%s err=%v", groupId, userId, err)
		return err
	}

	if exists {
		tlog.DebugfCtx(ctx, "AddMember", "用户已在群中 groupId=%s userId=%s", groupId, userId)
		return nil
	}

	// ���建群成员记录
	member := &model.GroupMember{
		GroupId: groupId,
		UserId:  userId,
	}

	err = s.svc.GroupMemberModel.Insert(ctx, member)
	if err != nil {
		tlog.ErrorfCtx(ctx, "AddMember", "添加群成员失败 groupId=%s userId=%s err=%v", groupId, userId, err)
		return err
	}

	tlog.InfofCtx(ctx, "AddMember", "成功添加群成员 groupId=%s userId=%s", groupId, userId)
	return nil
}

// RemoveMember 移除群成员
// 对应Java版本: removeMember方法
func (s *groupService) RemoveMember(ctx context.Context, groupId, userId string) error {
	tlog.InfofCtx(ctx, "RemoveMember", "groupId=%s userId=%s", groupId, userId)

	err := s.svc.GroupMemberModel.DeleteByGroupIdAndUserId(ctx, groupId, userId)
	if err != nil {
		tlog.ErrorfCtx(ctx, "RemoveMember", "移除群成员失败 groupId=%s userId=%s err=%v", groupId, userId, err)
		return err
	}

	return nil
}

// IsMember 检查是否是群成员
// 对应Java版本: isMember方法
func (s *groupService) IsMember(ctx context.Context, groupId, userId string) (bool, error) {
	return s.svc.GroupMemberModel.ExistsByGroupIdAndUserId(ctx, groupId, userId)
}

// GetMemberCount 获取群成员数量
// 对应Java版本: getMemberCount方法
func (s *groupService) GetMemberCount(ctx context.Context, groupId string) (int64, error) {
	return s.svc.GroupMemberModel.CountByGroupId(ctx, groupId)
}

// AddMembers 批量添加群成员
// 对应Java版本: addMembers方法
func (s *groupService) AddMembers(ctx context.Context, groupId string, userIds []string) error {
	tlog.InfofCtx(ctx, "AddMembers", "groupId=%s count=%d", groupId, len(userIds))

	for _, userId := range userIds {
		// 检查是否已存在
		exists, err := s.svc.GroupMemberModel.ExistsByGroupIdAndUserId(ctx, groupId, userId)
		if err != nil {
			tlog.ErrorfCtx(ctx, "AddMembers", "检查群成员失败 groupId=%s userId=%s err=%v", groupId, userId, err)
			continue
		}

		if exists {
			tlog.DebugfCtx(ctx, "AddMembers", "用户已在群中，跳过 groupId=%s userId=%s", groupId, userId)
			continue
		}

		// 创建群成员记录
		member := &model.GroupMember{
			GroupId: groupId,
			UserId:  userId,
		}

		err = s.svc.GroupMemberModel.Insert(ctx, member)
		if err != nil {
			tlog.ErrorfCtx(ctx, "AddMembers", "添加群成员失败 groupId=%s userId=%s err=%v", groupId, userId, err)
			continue
		}
	}

	tlog.InfofCtx(ctx, "AddMembers", "批量添加群成员完成 groupId=%s count=%d", groupId, len(userIds))
	return nil
}
