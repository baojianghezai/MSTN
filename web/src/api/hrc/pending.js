// 后台待办统计 + 推广投放审核（#3/#8）
import requestHrc from '@/utils/requestHrc'

/** 后台待处理数量 */
export const getPendingCounts = () =>
  requestHrc({
    url: '/admin/pending-counts',
    method: 'get'
  })

/** 推广投放列表（audit：-1 全部 / 0 待审 / 1 通过 / 3 不通过） */
export const getPromotions = (params) =>
  requestHrc({
    url: '/admin/promotions',
    method: 'get',
    params
  })

/** 审核推广投放（audit 1=通过 3=不通过，3 时 reason 必填） */
export const auditPromotion = (id, body) =>
  requestHrc({
    url: `/admin/promotions/${id}/audit`,
    method: 'put',
    data: body
  })