import requestHrc from '@/utils/requestHrc'

export const getSetmeals = () => requestHrc({ url: '/admin/setmeals', method: 'get' })
export const createSetmeal = (data) => requestHrc({ url: '/admin/setmeals', method: 'post', data })
export const updateSetmeal = (id, data) => requestHrc({ url: `/admin/setmeals/${id}`, method: 'put', data })
export const getOrders = (params) => requestHrc({ url: '/admin/orders', method: 'get', params })
export const confirmOrder = (id, data) => requestHrc({ url: `/admin/orders/${id}/confirm`, method: 'post', data })
