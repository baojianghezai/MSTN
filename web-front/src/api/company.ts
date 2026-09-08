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
