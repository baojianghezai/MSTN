<template>
  <div class="mx-auto max-w-3xl px-4 py-8">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-800">
        {{ isEdit ? '编辑简历' : '新建简历' }}
      </h1>
      <el-button @click="router.back()">返回</el-button>
    </div>

    <!-- 完善度（#57，仅编辑态；与列表 completePercent 同源） -->
    <div v-if="completeness" class="mt-4 max-w-md">
      <div class="flex items-center justify-between text-sm">
        <span class="text-slate-500">简历完善度</span>
        <span class="text-slate-600">
          {{ completeness.percent }}%{{ completeness.missing.length ? ` · 还缺 ${completeness.missing.length} 项` : ' · 已完善' }}
        </span>
      </div>
      <el-progress :percentage="completeness.percent" :stroke-width="10" class="mt-2" />
    </div>

    <!-- P2：长表单锚点导航（sticky 顶栏） -->
    <nav class="sticky top-14 z-10 -mx-4 mt-4 flex gap-2 overflow-x-auto border-b border-slate-200 bg-slate-50/95 px-4 py-2 shadow-sm backdrop-blur">
      <a
        v-for="s in sections"
        :key="s.id"
        :href="`#${s.id}`"
        class="whitespace-nowrap rounded-full border border-slate-200 bg-white px-3 py-1 text-sm text-slate-600 no-underline transition-colors hover:border-primary-300 hover:bg-primary-50 hover:text-primary-600"
      >
        {{ s.label }}
      </a>
    </nav>

    <el-form :model="form" label-width="90px" class="mt-4" @submit.prevent>
      <!-- 基本信息 -->
      <el-card id="sec-basic" shadow="never" class="scroll-mt-32">
        <template #header>
          <span class="font-semibold">基本信息</span>
        </template>

        <el-form-item label="简历模板">
          <el-radio-group v-model="form.template">
            <el-radio-button :value="1">经典</el-radio-button>
            <el-radio-button :value="2">简约</el-radio-button>
            <el-radio-button :value="3">紧凑</el-radio-button>
          </el-radio-group>
          <span class="ml-2 text-xs text-slate-400">企业下载本简历时按此模板生成 PDF</span>
        </el-form-item>

        <div class="grid grid-cols-1 gap-x-6 sm:grid-cols-2">
          <el-form-item label="姓名">
            <el-input v-model="form.fullname" maxlength="15" placeholder="请输入姓名" clearable />
          </el-form-item>
          <el-form-item label="性别">
            <el-select v-model="form.sex" placeholder="请选择" class="w-full" clearable :empty-values="[null, undefined, 0]">
              <el-option v-for="c in categories.sex" :key="c.id" :label="c.name" :value="c.id" />
            </el-select>
          </el-form-item>
        </div>

        <div class="grid grid-cols-1 gap-x-6 sm:grid-cols-2">
          <el-form-item label="出生年">
            <el-select v-model="form.birthdate" placeholder="请选择" class="w-full" clearable filterable :empty-values="[null, undefined, 0]">
              <el-option v-for="y in birthYears" :key="y" :label="y" :value="y" />
            </el-select>
          </el-form-item>
          <el-form-item label="籍贯">
            <div class="flex w-full gap-2">
              <el-select v-model="resProvinceId" placeholder="省" class="flex-1" clearable :empty-values="[null, undefined, 0]" @change="onResProvinceChange">
                <el-option v-for="p in resProvinces" :key="p.id" :label="p.name" :value="p.id" />
              </el-select>
              <el-select v-model="resCityId" placeholder="市" class="flex-1" clearable :disabled="!resProvinceId" :empty-values="[null, undefined, 0]" @change="onResCityChange">
                <el-option v-for="c in resCities" :key="c.id" :label="c.name" :value="c.id" />
              </el-select>
              <el-select v-model="resDistrictId" placeholder="区" class="flex-1" clearable :disabled="!resCityId" :empty-values="[null, undefined, 0]" @change="syncResidence">
                <el-option v-for="d in resDistricts" :key="d.id" :label="d.name" :value="d.id" />
              </el-select>
            </div>
          </el-form-item>
        </div>

        <div class="grid grid-cols-1 gap-x-6 sm:grid-cols-2">
          <el-form-item label="学历">
            <el-select v-model="form.education" placeholder="请选择" class="w-full" clearable :empty-values="[null, undefined, 0]">
              <el-option v-for="c in categories.education" :key="c.id" :label="c.name" :value="c.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="专业">
            <el-select v-model="form.major" placeholder="请选择" class="w-full" filterable clearable :empty-values="[null, undefined, 0]">
              <el-option v-for="c in categories.major" :key="c.id" :label="c.name" :value="c.id" />
            </el-select>
          </el-form-item>
        </div>

        <div class="grid grid-cols-1 gap-x-6 sm:grid-cols-2">
          <el-form-item label="工作年限">
            <el-select v-model="form.experience" placeholder="请选择" class="w-full" clearable :empty-values="[null, undefined, 0]">
              <el-option v-for="c in categories.experience" :key="c.id" :label="c.name" :value="c.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="期望薪资">
            <div class="flex w-full items-center gap-2">
              <el-input-number v-model="form.wageMin" :min="0" :max="999999" :controls="false" placeholder="最低" class="flex-1" />
              <span class="text-slate-400">-</span>
              <el-input-number v-model="form.wageMax" :min="0" :max="999999" :controls="false" placeholder="最高" class="flex-1" />
              <span class="whitespace-nowrap text-xs text-slate-400">元/月</span>
            </div>
          </el-form-item>
        </div>

        <el-form-item label="期望职位">
          <el-input v-model="form.intentionJobs" maxlength="255" placeholder="如：Go 后端开发" clearable />
        </el-form-item>

        <el-form-item label="期望地区">
          <div class="flex w-full gap-2">
            <el-select v-model="distProvinceId" placeholder="省" class="flex-1" clearable :empty-values="[null, undefined, 0]" @change="onDistProvinceChange">
              <el-option label="不限" :value="-1" />
              <el-option v-for="p in distProvinces" :key="p.id" :label="p.name" :value="p.id" />
            </el-select>
            <el-select v-model="distCityId" placeholder="市" class="flex-1" clearable :disabled="!distProvinceId || distProvinceId === -1" :empty-values="[null, undefined, 0]" @change="onDistCityChange">
              <el-option label="全部" :value="0" />
              <el-option v-for="c in distCities" :key="c.id" :label="c.name" :value="c.id" />
            </el-select>
            <el-select v-model="distDistrictId" placeholder="区" class="flex-1" clearable :disabled="!distCityId || distProvinceId === -1" :empty-values="[null, undefined, 0]" @change="syncDistrict">
              <el-option label="全部" :value="0" />
              <el-option v-for="d in distDistricts" :key="d.id" :label="d.name" :value="d.id" />
            </el-select>
          </div>
        </el-form-item>

        <el-form-item label="联系电话">
          <el-input v-model="form.telephone" maxlength="50" placeholder="请输入联系电话" clearable />
        </el-form-item>

        <el-form-item label="邮箱">
          <el-input v-model="form.email" maxlength="60" placeholder="请输入邮箱" clearable />
        </el-form-item>

        <el-form-item label="自我评价">
          <el-input
            v-model="form.specialty"
            type="textarea"
            :rows="3"
            maxlength="1000"
            show-word-limit
            placeholder="一句话介绍自己的优势"
          />
        </el-form-item>
      </el-card>

      <!-- 教育经历（子表 educations，不限条数） -->
      <el-card id="sec-education" shadow="never" class="mt-6 scroll-mt-32">
        <template #header>
          <div class="flex items-center justify-between">
            <span class="font-semibold">教育经历</span>
            <span class="text-xs text-slate-400">{{ form.educations.length }} 条</span>
          </div>
        </template>

        <div
          v-for="(e, i) in form.educations"
          :key="i"
          class="mb-4 rounded border border-slate-200 p-4"
        >
          <div class="flex items-center justify-between">
            <span class="text-sm font-medium text-slate-700">教育 {{ i + 1 }}</span>
            <el-button type="danger" text size="small" @click="removeRow(form.educations, i)">
              删除
            </el-button>
          </div>

          <div class="mt-3 grid grid-cols-1 gap-x-6 sm:grid-cols-2">
            <el-form-item label="学校" label-width="80px">
              <el-input v-model="e.school" maxlength="50" placeholder="学校名称" clearable />
            </el-form-item>
            <el-form-item label="专业" label-width="80px">
              <el-input v-model="e.speciality" maxlength="50" placeholder="专业名称" clearable />
            </el-form-item>
          </div>

          <div class="grid grid-cols-1 gap-x-6 sm:grid-cols-2">
            <el-form-item label="学历" label-width="80px">
              <el-select v-model="e.education" placeholder="请选择" class="w-full" clearable :empty-values="[null, undefined, 0]">
                <el-option v-for="c in categories.education" :key="c.id" :label="c.name" :value="c.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="起止时间" label-width="80px">
              <div class="flex w-full items-center gap-2">
                <el-select v-model="e.startyear" placeholder="开始年" class="flex-1" clearable filterable :empty-values="[null, undefined, 0]">
                  <el-option v-for="y in projectYears" :key="y" :label="y" :value="y" />
                </el-select>
                <el-select v-model="e.startmonth" placeholder="月" class="flex-1" clearable :empty-values="[null, undefined, 0]">
                  <el-option v-for="m in 12" :key="m" :label="`${m}月`" :value="m" />
                </el-select>
              </div>
            </el-form-item>
          </div>

          <div class="grid grid-cols-1 gap-x-6 sm:grid-cols-2">
            <el-form-item label="结束时间" label-width="80px">
              <div class="flex w-full items-center gap-2">
                <el-select
                  v-model="e.endyear"
                  placeholder="年"
                  class="flex-1"
                  clearable
                  filterable
                  :disabled="e.todate === 1"
                  :empty-values="[null, undefined, 0]"
                >
                  <el-option v-for="y in projectYears" :key="y" :label="y" :value="y" />
                </el-select>
                <el-select
                  v-model="e.endmonth"
                  placeholder="月"
                  class="flex-1"
                  clearable
                  :disabled="e.todate === 1"
                  :empty-values="[null, undefined, 0]"
                >
                  <el-option v-for="m in 12" :key="m" :label="`${m}月`" :value="m" />
                </el-select>
                <el-checkbox v-model="e.todate" :true-value="1" :false-value="0">至今</el-checkbox>
              </div>
            </el-form-item>
          </div>
        </div>

        <el-button class="w-full" @click="form.educations.push(emptyEducation())">
          添加教育经历
        </el-button>
      </el-card>

      <!-- 工作/实习经历（子表 work，不限条数） -->
      <el-card id="sec-work" shadow="never" class="mt-6 scroll-mt-32">
        <template #header>
          <div class="flex items-center justify-between">
            <span class="font-semibold">工作/实习经历</span>
            <span class="text-xs text-slate-400">{{ form.work.length }} 条</span>
          </div>
        </template>

        <div v-for="(w, i) in form.work" :key="i" class="mb-4 rounded border border-slate-200 p-4">
          <div class="flex items-center justify-between">
            <span class="text-sm font-medium text-slate-700">
              经历 {{ i + 1 }}
              <el-tag class="ml-1" size="small" :type="w.workType === 2 ? 'warning' : 'primary'">
                {{ w.workType === 2 ? '实习' : '工作' }}
              </el-tag>
            </span>
            <el-button type="danger" text size="small" @click="removeRow(form.work, i)">
              删除
            </el-button>
          </div>

          <div class="mt-3 grid grid-cols-1 gap-x-6 sm:grid-cols-2">
            <el-form-item label="经历类型" label-width="80px">
              <el-radio-group v-model="w.workType">
                <el-radio-button :value="1">工作</el-radio-button>
                <el-radio-button :value="2">实习</el-radio-button>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="公司" label-width="80px">
              <el-input v-model="w.companyname" maxlength="50" placeholder="公司名称" clearable />
            </el-form-item>
          </div>

          <div class="grid grid-cols-1 gap-x-6 sm:grid-cols-2">
            <el-form-item label="职位" label-width="80px">
              <el-input v-model="w.jobs" maxlength="30" placeholder="担任职位" clearable />
            </el-form-item>
          </div>

          <div class="grid grid-cols-1 gap-x-6 sm:grid-cols-2">
            <el-form-item label="起止时间" label-width="80px">
              <div class="flex w-full items-center gap-2">
                <el-select v-model="w.startyear" placeholder="开始年" class="flex-1" clearable filterable :empty-values="[null, undefined, 0]">
                  <el-option v-for="y in projectYears" :key="y" :label="y" :value="y" />
                </el-select>
                <el-select v-model="w.startmonth" placeholder="月" class="flex-1" clearable :empty-values="[null, undefined, 0]">
                  <el-option v-for="m in 12" :key="m" :label="`${m}月`" :value="m" />
                </el-select>
              </div>
            </el-form-item>
            <el-form-item label="结束时间" label-width="80px">
              <div class="flex w-full items-center gap-2">
                <el-select
                  v-model="w.endyear"
                  placeholder="年"
                  class="flex-1"
                  clearable
                  filterable
                  :disabled="w.todate === 1"
                  :empty-values="[null, undefined, 0]"
                >
                  <el-option v-for="y in projectYears" :key="y" :label="y" :value="y" />
                </el-select>
                <el-select
                  v-model="w.endmonth"
                  placeholder="月"
                  class="flex-1"
                  clearable
                  :disabled="w.todate === 1"
                  :empty-values="[null, undefined, 0]"
                >
                  <el-option v-for="m in 12" :key="m" :label="`${m}月`" :value="m" />
                </el-select>
                <el-checkbox v-model="w.todate" :true-value="1" :false-value="0">至今</el-checkbox>
              </div>
            </el-form-item>
          </div>

          <el-form-item label="工作业绩" label-width="80px">
            <el-input
              v-model="w.achievements"
              type="textarea"
              :rows="2"
              maxlength="1000"
              show-word-limit
              placeholder="职责、成果、数据"
            />
          </el-form-item>
        </div>

        <el-button class="w-full" @click="form.work.push(emptyWork())">
          添加工作/实习经历
        </el-button>
      </el-card>

      <!-- 项目经历（子表 projects，限 6 条） -->
      <el-card id="sec-project" shadow="never" class="mt-6 scroll-mt-32">
        <template #header>
          <div class="flex items-center justify-between">
            <span class="font-semibold">项目经历</span>
            <span class="text-xs text-slate-400">{{ form.projects.length }}/6 条</span>
          </div>
        </template>

        <div v-for="(p, i) in form.projects" :key="i" class="mb-4 rounded border border-slate-200 p-4">
          <div class="flex items-center justify-between">
            <span class="text-sm font-medium text-slate-700">项目 {{ i + 1 }}</span>
            <el-button type="danger" text size="small" @click="removeRow(form.projects, i)">
              删除
            </el-button>
          </div>

          <div class="mt-3 grid grid-cols-1 gap-x-6 sm:grid-cols-2">
            <el-form-item label="项目名称" label-width="80px">
              <el-input v-model="p.projectname" maxlength="50" placeholder="项目名称" clearable />
            </el-form-item>
            <el-form-item label="担任角色" label-width="80px">
              <el-input v-model="p.role" maxlength="50" placeholder="如：后端开发" clearable />
            </el-form-item>
          </div>

          <div class="grid grid-cols-1 gap-x-6 sm:grid-cols-2">
            <el-form-item label="开始时间" label-width="80px">
              <div class="flex w-full gap-2">
                <el-select v-model="p.startyear" placeholder="年" class="flex-1" clearable filterable :empty-values="[null, undefined, 0]">
                  <el-option v-for="y in projectYears" :key="y" :label="y" :value="y" />
                </el-select>
                <el-select v-model="p.startmonth" placeholder="月" class="flex-1" clearable :empty-values="[null, undefined, 0]">
                  <el-option v-for="m in 12" :key="m" :label="`${m}月`" :value="m" />
                </el-select>
              </div>
            </el-form-item>
            <el-form-item label="结束时间" label-width="80px">
              <div class="flex w-full items-center gap-2">
                <el-select
                  v-model="p.endyear"
                  placeholder="年"
                  class="flex-1"
                  clearable
                  filterable
                  :disabled="p.todate === 1"
                  :empty-values="[null, undefined, 0]"
                >
                  <el-option v-for="y in projectYears" :key="y" :label="y" :value="y" />
                </el-select>
                <el-select
                  v-model="p.endmonth"
                  placeholder="月"
                  class="flex-1"
                  clearable
                  :disabled="p.todate === 1"
                  :empty-values="[null, undefined, 0]"
                >
                  <el-option v-for="m in 12" :key="m" :label="`${m}月`" :value="m" />
                </el-select>
                <el-checkbox v-model="p.todate" :true-value="1" :false-value="0">至今</el-checkbox>
              </div>
            </el-form-item>
          </div>

          <el-form-item label="项目描述" label-width="80px">
            <el-input
              v-model="p.description"
              type="textarea"
              :rows="2"
              maxlength="1000"
              show-word-limit
              placeholder="项目背景、职责、成果"
            />
          </el-form-item>
        </div>

        <el-button class="w-full" :disabled="form.projects.length >= 6" @click="addProject">
          添加项目经历
        </el-button>
        <p v-if="form.projects.length >= 6" class="mt-2 text-center text-xs text-slate-400">
          项目经历最多 6 条
        </p>
      </el-card>

      <!-- 加分项（按需添加：专业技能/语言能力/培训经历/证书/个人作品/学生干部经历） -->
      <el-card id="sec-extra" shadow="never" class="mt-6 scroll-mt-32">
        <template #header>
          <div class="flex items-center justify-between">
            <span class="font-semibold">加分项（可选）</span>
            <span class="text-xs text-slate-400">按需添加，不强制填写</span>
          </div>
        </template>

        <el-collapse v-model="extraPanels">
          <!-- 专业技能 -->
          <el-collapse-item name="skill">
            <template #title>
              <span class="text-sm">专业技能</span>
              <span class="ml-2 text-xs text-slate-400">{{ form.skill.length }} 条</span>
            </template>
            <div v-for="(s, i) in form.skill" :key="i" class="mb-3 rounded border border-slate-200 p-3">
              <div class="flex items-center justify-between">
                <span class="text-sm text-slate-600">技能 {{ i + 1 }}</span>
                <el-button type="danger" text size="small" @click="removeRow(form.skill, i)">删除</el-button>
              </div>
              <div class="mt-2 grid grid-cols-1 gap-x-6 sm:grid-cols-2">
                <el-form-item label="名称" label-width="70px">
                  <el-input v-model="s.name" maxlength="50" placeholder="如：Go / Photoshop" clearable />
                </el-form-item>
                <el-form-item label="熟练度" label-width="70px">
                  <el-select v-model="s.level" placeholder="请选择" class="w-full" clearable :empty-values="[null, undefined, 0]">
                    <el-option v-for="o in skillLevelOptions" :key="o.id" :label="o.name" :value="o.id" />
                  </el-select>
                </el-form-item>
              </div>
            </div>
            <el-button class="w-full" @click="form.skill.push(emptySkill())">添加专业技能</el-button>
          </el-collapse-item>

          <!-- 语言能力 -->
          <el-collapse-item name="language">
            <template #title>
              <span class="text-sm">语言能力</span>
              <span class="ml-2 text-xs text-slate-400">{{ form.language.length }} 条</span>
            </template>
            <div v-for="(l, i) in form.language" :key="i" class="mb-3 rounded border border-slate-200 p-3">
              <div class="flex items-center justify-between">
                <span class="text-sm text-slate-600">语言 {{ i + 1 }}</span>
                <el-button type="danger" text size="small" @click="removeRow(form.language, i)">删除</el-button>
              </div>
              <div class="mt-2 grid grid-cols-1 gap-x-6 sm:grid-cols-2">
                <el-form-item label="语言" label-width="70px">
                  <el-select v-model="l.language" placeholder="请选择" class="w-full" clearable :empty-values="[null, undefined, 0]">
                    <el-option v-for="c in languageOptions" :key="c.id" :label="c.name" :value="c.id" />
                  </el-select>
                </el-form-item>
                <el-form-item label="熟练度" label-width="70px">
                  <el-select v-model="l.level" placeholder="请选择" class="w-full" clearable :empty-values="[null, undefined, 0]">
                    <el-option v-for="c in levelOptions" :key="c.id" :label="c.name" :value="c.id" />
                  </el-select>
                </el-form-item>
              </div>
            </div>
            <el-button class="w-full" @click="form.language.push(emptyLanguage())">添加语言能力</el-button>
          </el-collapse-item>

          <!-- 培训经历 -->
          <el-collapse-item name="training">
            <template #title>
              <span class="text-sm">培训经历</span>
              <span class="ml-2 text-xs text-slate-400">{{ form.training.length }} 条</span>
            </template>
            <div v-for="(t, i) in form.training" :key="i" class="mb-3 rounded border border-slate-200 p-3">
              <div class="flex items-center justify-between">
                <span class="text-sm text-slate-600">培训 {{ i + 1 }}</span>
                <el-button type="danger" text size="small" @click="removeRow(form.training, i)">删除</el-button>
              </div>
              <div class="mt-2 grid grid-cols-1 gap-x-6 sm:grid-cols-2">
                <el-form-item label="机构" label-width="70px">
                  <el-input v-model="t.agency" maxlength="50" placeholder="培训机构" clearable />
                </el-form-item>
                <el-form-item label="课程" label-width="70px">
                  <el-input v-model="t.course" maxlength="50" placeholder="课程名称" clearable />
                </el-form-item>
              </div>
              <div class="grid grid-cols-1 gap-x-6 sm:grid-cols-2">
                <el-form-item label="起止时间" label-width="70px">
                  <div class="flex w-full items-center gap-2">
                    <el-select v-model="t.startyear" placeholder="年" class="flex-1" clearable filterable :empty-values="[null, undefined, 0]">
                      <el-option v-for="y in projectYears" :key="y" :label="y" :value="y" />
                    </el-select>
                    <el-select v-model="t.startmonth" placeholder="月" class="flex-1" clearable :empty-values="[null, undefined, 0]">
                      <el-option v-for="m in 12" :key="m" :label="`${m}月`" :value="m" />
                    </el-select>
                  </div>
                </el-form-item>
                <el-form-item label="结束时间" label-width="70px">
                  <div class="flex w-full items-center gap-2">
                    <el-select v-model="t.endyear" placeholder="年" class="flex-1" clearable filterable :disabled="t.todate === 1" :empty-values="[null, undefined, 0]">
                      <el-option v-for="y in projectYears" :key="y" :label="y" :value="y" />
                    </el-select>
                    <el-select v-model="t.endmonth" placeholder="月" class="flex-1" clearable :disabled="t.todate === 1" :empty-values="[null, undefined, 0]">
                      <el-option v-for="m in 12" :key="m" :label="`${m}月`" :value="m" />
                    </el-select>
                    <el-checkbox v-model="t.todate" :true-value="1" :false-value="0">至今</el-checkbox>
                  </div>
                </el-form-item>
              </div>
              <el-form-item label="培训描述" label-width="70px">
                <el-input v-model="t.description" type="textarea" :rows="2" maxlength="1000" show-word-limit placeholder="培训内容、收获" />
              </el-form-item>
            </div>
            <el-button class="w-full" @click="form.training.push(emptyTraining())">添加培训经历</el-button>
          </el-collapse-item>

          <!-- 证书 -->
          <el-collapse-item name="credent">
            <template #title>
              <span class="text-sm">证书</span>
              <span class="ml-2 text-xs text-slate-400">{{ form.credent.length }} 条</span>
            </template>
            <div v-for="(c, i) in form.credent" :key="i" class="mb-3 rounded border border-slate-200 p-3">
              <div class="flex items-center justify-between">
                <span class="text-sm text-slate-600">证书 {{ i + 1 }}</span>
                <el-button type="danger" text size="small" @click="removeRow(form.credent, i)">删除</el-button>
              </div>
              <div class="mt-2 grid grid-cols-1 gap-x-6 sm:grid-cols-3">
                <el-form-item label="名称" label-width="70px">
                  <el-input v-model="c.name" maxlength="255" placeholder="证书名称" clearable />
                </el-form-item>
                <el-form-item label="年份" label-width="70px">
                  <el-select v-model="c.year" placeholder="年" class="w-full" clearable filterable :empty-values="[null, undefined, 0]">
                    <el-option v-for="y in projectYears" :key="y" :label="y" :value="y" />
                  </el-select>
                </el-form-item>
                <el-form-item label="月份" label-width="70px">
                  <el-select v-model="c.month" placeholder="月" class="w-full" clearable :empty-values="[null, undefined, 0]">
                    <el-option v-for="m in 12" :key="m" :label="`${m}月`" :value="m" />
                  </el-select>
                </el-form-item>
              </div>
            </div>
            <el-button class="w-full" @click="form.credent.push(emptyCredent())">添加证书</el-button>
          </el-collapse-item>

          <!-- 个人作品 -->
          <el-collapse-item name="portfolio">
            <template #title>
              <span class="text-sm">个人作品</span>
              <span class="ml-2 text-xs text-slate-400">{{ form.portfolio.length }} 条</span>
            </template>
            <div v-for="(p, i) in form.portfolio" :key="i" class="mb-3 rounded border border-slate-200 p-3">
              <div class="flex items-center justify-between">
                <span class="text-sm text-slate-600">作品 {{ i + 1 }}</span>
                <el-button type="danger" text size="small" @click="removeRow(form.portfolio, i)">删除</el-button>
              </div>
              <div class="mt-2 grid grid-cols-1 gap-x-6 sm:grid-cols-2">
                <el-form-item label="标题" label-width="70px">
                  <el-input v-model="p.title" maxlength="100" placeholder="作品名称" clearable />
                </el-form-item>
                <el-form-item label="链接" label-width="70px">
                  <el-input v-model="p.url" maxlength="255" placeholder="http(s)://" clearable />
                </el-form-item>
              </div>
              <el-form-item label="描述" label-width="70px">
                <el-input v-model="p.description" type="textarea" :rows="2" maxlength="1000" show-word-limit placeholder="作品简介" />
              </el-form-item>
            </div>
            <el-button class="w-full" @click="form.portfolio.push(emptyPortfolio())">添加个人作品</el-button>
          </el-collapse-item>

          <!-- 学生干部经历 -->
          <el-collapse-item name="studentLeader">
            <template #title>
              <span class="text-sm">学生干部经历</span>
              <span class="ml-2 text-xs text-slate-400">{{ form.studentLeader.length }} 条</span>
            </template>
            <div v-for="(s, i) in form.studentLeader" :key="i" class="mb-3 rounded border border-slate-200 p-3">
              <div class="flex items-center justify-between">
                <span class="text-sm text-slate-600">经历 {{ i + 1 }}</span>
                <el-button type="danger" text size="small" @click="removeRow(form.studentLeader, i)">删除</el-button>
              </div>
              <div class="mt-2 grid grid-cols-1 gap-x-6 sm:grid-cols-2">
                <el-form-item label="组织" label-width="70px">
                  <el-input v-model="s.organization" maxlength="100" placeholder="学生会/社团名称" clearable />
                </el-form-item>
                <el-form-item label="职务" label-width="70px">
                  <el-input v-model="s.role" maxlength="50" placeholder="担任职务" clearable />
                </el-form-item>
              </div>
              <div class="grid grid-cols-1 gap-x-6 sm:grid-cols-2">
                <el-form-item label="起止时间" label-width="70px">
                  <div class="flex w-full items-center gap-2">
                    <el-select v-model="s.startyear" placeholder="年" class="flex-1" clearable filterable :empty-values="[null, undefined, 0]">
                      <el-option v-for="y in projectYears" :key="y" :label="y" :value="y" />
                    </el-select>
                    <el-select v-model="s.startmonth" placeholder="月" class="flex-1" clearable :empty-values="[null, undefined, 0]">
                      <el-option v-for="m in 12" :key="m" :label="`${m}月`" :value="m" />
                    </el-select>
                  </div>
                </el-form-item>
                <el-form-item label="结束时间" label-width="70px">
                  <div class="flex w-full items-center gap-2">
                    <el-select v-model="s.endyear" placeholder="年" class="flex-1" clearable filterable :disabled="s.todate === 1" :empty-values="[null, undefined, 0]">
                      <el-option v-for="y in projectYears" :key="y" :label="y" :value="y" />
                    </el-select>
                    <el-select v-model="s.endmonth" placeholder="月" class="flex-1" clearable :disabled="s.todate === 1" :empty-values="[null, undefined, 0]">
                      <el-option v-for="m in 12" :key="m" :label="`${m}月`" :value="m" />
                    </el-select>
                    <el-checkbox v-model="s.todate" :true-value="1" :false-value="0">至今</el-checkbox>
                  </div>
                </el-form-item>
              </div>
              <el-form-item label="描述" label-width="70px">
                <el-input v-model="s.description" type="textarea" :rows="2" maxlength="1000" show-word-limit placeholder="主要工作与收获" />
              </el-form-item>
            </div>
            <el-button class="w-full" @click="form.studentLeader.push(emptyStudentLeader())">添加学生干部经历</el-button>
          </el-collapse-item>
        </el-collapse>
      </el-card>

      <!-- 附件简历（仅编辑态可上传；PDF） -->
      <el-card v-if="isEdit" id="sec-attachment" shadow="never" class="mt-6 scroll-mt-32">
        <template #header>
          <div class="flex items-center justify-between">
            <span class="font-semibold">附件简历</span>
            <span class="text-xs text-slate-400">仅支持 PDF</span>
          </div>
        </template>

        <div class="flex flex-wrap items-center gap-3">
          <el-upload
            :action="outwardAction"
            :headers="uploadHeaders"
            :show-file-list="false"
            accept="application/pdf,.pdf"
            :before-upload="beforeOutwardUpload"
            :on-success="onOutwardSuccess"
            :on-error="onOutwardError"
          >
            <el-button>{{ form.wordResume ? '重新上传' : '上传 PDF 简历' }}</el-button>
          </el-upload>
          <template v-if="form.wordResume">
            <a
              :href="outwardUrl"
              target="_blank"
              rel="noopener"
              class="text-sm text-primary-600 hover:underline"
            >
              {{ form.wordResumeTitle || '查看附件简历' }}
            </a>
            <el-button type="danger" text size="small" @click="handleRemoveOutward">删除</el-button>
          </template>
          <span v-else class="text-sm text-slate-400">未上传</span>
        </div>
        <p class="mt-2 text-xs text-slate-400">附件简历仅已投递/已下载的企业可见（防爬）。</p>
      </el-card>

      <!-- P2：长表单 sticky 保存栏，滚动任何位置都能直接保存 -->
      <div class="sticky bottom-0 z-10 -mx-4 mt-8 flex justify-end gap-3 border-t border-slate-200 bg-white/95 px-4 py-3 shadow-[0_-4px_12px_rgba(15,23,42,0.08)] backdrop-blur">
        <el-button @click="router.back()">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存简历</el-button>
      </div>
    </el-form>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    createResume,
    getResume,
    getResumeCompleteness,
    removeResumeOutward,
    updateResume
  } from '@/api/resume'
  import { getCategories, getDistricts } from '@/api/content'
  import { useUserStore } from '@/stores/user'
  import type {
    Categories,
    CategoryItem,
    Resume,
    ResumeCompleteness,
    ResumeCredent,
    ResumeEducation,
    ResumeLanguage,
    ResumePortfolio,
    ResumeSkill,
    ResumeStudentLeader,
    ResumeTraining,
    ResumeWork
  } from '@/types/api'

  const route = useRoute()
  const router = useRouter()
  const userStore = useUserStore()

  const resumeId = Number(route.params.id)
  const isEdit = !Number.isNaN(resumeId) && resumeId > 0

  // P2：长表单锚点导航（与模板里各 el-card 的 id 对应）
  const sections = computed(() => {
    const base = [
      { id: 'sec-basic', label: '基本信息' },
      { id: 'sec-education', label: '教育经历' },
      { id: 'sec-work', label: '工作/实习经历' },
      { id: 'sec-project', label: '项目经历' },
      { id: 'sec-extra', label: '加分项' }
    ]
    if (isEdit) base.push({ id: 'sec-attachment', label: '附件简历' })
    return base
  })

  // 语言/等级编码沿用 v6 字典（QS_language 208-213 / QS_language_level 291-293），
  // 后端只存 code+cn 不查表，前端下拉选中后把 label 写入 Cn 冗余字段
  const languageOptions = [
    { id: 208, name: '普通话' },
    { id: 209, name: '粤语' },
    { id: 210, name: '英语' },
    { id: 211, name: '法语' },
    { id: 212, name: '日语' },
    { id: 213, name: '其他' }
  ]
  const levelOptions = [
    { id: 291, name: '入门' },
    { id: 292, name: '熟练' },
    { id: 293, name: '精通' }
  ]
  const skillLevelOptions = [
    { id: 1, name: '入门' },
    { id: 2, name: '熟练' },
    { id: 3, name: '精通' }
  ]

  // 空行工厂（子表提交不带 id/pid/uid，全量替换语义）
  const emptyEducation = (): ResumeEducation => ({
    startyear: 0,
    startmonth: 0,
    endyear: 0,
    endmonth: 0,
    todate: 0,
    school: '',
    speciality: '',
    education: 0,
    educationCn: ''
  })
  const emptyWork = (): ResumeWork => ({
    startyear: 0,
    startmonth: 0,
    endyear: 0,
    endmonth: 0,
    todate: 0,
    workType: 1,
    companyname: '',
    jobs: '',
    achievements: ''
  })
  const emptyLanguage = (): ResumeLanguage => ({
    language: 0,
    languageCn: '',
    level: 0,
    levelCn: ''
  })
  const emptyTraining = (): ResumeTraining => ({
    startyear: 0,
    startmonth: 0,
    endyear: 0,
    endmonth: 0,
    todate: 0,
    agency: '',
    course: '',
    description: ''
  })
  const emptyCredent = (): ResumeCredent => ({
    name: '',
    year: 0,
    month: 0,
    images: ''
  })
  const emptySkill = (): ResumeSkill => ({ name: '', level: 0 })
  const emptyPortfolio = (): ResumePortfolio => ({ title: '', description: '', url: '' })
  const emptyStudentLeader = (): ResumeStudentLeader => ({
    organization: '',
    role: '',
    startyear: 0,
    startmonth: 0,
    endyear: 0,
    endmonth: 0,
    todate: 0,
    description: ''
  })
  const emptyProject = () => ({
    startyear: 0,
    startmonth: 0,
    endyear: 0,
    endmonth: 0,
    todate: 0,
    projectname: '',
    role: '',
    description: ''
  })

  const emptyResume = (): Resume => ({
    title: '',
    template: 1,
    fullname: '',
    sex: 0,
    sexCn: '',
    birthdate: 0,
    residence: '',
    education: 0,
    educationCn: '',
    major: 0,
    majorCn: '',
    experience: 0,
    experienceCn: '',
    district: '',
    districtCn: '',
    wageMin: 0,
    wageMax: 0,
    intentionJobs: '',
    specialty: '',
    telephone: '',
    email: '',
    displayName: 1,
    current: 0,
    currentCn: '',
    mobileAudit: 0,
    talent: 0,
    entrust: 0,
    wordResume: '',
    wordResumeTitle: '',
    educations: [],
    work: [],
    language: [],
    training: [],
    credent: [],
    projects: [],
    skill: [],
    portfolio: [],
    studentLeader: []
  })

  const form = reactive<Resume>(emptyResume())
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
  const saving = ref(false)
  const completeness = ref<ResumeCompleteness | null>(null)
  const extraPanels = ref<string[]>([])

  // 出生年范围：近 60 年倒序
  const currentYear = new Date().getFullYear()
  const birthYears = Array.from({ length: 60 }, (_, i) => currentYear - i)
  // 子表年份范围：近 50 年（可到当前年）
  const projectYears = Array.from({ length: 50 }, (_, i) => currentYear - i)

  // ---- 省/市/区三级联动（籍贯；现居地同款，见个人资料页）----
  const resProvinceId = ref<number | null>(null)
  const resCityId = ref<number | null>(null)
  const resDistrictId = ref<number | null>(null)
  const resProvinces = ref<CategoryItem[]>([])
  const resCities = ref<CategoryItem[]>([])
  const resDistricts = ref<CategoryItem[]>([])

  // ---- 期望地区（省/市/区；市、区级带「全部」，省级带「不限」）----
  const DIST_ANY_PROVINCE = -1 // 省级「不限」
  const DIST_ANY_CHILD = 0 // 市/区级「全部」
  const distProvinceId = ref<number | null>(null)
  const distCityId = ref<number | null>(null)
  const distDistrictId = ref<number | null>(null)
  const distProvinces = ref<CategoryItem[]>([])
  const distCities = ref<CategoryItem[]>([])
  const distDistricts = ref<CategoryItem[]>([])

  const fetchOptions = async (parentId: number): Promise<CategoryItem[]> => {
    const res = await getDistricts(parentId)
    return (res.data || []) as CategoryItem[]
  }

  const onResProvinceChange = async (v: number | null) => {
    resCities.value = []
    resCityId.value = null
    resDistricts.value = []
    resDistrictId.value = null
    if (v) resCities.value = await fetchOptions(v)
    syncResidence()
  }
  const onResCityChange = async (v: number | null) => {
    resDistricts.value = []
    resDistrictId.value = null
    if (v) resDistricts.value = await fetchOptions(v)
    syncResidence()
  }
  const syncResidence = () => {
    const parts: string[] = []
    if (resProvinceId.value) {
      const p = resProvinces.value.find((x) => x.id === resProvinceId.value)
      if (p) parts.push(p.name.trim())
    }
    if (resCityId.value) {
      const c = resCities.value.find((x) => x.id === resCityId.value)
      if (c) parts.push(c.name.trim())
    }
    if (resDistrictId.value) {
      const d = resDistricts.value.find((x) => x.id === resDistrictId.value)
      if (d) parts.push(d.name.trim())
    }
    form.residence = parts.join('/')
  }

  const onDistProvinceChange = async (v: number | null) => {
    distCities.value = []
    distCityId.value = null
    distDistricts.value = []
    distDistrictId.value = null
    if (v && v !== DIST_ANY_PROVINCE) distCities.value = await fetchOptions(v)
    syncDistrict()
  }
  const onDistCityChange = async (v: number | null) => {
    distDistricts.value = []
    distDistrictId.value = null
    if (v && v !== DIST_ANY_CHILD) distDistricts.value = await fetchOptions(v)
    syncDistrict()
  }
  const syncDistrict = () => {
    if (distProvinceId.value === DIST_ANY_PROVINCE) {
      form.district = '不限'
      form.districtCn = '不限'
      return
    }
    const parts: string[] = []
    if (distProvinceId.value) {
      const p = distProvinces.value.find((x) => x.id === distProvinceId.value)
      if (p) parts.push(p.name.trim())
    }
    if (distCityId.value) {
      const c = distCities.value.find((x) => x.id === distCityId.value)
      if (c) parts.push(c.name.trim())
    }
    if (distDistrictId.value) {
      const d = distDistricts.value.find((x) => x.id === distDistrictId.value)
      if (d) parts.push(d.name.trim())
    }
    form.district = parts.join('/')
    form.districtCn = form.district
  }

  // 编辑回显：按 "/" 拆分名称逐级匹配
  const restoreResidence = async (full: string) => {
    if (!full) return
    const parts = full.split('/').map((s) => s.trim()).filter(Boolean)
    await loadResProvinces()
    const p = resProvinces.value.find((x) => x.name.trim() === parts[0])
    if (!p) return
    resProvinceId.value = p.id
    resCities.value = await fetchOptions(p.id)
    if (parts[1]) {
      const c = resCities.value.find((x) => x.name.trim() === parts[1])
      if (c) {
        resCityId.value = c.id
        resDistricts.value = await fetchOptions(c.id)
        if (parts[2]) {
          const d = resDistricts.value.find((x) => x.name.trim() === parts[2])
          if (d) resDistrictId.value = d.id
        }
      }
    }
  }
  const restoreDistrict = async (full: string) => {
    if (!full) return
    await loadDistProvinces()
    if (full === '不限') {
      distProvinceId.value = DIST_ANY_PROVINCE
      return
    }
    const parts = full.split('/').map((s) => s.trim()).filter(Boolean)
    const p = distProvinces.value.find((x) => x.name.trim() === parts[0])
    if (!p) return
    distProvinceId.value = p.id
    distCities.value = await fetchOptions(p.id)
    if (!parts[1]) {
      distCityId.value = DIST_ANY_CHILD
      return
    }
    const c = distCities.value.find((x) => x.name.trim() === parts[1])
    if (!c) return
    distCityId.value = c.id
    distDistricts.value = await fetchOptions(c.id)
    if (!parts[2]) {
      distDistrictId.value = DIST_ANY_CHILD
      return
    }
    const d = distDistricts.value.find((x) => x.name.trim() === parts[2])
    if (d) distDistrictId.value = d.id
  }
  const loadResProvinces = async () => {
    resProvinces.value = await fetchOptions(0)
  }
  const loadDistProvinces = async () => {
    distProvinces.value = await fetchOptions(0)
  }

  const addProject = () => {
    if (form.projects.length >= 6) {
      ElMessage.warning('项目经历最多 6 条')
      return
    }
    form.projects.push(emptyProject())
  }

  // P2：子表行删除二次确认（误删一整行经历代价高）
  const removeRow = async (arr: unknown[], index: number, label = '这一条') => {
    try {
      await ElMessageBox.confirm(`确认删除${label}？保存后该行从简历移除`, '提示', {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }
    arr.splice(index, 1)
  }

  // X1：数字下拉清空后为 undefined（显示占位符），提交前统一还原为 0（后端契约：数值字段 0=空）
  const toNum = (v: unknown, d = 0): number => (typeof v === 'number' && !Number.isNaN(v) ? v : d)
  const normalizeNums = () => {
    form.sex = toNum(form.sex)
    form.birthdate = toNum(form.birthdate)
    form.education = toNum(form.education)
    form.major = toNum(form.major)
    form.experience = toNum(form.experience)
    form.wageMin = toNum(form.wageMin)
    form.wageMax = toNum(form.wageMax)
    form.template = toNum(form.template, 1) || 1
    form.educations.forEach((e) => {
      e.education = toNum(e.education)
      e.startyear = toNum(e.startyear)
      e.startmonth = toNum(e.startmonth)
      e.endyear = toNum(e.endyear)
      e.endmonth = toNum(e.endmonth)
    })
    form.work.forEach((w) => {
      w.workType = toNum(w.workType, 1) || 1
      w.startyear = toNum(w.startyear)
      w.startmonth = toNum(w.startmonth)
      w.endyear = toNum(w.endyear)
      w.endmonth = toNum(w.endmonth)
    })
    form.language.forEach((l) => {
      l.language = toNum(l.language)
      l.level = toNum(l.level)
    })
    form.training.forEach((t) => {
      t.startyear = toNum(t.startyear)
      t.startmonth = toNum(t.startmonth)
      t.endyear = toNum(t.endyear)
      t.endmonth = toNum(t.endmonth)
    })
    form.credent.forEach((c) => {
      c.year = toNum(c.year)
      c.month = toNum(c.month)
    })
    form.projects.forEach((p) => {
      p.startyear = toNum(p.startyear)
      p.startmonth = toNum(p.startmonth)
      p.endyear = toNum(p.endyear)
      p.endmonth = toNum(p.endmonth)
    })
    form.skill.forEach((s) => {
      s.level = toNum(s.level)
    })
    form.studentLeader.forEach((s) => {
      s.startyear = toNum(s.startyear)
      s.startmonth = toNum(s.startmonth)
      s.endyear = toNum(s.endyear)
      s.endmonth = toNum(s.endmonth)
    })
  }

  // 下拉选中后把中文 label 写入 xxxCn 冗余字段（后端不查表生成）
  const fillCn = () => {
    const find = (group: { id: number; name: string }[], id: number) =>
      group.find((c) => c.id === id)?.name || ''
    form.sexCn = find(categories.value.sex, form.sex)
    form.educationCn = find(categories.value.education, form.education)
    form.majorCn = find(categories.value.major, form.major)
    form.experienceCn = find(categories.value.experience, form.experience)
    form.educations.forEach((e) => {
      e.educationCn = find(categories.value.education, e.education)
    })
    form.language.forEach((l) => {
      l.languageCn = find(languageOptions, l.language)
      l.levelCn = find(levelOptions, l.level)
    })
  }

  // 附件简历（PDF）上传：el-upload 直传，成功后刷新主表字段
  const uploadHeaders = computed(() => ({ Authorization: `Bearer ${userStore.token}` }))
  const outwardAction = computed(() => `/api/v1/personal/resumes/${resumeId}/outward`)
  const outwardUrl = computed(() => {
    const url = form.wordResume
    if (!url) return ''
    return /^https?:\/\//.test(url) ? url : `/${url}`
  })
  const beforeOutwardUpload = (file: File) => {
    if (!file.name.toLowerCase().endsWith('.pdf')) {
      ElMessage.warning('附件简历仅支持 PDF 格式')
      return false
    }
    if (file.size > 10 * 1024 * 1024) {
      ElMessage.warning('附件简历不能超过 10MB')
      return false
    }
    return true
  }
  const onOutwardSuccess = (res: { code: number; message?: string; data?: { url: string; title: string } }) => {
    if (res.code === 0 && res.data) {
      form.wordResume = res.data.url
      form.wordResumeTitle = res.data.title
      ElMessage.success('附件简历已上传')
    } else {
      ElMessage.error(res.message || '上传失败')
    }
  }
  const onOutwardError = () => {
    ElMessage.error('附件简历上传失败')
  }
  const handleRemoveOutward = async () => {
    try {
      await removeResumeOutward(resumeId)
      form.wordResume = ''
      form.wordResumeTitle = ''
      ElMessage.success('附件简历已删除')
    } catch {
      // 错误已由拦截器统一弹出
    }
  }

  const handleSave = async () => {
    if (!form.fullname.trim()) {
      ElMessage.warning('请填写姓名')
      return
    }
    if (form.wageMin && form.wageMax && form.wageMax < form.wageMin) {
      ElMessage.warning('期望薪资上限不能低于下限')
      return
    }
    if (form.projects.length > 6) {
      ElMessage.warning('项目经历最多 6 条')
      return
    }
    normalizeNums()
    fillCn()
    saving.value = true
    try {
      if (isEdit) {
        await updateResume(resumeId, form)
        ElMessage.success('简历已保存')
      } else {
        const { data } = await createResume(form)
        ElMessage.success('简历已创建')
        router.replace(`/personal/resumes/${data.id}`)
      }
    } finally {
      saving.value = false
    }
  }

  onMounted(async () => {
    categories.value = (await getCategories()).data
    await loadResProvinces()
    await loadDistProvinces()
    if (isEdit) {
      const { data } = await getResume(resumeId)
      Object.assign(form, data)
      form.template = data.template || 1
      // 回显子表（后端可能返回 null，兜底为空数组）；注意教育经历读 educations 复数键
      form.educations = data.educations || []
      form.work = (data.work || []).map((w) => ({ ...w, workType: w.workType || 1 }))
      form.language = data.language || []
      form.training = data.training || []
      form.credent = data.credent || []
      form.projects = data.projects || []
      form.skill = data.skill || []
      form.portfolio = data.portfolio || []
      form.studentLeader = data.studentLeader || []
      // 已有数据的加分项默认展开，便于查看
      extraPanels.value = (['skill', 'language', 'training', 'credent', 'portfolio', 'studentLeader'] as const).filter(
        (k) => (form[k] as unknown[]).length > 0
      )
      await restoreResidence(data.residence)
      await restoreDistrict(data.district)
      // 完善度详情（失败不阻塞编辑）
      getResumeCompleteness(resumeId)
        .then(({ data: c }) => {
          completeness.value = c
        })
        .catch(() => {})
    }
  })
</script>