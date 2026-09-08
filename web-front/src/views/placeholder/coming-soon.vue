<template>
  <div class="flex flex-col items-center justify-center py-24">
    <el-empty :description="`${title}功能建设中，敬请期待`" />
    <el-button class="mt-4" type="primary" plain @click="router.push({ name: 'Home' })">
      返回首页
    </el-button>
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import { useRoute, useRouter } from 'vue-router'

  // 占位页：导航已挂出但后端接口尚未实现的入口（design/10 §3.3）
  // 页面名取路由 meta.title，后续接口就绪后替换本组件即可
  const route = useRoute()
  const router = useRouter()

  const title = computed(() => (route.meta.title as string) || '该页面')
</script>
