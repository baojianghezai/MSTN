<template>
  <div class="mx-auto max-w-6xl px-4 py-8">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold text-slate-800">套餐与订单</h1>
        <p class="mt-1 text-sm text-slate-500">购买套餐后即可获得相应的职位发布、简历下载和视频面试权益。</p>
      </div>
      <el-button :loading="loading" @click="load">刷新</el-button>
    </div>

    <el-alert
      class="mt-5"
      type="info"
      :closable="false"
      title="支付成功后自动开通"
      description="微信和支付宝到账后会自动校验并发放套餐权益，订单状态会更新为已生效。"
    />

    <el-card class="mt-5" shadow="never">
      <template #header>当前权益</template>
      <div v-if="current" class="grid gap-4 grid-cols-2 sm:grid-cols-3 lg:grid-cols-5">
        <div><p class="text-sm text-slate-500">当前套餐</p><p class="mt-1 font-semibold truncate">{{ current.setmealName }}</p></div>
        <div><p class="text-sm text-slate-500">有效期至</p><p class="mt-1 font-semibold">{{ formatTime(current.expireAt) }}</p></div>
        <div><p class="text-sm text-slate-500">在线职位</p><p class="mt-1 font-semibold">{{ current.jobsMeanwhile || '不限' }}</p></div>
        <div><p class="text-sm text-slate-500">简历下载</p><p class="mt-1 font-semibold">{{ current.resumeDownloadsTotal - current.resumeDownloadsUsed }} / {{ current.resumeDownloadsTotal }}</p></div>
        <div><p class="text-sm text-slate-500">首页推广</p><p class="mt-1 font-semibold">推流 {{ current.homePushSlots }} / 广告 {{ current.homeAdSlots }}</p></div>
      </div>
      <el-empty v-else :image-size="62" description="尚未购买套餐" />
    </el-card>

    <section class="mt-8">
      <h2 class="text-lg font-semibold text-slate-800">选择套餐</h2>
      <div v-if="plans.length" class="mt-4 grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
        <el-card v-for="plan in plans" :key="plan.id" shadow="never" class="flex flex-col">
          <h3 class="text-lg font-semibold text-slate-800">{{ plan.name }}</h3>
          <p class="mt-3 text-3xl font-bold text-primary-600">{{ formatPrice(plan.price) }}</p>
          <p class="mt-1 text-sm text-slate-500">{{ plan.durationDays }} 天有效</p>
          <ul class="mt-5 flex-1 space-y-2 text-sm text-slate-600">
            <li>在线职位：{{ plan.jobsMeanwhile || '不限' }}</li>
            <li>简历下载：{{ plan.resumeDownloads }} 次</li>
            <li>首页推流：{{ plan.homePushSlots }} 位</li>
            <li>首页广告：{{ plan.homeAdSlots }} 位</li>
            <li>视频面试：{{ plan.enableVideo ? '支持' : '暂不支持' }}</li>
          </ul>
          <p v-if="plan.description" class="mt-4 text-sm text-slate-500 line-clamp-2">{{ plan.description }}</p>
          <el-button class="mt-5" type="primary" :loading="creatingId === plan.id" @click="createOrder(plan)">创建订单</el-button>
        </el-card>
      </div>
      <el-empty v-else :image-size="70" description="暂未配置可购买套餐" />
    </section>

    <el-card class="mt-8" shadow="never">
      <template #header>我的订单</template>
      <el-table :data="orders" v-loading="loading" class="w-full">
        <el-table-column prop="oid" label="订单号" min-width="180" show-overflow-tooltip />
        <el-table-column prop="setmealName" label="套餐" min-width="100" show-overflow-tooltip />
        <el-table-column label="金额" width="90"><template #default="{ row }">{{ formatPrice(row.amount) }}</template></el-table-column>
        <el-table-column label="状态" width="80"><template #default="{ row }"><el-tag :type="orderTag(row.isPaid)" size="small">{{ orderStatus(row.isPaid) }}</el-tag></template></el-table-column>
        <el-table-column label="剩余时间" width="90">
          <template #default="{ row }">
            <template v-if="row.isPaid === 1 && row.expireAt > 0">
              <span :class="getRemaining(row.expireAt) <= 300 ? 'text-red-500 font-semibold' : 'text-slate-600'">
                {{ formatRemaining(getRemaining(row.expireAt)) }}
              </span>
            </template>
            <span v-else class="text-slate-400">-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="180" fixed="right">
          <template #default="{ row }">
            <template v-if="row.isPaid === 1">
              <el-button size="small" type="success" :loading="payingKey === `${row.id}-wechat_native`" @click="startGatewayPayment(row, 'wechat_native')">微信</el-button>
              <el-button size="small" type="primary" :loading="payingKey === `${row.id}-alipay_page`" @click="startGatewayPayment(row, 'alipay_page')">支付宝</el-button>
              <el-button size="small" type="danger" link @click="cancel(row.id)">取消</el-button>
            </template>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="wechatDialogVisible" title="微信支付" width="360px" @closed="stopPolling">
      <div class="flex flex-col items-center gap-4">
        <img v-if="wechatQRCode" :src="wechatQRCode" class="h-56 w-56" alt="微信支付二维码" />
        <p class="text-sm text-slate-500">请使用微信扫一扫完成支付</p>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
  import { onBeforeUnmount, onMounted, ref } from 'vue'
  import QRCode from 'qrcode'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    cancelOrder,
    createSetmealOrder,
    getCurrentSetmeal,
    getOrders,
    getSetmeals,
    startPayment,
    type BillingOrder,
    type CurrentSetmeal,
    type PaymentStart,
    type Setmeal
  } from '@/api/billing'

  const loading = ref(false)
  const creatingId = ref(0)
  const payingKey = ref('')
  const plans = ref<Setmeal[]>([])
  const current = ref<CurrentSetmeal | null>(null)
  const orders = ref<BillingOrder[]>([])
  const wechatDialogVisible = ref(false)
  const wechatQRCode = ref('')
  const now = ref(Date.now())
  let pollTimer: ReturnType<typeof setInterval> | undefined
  let countdownTimer: ReturnType<typeof setInterval> | undefined

  const formatPrice = (amount: number) => `￥${(amount / 100).toFixed(2)}`
  const formatTime = (value: number) => value ? new Date(value * 1000).toLocaleString() : '-'
  const orderStatus = (status: number) => ({ 1: '待支付', 2: '已生效', 3: '已取消' }[status] || '未知')
  const orderTag = (status: number) => ({ 1: 'warning', 2: 'success', 3: 'info' }[status] || 'info')

  // 后端 expireAt 是 unix 时间戳（秒），表示订单过期的绝对时间
  const getRemaining = (expireAt: number) => {
    if (!expireAt) return 0
    return Math.max(0, expireAt - Math.floor(now.value / 1000))
  }

  const formatRemaining = (seconds: number) => {
    if (seconds <= 0) return '已超时'
    const m = Math.floor(seconds / 60)
    const s = seconds % 60
    return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
  }

  const autoCancelExpired = async () => {
    const expiredOrders = orders.value.filter(
      (o) => o.isPaid === 1 && o.expireAt > 0 && getRemaining(o.expireAt) <= 0
    )
    for (const order of expiredOrders) {
      try {
        await cancelOrder(order.id)
        ElMessage.warning(`订单 ${order.oid} 已超时自动取消`)
      } catch { /* ignore */ }
    }
    if (expiredOrders.length > 0) {
      await load()
    }
  }

  const load = async () => {
    loading.value = true
    try {
      const [planRes, currentRes, orderRes] = await Promise.all([
        getSetmeals(),
        getCurrentSetmeal(),
        getOrders({ page: 1, pageSize: 20 })
      ])
      plans.value = planRes.data
      current.value = currentRes.data
      orders.value = orderRes.data.list
      if (wechatDialogVisible.value && orders.value.every((order) => order.isPaid !== 1)) {
        wechatDialogVisible.value = false
      }
    } finally {
      loading.value = false
    }
  }

  const createOrder = async (plan: Setmeal) => {
    creatingId.value = plan.id
    try {
      const result = await createSetmealOrder(plan.id)
      ElMessage.success(`订单已创建：${result.data.oid}`)
      await load()
    } finally {
      creatingId.value = 0
    }
  }

  const stopPolling = () => {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = undefined
    }
  }

  const showWechatQRCode = async (payment: PaymentStart) => {
    if (!payment.qrCodeUrl) return
    wechatQRCode.value = await QRCode.toDataURL(payment.qrCodeUrl, { width: 280, margin: 1 })
    wechatDialogVisible.value = true
    stopPolling()
    pollTimer = setInterval(() => { void load() }, 5000)
  }

  const startGatewayPayment = async (order: BillingOrder, provider: PaymentStart['provider']) => {
    payingKey.value = `${order.id}-${provider}`
    try {
      const result = await startPayment(order.id, provider)
      if (provider === 'wechat_native') {
        await showWechatQRCode(result.data)
      } else if (result.data.redirectUrl) {
        window.open(result.data.redirectUrl, '_blank', 'noopener')
        ElMessage.success('已打开支付宝支付页面，完成付款后订单会自动生效')
      }
    } finally {
      payingKey.value = ''
    }
  }

  const cancel = async (id: number) => {
    await ElMessageBox.confirm('取消后该订单不能恢复，是否继续？', '确认取消', { type: 'warning' })
    await cancelOrder(id)
    ElMessage.success('订单已取消')
    await load()
  }

  onMounted(() => {
    countdownTimer = setInterval(() => {
      now.value = Date.now()
      void autoCancelExpired()
    }, 1000)
    void load()
  })
  onBeforeUnmount(() => {
    stopPolling()
    if (countdownTimer) {
      clearInterval(countdownTimer)
      countdownTimer = undefined
    }
  })
</script>
