<template>
  <div class="mx-auto max-w-3xl px-4 py-8">
    <!-- 今日数据卡（design/10 §3.2.1 ② 企业中心：工作台数据卡） -->
    <div v-loading="statsLoading" class="grid grid-cols-1 gap-4 md:grid-cols-3">
      <el-card shadow="never" class="border-t-3 border-t-slate-400 text-center">
        <span class="i-lucide-briefcase-business text-xl text-slate-400" aria-hidden="true" />
        <p class="mt-2 text-3xl font-bold text-slate-800">{{ stats.jobs }}</p>
        <p class="mt-1 text-sm text-slate-500">在招职位</p>
      </el-card>
      <el-card shadow="never" class="border-t-3 border-t-primary-500 text-center">
        <span class="i-lucide-inbox text-xl text-primary-500" aria-hidden="true" />
        <p class="mt-2 text-3xl font-bold text-primary-600">{{ stats.applies }}</p>
        <p class="mt-1 text-sm text-slate-500">收到简历</p>
      </el-card>
      <el-card shadow="never" class="border-t-3 border-t-accent-500 text-center">
        <span class="i-lucide-mail-open text-xl text-accent-500" aria-hidden="true" />
        <p class="mt-2 text-3xl font-bold text-accent-500">{{ stats.unread }}</p>
        <p class="mt-1 text-sm text-slate-500">未读简历</p>
      </el-card>
    </div>

    <el-card shadow="never" class="mt-6">
      <h1 class="text-xl font-bold text-slate-800">企业中心</h1>
      <p class="mt-1 text-sm text-slate-500">
        uid={{ userStore.uid }}，手机号：{{ userStore.mobile }}
      </p>

      <div class="mt-4 flex items-center gap-3">
        <span class="text-sm text-slate-500">资质审核状态：</span>
        <el-tag :type="auditTag(audit.audit)">{{ audit.auditCn || '未提交' }}</el-tag>
      </div>

      <div class="mt-6 flex flex-wrap items-center gap-3">
        <el-button type="primary" @click="router.push({ name: 'CompanyJobs' })">
          <span class="mr-1 i-lucide-briefcase" />职位管理
        </el-button>
        <el-button type="warning" plain @click="router.push({ name: 'CompanyApplies' })">
          <span class="mr-1 i-lucide-inbox" />收到的简历
          <el-badge v-if="stats.unread > 0" :value="stats.unread" class="ml-1" />
        </el-button>
        <el-button @click="router.push({ name: 'CompanyProfile' })">
          <span class="mr-1 i-lucide-building-2" />维护企业资料
        </el-button>
        <el-button @click="router.push({ name: 'CompanyInterviews' })">
          <span class="mr-1 i-lucide-calendar-plus" />面试邀请
        </el-button>
        <el-button type="success" plain @click="router.push({ name: 'CompanyPlan' })">
          <span class="mr-1 i-lucide-credit-card" />套餐与订单
        </el-button>
        <el-button type="danger" plain @click="router.push({ name: 'CompanyCancellation' })">
          <span class="mr-1 i-lucide-shield-alert" />企业账号注销
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, reactive, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { useUserStore } from '@/stores/user'
  import { getCompanyAuditStatus } from '@/api/company'
  import type { CompanyAudit } from '@/types/api'
  import { getJobs } from '@/api/jobs'
  import { getCompanyApplies } from '@/api/companyApply'

  const router = useRouter()
  const userStore = useUserStore()

  const audit = ref<CompanyAudit>({ audit: 0, auditCn: '' })

  // 工作台数据卡：pageSize=1 只取 total，不拉全量列表
  const statsLoading = ref(false)
  const stats = reactive({ jobs: 0, applies: 0, unread: 0 })

  onMounted(async () => {
    try {
      const { data } = await getCompanyAuditStatus()
      audit.value = data
    } catch {
      // 审核状态拉取失败不阻塞页面
    }

    statsLoading.value = true
    try {
      const [jobs, applies, unread] = await Promise.all([
        getJobs({ page: 1, pageSize: 1 }),
        getCompanyApplies({ page: 1, pageSize: 1 }),
        getCompanyApplies({ page: 1, pageSize: 1, status: 1 })
      ])
      stats.jobs = jobs.data.total
      stats.applies = applies.data.total
      stats.unread = unread.data.total
    } catch {
      // 数据卡拉取失败不阻塞页面（显示 0）
    } finally {
      statsLoading.value = false
    }
  })

  const auditTag = (a: number) => {
    const map: Record<number, string> = { 0: 'info', 1: 'success', 2: 'warning', 3: 'danger' }
    return map[a] || 'info'
  }
</script>
