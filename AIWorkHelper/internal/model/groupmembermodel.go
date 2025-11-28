/**
 * @author: 公众号:IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package model 提供群聊成员的数据库操作接口
package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// GroupMemberModel 群聊成员数据库操作接口
type GroupMemberModel interface {
	// Insert 插入群成员记录
	Insert(ctx context.Context, data *GroupMember) error
	// FindByGroupId 根据群ID查询所有成员
	FindByGroupId(ctx context.Context, groupId string) ([]*GroupMember, error)
	// ExistsByGroupIdAndUserId 检查用户是否在群中
	ExistsByGroupIdAndUserId(ctx context.Context, groupId, userId string) (bool, error)
	// DeleteByGroupIdAndUserId 删除指定群的指定成员
	DeleteByGroupIdAndUserId(ctx context.Context, groupId, userId string) error
	// CountByGroupId 统计群成员数量
	CountByGroupId(ctx context.Context, groupId string) (int64, error)
}

type defaultGroupMemberModel struct {
	coll *mongo.Collection
}

// NewGroupMemberModel 创建群成员数据库操作实例
func NewGroupMemberModel(db *mongo.Database) GroupMemberModel {
	return &defaultGroupMemberModel{
		coll: db.Collection("group_member"),
	}
}

// Insert 插入群成员记录
func (m *defaultGroupMemberModel) Insert(ctx context.Context, data *GroupMember) error {
	now := time.Now().Unix()
	data.CreateAt = now
	data.UpdateAt = now

	_, err := m.coll.InsertOne(ctx, data)
	return err
}

// FindByGroupId 根据群ID查询所有成员
func (m *defaultGroupMemberModel) FindByGroupId(ctx context.Context, groupId string) ([]*GroupMember, error) {
	filter := bson.M{"groupId": groupId}

	cursor, err := m.coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var members []*GroupMember
	if err := cursor.All(ctx, &members); err != nil {
		return nil, err
	}

	return members, nil
}

// ExistsByGroupIdAndUserId 检查用户是否在群中
func (m *defaultGroupMemberModel) ExistsByGroupIdAndUserId(ctx context.Context, groupId, userId string) (bool, error) {
	filter := bson.M{
		"groupId": groupId,
		"userId":  userId,
	}

	count, err := m.coll.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// DeleteByGroupIdAndUserId 删除指定群的指定成员
func (m *defaultGroupMemberModel) DeleteByGroupIdAndUserId(ctx context.Context, groupId, userId string) error {
	filter := bson.M{
		"groupId": groupId,
		"userId":  userId,
	}

	_, err := m.coll.DeleteOne(ctx, filter)
	return err
}

// CountByGroupId 统计群成员数量
func (m *defaultGroupMemberModel) CountByGroupId(ctx context.Context, groupId string) (int64, error) {
	filter := bson.M{"groupId": groupId}
	return m.coll.CountDocuments(ctx, filter)
}
