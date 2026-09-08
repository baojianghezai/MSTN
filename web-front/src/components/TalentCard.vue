<template>
  <el-card shadow="never" class="h-full border-slate-200 transition-all duration-200 hover:-translate-y-0.5 hover:border-primary-200 hover:shadow-card-hover">
    <div class="flex h-full flex-col">
      <div class="flex items-start gap-3">
        <el-avatar :size="52" :src="resume.photoImg" class="flex-none bg-primary-50 text-primary-600">
          <span class="i-lucide-user-round" aria-hidden="true" />
        </el-avatar>
        <div class="min-w-0 flex-1">
          <div class="flex flex-wrap items-center gap-2">
            <h2 class="truncate text-base font-semibold text-slate-800">{{ resume.fullname }}</h2>
            <el-tag v-if="resume.talent === 1" type="warning" effect="light" size="small">高级人才</el-tag>
          </div>
          <p class="mt-1 truncate text-sm text-primary-600">{{ resume.intentionJobs || resume.title || '求职意向待补充' }}</p>
          <p class="mt-1 text-xs text-slate-500">{{ profileMeta }}</p>
        </div>
      </div>

      <div class="mt-4 flex flex-wrap gap-x-3 gap-y-2 text-sm text-slate-600">
        <span v-if="resume.educationCn" class="flex items-center gap-1">
          <span class="i-lucide-graduation-cap text-primary-500" aria-hidden="true" />
          {{ resume.educationCn }}
        </span>
        <span v-if="resume.experienceCn" class="flex items-center gap-1">
          <span class="i-lucide-briefcase-business text-primary-500" aria-hidden="true" />
          {{ resume.experienceCn }}
        </span>
        <span v-if="resume.districtCn" class="flex items-center gap-1">
          <span class="i-lucide-map-pin text-primary-500" aria-hidden="true" />
          {{ resume.districtCn }}
        </span>
      </div>

      <p v-if="resume.specialty" class="mt-3 line-clamp-2 min-h-10 text-sm leading-5 text-slate-500">
        {{ resume.specialty }}
      </p>
      <p v-else class="mt-3 min-h-10 text-sm text-slate-400">暂未填写个人技能说明</p>

      <div class="mt-4 flex items-center justify-between border-t border-slate-100 pt-3">
        <span class="font-medium text-accent-600">{{ resume.wageCn || '薪资面议' }}</span>
        <div class="flex items-center gap-3">
          <span class="text-xs text-slate-400">{{ formatDate(resume.refreshtime) }} 更新</span>
          <el-button type="primary" link @click="emit('view', resume)">查看简历</el-button>
        </div>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import type { PublicResume } from '@/api/talent'
  import { formatDate } from '@/utils/format'

  const props = defineProps<{
    resume: PublicResume
  }>()

  const emit = defineEmits<{
    view: [resume: PublicResume]
  }>()

  const profileMeta = computed(() => [props.resume.sexCn, props.resume.majorCn].filter(Boolean).join(' · ') || '公开简历')
</script>
