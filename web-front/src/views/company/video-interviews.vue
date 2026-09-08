<template>
  <div class="mx-auto max-w-4xl px-4 py-16">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-800">视频面试</h1>
      <el-button type="primary" @click="openCreate">发起视频面试</el-button>
    </div>
    <p class="mt-2 text-sm text-slate-500">向候选人发起视频面试邀请，双方凭房间码进入面试房间</p>

    <!-- 列表 -->
    <el-card class="mt-6">
      <div v-if="loading && tableData.length === 0" class="space-y-4 py-2">
        <el-skeleton v-for="i in 4" :key="i" :rows="1" animated />
      </div>
      <el-table v-else v-loading="loading" :data="tableData">
        <template #empty>
          <el-empty :image-size="80" description="还没有发起视频面试" />
        </template>
        <el-table-column prop="id" label="ID" width="70" align="center" />
        <el-table-column label="候选人" width="110">
          <template #default="{ row }">{{ row.fullname || '-' }}</template>
        </el-table-column>
        <el-table-column label="职位" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">{{ row.jobsName || '-' }}</template>
        </el-table-column>
        <el-table-column label="面试时间" width="170">
          <template #default="{ row }">{{ formatTime(row.interviewTime) }}</template>
        </el-table-column>
        <el-table-column label="房间状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="roomTag(row.roomStatus)">{{ roomStatusCn(row.roomStatus) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="企业房间码" width="130" align="center">
          <template #default="{ row }">
            <span v-if="row.companyCode" class="font-mono">{{ row.companyCode }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" align="center">
          <template #default="{ row }">
            <el-button type="primary" link @click="showDetail(row)">详情</el-button>
            <el-button type="danger" link @click="remove(row)">删除</el-button>
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

    <!-- 发起弹窗 -->
    <el-dialog v-model="createVisible" title="发起视频面试" width="480px" :close-on-click-modal="false">
      <el-form :model="createForm" label-width="90px" @submit.prevent>
        <el-form-item label="简历 ID" required>
          <el-input v-model.number="createForm.resumeId" placeholder="候选人简历 id" clearable />
        </el-form-item>
        <el-form-item label="职位 ID" required>
          <el-input v-model.number="createForm.jobsId" placeholder="关联职位 id" clearable />
        </el-form-item>
        <el-form-item label="职位名称" required>
          <el-input v-model="createForm.jobsName" maxlength="30" placeholder="职位名（快照）" clearable />
        </el-form-item>
        <el-form-item label="面试时间" required>
          <el-date-picker
            v-model="interviewTimeDate"
            type="datetime"
            placeholder="选择面试时间"
            value-format="x"
            class="w-full"
          />
        </el-form-item>
        <el-form-item label="联系人">
          <el-input v-model="createForm.contact" maxlength="30" placeholder="联系人姓名" clearable />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="createForm.telephone" maxlength="30" placeholder="联系电话" clearable />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="submitCreate">发起</el-button>
      </template>
    </el-dialog>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="视频面试详情" width="440px">
      <div v-if="detail.id" class="space-y-2 text-sm text-slate-600">
        <p><span class="font-semibold">候选人：</span>{{ detail.fullname || '-' }}</p>
        <p><span class="font-semibold">职位：</span>{{ detail.jobsName || '-' }}</p>
        <p><span class="font-semibold">面试时间：</span>{{ formatTime(detail.interviewTime) }}</p>
        <p><span class="font-semibold">房间状态：</span>{{ roomStatusCn(detail.roomStatus) }}</p>
        <p>
          <span class="font-semibold">企业房间码：</span>
          <span class="font-mono">{{ detail.companyCode || '-' }}</span>
        </p>
        <p>
          <span class="font-semibold">个人房间码：</span>
          <span class="font-mono">{{ detail.personalCode || '-' }}</span>
        </p>
        <p><span class="font-semibold">联系人：</span>{{ detail.contact || '-' }} {{ detail.contactTel || '' }}</p>
        <el-button
          type="primary"
          class="mt-4"
          :disabled="detail.roomStatus !== 'opened'"
          @click="enterRoom(detail.companyCode)"
        >
          进入面试房间
        </el-button>
        <p v-if="detail.roomStatus !== 'opened'" class="mt-2 text-xs text-slate-400">
          仅面试当天（opened）可进入房间
        </p>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, reactive, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    createVideoInterview,
    deleteCompanyVideoInterview,
    getCompanyVideoInterview,
    getCompanyVideoInterviews
  } from '@/api/videoInterview'
  import { formatTime } from '@/utils/format'
  import type { VideoInterview, VideoInterviewCreate } from '@/types/api'

  const router = useRouter()

  const loading = ref(false)
  const tableData = ref<VideoInterview[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)

  const getList = async () => {
    loading.value = true
    try {
      const { data } = await getCompanyVideoInterviews({
        page: page.value,
        pageSize: pageSize.value
      })
      tableData.value = data.list
      total.value = data.total
    } finally {
      loading.value = false
    }
  }

  onMounted(getList)

  // 发起弹窗
  const createVisible = ref(false)
  const creating = ref(false)
  const createForm = reactive<VideoInterviewCreate>({
    resumeId: 0,
    jobsId: 0,
    jobsName: '',
    interviewTime: 0,
    contact: '',
    telephone: ''
  })
  const interviewTimeDate = ref('')

  const openCreate = () => {
    Object.assign(createForm, {
      resumeId: 0,
      jobsId: 0,
      jobsName: '',
      interviewTime: 0,
      contact: '',
      telephone: ''
    })
    interviewTimeDate.value = ''
    createVisible.value = true
  }

  const submitCreate = async () => {
    if (!createForm.resumeId || !createForm.jobsId || !createForm.jobsName.trim() || !interviewTimeDate.value) {
      ElMessage.warning('请填写简历 ID、职位 ID、职位名称、面试时间')
      return
    }
    creating.value = true
    try {
      // el-date-picker value-format="x" 返回的是毫秒字符串，接口要 unix 秒
      createForm.interviewTime = Math.floor(Number(interviewTimeDate.value) / 1000)
      const { data } = await createVideoInterview(createForm)
      ElMessage.success(`已发起，企业房间码：${data.companyCode}`)
      createVisible.value = false
      getList()
    } finally {
      creating.value = false
    }
  }

  // 详情弹窗
  const detailVisible = ref(false)
  const detail = ref<Partial<VideoInterview>>({})
  const showDetail = async (row: VideoInterview) => {
    const { data } = await getCompanyVideoInterview(row.id as number)
    detail.value = data
    detailVisible.value = true
  }

  const remove = async (row: VideoInterview) => {
    try {
      await ElMessageBox.confirm('确认删除这条面试邀请？', '提示', {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }
    await deleteCompanyVideoInterview(row.id as number)
    ElMessage.success('已删除')
    getList()
  }

  const enterRoom = (code?: string) => {
    if (code) {
      router.push(`/video-interviews/room/${code}`)
    }
  }

  const roomStatusCn = (s?: string) => {
    const map: Record<string, string> = { nostart: '未开始', opened: '可进入', overtime: '已过期' }
    return map[s || ''] || s || '-'
  }
  const roomTag = (s?: string) => {
    const map: Record<string, string> = { nostart: 'info', opened: 'success', overtime: 'warning' }
    return map[s || ''] || 'info'
  }
</script>
