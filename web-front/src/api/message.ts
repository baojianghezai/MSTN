import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

export type MessageScope = 'personal' | 'company'

export interface MessageItem {
  id: number
  msgFrom: number
  msgTouid: number
  title: string
  message: string
  type: string
  link: string
  msgCheck: number
  addtime: number
}

interface MessagePage {
  list: MessageItem[]
  total: number
  page: number
  pageSize: number
}

interface MessageUnread {
  unread: number
}

/** #69 / #72a 个人或企业的站内信列表。 */
export const getMessages = (scope: MessageScope, params: { page: number; pageSize: number }) =>
  request.get<ApiResponse<MessagePage>>(`/${scope}/messages`, { params })

/** #70 / #72b 批量标记当前账号所属消息为已读。 */
export const markMessagesRead = (scope: MessageScope, ids: number[]) =>
  request.put<ApiResponse>(`/${scope}/messages/read`, { ids })

/** #71 / #72c 批量删除当前账号所属消息。 */
export const deleteMessages = (scope: MessageScope, ids: number[]) =>
  request.delete<ApiResponse>(`/${scope}/messages`, { data: { ids } })

/** #72 / #72d 获取当前账号的站内信未读数。 */
export const getMessageUnreadCount = (scope: MessageScope) =>
  request.get<ApiResponse<MessageUnread>>(`/${scope}/messages/unread-count`)
