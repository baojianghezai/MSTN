// 个人资料接口封装（02-account.md #76/#77）
import request from '@/utils/request'
import type { ApiResponse, PersonalProfile } from '@/types/api'

/** 个人资料读取 */
export const getPersonalProfile = () =>
  request.get<ApiResponse<PersonalProfile>>('/personal/profile')

/** 个人资料维护（整包覆盖写） */
export const updatePersonalProfile = (data: PersonalProfile) =>
  request.post<ApiResponse>('/personal/profile', data)
