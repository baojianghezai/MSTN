<template>
  <div class="auth-page mx-auto flex max-w-md flex-col px-4 py-16">
    <h1 class="text-center text-2xl font-bold text-slate-800">重置密码</h1>
    <p class="mt-2 text-center text-sm text-slate-500">
      输入注册手机号，短信验证后设置新密码
    </p>

    <el-form :model="form" label-position="top" class="mt-8" @submit.prevent>
      <el-form-item label="手机号">
        <el-input
          v-model="form.mobile"
          placeholder="请输入注册手机号"
          maxlength="11"
          clearable
        />
      </el-form-item>

      <el-form-item label="验证码">
        <div class="flex w-full gap-3">
          <el-input
            v-model="form.code"
            placeholder="6 位验证码"
            maxlength="6"
            clearable
          />
          <el-button
            class="w-36 flex-none border-primary-300 text-primary-600 hover:border-primary-400 hover:bg-primary-50 hover:text-primary-600"
            :disabled="sending || countdown > 0"
            @click="handleSendCode"
          >
            {{ countdown > 0 ? `${countdown}s 后重发` : '获取验证码' }}
          </el-button>
        </div>
      </el-form-item>

      <el-form-item label="新密码">
        <el-input
          v-model="form.password"
          type="password"
          show-password
          placeholder="至少 10 位，含三种字符类型"
        />
      </el-form-item>

      <el-button
        class="mt-2 w-full"
        type="primary"
        size="large"
        :loading="submitting"
        @click="handleReset"
      >
        重置密码
      </el-button>
    </el-form>

    <div class="mt-4 text-center text-sm">
      想起密码了？
      <router-link to="/login" class="text-primary-600 hover:underline">去登录</router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { reactive, ref } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { ElMessage } from 'element-plus'
  import { useSmsCode } from '@/composables/useSmsCode'
  import { resetPassword } from '@/api/auth'
  import { isStrongPassword, PASSWORD_RULE_HINT } from '@/utils/password'

  const router = useRouter()
  const route = useRoute()
  const { sending, countdown, send } = useSmsCode()

  const form = reactive({
    // 支持从「未设置密码」引导跳转时预填手机号
    mobile: (route.query.mobile as string) || '',
    code: '',
    password: ''
  })
  const submitting = ref(false)

  const handleSendCode = () => {
    send(form.mobile, 'reset')
  }

  const handleReset = async () => {
    if (!form.mobile || !form.code) {
      ElMessage.warning('请填写手机号和验证码')
      return
    }
    if (!isStrongPassword(form.password)) {
      ElMessage.warning(PASSWORD_RULE_HINT)
      return
    }
    submitting.value = true
    try {
      // 重置成功不自动登录，引导用户重新登录
      await resetPassword(form.mobile, form.code, form.password)
      ElMessage.success('密码已重置，请重新登录')
      router.push({ name: 'Login' })
    } finally {
      submitting.value = false
    }
  }
</script>
