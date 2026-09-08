<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-lg font-semibold">套餐管理</h2>
        <p class="mt-1 text-sm text-gray-500">金额单位为元；确认到账后将自动发放套餐权益。</p>
      </div>
      <el-button type="primary" @click="openCreate">新增套餐</el-button>
    </div>

    <el-table v-loading="plansLoading" :data="plans" border>
      <el-table-column prop="name" label="套餐" min-width="130" />
      <el-table-column label="价格" width="120"><template #default="{ row }">{{ yuan(row.price) }}</template></el-table-column>
      <el-table-column prop="durationDays" label="有效期" width="100"><template #default="{ row }">{{ row.durationDays }} 天</template></el-table-column>
      <el-table-column prop="jobsMeanwhile" label="在线职位" width="110" />
      <el-table-column prop="resumeDownloads" label="简历下载" width="110" />
      <el-table-column prop="homeAdSlots" label="首页广告位" width="120">
        <template #default="{ row }">{{ row.homeAdSlots || 0 }} 个</template>
      </el-table-column>
      <el-table-column prop="homePushSlots" label="首页推流位" width="120">
        <template #default="{ row }">{{ row.homePushSlots || 0 }} 个</template>
      </el-table-column>
      <el-table-column label="视频面试" width="100"><template #default="{ row }">{{ row.enableVideo ? '支持' : '不支持' }}</template></el-table-column>
      <el-table-column label="状态" width="90"><template #default="{ row }"><el-tag :type="row.display ? 'success' : 'info'">{{ row.display ? '上架' : '下架' }}</el-tag></template></el-table-column>
      <el-table-column label="操作" width="90"><template #default="{ row }"><el-button link type="primary" @click="openEdit(row)">编辑</el-button></template></el-table-column>
    </el-table>

    <el-card shadow="never">
      <template #header>订单确认</template>
      <div class="mb-4 flex gap-2">
        <el-select v-model="status" class="w-32" @change="loadOrders"><el-option label="全部" :value="0" /><el-option label="待确认" :value="1" /><el-option label="已生效" :value="2" /><el-option label="已取消" :value="3" /></el-select>
        <el-button @click="loadOrders">刷新</el-button>
      </div>
      <el-table v-loading="ordersLoading" :data="orders" border>
        <el-table-column prop="oid" label="订单号" min-width="210" />
        <el-table-column prop="setmealName" label="套餐" min-width="110" />
        <el-table-column prop="uid" label="企业 UID" width="100" />
        <el-table-column label="应收" width="110"><template #default="{ row }">{{ yuan(row.amount) }}</template></el-table-column>
        <el-table-column label="状态" width="100"><template #default="{ row }"><el-tag :type="tagType(row.isPaid)">{{ statusText(row.isPaid) }}</el-tag></template></el-table-column>
        <el-table-column label="创建时间" min-width="160"><template #default="{ row }">{{ time(row.createdAt) }}</template></el-table-column>
        <el-table-column label="操作" width="110"><template #default="{ row }"><el-button v-if="row.isPaid === 1" type="success" link @click="confirm(row)">确认到账</el-button></template></el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑套餐' : '新增套餐'" width="560px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="套餐名称" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="价格（元）" required><el-input-number v-model="form.priceYuan" :min="0" :precision="2" /></el-form-item>
        <el-form-item label="有效天数" required><el-input-number v-model="form.durationDays" :min="1" /></el-form-item>
        <el-form-item label="在线职位数"><el-input-number v-model="form.jobsMeanwhile" :min="0" /></el-form-item>
        <el-form-item label="简历下载数"><el-input-number v-model="form.resumeDownloads" :min="0" /></el-form-item>
        <el-form-item label="首页广告位">
          <el-input-number v-model="form.homeAdSlots" :min="0" />
          <span class="ml-2 text-xs text-gray-500">0 表示不支持首页广告展示</span>
        </el-form-item>
        <el-form-item label="首页推流位">
          <el-input-number v-model="form.homePushSlots" :min="0" />
          <span class="ml-2 text-xs text-gray-500">0 表示不支持职位首页推流</span>
        </el-form-item>
        <el-form-item label="视频面试"><el-switch v-model="form.enableVideo" /></el-form-item>
        <el-form-item label="上架"><el-switch v-model="form.display" /></el-form-item>
        <el-form-item label="说明"><el-input v-model="form.description" type="textarea" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { reactive, ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { confirmOrder, createSetmeal, getOrders, getSetmeals, updateSetmeal } from '@/api/hrc/billing'

  defineOptions({ name: 'BillingManage' })
  const plans = ref([])
  const orders = ref([])
  const plansLoading = ref(false)
  const ordersLoading = ref(false)
  const saving = ref(false)
  const status = ref(0)
  const dialogVisible = ref(false)
  const emptyForm = () => ({
    id: 0,
    name: '',
    priceYuan: 0,
    durationDays: 30,
    jobsMeanwhile: 0,
    resumeDownloads: 0,
    homeAdSlots: 0,
    homePushSlots: 0,
    enableVideo: false,
    display: true,
    description: ''
  })
  const form = reactive(emptyForm())
  const yuan = (cents) => `￥${(cents / 100).toFixed(2)}`
  const time = (value) => value ? new Date(value * 1000).toLocaleString() : '-'
  const statusText = (value) => ({ 1: '待确认', 2: '已生效', 3: '已取消' }[value] || '未知')
  const tagType = (value) => ({ 1: 'warning', 2: 'success', 3: 'info' }[value] || 'info')
  const resetForm = (value = emptyForm()) => Object.assign(form, value)
  const loadPlans = async () => { plansLoading.value = true; try { plans.value = (await getSetmeals()).data } finally { plansLoading.value = false } }
  const loadOrders = async () => { ordersLoading.value = true; try { orders.value = (await getOrders({ page: 1, pageSize: 100, status: status.value || undefined })).data.list } finally { ordersLoading.value = false } }
  const openCreate = () => { resetForm(); dialogVisible.value = true }
  const openEdit = (row) => {
    resetForm({ ...emptyForm(), ...row, homeAdSlots: row.homeAdSlots || 0, homePushSlots: row.homePushSlots || 0, priceYuan: row.price / 100 })
    dialogVisible.value = true
  }
  const save = async () => {
    if (!form.name || form.durationDays < 1) return ElMessage.warning('请完整填写套餐信息')
    saving.value = true
    try {
      const data = { ...form, price: Math.round(form.priceYuan * 100) }
      if (form.id) await updateSetmeal(form.id, data); else await createSetmeal(data)
      ElMessage.success('套餐已保存'); dialogVisible.value = false; await loadPlans()
    } finally { saving.value = false }
  }
  const confirm = async (row) => {
    await ElMessageBox.confirm(`确认订单 ${row.oid} 已收款 ${yuan(row.amount)}？确认后会立即发放套餐权益。`, '确认到账', { type: 'warning' })
    await confirmOrder(row.id, { payAmount: row.amount, payment: 'manual' })
    ElMessage.success('订单已确认，套餐权益已发放'); await loadOrders()
  }
  loadPlans(); loadOrders()
</script>
