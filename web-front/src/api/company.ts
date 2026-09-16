// 企业资料接口封装（02-account.md #107-#110）
import request from '@/utils/request'
import type { ApiResponse, CompanyAudit, CompanyCancellation, CompanyProfile } from '@/types/api'

/** 企业资料读取 */
export const getCompanyProfile = () =>
  request.get<ApiResponse<CompanyProfile>>('/company/profile')

/** 企业资料维护（提交即触发审核 audit=2） */
export const updateCompanyProfile = (data: CompanyProfile) =>
  request.post<ApiResponse>('/company/profile', data)

/** 企业审核状态 */
export const getCompanyAuditStatus = () =>
  request.get<ApiResponse<CompanyAudit>>('/company/profile/audit-status')

/** 提交企业注销申请（code 必须 type=cancellation 发送） */
export const companyCancel = (code: string) =>
  request.post<ApiResponse>('/company/cancel', { code })

/** 查询企业注销申请状态（最近一条） */
export const getCompanyCancelStatus = () =>
  request.get<ApiResponse<CompanyCancellation>>('/company/cancel')

// ---- 企业 HR 子账号（#21，仅企业主账号可管理）----

export interface CompanyHRItem {
  uid: number
  username: string
  mobile: string
  status: number // 1=启用 2=禁用
  regTime: string
}

/** HR 子账号列表 */
export const listCompanyHRs = () =>
  request.get<ApiResponse<CompanyHRItem[]>>('/company/hrs')

/** 新增 HR 子账号（手机号 + 密码登录，数据共享企业主体） */
export const createCompanyHR = (data: { mobile: string; password: string }) =>
  request.post<ApiResponse<CompanyHRItem>>('/company/hrs', data)

/** 启用/禁用 HR 子账号 */
export const setCompanyHRStatus = (uid: number, status: number) =>
  request.put<ApiResponse>(`/company/hrs/${uid}/status`, { status })

/** 重置 HR 子账号密码 */
export const resetCompanyHRPassword = (uid: number, password: string) =>
  request.put<ApiResponse>(`/company/hrs/${uid}/password`, { password })

/** 移除 HR 子账号 */
export const deleteCompanyHR = (uid: number) =>
  request.delete<ApiResponse>(`/company/hrs/${uid}`)
