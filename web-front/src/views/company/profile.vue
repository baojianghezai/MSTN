<template>
  <div class="mx-auto max-w-3xl px-4 py-16">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-800">企业资料</h1>
      <el-tag :type="auditTag(audit.audit)">{{ audit.auditCn || '未提交' }}</el-tag>
    </div>
    <p class="mt-2 text-sm text-slate-500">提交后进入审核中（audit=2），通过后前台展示企业信息</p>

    <el-form :model="form" label-width="100px" class="mt-8" @submit.prevent>
      <el-form-item label="企业名称" required>
        <el-input v-model="form.companyname" placeholder="请输入企业全称（唯一）" clearable />
      </el-form-item>

      <el-form-item label="企业性质">
        <el-select v-model="form.nature" placeholder="请选择企业性质" class="w-full" clearable :empty-values="[null, undefined, 0]">
          <el-option v-for="c in categories.nature" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>
      </el-form-item>

      <el-form-item label="所属行业">
        <el-select v-model="form.trade" placeholder="请选择行业" class="w-full" clearable :empty-values="[null, undefined, 0]">
          <el-option v-for="c in categories.trade" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>
      </el-form-item>

      <el-form-item label="企业规模">
        <el-select v-model="form.scale" placeholder="请选择规模" class="w-full" clearable :empty-values="[null, undefined, 0]">
          <el-option v-for="c in categories.scale" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>
      </el-form-item>

      <el-form-item label="注册资金">
        <el-input v-model="form.registered" placeholder="如 1000 万元" clearable />
      </el-form-item>

      <el-form-item label="所在地区">
        <el-input v-model="form.district" placeholder="地区数据暂未导入，请手填，如 北京市" clearable />
      </el-form-item>

      <el-form-item label="详细地址">
        <el-input v-model="form.address" placeholder="请输入办公地址" clearable />
      </el-form-item>

      <el-form-item label="联系人">
        <el-input v-model="form.contact" placeholder="请输入联系人姓名" clearable />
      </el-form-item>

      <el-form-item label="联系电话">
        <el-input v-model="form.telephone" placeholder="请输入联系电话" clearable />
      </el-form-item>

      <el-form-item label="座机">
        <el-input v-model="form.landlineTel" placeholder="选填" clearable />
      </el-form-item>

      <el-form-item label="邮箱">
        <el-input v-model="form.email" placeholder="选填" clearable />
      </el-form-item>

      <el-form-item label="官网">
        <el-input v-model="form.website" placeholder="选填" clearable />
      </el-form-item>

      <el-form-item label="企业 Logo">
        <div>
          <el-upload
            action="/api/v1/company/profile/logo"
            :headers="uploadHeaders"
            :show-file-list="false"
            accept="image/*"
            :on-success="onLogoUploaded"
            :before-upload="beforeUpload"
          >
            <el-button>上传 Logo</el-button>
          </el-upload>
          <img v-if="form.logo" :src="imgUrl(form.logo)" class="mt-2 h-16 w-16 rounded object-cover" />
        </div>
      </el-form-item>

      <el-form-item label="营业执照">
        <div>
          <el-upload
            action="/api/v1/company/profile/logo"
            :headers="uploadHeaders"
            :show-file-list="false"
            accept="image/*"
            :on-success="onCertUploaded"
            :before-upload="beforeUpload"
          >
            <el-button>上传证照</el-button>
          </el-upload>
          <img v-if="form.certificateImg" :src="imgUrl(form.certificateImg)" class="mt-2 h-16 w-16 rounded object-cover" />
        </div>
      </el-form-item>

      <el-form-item label="企业简介">
        <el-input v-model="form.contents" type="textarea" :rows="4" placeholder="请输入企业介绍" />
      </el-form-item>

      <el-form-item label="简称">
        <el-input v-model="form.shortName" placeholder="选填" clearable />
      </el-form-item>

      <el-form-item label="一句话简介">
        <el-input v-model="form.shortDesc" placeholder="选填" clearable />
      </el-form-item>

      <el-form-item label="标签">
        <el-input v-model="form.tag" placeholder="选填，如 五险一金,双休" clearable />
      </el-form-item>

      <el-button type="primary" :loading="saving" @click="handleSave">提交资料</el-button>
    </el-form>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { useUserStore } from '@/stores/user'
  import { getCompanyAuditStatus, getCompanyProfile, updateCompanyProfile } from '@/api/company'
  import { getCategories } from '@/api/content'
  import type { Categories, CompanyAudit, CompanyProfile } from '@/types/api'

  const userStore = useUserStore()

  const emptyCompany: CompanyProfile = {
    companyname: '',
    nature: 0,
    natureCn: '',
    trade: 0,
    tradeCn: '',
    district: '',
    districtCn: '',
    scale: 0,
    scaleCn: '',
    registered: '',
    address: '',
    contact: '',
    telephone: '',
    landlineTel: '',
    email: '',
    website: '',
    certificateImg: '',
    logo: '',
    contents: '',
    shortName: '',
    shortDesc: '',
    tag: '',
    audit: 0
  }

  const form = reactive<CompanyProfile>({ ...emptyCompany })
  const categories = ref<Categories>({
    education: [],
    experience: [],
    wage: [],
    trade: [],
    district: [],
    major: [],
    sex: [],
    marriage: [],
    nature: [],
    scale: []
  })
  const audit = ref<CompanyAudit>({ audit: 0, auditCn: '' })
  const saving = ref(false)

  // 上传接口走 el-upload 自己的 XHR，需手动带会员 token
  const uploadHeaders = computed(() => ({ Authorization: `Bearer ${userStore.token}` }))

  onMounted(async () => {
    const [profileRes, categoriesRes, auditRes] = await Promise.all([
      getCompanyProfile(),
      getCategories(),
      getCompanyAuditStatus()
    ])
    // 未提交过资料时后端返回空对象 {}，用默认值兜底
    Object.assign(form, profileRes.data || {})
    form.companyname = form.companyname || ''
    categories.value = categoriesRes.data
    audit.value = auditRes.data
  })

  const handleSave = async () => {
    if (!form.companyname?.trim()) {
      ElMessage.warning('企业名称不能为空')
      return
    }
    // X1：清空后的下拉为 undefined，提交前还原为 0（后端契约：数值字段 0=空）
    const toNum = (v: unknown, d = 0): number => (typeof v === 'number' && !Number.isNaN(v) ? v : d)
    form.nature = toNum(form.nature)
    form.trade = toNum(form.trade)
    form.scale = toNum(form.scale)
    saving.value = true
    try {
      await updateCompanyProfile({
        ...form,
        companyname: form.companyname.trim(),
        natureCn: categories.value.nature.find((c) => c.id === form.nature)?.name || '',
        tradeCn: categories.value.trade.find((c) => c.id === form.trade)?.name || '',
        scaleCn: categories.value.scale.find((c) => c.id === form.scale)?.name || '',
        districtCn: form.district // 地区暂手填，districtCn 同步存
      })
      ElMessage.success('资料已提交，等待审核')
      const { data } = await getCompanyAuditStatus()
      audit.value = data
    } finally {
      saving.value = false
    }
  }

  const onLogoUploaded = (res: { code: number; message: string; data: { url: string } }) => {
    form.logo = res.data.url
    ElMessage.success('Logo 已上传')
  }

  const onCertUploaded = (res: { code: number; message: string; data: { url: string } }) => {
    form.certificateImg = res.data.url
    ElMessage.success('证照已上传')
  }

  const beforeUpload = (file: File) => {
    if (!file.type.startsWith('image/')) {
      ElMessage.warning('只能上传图片')
      return false
    }
    if (file.size / 1024 / 1024 > 5) {
      ElMessage.warning('图片不能超过 5MB')
      return false
    }
    return true
  }

  // 后端返回相对路径 uploads/xxx，图片 src 需补前导斜杠走代理
  const imgUrl = (url: string) => (url ? `/${url}` : '')

  const auditTag = (a: number) => {
    const map: Record<number, string> = { 0: 'info', 1: 'success', 2: 'warning', 3: 'danger' }
    return map[a] || 'info'
  }
</script>
