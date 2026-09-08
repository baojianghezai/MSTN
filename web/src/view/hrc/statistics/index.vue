<template>
  <div class="grid grid-cols-1 gap-4 xl:grid-cols-2">
    <!-- 求职者分布 -->
    <el-card shadow="never">
      <template #header>
        <span class="font-semibold">求职者分布</span>
      </template>

      <div class="mb-4 flex items-center gap-2 text-sm text-slate-500">
        <span>性别</span>
        <el-divider direction="vertical" />
        <span v-for="s in resume.sex" :key="s.code" class="mr-3">
          {{ s.cn }} {{ s.count }}
        </span>
      </div>
      <Chart :option="pieOption(resume.sex)" height="220px" />

      <el-divider />

      <Chart :option="barOption('学历', resume.education)" height="220px" />
      <el-divider />
      <Chart :option="barOption('工作经验', resume.experience)" height="220px" />
    </el-card>

    <!-- 企业分布 -->
    <el-card shadow="never">
      <template #header>
        <span class="font-semibold">企业分布</span>
      </template>

      <el-divider />
      <Chart :option="barOption('企业性质', company.nature)" height="220px" />
      <el-divider />
      <Chart :option="barOption('企业规模', company.scale)" height="220px" />
    </el-card>
  </div>
</template>

<script setup>
  import { onMounted, ref } from 'vue'
  import Chart from '@/components/charts/index.vue'
  import { getCompanyStatistics, getResumeStatistics } from '@/api/hrc/dashboard'

  defineOptions({
    name: 'HrcStatistics'
  })

  const resume = ref({ sex: [], education: [], experience: [] })
  const company = ref({ nature: [], scale: [] })

  const load = async () => {
    const [r, c] = await Promise.all([getResumeStatistics(), getCompanyStatistics()])
    resume.value = r.data
    company.value = c.data
  }

  // 分布项统一转饼图（{code, cn, count}）
  const pieOption = (list) => ({
    tooltip: { trigger: 'item' },
    series: [
      {
        type: 'pie',
        radius: ['40%', '70%'],
        data: list.map((i) => ({ name: i.cn, value: i.count }))
      }
    ]
  })

  // 分布项统一转柱状图
  const barOption = (title, list) => ({
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: list.map((i) => i.cn) },
    yAxis: { type: 'value', minInterval: 1 },
    series: [
      {
        name: title,
        type: 'bar',
        data: list.map((i) => i.count)
      }
    ]
  })

  onMounted(load)
</script>
