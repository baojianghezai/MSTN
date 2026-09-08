// hrc 业务域接口请求封装（后台业务页专用）
//
// 与 GVA 原生 request.js 双轨并存（见 design/10_前端设计方案.md §1.3）：
//   - GVA 原生接口：baseURL /api，响应 {code, data, msg}（错误 code=7）→ 走 request.js
//   - hrc 业务接口：baseURL /api/v1，响应 {code, message, data}（成功 code=0）→ 走本文件
//
// 鉴权说明：后台 hrc 接口（/api/v1/admin/**）复用 GVA 后台 JWT，Header 仍是 x-token，
// 所以这里和 request.js 一样从 userStore 取 token。
import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/pinia/modules/user'
import router from '@/router/index'

const service = axios.create({
  baseURL: '/api/v1',
  timeout: 1000 * 60
})

// 请求拦截：统一携带后台 JWT
service.interceptors.request.use(
  (config) => {
    const userStore = useUserStore()
    config.headers = {
      'Content-Type': 'application/json',
      'x-token': userStore.token,
      ...config.headers
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截：统一处理 hrc 结构 {code, message, data}
service.interceptors.response.use(
  (response) => {
    const { code, message } = response.data

    // 成功：直接返回 response.data，调用方 .then(({ data }) => ...)
    if (code === 0) {
      return response.data
    }

    // HTTP 200 + 业务错误码（hrc 约定，见 docs/api/README.md）
    ElMessage.error(message || '请求失败')
    return Promise.reject(new Error(message || '请求失败'))
  },
  (error) => {
    // 未登录 / token 失效：清理登录态并回登录页（与 request.js 行为一致）
    if (error.response?.status === 401) {
      const userStore = useUserStore()
      userStore.ClearStorage()
      router.push({ name: 'Login', replace: true })
      return Promise.reject(error)
    }

    ElMessage.error(
      error.response?.data?.message || error.message || '请求失败'
    )
    return Promise.reject(error)
  }
)

export default service
