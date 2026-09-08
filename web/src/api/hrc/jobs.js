// 职位管理接口封装（后台，09-jobs.md #121-123）
import requestHrc from '@/utils/requestHrc'
import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/pinia/modules/user'

/**
 * #121 职位管理列表
 * @param {Object} params { page, pageSize, audit, keyword }
 *   audit 缺省=全部 0-3；keyword 职位名/企业名模糊搜索
 */
export const getAdminJobs = (params) =>
  requestHrc({
    url: '/admin/jobs',
    method: 'get',
    params
  })

/**
 * #122 待审职位列表（jobs_tmp，audit=2）
 * @param {Object} params { page, pageSize }
 */
export const getAdminJobsTmp = (params) =>
  requestHrc({
    url: '/admin/jobs/tmp',
    method: 'get',
    params
  })

/**
 * #123 职位审核
 * @param {number} id 职位 id（tmp 场景为 tmp id）
 * @param {Object} body { audit: 1|3, reason } audit=3 时 reason 必填
 */
/**
 * 职位审核详情（含职位全量字段 + 联系方式 + 标签 + 关联企业资质）
 * @param {number} id 职位 id（tmp 或 jobs）
 */
export const getAdminJobDetail = (id) =>
  requestHrc({
    url: `/admin/jobs/${id}`,
    method: 'get'
  })


export const auditAdminJob = (id, body) =>
  requestHrc({
    url: `/admin/jobs/${id}/audit`,
    method: 'put',
    data: body
  })

/**
 * 职位导出（07-export.md，CSV 文件流下载）
 * 成功是文件流，失败才是 JSON 1001，不能走 requestHrc。
 * @param {number[]} ids ms_jobs.id 列表
 */
export const exportAdminJobs = async (ids) => {
  const userStore = useUserStore()
  const res = await axios({
    url: '/api/v1/admin/jobs/export',
    method: 'post',
    data: { ids },
    responseType: 'blob',
    headers: {
      'Content-Type': 'application/json',
      'x-token': userStore.token
    }
  })

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

  const disposition = res.headers['content-disposition'] || ''
  const match = disposition.match(/filename="?([^";]+)"?/)
  const filename = match
    ? decodeURIComponent(match[1])
    : `jobs-export-${new Date().toISOString().slice(0, 10).replace(/-/g, '')}.csv`

  const url = URL.createObjectURL(res.data)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}