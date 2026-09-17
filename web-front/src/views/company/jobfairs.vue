<template>
  <div class="mx-auto max-w-6xl px-4 py-12">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-slate-800">举办招聘会</h1>
        <p class="mt-2 text-sm text-slate-500">发布本企业的线上线下招聘会，个人求职者可直接报名参加</p>
      </div>
      <el-button type="primary" :disabled="!canHost" @click="openCreate">举办招聘会</el-button>
    </div>

    <el-alert
      v-if="!canHost"
      class="mt-5"
      type="warning"
      :closable="false"
      show-icon
      title="当前套餐不含「举办招聘会」权益"
      description="升级到专业版/旗舰版后即可发布招聘会。"
    >
      <template #default>
        <el-button link type="primary" @click="router.push({ name: 'CompanyPlan' })">前往套餐订购</el-button>
      </template>
    </el-alert>

    <el-card class="mt-6" shadow="never">
      <div v-if="loading && list.length === 0" class="space-y-4 py-2">
        <el-skeleton v-for="i in 3" :key="i" :rows="1" animated />
      </div>
      <el-table v-else v-loading="loading" :data="list">
        <template #empty>
          <el-empty :image-size="80" description="还没有举办招聘会" />
        </template>
        <el-table-column label="招聘会" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">{{ row.title }}</template>
        </el-table-column>
        <el-table-column label="举办时间" width="180" align="center">
          <template #default="{ row }">{{ row.holdTime || '-' }}</template>
        </el-table-column>
        <el-table-column label="地点" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">{{ row.address || '-' }}</template>
        </el-table-column>
        <el-table-column label="报名人数" width="100" align="center">
          <template #default="{ row }">{{ row.signupCount || 0 }}</template>
        </el-table-column>
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.display === 1 ? 'success' : 'info'" size="small">
              {{ row.display === 1 ? '展示中' : '已下架' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="170" align="center">
          <template #default="{ row }">
            <el-button type="primary" link @click="openEdit(row)">编辑</el-button>
            <el-button v-if="row.display === 1" type="danger" link @click="handleDelete(row)">下架</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑招聘会' : '举办招聘会'" width="600px" :close-on-click-modal="false">
      <el-form :model="form" label-width="90px" @submit.prevent>
        <el-form-item label="招聘会名称" required>
          <el-input v-model="form.title" maxlength="150" placeholder="如：春季大型综合招聘会" clearable />
        </el-form-item>
        <el-form-item label="举办时间" required>
          <el-input v-model="form.holdTime" maxlength="60" placeholder="如：2026-03-15 09:00-16:00" clearable />
        </el-form-item>
        <el-form-item label="举办地点" required>
          <el-input v-model="form.address" maxlength="150" placeholder="如：市人才市场一楼大厅" clearable />
        </el-form-item>
        <el-form-item label="主办方">
          <el-input v-model="form.organizer" maxlength="100" placeholder="默认使用企业名称" clearable />
        </el-form-item>
        <el-form-item label="摘要">
          <el-input v-model="form.summary" maxlength="255" placeholder="一句话简介" clearable />
        </el-form-item>
        <el-form-item label="详情">
          <el-input v-model="form.content" type="textarea" :rows="4" maxlength="2000" show-word-limit placeholder="招聘会详情、参会企业、岗位方向等" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, reactive, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    createCompanyJobfair,
    deleteCompanyJobfair,
    listCompanyJobfairs,
    updateCompanyJobfair
  } from '@/api/company'
  import type { CompanyJobfairPayload } from '@/api/company'
  import { getCurrentSetmeal } from '@/api/billing'
  import type { ArticleItem } from '@/types/api'

  const router = useRouter()

  const list = ref<ArticleItem[]>([])
  const loading = ref(false)
  const saving = ref(false)
  const canHost = ref(true)
  const dialogVisible = ref(false)
  const isEdit = ref(false)
  const editId = ref(0)

  const emptyForm = (): CompanyJobfairPayload => ({
    title: '',
    summary: '',
    cover: '',
    content: '',
    holdTime: '',
    address: '',
    organizer: ''
  })
  const form = reactive<CompanyJobfairPayload>(emptyForm())

  const load = async () => {
    loading.value = true
    try {
      const { data } = await listCompanyJobfairs()
      list.value = data
    } finally {
      loading.value = false
    }
  }

  onMounted(async () => {
    try {
      const { data } = await getCurrentSetmeal()
      canHost.value = Boolean(data && data.enableJobfair)
    } catch {
      canHost.value = false
    }
    await load()
  })

  const openCreate = () => {
    if (!canHost.value) return
    isEdit.value = false
    editId.value = 0
    Object.assign(form, emptyForm())
    dialogVisible.value = true
  }

  const openEdit = (row: ArticleItem) => {
    isEdit.value = true
    editId.value = row.id
    Object.assign(form, {
      title: row.title,
      summary: row.summary,
      cover: row.cover,
      content: row.content,
      holdTime: row.holdTime,
      address: row.address,
      organizer: row.organizer
    })
    dialogVisible.value = true
  }

  const handleSave = async () => {
    if (!form.title.trim() || !form.holdTime.trim() || !form.address.trim()) {
      ElMessage.warning('请填写招聘会名称、举办时间和地点')
      return
    }
    saving.value = true
    try {
      if (isEdit.value) {
        await updateCompanyJobfair(editId.value, form)
        ElMessage.success('招聘会已更新')
      } else {
        await createCompanyJobfair(form)
        ElMessage.success('招聘会已发布')
      }
      dialogVisible.value = false
      await load()
    } catch {
      // 错误已由拦截器统一弹出
    } finally {
      saving.value = false
    }
  }

  const handleDelete = async (row: ArticleItem) => {
    try {
      await ElMessageBox.confirm(`确认下架招聘会「${row.title}」？`, '下架确认', {
        confirmButtonText: '下架',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }
    await deleteCompanyJobfair(row.id)
    ElMessage.success('已下架')
    await load()
  }
</script>