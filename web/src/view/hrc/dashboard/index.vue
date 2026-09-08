<template>
  <div>
    <!-- 指标卡：今日 -->
    <div class="grid grid-cols-2 gap-4 md:grid-cols-4 xl:grid-cols-7">
      <el-card shadow="never" v-for="card in metricCards" :key="card.key" class="text-center">
        <p class="text-xs text-slate-500">{{ card.label }}</p>
        <p class="mt-2 text-2xl font-semibold text-slate-800">{{ card.today }}</p>
        <p class="mt-1 text-xs text-slate-400">昨日 {{ card.yesterday }}</p>
      </el-card>
    </div>

    <div class="mt-4 grid grid-cols-1 gap-4 xl:grid-cols-3">
      <!-- 趋势图 -->
      <el-card shadow="never" class="xl:col-span-2">
        <template #header>
          <div class="flex items-center justify-between">
            <span class="font-semibold">趋势图</span>
            <div class="flex items-center gap-2">
              <el-select v-model="trend.metric" class="w-32" @change="loadTrend">
                <el-option label="注册" value="register" />
                <el-option label="简历" value="resume" />
                <el-option label="企业" value="company" />
                <el-option label="职位" value="job" />
                <el-option label="投递" value="application" />
              </el-select>
              <el-select v-model="trend.days" class="w-28" @change="loadTrend">
                <el-option label="近 7 天" :value="7" />
                <el-option label="近 30 天" :value="30" />
                <el-option label="近 90 天" :value="90" />
              </el-select>
            </div>
          </div>
        </template>
        <Chart :option="trendOption" height="320px" />
      </el-card>

      <!-- 待办 -->
      <el-card shadow="never">
        <template #header>
          <span class="font-semibold">待办事项</span>
        </template>
        <div class="space-y-3">
          <div
            v-for="todo in todoCards"
            :key="todo.key"
            class="flex cursor-pointer items-center justify-between rounded border border-slate-200 px-3 py-2 hover:bg-slate-50"
            @click="goTodo(todo)"
          >
            <span class="text-sm text-slate-700">{{ todo.label }}</span>
            <el-tag :type="todo.count > 0 ? 'danger' : 'info'" round>
              {{ todo.count }}
            </el-tag>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 收入占位 -->
    <el-card shadow="never" class="mt-4">
      <template #header>
        <span class="font-semibold">收入</span>
      </template>
      <p class="text-sm text-slate-400">
        今日 {{ income.today }} 元 · 本月 {{ income.month }} 元（订单模块 M4 建设后生效，当前恒 0）
      </p>
    </el-card>
  </div>
</template>

<script setup>
  import { computed, onMounted, reactive, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import Chart from '@/components/charts/index.vue'
  import { getDashboard, getTrend } from '@/api/hrc/dashboard'

  defineOptions({
    name: 'HrcDashboard'
  })

  const router = useRouter()

  const today = ref({})
  const yesterday = ref({})
  const todo = ref({})
  const income = ref({ today: 0, month: 0 })
  const trendData = ref([])

  // 筛选器切换即自动加载（filter 自动加载，进页面默认 days=30&metric=register）
  const trend = reactive({ days: 30, metric: 'register' })

  const metricCards = computed(() => [
    { key: 'personalUsers', label: '今日个人注册', today: today.value.personalUsers ?? 0, yesterday: yesterday.value.personalUsers ?? 0 },
    { key: 'companyUsers', label: '今日企业注册', today: today.value.companyUsers ?? 0, yesterday: yesterday.value.companyUsers ?? 0 },
    { key: 'resumes', label: '今日简历', today: today.value.resumes ?? 0, yesterday: yesterday.value.resumes ?? 0 },
    { key: 'companies', label: '今日企业资料', today: today.value.companies ?? 0, yesterday: yesterday.value.companies ?? 0 },
    { key: 'jobs', label: '今日职位', today: today.value.jobs ?? 0, yesterday: yesterday.value.jobs ?? 0 },
    { key: 'applications', label: '今日投递', today: today.value.applications ?? 0, yesterday: yesterday.value.applications ?? 0 },
    { key: 'videoInterviews', label: '今日视频面试', today: today.value.videoInterviews ?? 0, yesterday: yesterday.value.videoInterviews ?? 0 }
  ])

  const todoCards = computed(() => [
    { key: 'companyAudit', label: '待审企业', count: todo.value.companyAudit ?? 0, route: '/hrc/company-audit' },
    { key: 'resumeAudit', label: '待审简历', count: todo.value.resumeAudit ?? 0, route: '' },
    { key: 'jobAudit', label: '待审职位', count: todo.value.jobAudit ?? 0, route: '/hrc/jobs-manage?tab=tmp' },
    { key: 'appeal', label: '待处理申诉', count: todo.value.appeal ?? 0, route: '/hrc/appeal' },
    { key: 'companyCancellation', label: '待处理注销', count: todo.value.companyCancellation ?? 0, route: '/hrc/company-cancellation' }
  ])

  const goTodo = (t) => {
    if (t.route) {
      router.push(t.route)
    }
  }

  const loadDashboard = async () => {
    const { data } = await getDashboard()
    today.value = data.today || {}
    yesterday.value = data.yesterday || {}
    todo.value = data.todo || {}
    income.value = data.income || { today: 0, month: 0 }
  }

  const loadTrend = async () => {
    const { data } = await getTrend({ days: trend.days, metric: trend.metric })
    trendData.value = data || []
  }

  // 趋势图 option：register 双线（personal/company），resume/company 单线（count）
  const trendOption = computed(() => {
    const dates = trendData.value.map((d) => d.date)
    if (trend.metric === 'register') {
      return {
        tooltip: { trigger: 'axis' },
        legend: { data: ['个人', '企业'] },
        xAxis: { type: 'category', data: dates },
        yAxis: { type: 'value', minInterval: 1 },
        series: [
          { name: '个人', type: 'line', smooth: true, data: trendData.value.map((d) => d.personal) },
          { name: '企业', type: 'line', smooth: true, data: trendData.value.map((d) => d.company) }
        ]
      }
    }
    const nameMap = { resume: '简历', company: '企业', job: '职位', application: '投递' }
    return {
      tooltip: { trigger: 'axis' },
      legend: { data: [nameMap[trend.metric]] },
      xAxis: { type: 'category', data: dates },
      yAxis: { type: 'value', minInterval: 1 },
      series: [
        {
          name: nameMap[trend.metric],
          type: 'line',
          smooth: true,
          areaStyle: {},
          data: trendData.value.map((d) => d.count)
        }
      ]
    }
  })

  onMounted(() => {
    loadDashboard()
    loadTrend()
  })
</script>
