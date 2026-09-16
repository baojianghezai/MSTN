<template>
  <div class="mx-auto max-w-2xl px-4 py-16">
    <h1 class="text-2xl font-bold text-slate-800">个人资料</h1>
    <p class="mt-2 text-sm text-slate-500">完善资料有助于求职</p>

    <el-form :model="form" label-width="80px" class="mt-8" @submit.prevent>
      <el-form-item label="姓名">
        <el-input v-model="form.realname" placeholder="请输入真实姓名" clearable />
      </el-form-item>

      <el-form-item label="性别">
        <el-select v-model="form.sex" placeholder="请选择性别" style="width:100%" clearable :empty-values="[null, undefined, 0]">
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
        <el-select v-model="form.education" placeholder="请选择学历" style="width:100%" clearable :empty-values="[null, undefined, 0]">
          <el-option v-for="c in categories.education" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>
      </el-form-item>

      <el-form-item label="工作经验">
        <el-select v-model="form.experience" placeholder="请选择工作年限" style="width:100%" clearable :empty-values="[null, undefined, 0]">
          <el-option v-for="c in categories.experience" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>
      </el-form-item>

      <el-form-item label="专业">
        <el-select v-model="form.major" placeholder="请选择专业" style="width:100%" filterable clearable :empty-values="[null, undefined, 0]">
          <el-option v-for="c in categories.major" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>
      </el-form-item>

      <el-form-item label="婚姻">
        <el-select v-model="form.marriage" placeholder="请选择婚姻状态" style="width:100%" clearable :empty-values="[null, undefined, 0]">
          <el-option v-for="c in categories.marriage" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>
      </el-form-item>

      <el-form-item label="现居地">
        <div class="flex gap-2">
          <el-select v-model="provinceId" placeholder="省" style="flex:1;min-width:120px" clearable :empty-values="[null, undefined, 0]">
            <el-option v-for="p in provinces" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
          <el-select v-model="cityId" placeholder="市" style="flex:1;min-width:120px" clearable :empty-values="[null, undefined, 0]" :disabled="!provinceId">
            <el-option v-for="c in cities" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
          <el-select v-model="districtId" placeholder="区" style="flex:1;min-width:120px" clearable :empty-values="[null, undefined, 0]" :disabled="!cityId">
            <el-option v-for="d in districts" :key="d.id" :label="d.name" :value="d.id" />
          </el-select>
        </div>
      </el-form-item>

      <el-form-item label="联系电话">
        <el-input v-model="form.phone" placeholder="请输入联系电话" clearable />
      </el-form-item>

      <el-form-item label="身高">
        <el-input v-model="heightInput" placeholder="请输入身高" class="w-full" clearable>
          <template #append>cm</template>
        </el-input>
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
  import { onMounted, reactive, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import { getPersonalProfile, updatePersonalProfile } from '@/api/profile'
  import { getCategories, getDistricts } from '@/api/content'
  import { dateInputToISO, unixToDateInput } from '@/utils/format'
  import type { Categories, CategoryItem, PersonalProfile } from '@/types/api'

  const emptyProfile: PersonalProfile = {
    realname: '',
    sex: 0,
    sexCn: '',
    birthday: '',
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
  const heightInput = ref('')
  const saving = ref(false)

  // 省/市/区三级联动
  const provinceId = ref<number | null>(null)
  const cityId = ref<number | null>(null)
  const districtId = ref<number | null>(null)
  const provinces = ref<CategoryItem[]>([])
  const cities = ref<CategoryItem[]>([])
  const districts = ref<CategoryItem[]>([])

  const loadProvinces = async () => {
    const res = await getDistricts(0)
    provinces.value = (res.data || []) as CategoryItem[]
  }

  const loadCities = async (pid: number) => {
    const res = await getDistricts(pid)
    cities.value = (res.data || []) as CategoryItem[]
    cityId.value = null
    districtId.value = null
    districts.value = []
  }

  const loadDistricts = async (cid: number) => {
    const res = await getDistricts(cid)
    districts.value = (res.data || []) as CategoryItem[]
    districtId.value = null
  }

  const syncResidence = async () => {
    const parts: string[] = []
    if (provinceId.value) {
      const p = provinces.value.find((x) => x.id === provinceId.value)
      if (p) parts.push(p.name.trim())
    }
    if (cityId.value) {
      const c = cities.value.find((x) => x.id === cityId.value)
      if (c) parts.push(c.name.trim())
    }
    if (districtId.value) {
      const d = districts.value.find((x) => x.id === districtId.value)
      if (d) parts.push(d.name.trim())
    }
    form.residence = parts.join('/')
  }

  const restoreResidence = async (fullName: string) => {
    if (!fullName) return
    const parts = fullName.split('/').map((s) => s.trim())
    await loadProvinces()
    const p = provinces.value.find((x) => x.name.trim() === parts[0])
    if (!p) return
    provinceId.value = p.id
    if (parts.length > 1) {
      await loadCities(p.id)
      const c = cities.value.find((x) => x.name.trim() === parts[1])
      if (!c) return
      cityId.value = c.id
      if (parts.length > 2) {
        await loadDistricts(c.id)
        const d = districts.value.find((x) => x.name.trim() === parts[2])
        if (d) districtId.value = d.id
      }
    }
  }

  watch(provinceId, (v) => {
    if (v) loadCities(v)
    else { cities.value = []; cityId.value = null; districtId.value = null; districts.value = [] }
    syncResidence()
  })

  watch(cityId, (v) => {
    if (v) loadDistricts(v)
    else { districtId.value = null; districts.value = [] }
    syncResidence()
  })

  watch(districtId, syncResidence)

  onMounted(async () => {
    const [profileRes, categoriesRes] = await Promise.all([
      getPersonalProfile(),
      getCategories()
    ])
    Object.assign(form, profileRes.data)
    birthdayDate.value = unixToDateInput(profileRes.data.birthday)
    heightInput.value = profileRes.data.height || ''
    categories.value = categoriesRes.data
    await loadProvinces()
    if (profileRes.data.residence) {
      await restoreResidence(profileRes.data.residence)
    }
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
        birthday: dateInputToISO(birthdayDate.value),
        height: heightInput.value,
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
