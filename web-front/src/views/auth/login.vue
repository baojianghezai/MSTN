<template>
  <div class="auth-page mx-auto flex max-w-md flex-col px-4 py-16">
    <h1 class="text-center text-2xl font-bold text-slate-800">登录名硕人才网</h1>

    <el-tabs v-model="tab" class="mt-8">
      <!-- 短信登录 -->
      <el-tab-pane label="短信登录" name="sms">
        <el-form :model="smsForm" label-position="top" @submit.prevent>
          <el-form-item label="账号类型">
            <el-radio-group v-model="smsForm.utype">
              <el-radio :value="1">个人</el-radio>
              <el-radio :value="2">企业</el-radio>
            </el-radio-group>
            <div class="w-full text-xs text-slate-400">
              首次登录按所选身份注册；已注册手机号以库中身份为准
            </div>
          </el-form-item>
          <el-form-item label="手机号">
            <el-input
              v-model="smsForm.mobile"
              placeholder="请输入手机号"
              maxlength="11"
              clearable
            />
          </el-form-item>
          <el-form-item label="验证码">
            <div class="flex w-full gap-3">
              <el-input
                v-model="smsForm.code"
                placeholder="6 位验证码"
                maxlength="6"
                clearable
              />
              <el-button
                class="w-36 flex-none border-primary-300 text-primary-600 hover:border-primary-400 hover:bg-primary-50 hover:text-primary-600"
                :disabled="sending || countdown > 0"
                @click="handleSendSms"
              >
                {{ countdown > 0 ? `${countdown}s 后重发` : '获取验证码' }}
              </el-button>
            </div>
          </el-form-item>
          <el-button
            class="mt-2 w-full"
            type="primary"
            size="large"
            :loading="submitting"
            @click="handleSmsLogin"
          >
            登录
          </el-button>
        </el-form>
      </el-tab-pane>

      <!-- 密码登录 -->
      <el-tab-pane label="密码登录" name="password">
        <el-form :model="pwdForm" label-position="top" @submit.prevent>
          <el-form-item label="账号">
            <el-input
              v-model="pwdForm.account"
              placeholder="手机号 / 邮箱 / 用户名"
              clearable
            />
          </el-form-item>
          <el-form-item label="密码">
            <el-input
              v-model="pwdForm.password"
              type="password"
              show-password
              placeholder="请输入密码"
              @keyup.enter="handlePasswordLogin"
            />
          </el-form-item>
          <el-button
            class="mt-2 w-full"
            type="primary"
            size="large"
            :loading="submitting"
            @click="handlePasswordLogin"
          >
            登录
          </el-button>
        </el-form>
      </el-tab-pane>
    </el-tabs>

    <div class="mt-4 flex items-center justify-between text-sm">
      <router-link to="/register" class="text-primary-600 hover:underline">注册账号</router-link>
      <router-link to="/forgot-password" class="text-slate-500 hover:underline">忘记密码</router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { reactive, ref } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { ElMessage } from 'element-plus'
  import { useUserStore } from '@/stores/user'
  import { useSmsCode } from '@/composables/useSmsCode'
  import { loginByPassword, loginBySms } from '@/api/auth'
  import type { LoginResult } from '@/types/api'

  const router = useRouter()
  const route = useRoute()
  const userStore = useUserStore()
  const { sending, countdown, send } = useSmsCode()

  const tab = ref('sms')
  const submitting = ref(false)

  const smsForm = reactive({ mobile: '', code: '', utype: 1 })
  const pwdForm = reactive({ account: '', password: '' })

  // 登录成功后统一处理：存登录态 → 回跳或按身份进对应中心
  const afterLogin = (data: LoginResult) => {
    userStore.setLogin(data.token, data.uid, data.utype, data.mobile)

    // 无密码账号（如注销恢复后的账号）：引导先设置密码
    if (!data.passwordSet) {
      ElMessage.warning('该账号尚未设置密码，请先设置密码')
      router.push({ name: 'ForgotPassword', query: { mobile: data.mobile } })
      return
    }

    ElMessage.success('登录成功')
    const redirect = (route.query.redirect as string) || ''
    if (redirect) {
      router.push(redirect)
    } else {
      // 无回跳地址时按最终身份进对应中心（后端以库为准，utype 是最终身份）
      router.push(data.utype === 2 ? { name: 'Company' } : { name: 'Home' })
    }
  }

  const handleSendSms = () => {
    send(smsForm.mobile, 'login')
  }

  const handleSmsLogin = async () => {
    if (!smsForm.mobile || !smsForm.code) {
      ElMessage.warning('请填写手机号和验证码')
      return
    }
    submitting.value = true
    try {
      const { data } = await loginBySms(smsForm.mobile, smsForm.code, smsForm.utype)
      afterLogin(data)
    } finally {
      submitting.value = false
    }
  }

  const handlePasswordLogin = async () => {
    if (!pwdForm.account || !pwdForm.password) {
      ElMessage.warning('请填写账号和密码')
      return
    }
    submitting.value = true
    try {
      const { data } = await loginByPassword(pwdForm.account, pwdForm.password)
      afterLogin(data)
    } finally {
      submitting.value = false
    }
  }
</script>
