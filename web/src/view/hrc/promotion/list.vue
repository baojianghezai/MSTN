<template>
  <div>
    <div class="gva-search-box">
      <el-radio-group v-model="search.audit" @change="handleSearch">
        <el-radio-button :value="-1">全部</el-radio-button>
        <el-radio-button :value="0">待审核</el-radio-button>
        <el-radio-button :value="1">已通过</el-radio-button>
        <el-radio-button :value="3">未通过</el-radio-button>
      </el-radio-group>
      <el-button class="ml-3" @click="getList">刷新</el-button>
    </div>

    <div class="gva-table-box">
      <el-table v-loading="loading" :data="tableData" border>
        <el-table-column prop="id" label="ID" width="70" align="center" />
        <el-table-column label="企业" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.companyname || '-' }}</template>
        </el-table-column>
        <el-table-column label="职位" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.jobsName || '-' }}</template>
        </el-table-column>
        <el-table-column label="类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.type === 1 ? 'warning' : 'success'" size="small">
              {{ row.type === 1 ? '首页推流' : '广告位' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="广告创意" min-width="220">
          <template #default="{ row }">
            <div v-if="row.type === 2" class="flex items-center gap-3">
              <el-image
                v-if="row.adImage"
                :src="imgUrl(row.adImage)"
                fit="contain"
                :preview-src-list="[imgUrl(row.adImage)]"
                preview-teleported
                class="h-12 w-20 rounded border border-slate-200 bg-slate-50"
              />
              <div class="min-w-0">
                <p class="truncate text-sm text-slate-700">{{ row.adTitle || row.jobsName }}</p>
                <p class="truncate text-xs text-slate-400">{{ row.adSubtitle || '-' }}</p>
              </div>
            </div>
            <span v-else class="text-slate-400">-</span>
          </template>
        </el-table-column>
        <el-table-column label="审核状态" width="110" align="center">
          <template #default="{ row }">
            <el-tag :type="auditTag(row.audit)" size="small">{{ auditText(row.audit) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="提交时间" width="170" align="center">
          <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100" align="center">
          <template #default="{ row }">
            <el-button type="warning" link @click="openAudit(row)">审核</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="gva-pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="getList"
          @current-change="getList"
        />
      </div>
    </div>

    <el-dialog v-model="auditVisible" title="推广投放审核" width="460px" :close-on-click-modal="false">
      <el-form label-width="80px">
        <el-form-item label="审核结果">
          <el-radio-group v-model="auditForm.audit">
            <el-radio :value="1">通过</el-radio>
            <el-radio :value="3">不通过</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="auditForm.audit === 3" label="驳回原因">
          <el-input v-model="auditForm.reason" type="textarea" :rows="3" placeholder="必填，将同步给企业端" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="auditVisible = false">取消</el-button>
        <el-button type="primary" :loading="auditing" @click="submitAudit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { reactive, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { auditPromotion, getPromotions } from '@/api/hrc/pending'

  defineOptions({
    name: 'PromotionAudit'
  })

  const loading = ref(false)
  const tableData = ref([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)
  const search = reactive({ audit: 0 })

  const getList = async () => {
    loading.value = true
    try {
      const { data } = await getPromotions({ page: page.value, pageSize: pageSize.value, audit: search.audit })
      tableData.value = data.list
      total.value = data.total
    } finally {
      loading.value = false
    }
  }

  const handleSearch = () => {
    page.value = 1
    getList()
  }

  const auditVisible = ref(false)
  const auditing = ref(false)
  const auditForm = reactive({ id: 0, audit: 1, reason: '' })

  const openAudit = (row) => {
    auditForm.id = row.id
    auditForm.audit = 1
    auditForm.reason = ''
    auditVisible.value = true
  }

  const submitAudit = async () => {
    if (auditForm.audit === 3 && !auditForm.reason.trim()) {
      ElMessage.warning('不通过时必须填写原因')
      return
    }
    auditing.value = true
    try {
      await auditPromotion(auditForm.id, { audit: auditForm.audit, reason: auditForm.reason.trim() })
      ElMessage.success(auditForm.audit === 1 ? '已通过' : '已驳回')
      auditVisible.value = false
      getList()
    } finally {
      auditing.value = false
    }
  }

  const auditText = (a) => ({ 0: '待审核', 1: '已通过', 3: '未通过' }[a] || '待审核')
  const auditTag = (a) => ({ 0: 'warning', 1: 'success', 3: 'danger' }[a] || 'warning')

  const formatTime = (ts) => {
    if (!ts) return '-'
    const d = typeof ts === 'number' ? new Date(ts * 1000) : new Date(ts)
    const pad = (n) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  }

  const imgUrl = (url) => (url ? (url.startsWith('http') ? url : `/${url.replace(/^\/+/, '')}`) : '')

  getList()
</script>