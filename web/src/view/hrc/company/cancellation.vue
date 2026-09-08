<template>
  <div>
    <!-- 搜索区 -->
    <div class="gva-search-box">
      <el-select v-model="search.status" placeholder="状态筛选" class="w-36">
        <el-option label="全部" :value="0" />
        <el-option label="待处理" :value="1" />
        <el-option label="已处理" :value="2" />
      </el-select>
      <el-button type="primary" @click="handleSearch">查询</el-button>
      <el-button @click="handleReset">重置</el-button>
    </div>

    <!-- 表格区 -->
    <div class="gva-table-box">
      <el-table v-loading="loading" :data="tableData" border>
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column prop="companyname" label="企业名称" min-width="180" show-overflow-tooltip />
        <el-table-column prop="username" label="会员账号" width="140" show-overflow-tooltip />
        <el-table-column prop="mobile" label="手机号" width="130" />
        <el-table-column label="提交时间" width="170" align="center">
          <template #default="{ row }">{{ formatTime(row.addtime) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 0 ? 'warning' : 'success'">{{ row.statusCn }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" align="center">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 0"
              type="primary"
              link
              @click="handleProcess(row)"
            >
              处理
            </el-button>
            <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
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
  </div>
</template>

<script setup>
  import { reactive, ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    deleteCompanyCancellation,
    getCompanyCancellations,
    handleCompanyCancellation
  } from '@/api/hrc/company'

  defineOptions({
    name: 'CompanyCancellation'
  })

  const loading = ref(false)
  const tableData = ref([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)

  const search = reactive({ status: 0 })

  const getList = async () => {
    loading.value = true
    try {
      const { data } = await getCompanyCancellations({
        page: page.value,
        pageSize: pageSize.value,
        status: search.status
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
    search.status = 0
    page.value = 1
    getList()
  }

  // 处理：清除该企业全部业务数据，账号保留（二次确认）
  const handleProcess = async (row) => {
    try {
      await ElMessageBox.confirm(
        `确认处理「${row.companyname}」的注销申请？将清除该企业全部业务数据，会员账号保留。`,
        '提示',
        { confirmButtonText: '确认处理', cancelButtonText: '取消', type: 'warning' }
      )
    } catch {
      return
    }
    await handleCompanyCancellation(row.id)
    ElMessage.success('已处理')
    getList()
  }

  // 删除：硬删除申请记录（二次确认）
  const handleDelete = async (row) => {
    try {
      await ElMessageBox.confirm('确认删除该注销申请记录？删除后不可恢复。', '提示', {
        confirmButtonText: '确认删除',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }
    await deleteCompanyCancellation(row.id)
    ElMessage.success('已删除')
    getList()
  }

  // 后端 addtime 是 unix 秒，需先乘 1000 转毫秒再交给 Date
  const formatTime = (ts) => {
    if (!ts) return '-'
    const d = new Date(ts * 1000)
    const pad = (n) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  }

  getList()
</script>
