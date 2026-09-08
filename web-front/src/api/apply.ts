// 投递接口封装（个人端，10-apply.md #61-63）
import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

/** #61 投递职位 */
export const apply = (data: { jobsIds: number[]; resumeId: number; notes?: string }) =>
  request.post<ApiResponse>('/personal/applies', data)

/** #62 我的投递列表 */
export const getApplies = (params: { page: number; pageSize: number; status?: number }) =>
  request.get<ApiResponse<{ list: ApplyItem[]; total: number; page: number; pageSize: number }>>(
    '/personal/applies',
    { params }
  )

/** #63 删除投递记录 */
export const deleteApply = (did: number) =>
  request.delete<ApiResponse>(`/personal/applies/${did}`)

// 我的投递列表项（10-apply.md #62）
export interface ApplyItem {
  did: number
  resumeId: number
  resumeName: string
  jobsId: number
  jobsName: string
  companyId: number
  companyName: string
  companyUid: number
  applyAddtime: number
  personalLook: number // 1 未读 / 2 已读
  notes: string
  isReply: number // 0 待反馈 1 合适 2 不合适 3 待定 4 未接通（企业侧回复）
  replyTime: number
}
