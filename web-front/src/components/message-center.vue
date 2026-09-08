<template>
  <div class="mx-auto max-w-5xl px-4 py-10">
    <div class="flex items-start justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-slate-800">站内信</h1>
        <p class="mt-1 text-sm text-slate-500">{{ description }}</p>
      </div>
      <el-tag v-if="messageStore.unread > 0" type="danger" effect="light">{{ messageStore.unread }} 条未读</el-tag>
      <el-tag v-else type="success" effect="light">全部已读</el-tag>
    </div>

    <el-card class="mt-6" shadow="never">
      <div class="mb-4 flex items-center justify-between gap-3">
        <span class="text-sm text-slate-500">共 {{ total }} 条消息</span>
        <div class="flex items-center gap-2">
          <el-button :disabled="selected.length === 0" @click="handleMarkRead">标记已读</el-button>
          <el-button type="danger" :disabled="selected.length === 0" @click="handleDelete">删除</el-button>
        </div>
      </div>

      <div v-if="loading && tableData.length === 0" class="space-y-4 py-2">
        <el-skeleton v-for="i in 4" :key="i" :rows="2" animated />
      </div>
      <el-table
        v-else
        v-loading="loading"
        :data="tableData"
        row-key="id"
        @selection-change="handleSelectionChange"
      >
        <template #empty>
          <el-empty :image-size="80" description="暂无站内信" />
        </template>
        <el-table-column type="selection" width="48" />
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.msgCheck === 0 ? 'danger' : 'info'" size="small">
              {{ row.msgCheck === 0 ? '未读' : '已读' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="消息" min-width="360">
          <template #default="{ row }">
            <button
              type="button"
              class="block max-w-full text-left"
              @click="handleOpen(row)"
            >
              <span :class="row.msgCheck === 0 ? 'font-semibold text-slate-800' : 'text-slate-700'">
                {{ row.title }}
              </span>
              <span class="mt-1 block text-sm leading-6 text-slate-500">{{ row.message }}</span>
            </button>
          </template>
        </el-table-column>
        <el-table-column label="时间" width="175" align="center">
          <template #default="{ row }">{{ formatTime(row.addtime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="90" align="center">
          <template #default="{ row }">
            <el-button v-if="row.msgCheck === 0" type="primary" link @click="handleReadOne(row)">已读</el-button>
            <span v-else class="text-sm text-slate-400">-</span>
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
          @size-change="loadList"
          @current-change="loadList"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { deleteMessages, getMessages, markMessagesRead } from '@/api/message'
  import type { MessageItem, MessageScope } from '@/api/message'
  import { useMessageStore } from '@/stores/message'
  import { formatTime } from '@/utils/format'

  const props = defineProps<{ scope: MessageScope }>()

  const router = useRouter()
  const messageStore = useMessageStore()
  const loading = ref(false)
  const tableData = ref<MessageItem[]>([])
  const selected = ref<MessageItem[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)
  const description = computed(() => props.scope === 'personal' ? '投递进展、简历下载和面试邀请会在这里通知您' : '新职位投递及后续业务提醒会在这里通知您')

  const loadUnreadCount = async () => {
    await messageStore.refreshUnread(props.scope)
  }

  const loadList = async () => {
    loading.value = true
    try {
      const { data } = await getMessages(props.scope, {
        page: page.value,
        pageSize: pageSize.value
      })
      tableData.value = data.list
      total.value = data.total
      selected.value = []
      await loadUnreadCount()
    } finally {
      loading.value = false
    }
  }

  onMounted(loadList)

  const handleSelectionChange = (rows: MessageItem[]) => {
    selected.value = rows
  }

  const handleMarkRead = async () => {
    const ids = selected.value.filter((item) => item.msgCheck === 0).map((item) => item.id)
    if (ids.length === 0) {
      ElMessage.info('所选消息均已读')
      return
    }
    await markMessagesRead(props.scope, ids)
    ElMessage.success('已标记为已读')
    await loadList()
  }

  const handleReadOne = async (row: MessageItem) => {
    await markMessagesRead(props.scope, [row.id])
    await loadList()
  }

  const handleDelete = async () => {
    try {
      await ElMessageBox.confirm(`确认删除选中的 ${selected.value.length} 条消息？`, '删除站内信', {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }
    await deleteMessages(props.scope, selected.value.map((item) => item.id))
    ElMessage.success('已删除')
    if (tableData.value.length === selected.value.length && page.value > 1) {
      page.value -= 1
    }
    await loadList()
  }

  const handleOpen = async (row: MessageItem) => {
    if (row.msgCheck === 0) {
      await markMessagesRead(props.scope, [row.id])
      await loadList()
    }
    if (row.link.startsWith('/')) {
      await router.push(row.link)
    }
  }
</script>
