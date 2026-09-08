// 企业注销申请接口封装（后台，02-account.md #161a-c）
import requestHrc from '@/utils/requestHrc'
import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/pinia/modules/user'

/**
 * 企业注销申请列表
 * @param {Object} params { page, pageSize, status } status: 0 全部 / 1 待处理 / 2 已处理
 */
export const getCompanyCancellations = (params) =>
  requestHrc({
    url: '/admin/company/cancellations',
    method: 'get',
    params
  })

/**
 * 处理企业注销申请（清除企业业务数据，账号保留）
 * @param {number} id 申请 ID
 */
export const handleCompanyCancellation = (id) =>
  requestHrc({
    url: `/admin/company/cancellations/${id}/handle`,
    method: 'post'
  })

/**
 * 硬删除企业注销申请记录
 * @param {number} id 申请 ID
 */
export const deleteCompanyCancellation = (id) =>
  requestHrc({
    url: `/admin/company/cancellations/${id}`,
    method: 'delete'
  })

/**
 * 企业资料列表（#134）
 * @param {Object} params { page, pageSize, audit, keyword }
 *   audit 缺省=全部；0 草稿 / 1 已通过 / 2 审核中 / 3 未通过（注意 0 是「草稿」不是「全部」）
 */
export const getCompanyProfiles = (params) =>
  requestHrc({
    url: '/admin/company-profiles',
    method: 'get',
    params
  })

/**
 * 企业资料详情（#134a，含 logo/certificateImg 证照）
 * @param {number} id 企业资料 ID
 */
export const getCompanyProfile = (id) =>
  requestHrc({
    url: `/admin/company-profiles/${id}`,
    method: 'get'
  })

/**
 * 企业资质审核（#135）
 * @param {number} id 企业资料 ID
 * @param {Object} body { audit: 1|3, reason } audit=3 时 reason 必填
 */
export const auditCompanyProfile = (id, body) =>
  requestHrc({
    url: `/admin/company-profiles/${id}/audit`,
    method: 'put',
    data: body
  })

/**
 * 企业导出（07-export.md，CSV 文件流下载）
 * 注意：成功是文件流（text/csv），失败才是 JSON 1001，所以不能走 requestHrc 的 JSON 拦截器，
 * 用原生 axios + responseType: 'blob'，手动带 x-token。
 * @param {number[]} ids 企业资料 id 列表（来自列表勾选）
 */
export const exportCompanies = async (ids) => {
  const userStore = useUserStore()
  const res = await axios({
    url: '/api/v1/admin/companies/export',
    method: 'post',
    data: { ids },
    responseType: 'blob',
    headers: {
      'Content-Type': 'application/json',
      'x-token': userStore.token
    }
  })

  // 失败时后端返回 JSON（blob 类型），成功是 CSV
  const contentType = res.headers['content-type'] || ''
  if (contentType.includes('application/json')) {
    const text = await res.data.text()
    let msg = '导出失败'
    try {
      msg = JSON.parse(text).message || msg
    } catch {
      // 解析失败保留默认文案
    }
    ElMessage.error(msg)
    throw new Error(msg)
  }

  // 从 Content-Disposition 取文件名，拿不到就按日期兜底
  const disposition = res.headers['content-disposition'] || ''
  const match = disposition.match(/filename="?([^";]+)"?/)
  const filename = match
    ? decodeURIComponent(match[1])
    : `company-export-${new Date().toISOString().slice(0, 10).replace(/-/g, '')}.csv`

  const url = URL.createObjectURL(res.data)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}
