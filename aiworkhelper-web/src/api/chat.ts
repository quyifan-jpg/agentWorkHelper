import request from "@/utils/request";
import type {
  ApiResponse,
  ChatRequest,
  ChatResponse,
  ChatMessageListResponse,
} from "@/types";

// AI聊天
export function chat(data: ChatRequest): Promise<ApiResponse<ChatResponse>> {
  return request({
    url: "/v1/chat",
    method: "post",
    data,
  });
}

// 创建群聊
export function createGroup(data: {
  groupId: string;
  groupName: string;
  memberIds: string[];
}): Promise<ApiResponse<void>> {
  return request({
    url: "/v1/group/create",
    method: "post",
    data,
  });
}

// 获取群成员列表
export function getGroupMembers(
  groupId: string
): Promise<ApiResponse<string[]>> {
  return request({
    url: `/v1/group/${groupId}/members`,
    method: "get",
  });
}

// 添加群成员
export function addGroupMembers(data: {
  groupId: string;
  memberIds: string[];
}): Promise<ApiResponse<void>> {
  return request({
    url: "/v1/group/members/add",
    method: "post",
    data,
  });
}

// 移除群成员
export function removeGroupMember(
  groupId: string,
  userId: string
): Promise<ApiResponse<void>> {
  return request({
    url: `/v1/group/${groupId}/members/${userId}`,
    method: "delete",
  });
}

// 查询历史消息列表
export function getChatMessages(params: {
  conversationId: string;
  page?: number;
  count?: number;
  startTime?: number;
  endTime?: number;
}): Promise<ApiResponse<ChatMessageListResponse>> {
  return request({
    url: "/v1/chat/messages",
    method: "get",
    params,
  });
}
