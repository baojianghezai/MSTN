<template>
  <div class="mx-auto max-w-3xl px-4 py-16">
    <h1 class="text-2xl font-bold text-slate-800">视频面试邀请</h1>
    <p class="mt-2 text-sm text-slate-500">企业向你发起的视频面试，凭个人房间码进入面试房间</p>

    <el-card class="mt-6">
      <div v-if="loading && tableData.length === 0" class="space-y-4 py-2">
        <el-skeleton v-for="i in 4" :key="i" :rows="1" animated />
      </div>
      <el-table v-else v-loading="loading" :data="tableData">
        <template #empty>
          <el-empty :image-size="80" description="还没有视频面试安排" />
        </template>
        <el-table-column prop="id" label="ID" width="70" align="center" />
        <el-table-column label="职位" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.jobsName || '-' }}</template>
        </el-table-column>
        <el-table-column label="面试时间" width="170">
          <template #default="{ row }">{{ formatTime(row.interviewTime) }}</template>
        </el-table-column>
        <el-table-column label="房间状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="roomTag(row.roomStatus)">{{ roomStatusCn(row.roomStatus) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="130" align="center">
          <template #default="{ row }">
            <el-button type="primary" link @click="showDetail(row)">详情 / 进房间</el-button>
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

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="面试详情" width="440px">
      <div v-if="detail.id" class="space-y-2 text-sm text-slate-600">
        <p><span class="font-semibold">职位：</span>{{ detail.jobsName || '-' }}</p>
        <p><span class="font-semibold">面试时间：</span>{{ formatTime(detail.interviewTime) }}</p>
        <p><span class="font-semibold">房间状态：</span>{{ roomStatusCn(detail.roomStatus) }}</p>
        <p>
          <span class="font-semibold">个人房间码：</span>
          <span class="font-mono">{{ detail.personalCode || '-' }}</span>
        </p>
        <el-button
          type="primary"
          class="mt-4"
          :disabled="detail.roomStatus !== 'opened'"
          @click="enterRoom(detail.personalCode)"
        >
          进入面试房间
        </el-button>
        <p v-if="detail.roomStatus !== 'opened'" class="mt-2 text-xs text-slate-400">
          仅面试当天（opened）可进入房间
        </p>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { getPersonalVideoInterview, getPersonalVideoInterviews } from '@/api/videoInterview'
  import { formatTime } from '@/utils/format'
  import type { VideoInterview } from '@/types/api'

  const router = useRouter()

  const loading = ref(false)
  const tableData = ref<VideoInterview[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)

  const getList = async () => {
    loading.value = true
    try {
      const { data } = await getPersonalVideoInterviews({
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

  const detailVisible = ref(false)
  const detail = ref<Partial<VideoInterview>>({})
  const showDetail = async (row: VideoInterview) => {
    const { data } = await getPersonalVideoInterview(row.id as number)
    detail.value = data
    detailVisible.value = true
  }

  const enterRoom = (code?: string) => {
    if (code) {
      router.push(`/video-interviews/room/${code}`)
    }
  }

  const roomStatusCn = (s?: string) => {
    const map: Record<string, string> = { nostart: '未开始', opened: '可进入', overtime: '已过期' }
    return map[s || ''] || s || '-'
  }
  const roomTag = (s?: string) => {
    const map: Record<string, string> = { nostart: 'info', opened: 'success', overtime: 'warning' }
    return map[s || ''] || 'info'
  }
</script>
