import requestHrc from '@/utils/requestHrc'

/** 预览固定范围的数据清理，不修改数据。 */
export const previewDataCleanup = (data) =>
  requestHrc({
    url: '/admin/data-cleanup/preview',
    method: 'post',
    data
  })

/** 执行已确认的数据清理。 */
export const executeDataCleanup = (data) =>
  requestHrc({
    url: '/admin/data-cleanup/execute',
    method: 'post',
    data
  })

/** 获取数据清理执行审计记录。 */
export const getDataCleanupHistory = (params) =>
  requestHrc({
    url: '/admin/data-cleanup/history',
    method: 'get',
    params
  })
