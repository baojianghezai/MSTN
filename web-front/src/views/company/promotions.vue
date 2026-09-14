<template>
  <div class="mx-auto max-w-6xl px-4 py-8">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold text-slate-800">首页推广</h1>
        <p class="mt-1 text-sm text-slate-500">将已上线职位投放到首页推流或广告位，投放有效期与当前套餐一致。</p>
      </div>
      <el-button :loading="loading" @click="load">刷新</el-button>
    </div>

    <el-alert
      class="mt-5"
      type="info"
      :closable="false"
      title="推广位来自已购买套餐"
      description="推流职位在首页优先展示；广告位在首页顶部展示。撤下后会立即释放对应名额。"
    />

    <div class="mt-5 grid gap-4 sm:grid-cols-2">
      <el-card shadow="never">
        <p class="text-sm text-slate-500">首页推流位</p>
        <p class="mt-2 text-2xl font-semibold text-slate-800">{{ data.pushUsed }} / {{ data.homePushSlots }}</p>
      </el-card>
      <el-card shadow="never">
        <p class="text-sm text-slate-500">首页广告位</p>
        <p class="mt-2 text-2xl font-semibold text-slate-800">{{ data.adUsed }} / {{ data.homeAdSlots }}</p>
      </el-card>
    </div>

    <el-card class="mt-5" shadow="never">
      <template #header>投放职位</template>
      <el-form class="flex flex-wrap items-end gap-3" @submit.prevent>
        <el-form-item label="已上线职位" class="mb-0 min-w-60 flex-1">
          <el-select v-model="form.jobId" filterable placeholder="选择要投放的职位" class="w-full">
            <el-option v-for="job in eligibleJobs" :key="job.id" :label="job.jobsName" :value="job.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="推广类型" class="mb-0">
          <el-radio-group v-model="form.type">
            <el-radio-button :value="1">首页推流</el-radio-button>
            <el-radio-button :value="2">广告位</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-button type="primary" :loading="submitting" @click="submit">立即投放</el-button>
      </el-form>
      <div v-if="form.type === 2" class="mt-5 grid gap-4 border-t border-slate-200 pt-5 md:grid-cols-2">
        <el-form-item label="广告主标题" class="mb-0">
          <el-input v-model="form.adTitle" maxlength="60" show-word-limit placeholder="例如：加入我们，成就更好职业未来" />
        </el-form-item>
        <el-form-item label="广告副标题" class="mb-0">
          <el-input v-model="form.adSubtitle" maxlength="120" show-word-limit placeholder="补充招聘亮点或福利信息" />
        </el-form-item>
        <el-form-item label="广告横幅" class="mb-0 md:col-span-2">
          <div>
            <el-upload
              action="/api/v1/upload"
              :headers="uploadHeaders"
              :show-file-list="false"
              accept="image/*"
              :before-upload="beforeUpload"
              :on-success="onAdImageUploaded"
            >
              <el-button>上传横幅图片</el-button>
            </el-upload>
            <p class="mt-2 text-xs text-slate-500">建议尺寸 1200 × 280，图片不超过 5MB。</p>
            <img v-if="form.adImage" :src="imageUrl(form.adImage)" class="mt-3 h-28 w-full max-w-xl border border-slate-200 object-cover" />
          </div>
        </el-form-item>
      </div>
      <p v-if="eligibleJobs.length === 0" class="mt-3 text-sm text-slate-500">暂无可投放职位，请先发布并通过审核。</p>
    </el-card>

    <el-card class="mt-5" shadow="never">
      <template #header>当前投放</template>
      <el-table v-loading="loading" :data="data.list">
        <template #empty><el-empty :image-size="70" description="暂未投放首页推广" /></template>
        <el-table-column prop="jobsName" label="职位" min-width="180" />
        <el-table-column label="类型" width="130">
          <template #default="{ row }"><el-tag :type="row.type === 1 ? 'warning' : 'success'">{{ row.type === 1 ? '首页推流' : '广告位' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="广告创意" min-width="220">
          <template #default="{ row }">
            <div v-if="row.type === 2" class="flex items-center gap-2">
              <img v-if="row.adImage" :src="imageUrl(row.adImage)" class="h-10 w-20 border border-slate-200 object-cover" />
              <span class="truncate">{{ row.adTitle || row.jobsName }}</span>
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="投放时间" min-width="180"><template #default="{ row }">{{ formatTime(row.createdAt) }}</template></el-table-column>
        <el-table-column label="操作" width="110" fixed="right"><template #default="{ row }"><el-button type="danger" link @click="remove(row)">撤下</el-button></template></el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useUserStore } from '@/stores/user'
  import { getJobs } from '@/api/jobs'
  import type { JobItem } from '@/types/api'
  import { formatTime } from '@/utils/format'
  import {
    createPromotion,
    deletePromotion,
    getCompanyPromotions,
    type CompanyPromotionData,
    type CompanyPromotionItem
  } from '@/api/promotion'

  const userStore = useUserStore()
  const loading = ref(false)
  const submitting = ref(false)
  const jobs = ref<JobItem[]>([])
  const data = reactive<CompanyPromotionData>({ list: [], homePushSlots: 0, homeAdSlots: 0, pushUsed: 0, adUsed: 0 })
  const form = reactive({ jobId: 0, type: 1 as 1 | 2, adTitle: '', adSubtitle: '', adImage: '' })

  const uploadHeaders = computed(() => ({ Authorization: `Bearer ${userStore.token}` }))

  const eligibleJobs = computed(() => jobs.value.filter((job) => !job.pending && job.display === 1 && job.audit === 1))

  const load = async () => {
    loading.value = true
    try {
      const [promotionRes, jobRes] = await Promise.all([
        getCompanyPromotions(),
        getJobs({ page: 1, pageSize: 100 })
      ])
      Object.assign(data, promotionRes.data)
      jobs.value = jobRes.data.list
    } finally {
      loading.value = false
    }
  }

  const submit = async () => {
    if (!form.jobId) {
      ElMessage.warning('请选择已上线职位')
      return
    }
    submitting.value = true
    try {
      if (form.type === 2 && !form.adImage) {
        ElMessage.warning('请先上传广告横幅图片')
        return
      }
      await createPromotion(form.jobId, form.type, {
        adTitle: form.adTitle.trim(),
        adSubtitle: form.adSubtitle.trim(),
        adImage: form.adImage
      })
      ElMessage.success('已投放到首页')
      form.adTitle = ''
      form.adSubtitle = ''
      form.adImage = ''
      await load()
    } finally {
      submitting.value = false
    }
  }

  const remove = async (row: CompanyPromotionItem) => {
    try {
      await ElMessageBox.confirm(`确认撤下「${row.jobsName}」吗？`, '撤下推广', { type: 'warning' })
    } catch {
      return
    }
    await deletePromotion(row.id)
    ElMessage.success('已撤下，名额已释放')
    await load()
  }

  const onAdImageUploaded = (res: { code: number; message: string; data: { url: string } }) => {
    form.adImage = res.data.url
    ElMessage.success('广告横幅已上传')
  }

  const beforeUpload = (file: File) => {
    if (!file.type.startsWith('image/')) {
      ElMessage.warning('只能上传图片')
      return false
    }
    if (file.size / 1024 / 1024 > 5) {
      ElMessage.warning('图片不能超过 5MB')
      return false
    }
    return true
  }

  const imageUrl = (url: string) => {
    if (/^https?:\/\//i.test(url)) return url
    return `/${url.replace(/^\/+/, '')}`
  }

  onMounted(load)
</script>
