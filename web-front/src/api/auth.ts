// 认证接口封装（01-auth.md）
import request from '@/utils/request'
import type { ApiResponse, LoginResult } from '@/types/api'

/** 发送短信验证码（验证码打印在服务端日志，一期 mock） */
export const sendSmsCode = (mobile: string, type: string) =>
  request.post<ApiResponse>('/auth/sms-code', { mobile, type })

/** 短信验证码登录（未注册自动注册） */
export const loginBySms = (mobile: string, code: string, utype = 0) =>
  request.post<ApiResponse<LoginResult>>('/auth/login/sms', { mobile, code, utype })

/** 密码登录（账号自动识别：手机号/邮箱/用户名） */
export const loginByPassword = (account: string, password: string) =>
  request.post<ApiResponse<LoginResult>>('/auth/login', { account, password })

/** 注册（个人/企业） */
export const register = (utype: number, mobile: string, code: string, password: string) =>
  request.post<ApiResponse<LoginResult>>('/auth/register', { utype, mobile, code, password })

/** 忘记密码重置（code 必须 type=reset 发送） */
export const resetPassword = (mobile: string, code: string, newPassword: string) =>
  request.post<ApiResponse>('/auth/password/reset', { mobile, code, new_password: newPassword })

/** 修改密码（原密码校验；成功后踢出所有会话需重新登录） */
export const changePassword = (oldPassword: string, newPassword: string) =>
  request.put<ApiResponse>('/auth/password', { old_password: oldPassword, new_password: newPassword })

/** 绑定/换绑手机（code 必须 type=bind 发送） */
export const bindMobile = (mobile: string, code: string) =>
  request.put<ApiResponse>('/auth/bind/mobile', { mobile, code })

/** 解绑手机（无参） */
export const unbindMobile = () => request.put<ApiResponse>('/auth/unbind/mobile')

/** 账号注销（code 必须 type=bind 发送；两阶段匿名化，冷静期可申诉恢复） */
export const cancelAccount = (code: string) =>
  request.post<ApiResponse>('/auth/cancel', { code })
