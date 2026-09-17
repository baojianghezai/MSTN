<template>
  <div class="mx-auto max-w-5xl px-4 py-10">
    <header class="border-b border-slate-200 pb-5">
      <h1 class="text-2xl font-bold text-slate-800">{{ title }}</h1>
      <p class="mt-2 text-sm text-slate-500">{{ subtitle }}</p>
    </header>

    <div v-loading="loading" class="mt-6 space-y-4">
      <el-skeleton v-if="loading && list.length === 0" :rows="5" animated />
      <el-empty v-else-if="list.length === 0" :image-size="90" description="暂无内容" />
      <el-card
        v-for="item in list"
        v-else
        :key="item.id"
        shadow="hover"
        class="cursor-pointer transition-colors"
        @click="openDetail(item.id)"
      >
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <h2 class="text-base font-semibold text-slate-800">{{ item.title }}</h2>
            <p class="mt-1 line-clamp-2 text-sm text-slate-500">{{ item.summary || '点击查看详情' }}</p>
            <div v-if="item.holdTime || item.address" class="mt-2 flex flex-wrap gap-x-4 text-xs text-primary-600">
              <span v-if="item.holdTime">时间：{{ item.holdTime }}</span>
              <span v-if="item.address">地点：{{ item.address }}</span>
            </div>
          </div>
          <span class="flex-none text-xs text-slate-400">{{ formatDate(item.addtime) }}</span>
        </div>
        <div v-if="type === 2" class="mt-3 flex items-center justify-between border-t border-slate-100 pt-3">
          <span class="text-xs text-slate-400">已报名 {{ item.signupCount || 0 }} 人</span>
          <el-button
            size="small"
            :type="signedIds.includes(item.id) ? 'success' : 'primary'"
            plain
            @click.stop="toggleSignup(item)"
          >
            {{ signedIds.includes(item.id) ? '已报名（点击取消）' : '参加' }}
          </el-button>
        </div>
      </el-card>
    </div>

    <div class="mt-6 flex justify-end">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @size-change="load"
        @current-change="load"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, ref, watch } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { ElMessage } from 'element-plus'
  import { cancelJobfairSignup, getArticles, getMyJobfairSignups, signupJobfair } from '@/api/content'
  import type { ArticleItem } from '@/types/api'
  import { formatDate } from '@/utils/format'
  import { useUserStore } from '@/stores/user'

  const route = useRoute()
  const router = useRouter()
  const userStore = useUserStore()

  const type = computed(() => Number(route.meta.contentType) || 1)
  const title = computed(() => (route.meta.contentTitle as string) || '资讯')
  const detailName = computed(() => (route.meta.detailName as string) || 'NewsDetail')
  const subtitle = computed(() => {
    if (type.value === 2) return '线上线下招聘会，欢迎企业与求职者参与'
    if (type.value === 3) return '常见问题与使用帮助'
    return '行业动态、求职招聘资讯'
  })

  const list = ref<ArticleItem[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)
  const loading = ref(false)
  const signedIds = ref<number[]>([])

  const loadSigned = async () => {
    if (!userStore.token || userStore.utype !== 1) {
      signedIds.value = []
      return
    }
    try {
      const { data } = await getMyJobfairSignups()
      signedIds.value = data
    } catch {
      signedIds.value = []
    }
  }

  const toggleSignup = async (item: ArticleItem) => {
    if (!userStore.token) {
      router.push({ name: 'Login', query: { redirect: route.fullPath } })
      return
    }
    if (userStore.utype !== 1) {
      ElMessage.warning('请使用个人账号报名招聘会')
      return
    }
    if (signedIds.value.includes(item.id)) {
      await cancelJobfairSignup(item.id)
      signedIds.value = signedIds.value.filter((id) => id !== item.id)
      item.signupCount = Math.max(0, (item.signupCount || 0) - 1)
      ElMessage.success('已取消报名')
    } else {
      await signupJobfair(item.id)
      signedIds.value = [...signedIds.value, item.id]
      item.signupCount = (item.signupCount || 0) + 1
      ElMessage.success('报名成功')
    }
  }

  const load = async () => {
    loading.value = true
    try {
      const { data } = await getArticles({ type: type.value, page: page.value, pageSize: pageSize.value })
      list.value = data.list
      total.value = data.total
    } finally {
      loading.value = false
    }
  }

  const openDetail = (id: number) => {
    router.push({ name: detailName.value, params: { id } })
  }

  onMounted(() => {
    load()
    loadSigned()
  })
  watch(
    () => route.fullPath,
    () => {
      page.value = 1
      load()
    }
  )
</script>