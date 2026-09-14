<template>
  <div>
    <!-- 搜索区 -->
    <div class="gva-search-box">
      <el-select v-model="search.audit" placeholder="审核状态" class="w-36">
        <el-option label="全部" :value="null" />
        <el-option label="未提交" :value="0" />
        <el-option label="已通过" :value="1" />
        <el-option label="审核中" :value="2" />
        <el-option label="未通过" :value="3" />
      </el-select>
      <el-input
        v-model="search.keyword"
        placeholder="按企业名搜索"
        clearable
        class="w-48"
        @keyup.enter="handleSearch"
      />
      <el-button type="primary" @click="handleSearch">查询</el-button>
      <el-button @click="handleReset">重置</el-button>
      <el-button type="success" :disabled="selection.length === 0" @click="handleExport">
        导出选中（{{ selection.length }}）
      </el-button>
    </div>

    <!-- 表格区 -->
    <div class="gva-table-box">
      <el-table v-loading="loading" :data="tableData" border @selection-change="onSelectionChange">
        <el-table-column type="selection" width="48" align="center" />
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column prop="companyname" label="企业名称" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.companyname || '-' }}</template>
        </el-table-column>
        <el-table-column label="Logo" width="90" align="center">
          <template #default="{ row }">
            <el-image
              v-if="row.logo"
              :src="imgUrl(row.logo)"
              fit="cover"
              class="h-10 w-10 rounded"
            />
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="审核状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="auditTag(row.audit)">{{ row.auditCn }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="提交时间" width="170" align="center">
          <template #default="{ row }">{{ formatTime(row.refreshtime || row.addtime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160" align="center">
          <template #default="{ row }">
            <el-button type="primary" link @click="showDetail(row)">详情</el-button>
            <el-button type="warning" link @click="openAudit(row)">审核</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
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

    <!-- 详情抽屉 -->
    <el-drawer v-model="detailVisible" title="企业资料详情" size="480px">
      <div v-if="detail.id" class="space-y-2 text-sm text-slate-600">
        <p><span class="font-semibold">企业名称：</span>{{ detail.companyname || '-' }}</p>
        <p><span class="font-semibold">审核状态：</span>{{ detail.auditCn || '-' }}</p>
        <p><span class="font-semibold">企业性质：</span>{{ detail.natureCn || '-' }}</p>
        <p><span class="font-semibold">所属行业：</span>{{ detail.tradeCn || '-' }}</p>
        <p><span class="font-semibold">企业规模：</span>{{ detail.scaleCn || '-' }}</p>
        <p><span class="font-semibold">注册资金：</span>{{ detail.registered || '-' }}</p>
        <p><span class="font-semibold">详细地址：</span>{{ detail.address || '-' }}</p>
        <p><span class="font-semibold">联系人：</span>{{ detail.contact || '-' }}</p>
        <p><span class="font-semibold">联系电话：</span>{{ detail.telephone || '-' }}</p>
        <p><span class="font-semibold">邮箱：</span>{{ detail.email || '-' }}</p>
        <p><span class="font-semibold">官网：</span>{{ detail.website || '-' }}</p>
        <p><span class="font-semibold">企业简介：</span>{{ detail.contents || '-' }}</p>
        <p class="font-semibold">Logo：</p>
        <el-image
          v-if="detail.logo"
          :src="imgUrl(detail.logo)"
          fit="contain"
          class="max-h-40 rounded border"
        />
        <p class="font-semibold">营业执照：</p>
        <el-image
          v-if="detail.certificateImg"
          :src="imgUrl(detail.certificateImg)"
          fit="contain"
          class="max-h-40 rounded border"
        />
      </div>
    </el-drawer>

    <!-- 审核弹窗 -->
    <el-dialog
      v-model="auditVisible"
      title="企业资质审核"
      width="460px"
      :close-on-click-modal="false"
    >
      <el-form label-width="80px">
        <el-form-item label="审核结果">
          <el-radio-group v-model="auditForm.audit">
            <el-radio :value="1">通过</el-radio>
            <el-radio :value="3">不通过</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="auditForm.audit === 3" label="驳回原因">
          <el-input
            v-model="auditForm.reason"
            type="textarea"
            :rows="3"
            placeholder="必填，将同步给企业端"
          />
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
  import {
    auditCompanyProfile,
    exportCompanies,
    getCompanyProfile,
    getCompanyProfiles
  } from '@/api/hrc/company'

  defineOptions({
    name: 'CompanyAudit'
  })

  const loading = ref(false)
  const tableData = ref([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)

  // audit: null=全部（不传参数）；0 草稿 / 1 已通过 / 2 审核中 / 3 未通过
  // 默认筛「审核中」更贴合工作流
  const search = reactive({ audit: 2, keyword: '' })

  const getList = async () => {
    loading.value = true
    try {
      const params = { page: page.value, pageSize: pageSize.value }
      if (search.audit !== null) {
        params.audit = search.audit
      }
      if (search.keyword) {
        params.keyword = search.keyword.trim()
      }
      const { data } = await getCompanyProfiles(params)
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

  const handleReset = () => {
    search.audit = 2
    search.keyword = ''
    page.value = 1
    getList()
  }

  // 导出选中企业（CSV 文件流下载）
  const selection = ref([])
  const exporting = ref(false)
  const onSelectionChange = (rows) => {
    selection.value = rows
  }
  const handleExport = async () => {
    if (selection.value.length === 0) {
      ElMessage.warning('请先勾选要导出的企业')
      return
    }
    exporting.value = true
    try {
      await exportCompanies(selection.value.map((r) => r.id))
      ElMessage.success('导出成功')
    } finally {
      exporting.value = false
    }
  }

  // 详情抽屉
  const detailVisible = ref(false)
  const detail = ref({})
  const showDetail = async (row) => {
    const { data } = await getCompanyProfile(row.id)
    detail.value = data
    detailVisible.value = true
  }

  // 审核弹窗
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
      await auditCompanyProfile(auditForm.id, {
        audit: auditForm.audit,
        reason: auditForm.reason.trim()
      })
      ElMessage.success(auditForm.audit === 1 ? '已通过' : '已驳回')
      auditVisible.value = false
      getList()
    } finally {
      auditing.value = false
    }
  }

  // 审核状态数字 → el-tag 颜色
  const auditTag = (a) => {
    const map = { 0: 'info', 1: 'success', 2: 'warning', 3: 'danger' }
    return map[a] || 'info'
  }

  // 兼容 ISO8601 字符串和 unix 秒数
  const formatTime = (ts) => {
    if (!ts) return '-'
    const d = typeof ts === 'number' ? new Date(ts * 1000) : new Date(ts)
    const pad = (n) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  }

  // 后端返回相对路径 uploads/xxx，图片 src 需补前导斜杠走代理
  const imgUrl = (url) => (url ? `/${url}` : '')

  // 页面加载即拉一次列表（默认审核中）
  getList()
</script>
