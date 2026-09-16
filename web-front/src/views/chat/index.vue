<template>
  <div class="mx-auto max-w-6xl px-4 py-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-800">在线对话</h1>
      <el-tag :type="chatStore.connected ? 'success' : 'info'" effect="light" size="small">
        {{ chatStore.connected ? '实时在线' : '轮询同步' }}
      </el-tag>
    </div>

    <el-card class="mt-4 overflow-hidden" shadow="never">
      <div class="flex h-[600px]">
        <!-- 会话列表 -->
        <aside class="flex w-72 flex-none flex-col border-r border-slate-200">
          <div class="border-b border-slate-100 px-4 py-3 text-sm font-medium text-slate-600">
            对话列表
            <span class="ml-1 text-xs text-slate-400">({{ sessions.length }})</span>
          </div>
          <div v-loading="loadingSessions" class="flex-1 overflow-y-auto">
            <el-empty v-if="!loadingSessions && sessions.length === 0" :image-size="70" description="暂无对话" />
            <div
              v-for="s in sessions"
              :key="s.id"
              class="flex cursor-pointer items-center gap-3 border-b border-slate-50 px-4 py-3 transition-colors"
              :class="s.id === activeId ? 'bg-primary-50' : 'hover:bg-slate-50'"
              @click="selectSession(s.id)"
            >
              <el-avatar :size="36" :src="avatarUrl(s.peerLogo)" class="bg-primary-50 text-primary-600">
                <span class="i-lucide-user-round" aria-hidden="true" />
              </el-avatar>
              <div class="min-w-0 flex-1">
                <div class="flex items-center justify-between gap-2">
                  <span class="truncate text-sm font-medium text-slate-700">{{ s.peerName }}</span>
                  <span class="flex-none text-xs text-slate-400">{{ formatTime(s.lastTime || s.updateTime) }}</span>
                </div>
                <p class="mt-0.5 truncate text-xs text-slate-500">{{ s.lastContent || '开始对话吧' }}</p>
              </div>
              <el-badge v-if="s.unread > 0" :value="s.unread" :max="99" />
            </div>
          </div>
        </aside>

        <!-- 消息区 -->
        <section class="flex min-w-0 flex-1 flex-col">
          <template v-if="activeSession">
            <div class="flex items-center justify-between border-b border-slate-100 px-4 py-3">
              <div class="min-w-0">
                <div class="flex items-center gap-2">
                  <span class="font-medium text-slate-800">{{ activeSession.peerName }}</span>
                  <el-tag v-if="activeSession.jobsName" size="small" type="info">{{ activeSession.jobsName }}</el-tag>
                </div>
              </div>
              <el-button type="danger" text size="small" @click="handleDeleteSession">删除会话</el-button>
            </div>

            <div ref="scrollRef" v-loading="loadingMessages" class="flex-1 space-y-3 overflow-y-auto bg-slate-50 px-4 py-4">
              <el-empty v-if="!loadingMessages && messages.length === 0" :image-size="70" description="还没有消息，发一句问候吧" />
              <div
                v-for="m in messages"
                :key="m.id"
                class="flex"
                :class="m.mine ? 'justify-end' : 'justify-start'"
              >
                <div
                  class="max-w-[70%] rounded-lg px-3 py-2 text-sm leading-6"
                  :class="m.mine ? 'bg-primary-600 text-white' : 'bg-white text-slate-700 shadow-sm'"
                >
                  {{ m.content }}
                </div>
              </div>
            </div>

            <div class="border-t border-slate-100 p-3">
              <div class="flex items-end gap-2">
                <el-input
                  v-model="draft"
                  type="textarea"
                  :rows="2"
                  maxlength="1000"
                  resize="none"
                  placeholder="输入消息，Enter 发送，Shift+Enter 换行"
                  @keydown.enter.exact.prevent="handleSend"
                />
                <el-button type="primary" :loading="sending" :disabled="!draft.trim()" @click="handleSend">
                  发送
                </el-button>
              </div>
            </div>
          </template>
          <el-empty v-else class="m-auto" :image-size="90" description="从左侧选择一个对话，或从职位/简历发起沟通" />
        </section>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
  import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    deleteChatSession,
    getChatMessages,
    listChatSessions,
    markChatRead,
    openChatSession,
    sendChatMessage
  } from '@/api/chat'
  import type { ChatMessageItem, ChatScope, ChatSessionItem } from '@/api/chat'
  import { useChatStore } from '@/stores/chat'
  import { formatTime } from '@/utils/format'

  const route = useRoute()
  const router = useRouter()
  const chatStore = useChatStore()

  const scope = (route.meta.chatScope as ChatScope) || 'personal'

  const sessions = ref<ChatSessionItem[]>([])
  const messages = ref<ChatMessageItem[]>([])
  const activeId = ref<number | null>(null)
  const draft = ref('')
  const loadingSessions = ref(false)
  const loadingMessages = ref(false)
  const sending = ref(false)
  const scrollRef = ref<HTMLElement | null>(null)

  let pollTimer: number | null = null
  let unsubscribe: (() => void) | null = null

  const activeSession = computed(() => sessions.value.find((s) => s.id === activeId.value) || null)
  const avatarUrl = (url: string) => (url ? (url.startsWith('http') ? url : `/${url}`) : '')

  const scrollToBottom = async () => {
    await nextTick()
    if (scrollRef.value) scrollRef.value.scrollTop = scrollRef.value.scrollHeight
  }

  const loadSessions = async () => {
    loadingSessions.value = true
    try {
      const { data } = await listChatSessions(scope, { page: 1, pageSize: 50 })
      sessions.value = data.list
    } finally {
      loadingSessions.value = false
    }
  }

  const loadMessages = async (sessionId: number) => {
    loadingMessages.value = true
    try {
      const { data } = await getChatMessages(scope, sessionId, { page: 1, pageSize: 50 })
      messages.value = data.list
      await scrollToBottom()
    } finally {
      loadingMessages.value = false
    }
  }

  const selectSession = async (sessionId: number) => {
    activeId.value = sessionId
    router.replace({ query: { ...route.query, sessionId: String(sessionId) } })
    await loadMessages(sessionId)
    await markRead(sessionId)
  }

  const markRead = async (sessionId: number) => {
    try {
      await markChatRead(scope, sessionId)
      const target = sessions.value.find((s) => s.id === sessionId)
      if (target && target.unread > 0) {
        target.unread = 0
      }
      await chatStore.refreshUnread()
    } catch {
      // 忽略
    }
  }

  const handleSend = async () => {
    const content = draft.value.trim()
    if (!content || !activeId.value) return
    sending.value = true
    try {
      const { data } = await sendChatMessage(scope, activeId.value, content)
      messages.value.push(data)
      draft.value = ''
      const target = sessions.value.find((s) => s.id === activeId.value)
      if (target) {
        target.lastContent = content
        target.lastTime = data.addtime
      }
      await scrollToBottom()
    } finally {
      sending.value = false
    }
  }

  const handleDeleteSession = async () => {
    if (!activeId.value) return
    try {
      await ElMessageBox.confirm('确认删除该会话？仅从您的列表中移除，对方仍可见。', '删除会话', {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }
    await deleteChatSession(scope, activeId.value)
    ElMessage.success('已删除')
    activeId.value = null
    await loadSessions()
    await chatStore.refreshUnread()
  }

  // 处理来自其他入口的发起参数：peerUid + jobsId/jobsName 或 sessionId
  const applyRouteQuery = async () => {
    const peerUid = Number(route.query.peerUid)
    if (peerUid > 0) {
      const { data } = await openChatSession(scope, {
        peerUid,
        jobsId: Number(route.query.jobsId) || 0,
        jobsName: (route.query.jobsName as string) || ''
      })
      await loadSessions()
      await selectSession(data.id)
      return
    }
    const sessionId = Number(route.query.sessionId)
    if (sessionId > 0) {
      await selectSession(sessionId)
      return
    }
    if (sessions.value.length > 0) {
      await selectSession(sessions.value[0].id)
    }
  }

  onMounted(async () => {
    await chatStore.refreshUnread()
    await loadSessions()

    // 实时消息：命中当前会话则追加，否则只更新列表
    unsubscribe = chatStore.onMessage(async (msg) => {
      if (msg.sessionId === activeId.value) {
        if (!messages.value.some((m) => m.id === msg.id)) {
          messages.value.push({ ...msg, mine: false })
          await scrollToBottom()
        }
        await markRead(msg.sessionId)
      }
      await loadSessions()
    })

    await applyRouteQuery()

    // 轮询兜底：WebSocket 未连通时同步未读与当前会话
    pollTimer = window.setInterval(async () => {
      if (chatStore.connected) return
      await chatStore.refreshUnread()
      await loadSessions()
      if (activeId.value) {
        const { data } = await getChatMessages(scope, activeId.value, { page: 1, pageSize: 50 })
        const grew = data.list.length !== messages.value.length
        messages.value = data.list
        if (grew) await scrollToBottom()
      }
    }, 5000)
  })

  onUnmounted(() => {
    if (pollTimer !== null) window.clearInterval(pollTimer)
    if (unsubscribe) unsubscribe()
  })
</script>