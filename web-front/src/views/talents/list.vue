<template>
  <div class="mx-auto max-w-7xl px-4 py-8">
    <section class="flex flex-col gap-3 border-b border-slate-200 pb-6 md:flex-row md:items-end md:justify-between">
      <div>
        <p class="flex items-center gap-2 text-sm font-medium text-primary-600">
          <span class="i-lucide-users-round" aria-hidden="true" />
          公开人才市场
        </p>
        <h1 class="mt-2 text-2xl font-bold text-slate-800">找人才</h1>
        <p class="mt-2 text-sm text-slate-500">筛选公开简历，找到匹配岗位需求的候选人</p>
      </div>
      <el-button v-if="userStore.utype === 2" @click="router.push({ name: 'CompanyTalentLibrary' })">
        <span class="mr-1 i-lucide-bookmark-check" aria-hidden="true" />
        人才库
      </el-button>
    </section>

    <el-card shadow="never" class="mt-6">
      <el-form :model="query" @submit.prevent>
        <div class="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-4">
          <el-input
            v-model="query.keyword"
            class="lg:col-span-2"
            placeholder="搜索职位意向、技能或简历标题"
            clearable
            @keyup.enter="handleSearch"
          >
            <template #prefix><span class="i-lucide-search" aria-hidden="true" /></template>
          </el-input>
          <el-select v-model="query.district" placeholder="期望地区" clearable>
            <el-option v-for="item in filters.district" :key="item.id" :label="item.name" :value="String(item.id)" />
          </el-select>
          <el-radio-group v-model="query.talentOnly" class="flex items-center">
            <el-radio-button :value="false">全部人才</el-radio-button>
            <el-radio-button :value="true">高级人才</el-radio-button>
          </el-radio-group>
          <el-select v-model="query.education" placeholder="学历" clearable>
            <el-option v-for="item in filters.education" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
          <el-select v-model="query.experience" placeholder="工作经验" clearable>
            <el-option v-for="item in filters.experience" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
          <el-select v-model="query.wage" placeholder="期望薪资" clearable>
            <el-option v-for="item in filters.wage" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
          <div class="flex gap-2">
            <el-button type="primary" @click="handleSearch">
              <span class="mr-1 i-lucide-search" aria-hidden="true" />
              搜索
            </el-button>
            <el-button @click="handleReset">重置</el-button>
          </div>
        </div>
      </el-form>
    </el-card>

    <section class="mt-6">
      <div class="flex items-center justify-between">
        <h2 class="text-lg font-semibold text-slate-800">候选人简历</h2>
        <span v-if="total > 0" class="text-sm text-slate-500">共 {{ total }} 份公开简历</span>
      </div>

      <div v-loading="loading" class="mt-4">
        <template v-if="loading && resumes.length === 0">
          <div class="grid gap-4 md:grid-cols-2">
            <el-skeleton v-for="item in 6" :key="item" :rows="4" animated />
          </div>
        </template>
        <el-empty v-else-if="resumes.length === 0" description="暂无符合条件的公开简历" />
        <div v-else class="grid gap-4 md:grid-cols-2">
          <TalentCard v-for="resume in resumes" :key="resume.id" :resume="resume" @view="openDetail" />
        </div>
      </div>
    </section>

    <div v-if="isGuest" class="mt-7 flex justify-center">
      <el-button type="primary" @click="router.push({ name: 'Login', query: { redirect: route.fullPath } })">
        登录查看更多人才
      </el-button>
    </div>
    <div v-else class="mt-7 flex justify-end">
      <el-pagination
        v-model:current-page="query.page"
        v-model:page-size="query.pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @size-change="loadList"
        @current-change="loadList"
      />
    </div>

    <el-drawer v-model="detailVisible" title="候选人简历" size="min(520px, 100%)">
      <template v-if="detailLoading">
        <el-skeleton :rows="8" animated />
      </template>
      <template v-else-if="detail">
        <div class="flex items-start gap-3">
          <el-avatar :size="58" :src="detail.resume.photoImg" class="bg-primary-50 text-primary-600">
            <span class="i-lucide-user-round" aria-hidden="true" />
          </el-avatar>
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <h2 class="text-lg font-semibold text-slate-800">{{ detail.resume.fullname }}</h2>
              <el-tag v-if="detail.resume.talent === 1" type="warning" size="small">高级人才</el-tag>
            </div>
            <p class="mt-1 text-sm text-primary-600">{{ detail.resume.intentionJobs || detail.resume.title }}</p>
          </div>
        </div>

        <el-descriptions class="mt-6" :column="1" border>
          <el-descriptions-item label="学历">{{ detail.resume.educationCn || '-' }}</el-descriptions-item>
          <el-descriptions-item label="工作经验">{{ detail.resume.experienceCn || '-' }}</el-descriptions-item>
          <el-descriptions-item label="期望地区">{{ detail.resume.districtCn || '-' }}</el-descriptions-item>
          <el-descriptions-item label="期望薪资">{{ detail.resume.wageCn || '-' }}</el-descriptions-item>
          <el-descriptions-item label="专业">{{ detail.resume.majorCn || '-' }}</el-descriptions-item>
        </el-descriptions>

        <section class="mt-6">
          <h3 class="text-sm font-semibold text-slate-800">技能说明</h3>
          <p class="mt-2 whitespace-pre-line text-sm leading-6 text-slate-600">{{ detail.resume.specialty || '暂未填写' }}</p>
        </section>

        <section v-if="detail.work.length" class="mt-6">
          <h3 class="text-sm font-semibold text-slate-800">工作经历</h3>
          <div v-for="item in detail.work.slice(0, 3)" :key="item.id" class="mt-3 border-l-2 border-primary-200 pl-3">
            <p class="font-medium text-slate-700">{{ item.jobs || '工作经历' }}</p>
            <p class="mt-1 text-sm text-slate-500">{{ item.companyname }}</p>
          </div>
        </section>

        <el-alert class="mt-6" type="info" :closable="false" show-icon title="联系方式已受保护，企业解锁后才能查看。" />
      </template>

      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
        <el-button type="primary" :loading="unlocking" @click="handleContact">
          <span class="mr-1 i-lucide-unlock" aria-hidden="true" />
          {{ userStore.utype === 2 ? '解锁联系方式' : '企业账号联系候选人' }}
        </el-button>
      </template>
    </el-drawer>

    <el-dialog v-model="contactVisible" title="候选人联系方式" width="min(480px, calc(100% - 32px))">
      <el-descriptions v-if="unlockedDetail" :column="1" border>
        <el-descriptions-item label="候选人">{{ unlockedDetail.resume.fullname }}</el-descriptions-item>
        <el-descriptions-item label="联系电话">{{ unlockedDetail.resume.telephone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="邮箱">{{ unlockedDetail.resume.email || '-' }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="contactVisible = false">关闭</el-button>
        <el-button type="primary" @click="router.push({ name: 'CompanyTalentLibrary' })">前往人才库</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref, watch } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { ElMessage } from 'element-plus'
  import TalentCard from '@/components/TalentCard.vue'
  import { getJobFilters } from '@/api/jobs'
  import { getDistricts } from '@/api/content'
  import { getPublicTalent, searchPremiumTalents, searchTalents, unlockTalent } from '@/api/talent'
  import type { PublicResume, PublicResumeDetail, TalentUnlockedDetail } from '@/api/talent'
  import { useUserStore } from '@/stores/user'
  import type { Categories } from '@/types/api'

  const route = useRoute()
  const router = useRouter()
  const userStore = useUserStore()
  const GUEST_PREVIEW_SIZE = 4
  const isGuest = computed(() => !userStore.token)
  const loading = ref(false)
  const resumes = ref<PublicResume[]>([])
  const total = ref(0)
  const detailVisible = ref(false)
  const detailLoading = ref(false)
  const detail = ref<PublicResumeDetail | null>(null)
  const unlocking = ref(false)
  const contactVisible = ref(false)
  const unlockedDetail = ref<TalentUnlockedDetail | null>(null)
  const filters = ref<Categories>({
    education: [],
    experience: [],
    wage: [],
    trade: [],
    district: [],
    major: [],
    sex: [],
    marriage: [],
    nature: [],
    scale: []
  })
  const query = reactive({
    keyword: '',
    district: undefined as string | undefined,
    education: undefined as number | undefined,
    experience: undefined as number | undefined,
    wage: undefined as number | undefined,
    talentOnly: false,
    page: 1,
    pageSize: 10
  })

  const loadList = async () => {
    loading.value = true
    try {
      const params = {
        page: query.page,
        pageSize: isGuest.value ? GUEST_PREVIEW_SIZE : query.pageSize,
        keyword: query.keyword || undefined,
        district: query.district,
        education: query.education,
        experience: query.experience,
        wage: query.wage
      }
      const { data } = query.talentOnly ? await searchPremiumTalents(params) : await searchTalents(params)
      resumes.value = data.list
      total.value = isGuest.value ? Math.min(data.total, GUEST_PREVIEW_SIZE) : data.total
    } finally {
      loading.value = false
    }
  }

  const loadFilters = async () => {
    try {
      const [filtersResponse, districtsResponse] = await Promise.all([getJobFilters(), getDistricts()])
      filters.value = { ...filtersResponse.data, district: districtsResponse.data }
    } catch {
      // 筛选项加载失败时仍可使用关键字和高级人才筛选。
    }
  }

  const handleSearch = () => {
    query.page = 1
    loadList()
  }

  const handleReset = () => {
    query.keyword = ''
    query.district = undefined
    query.education = undefined
    query.experience = undefined
    query.wage = undefined
    query.talentOnly = false
    query.page = 1
    loadList()
  }

  const openDetail = async (resume: PublicResume) => {
    detail.value = null
    detailVisible.value = true
    detailLoading.value = true
    try {
      const { data } = await getPublicTalent(resume.id)
      detail.value = data
    } finally {
      detailLoading.value = false
    }
  }

  const handleContact = async () => {
    if (!detail.value) return
    if (!userStore.token) {
      router.push({ name: 'Login', query: { redirect: route.fullPath } })
      return
    }
    if (userStore.utype !== 2) {
      ElMessage.warning('请使用企业账号解锁候选人联系方式')
      return
    }
    unlocking.value = true
    try {
      const { data } = await unlockTalent(detail.value.resume.id)
      unlockedDetail.value = data.detail
      contactVisible.value = true
      ElMessage.success(data.newlyUnlocked ? '简历已解锁，已扣除一次下载权益' : '该候选人已在人才库中')
    } finally {
      unlocking.value = false
    }
  }

  onMounted(() => {
    if (typeof route.query.keyword === 'string') query.keyword = route.query.keyword
    if (typeof route.query.district === 'string') query.district = route.query.district
    loadList()
    loadFilters()
  })

  watch(
    () => userStore.token,
    () => {
      query.page = 1
      loadList()
    }
  )
</script>
