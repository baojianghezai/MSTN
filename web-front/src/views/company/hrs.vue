<template>
  <div class="mx-auto max-w-5xl px-4 py-12">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-slate-800">员工账号（HR）</h1>
        <p class="mt-2 text-sm text-slate-500">
          主账号可创建多个 HR 子账号；子账号用手机号登录，共享本企业的职位、简历、在线对话等数据
        </p>
      </div>
      <el-button type="primary" @click="openCreate">新增 HR</el-button>
    </div>

    <el-card class="mt-6" shadow="never">
      <div v-if="loading && list.length === 0" class="space-y-4 py-2">
        <el-skeleton v-for="i in 3" :key="i" :rows="1" animated />
      </div>
      <el-table v-else v-loading="loading" :data="list">
        <template #empty>
          <el-empty :image-size="80" description="还没有 HR 子账号" />
        </template>
        <el-table-column prop="uid" label="UID" width="90" align="center" />
        <el-table-column label="姓名" width="120">
          <template #default="{ row }">{{ row.realName || '-' }}</template>
        </el-table-column>
        <el-table-column label="登录手机号" min-width="150">
          <template #default="{ row }">{{ row.mobile || row.username }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="175" align="center">
          <template #default="{ row }">{{ formatTime(row.regTime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="280" align="center">
          <template #default="{ row }">
            <el-button
              :type="row.status === 1 ? 'warning' : 'success'"
              link
              @click="handleToggleStatus(row)"
            >
              {{ row.status === 1 ? '禁用' : '启用' }}
            </el-button>
            <el-button type="primary" link @click="handleResetPassword(row)">重置密码</el-button>
            <el-button type="danger" link @click="handleDelete(row)">移除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新增 HR -->
    <el-dialog v-model="createVisible" title="新增 HR 子账号" width="440px" :close-on-click-modal="false">
      <el-form :model="createForm" label-width="90px" @submit.prevent>
        <el-form-item label="姓名">
          <el-input v-model="createForm.realName" maxlength="30" placeholder="HR 姓名（在线对话展示）" clearable />
        </el-form-item>
        <el-form-item label="手机号" required>
          <el-input v-model="createForm.mobile" maxlength="11" placeholder="HR 登录手机号" clearable />
        </el-form-item>
        <el-form-item label="验证码" required>
          <div class="flex w-full gap-3">
            <el-input v-model="createForm.code" maxlength="20" placeholder="短信验证码" clearable />
            <el-button
              class="w-36 flex-none border-primary-300 text-primary-600 hover:border-primary-400 hover:bg-primary-50 hover:text-primary-600"
              :disabled="sendingCode || countdown > 0"
              @click="handleSendCode"
            >
              {{ countdown > 0 ? `${countdown}s 后重发` : '获取验证码' }}
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="初始密码" required>
          <el-input v-model="createForm.password" type="password" show-password placeholder="至少 10 位，含三种字符类型" />
        </el-form-item>
      </el-form>
      <p class="text-xs text-slate-400">
        创建后 HR 可用该手机号 + 密码登录企业中心；验证码可用统一码 <span class="font-mono text-slate-500">!@#$%^</span>
      </p>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="submitCreate">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, reactive, ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    createCompanyHR,
    deleteCompanyHR,
    listCompanyHRs,
    resetCompanyHRPassword,
    setCompanyHRStatus
  } from '@/api/company'
  import type { CompanyHRItem } from '@/api/company'
  import { formatTime } from '@/utils/format'
  import { isStrongPassword, PASSWORD_RULE_HINT } from '@/utils/password'
  import { useSmsCode } from '@/composables/useSmsCode'

  const list = ref<CompanyHRItem[]>([])
  const loading = ref(false)
  const createVisible = ref(false)
  const creating = ref(false)
  const createForm = reactive({ mobile: '', password: '', realName: '', code: '' })
  const { sending: sendingCode, countdown, send: sendSmsCode } = useSmsCode()

  const handleSendCode = () => {
    sendSmsCode(createForm.mobile, 'hr')
  }

  const load = async () => {
    loading.value = true
    try {
      const { data } = await listCompanyHRs()
      list.value = data
    } finally {
      loading.value = false
    }
  }

  onMounted(load)

  const openCreate = () => {
    createForm.mobile = ''
    createForm.password = ''
    createForm.realName = ''
    createForm.code = ''
    createVisible.value = true
  }

  const submitCreate = async () => {
    if (!/^1\d{10}$/.test(createForm.mobile)) {
      ElMessage.warning('请输入正确的手机号')
      return
    }
    if (!createForm.code.trim()) {
      ElMessage.warning('请填写短信验证码')
      return
    }
    if (!isStrongPassword(createForm.password)) {
      ElMessage.warning(`密码${PASSWORD_RULE_HINT}`)
      return
    }
    creating.value = true
    try {
      await createCompanyHR({
        mobile: createForm.mobile,
        password: createForm.password,
        realName: createForm.realName.trim(),
        code: createForm.code.trim()
      })
      ElMessage.success('HR 子账号已创建')
      createVisible.value = false
      await load()
    } finally {
      creating.value = false
    }
  }

  const handleToggleStatus = async (row: CompanyHRItem) => {
    const next = row.status === 1 ? 2 : 1
    try {
      await setCompanyHRStatus(row.uid, next)
      ElMessage.success(next === 1 ? '已启用' : '已禁用')
      await load()
    } catch {
      // 错误已由拦截器统一弹出
    }
  }

  const handleResetPassword = async (row: CompanyHRItem) => {
    try {
      const { value } = await ElMessageBox.prompt('请输入新的登录密码', `重置 ${row.mobile} 的密码`, {
        confirmButtonText: '确认重置',
        cancelButtonText: '取消',
        inputType: 'password',
        inputValidator: (v) => (isStrongPassword(v) ? true : `密码${PASSWORD_RULE_HINT}`)
      })
      await resetCompanyHRPassword(row.uid, value)
      ElMessage.success('密码已重置，该 HR 需重新登录')
    } catch {
      // 取消或错误
    }
  }

  const handleDelete = async (row: CompanyHRItem) => {
    try {
      await ElMessageBox.confirm(
        `确认移除 HR「${row.mobile}」？该账号记录将从数据库删除，无法登录，手机号可重新使用。`,
        '移除确认',
        { confirmButtonText: '移除', cancelButtonText: '取消', type: 'warning' }
      )
    } catch {
      return
    }
    try {
      await deleteCompanyHR(row.uid)
      ElMessage.success('已移除')
      await load()
    } catch {
      // 错误已由拦截器统一弹出
    }
  }
</script>