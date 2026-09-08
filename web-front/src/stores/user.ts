import { defineStore } from 'pinia'
import { ref } from 'vue'

// 会员登录态。
// Token 存储 key 用 ms_token / ms_utype / ms_uid —— 与后台 GVA 的 token 隔离
// （design/10 §1.3：同一浏览器同时开两个前端也不串号）
const TOKEN_KEY = 'ms_token'
const UTYPE_KEY = 'ms_utype'
const UID_KEY = 'ms_uid'
const MOBILE_KEY = 'ms_mobile'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem(TOKEN_KEY) || '')
  const uid = ref(Number(localStorage.getItem(UID_KEY) || 0))
  const utype = ref(Number(localStorage.getItem(UTYPE_KEY) || 0))
  const mobile = ref(localStorage.getItem(MOBILE_KEY) || '')

  function setLogin(tokenStr: string, uidNum: number, utypeNum: number, mobileStr: string) {
    token.value = tokenStr
    uid.value = uidNum
    utype.value = utypeNum
    mobile.value = mobileStr
    localStorage.setItem(TOKEN_KEY, tokenStr)
    localStorage.setItem(UID_KEY, String(uidNum))
    localStorage.setItem(UTYPE_KEY, String(utypeNum))
    localStorage.setItem(MOBILE_KEY, mobileStr)
  }

  function logout() {
    token.value = ''
    uid.value = 0
    utype.value = 0
    mobile.value = ''
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(UID_KEY)
    localStorage.removeItem(UTYPE_KEY)
    localStorage.removeItem(MOBILE_KEY)
  }

  return { token, uid, utype, mobile, setLogin, logout }
})
