<template>
  <div>
    <!-- Tab：职位管理 / 待审职位 -->
    <el-tabs v-model="activeTab" @tab-change="handleTabChange">
      <!-- 职位管理 -->
      <el-tab-pane label="职位管理" name="jobs">
        <!-- 搜索区 -->
        <div class="gva-search-box">
          <el-select v-model="search.audit" placeholder="审核状态" class="w-36">
            <el-option label="全部" :value="null" />
            <el-option label="草稿" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="审核中" :value="2" />
            <el-option label="不通过" :value="3" />
          </el-select>
          <el-input
            v-model="search.keyword"
            placeholder="按职位名/企业名搜索"
            clearable
            class="w-56"
            @keyup.enter="handleSearch"
          />
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </div>

        <!-- 表格区 -->
        <div class="gva-table-box">
          <div class="mb-3">
            <el-button
              type="success"
              :disabled="selection.length === 0"
              @click="handleExport"
            >
              导出选中（{{ selection.length }}）
            </el-button>
          </div>
          <el-table
            v-loading="loading"
            :data="tableData"
            border
            @selection-change="onSelectionChange"
          >
            <el-table-column type="selection" width="48" align="center" />
            <el-table-column prop="id" label="ID" width="80" align="center" />
            <el-table-column prop="jobsName" label="职位名称" min-width="160" show-overflow-tooltip />
            <el-table-column prop="companyname" label="企业名称" min-width="160" show-overflow-tooltip>
              <template #default="{ row }">{{ row.companyname || '-' }}</template>
            </el-table-column>
            <el-table-column label="薪资" width="150" align="center">
              <template #default="{ row }">
                <span v-if="row.negotiable === 1">面议</span>
                <span v-else>{{ row.minwage }}-{{ row.maxwage }}</span>
              </template>
            </el-table-column>
            <el-table-column label="审核状态" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="auditTag(row.audit)">{{ auditCn(row.audit) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="发布/刷新时间" width="170" align="center">
              <template #default="{ row }">{{ formatTime(row.refreshtime || row.addtime) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="100" align="center">
              <template #default="{ row }">
                <el-button type="primary" link @click="openAudit(row)">审核</el-button>
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
      </el-tab-pane>

      <!-- 待审职位 -->
      <el-tab-pane label="待审职位" name="tmp">
        <div class="gva-table-box">
          <el-table v-loading="loading" :data="tableData" border>
            <el-table-column prop="id" label="ID" width="80" align="center" />
            <el-table-column prop="jobsName" label="职位名称" min-width="160" show-overflow-tooltip />
            <el-table-column prop="companyname" label="企业名称" min-width="160" show-overflow-tooltip>
              <template #default="{ row }">{{ row.companyname || '-' }}</template>
            </el-table-column>
            <el-table-column label="薪资" width="150" align="center">
              <template #default="{ row }">
                <span v-if="row.negotiable === 1">面议</span>
                <span v-else>{{ row.minwage }}-{{ row.maxwage }}</span>
              </template>
            </el-table-column>
            <el-table-column label="提交时间" width="170" align="center">
              <template #default="{ row }">{{ formatTime(row.addtime) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="100" align="center">
              <template #default="{ row }">
                <el-button type="primary" link @click="openAudit(row)">审核</el-button>
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
      </el-tab-pane>
    </el-tabs>

    <!-- 审核弹窗 -->
    <el-dialog
      v-model="auditVisible"
      title="职位审核"
      width="860px"
      top="5vh"
      :close-on-click-modal="false"
    >
      <div v-loading="detailLoading" class="max-h-[65vh] overflow-y-auto pr-2">
        <!-- 不通过原因（历史审核记录） -->
        <el-alert
          v-if="auditDetail && auditDetail.reason"
          type="error"
          :closable="false"
          show-icon
          class="mb-4"
          title="历史不通过原因"
          :description="auditDetail.reason"
        />

        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="职位名称" :span="2">
            {{ auditDetail?.jobsName || auditRow.jobsName || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="企业名称">
            {{ auditDetail?.companyname || auditRow.companyname || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="企业ID / 职位ID">
            {{ auditDetail?.companyId || '-' }} / {{ auditDetail?.id || auditRow.id || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="工作性质">
            {{ auditDetail?.natureCn || '-' }}（{{ auditDetail?.nature ?? '-' }}）
          </el-descriptions-item>
          <el-descriptions-item label="招聘人数">
            {{ auditDetail?.amount ?? '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="职位分类">
            {{ auditDetail?.categoryCn || '-' }}（{{ auditDetail?.topclass }}-{{ auditDetail?.category }}-{{ auditDetail?.subclass }}）
          </el-descriptions-item>
          <el-descriptions-item label="所属行业">
            {{ auditDetail?.trade || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="工作地区">
            {{ auditDetail?.districtCn || auditDetail?.district || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="学历要求">
            {{ auditDetail?.education || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="经验要求">
            {{ auditDetail?.experience || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="招聘部门">
            {{ auditDetail?.department || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="性别要求">
            {{ sexCn(auditDetail?.sex) }}
          </el-descriptions-item>
          <el-descriptions-item label="紧急 / 置顶">
            {{ auditDetail?.emergency === 1 ? '紧急' : '否' }} / {{ auditDetail?.stick === 1 ? '置顶' : '否' }}
          </el-descriptions-item>
          <el-descriptions-item label="薪资">
            <span v-if="auditDetail?.negotiable === 1">面议</span>
            <span v-else>{{ auditDetail?.minwage ?? '-' }}-{{ auditDetail?.maxwage ?? '-' }} 元/月</span>
          </el-descriptions-item>
          <el-descriptions-item label="有效期至">
            {{ formatTime(auditDetail?.deadline) }}
          </el-descriptions-item>
          <el-descriptions-item label="发布时间">
            {{ formatTime(auditDetail?.addtime) }}
          </el-descriptions-item>
          <el-descriptions-item label="刷新时间">
            {{ formatTime(auditDetail?.refreshtime) }}
          </el-descriptions-item>
          <el-descriptions-item label="标签" :span="2">
            {{ auditDetail?.tag || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="职位描述" :span="2">
            <div class="whitespace-pre-wrap break-words">
              {{ auditDetail?.contents || '-' }}
            </div>
          </el-descriptions-item>
        </el-descriptions>

        <!-- 联系方式 -->
        <el-divider content-position="left">联系方式</el-divider>
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="联系人">
            {{ auditDetail?.contact?.contact || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="手机号">
            {{ auditDetail?.contact?.telephone || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="QQ">
            {{ auditDetail?.contact?.qq || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="座机">
            {{ auditDetail?.contact?.landlineTel || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="邮箱">
            {{ auditDetail?.contact?.email || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="工作地址">
            {{ auditDetail?.contact?.address || '-' }}
          </el-descriptions-item>
        </el-descriptions>

        <!-- 关联企业资质 -->
        <el-divider content-position="left">关联企业资质</el-divider>
        <el-descriptions v-if="auditDetail?.company" :column="2" border size="small">
          <el-descriptions-item label="企业名称">
            {{ auditDetail.company.companyname || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="资质状态">
            {{ auditCn(auditDetail.company.audit) }}
          </el-descriptions-item>
          <el-descriptions-item label="企业性质">
            {{ auditDetail.company.natureCn || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="所属行业">
            {{ auditDetail.company.tradeCn || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="企业规模">
            {{ auditDetail.company.scaleCn || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="注册资金">
            {{ auditDetail.company.registered || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="详细地址" :span="2">
            {{ auditDetail.company.address || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="联系人">
            {{ auditDetail.company.contact || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="联系电话">
            {{ auditDetail.company.telephone || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="邮箱">
            {{ auditDetail.company.email || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="官网">
            {{ auditDetail.company.website || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="企业简介" :span="2">
            <div class="whitespace-pre-wrap break-words">
              {{ auditDetail.company.contents || '-' }}
            </div>
          </el-descriptions-item>
        </el-descriptions>
        <el-empty v-else description="未关联到企业资质" :image-size="60" />
      </div>

      <!-- 审核操作 -->
      <el-divider content-position="left">审核操作</el-divider>
      <el-form label-width="80px">
        <el-form-item label="审核结果">
          <el-radio-group v-model="auditForm.audit">
            <el-radio :value="1">通过</el-radio>
            <el-radio :value="3">不通过</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="auditForm.audit === 3" label="不通过原因">
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
  import { useRoute } from 'vue-router'
  import { ElMessage } from 'element-plus'
  import {
    auditAdminJob,
    exportAdminJobs,
    getAdminJobDetail,
    getAdminJobs,
    getAdminJobsTmp
  } from '@/api/hrc/jobs'

  defineOptions({
    name: 'JobsManage'
  })

  const route = useRoute()
  const loading = ref(false)
  const tableData = ref([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)
  const selection = ref([])

  const activeTab = ref(route.query.tab === 'tmp' ? 'tmp' : 'jobs')
  // audit: null=全部（不传参数）；0 草稿 / 1 已通过 / 2 审核中 / 3 不通过
  const search = reactive({ audit: null, keyword: '' })

  const onSelectionChange = (rows) => {
    selection.value = rows
  }

  const handleExport = async () => {
    if (selection.value.length === 0) {
      ElMessage.warning('请先勾选要导出的职位')
      return
    }
    try {
      await exportAdminJobs(selection.value.map((r) => r.id))
      ElMessage.success('导出成功')
    } catch {
      // 错误提示已在 api 内弹出
    }
  }

  const getList = async () => {
    loading.value = true
    selection.value = []
    try {
      if (activeTab.value === 'tmp') {
        const { data } = await getAdminJobsTmp({ page: page.value, pageSize: pageSize.value })
        tableData.value = data.list
        total.value = data.total
        return
      }
      const params = { page: page.value, pageSize: pageSize.value }
      if (search.audit !== null) {
        params.audit = search.audit
      }
      if (search.keyword) {
        params.keyword = search.keyword.trim()
      }
      const { data } = await getAdminJobs(params)
      tableData.value = data.list
      total.value = data.total
    } finally {
      loading.value = false
    }
  }

  const handleTabChange = () => {
    page.value = 1
    getList()
  }

  const handleSearch = () => {
    page.value = 1
    getList()
  }

  const handleReset = () => {
    search.audit = null
    search.keyword = ''
    page.value = 1
    getList()
  }

  // 审核弹窗
  const auditVisible = ref(false)
  const auditing = ref(false)
  const detailLoading = ref(false)
  const auditRow = ref({})
  const auditDetail = ref(null)
  const auditForm = reactive({ id: 0, audit: 1, reason: '' })

  const openAudit = async (row) => {
    auditRow.value = row
    auditForm.id = row.id
    auditForm.audit = 1
    auditForm.reason = ''
    auditDetail.value = null
    auditVisible.value = true
    detailLoading.value = true
    try {
      const { data } = await getAdminJobDetail(row.id)
      auditDetail.value = data
    } catch {
      // 详情加载失败不阻断审核，仅提示
      ElMessage.warning('职位详情加载失败，请稍后重试')
    } finally {
      detailLoading.value = false
    }
  }

  const submitAudit = async () => {
    if (auditForm.audit === 3 && !auditForm.reason.trim()) {
      ElMessage.warning('不通过时必须填写原因')
      return
    }
    auditing.value = true
    try {
      await auditAdminJob(auditForm.id, {
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

  const auditCn = (a) => {
    const map = { 0: '草稿', 1: '已通过', 2: '审核中', 3: '不通过' }
    return map[a] ?? '未知'
  }
  const auditTag = (a) => {
    const map = { 0: 'info', 1: 'success', 2: 'warning', 3: 'danger' }
    return map[a] || 'info'
  }

  const sexCn = (v) => {
    const map = { 1: '男', 2: '女', 3: '不限' }
    return map[v] ?? '-'
  }

  const formatTime = (ts) => {
    if (!ts) return '-'
    const d = typeof ts === 'number' ? new Date(ts * 1000) : new Date(ts)
    const pad = (n) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  }

  getList()
</script>
