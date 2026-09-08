<template>
  <article
    class="cursor-pointer rounded-xl border border-slate-200 bg-white p-5 shadow-card transition-all duration-200 hover:-translate-y-0.5 hover:border-primary-200 hover:shadow-card-hover"
    @click="goDetail"
  >
    <div class="flex items-start gap-4">
      <!-- 公司首字 logo 占位（无企业 logo 接口，一期用首字色块） -->
      <div
        class="flex h-12 w-12 flex-none items-center justify-center rounded-xl bg-primary-50 text-lg font-bold text-primary-600"
      >
        {{ job.companyname.charAt(0) }}
      </div>

      <div class="min-w-0 flex-1">
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0 flex-1">
            <div class="flex flex-wrap items-center gap-2">
              <h2 class="truncate text-lg font-semibold text-slate-800">{{ job.jobsName }}</h2>
              <el-tag v-if="job.emergency === 1" type="danger" size="small">急聘</el-tag>
              <el-tag v-if="job.stick === 1" type="warning" size="small">置顶</el-tag>
            </div>
            <p class="mt-1 truncate text-sm text-slate-600">{{ job.companyname }}</p>
          </div>

          <div class="flex-none text-right">
            <div class="text-lg font-bold text-accent-600">{{ salaryText }}</div>
            <div class="mt-1 text-xs text-slate-400">{{ formatDate(job.refreshtime || job.addtime) }}</div>
          </div>
        </div>

        <div class="mt-3 flex flex-wrap items-center gap-2 text-xs text-slate-500">
          <span
            v-if="job.districtCn"
            class="flex items-center gap-1 rounded-md bg-slate-100 px-2 py-1"
          >
            <span class="i-lucide-map-pin text-primary-500" aria-hidden="true" />
            {{ job.districtCn }}
          </span>
          <span v-if="job.categoryCn" class="rounded-md bg-slate-100 px-2 py-1">{{ job.categoryCn }}</span>
          <span v-if="job.natureCn" class="rounded-md bg-slate-100 px-2 py-1">{{ job.natureCn }}</span>
          <span v-if="eduName" class="rounded-md bg-slate-100 px-2 py-1">{{ eduName }}</span>
          <span v-if="expName" class="rounded-md bg-slate-100 px-2 py-1">{{ expName }}</span>
        </div>
      </div>
    </div>
  </article>
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import { useRouter } from 'vue-router'
  import type { PublicJobItem } from '@/api/jobs'
  import { formatDate } from '@/utils/format'

  // 职位卡片：首页「最新招聘」「优先推荐职位」与职位列表页共用
  const props = defineProps<{
    job: PublicJobItem
    eduName?: string
    expName?: string
  }>()

  const router = useRouter()

  const salaryText = computed(() =>
    props.job.negotiable === 1 ? '面议' : `${props.job.minwage}-${props.job.maxwage} 元`
  )

  const goDetail = () => {
    router.push(`/jobs/${props.job.id}`)
  }
</script>
