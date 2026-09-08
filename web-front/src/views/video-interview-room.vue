<template>
  <div class="mx-auto max-w-xl px-4 py-16">
    <h1 class="text-2xl font-bold text-slate-800">视频面试房间</h1>

    <el-card v-loading="loading" class="mt-6">
      <template v-if="room">
        <div class="space-y-3 text-sm text-slate-600">
          <p><span class="font-semibold">职位：</span>{{ room.jobsName || '-' }}</p>
          <p><span class="font-semibold">面试时间：</span>{{ formatTime(room.interviewTime) }}</p>
          <p>
            <span class="font-semibold">房间状态：</span>
            <el-tag :type="roomTag(room.roomStatus)">{{ roomStatusCn(room.roomStatus) }}</el-tag>
          </p>
          <p><span class="font-semibold">身份：</span>{{ room.utype === 1 ? '个人' : '企业' }}</p>
        </div>

        <el-button
          type="primary"
          class="mt-6 w-full"
          :disabled="room.roomStatus !== 'opened'"
          @click="handleEnter"
        >
          {{ room.roomStatus === 'opened' ? '进入面试房间' : roomStatusCn(room.roomStatus) }}
        </el-button>
        <p class="mt-3 text-center text-xs text-slate-400">
          房间有效期为面试时间起 15 天；仅面试当天（opened）可进入
        </p>
      </template>
      <el-empty v-else-if="!loading" description="房间码无效或面试不存在" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, ref } from 'vue'
  import { useRoute } from 'vue-router'
  import { ElMessage } from 'element-plus'
  import { getVideoInterviewRoom } from '@/api/videoInterview'
  import { formatTime } from '@/utils/format'
  import type { VideoInterviewRoom } from '@/types/api'

  const route = useRoute()
  const code = String(route.params.code || '')

  const loading = ref(false)
  const room = ref<VideoInterviewRoom | null>(null)

  onMounted(async () => {
    loading.value = true
    try {
      const { data } = await getVideoInterviewRoom(code)
      room.value = data
    } catch {
      // 房间码无效，展示空态
      room.value = null
    } finally {
      loading.value = false
    }
  })

  const handleEnter = () => {
    // TRTC 入房：SDK 集成前端侧（一期骨架先占位提示，M3/M4 联调时接入 TRTC）
    ElMessage.info('视频通话（TRTC）接入待联调，房间码已就绪')
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
