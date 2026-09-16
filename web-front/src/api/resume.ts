// 简历接口封装（04-resume.md，个人端 utype=1）
// M3 收尾批（handoff/2026-08-20-resume-m3.md）：#48/#49/#50/#51/#52/#54/#55/#57
// 后置（勿按编号猜写）：#53 刷新 / #56 复制 / #58 附件 / #59 照片 / #60 发邮箱
import request from '@/utils/request'
import type {
  ApiResponse,
  Resume,
  ResumeCompleteness,
  ResumeLite
} from '@/types/api'

/** #48 创建简历（含子表全量写入，project 限 6 条；首份自动 def=1） */
export const createResume = (data: Resume) =>
  request.post<ApiResponse<{ id: number }>>('/personal/resumes', data)

/** #49 我的简历列表（轻量字段不返子表；默认简历在前，创建时间倒序） */
export const listResumes = (params: { page: number; pageSize: number }) =>
  request.get<ApiResponse<{ list: ResumeLite[]; total: number; page: number; pageSize: number }>>(
    '/personal/resumes',
    { params }
  )

/** #50 简历详情（编辑回显，含子表数组；⚠️ 教育经历键名 educations 复数） */
export const getResume = (id: number) =>
  request.get<ApiResponse<Resume>>(`/personal/resumes/${id}`)

/** #51 编辑简历（子表全量替换，project 仍限 6 条） */
export const updateResume = (id: number, data: Resume) =>
  request.put<ApiResponse>(`/personal/resumes/${id}`, data)

/** #58 删除附件简历（清空 word_resume 系列字段；上传走 el-upload 直传 /personal/resumes/{id}/outward） */
export const removeResumeOutward = (id: number) =>
  request.delete<ApiResponse>(`/personal/resumes/${id}/outward`)

/** #52 删除简历（软删；删默认简历自动升级最近一份为默认） */
export const deleteResume = (id: number) =>
  request.delete<ApiResponse>(`/personal/resumes/${id}`)

/** #54 公开/隐藏（1=公开 2=不公开） */
export const setResumeDisplay = (id: number, display: number) =>
  request.put<ApiResponse>(`/personal/resumes/${id}/display`, { display })

/** #55 设为默认（同 uid 内 def 互斥事务） */
export const setResumeDefault = (id: number) =>
  request.put<ApiResponse>(`/personal/resumes/${id}/default`)

/** #57 完善度详情（percent + 分组得分 + 缺失字段清单，与主表 completePercent 同源） */
export const getResumeCompleteness = (id: number) =>
  request.get<ApiResponse<ResumeCompleteness>>(`/personal/resumes/${id}/completeness`)
