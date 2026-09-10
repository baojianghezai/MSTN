<template>
  <div class="home-page min-h-screen bg-gradient-to-b from-blue-50 via-white to-white">
    <!-- 搜索头部区 -->
    <section class="home-search-header">
      <div class="mx-auto flex max-w-[1200px] items-center gap-4 px-4 py-6">
        <router-link to="/" class="home-logo flex-none text-3xl font-black tracking-tight text-[#0066ff]">
          名硕人才网
        </router-link>

        <div class="mx-auto flex w-full max-w-2xl items-center gap-2 rounded-full border border-blue-200 bg-white p-2 shadow-sm px-4">
          <span class="i-lucide-search hidden text-xl text-slate-400 sm:block" aria-hidden="true" />
          <el-input
            v-model="keyword"
            size="large"
            class="flex-1"
            placeholder="搜索职位、公司"
            clearable
            @keyup.enter="goSearch"
          />
          <el-select
            v-model="district"
            size="large"
            class="w-28 flex-none"
            placeholder="城市"
            clearable
          >
            <el-option v-for="d in districts" :key="d.id" :label="d.name" :value="d.name" />
          </el-select>
          <el-button type="primary" size="large" class="flex-none rounded-full px-6 font-semibold" @click="goSearch">
            搜索
          </el-button>
        </div>

        <div class="flex flex-none items-center gap-3">
          <button
            type="button"
            class="flex h-10 items-center gap-1.5 rounded-lg bg-green-500 px-3 text-sm font-medium text-white shadow-sm transition-transform hover:-translate-y-0.5"
            @click="comingSoon('关注公众号')"
          >
            <span class="i-lucide-message-circle" aria-hidden="true" />
            公众号
          </button>
        </div>
      </div>
    </section>

    <!-- 主视觉区：左侧垂直菜单 + 中间 Banner + 右侧登录面板 -->
    <section class="home-main-visual">
      <div class="mx-auto max-w-[1200px] px-4">
        <div class="grid grid-wrapper relative overflow-visible rounded-2xl border border-slate-100 bg-white shadow-[0_18px_50px_rgba(15,23,42,0.08)] lg:h-[400px] lg:grid-cols-[200px_1fr_300px]">
          <aside
            class="hidden bg-slate-800 text-white lg:block lg:min-h-0 relative z-20"
            @mouseleave="onMenuLeave"
          >
            <ul class="flex h-full flex-col justify-start list-none p-0 m-0 overflow-y-auto">
              <li
                v-for="cat in menuCategories"
                :key="cat.id"
                @mouseenter="onMenuEnter(cat)"
              >
                <button
                  type="button"
                  class="flex w-full items-center justify-between gap-2 px-4 py-2.5 text-left text-sm transition-colors"
                  :class="hoveredCategoryId === cat.id ? 'bg-white text-slate-800' : 'bg-slate-600 text-slate-100 hover:bg-slate-700 hover:text-white'"
                  @click="goCategory(cat)"
                >
                  <span class="truncate">{{ cat.name }}</span>
                  <span class="i-lucide-chevron-right flex-none text-slate-400" aria-hidden="true" />
                </button>
              </li>
            </ul>
          </aside>

          <!-- 二三级 flyout 面板（fixed 定位，覆盖 banner 区域） -->
          <Teleport to="body">
            <Transition name="flyout">
              <div
                v-if="hoveredCategory && hoveredCategory.children?.length"
                ref="flyoutRef"
                class="flyout-panel fixed z-[9999] flex h-[400px] flex-col border-l border-slate-200 bg-white shadow-2xl"
                :style="{ top: flyoutPos.top + 'px', left: flyoutPos.left + 'px', width: flyoutPos.width + 'px' }"
                @mouseenter="onFlyoutEnter"
                @mouseleave="onMenuLeave"
              >
                <div class="flex-1 overflow-y-auto p-5 no-scrollbar">
                  <div
                    v-for="sub in hoveredCategory.children"
                    :key="sub.id"
                    class="mb-4 last:mb-0"
                  >
                    <h4 class="mb-1.5 border-b border-slate-100 pb-1.5 text-sm font-bold text-slate-700">
                      {{ sub.name }}
                    </h4>
                    <div class="flex flex-wrap gap-1">
                      <button
                        v-for="child in sub.children"
                        :key="child.id"
                        type="button"
                        class="rounded px-2.5 py-1 text-xs text-slate-500 transition-colors hover:bg-blue-50 hover:text-blue-600"
                        @click="goCategory(child)"
                      >
                        {{ child.name }}
                      </button>
                    </div>
                  </div>
                </div>
                <div class="border-t border-slate-100 px-5 py-2.5 text-right">
                  <button
                    type="button"
                    class="text-xs font-medium text-blue-500 hover:text-blue-600"
                    @click="goCategory(hoveredCategory)"
                  >
                    查看全部「{{ hoveredCategory.name }}」职位 &gt;
                  </button>
                </div>
              </div>
            </Transition>
          </Teleport>

          <div class="min-w-0 lg:h-full">
            <el-carousel
              v-if="heroSlides.length"
              :autoplay="heroSlides.length > 1"
              :interval="5000"
              height="400px"
              class="home-banner-carousel"
            >
              <el-carousel-item v-for="(slide, index) in heroSlides" :key="index">
                <button
                  type="button"
                  class="relative block h-full w-full overflow-hidden text-left text-white"
                  @click="router.push(slide.to)"
                >
                  <img
                    v-if="slide.image"
                    :src="imageUrl(slide.image)"
                    :alt="slide.title"
                    class="absolute inset-0 h-full w-full object-cover"
                  />
                  <div
                    class="absolute inset-0"
                    :class="slide.gradient || 'bg-gradient-to-br from-blue-700 via-blue-600 to-indigo-800'"
                  />
                  <div class="absolute inset-0 flex flex-col justify-center px-8 sm:px-10">
                    <span class="w-fit rounded border border-white/50 px-2 py-1 text-xs">{{ slide.tag }}</span>
                    <h2 class="mt-4 text-3xl font-black leading-tight sm:text-4xl">{{ slide.title }}</h2>
                    <p v-if="slide.subtitle" class="mt-3 text-2xl font-black text-amber-400">{{ slide.subtitle }}</p>
                    <p v-if="slide.aux" class="mt-3 text-sm text-blue-100">{{ slide.aux }}</p>
                    <div v-if="slide.features?.length" class="mt-6 grid max-w-xl grid-cols-2 gap-2 text-xs sm:grid-cols-4">
                      <span
                        v-for="feature in slide.features"
                        :key="feature"
                        class="flex items-center gap-1.5 rounded bg-white/12 px-2 py-1.5"
                      >
                        <span class="i-lucide-badge-check text-amber-300" aria-hidden="true" />
                        {{ feature }}
                      </span>
                    </div>
                  </div>
                  <span class="absolute bottom-3 right-4 text-xs text-white/60">广告</span>
                </button>
              </el-carousel-item>
            </el-carousel>
          </div>

          <aside class="border-t border-slate-100 p-5 lg:h-full lg:min-h-0 lg:overflow-y-auto lg:border-l lg:border-t-0">
            <div v-if="userStore.token" class="flex h-full flex-col justify-center">
              <div class="flex items-center gap-3">
                <el-avatar :size="44" class="bg-primary-50 text-primary-600">
                  <span class="i-lucide-user-round" aria-hidden="true" />
                </el-avatar>
                <div class="min-w-0">
                  <p class="truncate text-base font-semibold text-slate-800">{{ userStore.mobile || '欢迎回来' }}</p>
                  <p class="text-xs text-slate-500">{{ userStore.utype === 2 ? '企业会员' : '个人会员' }}</p>
                </div>
              </div>
              <button
                type="button"
                class="mt-5 w-full rounded-lg bg-slate-600 px-4 py-2 text-sm font-medium text-slate-100 transition-colors hover:bg-slate-700 hover:text-white"
                @click="router.push(userStore.utype === 2 ? { name: 'Company' } : { name: 'Personal' })"
              >
                进入{{ userStore.utype === 2 ? '企业' : '个人' }}中心
              </button>
              <button
                type="button"
                class="mt-2 w-full rounded-lg bg-slate-600 px-4 py-2 text-sm font-medium text-slate-100 transition-colors hover:bg-slate-700 hover:text-white"
                @click="handleLogout"
              >
                退出登录
              </button>
            </div>

            <template v-else>
              <div class="flex items-center justify-between">
                <h3 class="my-0 text-lg font-bold text-slate-800">验证码登录/注册</h3>
                <router-link to="/company/jobs" class="flex-none text-sm text-primary-600 hover:underline">
                  我要招人 &gt;
                </router-link>
              </div>

              <el-form :model="smsForm" label-position="top" class="home-login-form mt-3" @submit.prevent>
                <el-form-item label="账号类型">
                  <el-radio-group v-model="smsForm.utype" size="small">
                    <el-radio-button :value="1">个人</el-radio-button>
                    <el-radio-button :value="2">企业</el-radio-button>
                  </el-radio-group>
                </el-form-item>
                <el-form-item label="手机号">
                  <el-input v-model="smsForm.mobile" maxlength="11" placeholder="手机号">
                    <template #prepend>+86</template>
                  </el-input>
                </el-form-item>
                <el-form-item label="验证码">
                  <div class="flex w-full gap-2">
                    <el-input v-model="smsForm.code" maxlength="6" placeholder="短信验证码" />
                    <el-button
                      class="flex-none"
                      :disabled="sending || countdown > 0"
                      @click="handleSendSms"
                    >
                      {{ countdown > 0 ? `${countdown}s 后重发` : '获取验证码' }}
                    </el-button>
                  </div>
                </el-form-item>
                <el-button
                  type="primary"
                  size="large"
                  class="mt-2 w-full"
                  :loading="submitting"
                  @click="handleSmsLogin"
                >
                  登录 / 注册
                </el-button>
              </el-form>

              <p class="mt-4 text-xs leading-5 text-slate-400">
                已阅读并同意《用户服务协议》和《隐私政策》
              </p>
            </template>
          </aside>
        </div>
      </div>
    </section>

    <!-- 下方企业展示区 -->
    <section class="home-showcase bg-white py-10">
      <div class="mx-auto max-w-[1200px] px-4">
        <h2 class="text-center text-2xl font-black text-slate-800">本季热招企业</h2>
        <div class="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <button
            v-for="item in showcaseCompanies"
            :key="item.id"
            type="button"
            class="group overflow-hidden rounded-xl text-left text-white shadow-card transition-all duration-200 hover:-translate-y-1 hover:shadow-card-hover"
            @click="router.push(item.to)"
          >
            <div class="h-32 p-5" :class="item.gradient">
              <p class="text-lg font-bold leading-snug">{{ item.title }}</p>
              <p class="mt-2 text-sm text-white/85">{{ item.subtitle }}</p>
            </div>
            <div class="flex items-center justify-between bg-white px-5 py-2 text-sm text-primary-600">
              <span>查看详情</span>
              <span class="i-lucide-arrow-right transition-transform group-hover:translate-x-0.5" aria-hidden="true" />
            </div>
          </button>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
  import { computed, nextTick, onMounted, reactive, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { ElMessage } from 'element-plus'
  import { getJobList } from '@/api/category'
  import { getDistricts } from '@/api/content'
  import { getHomePromotions, type HomeAdItem } from '@/api/promotion'
  import { searchCompanies, type PublicCompany } from '@/api/companies'
  import { loginBySms } from '@/api/auth'
  import { useUserStore } from '@/stores/user'
  import { useSmsCode } from '@/composables/useSmsCode'
  import { CategoryTreeNode } from '@/types/api'
  import type { CategoryItem } from '@/types/api'

  interface HeroSlide {
    tag: string
    title: string
    subtitle: string
    aux: string
    image?: string
    gradient: string
    to: string
    features?: string[]
  }

  interface ShowcaseItem {
    id: number | string
    title: string
    subtitle: string
    gradient: string
    to: string
  }

  const router = useRouter()
  const userStore = useUserStore()
  const { sending, countdown, send } = useSmsCode()

  const keyword = ref('')
  const district = ref('')
  const districts = ref<CategoryItem[]>([])
  const menuCategories = ref<CategoryTreeNode[]>([])
  const hoveredCategoryId = ref<number | null>(null)
  const flyoutPos = ref({ top: 0, left: 0, width: 520 })
  const flyoutRef = ref<HTMLElement | null>(null)
  const hoveredCategory = computed(() => {
    if (hoveredCategoryId.value === null) return null
    return menuCategories.value.find(c => c.id === hoveredCategoryId.value) || null
  })
  let menuLeaveTimer: ReturnType<typeof setTimeout> | null = null

  const onMenuEnter = (cat: CategoryTreeNode) => {
    if (menuLeaveTimer) { clearTimeout(menuLeaveTimer); menuLeaveTimer = null }
    hoveredCategoryId.value = cat.id
    // flyout 固定在 banner 区域：取 grid-wrapper 中间列的位置
    nextTick(() => {
      const grid = document.querySelector('.grid-wrapper') as HTMLElement
      if (!grid) return
      const gridRect = grid.getBoundingClientRect()
      flyoutPos.value = { top: gridRect.top, left: gridRect.left + 200, width: Math.max(gridRect.width - 200 - 300, 520) }
    })
  }

  const onFlyoutEnter = () => {
    if (menuLeaveTimer) { clearTimeout(menuLeaveTimer); menuLeaveTimer = null }
  }

  const onMenuLeave = () => {
    menuLeaveTimer = setTimeout(() => { hoveredCategoryId.value = null }, 200)
  }

  const ads = ref<HomeAdItem[]>([])
  const companies = ref<PublicCompany[]>([])
  const submitting = ref(false)

  const smsForm = reactive({ mobile: '', code: '', utype: 1 })

  // 左侧垂直菜单固定取 9 条（首页描述.md §4）：主视觉区高度锁定 400px，
  // 而 /jobs/filters 的一级职位类目实际有 20 条，全量渲染会把网格行撑到 1100px+ 、
  // 顶掉 Banner 与登录面板的对齐，故只展示前 9 条，其余类目由「找工作」页承载
  const MENU_LIMIT = 9

  const FALLBACK_MENU = [
    "销售",
    "人事/行政/党群",
    "财务/法务",
    "技术",
    "电子/通信/半导体",
    "产品",
    "设计",
    "游戏",
    "运营/客服",
    "市场/公关/广告",
    "项目管理",
    "高级管理",
    "房地产/建筑",
    "金融",
    "采购/贸易",
    "物流/仓储/司机",
    "汽车",
    "普工/技工",
    "生产制造",
    "能源/环保",
    "农/林/牧/渔",
    "医疗健康",
    "教育培训",
    "直播/影视/传媒",
    "咨询/翻译/法律",
    "生活服务",
    "餐饮",
    "管培生/非企业从业者"
  ].map((name, index) => ({ id: -index - 1, name }))

  const DEFAULT_SLIDES: HeroSlide[] = [
    {
      tag: '名硕人才网 · 产业升级风口技能专题',
      title: '本地好岗直招',
      subtitle: '月薪 8000 起',
      aux: '技能变现，从现在开始！',
      features: ['名企直招 高薪热岗', '技能认证 提升竞争力', '职业发展 前景广阔', '安全保障 安心求职'],
      gradient: 'bg-gradient-to-br from-blue-700 via-blue-600 to-indigo-800',
      to: '/jobs'
    },
    {
      tag: '名硕人才网 · 名企热招',
      title: '名企直招 高薪热岗',
      subtitle: '五险一金 福利齐全',
      aux: '好工作不等待',
      features: ['真实职位', '企业直招', '高效匹配', '隐私保护'],
      gradient: 'bg-gradient-to-br from-cyan-700 via-blue-600 to-blue-800',
      to: '/companies'
    }
  ]

  const heroSlides = computed<HeroSlide[]>(() => {
    if (!ads.value.length) return DEFAULT_SLIDES
    return ads.value.map((ad) => ({
      tag: '企业招聘广告',
      title: ad.adTitle,
      subtitle: ad.adSubtitle,
      aux: '查看职位',
      image: ad.adImage,
      gradient: 'bg-gradient-to-br from-blue-700 via-blue-600 to-indigo-800',
      to: `/jobs/${ad.id}`
    }))
  })

  const gradients = [
    'bg-gradient-to-br from-emerald-500 to-cyan-600',
    'bg-gradient-to-br from-rose-500 to-red-600',
    'bg-gradient-to-br from-blue-500 to-sky-500'
  ]

  const FALLBACK_SHOWCASE: ShowcaseItem[] = [
    { id: 'companies', title: '认证企业入驻', subtitle: '真实资质审核 · 安心求职', gradient: gradients[0], to: '/companies' },
    { id: 'hot-jobs', title: '名企热招专场', subtitle: '高薪岗位每日更新', gradient: gradients[1], to: '/jobs' },
    { id: 'jobfairs', title: '招聘会 / 校园招聘', subtitle: '线上线下招聘会预告', gradient: gradients[2], to: '/jobfairs' }
  ]

  const showcaseCompanies = computed<ShowcaseItem[]>(() => {
    if (!companies.value.length) return FALLBACK_SHOWCASE
    const items = companies.value.slice(0, 3).map((company, index) => ({
      id: company.id,
      title: company.companyname,
      subtitle: `${company.tradeCn || '认证企业'} · ${company.jobsCount} 个在招职位`,
      gradient: gradients[index % gradients.length],
      to: `/companies/${company.id}`
    }))
    return [...items, ...FALLBACK_SHOWCASE.slice(items.length, 3)]
  })

  const imageUrl = (url: string) => {
    if (/^https?:\/\//i.test(url)) return url
    return `/${url.replace(/^\/+/, '')}`
  }

  const goSearch = () => {
    const query: Record<string, string> = {}
    if (keyword.value) query.keyword = keyword.value
    if (district.value) query.district = district.value
    router.push({ path: '/jobs', query })
  }

  const goCategory = (cat: { id: number; name: string }) => {
    router.push({ path: '/jobs', query: { keyword: cat.name } })
  }

  const comingSoon = (label: string) => {
    ElMessage.info(`${label}功能建设中`)
  }

  const handleSendSms = () => {
    send(smsForm.mobile, 'login')
  }

  const handleSmsLogin = async () => {
    if (!smsForm.mobile || !smsForm.code) {
      ElMessage.warning('请填写手机号和验证码')
      return
    }
    submitting.value = true
    try {
      const { data } = await loginBySms(smsForm.mobile, smsForm.code, smsForm.utype)
      userStore.setLogin(data.token, data.uid, data.utype, data.mobile)

      if (!data.passwordSet) {
        ElMessage.warning('该账号尚未设置密码，请先设置密码')
        router.push({ name: 'ForgotPassword', query: { mobile: data.mobile } })
        return
      }

      ElMessage.success('登录成功')
      router.push(data.utype === 2 ? { name: 'Company' } : { name: 'Home' })
    } finally {
      submitting.value = false
    }
  }

  const handleLogout = () => {
    userStore.logout()
    router.push({ name: 'Home' })
  }

  onMounted(() => {
    getJobList()
      .then(({data}) => {
        const topLevel = data.filter(item => item.parentId === 0)
        menuCategories.value = topLevel
            .sort((a, b) => ( a.id - b.id ))
      })
        .catch(() => {
          // 出错时用 fallback（注意 fallback 数据需要符合树形结构，这里简化用空数组）
          menuCategories.value = []
        })

    getDistricts()
      .then(({ data }) => {
        districts.value = data
      })
      .catch(() => {})

    getHomePromotions()
      .then(({ data }) => {
        ads.value = data.ads
      })
      .catch(() => {})

    searchCompanies({ page: 1, pageSize: 3 })
      .then(({ data }) => {
        companies.value = data.list
      })
      .catch(() => {})
  })
</script>

<style scoped>
/* 隐藏滚动条（WebKit + Firefox） */
.no-scrollbar::-webkit-scrollbar { width: 0; height: 0; background: transparent; }
.no-scrollbar { scrollbar-width: none; -ms-overflow-style: none; }
ul.list-none::-webkit-scrollbar { width: 0; height: 0; background: transparent; }

/* flyout 过渡动画 */
.flyout-enter-active,
.flyout-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}
.flyout-enter-from,
.flyout-leave-to {
  opacity: 0;
  transform: translateX(-6px);
}

.home-login-form :deep(.el-form-item) { margin-bottom: 10px; }
.home-login-form :deep(.el-form-item__label) { margin-bottom: 0; padding-bottom: 2px; line-height: 18px; }

@media (max-width: 767px) {
  .home-banner-carousel :deep(.el-carousel__container) { height: 320px !important; }
  .home-main-visual .grid { border-radius: 0; }
}
</style>

<style>
/* flyout panel 不受 scoped 限制（Teleport 到 body） */
.flyout-panel {
  max-height: 400px;
}
</style>
