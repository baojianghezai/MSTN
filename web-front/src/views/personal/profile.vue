<template>
  <div class="mx-auto max-w-2xl px-4 py-16">
    <h1 class="text-2xl font-bold text-slate-800">个人资料</h1>
    <p class="mt-2 text-sm text-slate-500">完善资料有助于求职</p>

    <el-form :model="form" label-width="80px" class="mt-8" @submit.prevent>
      <el-form-item label="姓名">
        <el-input v-model="form.realname" placeholder="请输入真实姓名" clearable />
      </el-form-item>

      <el-form-item label="性别">
        <el-select v-model="form.sex" placeholder="请选择性别" class="w-full" clearable :empty-values="[null, undefined, 0]">
          <el-option v-for="c in categories.sex" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>
      </el-form-item>

      <el-form-item label="生日">
        <el-date-picker
          v-model="birthdayDate"
          type="date"
          placeholder="选择生日"
          value-format="YYYY-MM-DD"
          class="w-full"
        />
      </el-form-item>

      <el-form-item label="学历">
        <el-select v-model="form.education" placeholder="请选择学历" class="w-full" clearable :empty-values="[null, undefined, 0]">
          <el-option v-for="c in categories.education" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>
      </el-form-item>

      <el-form-item label="工作经验">
        <el-select v-model="form.experience" placeholder="请选择工作年限" class="w-full" clearable :empty-values="[null, undefined, 0]">
          <el-option v-for="c in categories.experience" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>
      </el-form-item>

      <el-form-item label="专业">
        <el-select v-model="form.major" placeholder="请选择专业" class="w-full" clearable :empty-values="[null, undefined, 0]">
          <el-option v-for="c in categories.major" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>
      </el-form-item>

      <el-form-item label="婚姻">
        <el-select v-model="form.marriage" placeholder="请选择婚姻状态" class="w-full" clearable :empty-values="[null, undefined, 0]">
          <el-option v-for="c in categories.marriage" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>
      </el-form-item>

      <el-form-item label="现居地">
        <el-input v-model="form.residence" placeholder="请输入现居城市" clearable />
      </el-form-item>

      <el-form-item label="联系电话">
        <el-input v-model="form.phone" placeholder="请输入联系电话" clearable />
      </el-form-item>

      <el-form-item label="身高">
        <div class="flex w-full items-center gap-3">
          <el-slider v-model="heightNum" :min="140" :max="220" :step="1" class="flex-1" />
          <span class="w-12 flex-none text-slate-500">{{ heightNum }}cm</span>
        </div>
      </el-form-item>

      <el-form-item label="QQ">
        <el-input v-model="form.qq" placeholder="选填" clearable />
      </el-form-item>

      <el-form-item label="微信">
        <el-input v-model="form.weixin" placeholder="选填" clearable />
      </el-form-item>

      <el-button type="primary" :loading="saving" @click="handleSave">保存资料</el-button>
    </el-form>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, reactive, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { getPersonalProfile, updatePersonalProfile } from '@/api/profile'
  import { getCategories } from '@/api/content'
  import { dateInputToUnix, unixToDateInput } from '@/utils/format'
  import type { Categories, PersonalProfile } from '@/types/api'

  const emptyProfile: PersonalProfile = {
    realname: '',
    sex: 0,
    sexCn: '',
    birthday: 0,
    residence: '',
    education: 0,
    educationCn: '',
    major: 0,
    majorCn: '',
    experience: 0,
    experienceCn: '',
    phone: '',
    height: '',
    marriage: 0,
    marriageCn: '',
    displayName: 1,
    qq: '',
    weixin: ''
  }

  const form = reactive<PersonalProfile>({ ...emptyProfile })
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
  const birthdayDate = ref('')
  const heightNum = ref(170)
  const saving = ref(false)

  onMounted(async () => {
    const [profileRes, categoriesRes] = await Promise.all([
      getPersonalProfile(),
      getCategories()
    ])
    Object.assign(form, profileRes.data)
    birthdayDate.value = unixToDateInput(profileRes.data.birthday)
    heightNum.value = parseInt(profileRes.data.height) || 170
    categories.value = categoriesRes.data
  })

  const handleSave = async () => {
    // X1：清空后的下拉为 undefined，提交前还原为 0（后端契约：数值字段 0=空）
    const toNum = (v: unknown, d = 0): number => (typeof v === 'number' && !Number.isNaN(v) ? v : d)
    form.sex = toNum(form.sex)
    form.education = toNum(form.education)
    form.experience = toNum(form.experience)
    form.major = toNum(form.major)
    form.marriage = toNum(form.marriage)
    saving.value = true
    try {
      await updatePersonalProfile({
        ...form,
        birthday: dateInputToUnix(birthdayDate.value),
        height: String(heightNum.value),
        // 选中下拉后，把中文 label 一起写入 xxxCn 冗余字段（后端不查表生成）
        educationCn: categories.value.education.find((c) => c.id === form.education)?.name || '',
        experienceCn: categories.value.experience.find((c) => c.id === form.experience)?.name || '',
        sexCn: categories.value.sex.find((c) => c.id === form.sex)?.name || '',
        marriageCn: categories.value.marriage.find((c) => c.id === form.marriage)?.name || '',
        majorCn: categories.value.major.find((c) => c.id === form.major)?.name || ''
      })
      ElMessage.success('资料已保存')
    } finally {
      saving.value = false
    }
  }
</script>
