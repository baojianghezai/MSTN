<template>
  <el-popover placement="bottom-end" :width="360" trigger="click" @show="handleShow">
    <template #reference>
      <el-badge :value="totalUnread" :max="99" :hidden="totalUnread === 0">
        <el-button circle aria-label="消息中心">
          <span class="i-lucide-bell" aria-hidden="true" />
        </el-button>
      </el-badge>
    </template>

    <div>
      <el-tabs v-model="activeTab">
        <el-tab-pane :label="`站内信${messageStore.unread ? ` (${messageStore.unread})` : ''}`" name="system" />
        <el-tab-pane :label="`在线对话${chatStore.unread ? ` (${chatStore.unread})` : ''}`" name="chat" />
      </el-tabs>

      <!-- 站内信 -->
      <div v-if="activeTab === 'system'">
        <el-skeleton v-if="loadingSystem" :rows="3" animated />
        <el-empty v-else-if="systemList.length === 0" :image-size="60" description="暂无站内信" />
        <ul v-else class="m-0 list-none p-0">
          <li
            v-for="item in systemList"
            :key="item.id"
            class="cursor-pointer rounded px-2 py-2 hover:bg-slate-50"
            @click="openSystemMessage(item)"
          >
            <div class="flex items-center justify-between gap-2">
              <span class="truncate text-sm" :class="item.msgCheck === 0 ? 'font-semibold text-slate-800' : 'text-slate-600'">
                {{ item.title }}
              </span>
              <span class="flex-none text-xs text-slate-400">{{ formatTime(item.addtime) }}</span>
            </div>
            <p class="mt-0.5 line-clamp-2 text-xs text-slate-500">{{ item.message }}</p>
          </li>
        </ul>
        <div class="mt-2 flex justify-end border-t border-slate-100 pt-2">
          <el-button link type="primary" @click="goMessages">查看全部站内信</el-button>
        </div>
      </div>

      <!-- 在线对话 -->
      <div v-else>
        <el-skeleton v-if="loadingChat" :rows="3" animated />
        <el-empty v-else-if="chatList.length === 0" :image-size="60" description="暂无对话" />
        <ul v-else class="m-0 list-none p-0">
          <li
            v-for="item in chatList"
            :key="item.id"
            class="flex cursor-pointer items-center gap-3 rounded px-2 py-2 hover:bg-slate-50"
            @click="openChat(item)"
          >
            <el-avatar :size="34" :src="avatarUrl(item.peerLogo)" class="bg-primary-50 text-primary-600">
              <span class="i-lucide-user-round" aria-hidden="true" />
            </el-avatar>
            <div class="min-w-0 flex-1">
              <div class="flex items-center justify-between gap-2">
                <span class="truncate text-sm font-medium text-slate-700">{{ item.peerName }}</span>
                <span class="flex-none text-xs text-slate-400">{{ formatTime(item.lastTime || item.updateTime) }}</span>
              </div>
              <p class="mt-0.5 truncate text-xs text-slate-500">{{ item.lastContent || '开始对话吧' }}</p>
            </div>
            <el-badge v-if="item.unread > 0" :value="item.unread" :max="99" />
          </li>
        </ul>
        <div class="mt-2 flex justify-end border-t border-slate-100 pt-2">
          <el-button link type="primary" @click="goChat">进入在线对话</el-button>
        </div>
      </div>
    </div>
  </el-popover>
</template>

<script setup lang="ts">
  import { computed, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { getMessages, markMessagesRead } from '@/api/message'
  import type { MessageItem } from '@/api/message'
  import { listChatSessions } from '@/api/chat'
  import type { ChatSessionItem, ChatScope } from '@/api/chat'
  import { useMessageStore } from '@/stores/message'
  import { useChatStore } from '@/stores/chat'
  import { formatTime } from '@/utils/format'

  const props = defineProps<{ scope: ChatScope }>()
  const router = useRouter()
  const messageStore = useMessageStore()
  const chatStore = useChatStore()

  const activeTab = ref<'system' | 'chat'>('system')
  const systemList = ref<MessageItem[]>([])
  const chatList = ref<ChatSessionItem[]>([])
  const loadingSystem = ref(false)
  const loadingChat = ref(false)

  const totalUnread = computed(() => messageStore.unread + chatStore.unread)

  const messagesRoute = computed(() => (props.scope === 'personal' ? 'PersonalMessages' : 'CompanyMessages'))
  const chatRoute = computed(() => (props.scope === 'personal' ? 'PersonalChat' : 'CompanyChat'))

  const avatarUrl = (url: string) => (url ? (url.startsWith('http') ? url : `/${url}`) : '')

  const handleShow = async () => {
    loadingSystem.value = true
    loadingChat.value = true
    await Promise.all([
      messageStore.refreshUnread(props.scope),
      chatStore.refreshUnread()
    ])
    try {
      const { data } = await getMessages(props.scope, { page: 1, pageSize: 5 })
      systemList.value = data.list
    } finally {
      loadingSystem.value = false
    }
    try {
      const { data } = await listChatSessions(props.scope, { page: 1, pageSize: 5 })
      chatList.value = data.list
    } finally {
      loadingChat.value = false
    }
  }

  const openSystemMessage = async (item: MessageItem) => {
    if (item.msgCheck === 0) {
      await markMessagesRead(props.scope, [item.id])
      item.msgCheck = 1
      await messageStore.refreshUnread(props.scope)
    }
    if (item.link && item.link.startsWith('/')) {
      router.push(item.link)
    } else {
      goMessages()
    }
  }

  const openChat = (item: ChatSessionItem) => {
    router.push({ name: chatRoute.value, query: { sessionId: String(item.id) } })
  }

  const goMessages = () => router.push({ name: messagesRoute.value })
  const goChat = () => router.push({ name: chatRoute.value })
</script>