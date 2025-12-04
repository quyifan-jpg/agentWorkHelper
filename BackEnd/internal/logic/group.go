package logic

import (
	"BackEnd/internal/model"
	"BackEnd/internal/svc"
	"BackEnd/pkg/util"
	"BackEnd/pkg/xerr"
	"context"
	"errors"

	"github.com/rs/zerolog/log"
)

type Group interface {
	// GetGroupMemberIds 获取群成员ID列表
	GetGroupMemberIds(ctx context.Context, groupId string) ([]string, error)
	// AddMember 添加单个群成员
	AddMember(ctx context.Context, groupId string, userId string) error
	// AddMembers 批量添加群成员
	AddMembers(ctx context.Context, groupId string, userIds []string) error
	// RemoveMember 移除群成员
	RemoveMember(ctx context.Context, groupId string, userId string) error
	// IsMember 检查是否是群成员
	IsMember(ctx context.Context, groupId string, userId string) (bool, error)
	// GetMemberCount 获取群成员数量
	GetMemberCount(ctx context.Context, groupId string) (int64, error)
}

type group struct {
	svcCtx *svc.ServiceContext
}

func NewGroup(svcCtx *svc.ServiceContext) Group {
	return &group{
		svcCtx: svcCtx,
	}
}

// GetGroupMemberIds 获取群成员ID列表
func (l *group) GetGroupMemberIds(ctx context.Context, groupId string) ([]string, error) {
	var members []model.GroupMember
	if err := l.svcCtx.DB.WithContext(ctx).
		Where("group_id = ?", groupId).
		Find(&members).Error; err != nil {
		log.Error().Err(err).Str("groupId", groupId).Msg("查询群成员失败")
		return nil, xerr.New(err)
	}

	// 提取用户ID列表并转换为字符串
	userIds := make([]string, 0, len(members))
	for _, member := range members {
		userIds = append(userIds, util.UintToString(member.UserId))
	}

	return userIds, nil
}

// AddMember 添加单个群成员
func (l *group) AddMember(ctx context.Context, groupId string, userId string) error {
	// 转换用户ID
	userID, err := util.StringToUint(userId)
	if err != nil {
		return xerr.New(errors.New("无效的用户ID"))
	}

	// 检查是否已存在
	var count int64
	if err := l.svcCtx.DB.WithContext(ctx).
		Model(&model.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupId, userID).
		Count(&count).Error; err != nil {
		log.Error().Err(err).Str("groupId", groupId).Str("userId", userId).Msg("检查群成员失败")
		return xerr.New(err)
	}

	if count > 0 {
		log.Debug().Str("groupId", groupId).Str("userId", userId).Msg("用户已在群中")
		return nil // 已存在，不报错
	}

	// 创建群成员记录
	member := &model.GroupMember{
		GroupId: groupId,
		UserId:  userID,
	}

	if err := l.svcCtx.DB.WithContext(ctx).Create(member).Error; err != nil {
		log.Error().Err(err).Str("groupId", groupId).Str("userId", userId).Msg("添加群成员失败")
		return xerr.New(err)
	}

	log.Info().Str("groupId", groupId).Str("userId", userId).Msg("成功添加群成员")
	return nil
}

// AddMembers 批量添加群成员
func (l *group) AddMembers(ctx context.Context, groupId string, userIds []string) error {
	for _, userId := range userIds {
		if err := l.AddMember(ctx, groupId, userId); err != nil {
			log.Error().Err(err).Str("groupId", groupId).Str("userId", userId).Msg("批量添加群成员失败")
			// 继续处理其他成员，不中断
		}
	}

	log.Info().Str("groupId", groupId).Int("count", len(userIds)).Msg("批量添加群成员完成")
	return nil
}

// RemoveMember 移除群成员
func (l *group) RemoveMember(ctx context.Context, groupId string, userId string) error {
	// 转换用户ID
	userID, err := util.StringToUint(userId)
	if err != nil {
		return xerr.New(errors.New("无效的用户ID"))
	}

	if err := l.svcCtx.DB.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupId, userID).
		Delete(&model.GroupMember{}).Error; err != nil {
		log.Error().Err(err).Str("groupId", groupId).Str("userId", userId).Msg("移除群成员失败")
		return xerr.New(err)
	}

	log.Info().Str("groupId", groupId).Str("userId", userId).Msg("成功移除群成员")
	return nil
}

// IsMember 检查是否是群成员
func (l *group) IsMember(ctx context.Context, groupId string, userId string) (bool, error) {
	// 转换用户ID
	userID, err := util.StringToUint(userId)
	if err != nil {
		return false, xerr.New(errors.New("无效的用户ID"))
	}

	var count int64
	if err := l.svcCtx.DB.WithContext(ctx).
		Model(&model.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupId, userID).
		Count(&count).Error; err != nil {
		log.Error().Err(err).Str("groupId", groupId).Str("userId", userId).Msg("检查群成员失败")
		return false, xerr.New(err)
	}

	return count > 0, nil
}

// GetMemberCount 获取群成员数量
func (l *group) GetMemberCount(ctx context.Context, groupId string) (int64, error) {
	var count int64
	if err := l.svcCtx.DB.WithContext(ctx).
		Model(&model.GroupMember{}).
		Where("group_id = ?", groupId).
		Count(&count).Error; err != nil {
		log.Error().Err(err).Str("groupId", groupId).Msg("统计群成员数量失败")
		return 0, xerr.New(err)
	}

	return count, nil
}

