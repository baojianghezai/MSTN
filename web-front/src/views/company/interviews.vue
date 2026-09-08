<template>
  <div class="mx-auto max-w-6xl px-4 py-10">
    <div class="flex items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-slate-800">面试邀请</h1>
        <p class="mt-1 text-sm text-slate-500">已发出的线下面试安排</p>
      </div>
      <el-button type="primary" @click="router.push({ name: 'CompanyApplies' })">
        <span class="mr-1 i-lucide-inbox" />从收简历发起
      </el-button>
    </div>

    <el-card class="mt-6" shadow="never">
      <div v-if="loading && tableData.length === 0" class="space-y-4 py-2">
        <el-skeleton v-for="i in 4" :key="i" :rows="1" animated />
      </div>
      <el-table v-else v-loading="loading" :data="tableData">
        <template #empty>
          <el-empty :image-size="80" description="暂无面试邀请" />
        </template>
        <el-table-column label="候选人" min-width="110" show-overflow-tooltip>
          <template #default="{ row }">{{ row.resumeName }}</template>
        </el-table-column>
        <el-table-column label="职位" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">{{ row.jobsName }}</template>
        </el-table-column>
        <el-table-column label="面试时间" width="170">
          <template #default="{ row }">{{ formatTime(row.interviewTime) }}</template>
        </el-table-column>
        <el-table-column label="地址" min-width="170" show-overflow-tooltip>
          <template #default="{ row }">{{ row.address }}</template>
        </el-table-column>
        <el-table-column label="联系人" width="150">
          <template #default="{ row }">{{ row.contact }} {{ row.telephone }}</template>
        </el-table-column>
        <el-table-column label="阅读状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.personalLook === 2 ? 'success' : 'info'" size="small">
              {{ row.personalLook === 2 ? '已读' : '未读' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="90" align="center">
          <template #default="{ row }">
            <el-button type="danger" link @click="handleWithdraw(row)">撤回</el-button>
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
  import { useRouter } from 'vue-router'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { getCompanyInterviews, withdrawInterview } from '@/api/interview'
  import type { CompanyInterview } from '@/api/interview'
  import { formatTime } from '@/utils/format'

  const router = useRouter()
  const loading = ref(false)
  const tableData = ref<CompanyInterview[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)

  const getList = async () => {
    loading.value = true
    try {
      const { data } = await getCompanyInterviews({
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

  const handleWithdraw = async (row: CompanyInterview) => {
    try {
      await ElMessageBox.confirm('撤回后候选人将无法再查看该邀请，确认撤回？', '撤回面试邀请', {
        confirmButtonText: '撤回',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }
    await withdrawInterview(row.did)
    ElMessage.success('面试邀请已撤回')
    await getList()
  }
</script>
