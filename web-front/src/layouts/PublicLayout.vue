<template>
  <div class="public-shell flex min-h-screen flex-col bg-slate-50">
    <!-- 顶部蓝色导航栏 -->
    <header class="site-topbar sticky top-0 z-30 bg-[#0066ff] text-white">
      <div class="mx-auto flex h-10 max-w-[1200px] items-center gap-4 px-4">
        <div class="flex flex-none items-center gap-1.5 text-sm">
          <span class="i-lucide-map-pin" aria-hidden="true" />
          <el-select
            v-model="city"
            size="small"
            class="!w-24"
            popper-class="site-city-select"
            @change="applyCity"
          >
            <el-option v-for="c in cityOptions" :key="c" :label="c" :value="c" />
          </el-select>
          <button type="button" class="rounded bg-white/15 px-2 py-0.5 text-xs text-white transition-colors hover:bg-white/25" @click="applyCity">
            切换
          </button>
        </div>

        <nav class="hidden flex-1 items-center gap-4 overflow-x-auto text-sm md:flex">
          <router-link
            v-for="nav in navItems"
            :key="nav.id"
            :to="nav.url || '/'"
            class="flex-none py-1 transition-colors"
            :class="isNavActive(nav) ? 'font-semibold text-white underline decoration-white decoration-2 underline-offset-4' : 'text-blue-100 hover:text-white'"
          >
            {{ nav.title }}
          </router-link>
        </nav>

        <div class="ml-auto flex flex-none items-center gap-2 text-sm">
          <router-link to="/company/jobs" class="hidden text-blue-50 transition-colors hover:text-white sm:block">
            我要招人
          </router-link>

          <template v-if="userStore.token">
            <MessageInbox :scope="chatScope" />
            <router-link
              :to="userStore.utype === 2 ? { name: 'Company' } : { name: 'Personal' }"
              class="rounded px-3 py-1 text-blue-50 transition-colors hover:bg-white/10 hover:text-white"
            >
              {{ userStore.utype === 2 ? '企业中心' : '个人中心' }}
            </router-link>
            <button
              type="button"
              class="rounded bg-white/15 px-3 py-1 text-sm text-white transition-colors hover:bg-white/25"
              @click="handleLogout"
            >
              退出
            </button>
          </template>
          <template v-else>
            <router-link to="/login" class="rounded px-3 py-1 text-blue-50 transition-colors hover:bg-white/10 hover:text-white">
              登录
            </router-link>
            <router-link
              to="/register"
              class="rounded bg-white px-3 py-1 font-semibold text-[#0066ff] transition-colors hover:bg-blue-50"
            >
              注册
            </router-link>
          </template>

          <el-popover placement="bottom-end" :width="176" trigger="click" popper-class="public-nav-popover">
            <template #reference>
              <button class="flex h-7 w-7 items-center justify-center rounded md:hidden" type="button" aria-label="打开导航">
                <span class="i-lucide-menu" aria-hidden="true" />
              </button>
            </template>
            <nav class="flex flex-col py-1">
              <router-link
                v-for="nav in navItems"
                :key="nav.id"
                :to="nav.url || '/'"
                class="rounded px-3 py-2.5 text-sm"
                :class="isNavActive(nav) ? 'bg-primary-50 font-semibold text-primary-600' : 'text-slate-600 hover:bg-slate-50'"
              >
                {{ nav.title }}
              </router-link>
            </nav>
          </el-popover>
        </div>
      </div>
    </header>

    <!-- 内容区 -->
    <main class="flex-1">
      <router-view />
    </main>

    <!-- 页脚 -->
    <footer class="mt-10 bg-slate-900 text-slate-400">
      <div class="mx-auto max-w-[1200px] px-4 py-10">
        <div class="grid gap-8 md:grid-cols-[1.3fr_1fr_1fr_1fr]">
          <div>
            <div class="flex items-center gap-2 text-lg font-bold text-white">
              <span class="site-brand__mark">名</span>
              <span>名硕人才网</span>
            </div>
            <p class="mt-3 max-w-xs text-sm leading-6">
              专注真实职位直投、企业招聘与人才服务，让求职更简单，招聘更高效。
            </p>
          </div>

          <div>
            <h3 class="text-sm font-semibold text-white">求职者</h3>
            <ul class="mt-3 space-y-2 text-sm">
              <li><router-link to="/jobs" class="transition-colors hover:text-white">职位搜索</router-link></li>
              <li><router-link to="/companies" class="transition-colors hover:text-white">企业搜索</router-link></li>
              <li><router-link to="/talents" class="transition-colors hover:text-white">人才市场</router-link></li>
              <li><router-link to="/jobfairs" class="transition-colors hover:text-white">招聘会</router-link></li>
              <li><router-link to="/help" class="transition-colors hover:text-white">帮助中心</router-link></li>
            </ul>
          </div>

          <div>
            <h3 class="text-sm font-semibold text-white">企业服务</h3>
            <ul class="mt-3 space-y-2 text-sm">
              <li><router-link to="/company" class="transition-colors hover:text-white">企业中心</router-link></li>
              <li><router-link to="/company/jobs" class="transition-colors hover:text-white">发布职位</router-link></li>
              <li><router-link to="/company/applies" class="transition-colors hover:text-white">收到简历</router-link></li>
              <li><router-link to="/company/promotions" class="transition-colors hover:text-white">首页推广</router-link></li>
            </ul>
          </div>

          <div>
            <h3 class="text-sm font-semibold text-white">关于我们</h3>
            <ul class="mt-3 space-y-2 text-sm">
              <li><router-link to="/appeal" class="transition-colors hover:text-white">账号申诉</router-link></li>
              <li><router-link to="/news" class="transition-colors hover:text-white">资讯</router-link></li>
              <li><router-link to="/help" class="transition-colors hover:text-white">联系我们</router-link></li>
              <li><router-link to="/login" class="transition-colors hover:text-white">会员登录</router-link></li>
            </ul>
          </div>
        </div>

        <div class="mt-8 flex flex-col items-center justify-between gap-3 border-t border-white/10 pt-5 text-xs sm:flex-row">
          <p>© {{ currentYear }} 名硕人才网 · PC 前台</p>
          <div class="flex items-center gap-4">
            <router-link to="/appeal" class="transition-colors hover:text-white">账号申诉</router-link>
            <router-link to="/help" class="transition-colors hover:text-white">帮助中心</router-link>
          </div>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, ref, watch } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import MessageInbox from '@/components/message-inbox.vue'
  import { useUserStore } from '@/stores/user'
  import { useMessageStore } from '@/stores/message'
  import { useChatStore } from '@/stores/chat'
  import { getNavigations } from '@/api/content'
  import type { NavItem } from '@/types/api'

  const route = useRoute()
  const router = useRouter()
  const userStore = useUserStore()
  const messageStore = useMessageStore()
  const chatStore = useChatStore()
  const currentYear = new Date().getFullYear()

  const chatScope = computed<'personal' | 'company'>(() => (userStore.utype === 2 ? 'company' : 'personal'))

  const city = ref('青岛')
  const cityOptions = ['青岛', '北京', '上海', '广州', '深圳', '杭州', '成都', '武汉']

  const DEFAULT_NAVS: NavItem[] = [
    { id: 0, title: '首页', url: '/', sort: 1, display: 1, addTime: '' },
    { id: -1, title: '找工作', url: '/jobs', sort: 2, display: 1, addTime: '' },
    { id: -2, title: '找企业', url: '/companies', sort: 3, display: 1, addTime: '' },
    { id: -3, title: '找人才', url: '/talents', sort: 4, display: 1, addTime: '' },
    { id: -4, title: '招聘会', url: '/jobfairs', sort: 5, display: 1, addTime: '' },
    { id: -5, title: '资讯', url: '/news', sort: 6, display: 1, addTime: '' },
    { id: -6, title: '帮助', url: '/help', sort: 7, display: 1, addTime: '' }
  ]

  const ensureTalentNav = (items: NavItem[]) => {
    const talentNav = items.find((item) => item.url === '/talents') || DEFAULT_NAVS[3]
    const remainingItems = items.filter((item) => item.url !== '/talents')
    const jobsIndex = remainingItems.findIndex((item) => item.url === '/jobs')
    if (jobsIndex < 0) return [...remainingItems, talentNav]
    return [...remainingItems.slice(0, jobsIndex + 1), talentNav, ...remainingItems.slice(jobsIndex + 1)]
  }

  const navItems = ref<NavItem[]>(ensureTalentNav(DEFAULT_NAVS))

  onMounted(async () => {
    try {
      const { data } = await getNavigations()
      if (data.length) navItems.value = ensureTalentNav(data)
    } catch {
      // 拉取失败保留兜底导航
    }
  })

  // 登录态变化时建立/断开在线对话连接，并刷新未读
  watch(
    () => userStore.token,
    async (token) => {
      if (token) {
        chatStore.connect(chatScope.value, token)
        await messageStore.refreshUnread(chatScope.value)
      } else {
        chatStore.disconnect()
      }
    },
    { immediate: true }
  )

  const isNavActive = (nav: NavItem) => {
    const url = nav.url || '/'
    if (url === '/') return route.path === '/'
    return route.path === url || route.path.startsWith(`${url}/`)
  }

  const applyCity = () => {
    router.push({
      path: '/jobs',
      query: city.value ? { district: city.value } : {}
    })
  }

  const handleLogout = () => {
    chatStore.disconnect()
    userStore.logout()
    router.push({ name: 'Home' })
  }
</script>
