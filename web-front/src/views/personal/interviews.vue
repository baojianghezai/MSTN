<template>
  <div class="mx-auto max-w-5xl px-4 py-10">
    <div>
      <h1 class="text-2xl font-bold text-slate-800">面试邀请</h1>
      <p class="mt-1 text-sm text-slate-500">企业发来的线下面试安排</p>
    </div>

    <el-card class="mt-6" shadow="never">
      <div v-if="loading && tableData.length === 0" class="space-y-4 py-2">
        <el-skeleton v-for="i in 4" :key="i" :rows="1" animated />
      </div>
      <el-table v-else v-loading="loading" :data="tableData">
        <template #empty>
          <el-empty :image-size="80" description="暂无面试邀请" />
        </template>
        <el-table-column label="企业" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">{{ row.companyName }}</template>
        </el-table-column>
        <el-table-column label="职位" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">{{ row.jobsName }}</template>
        </el-table-column>
        <el-table-column label="面试时间" width="170">
          <template #default="{ row }">{{ formatTime(row.interviewTime) }}</template>
        </el-table-column>
        <el-table-column label="地点" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.address }}</template>
        </el-table-column>
        <el-table-column label="联系人" width="150">
          <template #default="{ row }">{{ row.contact }} {{ row.telephone }}</template>
        </el-table-column>
        <el-table-column label="备注" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">{{ row.notes || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-button v-if="row.personalLook !== 2" type="primary" link @click="handleRead(row)">标记已读</el-button>
            <el-tag v-else type="success" size="small">已读</el-tag>
          </template>
        </el-table-column>
      </el-table>

      <div class="mt-4 flex justify-end">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @size-change="getList"
          @current-change="getList"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { getPersonalInterviews, markInterviewRead } from '@/api/interview'
  import type { CompanyInterview } from '@/api/interview'
  import { formatTime } from '@/utils/format'

  const loading = ref(false)
  const tableData = ref<CompanyInterview[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)

  const getList = async () => {
    loading.value = true
    try {
      const { data } = await getPersonalInterviews({
        page: page.value,
        pageSize: pageSize.value
      })
      tableData.value = data.list
      total.value = data.total
    } finally {
      loading.value = false
    }
  }

  onMounted(getList)

  const handleRead = async (row: CompanyInterview) => {
    await markInterviewRead(row.did)
    ElMessage.success('已标记为已读')
    await getList()
  }
</script>
