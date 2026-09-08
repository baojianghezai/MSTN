import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

export interface Setmeal {
  id: number
  name: string
  price: number
  durationDays: number
  jobsMeanwhile: number
  resumeDownloads: number
  homePushSlots: number
  homeAdSlots: number
  enableVideo: boolean
  display: boolean
  description: string
}

export interface CurrentSetmeal {
  setmealId: number
  setmealName: string
  expireAt: number
  jobsMeanwhile: number
  resumeDownloadsTotal: number
  resumeDownloadsUsed: number
  homePushSlots: number
  homeAdSlots: number
  enableVideo: boolean
}

export interface BillingOrder {
  id: number
  oid: string
  setmealName: string
  amount: number
  payAmount: number
  isPaid: number
  createdAt: number
  paidAt: number
  payment: string
  transactionId: string | null
}

export interface PaymentStart {
  provider: 'wechat_native' | 'alipay_page'
  qrCodeUrl?: string
  redirectUrl?: string
  expiresAt: number
}

export const getSetmeals = () => request.get<ApiResponse<Setmeal[]>>('/setmeals')

export const getCurrentSetmeal = () =>
  request.get<ApiResponse<CurrentSetmeal | null>>('/company/members/setmeal')

export const createSetmealOrder = (setmealId: number) =>
  request.post<ApiResponse<BillingOrder>>('/orders', { setmealId })

export const getOrders = (params: { page: number; pageSize: number }) =>
  request.get<ApiResponse<{ list: BillingOrder[]; total: number }>>('/orders', { params })

export const cancelOrder = (id: number) => request.post<ApiResponse>(`/orders/${id}/cancel`)

export const startPayment = (id: number, provider: PaymentStart['provider']) =>
  request.post<ApiResponse<PaymentStart>>(`/orders/${id}/pay`, { provider })
