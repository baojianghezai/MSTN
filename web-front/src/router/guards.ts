import router from './index'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'

// 登录守卫 + utype 角色守卫：
// - 访问需要登录的路由（meta.requiresAuth）且未登录 → 跳登录页，带上回跳地址
// - 已登录访问登录页 → 直接回首页
// - 路由 meta.utype 限定角色：/personal/** 需 utype=1、/company/** 需 utype=2
router.beforeEach((to) => {
  const userStore = useUserStore()

  if (to.meta.requiresAuth && !userStore.token) {
    return {
      name: 'Login',
      query: { redirect: to.fullPath }
    }
  }

  if (to.name === 'Login' && userStore.token) {
    return { name: 'Home' }
  }

  if (to.meta.utype && userStore.utype !== to.meta.utype) {
    ElMessage.warning('无权限访问该页面')
    return { name: 'Home' }
  }

  return true
})

// 页面标题同步（路由 meta.title → 浏览器标签页）
router.afterEach((to) => {
  const base = '名硕人才网'
  document.title = to.meta.title ? `${to.meta.title} - ${base}` : base
})
