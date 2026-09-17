<template>
  <!-- 工具按钮组：扁平图标，无独立描边/阴影，hover 时图标 morph 为文案（取代 tooltip）。 -->
  <div class="flex items-center gap-1 mx-3">
    <g-dropdown-menu
      v-if="isDev"
      trigger="hover"
      :items="videoList"
      @select="(item) => toDoc(item.value)"
    >
      <icon-button icon="lucide:clapperboard" label="教程" />
    </g-dropdown-menu>

    <icon-button
      v-if="settings.header.search.visible"
      icon="lucide:search"
      label="搜索"
      @click="handleCommand"
    />

    <!-- 待办红点（#3）：职位审核 / 企业审核 / 待处理订单 / 推广审核 -->
    <el-popover placement="bottom-end" :width="240" trigger="hover">
      <template #reference>
        <el-badge :value="pendingTotal" :max="99" :hidden="pendingTotal === 0" class="mx-1">
          <icon-button icon="lucide:bell" label="待办" @click="goPending('companyAudit')" />
        </el-badge>
      </template>
      <ul class="m-0 list-none p-0">
        <li
          v-for="item in pendingItems"
          :key="item.name"
          class="flex cursor-pointer items-center justify-between rounded px-2 py-1.5 hover:bg-slate-50"
          @click="goPending(item.name)"
        >
          <span class="text-sm text-slate-600">{{ item.label }}</span>
          <span class="text-sm font-semibold" :class="item.count > 0 ? 'text-red-500' : 'text-slate-300'">
            {{ item.count }}
          </span>
        </li>
      </ul>
    </el-popover>

    <icon-button
      icon="lucide:settings"
      label="设置"
      @click="toggleSetting"
    />

    <icon-button
      v-if="settings.header.refresh.visible"
      icon="lucide:refresh-cw"
      label="刷新"
      :icon-class="showRefreshAnmite ? 'animate-spin' : ''"
      @click="toggleRefresh"
    />

    <icon-button
      :icon="themeStore.isDark ? 'lucide:sun' : 'lucide:moon'"
      label="主题"
      @click="themeStore.toggleTheme(!themeStore.isDark)"
    />

    <gva-setting v-model:drawer="showSettingDrawer"></gva-setting>
    <command-menu ref="command" />
  </div>
</template>

<script setup>
  import { useThemeStore } from '@/pinia'
  import { storeToRefs } from 'pinia'
  import GvaSetting from '@/view/layout/setting/index.vue'
  import { computed, onMounted, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { emitter } from '@/utils/bus.js'
  import CommandMenu from '@/components/commandMenu/index.vue'
  import IconButton from '@/components/iconButton/index.vue'
  import { toDoc } from '@/utils/doc'
  import { isDev } from '@/utils/env.js'
  import { getPendingCounts } from '@/api/hrc/pending'

  const themeStore = useThemeStore()
  const { settings } = storeToRefs(themeStore)
  const router = useRouter()

  // 待办统计
  const pending = ref({ jobAudit: 0, companyAudit: 0, orderPending: 0, promotionAudit: 0, total: 0 })
  const pendingItems = computed(() => [
    { name: 'jobsManage', label: '职位待审核', count: pending.value.jobAudit },
    { name: 'companyAudit', label: '企业待审核', count: pending.value.companyAudit },
    { name: 'billingManage', label: '待处理订单', count: pending.value.orderPending },
    { name: 'promotionAudit', label: '推广待审核', count: pending.value.promotionAudit }
  ])
  const pendingTotal = computed(() => pending.value.total || 0)

  const goPending = (name) => {
    if (name) router.push({ name })
  }

  const loadPending = async () => {
    try {
      const res = await getPendingCounts()
      pending.value = res.data || pending.value
    } catch {
      // 未登录/无权限时忽略
    }
  }

  onMounted(loadPending)
  const showSettingDrawer = ref(false)
  const showRefreshAnmite = ref(false)
  const toggleRefresh = () => {
    showRefreshAnmite.value = true
    emitter.emit('reload')
    setTimeout(() => {
      showRefreshAnmite.value = false
    }, 1000)
  }

  const toggleSetting = () => {
    showSettingDrawer.value = true
  }

  const first = ref('')
  const command = ref()

  const handleCommand = () => {
    command.value.open()
  }
  const initPage = () => {
    // 判断当前用户的操作系统
    if (window.localStorage.getItem('osType') === 'WIN') {
      first.value = 'Ctrl'
    } else {
      first.value = '⌘'
    }
    // 当用户同时按下ctrl和k键的时候
    const handleKeyDown = (e) => {
      if (e.ctrlKey && e.key === 'k') {
        // 阻止浏览器默认事件
        e.preventDefault()
        handleCommand()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
  }

  initPage()

  const videoList = [
    {
      label: '1.clone项目和安装依赖',
      value: 'https://www.bilibili.com/video/BV1jx4y1s7xx'
    },
    {
      label: '2.初始化项目',
      value: 'https://www.bilibili.com/video/BV1sr421K7sv'
    },
    {
      label: '3.开启调试工具+创建初始化包',
      value: 'https://www.bilibili.com/video/BV1iH4y1c7Na'
    },
    {
      label: '4.手动使用自动化创建功能',
      value: 'https://www.bilibili.com/video/BV1UZ421T7fV'
    },
    {
      label: '5.使用已有表格创建业务',
      value: 'https://www.bilibili.com/video/BV1NE4m1977s'
    },
    {
      label: '6.使用AI创建业务和创建数据源模式的可选项',
      value: 'https://www.bilibili.com/video/BV17i421a7DE'
    },
    {
      label: '7.创建自己的后端方法',
      value: 'https://www.bilibili.com/video/BV1Yw4m1k7fg'
    },
    {
      label: '8.新增一个前端页面',
      value: 'https://www.bilibili.com/video/BV12y411i7oE'
    },
    {
      label: '9.配置一个前端二级页面',
      value: 'https://www.bilibili.com/video/BV1ZM4m1y7i3'
    },
    {
      label: '10.配置一个前端菜单参数',
      value: 'https://www.bilibili.com/video/BV1WS42197DZ'
    },
    {
      label: '11.菜单参数实战+动态菜单标题+菜单高亮配置',
      value: 'https://www.bilibili.com/video/BV1NE4m1979c'
    },
    {
      label: '12.增加菜单可控按钮',
      value: 'https://www.bilibili.com/video/BV1Sw4m1k746'
    },
    {
      label: '14.新增客户角色和其相关配置教学',
      value: 'https://www.bilibili.com/video/BV1Ki421a7X2'
    },
    {
      label: '15.发布项目上线',
      value: 'https://www.bilibili.com/video/BV1Lx4y1s77D'
    }
  ]
</script>
