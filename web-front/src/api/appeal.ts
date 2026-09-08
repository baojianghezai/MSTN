// 申诉接口封装（02-account.md #45/#46）
import request from '@/utils/request'
import type { ApiResponse, AppealStatusItem } from '@/types/api'

/** 提交账号申诉（公开，已登录自动带 uid） */
export const submitAppeal = (
  realname: string,
  mobile: string,
  email: string,
  description: string
) =>
  request.post<ApiResponse<{ id: number }>>('/appeal', {
    realname,
    mobile,
    email,
    description
  })

/** 申诉进度查询（按手机号返回最近 10 条倒序） */
export const getAppealStatus = (mobile: string) =>
  request.get<ApiResponse<AppealStatusItem[]>>('/appeal/status', {
    params: { mobile }
  })
