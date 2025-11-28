/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package ws 提供WebSocket聊天功能的会话处理
package ws

import (
	"AIWorkHelper/internal/domain"
	"AIWorkHelper/internal/model"
	"context"

	"gitee.com/dn-jinmin/tlog"
	"github.com/gorilla/websocket"
)

// privateChat 处理私聊消息
// 将消息保存到数据库并发送给指定的接收者
func (s *Ws) privateChat(ctx context.Context, conn *websocket.Conn, req *domain.Message) error {
	// 调用聊天业务逻辑，保存私聊消息到数据库
	if err := s.chat.PrivateChat(ctx, req); err != nil {
		return err
	}

	// 将消息发送给接收者（注意：当前逻辑发送者看不到自己发送的消息）
	return s.sendByUids(ctx, req, req.RecvId)
}

// groupChat 处理群聊消息
// 将消息保存到数据库并根据群成员列表推送消息
// 支持多群聊：根据conversationId从数据库获取群成员列表，只向该群的成员推送消息
// 自动添加发送者到群成员（如果还不是成员），实现兼容性
func (s *Ws) groupChat(ctx context.Context, conn *websocket.Conn, req *domain.Message) error {
	// 调用聊天业务逻辑，保存群聊消息到数据库
	_, err := s.chat.GroupChat(ctx, req)
	if err != nil {
		return err
	}

	// 获取群聊ID (conversationId)
	groupId := req.ConversationId
	if groupId == "" {
		tlog.InfofCtx(ctx, "groupChat", "群聊消息缺少conversationId，无法广播: from=%s", req.SendId)
		return nil
	}

	// 自动将发送者添加到群成员（如果还不是成员）
	// 这样可以兼容前端直接发送群聊消息而不调用API创建群的情况
	senderId := req.SendId
	if senderId != "" {
		isMember, err := s.svc.GroupMemberModel.ExistsByGroupIdAndUserId(ctx, groupId, senderId)
		if err != nil {
			tlog.ErrorfCtx(ctx, "groupChat", "检查群成员失败 groupId=%s userId=%s err=%v", groupId, senderId, err)
		} else if !isMember {
			// 添加到群成员
			member := &model.GroupMember{
				GroupId: groupId,
				UserId:  senderId,
			}
			if err := s.svc.GroupMemberModel.Insert(ctx, member); err != nil {
				tlog.ErrorfCtx(ctx, "groupChat", "自动添加群成员失败 groupId=%s userId=%s err=%v", groupId, senderId, err)
			} else {
				tlog.InfofCtx(ctx, "groupChat", "自动添加发送者到群: groupId=%s userId=%s", groupId, senderId)
			}
		}
	}

	// 从数据库获取该群的所有成员ID
	members, err := s.svc.GroupMemberModel.FindByGroupId(ctx, groupId)
	if err != nil {
		tlog.ErrorfCtx(ctx, "groupChat", "查询群成员失败 groupId=%s err=%v", groupId, err)
		// 如果查询失败，回退到广播给所有人（兼容旧数据）
		return s.sendByUids(ctx, req)
	}

	if len(members) == 0 {
		tlog.InfofCtx(ctx, "groupChat", "群聊%s没有成员记录，将广播给所有在线用户", groupId)
		// 如果数据库中没有群成员记录，回退到广播给所有人（兼容旧数据）
		return s.sendByUids(ctx, req)
	}

	// 提取成员的用户ID列表
	memberIds := make([]string, 0, len(members))
	for _, member := range members {
		memberIds = append(memberIds, member.UserId)
	}

	// 只向该群的成员发送消息
	err = s.sendByUids(ctx, req, memberIds...)
	if err != nil {
		tlog.ErrorfCtx(ctx, "groupChat", "发送群消息失败 groupId=%s err=%v", groupId, err)
		return err
	}

	// 统计在线成员数量
	onlineCount := 0
	for _, memberId := range memberIds {
		if _, ok := s.uidToConn[memberId]; ok {
			onlineCount++
		}
	}

	tlog.InfofCtx(ctx, "groupChat", "群聊消息已发送给群成员: groupId=%s from=%s memberCount=%d onlineMemberCount=%d",
		groupId, req.SendId, len(members), onlineCount)

	return nil
}
