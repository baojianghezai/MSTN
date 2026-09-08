<template>
  <div class="mx-auto max-w-7xl px-4 py-8" v-loading="loading">
    <template v-if="detail">
      <section class="flex flex-col gap-5 border-b border-slate-200 pb-6 sm:flex-row sm:items-start">
        <img v-if="detail.company.logo" :src="imageUrl(detail.company.logo)" :alt="detail.company.companyname" class="h-20 w-20 border border-slate-100 object-cover" />
        <div v-else class="flex h-20 w-20 flex-none items-center justify-center bg-primary-50 text-2xl font-semibold text-primary-600">{{ detail.company.companyname.slice(0, 1) }}</div>
        <div class="min-w-0 flex-1">
          <div class="flex flex-wrap items-center gap-2">
            <h1 class="text-2xl font-bold text-slate-800">{{ detail.company.companyname }}</h1>
            <el-tag type="success" effect="plain"><span class="mr-1 i-lucide-badge-check" aria-hidden="true" />已认证</el-tag>
          </div>
          <p v-if="detail.company.shortDesc" class="mt-2 text-sm text-slate-600">{{ detail.company.shortDesc }}</p>
          <div class="mt-3 flex flex-wrap gap-x-4 gap-y-2 text-sm text-slate-500">
            <span v-if="detail.company.districtCn" class="flex items-center gap-1"><span class="i-lucide-map-pin" aria-hidden="true" />{{ detail.company.districtCn }}</span>
            <span v-if="detail.company.tradeCn">{{ detail.company.tradeCn }}</span>
            <span v-if="detail.company.scaleCn">{{ detail.company.scaleCn }}</span>
          </div>
        </div>
      </section>

      <div class="mt-7 grid gap-8 lg:grid-cols-[minmax(0,1fr)_260px]">
        <div>
          <section>
            <h2 class="text-lg font-semibold text-slate-800">企业介绍</h2>
            <p class="mt-3 whitespace-pre-line text-sm leading-7 text-slate-600">{{ detail.contents || '该企业暂未完善介绍。' }}</p>
          </section>

          <section class="mt-8">
            <div class="flex items-center justify-between gap-4">
              <h2 class="text-lg font-semibold text-slate-800">在招职位</h2>
              <span class="text-sm text-slate-500">{{ jobsTotal }} 个职位</span>
            </div>
            <div v-if="jobs.length" class="mt-3 divide-y divide-slate-200 border-y border-slate-200">
              <router-link v-for="job in jobs" :key="job.id" :to="`/jobs/${job.id}`" class="flex items-center justify-between gap-4 px-1 py-4 transition-colors hover:bg-slate-50">
                <div class="min-w-0">
                  <div class="flex flex-wrap items-center gap-2"><span class="font-medium text-slate-800">{{ job.jobsName }}</span><el-tag v-if="job.stick === 1" size="small" type="warning">置顶</el-tag><el-tag v-if="job.emergency === 1" size="small" type="danger">急聘</el-tag></div>
                  <p class="mt-1 truncate text-sm text-slate-500">{{ [job.districtCn, job.categoryCn, job.natureCn].filter(Boolean).join(' · ') || '招聘中' }}</p>
                </div>
                <span class="flex-none text-sm font-semibold text-primary-600">{{ salaryText(job) }}</span>
              </router-link>
            </div>
            <el-empty v-else description="该企业暂时没有在招职位" :image-size="96" />
            <div v-if="isGuest && jobsTotal > jobs.length" class="mt-5 flex justify-center"><el-button type="primary" @click="goLogin">登录查看全部职位</el-button></div>
            <div v-else-if="!isGuest && jobsTotal > 0" class="mt-5 flex justify-end"><el-pagination v-model:current-page="jobPage.page" v-model:page-size="jobPage.pageSize" :total="jobsTotal" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next" @size-change="loadJobs" @current-change="loadJobs" /></div>
          </section>
        </div>

        <aside class="border-l border-slate-200 pl-5 text-sm">
          <h2 class="font-semibold text-slate-800">企业资料</h2>
          <dl class="mt-4 space-y-4 text-slate-600">
            <div v-if="detail.company.natureCn"><dt class="text-slate-400">企业性质</dt><dd class="mt-1">{{ detail.company.natureCn }}</dd></div>
            <div v-if="detail.company.tradeCn"><dt class="text-slate-400">所属行业</dt><dd class="mt-1">{{ detail.company.tradeCn }}</dd></div>
            <div v-if="detail.company.scaleCn"><dt class="text-slate-400">企业规模</dt><dd class="mt-1">{{ detail.company.scaleCn }}</dd></div>
            <div v-if="detail.address"><dt class="text-slate-400">办公地址</dt><dd class="mt-1 leading-6">{{ detail.address }}</dd></div>
            <div v-if="detail.website"><dt class="text-slate-400">企业官网</dt><dd class="mt-1"><a :href="detail.website" target="_blank" rel="noopener noreferrer" class="break-all text-primary-600 hover:underline">访问官网</a></dd></div>
          </dl>
          <div v-if="tags.length" class="mt-6 border-t border-slate-200 pt-5"><h3 class="font-medium text-slate-700">企业标签</h3><div class="mt-3 flex flex-wrap gap-2"><el-tag v-for="tag in tags" :key="tag" effect="plain" type="info">{{ tag }}</el-tag></div></div>
        </aside>
      </div>
    </template>
    <el-empty v-else-if="!loading" description="企业不存在或暂不可见" />
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref, watch } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { getPublicCompany, getPublicCompanyJobs, type PublicCompanyDetail, type PublicCompanyJob } from '@/api/companies'
  import { useUserStore } from '@/stores/user'

  const route = useRoute()
  const router = useRouter()
  const userStore = useUserStore()
  const companyId = Number(route.params.id)
  const loading = ref(false)
  const detail = ref<PublicCompanyDetail | null>(null)
  const jobs = ref<PublicCompanyJob[]>([])
  const jobsTotal = ref(0)
  const isGuest = computed(() => !userStore.token)
  const jobPage = reactive({ page: 1, pageSize: 10 })
  const tags = computed(() => detail.value?.company.tag.split(/[,，]/).map((item) => item.trim()).filter(Boolean).slice(0, 6) || [])
  const imageUrl = (url: string) => (url.startsWith('http') ? url : `/${url.replace(/^\//, '')}`)
  const salaryText = (job: PublicCompanyJob) => job.negotiable === 1 ? '面议' : `${job.minwage}-${job.maxwage} 元`
  const goLogin = () => router.push({ name: 'Login', query: { redirect: route.fullPath } })

  const loadJobs = async () => {
    if (isGuest.value) return
    const { data } = await getPublicCompanyJobs(companyId, jobPage)
    jobs.value = data.list
    jobsTotal.value = data.total
  }

  const loadDetail = async () => {
    loading.value = true
    try {
      const { data } = await getPublicCompany(companyId)
      detail.value = data
      jobs.value = data.jobs
      jobsTotal.value = data.jobsTotal
      if (!isGuest.value) await loadJobs()
    } catch {
      detail.value = null
      jobs.value = []
      jobsTotal.value = 0
    } finally {
      loading.value = false
    }
  }

  onMounted(loadDetail)
  watch(() => userStore.token, () => { jobPage.page = 1; loadDetail() })
</script>
