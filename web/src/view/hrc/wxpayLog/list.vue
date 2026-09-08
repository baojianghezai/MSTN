<template>
  <div>
    <!-- 搜索区 -->
    <div class="gva-search-box">
      <el-select v-model="search.status" placeholder="状态筛选" class="w-36">
        <el-option label="全部" :value="null" />
        <el-option label="成功" :value="1" />
        <el-option label="失败" :value="0" />
      </el-select>
      <el-button type="primary" @click="handleSearch">查询</el-button>
      <el-button @click="handleReset">重置</el-button>
    </div>

    <!-- 表格区 -->
    <div class="gva-table-box">
      <el-table v-loading="loading" :data="tableData" border>
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column prop="tradeNo" label="交易号" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.tradeNo || '-' }}</template>
        </el-table-column>
        <el-table-column prop="openid" label="openid" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.openid || '-' }}</template>
        </el-table-column>
        <el-table-column prop="amount" label="金额" width="100" align="right">
          <template #default="{ row }">{{ row.amount }}</template>
        </el-table-column>
        <el-table-column label="时间" width="170" align="center">
          <template #default="{ row }">{{ formatTime(row.addtime) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="失败原因" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.failReason || '-' }}</template>
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
  import { getWxpayLogs } from '@/api/hrc/wxpayLog'

  defineOptions({
    name: 'WxpayLogList'
  })

  const loading = ref(false)
  const tableData = ref([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)

  // status: null=全部（不传参数）；0=失败 1=成功
  const search = reactive({ status: null })

  const getList = async () => {
    loading.value = true
    try {
      const params = { page: page.value, pageSize: pageSize.value }
      if (search.status !== null) {
        params.status = search.status
      }
      const { data } = await getWxpayLogs(params)
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
    search.status = null
    page.value = 1
    getList()
  }

  const formatTime = (ts) => {
    if (!ts) return '-'
    const d = new Date(ts * 1000)
    const pad = (n) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  }

  getList()
</script>
