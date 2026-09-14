import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

export interface CompanyInterview {
  did: number
  resumeId: number
  resumeName: string
  resumeUid: number
  jobsId: number
  jobsName: string
  companyId: number
  companyName: string
  companyUid: number
  interviewTime: string
  address: string
  contact: string
  telephone: string
  notes: string
  interviewAddtime: string
  personalLook: number
}

export interface InterviewCreate {
  resumeId: number
  jobsId: number
  interviewTime: number
  address: string
  contact: string
  telephone: string
  notes: string
}

type InterviewPage = { list: CompanyInterview[]; total: number; page: number; pageSize: number }

/** #98 企业发起标准线下面试邀请。 */
export const createInterview = (data: InterviewCreate) =>
  request.post<ApiResponse<CompanyInterview>>('/company/interviews', data)

/** #99 企业查看已发出的邀请。 */
export const getCompanyInterviews = (params: { page: number; pageSize: number }) =>
  request.get<ApiResponse<InterviewPage>>('/company/interviews', { params })

/** #100 企业撤回邀请。 */
export const withdrawInterview = (did: number) =>
  request.delete<ApiResponse>(`/company/interviews/${did}`)

/** #64 个人查看收到的邀请。 */
export const getPersonalInterviews = (params: { page: number; pageSize: number }) =>
  request.get<ApiResponse<InterviewPage>>('/personal/interviews', { params })

/** #65 个人将邀请标记为已读。 */
export const markInterviewRead = (did: number) =>
  request.put<ApiResponse>(`/personal/interviews/${did}/read`)
