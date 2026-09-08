<template>
  <div class="auth-page mx-auto flex max-w-md flex-col px-4 py-16">
    <h1 class="text-center text-2xl font-bold text-slate-800">注册账号</h1>

    <el-form :model="form" label-position="top" class="mt-8" @submit.prevent>
      <el-form-item label="账号类型">
        <el-radio-group v-model="form.utype">
          <el-radio :value="1">个人</el-radio>
          <el-radio :value="2">企业</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item label="手机号">
        <el-input
          v-model="form.mobile"
          placeholder="请输入手机号"
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

      <el-form-item label="密码">
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
        @click="handleRegister"
      >
        注册并登录
      </el-button>
    </el-form>

    <div class="mt-4 text-center text-sm">
      已有账号？
      <router-link to="/login" class="text-primary-600 hover:underline">去登录</router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { reactive, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { ElMessage } from 'element-plus'
  import { useUserStore } from '@/stores/user'
  import { useSmsCode } from '@/composables/useSmsCode'
  import { register } from '@/api/auth'
  import { isStrongPassword, PASSWORD_RULE_HINT } from '@/utils/password'

  const router = useRouter()
  const userStore = useUserStore()
  const { sending, countdown, send } = useSmsCode()

  const form = reactive({
    utype: 1, // 1=个人 2=企业
    mobile: '',
    code: '',
    password: ''
  })
  const submitting = ref(false)

  const handleSendCode = () => {
    send(form.mobile, 'register')
  }

  const handleRegister = async () => {
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
      // 注册成功直接返回 token，自动登录
      const { data } = await register(form.utype, form.mobile, form.code, form.password)
      userStore.setLogin(data.token, data.uid, data.utype, data.mobile)
      ElMessage.success('注册成功')
      router.push('/')
    } finally {
      submitting.value = false
    }
  }
</script>
