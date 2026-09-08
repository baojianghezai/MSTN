<template>
  <div class="mx-auto max-w-6xl px-4 py-10">
    <div class="flex items-start justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-slate-800">找人才</h1>
        <p class="mt-1 text-sm text-slate-500">搜索公开候选人，解锁后可查看完整联系方式</p>
      </div>
      <el-button @click="router.push({ name: 'CompanyTalentLibrary' })">人才库</el-button>
    </div>

    <el-card class="mt-6" shadow="never">
      <el-form inline @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="职位、技能或简历标题" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="地区">
          <el-input v-model="filters.district" placeholder="地区编码或名称" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="premiumOnly">仅高级人才</el-checkbox>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="mt-4" shadow="never">
      <div v-if="loading && tableData.length === 0" class="space-y-4 py-2">
        <el-skeleton v-for="i in 4" :key="i" :rows="2" animated />
      </div>
      <el-table v-else v-loading="loading" :data="tableData">
        <template #empty>
          <el-empty :image-size="80" description="没有符合条件的公开简历" />
        </template>
        <el-table-column label="候选人" min-width="150">
          <template #default="{ row }">
            <div class="flex items-center gap-2">
              <el-avatar v-if="row.photoImg" :src="row.photoImg" :size="30" />
              <span>{{ row.fullname }}</span>
              <el-tag v-if="row.talent === 1" type="warning" size="small">高级</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="求职意向" min-width="170" show-overflow-tooltip>
          <template #default="{ row }">{{ row.intentionJobs || row.title || '-' }}</template>
        </el-table-column>
        <el-table-column label="学历 / 经验" min-width="160">
          <template #default="{ row }">{{ row.educationCn || '-' }} / {{ row.experienceCn || '-' }}</template>
        </el-table-column>
        <el-table-column label="期望地区" min-width="130" show-overflow-tooltip>
          <template #default="{ row }">{{ row.districtCn || '-' }}</template>
        </el-table-column>
        <el-table-column label="期望薪资" width="130">
          <template #default="{ row }">{{ row.wageCn || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="170" align="center">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleUnlock(row)">解锁联系</el-button>
            <el-button link @click="handleFavorite(row)">收藏</el-button>
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
          @size-change="loadList"
          @current-change="loadList"
        />
      </div>
    </el-card>

    <el-dialog v-model="contactVisible" title="候选人联系方式" width="480px">
      <template v-if="unlockedDetail">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="候选人">{{ unlockedDetail.resume.fullname }}</el-descriptions-item>
          <el-descriptions-item label="求职意向">{{ unlockedDetail.resume.intentionJobs || unlockedDetail.resume.title }}</el-descriptions-item>
          <el-descriptions-item label="联系电话">{{ unlockedDetail.resume.telephone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="邮箱">{{ unlockedDetail.resume.email || '-' }}</el-descriptions-item>
        </el-descriptions>
      </template>
      <template #footer>
        <el-button @click="contactVisible = false">关闭</el-button>
        <el-button type="primary" @click="router.push({ name: 'CompanyTalentLibrary' })">前往人才库</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, reactive, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { ElMessage } from 'element-plus'
  import { favoriteTalent, searchPremiumTalents, searchTalents, unlockTalent } from '@/api/talent'
  import type { PublicResume, TalentUnlockedDetail } from '@/api/talent'

  const router = useRouter()
  const loading = ref(false)
  const tableData = ref<PublicResume[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)
  const premiumOnly = ref(false)
  const contactVisible = ref(false)
  const unlockedDetail = ref<TalentUnlockedDetail | null>(null)
  const filters = reactive({ keyword: '', district: '' })

  const loadList = async () => {
    loading.value = true
    try {
      const params = {
        page: page.value,
        pageSize: pageSize.value,
        keyword: filters.keyword || undefined,
        district: filters.district || undefined
      }
      const { data } = premiumOnly.value ? await searchPremiumTalents(params) : await searchTalents(params)
      tableData.value = data.list
      total.value = data.total
    } finally {
      loading.value = false
    }
  }

  onMounted(loadList)

  const handleSearch = () => {
    page.value = 1
    loadList()
  }

  const handleReset = () => {
    filters.keyword = ''
    filters.district = ''
    premiumOnly.value = false
    handleSearch()
  }

  const handleUnlock = async (row: PublicResume) => {
    const { data } = await unlockTalent(row.id)
    unlockedDetail.value = data.detail
    contactVisible.value = true
    ElMessage.success(data.newlyUnlocked ? '简历已解锁，已扣除一次下载权益' : '该候选人已在人才库中')
  }

  const handleFavorite = async (row: PublicResume) => {
    await favoriteTalent(row.id)
    ElMessage.success('已收藏')
  }
</script>
