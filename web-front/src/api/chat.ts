// 在线对话接口封装（个人/企业双端共用；REST 落库 + WebSocket 实时推送）
import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

export type ChatScope = 'personal' | 'company'

// 会话列表项（按会话方向返回对端信息）
export interface ChatSessionItem {
  id: number
  peerUid: number
  peerName: string
  peerLogo: string
  jobsId: number
  jobsName: string
  lastContent: string
  lastTime: string | null
  unread: number
  updateTime: string
}

// 消息项（mine 便于前端左右气泡）
export interface ChatMessageItem {
  id: number
  sessionId: number
  fromUid: number
  toUid: number
  mine: boolean
  content: string
  isRead: number
  addtime: string
}

interface ChatPage<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

/** 会话列表（按最后消息时间倒序） */
export const listChatSessions = (scope: ChatScope, params: { page: number; pageSize: number }) =>
  request.get<ApiResponse<ChatPage<ChatSessionItem>>>(`/${scope}/chat/sessions`, { params })

/** 打开/创建会话（幂等；peerUid=对方 uid，jobsId/jobsName 为上下文） */
export const openChatSession = (
  scope: ChatScope,
  data: { peerUid: number; jobsId?: number; jobsName?: string }
) => request.post<ApiResponse<{ id: number }>>(`/${scope}/chat/sessions`, data)

/** 会话消息（分页，返回按时间正序） */
export const getChatMessages = (
  scope: ChatScope,
  sessionId: number,
  params: { page: number; pageSize: number }
) => request.get<ApiResponse<ChatPage<ChatMessageItem>>>(`/${scope}/chat/sessions/${sessionId}/messages`, { params })

/** 发送消息（落库并实时推送对端） */
export const sendChatMessage = (scope: ChatScope, sessionId: number, content: string) =>
  request.post<ApiResponse<ChatMessageItem>>(`/${scope}/chat/sessions/${sessionId}/messages`, { content })

/** 标记会话已读（清零己方未读） */
export const markChatRead = (scope: ChatScope, sessionId: number) =>
  request.put<ApiResponse>(`/${scope}/chat/sessions/${sessionId}/read`)

/** 删除会话（仅己方隐藏） */
export const deleteChatSession = (scope: ChatScope, sessionId: number) =>
  request.delete<ApiResponse>(`/${scope}/chat/sessions/${sessionId}`)

/** 未读总数（顶栏红点） */
export const getChatUnread = (scope: ChatScope) =>
  request.get<ApiResponse<{ unread: number }>>(`/${scope}/chat/unread`)