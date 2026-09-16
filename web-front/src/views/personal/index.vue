<template>
  <div class="mx-auto max-w-3xl px-4 py-8">
    <!-- 用户信息 -->
    <el-card shadow="never" class="border-t-3 border-t-primary-500">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-4">
          <el-avatar :size="52" :src="avatarUrl" class="bg-primary-50 text-primary-600">
            <span class="i-lucide-user-round" aria-hidden="true" />
          </el-avatar>
          <div>
          <h1 class="text-xl font-bold text-slate-800">个人中心</h1>
          <p class="mt-1 text-sm text-slate-500">
            uid={{ userStore.uid }}，utype={{ userStore.utype }}（1=个人 2=企业）
          </p>
          <p class="mt-1 text-sm text-slate-500">手机号：{{ userStore.mobile || '未绑定' }}</p>
          </div>
        </div>
      </div>
    </el-card>

    <!-- 快捷入口（design/10 §3.2.1 ③：九宫格快捷入口） -->
    <el-card shadow="never" class="mt-6">
      <template #header>
        <span class="font-semibold">快捷入口</span>
      </template>
      <div class="grid grid-cols-3 gap-3">
        <button
          v-for="item in quickEntries"
          :key="item.label"
          class="group flex flex-col items-center gap-1.5 rounded-lg border border-slate-200 py-4 text-sm text-slate-600 transition-colors hover:border-primary-400 hover:bg-primary-50/50"
          @click="item.action()"
        >
          <span
            :class="[
              'text-2xl transition-colors',
              item.icon,
              item.danger
                ? 'text-danger-400 group-hover:text-danger-600'
                : 'text-slate-400 group-hover:text-primary-600'
            ]"
          />
          <span :class="item.danger ? 'text-danger-600 group-hover:text-danger-700' : 'group-hover:text-primary-700'">
            {{ item.label }}
          </span>
        </button>
      </div>
    </el-card>

    <!-- 账号安全 -->
    <h2 class="mt-8 text-sm font-semibold text-slate-500">账号安全</h2>

    <div class="mt-3 grid grid-cols-1 gap-6 md:grid-cols-2">
      <!-- 修改密码 -->
      <el-card id="card-password" class="h-full">
        <template #header>
          <span class="font-semibold">修改密码</span>
        </template>
        <p class="text-xs text-slate-400">改密成功会踢出所有会话，需要重新登录</p>
        <el-form :model="pwdForm" label-width="80px" class="mt-4" @submit.prevent>
          <el-form-item label="原密码">
            <el-input v-model="pwdForm.oldPassword" type="password" show-password placeholder="请输入原密码" />
          </el-form-item>
          <el-form-item label="新密码">
            <el-input v-model="pwdForm.newPassword" type="password" show-password placeholder="至少 10 位，含三种字符类型" />
          </el-form-item>
          <el-button type="primary" :loading="changing" @click="handleChangePassword">
            修改密码
          </el-button>
        </el-form>
      </el-card>

      <!-- 换绑手机 -->
      <el-card id="card-bind" class="h-full">
        <template #header>
          <span class="font-semibold">换绑手机</span>
        </template>
        <p class="text-xs text-slate-400">更换登录手机号，需向新手机号发送短信验证</p>
        <el-form :model="bindForm" label-width="80px" class="mt-4" @submit.prevent>
          <el-form-item label="新手机号">
            <el-input v-model="bindForm.mobile" placeholder="请输入新手机号" maxlength="11" clearable />
          </el-form-item>
          <el-form-item label="验证码">
            <div class="flex w-full gap-3">
              <el-input v-model="bindForm.code" placeholder="6 位验证码" maxlength="6" clearable />
              <el-button
                class="w-36 flex-none border-primary-300 text-primary-600 hover:border-primary-400 hover:bg-primary-50 hover:text-primary-600"
                :disabled="bindSending || bindCountdown > 0"
                @click="handleSendBindCode"
              >
                {{ bindCountdown > 0 ? `${bindCountdown}s 后重发` : '获取验证码' }}
              </el-button>
            </div>
          </el-form-item>
          <el-button type="primary" :loading="binding" @click="handleBind">确认换绑</el-button>
        </el-form>
      </el-card>

      <!-- 解绑手机 -->
      <el-card id="card-unbind" class="h-full">
        <template #header>
          <span class="font-semibold">解绑手机</span>
        </template>
        <p class="text-xs text-slate-400">解绑后登录手机号将释放，需重新绑定才能用手机号登录</p>
        <el-button class="mt-4" type="danger" plain @click="handleUnbind">解绑手机</el-button>
      </el-card>

      <!-- 注销账号（危险区：danger 色 token 与 accent 橙区分） -->
      <el-card id="card-cancel" class="h-full border-danger-200">
        <template #header>
          <span class="font-semibold text-danger-600">注销账号（危险）</span>
        </template>
        <p class="text-xs text-danger-400">
          注销后进入 30 天冷静期，期间可提交申诉恢复；冷静期后数据匿名化，不可恢复
        </p>
        <el-form :model="cancelForm" label-width="80px" class="mt-4" @submit.prevent>
          <el-form-item label="手机号">
            <span class="text-slate-700">{{ userStore.mobile || '未绑定手机号' }}</span>
          </el-form-item>
          <el-form-item label="验证码">
            <div class="flex w-full gap-3">
              <el-input v-model="cancelForm.code" placeholder="6 位验证码" maxlength="6" clearable />
              <el-button
                class="w-36 flex-none border-primary-300 text-primary-600 hover:border-primary-400 hover:bg-primary-50 hover:text-primary-600"
                :disabled="cancelSending || cancelCountdown > 0"
                @click="handleSendCancelCode"
              >
                {{ cancelCountdown > 0 ? `${cancelCountdown}s 后重发` : '获取验证码' }}
              </el-button>
            </div>
          </el-form-item>
          <el-button type="danger" :loading="canceling" @click="handleCancel">注销账号</el-button>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useUserStore } from '@/stores/user'
  import { useChatStore } from '@/stores/chat'
  import { useSmsCode } from '@/composables/useSmsCode'
  import { getPersonalProfile } from '@/api/profile'
  import { bindMobile, cancelAccount, changePassword, unbindMobile } from '@/api/auth'
  import { isStrongPassword, PASSWORD_RULE_HINT } from '@/utils/password'

  const router = useRouter()
  const userStore = useUserStore()
  const chatStore = useChatStore()

  // 头像（#16）：有则显示，无则默认头像
  const avatar = ref('')
  const avatarUrl = computed(() => {
    if (!avatar.value) return '/default-avatar.svg'
    return /^https?:\/\//.test(avatar.value) ? avatar.value : `/${avatar.value}`
  })
  onMounted(async () => {
    try {
      const { data } = await getPersonalProfile()
      avatar.value = data.avatar || ''
    } catch {
      // 忽略
    }
  })

  // 退出登录（#15：退出入口收敛到个人中心，不再放顶栏）
  const handleLogout = () => {
    chatStore.disconnect()
    userStore.logout()
    router.push({ name: 'Home' })
  }

  // 滚动到页内卡片（九宫格里的安全类入口）
  const scrollTo = (id: string) => {
    document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }

  // 九宫格快捷入口：功能页走路由，安全操作滚到下方对应卡片（图标 lucide 线框，i-lucide-* 原子类）
  const quickEntries = [
    { label: '我的简历', icon: 'i-lucide-file-text', action: () => router.push({ name: 'PersonalResumes' }) },
    { label: '我的投递', icon: 'i-lucide-send', action: () => router.push({ name: 'PersonalApplies' }) },
    { label: '面试邀请', icon: 'i-lucide-calendar-check', action: () => router.push({ name: 'PersonalInterviews' }) },
    { label: '个人资料', icon: 'i-lucide-user-round', action: () => router.push({ name: 'PersonalProfile' }) },
    { label: '职位广场', icon: 'i-lucide-briefcase', action: () => router.push({ name: 'Jobs' }) },
    { label: '修改密码', icon: 'i-lucide-key-round', action: () => scrollTo('card-password') },
    { label: '换绑手机', icon: 'i-lucide-smartphone', action: () => scrollTo('card-bind') },
    { label: '解绑手机', icon: 'i-lucide-unplug', action: () => scrollTo('card-unbind') },
    { label: '注销账号', icon: 'i-lucide-shield-alert', danger: true, action: () => scrollTo('card-cancel') },
    { label: '退出登录', icon: 'i-lucide-log-out', danger: true, action: handleLogout }
  ]

  // 换绑手机 & 注销各用一份独立的验证码倒计时
  const { sending: bindSending, countdown: bindCountdown, send: sendBindCode } = useSmsCode()
  const { sending: cancelSending, countdown: cancelCountdown, send: sendCancelCode } = useSmsCode()

  const pwdForm = reactive({ oldPassword: '', newPassword: '' })
  const bindForm = reactive({ mobile: '', code: '' })
  const cancelForm = reactive({ code: '' })

  const changing = ref(false)
  const binding = ref(false)
  const canceling = ref(false)

  const handleChangePassword = async () => {
    if (!pwdForm.oldPassword || !isStrongPassword(pwdForm.newPassword)) {
      ElMessage.warning(`请填写原密码；新密码${PASSWORD_RULE_HINT}`)
      return
    }
    changing.value = true
    try {
      await changePassword(pwdForm.oldPassword, pwdForm.newPassword)
      ElMessage.success('密码已修改，请重新登录')
      userStore.logout()
      router.push({ name: 'Login' })
    } finally {
      changing.value = false
    }
  }

  const handleSendBindCode = () => {
    sendBindCode(bindForm.mobile, 'bind')
  }

  const handleBind = async () => {
    if (!bindForm.mobile || !bindForm.code) {
      ElMessage.warning('请填写新手机号和验证码')
      return
    }
    binding.value = true
    try {
      await bindMobile(bindForm.mobile, bindForm.code)
      ElMessage.success('手机号已更新')
      bindForm.mobile = ''
      bindForm.code = ''
    } finally {
      binding.value = false
    }
  }

  const handleUnbind = async () => {
    try {
      await ElMessageBox.confirm('解绑后需重新绑定才能用手机号登录，确认解绑？', '提示', {
        confirmButtonText: '确认解绑',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return // 用户取消
    }
    try {
      await unbindMobile()
      ElMessage.success('已解绑手机')
    } catch {
      // 错误已由拦截器统一弹出
    }
  }

  const handleSendCancelCode = () => {
    sendCancelCode(userStore.mobile, 'bind')
  }

  const handleCancel = async () => {
    if (!userStore.mobile || !cancelForm.code) {
      ElMessage.warning('请填写验证码')
      return
    }
    try {
      await ElMessageBox.confirm(
        '注销后账号进入 30 天冷静期，数据将在冷静期后匿名化。确认注销？',
        '严重警告',
        {
          confirmButtonText: '确认注销',
          cancelButtonText: '再想想',
          type: 'error'
        }
      )
    } catch {
      return // 用户取消
    }
    canceling.value = true
    try {
      await cancelAccount(cancelForm.code)
      ElMessage.success('账号已注销，30 天内可提交申诉恢复')
      userStore.logout()
      router.push({ name: 'Home' })
    } finally {
      canceling.value = false
    }
  }
</script>
