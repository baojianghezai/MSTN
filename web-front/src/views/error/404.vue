<template>
  <div class="flex flex-col items-center justify-center py-24">
    <p class="text-7xl font-bold text-slate-200">404</p>
    <p class="mt-4 text-slate-500">页面不存在或已被移除</p>
    <div class="mt-8 flex gap-4">
      <el-button type="primary" @click="router.push({ name: 'Home' })">返回首页</el-button>
      <el-button @click="router.push({ name: 'Jobs' })">职位广场</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { useRouter } from 'vue-router'

  const router = useRouter()
</script>
