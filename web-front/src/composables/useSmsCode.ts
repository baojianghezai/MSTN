// 可复用的「发送短信验证码 + 倒计时」逻辑
// composable = 把响应式状态和逻辑抽成可复用函数（相当于后端的公共函数，只是它带着响应式）
import { onUnmounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { sendSmsCode } from '@/api/auth'

export function useSmsCode() {
  const sending = ref(false)
  const countdown = ref(0)
  let timer: ReturnType<typeof setInterval> | null = null

  const send = async (mobile: string, type: string): Promise<boolean> => {
    if (!/^1[3-9]\d{9}$/.test(mobile)) {
      ElMessage.warning('请输入正确的手机号')
      return false
    }
    sending.value = true
    try {
      await sendSmsCode(mobile, type)
      ElMessage.success('验证码已发送，请查看后端日志')
      countdown.value = 60
      timer = setInterval(() => {
        countdown.value--
        if (countdown.value <= 0 && timer) {
          clearInterval(timer)
          timer = null
        }
      }, 1000)
      return true
    } finally {
      sending.value = false
    }
  }

  onUnmounted(() => {
    if (timer) {
      clearInterval(timer)
    }
  })

  return { sending, countdown, send }
}
