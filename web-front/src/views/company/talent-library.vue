<template>
  <div class="mx-auto max-w-6xl px-4 py-10">
    <div class="flex items-start justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-slate-800">人才库</h1>
        <p class="mt-1 text-sm text-slate-500">管理已解锁候选人、联系方式与跟进进度</p>
      </div>
      <el-button type="primary" @click="router.push({ name: 'CompanyTalents' })">继续找人才</el-button>
    </div>

    <el-tabs v-model="activeTab" class="mt-6" @tab-change="loadCurrentTab">
      <el-tab-pane label="已解锁人才" name="unlocked">
        <el-card shadow="never">
          <el-table v-loading="loading" :data="talents">
            <template #empty>
              <el-empty :image-size="80" description="暂未解锁候选人" />
            </template>
            <el-table-column label="候选人" min-width="130">
              <template #default="{ row }">{{ row.resume.fullname }}</template>
            </el-table-column>
            <el-table-column label="求职意向" min-width="170" show-overflow-tooltip>
              <template #default="{ row }">{{ row.resume.intentionJobs || row.resume.title || '-' }}</template>
            </el-table-column>
            <el-table-column label="联系方式" min-width="210">
              <template #default="{ row }">{{ row.resume.telephone || '-' }} {{ row.resume.email || '' }}</template>
            </el-table-column>
            <el-table-column label="跟进状态" width="130" align="center">
              <template #default="{ row }">
                <el-select
                  :model-value="row.download.followUp"
                  size="small"
                  @change="(value: number) => handleFollowUp(row, value)"
                >
                  <el-option v-for="item in followUpOptions" :key="item.value" :label="item.label" :value="item.value" />
                </el-select>
              </template>
            </el-table-column>
            <el-table-column label="解锁时间" width="170" align="center">
              <template #default="{ row }">{{ formatTime(row.download.downloadedAt) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="100" align="center">
              <template #default="{ row }">
                <el-button type="primary" link @click="openInterview(row)">邀请面试</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div class="mt-4 flex justify-end">
            <el-pagination
              v-model:current-page="talentPage"
              v-model:page-size="pageSize"
              :total="talentTotal"
              :page-sizes="[10, 20, 50]"
              layout="total, sizes, prev, pager, next"
              @size-change="loadTalents"
              @current-change="loadTalents"
            />
          </div>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="我的收藏" name="favorites">
        <el-card shadow="never">
          <el-table v-loading="loading" :data="favorites">
            <template #empty>
              <el-empty :image-size="80" description="暂未收藏候选人" />
            </template>
            <el-table-column label="候选人" min-width="130">
              <template #default="{ row }">{{ row.resume.fullname }}</template>
            </el-table-column>
            <el-table-column label="求职意向" min-width="180" show-overflow-tooltip>
              <template #default="{ row }">{{ row.resume.intentionJobs || row.resume.title || '-' }}</template>
            </el-table-column>
            <el-table-column label="学历 / 经验" min-width="160">
              <template #default="{ row }">{{ row.resume.educationCn || '-' }} / {{ row.resume.experienceCn || '-' }}</template>
            </el-table-column>
            <el-table-column label="收藏时间" width="170" align="center">
              <template #default="{ row }">{{ formatTime(row.favorite.addtime) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="100" align="center">
              <template #default="{ row }">
                <el-button type="danger" link @click="handleUnfavorite(row)">取消收藏</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div class="mt-4 flex justify-end">
            <el-pagination
              v-model:current-page="favoritePage"
              v-model:page-size="pageSize"
              :total="favoriteTotal"
              :page-sizes="[10, 20, 50]"
              layout="total, sizes, prev, pager, next"
              @size-change="loadFavorites"
              @current-change="loadFavorites"
            />
          </div>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="interviewVisible" title="发起线下面试邀请" width="520px" :close-on-click-modal="false">
      <el-form :model="interviewForm" label-width="96px" @submit.prevent>
        <el-form-item label="候选人">
          <el-input :model-value="selectedTalent?.resume.fullname || ''" disabled />
        </el-form-item>
        <el-form-item label="应聘职位" required>
          <el-select v-model="interviewForm.jobsId" placeholder="选择企业在招职位" class="w-full">
            <el-option v-for="job in availableJobs" :key="job.id" :label="job.jobsName" :value="job.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="面试时间" required>
          <el-date-picker v-model="interviewTimeValue" type="datetime" value-format="x" placeholder="选择面试时间" class="w-full" />
        </el-form-item>
        <el-form-item label="面试地点" required>
          <el-input v-model="interviewForm.address" maxlength="200" clearable />
        </el-form-item>
        <el-form-item label="联系人" required>
          <el-input v-model="interviewForm.contact" maxlength="30" clearable />
        </el-form-item>
        <el-form-item label="联系电话" required>
          <el-input v-model="interviewForm.telephone" maxlength="30" clearable />
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
  import { useRouter } from 'vue-router'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { getJobs } from '@/api/jobs'
  import { createInterview } from '@/api/interview'
  import { getCompanyTalents, getFavoriteTalents, setTalentFollowUp, unfavoriteTalent } from '@/api/talent'
  import type { CompanyTalentItem, FavoriteTalentItem } from '@/api/talent'
  import type { InterviewCreate } from '@/api/interview'
  import type { JobItem } from '@/types/api'
  import { formatTime } from '@/utils/format'

  const router = useRouter()
  const loading = ref(false)
  const activeTab = ref('unlocked')
  const pageSize = ref(10)
  const talentPage = ref(1)
  const talentTotal = ref(0)
  const talents = ref<CompanyTalentItem[]>([])
  const favoritePage = ref(1)
  const favoriteTotal = ref(0)
  const favorites = ref<FavoriteTalentItem[]>([])
  const interviewVisible = ref(false)
  const inviting = ref(false)
  const selectedTalent = ref<CompanyTalentItem | null>(null)
  const availableJobs = ref<JobItem[]>([])
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
  const followUpOptions = [
    { value: 0, label: '待跟进' },
    { value: 1, label: '合适' },
    { value: 2, label: '不合适' },
    { value: 3, label: '待定' },
    { value: 4, label: '未接通' }
  ]

  const loadTalents = async () => {
    loading.value = true
    try {
      const { data } = await getCompanyTalents({ page: talentPage.value, pageSize: pageSize.value })
      talents.value = data.list
      talentTotal.value = data.total
    } finally {
      loading.value = false
    }
  }

  const loadFavorites = async () => {
    loading.value = true
    try {
      const { data } = await getFavoriteTalents({ page: favoritePage.value, pageSize: pageSize.value })
      favorites.value = data.list
      favoriteTotal.value = data.total
    } finally {
      loading.value = false
    }
  }

  const loadCurrentTab = () => {
    if (activeTab.value === 'favorites') {
      loadFavorites()
      return
    }
    loadTalents()
  }

  onMounted(loadTalents)

  const handleFollowUp = async (row: CompanyTalentItem, followUp: number) => {
    await setTalentFollowUp(row.resume.id, followUp)
    ElMessage.success('跟进状态已更新')
    await loadTalents()
  }

  const handleUnfavorite = async (row: FavoriteTalentItem) => {
    try {
      await ElMessageBox.confirm('确认取消收藏该候选人？', '取消收藏', {
        confirmButtonText: '取消收藏',
        cancelButtonText: '保留',
        type: 'warning'
      })
    } catch {
      return
    }
    await unfavoriteTalent(row.favorite.id)
    ElMessage.success('已取消收藏')
    await loadFavorites()
  }

  const openInterview = async (row: CompanyTalentItem) => {
    const { data } = await getJobs({ page: 1, pageSize: 100 })
    availableJobs.value = data.list.filter((job) => !job.pending && job.audit === 1 && job.display === 1)
    if (availableJobs.value.length === 0) {
      ElMessage.warning('请先发布并审核通过至少一个在招职位')
      return
    }
    selectedTalent.value = row
    Object.assign(interviewForm, {
      resumeId: row.resume.id,
      jobsId: 0,
      interviewTime: 0,
      address: '',
      contact: '',
      telephone: '',
      notes: ''
    })
    interviewTimeValue.value = ''
    interviewVisible.value = true
  }

  const submitInterview = async () => {
    if (!interviewForm.jobsId || !interviewTimeValue.value || !interviewForm.address.trim() || !interviewForm.contact.trim() || !interviewForm.telephone.trim()) {
      ElMessage.warning('请完整填写面试信息')
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
</script>
