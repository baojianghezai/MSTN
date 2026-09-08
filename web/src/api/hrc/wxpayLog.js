// 微信支付回调日志接口封装（后台，08-wxpay-log.md）
import requestHrc from '@/utils/requestHrc'

/**
 * 支付日志列表
 * @param {Object} params { page, pageSize, status }
 *   status 缺省=全部；0=失败 1=成功
 */
export const getWxpayLogs = (params) =>
  requestHrc({
    url: '/admin/wxpay-logs',
    method: 'get',
    params
  })
