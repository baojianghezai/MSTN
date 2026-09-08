<template>
  <div class="h-full gva-container2 overflow-auto bg-main">
    <div class="space-y-2 py-2">
      <gva-card
        class="relative overflow-hidden rounded-xl border border-slate-200/80 bg-white px-5 py-6 shadow-sm dark:border-slate-700 dark:bg-slate-900"
      >
        
        <div class="relative flex flex-col gap-2 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <p class="text-xs tracking-[0.2em] text-muted-foreground">DASHBOARD</p>
            <h1 class="mt-2 text-xl font-semibold text-base-text lg:text-2xl">
              欢迎回来，开始今天的Coding节奏
            </h1>
            <p class="mt-2 text-sm text-muted-foreground">
              {{ today }} · 已为你聚合核心业务数据、插件动态和系统公告
            </p>
          </div>
        </div>
      </gva-card>

      <div class="grid grid-cols-1 gap-2 sm:grid-cols-2 xl:grid-cols-3">
        <gva-card>
          <gva-chart :type="1" title="访问人数" />
        </gva-card>
        <gva-card>
          <gva-chart :type="2" title="新增客户" />
        </gva-card>
        <gva-card>
          <gva-chart :type="3" title="解决数量" />
        </gva-card>
      </div>

      <div class="grid grid-cols-1 items-stretch gap-2 xl:grid-cols-12">
        <div class="grid grid-cols-1 gap-2 content-start xl:col-span-8 xl:h-full">
          <gva-card title="内容数据">
            <gva-chart :type="4" />
          </gva-card>

          <gva-card title="最新插件">
            <gva-plugin-table />
          </gva-card>

          <gva-card title="最新更新">
            <gva-table />
          </gva-card>
        </div>

        <div class="flex flex-col gap-2 xl:col-span-4 xl:h-full">
          <gva-card title="快捷功能" show-action custom-class="min-h-[300px]">
            <gva-quick-link />
          </gva-card>
          <gva-card title="公告" show-action custom-class="min-h-[300px]">
            <gva-notice />
          </gva-card>
          <gva-card title="文档" show-action custom-class="min-h-[120px]">
            <gva-wiki />
          </gva-card>
          <div
            class="relative min-h-[200px] flex-1 overflow-hidden rounded-lg border border-slate-200 bg-slate-900 p-5 text-white shadow-sm dark:border-slate-700"
          >
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { computed } from 'vue'
  import {
    GvaPluginTable,
    GvaTable,
    GvaChart,
    GvaWiki,
    GvaNotice,
    GvaQuickLink,
    GvaCard
  } from './components'

  const today = computed(() => {
    try {
      const d = new Date()
      return d.toLocaleDateString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit'
      })
    } catch (e) {
      return new Date().toISOString().slice(0, 10)
    }
  })
  defineOptions({
    name: 'Dashboard'
  })
</script>

<style lang="scss" scoped></style>

