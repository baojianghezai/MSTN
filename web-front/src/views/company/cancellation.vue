<template>
  <div class="mx-auto max-w-2xl px-4 py-16">
    <h1 class="text-2xl font-bold text-slate-800">企业账号注销</h1>
    <p class="mt-2 text-sm text-slate-500">
      提交注销申请后由后台审批；审批通过将清除企业业务数据，会员账号保留可复用
    </p>

    <!-- 待处理申请：阻塞重复提交，展示状态 -->
    <el-card v-if="pendingCancel" class="mt-6">
      <div class="flex items-center justify-between">
        <span class="font-semibold text-slate-700">注销申请状态</span>
        <el-tag :type="statusTag(pendingCancel.status)">{{ pendingCancel.statusCn }}</el-tag>
      </div>
      <div class="mt-4 space-y-2 text-sm text-slate-600">
        <p>企业名称：{{ pendingCancel.companyname }}</p>
        <p>提交时间：{{ formatTime(pendingCancel.addtime) }}</p>
        <p v-if="pendingCancel.finishtime">处理时间：{{ formatTime(pendingCancel.finishtime) }}</p>
      </div>
      <p class="mt-4 text-xs text-slate-400">
        申请正在处理中，请耐心等待；处理完成后可再次提交申请
      </p>
    </el-card>

    <!-- 可提交：无申请 或 上次已处理（允许再次申请） -->
    <el-card v-else class="mt-6">
      <div v-if="lastCancelled" class="mb-4 rounded bg-green-50 p-3 text-xs text-green-600">
        上次注销申请已于 {{ formatTime(lastCancelled.finishtime) }} 处理完成，企业数据已清除，可再次提交申请
      </div>
      <p class="text-sm text-slate-500">
        需先用当前绑定手机号接收验证码（{{ userStore.mobile || '未绑定手机号' }}）
      </p>
      <el-form label-position="top" class="mt-4" @submit.prevent>
        <el-form-item label="验证码">
          <div class="flex w-full gap-3">
            <el-input v-model="code" placeholder="6 位验证码" maxlength="6" clearable />
            <el-button
              class="w-36 flex-none"
              :disabled="sending || countdown > 0"
              @click="handleSendCode"
            >
              {{ countdown > 0 ? `${countdown}s 后重发` : '获取验证码' }}
            </el-button>
          </div>
        </el-form-item>
        <el-button type="danger" :loading="submitting" @click="handleSubmit">
          提交注销申请
        </el-button>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { useUserStore } from '@/stores/user'
  import { useSmsCode } from '@/composables/useSmsCode'
  import { companyCancel, getCompanyCancelStatus } from '@/api/company'
  import { formatTime } from '@/utils/format'
  import type { CompanyCancellation } from '@/types/api'

  const userStore = useUserStore()
  const { sending, countdown, send } = useSmsCode()

  // 待处理申请（status=0）：阻塞重复提交
  const pendingCancel = ref<CompanyCancellation | null>(null)
  // 上次已处理申请（status=1）：仅保留历史提示，不阻塞再次申请
  const lastCancelled = ref<CompanyCancellation | null>(null)
  const code = ref('')
  const submitting = ref(false)

  // 只有 status=0（待处理）才阻塞；已处理（status=1）允许再次申请；
  // 无申请时后端返回错误「注销申请不存在」，用 try/catch 兜底
  const loadStatus = async () => {
    try {
      const { data } = await getCompanyCancelStatus()
      if (data && data.id && data.status === 0) {
        pendingCancel.value = data
        lastCancelled.value = null
      } else if (data && data.id) {
        pendingCancel.value = null
        lastCancelled.value = data
      } else {
        pendingCancel.value = null
        lastCancelled.value = null
      }
    } catch {
      pendingCancel.value = null
      lastCancelled.value = null
    }
  }

  onMounted(loadStatus)

  const handleSendCode = () => {
    send(userStore.mobile, 'cancellation')
  }

  const handleSubmit = async () => {
    if (!code.value) {
      ElMessage.warning('请输入验证码')
      return
    }
    submitting.value = true
    try {
      await companyCancel(code.value)
      ElMessage.success('注销申请已提交，等待后台审批')
      code.value = ''
      await loadStatus()
    } finally {
      submitting.value = false
    }
  }

  const statusTag = (s: number) => (s === 0 ? 'warning' : 'success')
</script>
