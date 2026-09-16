<template>
  <div class="mx-auto max-w-3xl px-4 py-10">
    <el-button link class="mb-4" @click="router.back()">← 返回</el-button>

    <div v-loading="loading">
      <template v-if="article">
        <h1 class="text-2xl font-bold text-slate-800">{{ article.title }}</h1>
        <div class="mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-slate-400">
          <span v-if="article.source">来源：{{ article.source }}</span>
          <span>{{ formatDate(article.addtime) }}</span>
          <span>阅读 {{ article.click }}</span>
        </div>

        <!-- 招聘会信息 -->
        <el-descriptions v-if="article.type === 2" class="mt-5" :column="1" border>
          <el-descriptions-item label="举办时间">{{ article.holdTime || '-' }}</el-descriptions-item>
          <el-descriptions-item label="举办地点">{{ article.address || '-' }}</el-descriptions-item>
          <el-descriptions-item label="主办方">{{ article.organizer || '-' }}</el-descriptions-item>
        </el-descriptions>

        <div class="mt-5 whitespace-pre-wrap text-sm leading-7 text-slate-700">{{ article.content }}</div>

        <!-- 帮助中心：公众号二维码（#20） -->
        <el-card v-if="article.type === 3" shadow="never" class="mt-8 max-w-sm">
          <div class="flex items-center gap-4">
            <img
              v-if="qrOk"
              src="/mp-qrcode.png"
              alt="名硕人才网公众号二维码"
              class="h-24 w-24 rounded object-contain"
              @error="qrOk = false"
            />
            <div v-else class="flex h-24 w-24 items-center justify-center rounded bg-slate-100 text-xs text-slate-400">
              公众号二维码
            </div>
            <div class="text-sm text-slate-600">
              <p class="font-medium">还有疑问？</p>
              <p class="mt-1 text-xs text-slate-400">扫码关注公众号，获取更多帮助</p>
            </div>
          </div>
        </el-card>
      </template>
      <el-empty v-else-if="!loading" description="内容不存在或已下架" />
    </div>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, ref, watch } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { getArticle } from '@/api/content'
  import type { ArticleItem } from '@/types/api'
  import { formatDate } from '@/utils/format'

  const route = useRoute()
  const router = useRouter()

  const article = ref<ArticleItem | null>(null)
  const loading = ref(false)
  const qrOk = ref(true)

  const load = async () => {
    const id = Number(route.params.id)
    if (!id) return
    loading.value = true
    try {
      const { data } = await getArticle(id)
      article.value = data
    } catch {
      article.value = null
    } finally {
      loading.value = false
    }
  }

  onMounted(load)
  watch(() => route.params.id, load)
</script>