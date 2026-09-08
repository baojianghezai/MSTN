// 申诉模块接口封装（后台业务页）
// 契约：docs/api/02-account.md #158/#159
import requestHrc from '@/utils/requestHrc'

/**
 * 申诉列表
 * @param {Object} params 查询参数
 * @param {number} [params.page] 页码，默认 1
 * @param {number} [params.pageSize] 每页条数，默认 10，最大 100
 * @param {number} [params.status] 状态：0 全部 / 1 已处理 / 2 已驳回
 * @param {string} [params.mobile] 手机号模糊筛选
 * @returns {Promise<{code:number, message:string, data:{list:Array, total:number, page:number, pageSize:number}}>}
 */
export const getAppealList = (params) =>
  requestHrc({
    url: '/admin/appeals',
    method: 'get',
    params
  })

/**
 * 处理申诉（可恢复账号）
 * @param {number} id 申诉 ID
 * @param {Object} body 处理参数
 * @param {number} body.status 1 已处理 / 2 已驳回
 * @param {boolean} [body.restore] 仅 status=1 时生效：是否按申诉手机号恢复注销账号
 * @returns {Promise<{code:number, message:string, data:Object}>}
 */
export const processAppeal = (id, body) =>
  requestHrc({
    url: `/admin/appeals/${id}`,
    method: 'put',
    data: body
  })
