<template>
  <div class="mx-auto max-w-4xl px-4 py-16">
    <h1 class="text-2xl font-bold text-slate-800">我的投递</h1>
    <p class="mt-2 text-sm text-slate-500">查看投递进度与企业的回复状态</p>

    <!-- 状态筛选 -->
    <div class="mt-6 flex items-center gap-3">
      <el-radio-group v-model="status" @change="handleSearch">
        <el-radio-button :value="0">全部</el-radio-button>
        <el-radio-button :value="1">未读</el-radio-button>
        <el-radio-button :value="2">已读</el-radio-button>
      </el-radio-group>
    </div>

    <el-card class="mt-4">
      <div v-if="loading && tableData.length === 0" class="space-y-4 py-2">
        <el-skeleton v-for="i in 4" :key="i" :rows="1" animated />
      </div>
      <el-table v-else v-loading="loading" :data="tableData">
        <template #empty>
          <el-empty :image-size="80" description="还没有投递记录">
            <el-button type="primary" @click="router.push('/jobs')">去职位列表</el-button>
          </el-empty>
        </template>
        <el-table-column label="职位" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.jobsName }}</template>
        </el-table-column>
        <el-table-column label="企业" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.companyName }}</template>
        </el-table-column>
        <el-table-column label="投递时间" width="170" align="center">
          <template #default="{ row }">{{ formatTime(row.applyAddtime) }}</template>
        </el-table-column>
        <el-table-column label="已读" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.personalLook === 2 ? 'success' : 'info'" size="small">
              {{ row.personalLook === 2 ? '已读' : '未读' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="企业回复" width="100" align="center">
          <template #default="{ row }">
            <span :class="replyClass(row.isReply)">{{ replyCn(row.isReply) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="90" align="center">
          <template #default="{ row }">
            <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
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
  import { deleteApply, getApplies } from '@/api/apply'
  import type { ApplyItem } from '@/api/apply'
  import { formatTime } from '@/utils/format'

  const router = useRouter()
  const loading = ref(false)
  const tableData = ref<ApplyItem[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)
  const status = ref(0)

  const getList = async () => {
    loading.value = true
    try {
      const { data } = await getApplies({
        page: page.value,
        pageSize: pageSize.value,
        status: status.value
      })
      tableData.value = data.list
      total.value = data.total
    } finally {
      loading.value = false
    }
  }

  onMounted(getList)

  const handleSearch = () => {
    page.value = 1
    getList()
  }

  const handleDelete = async (row: ApplyItem) => {
    try {
      await ElMessageBox.confirm('确认删除这条投递记录？', '提示', {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }
    await deleteApply(row.did)
    ElMessage.success('已删除')
    getList()
  }

  const replyCn = (isReply: number) => {
    const map: Record<number, string> = { 0: '待反馈', 1: '合适', 2: '不合适', 3: '待定', 4: '未接通' }
    return map[isReply] ?? '待反馈'
  }
  const replyClass = (isReply: number) => {
    const map: Record<number, string> = { 0: 'text-slate-400', 1: 'text-green-600', 2: 'text-danger-500', 3: 'text-amber-500', 4: 'text-slate-500' }
    return map[isReply] ?? 'text-slate-400'
  }
</script>
