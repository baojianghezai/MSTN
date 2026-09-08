<template>
  <div class="mx-auto max-w-2xl px-4 py-16">
    <h1 class="text-2xl font-bold text-slate-800">账号申诉</h1>
    <p class="mt-2 text-sm text-slate-500">
      收不到验证码、误注销想恢复、换绑失败等账号问题，可在此提交申诉，客服处理后短信通知结果
    </p>

    <el-tabs v-model="tab" class="mt-6">
      <!-- 提交申诉 -->
      <el-tab-pane label="提交申诉" name="submit">
        <el-form :model="form" label-position="top" @submit.prevent>
          <el-form-item label="姓名">
            <el-input v-model="form.realname" placeholder="请输入姓名" clearable />
          </el-form-item>
          <el-form-item label="手机号">
            <el-input v-model="form.mobile" placeholder="请输入手机号" maxlength="11" clearable />
          </el-form-item>
          <el-form-item label="邮箱（选填）">
            <el-input v-model="form.email" placeholder="选填，便于客服联系" clearable />
          </el-form-item>
          <el-form-item label="申诉内容">
            <el-input
              v-model="form.description"
              type="textarea"
              :rows="4"
              placeholder="请描述遇到的问题，例如：账号被误注销，申请恢复"
            />
          </el-form-item>
          <el-button type="primary" :loading="submitting" @click="handleSubmit">
            提交申诉
          </el-button>
        </el-form>
      </el-tab-pane>

      <!-- 查询进度 -->
      <el-tab-pane label="查询进度" name="status">
        <div class="flex gap-3">
          <el-input
            v-model="statusMobile"
            placeholder="输入手机号查询申诉进度"
            maxlength="11"
            clearable
            @keyup.enter="handleQuery"
          />
          <el-button type="primary" :loading="querying" @click="handleQuery">查询</el-button>
        </div>

        <el-table v-if="statusList.length" :data="statusList" class="mt-4" border>
          <el-table-column prop="id" label="ID" width="70" align="center" />
          <el-table-column prop="realname" label="姓名" width="110" />
          <el-table-column prop="mobile" label="手机号" width="130" />
          <el-table-column label="状态" width="100" align="center">
            <template #default="{ row }">
              <el-tag :type="statusTag(row.status)">{{ row.statusCn }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="提交时间">
            <template #default="{ row }">{{ formatTime(row.addtime) }}</template>
          </el-table-column>
        </el-table>
        <el-empty v-else-if="queried" description="未查询到申诉记录" />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
  import { reactive, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { useUserStore } from '@/stores/user'
  import { getAppealStatus, submitAppeal } from '@/api/appeal'
  import { formatTime } from '@/utils/format'
  import type { AppealStatusItem } from '@/types/api'

  const userStore = useUserStore()

  const tab = ref('submit')
  const submitting = ref(false)
  const querying = ref(false)
  const queried = ref(false)

  // 已登录时预填当前手机号，减少用户输入
  const form = reactive({
    realname: '',
    mobile: userStore.mobile,
    email: '',
    description: ''
  })
  const statusMobile = ref(userStore.mobile)
  const statusList = ref<AppealStatusItem[]>([])

  const handleSubmit = async () => {
    if (!form.realname || !form.mobile || !form.description) {
      ElMessage.warning('请填写姓名、手机号和申诉内容')
      return
    }
    if (!/^1[3-9]\d{9}$/.test(form.mobile)) {
      ElMessage.warning('请输入正确的手机号')
      return
    }
    submitting.value = true
    try {
      await submitAppeal(form.realname, form.mobile, form.email, form.description)
      ElMessage.success('申诉已提交，客服会尽快处理')
      form.realname = ''
      form.email = ''
      form.description = ''
    } finally {
      submitting.value = false
    }
  }

  const handleQuery = async () => {
    if (!/^1[3-9]\d{9}$/.test(statusMobile.value)) {
      ElMessage.warning('请输入正确的手机号')
      return
    }
    querying.value = true
    try {
      const { data } = await getAppealStatus(statusMobile.value)
      statusList.value = data
      queried.value = true
    } finally {
      querying.value = false
    }
  }

  const statusTag = (status: number) => {
    const map: Record<number, string> = { 0: 'warning', 1: 'success', 2: 'danger' }
    return map[status] || 'info'
  }
</script>
