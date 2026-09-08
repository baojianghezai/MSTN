// 统计看板与报表接口封装（后台，06-dashboard.md #119/#120 + statistics）
import requestHrc from '@/utils/requestHrc'

/**
 * #119 看板指标（今日/昨日/待办/收入）
 * @returns {Promise<{data:{today:Object, yesterday:Object, todo:Object, income:Object}}>}
 */
export const getDashboard = () =>
  requestHrc({
    url: '/admin/dashboard',
    method: 'get'
  })

/**
 * #120 趋势图数据
 * @param {Object} params { days, metric } metric: register/resume/company/job/application
 * @returns {Promise<{data:Array<{date, personal, company, count}>}>}
 */
export const getTrend = (params) =>
  requestHrc({
    url: '/admin/dashboard/trend',
    method: 'get',
    params
  })

/**
 * 求职者分布（性别/学历/经验）
 * @returns {Promise<{data:{sex:Array, education:Array, experience:Array}}>}
 */
export const getResumeStatistics = () =>
  requestHrc({
    url: '/admin/statistics/resume',
    method: 'get'
  })

/**
 * 企业分布（性质/规模）
 * @returns {Promise<{data:{nature:Array, scale:Array}}>}
 */
export const getCompanyStatistics = () =>
  requestHrc({
    url: '/admin/statistics/company',
    method: 'get'
  })
