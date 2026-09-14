import request from '@/utils/request'
import type { PublicJobItem } from '@/api/jobs'
import type { ApiResponse } from '@/types/api'

export interface CompanyPromotionItem {
  id: number
  uid: number
  jobId: number
  type: 1 | 2
  adTitle: string
  adSubtitle: string
  adImage: string
  sort: number
  createdAt: string
  jobsName: string
  companyname: string
}

export interface HomeAdItem extends PublicJobItem {
  promotionId: number
  adTitle: string
  adSubtitle: string
  adImage: string
}

export interface CompanyPromotionData {
  list: CompanyPromotionItem[]
  homePushSlots: number
  homeAdSlots: number
  pushUsed: number
  adUsed: number
}

/** 公开首页的有效推流和广告位。 */
export const getHomePromotions = () =>
  request.get<ApiResponse<{ push: PublicJobItem[]; ads: HomeAdItem[] }>>('/home/promotions')

/** 企业当前投放和套餐名额。 */
export const getCompanyPromotions = () =>
  request.get<ApiResponse<CompanyPromotionData>>('/company/promotions')

/** 创建首页投放：type 1=推流，2=广告位。 */
export const createPromotion = (
  jobId: number,
  type: 1 | 2,
  creative: { adTitle?: string; adSubtitle?: string; adImage?: string } = {}
) =>
  request.post<ApiResponse<CompanyPromotionItem>>('/company/promotions', { jobId, type, ...creative })

/** 撤下首页投放，名额立即释放。 */
export const deletePromotion = (id: number) =>
  request.delete<ApiResponse>(`/company/promotions/${id}`)
