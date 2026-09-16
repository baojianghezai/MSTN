<template>
  <div class="personal-shell flex min-h-screen flex-col bg-slate-50">
    <!-- 顶栏（design/10 §3.2.1 ③：与 CompanyLayout 共用同一套顶部导航骨架） -->
    <header class="site-header sticky top-0 z-20 bg-white">
      <div class="site-header__inner mx-auto flex max-w-6xl items-center gap-4 px-4">
        <div class="flex min-w-0 items-center gap-3">
          <router-link to="/" class="site-brand flex items-center gap-2 text-lg font-bold"><span class="site-brand__mark">名</span><span class="hidden sm:inline">名硕人才网</span></router-link>
          <nav class="workspace-nav flex min-w-0 items-center gap-1 overflow-x-auto">
            <router-link
              v-for="item in navItems"
              :key="item.name"
              :to="{ name: item.name }"
              class="workspace-nav-link rounded px-2.5 py-1.5 text-sm transition-colors"
              :class="
                route.name === item.name
                  ? 'bg-primary-50 font-medium text-primary-600'
                  : 'text-slate-600 hover:bg-slate-100'
              "
            >
              <el-badge v-if="item.name === 'PersonalMessages' && messageStore.unread > 0" :value="messageStore.unread" :max="99">
                {{ item.label }}
              </el-badge>
              <span v-else>{{ item.label }}</span>
            </router-link>
          </nav>
        </div>
        <div class="ml-auto flex flex-none items-center gap-3">
          <span class="hidden text-sm text-slate-500 md:inline">uid: {{ userStore.uid }}</span>
          <MessageInbox scope="personal" />
          <el-button @click="handleLogout">退出</el-button>
        </div>
      </div>
    </header>

    <!-- 内容区 -->
    <main class="flex-1">
      <router-view />
    </main>

    <footer class="mt-8 border-t bg-white py-5 text-center text-xs text-slate-400">
      名硕人才网 · 个人中心
    </footer>
  </div>
</template>

<script setup lang="ts">
  import { onMounted } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import MessageInbox from '@/components/message-inbox.vue'
  import { useMessageStore } from '@/stores/message'
  import { useChatStore } from '@/stores/chat'
  import { useUserStore } from '@/stores/user'

  const route = useRoute()
  const router = useRouter()
  const userStore = useUserStore()
  const messageStore = useMessageStore()
  const chatStore = useChatStore()

  // 顶部导航（design/10 §3.2.1 ③）
  const navItems = [
    { name: 'Personal', label: '个人中心' },
    { name: 'PersonalResumes', label: '我的简历' },
    { name: 'PersonalApplies', label: '我的投递' },
    { name: 'PersonalInterviews', label: '面试邀请' },
    { name: 'PersonalChat', label: '在线对话' },
    { name: 'PersonalMessages', label: '站内信' },
    { name: 'PersonalProfile', label: '个人资料' }
  ]

  onMounted(async () => {
    await messageStore.refreshUnread('personal')
    chatStore.connect('personal', userStore.token)
  })

  const handleLogout = () => {
    chatStore.disconnect()
    userStore.logout()
    router.push({ name: 'Home' })
  }
</script>
