<template>
  <div class="mx-auto max-w-7xl px-4 py-8">
    <section class="border-b border-slate-200 pb-6">
      <p class="flex items-center gap-2 text-sm font-medium text-primary-600"><span class="i-lucide-briefcase-business" aria-hidden="true" />职位广场</p>
      <h1 class="mt-2 text-2xl font-bold text-slate-800">发现合适的工作</h1>
      <p class="mt-2 text-sm text-slate-500">按职位、地区、行业和薪资筛选正在招聘的岗位</p>
    </section>

    <!-- 筛选区 -->
    <el-card shadow="never" class="mt-6">
      <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
        <el-input v-model="query.keyword" placeholder="职位名/企业名" clearable @keyup.enter="handleSearch" />
        <el-select v-model="query.district" placeholder="城市" clearable>
          <el-option v-for="c in filters.district" :key="c.id" :label="c.name" :value="c.name" />
        </el-select>
        <el-select v-model="query.trade" placeholder="行业" clearable>
          <el-option v-for="c in filters.trade" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>
        <el-select v-model="query.education" placeholder="学历" clearable>
          <el-option v-for="c in filters.education" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>
        <el-select v-model="query.experience" placeholder="经验" clearable>
          <el-option v-for="c in filters.experience" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>

        <el-input v-model.number="query.minwage" placeholder="最低薪资（元）" clearable />
        <el-input v-model.number="query.maxwage" placeholder="最高薪资（元）" clearable />
        <el-select v-model="query.order" placeholder="排序">
          <el-option label="最新刷新" value="last" />
          <el-option label="发布时间" value="addtime" />
          <el-option label="薪资最高" value="salary" />
          <el-option label="置顶优先" value="stick" />
        </el-select>
        <div class="flex gap-2">
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </div>
      </div>
    </el-card>

    <!-- 列表（卡片组件与首页共用，design/10 §3.2 components/JobsCard.vue） -->
    <div v-loading="loading" class="mt-4">
      <template v-if="loading && tableData.length === 0">
        <el-skeleton v-for="i in 6" :key="i" :rows="2" animated class="mb-4" />
      </template>
      <el-empty v-else-if="tableData.length === 0" description="暂无符合条件的职位" />
      <div v-else class="space-y-3">
        <JobsCard
          v-for="job in tableData"
          :key="job.id"
          :job="job"
          :edu-name="eduName(job.education)"
          :exp-name="expName(job.experience)"
        />
      </div>
    </div>

    <!-- 分页 -->
    <div v-if="isGuest" class="mt-6 flex justify-center">
      <el-button type="primary" @click="router.push({ name: 'Login', query: { redirect: route.fullPath } })">
        登录查看更多岗位
      </el-button>
    </div>
    <div v-else class="mt-6 flex justify-end">
      <el-pagination
        v-model:current-page="query.page"
        v-model:page-size="query.pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @size-change="getList"
        @current-change="getList"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref, watch } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import JobsCard from '@/components/JobsCard.vue'
  import { getJobFilters, searchJobs } from '@/api/jobs'
  import { getDistricts } from '@/api/content'
  import type { PublicJobItem } from '@/api/jobs'
  import { useUserStore } from '@/stores/user'
  import type { Categories } from '@/types/api'

  const route = useRoute()
  const router = useRouter()
  const userStore = useUserStore()
  const GUEST_PREVIEW_SIZE = 4
  const isGuest = computed(() => !userStore.token)

  const loading = ref(false)
  const tableData = ref<PublicJobItem[]>([])
  const total = ref(0)

  const filters = ref<Categories>({
    education: [],
    experience: [],
    wage: [],
    trade: [],
    district: [],
    major: [],
    sex: [],
    marriage: [],
    nature: [],
    scale: []
  })

  const query = reactive({
    keyword: '',
    district: undefined as string | undefined,
    trade: undefined as number | undefined,
    education: undefined as number | undefined,
    experience: undefined as number | undefined,
    minwage: undefined as number | undefined,
    maxwage: undefined as number | undefined,
    order: 'last',
    page: 1,
    pageSize: 10
  })

  const getList = async () => {
    loading.value = true
    try {
      const { data } = await searchJobs({
        keyword: query.keyword || undefined,
        district: query.district,
        trade: query.trade,
        education: query.education,
        experience: query.experience,
        minwage: query.minwage,
        maxwage: query.maxwage,
        order: query.order,
        page: query.page,
        pageSize: isGuest.value ? GUEST_PREVIEW_SIZE : query.pageSize
      })
      tableData.value = data.list
      total.value = isGuest.value ? Math.min(data.total, GUEST_PREVIEW_SIZE) : data.total
    } finally {
      loading.value = false
    }
  }

  onMounted(async () => {
    // 深链初始化：首页大搜索框/顶栏搜索/行业快跳带 query 进来（/jobs?keyword=&district=&trade=）
    const q = route.query
    if (typeof q.keyword === 'string' && q.keyword) query.keyword = q.keyword
    if (typeof q.district === 'string' && q.district) query.district = q.district
    if (q.trade) query.trade = Number(q.trade)

    getList()
    try {
      const [filtersResponse, districtsResponse] = await Promise.all([getJobFilters(), getDistricts()])
      filters.value = { ...filtersResponse.data, district: districtsResponse.data }
    } catch {
      // 筛选元数据拉取失败不阻塞列表
    }
  })

  watch(
    () => userStore.token,
    () => {
      query.page = 1
      getList()
    }
  )

  const handleSearch = () => {
    query.page = 1
    getList()
  }

  const handleReset = () => {
    query.keyword = ''
    query.district = undefined
    query.trade = undefined
    query.education = undefined
    query.experience = undefined
    query.minwage = undefined
    query.maxwage = undefined
    query.order = 'last'
    query.page = 1
    getList()
  }

  const eduName = (code: number) => {
    const name = filters.value.education.find((c) => c.id === code)?.name
    return name ? `学历：${name}` : ''
  }
  const expName = (code: number) => {
    const name = filters.value.experience.find((c) => c.id === code)?.name
    return name ? `经验：${name}` : ''
  }
</script>
