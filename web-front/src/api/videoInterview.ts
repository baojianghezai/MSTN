// 视频面试接口封装（05-video-interview.md）
import request from '@/utils/request'
import type { ApiResponse, VideoInterview, VideoInterviewCreate, VideoInterviewRoom } from '@/types/api'

/** 企业发起视频面试邀请 */
export const createVideoInterview = (data: VideoInterviewCreate) =>
  request.post<ApiResponse<VideoInterview>>('/company/video-interviews', data)

/** 企业：我发出的列表（分页） */
export const getCompanyVideoInterviews = (params: { page: number; pageSize: number }) =>
  request.get<ApiResponse<{ list: VideoInterview[]; total: number; page: number; pageSize: number }>>(
    '/company/video-interviews',
    { params }
  )

/** 企业：详情 */
export const getCompanyVideoInterview = (id: number) =>
  request.get<ApiResponse<VideoInterview>>(`/company/video-interviews/${id}`)

/** 企业：删除邀请 */
export const deleteCompanyVideoInterview = (id: number) =>
  request.delete<ApiResponse>(`/company/video-interviews/${id}`)

/** 个人：我收到的列表（分页） */
export const getPersonalVideoInterviews = (params: { page: number; pageSize: number }) =>
  request.get<ApiResponse<{ list: VideoInterview[]; total: number; page: number; pageSize: number }>>(
    '/personal/video-interviews',
    { params }
  )

/** 个人：详情（含 personalCode） */
export const getPersonalVideoInterview = (id: number) =>
  request.get<ApiResponse<VideoInterview>>(`/personal/video-interviews/${id}`)

/** 公开：房间码查询（TRTC 入房，不含联系方式） */
export const getVideoInterviewRoom = (code: string) =>
  request.get<ApiResponse<VideoInterviewRoom>>(`/video-interviews/room/${code}`)
