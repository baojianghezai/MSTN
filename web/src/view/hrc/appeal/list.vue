<template>
  <div>
    <!-- 搜索区 -->
    <div class="gva-search-box">
      <el-select v-model="search.status" placeholder="状态筛选" class="w-36">
        <el-option label="全部" :value="0" />
        <el-option label="已处理" :value="1" />
        <el-option label="已驳回" :value="2" />
      </el-select>
      <el-input
        v-model="search.mobile"
        placeholder="按手机号搜索"
        clearable
        class="w-48"
        @keyup.enter="handleSearch"
      />
      <el-button type="primary" @click="handleSearch">查询</el-button>
      <el-button @click="handleReset">重置</el-button>
    </div>

    <!-- 表格区 -->
    <div class="gva-table-box">
      <el-table v-loading="loading" :data="tableData" border>
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column prop="mobile" label="手机号" width="130" />
        <el-table-column prop="realname" label="姓名" width="110">
          <template #default="{ row }">{{ row.realname || '-' }}</template>
        </el-table-column>
        <el-table-column prop="email" label="邮箱" min-width="170" show-overflow-tooltip>
          <template #default="{ row }">{{ row.email || '-' }}</template>
        </el-table-column>
        <el-table-column prop="description" label="申诉内容" min-width="240" show-overflow-tooltip />
        <el-table-column label="提交时间" width="170" align="center">
          <template #default="{ row }">{{ formatTime(row.addtime) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)">{{ row.statusCn }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="110" align="center">
          <template #default="{ row }">
            <el-button type="primary" link @click="openDialog(row)">处理</el-button>
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

    <!-- 处理弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      title="处理申诉"
      width="480px"
      :close-on-click-modal="false"
    >
      <!-- 申诉信息（只读展示） -->
      <div class="mb-4 rounded bg-gray-50 p-3 text-sm leading-6">
        <p>手机号：{{ dialogRow.mobile }}</p>
        <p>姓名：{{ dialogRow.realname || '-' }}</p>
        <p class="break-all">申诉内容：{{ dialogRow.description }}</p>
      </div>

      <el-form label-width="80px">
        <el-form-item label="处理结果">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">已处理</el-radio>
            <el-radio :value="2">已驳回</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="恢复账号">
          <el-checkbox v-model="form.restore" :disabled="form.status !== 1">
            恢复该手机号对应的注销账号
          </el-checkbox>
          <div class="w-full text-xs text-gray-400">
            仅「已处理」可勾选；勾选后按申诉手机号匹配冷静期内的注销账号并恢复
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { reactive, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import { getAppealList, processAppeal } from '@/api/hrc/appeal'

  defineOptions({
    name: 'AppealList'
  })

  const loading = ref(false)
  const tableData = ref([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)

  // 搜索条件（status: 0 全部 / 1 已处理 / 2 已驳回，与接口契约一致）
  const search = reactive({
    status: 0,
    mobile: ''
  })

  // 拉取列表：接口返回 {code, message, data}，requestHrc 拦截器已剥离外层，这里直接拿 data
  const getList = async () => {
    loading.value = true
    try {
      const { data } = await getAppealList({
        page: page.value,
        pageSize: pageSize.value,
        status: search.status,
        mobile: search.mobile.trim()
      })
      tableData.value = data.list
      total.value = data.total
    } finally {
      loading.value = false
    }
  }

  // 查询：回到第一页再拉数据
  const handleSearch = () => {
    page.value = 1
    getList()
  }

  // 重置：清空条件后回到第一页
  const handleReset = () => {
    search.status = 0
    search.mobile = ''
    page.value = 1
    getList()
  }

  // 状态数字 → el-tag 颜色
  const statusTag = (status) => {
    const map = { 0: 'warning', 1: 'success', 2: 'danger' }
    return map[status] || 'info'
  }

  // 兼容 ISO8601 字符串和 unix 秒数
  const formatTime = (ts) => {
    if (!ts) return '-'
    const d = typeof ts === 'number' ? new Date(ts * 1000) : new Date(ts)
    const pad = (n) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  }

  // ========== 处理弹窗 ==========

  const dialogVisible = ref(false)
  const submitting = ref(false)
  const dialogRow = ref({}) // 当前处理的申诉行（弹窗内只读展示）
  const form = reactive({
    status: 1, // 1 已处理 / 2 已驳回
    restore: false // 是否恢复注销账号（仅 status=1 生效）
  })

  // 打开弹窗：带上行数据，表单用默认值（避免上次残留）
  const openDialog = (row) => {
    dialogRow.value = row
    form.status = 1
    form.restore = false
    dialogVisible.value = true
  }

  // 切到「已驳回」时清掉恢复勾选（后端只有 status=1 才恢复）
  watch(
    () => form.status,
    (val) => {
      if (val === 2) {
        form.restore = false
      }
    }
  )

  // 提交处理：PUT /api/v1/admin/appeals/{id}
  const handleSubmit = async () => {
    submitting.value = true
    try {
      await processAppeal(dialogRow.value.id, {
        status: form.status,
        restore: form.restore
      })
      ElMessage.success('处理成功')
      dialogVisible.value = false
      getList() // 刷新列表
    } finally {
      submitting.value = false
    }
  }

  // 页面加载即拉一次列表
  getList()
</script>
