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

        <el-form-item label="简历标题">
          <el-input v-model="form.title" maxlength="80" placeholder="如：后端开发工程师" clearable />
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
            <el-input v-model="form.residence" maxlength="30" placeholder="如：广东广州" clearable />
          </el-form-item>
        </div>

        <div class="grid grid-cols-1 gap-x-6 sm:grid-cols-2">
          <el-form-item label="学历">
            <el-select v-model="form.education" placeholder="请选择" class="w-full" clearable :empty-values="[null, undefined, 0]">
              <el-option v-for="c in categories.education" :key="c.id" :label="c.name" :value="c.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="专业">
            <el-select v-model="form.major" placeholder="请选择" class="w-full" clearable :empty-values="[null, undefined, 0]">
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
            <el-select v-model="form.wage" placeholder="请选择" class="w-full" clearable :empty-values="[null, undefined, 0]">
              <el-option v-for="c in categories.wage" :key="c.id" :label="c.name" :value="c.id" />
            </el-select>
          </el-form-item>
        </div>

        <el-form-item label="期望职位">
          <el-input v-model="form.intentionJobs" maxlength="255" placeholder="如：Go 后端开发" clearable />
        </el-form-item>

        <el-form-item label="期望地区">
          <el-input
            v-model="form.districtCn"
            maxlength="30"
            placeholder="如：广州"
            clearable
            @change="form.district = form.districtCn"
          />
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

      <!-- 工作经历（子表 work，不限条数） -->
      <el-card id="sec-work" shadow="never" class="mt-6 scroll-mt-32">
        <template #header>
          <div class="flex items-center justify-between">
            <span class="font-semibold">工作经历</span>
            <span class="text-xs text-slate-400">{{ form.work.length }} 条</span>
          </div>
        </template>

        <div v-for="(w, i) in form.work" :key="i" class="mb-4 rounded border border-slate-200 p-4">
          <div class="flex items-center justify-between">
            <span class="text-sm font-medium text-slate-700">工作 {{ i + 1 }}</span>
            <el-button type="danger" text size="small" @click="removeRow(form.work, i)">
              删除
            </el-button>
          </div>

          <div class="mt-3 grid grid-cols-1 gap-x-6 sm:grid-cols-2">
            <el-form-item label="公司" label-width="80px">
              <el-input v-model="w.companyname" maxlength="50" placeholder="公司名称" clearable />
            </el-form-item>
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
          添加工作经历
        </el-button>
      </el-card>

      <!-- 语言能力（子表 language，不限条数；编码沿用 v6 字典） -->
      <el-card id="sec-language" shadow="never" class="mt-6 scroll-mt-32">
        <template #header>
          <div class="flex items-center justify-between">
            <span class="font-semibold">语言能力</span>
            <span class="text-xs text-slate-400">{{ form.language.length }} 条</span>
          </div>
        </template>

        <div v-for="(l, i) in form.language" :key="i" class="mb-4 rounded border border-slate-200 p-4">
          <div class="flex items-center justify-between">
            <span class="text-sm font-medium text-slate-700">语言 {{ i + 1 }}</span>
            <el-button type="danger" text size="small" @click="removeRow(form.language, i)">
              删除
            </el-button>
          </div>

          <div class="mt-3 grid grid-cols-1 gap-x-6 sm:grid-cols-2">
            <el-form-item label="语言" label-width="80px">
              <el-select v-model="l.language" placeholder="请选择" class="w-full" clearable :empty-values="[null, undefined, 0]">
                <el-option v-for="c in languageOptions" :key="c.id" :label="c.name" :value="c.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="熟练度" label-width="80px">
              <el-select v-model="l.level" placeholder="请选择" class="w-full" clearable :empty-values="[null, undefined, 0]">
                <el-option v-for="c in levelOptions" :key="c.id" :label="c.name" :value="c.id" />
              </el-select>
            </el-form-item>
          </div>
        </div>

        <el-button class="w-full" @click="form.language.push(emptyLanguage())">
          添加语言能力
        </el-button>
      </el-card>

      <!-- 培训经历（子表 training，不限条数） -->
      <el-card id="sec-training" shadow="never" class="mt-6 scroll-mt-32">
        <template #header>
          <div class="flex items-center justify-between">
            <span class="font-semibold">培训经历</span>
            <span class="text-xs text-slate-400">{{ form.training.length }} 条</span>
          </div>
        </template>

        <div
          v-for="(t, i) in form.training"
          :key="i"
          class="mb-4 rounded border border-slate-200 p-4"
        >
          <div class="flex items-center justify-between">
            <span class="text-sm font-medium text-slate-700">培训 {{ i + 1 }}</span>
            <el-button type="danger" text size="small" @click="removeRow(form.training, i)">
              删除
            </el-button>
          </div>

          <div class="mt-3 grid grid-cols-1 gap-x-6 sm:grid-cols-2">
            <el-form-item label="机构" label-width="80px">
              <el-input v-model="t.agency" maxlength="50" placeholder="培训机构" clearable />
            </el-form-item>
            <el-form-item label="课程" label-width="80px">
              <el-input v-model="t.course" maxlength="50" placeholder="课程名称" clearable />
            </el-form-item>
          </div>

          <div class="grid grid-cols-1 gap-x-6 sm:grid-cols-2">
            <el-form-item label="起止时间" label-width="80px">
              <div class="flex w-full items-center gap-2">
                <el-select v-model="t.startyear" placeholder="开始年" class="flex-1" clearable filterable :empty-values="[null, undefined, 0]">
                  <el-option v-for="y in projectYears" :key="y" :label="y" :value="y" />
                </el-select>
                <el-select v-model="t.startmonth" placeholder="月" class="flex-1" clearable :empty-values="[null, undefined, 0]">
                  <el-option v-for="m in 12" :key="m" :label="`${m}月`" :value="m" />
                </el-select>
              </div>
            </el-form-item>
            <el-form-item label="结束时间" label-width="80px">
              <div class="flex w-full items-center gap-2">
                <el-select
                  v-model="t.endyear"
                  placeholder="年"
                  class="flex-1"
                  clearable
                  filterable
                  :disabled="t.todate === 1"
                  :empty-values="[null, undefined, 0]"
                >
                  <el-option v-for="y in projectYears" :key="y" :label="y" :value="y" />
                </el-select>
                <el-select
                  v-model="t.endmonth"
                  placeholder="月"
                  class="flex-1"
                  clearable
                  :disabled="t.todate === 1"
                  :empty-values="[null, undefined, 0]"
                >
                  <el-option v-for="m in 12" :key="m" :label="`${m}月`" :value="m" />
                </el-select>
                <el-checkbox v-model="t.todate" :true-value="1" :false-value="0">至今</el-checkbox>
              </div>
            </el-form-item>
          </div>

          <el-form-item label="培训描述" label-width="80px">
            <el-input
              v-model="t.description"
              type="textarea"
              :rows="2"
              maxlength="1000"
              show-word-limit
              placeholder="培训内容、收获"
            />
          </el-form-item>
        </div>

        <el-button class="w-full" @click="form.training.push(emptyTraining())">
          添加培训经历
        </el-button>
      </el-card>

      <!-- 证书（子表 credent，不限条数；images 一期无上传 UI，预留字段） -->
      <el-card id="sec-credential" shadow="never" class="mt-6 scroll-mt-32">
        <template #header>
          <div class="flex items-center justify-between">
            <span class="font-semibold">证书</span>
            <span class="text-xs text-slate-400">{{ form.credent.length }} 条</span>
          </div>
        </template>

        <div v-for="(c, i) in form.credent" :key="i" class="mb-4 rounded border border-slate-200 p-4">
          <div class="flex items-center justify-between">
            <span class="text-sm font-medium text-slate-700">证书 {{ i + 1 }}</span>
            <el-button type="danger" text size="small" @click="removeRow(form.credent, i)">
              删除
            </el-button>
          </div>

          <div class="mt-3 grid grid-cols-1 gap-x-6 sm:grid-cols-3">
            <el-form-item label="名称" label-width="80px">
              <el-input v-model="c.name" maxlength="255" placeholder="证书名称" clearable />
            </el-form-item>
            <el-form-item label="年份" label-width="80px">
              <el-select v-model="c.year" placeholder="年" class="w-full" clearable filterable :empty-values="[null, undefined, 0]">
                <el-option v-for="y in projectYears" :key="y" :label="y" :value="y" />
              </el-select>
            </el-form-item>
            <el-form-item label="月份" label-width="80px">
              <el-select v-model="c.month" placeholder="月" class="w-full" clearable :empty-values="[null, undefined, 0]">
                <el-option v-for="m in 12" :key="m" :label="`${m}月`" :value="m" />
              </el-select>
            </el-form-item>
          </div>
        </div>

        <el-button class="w-full" @click="form.credent.push(emptyCredent())">
          添加证书
        </el-button>
        <p class="mt-2 text-xs text-slate-400">证书图片上传一期暂不支持，仅登记名称与时间。</p>
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

      <!-- P2：长表单 sticky 保存栏，滚动任何位置都能直接保存 -->
      <div class="sticky bottom-0 z-10 -mx-4 mt-8 flex justify-end gap-3 border-t border-slate-200 bg-white/95 px-4 py-3 shadow-[0_-4px_12px_rgba(15,23,42,0.08)] backdrop-blur">
        <el-button @click="router.back()">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存简历</el-button>
      </div>
    </el-form>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, reactive, ref } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    createResume,
    getResume,
    getResumeCompleteness,
    updateResume
  } from '@/api/resume'
  import { getCategories } from '@/api/content'
  import type {
    Categories,
    Resume,
    ResumeCompleteness,
    ResumeCredent,
    ResumeEducation,
    ResumeLanguage,
    ResumeTraining,
    ResumeWork
  } from '@/types/api'

  const route = useRoute()
  const router = useRouter()

  const resumeId = Number(route.params.id)
  const isEdit = !Number.isNaN(resumeId) && resumeId > 0

  // P2：长表单锚点导航（与模板里各 el-card 的 id 对应）
  const sections = [
    { id: 'sec-basic', label: '基本信息' },
    { id: 'sec-education', label: '教育经历' },
    { id: 'sec-work', label: '工作经历' },
    { id: 'sec-language', label: '语言能力' },
    { id: 'sec-training', label: '培训经历' },
    { id: 'sec-credential', label: '证书' },
    { id: 'sec-project', label: '项目经历' }
  ]

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

  const emptyResume = (): Resume => ({
    title: '',
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
    wage: 0,
    wageCn: '',
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
    educations: [],
    work: [],
    language: [],
    training: [],
    credent: [],
    projects: []
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

  // 出生年范围：近 60 年倒序
  const currentYear = new Date().getFullYear()
  const birthYears = Array.from({ length: 60 }, (_, i) => currentYear - i)
  // 子表年份范围：近 50 年（可到当前年）
  const projectYears = Array.from({ length: 50 }, (_, i) => currentYear - i)

  const addProject = () => {
    if (form.projects.length >= 6) {
      ElMessage.warning('项目经历最多 6 条')
      return
    }
    form.projects.push(emptyProject())
  }

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
    form.wage = toNum(form.wage)
    form.educations.forEach((e) => {
      e.education = toNum(e.education)
      e.startyear = toNum(e.startyear)
      e.startmonth = toNum(e.startmonth)
      e.endyear = toNum(e.endyear)
      e.endmonth = toNum(e.endmonth)
    })
    form.work.forEach((w) => {
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
  }

  // 下拉选中后把中文 label 写入 xxxCn 冗余字段（后端不查表生成）
  const fillCn = () => {
    const find = (group: { id: number; name: string }[], id: number) =>
      group.find((c) => c.id === id)?.name || ''
    form.sexCn = find(categories.value.sex, form.sex)
    form.educationCn = find(categories.value.education, form.education)
    form.majorCn = find(categories.value.major, form.major)
    form.experienceCn = find(categories.value.experience, form.experience)
    form.wageCn = find(categories.value.wage, form.wage)
    form.educations.forEach((e) => {
      e.educationCn = find(categories.value.education, e.education)
    })
    form.language.forEach((l) => {
      l.languageCn = find(languageOptions, l.language)
      l.levelCn = find(levelOptions, l.level)
    })
  }

  const handleSave = async () => {
    if (!form.title.trim()) {
      ElMessage.warning('请填写简历标题')
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
    if (isEdit) {
      const { data } = await getResume(resumeId)
      Object.assign(form, data)
      // 回显 6 子表（后端可能返回 null，兜底为空数组）；注意教育经历读 educations 复数键
      form.educations = data.educations || []
      form.work = data.work || []
      form.language = data.language || []
      form.training = data.training || []
      form.credent = data.credent || []
      form.projects = data.projects || []
      // 完善度详情（失败不阻塞编辑）
      getResumeCompleteness(resumeId)
        .then(({ data: c }) => {
          completeness.value = c
        })
        .catch(() => {})
    }
  })
</script>
