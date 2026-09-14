// 职位管理接口封装（企业端 #79-86 + 前台公开 #22-25，09-jobs.md）
import request from '@/utils/request'
import type { ApiResponse, Categories, JobDetail, JobItem, JobsRequest } from '@/types/api'

/** #79 发布职位 */
export const createJob = (data: JobsRequest) =>
  request.post<ApiResponse<{ id: number }>>('/company/jobs', data)

/** #80 我的职位列表（jobs + jobs_tmp 合并，分页） */
export const getJobs = (params: { page: number; pageSize: number }) =>
  request.get<ApiResponse<{ list: JobItem[]; total: number; page: number; pageSize: number }>>(
    '/company/jobs',
    { params }
  )

/** #81 职位详情（编辑回显，含 reason/contact/tags）
 * @param pending true=按 jobs_tmp 表查（待审项）；缺省按 jobs 表查
 *   两表 id 各自自增会重叠，必须由 pending 区分（2026-08-18 缺陷修复）
 */
export const getJob = (id: number, pending?: boolean) =>
  request.get<ApiResponse<JobDetail>>(`/company/jobs/${id}`, {
    params: pending ? { pending: 1 } : undefined
  })

/** #82 编辑职位
 * @param pending true=按 jobs_tmp 表原地重提（待审项）；缺省按 jobs 表编辑
 *   两表 id 各自自增会重叠，显式透传可避免歧义（D1，2026-08-21 后端建议）
 */
export const updateJob = (id: number, data: JobsRequest, pending?: boolean) =>
  request.put<ApiResponse>(`/company/jobs/${id}`, data, {
    params: pending ? { pending: 1 } : undefined
  })

/** #83 删除职位（逻辑删除）
 * @param pending true=按 jobs_tmp 表软删（待审项）；缺省按 jobs 表删
 *   两表 id 各自自增会重叠，显式透传可避免歧义（D1，2026-08-21 后端建议）
 */
export const deleteJob = (id: number, pending?: boolean) =>
  request.delete<ApiResponse>(`/company/jobs/${id}`, {
    params: pending ? { pending: 1 } : undefined
  })

/** #84 暂停职位 */
export const pauseJob = (id: number) =>
  request.put<ApiResponse>(`/company/jobs/${id}/pause`)

/** #85 恢复职位 */
export const resumeJob = (id: number) =>
  request.put<ApiResponse>(`/company/jobs/${id}/resume`)

/** #86 刷新职位 */
export const refreshJob = (id: number) =>
  request.put<ApiResponse>(`/company/jobs/${id}/refresh`)

// ========== 前台公开（#22-25）==========

// 公开职位列表查询参数（#22）
export interface JobSearchParams {
  keyword?: string
  trade?: number
  category?: number
  district?: string
  education?: number
  experience?: number
  minwage?: number
  maxwage?: number
  order?: string // last 默认 / addtime / salary / stick
  page?: number
  pageSize?: number
}

// 公开职位列表项（#22，含公司名/分类中文/薪资等展示字段）
export interface PublicJobItem {
  id: number
  jobsName: string
  companyname: string
  logo: string
  natureCn: string
  categoryCn: string
  districtCn: string
  minwage: number
  maxwage: number
  negotiable: number
  education: number
  experience: number
  addtime: string
  refreshtime: string
  stick: number
  emergency: number
}

/** #22 前台职位列表（筛选+排序+分页） */
export const searchJobs = (params: JobSearchParams) =>
  request.get<ApiResponse<{ list: PublicJobItem[]; total: number; page: number; pageSize: number }>>(
    '/jobs',
    { params }
  )

/** #23 前台职位详情（click 自增） */
export const getPublicJob = (id: number) =>
  request.get<ApiResponse<JobDetail>>(`/jobs/${id}`)

/** #24 热门搜索词 */
export const getHotWords = () =>
  request.get<ApiResponse<string[]>>('/jobs/hot-words')

/** #25 筛选元数据（分类聚合，含 jobcategory 三级层级） */
export const getJobFilters = () =>
  request.get<ApiResponse<Categories>>('/jobs/filters')

