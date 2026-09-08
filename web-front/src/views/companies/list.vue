<template>
  <div class="mx-auto max-w-7xl px-4 py-8">
    <section class="border-b border-slate-200 pb-6">
      <p class="flex items-center gap-2 text-sm font-medium text-primary-600">
        <span class="i-lucide-building-2" aria-hidden="true" /> 企业招聘库
      </p>
      <h1 class="mt-2 text-2xl font-bold text-slate-800">找企业</h1>
      <p class="mt-2 text-sm text-slate-500">浏览已认证企业与正在招聘的岗位</p>
    </section>

    <el-card shadow="never" class="mt-6">
      <el-form :model="query" @submit.prevent>
        <div class="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-4">
          <el-input v-model="query.keyword" class="lg:col-span-2" placeholder="搜索企业名称或简介" clearable @keyup.enter="handleSearch">
            <template #prefix><span class="i-lucide-search" aria-hidden="true" /></template>
          </el-input>
          <el-select v-model="query.nature" placeholder="企业性质" clearable>
            <el-option v-for="item in filters.nature" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
          <el-select v-model="query.trade" placeholder="所属行业" clearable>
            <el-option v-for="item in filters.trade" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
          <el-select v-model="query.scale" placeholder="企业规模" clearable>
            <el-option v-for="item in filters.scale" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
          <el-input v-model="query.district" placeholder="所在地区" clearable @keyup.enter="handleSearch" />
          <div class="flex gap-2">
            <el-button type="primary" @click="handleSearch"><span class="mr-1 i-lucide-search" aria-hidden="true" />搜索</el-button>
            <el-button @click="handleReset">重置</el-button>
          </div>
        </div>
      </el-form>
    </el-card>

    <section class="mt-6">
      <div class="flex items-center justify-between">
        <h2 class="text-lg font-semibold text-slate-800">认证企业</h2>
        <span v-if="total > 0" class="text-sm text-slate-500">共 {{ total }} 家企业</span>
      </div>
      <div v-loading="loading" class="mt-4">
        <template v-if="loading && companies.length === 0">
          <el-skeleton v-for="item in 6" :key="item" :rows="3" animated class="mb-3" />
        </template>
        <el-empty v-else-if="companies.length === 0" description="暂无符合条件的企业" />
        <div v-else class="grid gap-3 md:grid-cols-2">
          <CompanyCard v-for="company in companies" :key="company.id" :company="company" />
        </div>
      </div>
    </section>

    <div v-if="isGuest" class="mt-7 flex justify-center">
      <el-button type="primary" @click="router.push({ name: 'Login', query: { redirect: route.fullPath } })">登录查看更多企业</el-button>
    </div>
    <div v-else class="mt-7 flex justify-end">
      <el-pagination v-model:current-page="query.page" v-model:page-size="query.pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next" @size-change="loadList" @current-change="loadList" />
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref, watch } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import CompanyCard from '@/components/CompanyCard.vue'
  import { searchCompanies, type PublicCompany } from '@/api/companies'
  import { getCategories } from '@/api/content'
  import { useUserStore } from '@/stores/user'
  import type { Categories } from '@/types/api'

  const GUEST_PREVIEW_SIZE = 4
  const route = useRoute()
  const router = useRouter()
  const userStore = useUserStore()
  const isGuest = computed(() => !userStore.token)
  const loading = ref(false)
  const companies = ref<PublicCompany[]>([])
  const total = ref(0)
  const filters = ref<Categories>({ education: [], experience: [], wage: [], trade: [], district: [], major: [], sex: [], marriage: [], nature: [], scale: [] })
  const query = reactive({ keyword: '', nature: undefined as number | undefined, trade: undefined as number | undefined, scale: undefined as number | undefined, district: '', page: 1, pageSize: 10 })

  const loadList = async () => {
    loading.value = true
    try {
      const { data } = await searchCompanies({
        page: query.page,
        pageSize: isGuest.value ? GUEST_PREVIEW_SIZE : query.pageSize,
        keyword: query.keyword.trim() || undefined,
        nature: query.nature,
        trade: query.trade,
        scale: query.scale,
        district: query.district.trim() || undefined
      })
      companies.value = data.list
      total.value = isGuest.value ? Math.min(data.total, GUEST_PREVIEW_SIZE) : data.total
    } finally {
      loading.value = false
    }
  }

  const loadFilters = async () => {
    try {
      const { data } = await getCategories()
      filters.value = data
    } catch {
      // The directory remains searchable by keyword when filter metadata is unavailable.
    }
  }

  const handleSearch = () => { query.page = 1; loadList() }
  const handleReset = () => {
    query.keyword = ''
    query.nature = undefined
    query.trade = undefined
    query.scale = undefined
    query.district = ''
    query.page = 1
    loadList()
  }

  onMounted(() => {
    if (typeof route.query.keyword === 'string') query.keyword = route.query.keyword
    loadList()
    loadFilters()
  })

  watch(() => userStore.token, () => { query.page = 1; loadList() })
</script>
