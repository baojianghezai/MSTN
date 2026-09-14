// 企业收简历接口封装（企业端，11-company-apply.md #90/#92-94）
import axios from 'axios'
import { ElMessage } from 'element-plus'
import request from '@/utils/request'
import { useUserStore } from '@/stores/user'
import type { ApiResponse } from '@/types/api'

/** #92 收简历列表（脱敏快照，分页 + 状态筛选） */
export const getCompanyApplies = (params: { page: number; pageSize: number; status?: number }) =>
  request.get<ApiResponse<{ list: CompanyApplyItem[]; total: number; page: number; pageSize: number }>>(
    '/company/applies',
    { params }
  )

/** #90 单职位收到的投递 */
export const getJobApplies = (jobId: number, params: { page: number; pageSize: number }) =>
  request.get<ApiResponse<{ list: CompanyApplyItem[]; total: number; page: number; pageSize: number }>>(
    `/company/jobs/${jobId}/applies`,
    { params }
  )

/** #93 标记已看 */
export const markLooked = (did: number) =>
  request.put<ApiResponse>(`/company/applies/${did}/looked`)

/** #94 回复状态 { isReply: 0-4 } */
export const replyApply = (did: number, isReply: number) =>
  request.put<ApiResponse>(`/company/applies/${did}/reply`, { isReply })

/** #95 下载完整简历。首次下载扣套餐额度，重复下载不扣。 */
export const downloadResume = async (did: number) => {
  const { token } = useUserStore()
  const apiBase = import.meta.env.VITE_API_BASE || '/api/v1'
  const response = await axios.get(`${apiBase}/company/applies/${did}/resume/download`, {
    headers: { Authorization: `Bearer ${token}` },
    responseType: 'blob'
  })

  const contentType = response.headers['content-type'] || ''
  if (contentType.includes('application/json')) {
    const error = JSON.parse(await response.data.text()) as ApiResponse
    ElMessage.error(error.message || '下载失败')
    throw new Error(error.message || '下载失败')
  }

  const filename = response.headers['content-disposition']?.match(/filename="?([^";]+)"?/)?.[1] || `resume-${did}.html`
  const url = URL.createObjectURL(response.data)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}

// 收简历列表项（11-company-apply.md #92）
export interface CompanyApplyItem {
  did: number
  resumeId: number
  resumeName: string
  jobsId: number
  jobsName: string
  companyId: number
  companyName: string
  applyAddtime: string
  personalLook: number // 1 未读 / 2 已读
  notes: string
  isReply: number // 0 待反馈 1 合适 2 不合适 3 待定 4 未接通
  replyTime: string
}
