// 前台统一请求封装（hrc 接口域）
// 契约：docs/api/README.md —— 响应 {code, message, data}，成功 code=0
// 鉴权：会员 JWT，Header Authorization: Bearer <token>（与后台 x-token 完全不同体系）
import axios from 'axios'
import type { AxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import router from '@/router'

const service = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || '/api/v1',
  timeout: 1000 * 60
})

// 请求拦截：自动带会员 token
service.interceptors.request.use((config) => {
  const userStore = useUserStore()
  if (userStore.token) {
    config.headers.Authorization = `Bearer ${userStore.token}`
  }
  return config
})

// 响应拦截：统一处理 {code, message, data}，401 清登录态回登录页
service.interceptors.response.use(
  (response) => {
    const { code, message } = response.data
    if (code === 0) {
      return response.data
    }
    ElMessage.error(message || '请求失败')
    return Promise.reject(new Error(message || '请求失败'))
  },
  (error) => {
    if (error.response?.status === 401) {
      const userStore = useUserStore()
      userStore.logout()
      router.push({ name: 'Login' })
      return Promise.reject(error)
    }
    ElMessage.error(error.response?.data?.message || error.message || '请求失败')
    return Promise.reject(error)
  }
)

// 类型化包装：响应拦截器会把 response.data 直接返回，
// 所以这里的 T 就是「最终拿到的数据」，返回 Promise<T>（而非 axios 默认的 AxiosResponse<T>）
// —— 类型与运行时行为保持一致，调用方 await 后拿到的就是 {code, message, data}
const http = {
  get: <T>(url: string, config?: AxiosRequestConfig) =>
    service.get<T, T>(url, config),
  post: <T>(url: string, data?: unknown, config?: AxiosRequestConfig) =>
    service.post<T, T>(url, data, config),
  put: <T>(url: string, data?: unknown, config?: AxiosRequestConfig) =>
    service.put<T, T>(url, data, config),
  delete: <T>(url: string, config?: AxiosRequestConfig) =>
    service.delete<T, T>(url, config)
}

export default http
