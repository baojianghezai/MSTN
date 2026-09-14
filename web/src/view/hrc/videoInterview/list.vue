<template>
  <div>
    <!-- 搜索区 -->
    <div class="gva-search-box">
      <el-input
        v-model="search.keyword"
        placeholder="按职位名/公司名/简历姓名搜索"
        clearable
        class="w-64"
        @keyup.enter="handleSearch"
      />
      <el-button type="primary" @click="handleSearch">查询</el-button>
      <el-button @click="handleReset">重置</el-button>
    </div>

    <!-- 表格区 -->
    <div class="gva-table-box">
      <el-table v-loading="loading" :data="tableData" border>
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column prop="companyname" label="企业名称" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.companyname || '-' }}</template>
        </el-table-column>
        <el-table-column prop="fullname" label="简历姓名" width="110">
          <template #default="{ row }">{{ row.fullname || '-' }}</template>
        </el-table-column>
        <el-table-column prop="jobsName" label="职位" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">{{ row.jobsName || '-' }}</template>
        </el-table-column>
        <el-table-column label="面试时间" width="170" align="center">
          <template #default="{ row }">{{ formatTime(row.interviewTime) }}</template>
        </el-table-column>
        <el-table-column label="房间状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="roomTag(row.roomStatus)">{{ roomStatusCn(row.roomStatus) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="contact" label="联系人" width="100">
          <template #default="{ row }">{{ row.contact || '-' }}</template>
        </el-table-column>
        <el-table-column prop="contactTel" label="联系电话" width="130">
          <template #default="{ row }">{{ row.contactTel || '-' }}</template>
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
  </div>
</template>

<script setup>
  import { reactive, ref } from 'vue'
  import { getVideoInterviews } from '@/api/hrc/videoInterview'

  defineOptions({
    name: 'VideoInterviewList'
  })

  const loading = ref(false)
  const tableData = ref([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)

  const search = reactive({ keyword: '' })

  const getList = async () => {
    loading.value = true
    try {
      const { data } = await getVideoInterviews({
        page: page.value,
        pageSize: pageSize.value,
        keyword: search.keyword.trim()
      })
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
    search.keyword = ''
    page.value = 1
    getList()
  }

  // 房间状态机（后端计算，前端只读）：nostart 未到 / opened 当天可进 / overtime 已过期
  const roomStatusCn = (s) => {
    const map = { nostart: '未开始', opened: '可进入', overtime: '已过期' }
    return map[s] || s || '-'
  }
  const roomTag = (s) => {
    const map = { nostart: 'info', opened: 'success', overtime: 'warning' }
    return map[s] || 'info'
  }

  const formatTime = (ts) => {
    if (!ts) return '-'
    const d = typeof ts === 'number' ? new Date(ts * 1000) : new Date(ts)
    const pad = (n) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  }

  getList()
</script>
