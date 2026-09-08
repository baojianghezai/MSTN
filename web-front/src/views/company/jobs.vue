<template>
  <div class="mx-auto max-w-4xl px-4 py-16">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-800">职位管理</h1>
      <el-button type="primary" @click="openCreate">发布职位</el-button>
    </div>
    <p class="mt-2 text-sm text-slate-500">发布、编辑、暂停、恢复、刷新、删除职位</p>

    <!-- 列表 -->
    <el-card class="mt-6">
      <div v-if="loading && tableData.length === 0" class="space-y-4 py-2">
        <el-skeleton v-for="i in 4" :key="i" :rows="1" animated />
      </div>
      <el-table v-else v-loading="loading" :data="tableData">
        <template #empty>
          <el-empty :image-size="80" description="还没有发布职位">
            <el-button type="primary" @click="openCreate">发布职位</el-button>
          </el-empty>
        </template>
        <el-table-column prop="id" label="ID" width="70" align="center" />
        <el-table-column label="职位名称" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.jobsName }}</template>
        </el-table-column>
        <el-table-column label="薪资" width="150" align="center">
          <template #default="{ row }">
            <span v-if="row.negotiable === 1">面议</span>
            <span v-else>{{ row.minwage }}-{{ row.maxwage }}</span>
          </template>
        </el-table-column>
        <el-table-column label="招聘人数" width="90" align="center">
          <template #default="{ row }">{{ row.amount }}</template>
        </el-table-column>
        <el-table-column label="状态" width="140" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.pending" type="warning">审核中</el-tag>
            <el-tag v-else-if="row.audit === 3" type="danger">不通过</el-tag>
            <el-tag v-else-if="row.display === 2" type="info">已暂停</el-tag>
            <el-tag v-else type="success">展示中</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="刷新时间" width="170" align="center">
          <template #default="{ row }">{{ formatTime(row.refreshtime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="230" align="center">
          <template #default="{ row }">
            <el-button type="primary" link @click="openEdit(row)">编辑</el-button>
            <el-button v-if="row.display === 1" type="warning" link @click="togglePause(row, 2)">暂停</el-button>
            <el-button v-else type="success" link @click="togglePause(row, 1)">恢复</el-button>
            <el-button type="primary" link @click="handleRefresh(row)">刷新</el-button>
            <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="mt-4 flex justify-end">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @size-change="getList"
          @current-change="getList"
        />
      </div>
    </el-card>

    <!-- 发布/编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑职位' : '发布职位'"
      width="720px"
      :close-on-click-modal="false"
      top="4vh"
    >
      <!-- 不通过原因提示 -->
      <el-alert
        v-if="isEdit && detail.reason"
        type="error"
        :closable="false"
        class="mb-4"
        title="审核不通过原因"
        :description="detail.reason"
      />

      <el-form :model="form" label-width="90px" @submit.prevent>
        <el-form-item label="职位名称" required>
          <!-- X3：职位名称列表化（categories.jobtitle，纯选择不允许自由输入；
               编辑回显时历史名字不在列表内则加虚拟 option，不丢数据） -->
          <el-select
            v-model="form.jobsName"
            filterable
            :allow-create="false"
            default-first-option
            placeholder="请选择职位名称"
            class="w-full"
          >
            <el-option v-if="historyTitleOption" :label="form.jobsName" :value="form.jobsName" />
            <el-option v-for="t in jobTitles" :key="t.id" :label="t.name" :value="t.name" />
          </el-select>
          <div v-if="!jobTitles.length" class="mt-1 text-xs text-slate-400">
            职位名称列表暂未配置（后端 jobtitle 分组 seed 后自动出现）
          </div>
        </el-form-item>

        <div class="grid grid-cols-1 gap-x-6 sm:grid-cols-2">
          <el-form-item label="工作性质">
            <el-select v-model="form.nature" placeholder="请选择" class="w-full" clearable :empty-values="[null, undefined, 0]" @change="onNatureChange">
              <el-option v-for="n in jobNatures" :key="n.value" :label="n.label" :value="n.value" />
            </el-select>
          </el-form-item>
          <el-form-item label="性别要求">
            <el-select v-model="form.sex" placeholder="请选择" class="w-full" :empty-values="[null, undefined, 0]">
              <el-option label="不限" :value="3" />
              <el-option label="男" :value="1" />
              <el-option label="女" :value="2" />
            </el-select>
          </el-form-item>
        </div>

        <div class="grid grid-cols-1 gap-x-6 sm:grid-cols-2">
          <el-form-item label="招聘人数">
            <el-input-number v-model="form.amount" :min="0" :max="99" class="w-full" />
          </el-form-item>
          <el-form-item label="有效期">
            <el-date-picker
              v-model="deadlineDate"
              type="date"
              placeholder="到期日（不选默认 30 天）"
              value-format="x"
              class="w-full"
            />
          </el-form-item>
        </div>

        <!-- 三级分类级联 -->
        <el-form-item label="职位分类" required>
          <div class="flex w-full gap-2">
            <el-select v-model="form.topclass" placeholder="一级分类" class="flex-1" clearable :empty-values="[null, undefined, 0]" @change="onTopChange">
              <el-option v-for="c in level1" :key="c.id" :label="c.name" :value="c.id" />
            </el-select>
            <el-select v-model="form.category" placeholder="二级分类" class="flex-1" clearable :disabled="!form.topclass" :empty-values="[null, undefined, 0]" @change="onCatChange">
              <el-option v-for="c in level2" :key="c.id" :label="c.name" :value="c.id" />
            </el-select>
            <el-select v-model="form.subclass" placeholder="三级分类" class="flex-1" clearable :disabled="!form.category" :empty-values="[null, undefined, 0]">
              <el-option v-for="c in level3" :key="c.id" :label="c.name" :value="c.id" />
            </el-select>
          </div>
          <div v-if="!jobCategories.length" class="mt-1 text-xs text-slate-400">
            职位分类数据暂未配置（后端 seed 补「职位分类」分组后自动出现）
          </div>
        </el-form-item>

        <div class="grid grid-cols-1 gap-x-6 sm:grid-cols-2">
          <el-form-item label="所属行业">
            <el-select v-model="form.trade" placeholder="请选择" class="w-full" clearable :empty-values="[null, undefined, 0]">
              <el-option v-for="c in categories.trade" :key="c.id" :label="c.name" :value="c.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="学历要求">
            <el-select v-model="form.education" placeholder="请选择" class="w-full" clearable :empty-values="[null, undefined, 0]">
              <el-option v-for="c in categories.education" :key="c.id" :label="c.name" :value="c.id" />
            </el-select>
          </el-form-item>
        </div>

        <div class="grid grid-cols-1 gap-x-6 sm:grid-cols-2">
          <el-form-item label="经验要求">
            <el-select v-model="form.experience" placeholder="请选择" class="w-full" clearable :empty-values="[null, undefined, 0]">
              <el-option v-for="c in categories.experience" :key="c.id" :label="c.name" :value="c.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="工作地区">
            <el-input v-model="form.districtCn" placeholder="如：广州" clearable @change="form.district = form.districtCn" />
          </el-form-item>
        </div>

        <el-form-item label="职位福利">
          <el-select v-model="form.tags" multiple collapse-tags collapse-tags-tooltip placeholder="请选择福利标签" class="w-full">
            <el-option v-for="tag in jobTags" :key="tag.id" :label="tag.name" :value="tag.id" />
          </el-select>
        </el-form-item>

        <!-- 薪资 -->
        <el-form-item label="薪资范围">
          <div class="flex w-full items-center gap-2">
            <el-checkbox v-model="form.negotiable" :true-value="1" :false-value="0">面议</el-checkbox>
            <template v-if="form.negotiable !== 1">
              <el-input-number v-model="form.minwage" :min="0" :max="999999" :step="1000" placeholder="最低" class="flex-1" />
              <span class="text-slate-400">至</span>
              <el-input-number v-model="form.maxwage" :min="0" :max="999999" :step="1000" placeholder="最高" class="flex-1" />
            </template>
          </div>
        </el-form-item>

        <el-form-item label="职位描述">
          <el-input
            v-model="form.contents"
            type="textarea"
            :rows="5"
            maxlength="4000"
            show-word-limit
            placeholder="岗位职责、任职要求等"
          />
        </el-form-item>

        <el-divider content-position="left">联系方式</el-divider>
        <div class="grid grid-cols-1 gap-x-6 sm:grid-cols-2">
          <el-form-item label="联系人">
            <el-input v-model="form.contact.contact" maxlength="80" placeholder="联系人姓名" clearable />
          </el-form-item>
          <el-form-item label="联系电话">
            <el-input v-model="form.contact.telephone" maxlength="80" placeholder="联系电话" clearable />
          </el-form-item>
        </div>
        <div class="grid grid-cols-1 gap-x-6 sm:grid-cols-2">
          <el-form-item label="邮箱">
            <el-input v-model="form.contact.email" maxlength="80" placeholder="联系邮箱" clearable />
          </el-form-item>
          <el-form-item label="地址">
            <el-input v-model="form.contact.address" maxlength="80" placeholder="联系地址" clearable />
          </el-form-item>
        </div>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { createJob, deleteJob, getJob, getJobs, pauseJob, refreshJob, resumeJob, updateJob } from '@/api/jobs'
  import { getCategories } from '@/api/content'
  import { dateInputToUnix } from '@/utils/format'
  import type { Categories, CategoryItem, JobDetail, JobItem, JobsRequest } from '@/types/api'

  // 工作性质固定枚举（v6 语义：1 全职 2 兼职 3 实习）
  const jobNatures = [
    { value: 1, label: '全职' },
    { value: 2, label: '兼职' },
    { value: 3, label: '实习' }
  ]

  const loading = ref(false)
  const tableData = ref<JobItem[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)

  const categories = ref<Categories>({
    education: [],
    experience: [],
    wage: [],
    trade: [],
    district: [],
    major: [],
    sex: [],
    marriage: [],
    nature: [],
    scale: []
  })

  const getList = async () => {
    loading.value = true
    try {
      const { data } = await getJobs({ page: page.value, pageSize: pageSize.value })
      tableData.value = data.list
      total.value = data.total
    } finally {
      loading.value = false
    }
  }

  onMounted(async () => {
    getList()
    const { data } = await getCategories()
    categories.value = data
  })

  // ========== 三级分类级联 ==========
  // 职位分类走 jobcategory 分组（parentId 三级层级）；后端 seed 未补前为空数组
  const jobCategories = computed(() => categories.value.jobcategory || [])
  const level1 = computed(() => jobCategories.value.filter((c) => c.parentId === 0))
  const level2 = computed(() => jobCategories.value.filter((c) => c.parentId === form.topclass))
  const level3 = computed(() => jobCategories.value.filter((c) => c.parentId === form.category))

  // ========== 职位名称（X3：jobtitle 扁平列表，纯选择） ==========
  const jobTitles = computed(() => categories.value.jobtitle || [])
  const jobTags = computed(() => categories.value.jobtag || [])
  // 编辑回显：已存名字不在当前列表内（后期岗位被下线等）→ 显示虚拟 option 保数据
  const historyTitleOption = computed(
    () =>
      isEdit.value &&
      !!form.jobsName &&
      !jobTitles.value.some((t) => t.name === form.jobsName)
  )

  const onTopChange = () => {
    form.category = 0
    form.subclass = 0
    form.categoryCn = ''
  }
  const onCatChange = () => {
    form.subclass = 0
    form.categoryCn = ''
  }
  const onSubClassChange = () => {
    fillCategoryCn()
  }

  const fillCategoryCn = () => {
    const find = (list: CategoryItem[], id: number) => list.find((c) => c.id === id)?.name || ''
    const parts = [
      find(level1.value, form.topclass),
      find(level2.value, form.category),
      find(level3.value, form.subclass)
    ].filter(Boolean)
    form.categoryCn = parts.join('/')
  }

  const onNatureChange = () => {
    form.natureCn = jobNatures.find((n) => n.value === form.nature)?.label || ''
  }

  // ========== 发布/编辑 ==========
  const dialogVisible = ref(false)
  const saving = ref(false)
  const isEdit = ref(false)
  const editId = ref(0)
  // D1：编辑待审项时 #82 需显式带 ?pending=1 走 tmp 表重提（两表 id 重叠时避免错表）
  const editPending = ref(false)
  const detail = ref<JobDetail>({
    id: 0, jobsId: 0, jobsName: '', companyname: '', companyId: 0, emergency: 0, stick: 0,
    nature: 1, natureCn: '', sex: 3, amount: 1, topclass: 0, category: 0, subclass: 0,
    categoryCn: '', trade: 0, district: '', districtCn: '', tag: '', education: 0, experience: 0,
    minwage: 0, maxwage: 0, negotiable: 0, contents: '', addtime: 0, deadline: 0, refreshtime: 0,
    audit: 0, display: 1, click: 0, department: '', mapX: 0, mapY: 0, mapZoom: 0, pending: false,
    contact: { contact: '', qq: '', telephone: '', landlineTel: '', address: '', email: '' },
    tags: [], reason: ''
  })

  const emptyForm = (): JobsRequest => ({
    jobsName: '',
    nature: 1,
    natureCn: '全职',
    sex: 3,
    amount: 1,
    topclass: 0,
    category: 0,
    subclass: 0,
    categoryCn: '',
    trade: 0,
    district: '',
    districtCn: '',
    tag: '',
    education: 0,
    experience: 0,
    minwage: 0,
    maxwage: 0,
    negotiable: 0,
    contents: '',
    deadline: 0,
    department: '',
    mapX: 0,
    mapY: 0,
    mapZoom: 0,
    contact: { contact: '', qq: '', telephone: '', landlineTel: '', address: '', email: '' },
    tags: []
  })

  const form = reactive<JobsRequest>(emptyForm())
  const deadlineDate = ref('')

  const openCreate = () => {
    isEdit.value = false
    editId.value = 0
    editPending.value = false
    detail.value.reason = ''
    Object.assign(form, emptyForm())
    deadlineDate.value = ''
    dialogVisible.value = true
  }

  const openEdit = async (row: JobItem) => {
    // 待审项（pending=true）按 tmp 表查，缺省按 jobs 表查——两表 id 会重叠，必须透传 pending 区分
    const { data } = await getJob(row.id, row.pending)
    detail.value = data
    isEdit.value = true
    editId.value = row.id
    editPending.value = row.pending
    Object.assign(form, {
      jobsName: data.jobsName,
      nature: data.nature,
      natureCn: data.natureCn,
      sex: data.sex,
      amount: data.amount,
      topclass: data.topclass,
      category: data.category,
      subclass: data.subclass,
      categoryCn: data.categoryCn,
      trade: data.trade,
      district: data.district,
      districtCn: data.districtCn,
      tag: data.tag,
      education: data.education,
      experience: data.experience,
      minwage: data.minwage,
      maxwage: data.maxwage,
      negotiable: data.negotiable,
      contents: data.contents,
      deadline: data.deadline,
      department: data.department || '',
      mapX: data.mapX || 0,
      mapY: data.mapY || 0,
      mapZoom: data.mapZoom || 0,
      contact: data.contact || { contact: '', qq: '', telephone: '', landlineTel: '', address: '', email: '' },
      tags: data.tags || []
    })
    deadlineDate.value = ''
    dialogVisible.value = true
  }

  const handleSave = async () => {
    if (!form.jobsName.trim() || form.jobsName.trim().length < 2) {
      ElMessage.warning('职位名称需 2-50 字符')
      return
    }
    if (!form.topclass || !form.category || !form.subclass) {
      ElMessage.warning('请选择职位分类（三级必填）')
      return
    }
    if (form.negotiable !== 1 && form.maxwage <= form.minwage) {
      ElMessage.warning('最高薪资需大于最低薪资')
      return
    }
    // 08-21 PM 裁决：与后端同口径的 2 倍上限（jobs_service.go L439 隐藏规则 max/min 限制），
    // 前端先行提示，避免用户提交后吃后端「薪资范围不合法」
    if (form.negotiable !== 1 && form.minwage > 0 && form.maxwage > form.minwage * 2) {
      ElMessage.warning('最高薪资不能超过最低薪资的 2 倍')
      return
    }
    // X1：清空后的下拉为 undefined，提交前还原为 0（后端契约：数值字段 0=空）
    const toNum = (v: unknown, d = 0): number => (typeof v === 'number' && !Number.isNaN(v) ? v : d)
    form.nature = toNum(form.nature)
    form.sex = toNum(form.sex)
    form.topclass = toNum(form.topclass)
    form.category = toNum(form.category)
    form.subclass = toNum(form.subclass)
    form.trade = toNum(form.trade)
    form.education = toNum(form.education)
    form.experience = toNum(form.experience)
    form.amount = toNum(form.amount)
    form.minwage = toNum(form.minwage)
    form.maxwage = toNum(form.maxwage)
    fillCategoryCn()
    form.deadline = deadlineDate.value ? dateInputToUnix(deadlineDate.value) : 0
    saving.value = true
    try {
      if (isEdit.value) {
        // D1：显式透传 pending，避免两表 id 重叠时编辑错表（缺省回退行为也正确）
        await updateJob(editId.value, form, editPending.value)
        ElMessage.success('已提交保存')
      } else {
        await createJob(form)
        // X2：发布默认进后台审核（jobsDisplayMode 缺省=1），提示语与机制一致
        ElMessage.success('已提交，待审核')
      }
      dialogVisible.value = false
      getList()
    } finally {
      saving.value = false
    }
  }

  // ========== 列表操作 ==========
  const togglePause = async (row: JobItem, target: number) => {
    try {
      if (target === 2) {
        await pauseJob(row.id)
        ElMessage.success('已暂停')
      } else {
        await resumeJob(row.id)
        ElMessage.success('已恢复')
      }
      getList()
    } catch {
      // 错误已由拦截器弹出
    }
  }

  const handleRefresh = async (row: JobItem) => {
    await refreshJob(row.id)
    ElMessage.success('已刷新')
    getList()
  }

  const handleDelete = async (row: JobItem) => {
    try {
      await ElMessageBox.confirm('确认删除该职位？投递记录会保留', '提示', {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }
    // D1：显式透传 pending，待审项走 jobs_tmp 软删
    await deleteJob(row.id, row.pending)
    ElMessage.success('已删除')
    getList()
  }

  const formatTime = (ts?: number) => {
    if (!ts) return '-'
    const d = new Date(ts * 1000)
    const pad = (n: number) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  }
</script>
