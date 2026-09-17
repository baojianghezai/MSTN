<template>
  <div class="mx-auto max-w-4xl px-4 py-8">
    <el-card v-loading="loading" shadow="never" class="overflow-hidden">
      <template v-if="job.id">
        <!-- 头部 -->
        <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div class="min-w-0">
            <div class="flex items-center gap-2">
              <h1 class="text-2xl font-bold text-slate-800">{{ job.jobsName }}</h1>
              <el-tag v-if="job.emergency === 1" type="danger" size="small">紧急</el-tag>
              <el-tag v-if="job.stick === 1" type="warning" size="small">置顶</el-tag>
            </div>
            <p class="mt-1 text-sm text-slate-600">{{ job.companyname }}</p>
          </div>
          <span class="flex-none rounded-md bg-accent-50 px-3 py-2 text-xl font-semibold text-accent-600">{{ salaryText }}</span>
        </div>

        <!-- 基础信息 -->
        <div class="mt-4 flex flex-wrap gap-x-6 gap-y-2 text-sm text-slate-600">
          <span v-if="job.natureCn">工作性质：{{ job.natureCn }}</span>
          <span v-if="job.districtCn">工作地区：{{ job.districtCn }}</span>
          <span v-if="job.categoryCn">职位分类：{{ job.categoryCn }}</span>
          <span>招聘人数：{{ job.amount }} 人</span>
          <span v-if="eduName">学历：{{ eduName }}</span>
          <span v-if="expName">经验：{{ expName }}</span>
          <span v-if="job.department">部门：{{ job.department }}</span>
        </div>

        <el-divider />

        <!-- 职位描述 -->
        <h2 class="font-semibold text-slate-800">职位描述</h2>
        <div class="mt-2 whitespace-pre-wrap text-sm leading-6 text-slate-600">
          {{ job.contents || '暂无描述' }}
        </div>

        <!-- 福利待遇（tag 存的是分类编码，需映射为名称） -->
        <div v-if="tagNames.length" class="mt-4">
          <h2 class="text-sm font-semibold text-slate-800">福利待遇</h2>
          <div class="mt-2 flex flex-wrap gap-2">
            <el-tag v-for="t in tagNames" :key="t" size="small" type="info">{{ t }}</el-tag>
          </div>
        </div>

        <!-- 联系方式 -->
        <el-divider />
        <h2 class="font-semibold text-slate-800">联系方式</h2>
        <div class="mt-2 space-y-1 text-sm text-slate-600">
          <p v-if="job.contact?.contact">联系人：{{ job.contact.contact }}</p>
          <p v-if="job.contact?.telephone">电话：{{ job.contact.telephone }}</p>
          <p v-if="job.contact?.email">邮箱：{{ job.contact.email }}</p>
          <p v-if="job.contact?.address">地址：{{ job.contact.address }}</p>
          <p v-if="!hasContact" class="text-slate-400">暂无联系方式</p>
        </div>

        <!-- 投递按钮 -->
        <div class="mt-7 flex justify-center gap-3 border-t border-slate-100 pt-6">
          <el-button
            type="primary"
            size="large"
            class="min-w-36"
            :loading="checkingResume"
            @click="handleApply"
          >
            <span v-if="!checkingResume" class="mr-1 i-lucide-send" aria-hidden="true" />
            投递简历
          </el-button>
          <el-button size="large" class="min-w-36" @click="handleChat">
            <span class="mr-1 i-lucide-message-circle" aria-hidden="true" />
            在线沟通
          </el-button>
        </div>
      </template>

      <el-empty v-else-if="!loading" description="职位不存在或已下架" />
    </el-card>

    <!-- 投递弹窗 -->
    <el-dialog v-model="applyVisible" title="投递职位" width="440px" :close-on-click-modal="false">
      <p class="text-sm text-slate-600">投递职位：<span class="font-semibold">{{ job.jobsName }}</span></p>
      <el-form label-width="80px" class="mt-4" @submit.prevent>
        <el-form-item label="投递简历">
          <el-select v-model="selectedResumeId" placeholder="请选择简历" style="width: 100%">
            <el-option
              v-for="r in resumeList"
              :key="r.id"
              :label="`${r.title || '未命名简历'}（完善度 ${r.completePercent}%）`"
              :value="r.id"
              :disabled="r.completePercent < MIN_APPLY_RESUME_PERCENT"
            />
          </el-select>
          <p class="mt-1 text-xs text-slate-400">
            完善度不足 {{ MIN_APPLY_RESUME_PERCENT }}% 的简历不可投递，可选择其他简历
          </p>
        </el-form-item>
        <el-form-item label="附言">
          <el-input
            v-model="applyForm.notes"
            type="textarea"
            :rows="3"
            placeholder="选填，如：随时到岗"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="applyVisible = false">取消</el-button>
        <el-button type="primary" :loading="applying" @click="submitApply">确认投递</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="noResumeVisible" title="暂无简历" width="400px">
      <el-empty :image-size="72" description="请先制作一份简历，再投递职位" />
      <template #footer>
        <el-button @click="noResumeVisible = false">取消</el-button>
        <el-button type="primary" @click="goCreateResume">去制作简历</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="incompleteResumeVisible" title="简历待完善" width="420px">
      <el-empty :image-size="72" description="默认简历信息较少，暂时不能投递">
        <p class="text-sm text-slate-500">
          当前完善度 {{ resumeToComplete?.completePercent ?? 0 }}%，完善至 {{ MIN_APPLY_RESUME_PERCENT }}% 后即可投递
        </p>
      </el-empty>
      <template #footer>
        <el-button @click="incompleteResumeVisible = false">取消</el-button>
        <el-button type="primary" @click="goCompleteResume">去完善简历</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { ElMessage } from 'element-plus'
  import { getPublicJob } from '@/api/jobs'
  import { getCategories } from '@/api/content'
  import { apply } from '@/api/apply'
  import { listResumes } from '@/api/resume'
  import { useUserStore } from '@/stores/user'
  import type { Categories, JobDetail, ResumeLite } from '@/types/api'

  const MIN_APPLY_RESUME_PERCENT = 40

  const route = useRoute()
  const router = useRouter()
  const userStore = useUserStore()

  const jobId = Number(route.params.id)
  const loading = ref(false)
  const job = ref<Partial<JobDetail>>({})
  const categories = ref<Categories>({
    education: [], experience: [], wage: [], trade: [], district: [], major: [],
    sex: [], marriage: [], nature: [], scale: []
  })

  onMounted(async () => {
    loading.value = true
    try {
      const [jobRes, catRes] = await Promise.all([getPublicJob(jobId), getCategories()])
      job.value = jobRes.data
      categories.value = catRes.data
    } catch {
      job.value = {}
    } finally {
      loading.value = false
    }
  })

  const salaryText = computed(() => {
    if (!job.value.id) return ''
    if (job.value.negotiable === 1) return '面议'
    return `${job.value.minwage}-${job.value.maxwage} 元`
  })

  const hasContact = computed(() =>
    Boolean(job.value.contact && (job.value.contact.contact || job.value.contact.telephone || job.value.contact.email || job.value.contact.address))
  )

  const eduName = computed(() => categories.value.education.find((c) => c.id === job.value.education)?.name || '')
  const expName = computed(() => categories.value.experience.find((c) => c.id === job.value.experience)?.name || '')
  // 福利待遇：tag 为分类编码数组，映射为中文名（映射不到时兜底展示原值）
  const tagNames = computed(() =>
    (job.value.tags || []).map(
      (code) => categories.value.jobtag?.find((t) => t.id === code)?.name || String(code)
    )
  )

  const applyVisible = ref(false)
  const noResumeVisible = ref(false)
  const incompleteResumeVisible = ref(false)
  const resumeToComplete = ref<ResumeLite | null>(null)
  const checkingResume = ref(false)
  const applying = ref(false)
  const applyForm = reactive({ notes: '' })
  const resumeList = ref<ResumeLite[]>([])
  const selectedResumeId = ref<number | null>(null)

  const handleApply = async () => {
    if (!userStore.token || userStore.utype !== 1) {
      ElMessage.warning('请先登录个人账号')
      router.push({ name: 'Login', query: { redirect: route.fullPath } })
      return
    }

    checkingResume.value = true
    try {
      const { data } = await listResumes({ page: 1, pageSize: 50 })
      resumeList.value = data.list || []
      if (resumeList.value.length === 0) {
        noResumeVisible.value = true
        return
      }
      const defaultResume = resumeList.value.find((resume) => resume.def === 1) || resumeList.value[0]
      selectedResumeId.value = defaultResume.id
      if (defaultResume.completePercent < MIN_APPLY_RESUME_PERCENT) {
        resumeToComplete.value = defaultResume
        incompleteResumeVisible.value = true
        return
      }
      applyForm.notes = ''
      applyVisible.value = true
    } finally {
      checkingResume.value = false
    }
  }

  const handleChat = () => {
    if (!userStore.token || userStore.utype !== 1) {
      ElMessage.warning('请先登录个人账号')
      router.push({ name: 'Login', query: { redirect: route.fullPath } })
      return
    }
    const peerUid = Number(job.value.hrUid || job.value.uid)
    if (!peerUid) {
      ElMessage.warning('企业信息不完整，暂不能发起沟通')
      return
    }
    router.push({
      name: 'PersonalChat',
      query: { peerUid: String(peerUid), jobsId: String(jobId), jobsName: job.value.jobsName || '' }
    })
  }

  const goCreateResume = () => {
    noResumeVisible.value = false
    router.push({ name: 'ResumeNew' })
  }

  const goCompleteResume = () => {
    if (!resumeToComplete.value) return
    incompleteResumeVisible.value = false
    router.push({ name: 'ResumeEdit', params: { id: resumeToComplete.value.id } })
  }

  const submitApply = async () => {
    applying.value = true
    try {
      await apply({
        jobsIds: [jobId],
        resumeId: selectedResumeId.value ?? 0,
        notes: applyForm.notes.trim()
      })
      ElMessage.success('投递成功')
      applyVisible.value = false
    } finally {
      applying.value = false
    }
  }
</script>
