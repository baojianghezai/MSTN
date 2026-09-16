<template>
  <div class="company-shell flex min-h-screen flex-col bg-slate-50">
    <!-- 顶栏 -->
    <header class="site-header sticky top-0 z-20 bg-white">
      <div class="site-header__inner mx-auto flex max-w-6xl items-center gap-4 px-4">
        <div class="flex min-w-0 items-center gap-3">
          <router-link to="/" class="site-brand flex items-center gap-2 text-lg font-bold"><span class="site-brand__mark">名</span><span class="hidden sm:inline">名硕人才网</span></router-link>
          <span class="whitespace-nowrap border-l border-slate-200 pl-3 text-sm font-medium text-slate-500">企业中心</span>
        </div>
        <nav class="workspace-nav order-3 flex w-full gap-1 overflow-x-auto border-t border-slate-100 py-2 sm:order-none sm:ml-auto sm:w-auto sm:border-0 sm:py-0">
          <router-link :to="{ name: 'CompanyJobs' }" class="workspace-nav-link rounded px-2.5 py-1.5 text-sm text-slate-600 hover:bg-slate-100">职位管理</router-link>
          <router-link :to="{ name: 'CompanyApplies' }" class="workspace-nav-link rounded px-2.5 py-1.5 text-sm text-slate-600 hover:bg-slate-100">收到简历</router-link>
          <router-link :to="{ name: 'CompanyTalents' }" class="workspace-nav-link rounded px-2.5 py-1.5 text-sm text-slate-600 hover:bg-slate-100">找人才</router-link>
          <router-link :to="{ name: 'CompanyPlan' }" class="workspace-nav-link rounded px-2.5 py-1.5 text-sm text-slate-600 hover:bg-slate-100">套餐订购</router-link>
        </nav>
        <div class="ml-auto flex items-center gap-2 sm:ml-0 sm:gap-3">
          <router-link to="/" class="text-sm text-slate-500 hover:text-primary-600">
            返回首页
          </router-link>
          <el-dropdown trigger="click">
            <el-button circle aria-label="企业功能菜单"><span class="i-lucide-ellipsis" aria-hidden="true" /></el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="router.push({ name: 'CompanyProfile' })">企业资料</el-dropdown-item>
                <el-dropdown-item @click="router.push({ name: 'CompanyPromotions' })">首页推广</el-dropdown-item>
                <el-dropdown-item @click="router.push({ name: 'CompanyTalentLibrary' })">人才库</el-dropdown-item>
                <el-dropdown-item @click="router.push({ name: 'CompanyInterviews' })">面试邀请</el-dropdown-item>
                <el-dropdown-item @click="router.push({ name: 'CompanyChat' })">在线对话</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <MessageInbox scope="company" />
          <el-button @click="handleLogout">退出</el-button>
        </div>
      </div>
    </header>

    <!-- 内容区 -->
    <main class="flex-1">
      <router-view />
    </main>

    <footer class="mt-8 border-t bg-white py-5 text-center text-xs text-slate-400">
      名硕人才网 · 企业中心
    </footer>
  </div>
</template>

<style scoped>
@media (max-width: 639px) {
  .site-header__inner { flex-wrap: wrap; gap: 0.25rem; padding-top: 0.35rem; }
}
</style>

<script setup lang="ts">
  import { onMounted } from 'vue'
  import { useRouter } from 'vue-router'
  import MessageInbox from '@/components/message-inbox.vue'
  import { useMessageStore } from '@/stores/message'
  import { useChatStore } from '@/stores/chat'
  import { useUserStore } from '@/stores/user'

  const router = useRouter()
  const userStore = useUserStore()
  const messageStore = useMessageStore()
  const chatStore = useChatStore()

  onMounted(async () => {
    await messageStore.refreshUnread('company')
    chatStore.connect('company', userStore.token)
  })

  const handleLogout = () => {
    chatStore.disconnect()
    userStore.logout()
    router.push({ name: 'Home' })
  }
</script>
