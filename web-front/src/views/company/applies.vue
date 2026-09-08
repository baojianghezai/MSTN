<template>
  <div class="mx-auto max-w-5xl px-4 py-16">
    <h1 class="text-2xl font-bold text-slate-800">收到的简历</h1>
    <p class="mt-2 text-sm text-slate-500">查看求职者投递，标记已看并回复</p>

    <!-- 状态筛选 -->
    <div class="mt-6 flex items-center gap-3">
      <el-radio-group v-model="status" @change="handleSearch">
        <el-radio-button :value="0">全部</el-radio-button>
        <el-radio-button :value="1">未读</el-radio-button>
        <el-radio-button :value="2">已读</el-radio-button>
      </el-radio-group>
    </div>

    <el-card class="mt-4">
      <div v-if="loading && tableData.length === 0" class="space-y-4 py-2">
        <el-skeleton v-for="i in 4" :key="i" :rows="1" animated />
      </div>
      <el-table v-else v-loading="loading" :data="tableData">
        <template #empty>
          <el-empty :image-size="80" description="还没有收到简历投递" />
        </template>
        <el-table-column label="简历" min-width="120" show-overflow-tooltip>
          <template #default="{ row }">{{ row.resumeName }}</template>
        </el-table-column>
        <el-table-column label="投递职位" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.jobsName }}</template>
        </el-table-column>
        <el-table-column label="投递时间" width="170" align="center">
          <template #default="{ row }">{{ formatTime(row.applyAddtime) }}</template>
        </el-table-column>
        <el-table-column label="附言" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">{{ row.notes || '-' }}</template>
        </el-table-column>
        <el-table-column label="已读" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.personalLook === 2 ? 'success' : 'info'" size="small">
              {{ row.personalLook === 2 ? '已读' : '未读' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="回复状态" width="130" align="center">
          <template #default="{ row }">
            <el-select
              :model-value="row.isReply"
              size="small"
              class="w-24"
              @change="(v: number) => handleReply(row, v)"
            >
              <el-option v-for="o in replyOptions" :key="o.value" :label="o.label" :value="o.value" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250" align="center">
          <template #default="{ row }">
            <el-button v-if="row.personalLook !== 2" type="primary" link @click="handleLooked(row)">标记已看</el-button>
            <el-button type="warning" link @click="openInterview(row)">邀请面试</el-button>
            <el-button type="primary" link @click="handleDownload(row)">下载简历</el-button>
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

    <el-dialog v-model="interviewVisible" title="发起线下面试邀请" width="520px" :close-on-click-modal="false">
      <el-form :model="interviewForm" label-width="96px" @submit.prevent>
        <el-form-item label="候选人">
          <el-input :model-value="interviewCandidate" disabled />
        </el-form-item>
        <el-form-item label="应聘职位">
          <el-input :model-value="interviewJob" disabled />
        </el-form-item>
        <el-form-item label="面试时间" required>
          <el-date-picker
            v-model="interviewTimeValue"
            type="datetime"
            value-format="x"
            placeholder="选择面试时间"
            class="w-full"
          />
        </el-form-item>
        <el-form-item label="面试地点" required>
          <el-input v-model="interviewForm.address" maxlength="200" placeholder="请填写详细地址" clearable />
        </el-form-item>
        <el-form-item label="联系人" required>
          <el-input v-model="interviewForm.contact" maxlength="30" placeholder="联系人姓名" clearable />
        </el-form-item>
        <el-form-item label="联系电话" required>
          <el-input v-model="interviewForm.telephone" maxlength="30" placeholder="联系电话" clearable />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="interviewForm.notes" type="textarea" :rows="3" maxlength="500" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="interviewVisible = false">取消</el-button>
        <el-button type="primary" :loading="inviting" @click="submitInterview">发送邀请</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, reactive, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { downloadResume, getCompanyApplies, markLooked, replyApply } from '@/api/companyApply'
  import type { CompanyApplyItem } from '@/api/companyApply'
  import { createInterview } from '@/api/interview'
  import type { InterviewCreate } from '@/api/interview'

  const replyOptions = [
    { value: 0, label: '待反馈' },
    { value: 1, label: '合适' },
    { value: 2, label: '不合适' },
    { value: 3, label: '待定' },
    { value: 4, label: '未接通' }
  ]

  const loading = ref(false)
  const tableData = ref<CompanyApplyItem[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)
  const status = ref(0)
  const interviewVisible = ref(false)
  const inviting = ref(false)
  const interviewCandidate = ref('')
  const interviewJob = ref('')
  const interviewTimeValue = ref('')
  const interviewForm = reactive<InterviewCreate>({
    resumeId: 0,
    jobsId: 0,
    interviewTime: 0,
    address: '',
    contact: '',
    telephone: '',
    notes: ''
  })

  const getList = async () => {
    loading.value = true
    try {
      const { data } = await getCompanyApplies({
        page: page.value,
        pageSize: pageSize.value,
        status: status.value
      })
      tableData.value = data.list
      total.value = data.total
    } finally {
      loading.value = false
    }
  }

  onMounted(getList)

  const handleSearch = () => {
    page.value = 1
    getList()
  }

  const handleLooked = async (row: CompanyApplyItem) => {
    await markLooked(row.did)
    ElMessage.success('已标记已看')
    getList()
  }

  const handleReply = async (row: CompanyApplyItem, isReply: number) => {
    await replyApply(row.did, isReply)
    ElMessage.success('回复状态已更新')
    getList()
  }

  const handleDownload = async (row: CompanyApplyItem) => {
    try {
      await downloadResume(row.did)
      ElMessage.success('简历已开始下载')
      if (row.personalLook !== 2) {
        await getList()
      }
    } catch {
      // 下载接口会自行显示业务错误
    }
  }

  const openInterview = (row: CompanyApplyItem) => {
    Object.assign(interviewForm, {
      resumeId: row.resumeId,
      jobsId: row.jobsId,
      interviewTime: 0,
      address: '',
      contact: '',
      telephone: '',
      notes: ''
    })
    interviewCandidate.value = row.resumeName
    interviewJob.value = row.jobsName
    interviewTimeValue.value = ''
    interviewVisible.value = true
  }

  const submitInterview = async () => {
    if (!interviewTimeValue.value || !interviewForm.address.trim() || !interviewForm.contact.trim() || !interviewForm.telephone.trim()) {
      ElMessage.warning('请填写面试时间、地点、联系人和联系电话')
      return
    }
    interviewForm.interviewTime = Math.floor(Number(interviewTimeValue.value) / 1000)
    inviting.value = true
    try {
      await createInterview(interviewForm)
      ElMessage.success('面试邀请已发送')
      interviewVisible.value = false
    } finally {
      inviting.value = false
    }
  }

  const formatTime = (ts?: number) => {
    if (!ts) return '-'
    const d = new Date(ts * 1000)
    const pad = (n: number) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  }
</script>
