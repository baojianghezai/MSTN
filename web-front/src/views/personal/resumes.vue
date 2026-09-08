<template>
  <div class="mx-auto max-w-3xl px-4 py-8">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-800">我的简历</h1>
      <el-button type="primary" @click="router.push({ name: 'ResumeNew' })">
        新建简历
      </el-button>
    </div>

    <el-card shadow="never" class="mt-6">
      <div v-loading="loading">
        <div v-if="loading && tableData.length === 0" class="space-y-4 py-2">
          <el-skeleton v-for="i in 4" :key="i" :rows="1" animated />
        </div>
        <el-empty
          v-else-if="tableData.length === 0"
          description="还没有简历，先建一份吧"
        >
          <el-button type="primary" @click="router.push({ name: 'ResumeNew' })">
            新建简历
          </el-button>
        </el-empty>

        <el-table v-else :data="tableData" row-key="id">
          <el-table-column label="简历标题" min-width="140">
            <template #default="{ row }">
              <div class="font-medium text-slate-700">{{ row.title || '未命名简历' }}</div>
            </template>
          </el-table-column>
          <el-table-column label="完善度" width="120">
            <template #default="{ row }">
              <el-progress
                :percentage="row.completePercent"
                :stroke-width="8"
                :color="row.completePercent >= 80 ? '#16a34a' : row.completePercent >= 40 ? '#3b82f6' : '#f97316'"
              />
            </template>
          </el-table-column>
          <el-table-column label="状态" width="105">
            <template #default="{ row }">
              <el-tag v-if="row.def === 1" type="primary" size="small">默认</el-tag>
              <el-tag :type="row.display === 1 ? 'success' : 'info'" size="small" class="ml-1">
                {{ row.display === 1 ? '公开' : '隐藏' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="最后更新" width="95">
            <template #default="{ row }">
              <span class="text-xs text-slate-400">{{ formatDate(row.refreshtime || row.addtime) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="230" align="right">
            <template #default="{ row }">
              <el-button type="primary" link @click="goEdit(row.id)">编辑</el-button>
              <el-button v-if="row.def !== 1" link @click="handleSetDefault(row)">
                设默认
              </el-button>
              <el-button link @click="handleToggleDisplay(row)">
                {{ row.display === 1 ? '设为隐藏' : '设为公开' }}
              </el-button>
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
      </div>
    </el-card>

    <p class="mt-3 text-xs text-slate-400">
      投递时使用「默认」简历；企业端仅能查看「公开」状态的简历。删除为软删除，投递记录保留。
    </p>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    deleteResume,
    listResumes,
    setResumeDefault,
    setResumeDisplay
  } from '@/api/resume'
  import type { ResumeLite } from '@/types/api'
  import { formatDate } from '@/utils/format'

  const router = useRouter()

  const loading = ref(false)
  const tableData = ref<ResumeLite[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)

  const getList = async () => {
    loading.value = true
    try {
      const { data } = await listResumes({ page: page.value, pageSize: pageSize.value })
      tableData.value = data.list
      total.value = data.total
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    getList()
  })

  const goEdit = (id: number) => {
    router.push(`/personal/resumes/${id}`)
  }

  const handleSetDefault = async (row: ResumeLite) => {
    try {
      await setResumeDefault(row.id)
      ElMessage.success('已设为默认简历')
      await getList()
    } catch {
      // 错误已由拦截器统一弹出
    }
  }

  const handleToggleDisplay = async (row: ResumeLite) => {
    const next = row.display === 1 ? 2 : 1
    try {
      await setResumeDisplay(row.id, next)
      ElMessage.success(next === 1 ? '已设为公开' : '已设为隐藏')
      await getList()
    } catch {
      // 错误已由拦截器统一弹出
    }
  }

  const handleDelete = async (row: ResumeLite) => {
    const warn =
      row.def === 1 ? '这是默认简历，删除后最近创建的另一份简历将自动成为默认。' : ''
    try {
      await ElMessageBox.confirm(
        `确定删除简历「${row.title || '未命名简历'}」？${warn}删除后企业端不可见，但投递记录保留。`,
        '删除确认',
        {
          confirmButtonText: '删除',
          cancelButtonText: '取消',
          type: 'warning'
        }
      )
    } catch {
      return // 用户取消
    }
    try {
      await deleteResume(row.id)
      ElMessage.success('已删除')
      await getList()
    } catch {
      // 错误已由拦截器统一弹出
    }
  }
</script>
