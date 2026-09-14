<template>
  <div class="space-y-4">
    <el-alert
      title="数据清理"
      description="仅清理固定范围的非核心数据。请先预览，再确认执行。"
      type="warning"
      :closable="false"
      show-icon
    />

    <el-card shadow="never">
      <template #header>
        <span class="font-semibold">清理范围</span>
      </template>
      <div class="flex flex-wrap items-end gap-3">
        <el-form-item label="清理目标" class="mb-0">
          <el-select v-model="form.target" class="w-64" @change="resetPreview">
            <el-option
              v-for="item in targets"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item v-if="requiresRetention" label="保留天数" class="mb-0">
          <el-input-number
            v-model="form.retentionDays"
            :min="7"
            :max="3650"
            controls-position="right"
            @change="resetPreview"
          />
        </el-form-item>
        <el-button type="primary" :loading="previewLoading" @click="loadPreview">
          预览清理范围
        </el-button>
      </div>

      <div v-if="preview" class="mt-5 border-t border-slate-200 pt-4">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="目标">{{ preview.title }}</el-descriptions-item>
          <el-descriptions-item label="预计影响">
            <span class="font-semibold text-red-600">{{ preview.affectedCount }} 条</span>
          </el-descriptions-item>
          <el-descriptions-item label="截止时间">{{ formatTime(preview.cutoffAt) }}</el-descriptions-item>
          <el-descriptions-item label="处理方式">{{ preview.softDelete ? '逻辑下线' : '物理删除' }}</el-descriptions-item>
          <el-descriptions-item label="范围说明" :span="2">{{ preview.description }}</el-descriptions-item>
        </el-descriptions>
        <div class="mt-4 flex items-center justify-between gap-3">
          <span class="text-sm text-slate-500">执行时会按当前数据重新计算，并写入操作审计。</span>
          <el-button type="danger" :disabled="preview.affectedCount === 0" @click="openConfirm">
            执行清理
          </el-button>
        </div>
      </div>
    </el-card>

    <el-card shadow="never">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="font-semibold">近期执行记录</span>
          <el-button :loading="historyLoading" @click="loadHistory">刷新</el-button>
        </div>
      </template>
      <el-table v-loading="historyLoading" :data="history" border>
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column label="清理目标" min-width="150">
          <template #default="{ row }">{{ targetLabel(row.target) }}</template>
        </el-table-column>
        <el-table-column label="保留天数" width="110" align="center">
          <template #default="{ row }">{{ row.retentionDays || '-' }}</template>
        </el-table-column>
        <el-table-column label="影响数量" width="110" align="center">
          <template #default="{ row }">{{ row.affectedCount }}</template>
        </el-table-column>
        <el-table-column label="执行人" min-width="120">
          <template #default="{ row }">{{ row.operatorName || `用户 ${row.operatorId || '-'}` }}</template>
        </el-table-column>
        <el-table-column label="执行时间" min-width="170">
          <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
        </el-table-column>
      </el-table>
      <div class="gva-pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @current-change="loadHistory"
          @size-change="loadHistory"
        />
      </div>
    </el-card>

    <el-dialog v-model="confirmVisible" title="确认执行清理" width="440px" :close-on-click-modal="false">
      <p class="mb-4 text-sm text-slate-600">
        将处理 {{ preview?.affectedCount || 0 }} 条 {{ preview?.title || '' }}。请输入 CONFIRM 继续。
      </p>
      <el-input v-model="confirmation" placeholder="CONFIRM" autocomplete="off" />
      <template #footer>
        <el-button @click="confirmVisible = false">取消</el-button>
        <el-button type="danger" :loading="executing" :disabled="confirmation !== 'CONFIRM'" @click="execute">
          确认执行
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { computed, reactive, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import {
    executeDataCleanup,
    getDataCleanupHistory,
    previewDataCleanup
  } from '@/api/hrc/dataCleanup'

  defineOptions({ name: 'HrcDataCleanup' })

  const targets = [
    { value: 'expired_jobs', label: '过期职位', requiresRetention: false },
    { value: 'rejected_job_drafts', label: '已驳回职位草稿', requiresRetention: false },
    { value: 'expired_promotions', label: '失效首页推广', requiresRetention: false },
    { value: 'old_payment_notify_logs', label: '历史支付通知日志', requiresRetention: true },
    { value: 'old_wxpay_logs', label: '历史微信支付日志', requiresRetention: true }
  ]

  const form = reactive({ target: 'expired_jobs', retentionDays: 30 })
  const preview = ref(null)
  const previewLoading = ref(false)
  const confirmVisible = ref(false)
  const confirmation = ref('')
  const executing = ref(false)
  const history = ref([])
  const historyLoading = ref(false)
  const page = ref(1)
  const pageSize = ref(10)
  const total = ref(0)

  const selectedTarget = computed(() => targets.find((item) => item.value === form.target))
  const requiresRetention = computed(() => selectedTarget.value?.requiresRetention === true)
  const requestPayload = () => ({
    target: form.target,
    retentionDays: requiresRetention.value ? form.retentionDays : 0
  })
  const targetLabel = (value) => targets.find((item) => item.value === value)?.label || value

  const resetPreview = () => {
    preview.value = null
  }

  const loadPreview = async () => {
    previewLoading.value = true
    try {
      const { data } = await previewDataCleanup(requestPayload())
      preview.value = data
    } finally {
      previewLoading.value = false
    }
  }

  const openConfirm = () => {
    confirmation.value = ''
    confirmVisible.value = true
  }

  const execute = async () => {
    executing.value = true
    try {
      const { data } = await executeDataCleanup({
        ...requestPayload(),
        confirmation: confirmation.value
      })
      ElMessage.success(`已处理 ${data.affectedCount} 条记录`)
      confirmVisible.value = false
      await Promise.all([loadPreview(), loadHistory()])
    } finally {
      executing.value = false
    }
  }

  const loadHistory = async () => {
    historyLoading.value = true
    try {
      const { data } = await getDataCleanupHistory({
        page: page.value,
        pageSize: pageSize.value
      })
      history.value = data.list || []
      total.value = data.total || 0
    } finally {
      historyLoading.value = false
    }
  }

  const formatTime = (value) => {
    if (!value) return '-'
    const d = typeof value === 'number' ? new Date(value * 1000) : new Date(value)
    return d.toLocaleString()
  }

  loadPreview()
  loadHistory()
</script>
