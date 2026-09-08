<template>
  <article
    class="cursor-pointer rounded-lg border border-slate-200 bg-white p-5 shadow-card transition-all duration-200 hover:-translate-y-0.5 hover:border-primary-300 hover:bg-primary-50/30 hover:shadow-card-hover"
    @click="router.push(`/companies/${company.id}`)"
  >
    <div class="flex items-start gap-4">
      <img v-if="company.logo" :src="imageUrl(company.logo)" :alt="company.companyname" class="h-14 w-14 flex-none rounded-md border border-slate-100 object-cover" />
      <div v-else class="flex h-14 w-14 flex-none items-center justify-center rounded-md bg-primary-50 text-lg font-semibold text-primary-600">
        {{ company.companyname.slice(0, 1) }}
      </div>
      <div class="min-w-0 flex-1">
        <div class="flex items-start justify-between gap-3">
          <h2 class="truncate text-base font-semibold text-slate-800">{{ company.companyname }}</h2>
          <span class="flex-none text-xs text-slate-400">{{ formatDate(company.refreshtime) }}</span>
        </div>
        <p v-if="company.shortDesc" class="mt-1 line-clamp-1 text-sm text-slate-600">{{ company.shortDesc }}</p>
        <div class="mt-3 flex flex-wrap gap-x-3 gap-y-1 text-xs text-slate-500">
          <span v-if="company.districtCn">{{ company.districtCn }}</span>
          <span v-if="company.natureCn">{{ company.natureCn }}</span>
          <span v-if="company.tradeCn">{{ company.tradeCn }}</span>
          <span v-if="company.scaleCn">{{ company.scaleCn }}</span>
        </div>
        <div class="mt-3 flex items-center justify-between gap-3">
          <div class="flex flex-wrap gap-1.5">
            <el-tag v-for="tag in tags" :key="tag" size="small" effect="plain" type="info">{{ tag }}</el-tag>
          </div>
          <span class="flex flex-none items-center gap-1 text-sm font-medium text-primary-600">
            <span class="i-lucide-briefcase-business" aria-hidden="true" />
            {{ company.jobsCount }} 个在招职位
          </span>
        </div>
      </div>
    </div>
  </article>
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import { useRouter } from 'vue-router'
  import type { PublicCompany } from '@/api/companies'
  import { formatDate } from '@/utils/format'

  const props = defineProps<{ company: PublicCompany }>()
  const router = useRouter()
  const tags = computed(() => props.company.tag.split(/[,，]/).map((item) => item.trim()).filter(Boolean).slice(0, 3))
  const imageUrl = (url: string) => (url.startsWith('http') ? url : `/${url.replace(/^\//, '')}`)
</script>
