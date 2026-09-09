// 与 docs/api 契约对应的类型（字段小驼峰，与后端 json tag 一致）

// hrc 统一响应结构 {code, message, data}
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

// 登录/注册成功响应（01-auth.md #3/#5）
export interface LoginResult {
  token: string
  uid: number
  utype: number
  mobile: string // 当前登录手机号（解绑后为 unbound_<uid> 占位）
  passwordSet: boolean // 是否已设置密码（false 时引导走忘记密码重置设密）
}

// 前台导航项（03-cms.md #44）
export interface NavItem {
  id: number
  title: string
  url: string
  sort: number
  display: number
  addTime: number
}

// 申诉进度记录（02-account.md #46）
export interface AppealStatusItem {
  id: number
  uid: number
  realname: string
  mobile: string
  status: number // 0 待处理 / 1 已处理 / 2 已驳回
  statusCn: string
  addtime: number
}

// 个人资料（02-account.md #76）
export interface PersonalProfile {
  realname: string
  sex: number
  sexCn: string
  birthday: number // unix 秒
  residence: string
  education: number
  educationCn: string
  major: number
  majorCn: string
  experience: number
  experienceCn: string
  phone: string
  height: string
  marriage: number
  marriageCn: string
  displayName: number
  qq: string
  weixin: string
}

// 分类项（03-cms.md #21）
export interface CategoryItem {
  id: number
  groupId: number
  parentId: number
  name: string
  sort: number
  display: number
}

// 分类聚合（按分组别名）
export interface Categories {
  education: CategoryItem[]
  experience: CategoryItem[]
  wage: CategoryItem[]
  trade: CategoryItem[]
  district: CategoryItem[]
  major: CategoryItem[]
  sex: CategoryItem[]
  marriage: CategoryItem[]
  nature: CategoryItem[]
  scale: CategoryItem[]
  // 职位三级分类（ms_category_jobs 平移，parentId 层级；后端 seed 补后才有）
  jobcategory?: CategoryItem[]
  // 职位名称扁平列表（X3 修改批，50 常用岗位，#25 jobtitle 分组；用于发布表单 el-select 纯选择）
  jobtitle?: CategoryItem[]
  jobnature?: CategoryItem[]
  jobtag?: CategoryItem[]
  resumetag?: CategoryItem[]
  language?: CategoryItem[]
  languagelevel?: CategoryItem[]
  current?: CategoryItem[]
  age?: CategoryItem[]
}

// 职位三级分类
export interface CategoryTreeNode {
    id: number
    parentId: number
    level: number
    name: string
    children?: CategoryTreeNode[] // 递归引用
}

// 企业资料（02-account.md #107）
export interface CompanyProfile {
  companyname: string | null
  nature: number
  natureCn: string
  trade: number
  tradeCn: string
  district: string
  districtCn: string
  scale: number
  scaleCn: string
  registered: string
  address: string
  contact: string
  telephone: string
  landlineTel: string
  email: string
  website: string
  certificateImg: string
  logo: string
  contents: string
  shortName: string
  shortDesc: string
  tag: string
  audit: number
}

// 企业审核状态（02-account.md #110）
export interface CompanyAudit {
  audit: number // 0 未提交 / 1 已通过 / 2 审核中 / 3 未通过
  auditCn: string
}

// 企业注销申请（02-account.md #110b）
export interface CompanyCancellation {
  id: number
  companyname: string
  addtime: number
  status: number // 0 待处理 / 1 已处理
  statusCn: string
  finishtime: number
}

// 上传响应（03-cms.md #20 / 02-account.md #109）
export interface UploadResult {
  url: string
}

// 简历项目经历子表（04-resume.md #48/#50/#51，ms_resume_project，限 6 条）
export interface ResumeProject {
  id?: number
  pid?: number
  uid?: number
  startyear: number // 开始年
  startmonth: number // 开始月
  endyear: number // 结束年
  endmonth: number // 结束月
  todate: number // 至今标记（0=已结束 1=至今）
  projectname: string
  role: string
  description: string
}

// 简历教育经历子表（04-resume.md M3 收尾，ms_resume_education；数组键名 educations 复数）
export interface ResumeEducation {
  id?: number
  pid?: number
  uid?: number
  startyear: number
  startmonth: number
  endyear: number
  endmonth: number
  todate: number
  school: string
  speciality: string
  education: number // 学历编码（枚举同主表 education）
  educationCn: string
}

// 简历工作经历子表（ms_resume_work，数组键名 work）
export interface ResumeWork {
  id?: number
  pid?: number
  uid?: number
  startyear: number
  startmonth: number
  endyear: number
  endmonth: number
  todate: number
  companyname: string
  jobs: string
  achievements: string
}

// 简历语言能力子表（ms_resume_language，数组键名 language）
// 编码沿用 v6 字典：QS_language 208普通话/209粤语/210英语/211法语/212日语/213其他
//             QS_language_level 291入门/292熟练/293精通
export interface ResumeLanguage {
  id?: number
  pid?: number
  uid?: number
  language: number
  languageCn: string
  level: number
  levelCn: string
}

// 简历培训经历子表（ms_resume_training，数组键名 training）
export interface ResumeTraining {
  id?: number
  pid?: number
  uid?: number
  startyear: number
  startmonth: number
  endyear: number
  endmonth: number
  todate: number
  agency: string
  course: string
  description: string
}

// 简历证书子表（ms_resume_credent，数组键名 credent；images 一期预留无上传 UI）
export interface ResumeCredent {
  id?: number
  pid?: number
  uid?: number
  name: string
  year: number
  month: number
  images: string
}

// 简历列表轻量项（#49，不返子表）
export interface ResumeLite {
  id: number
  title: string
  completePercent: number
  def: number // 1=默认简历
  display: number // 1=公开 2=不公开
  addtime: number
  refreshtime: number
}

// 简历完善度详情（#57，与主表 completePercent 同源）
export interface ResumeCompleteness {
  percent: number
  categories: { key: string; name: string; max: number; score: number }[]
  missing: string[]
}

// 简历主表可编辑字段 + 项目经历子表（04-resume.md #48/#50/#51）
export interface Resume {
  id?: number
  title: string
  fullname: string
  sex: number
  sexCn: string
  birthdate: number // 出生年
  residence: string
  education: number
  educationCn: string
  major: number
  majorCn: string
  experience: number
  experienceCn: string
  district: string
  districtCn: string
  wage: number
  wageCn: string
  intentionJobs: string
  specialty: string
  telephone: string
  email: string
  displayName: number // 1=显示姓名 2=匿名
  current: number
  currentCn: string
  mobileAudit: number
  talent: number
  entrust: number
  // 6 子表（M3 收尾 #48/#50/#51 全量替换；注意教育经历数组键是 educations 复数，
  // 与主表学历编码 education 区分——handoff/2026-08-20-resume-m3.md ⚠️ 易踩点）
  educations: ResumeEducation[]
  work: ResumeWork[]
  language: ResumeLanguage[]
  training: ResumeTraining[]
  credent: ResumeCredent[]
  projects: ResumeProject[]
}

// 视频面试（05-video-interview.md，ms_video_interview）
export interface VideoInterview {
  id?: number
  companyUid?: number
  personalUid?: number
  jobsId?: number
  jobsName: string
  interviewTime: number // 面试时间 unix 秒
  deadline?: number // 房间有效期（面试时间+15天）
  contact?: string
  contactTel?: string
  addtime?: number
  companyCode?: string // 企业房间码（6 位）
  personalCode?: string // 个人房间码（6 位）
  roomStatus?: string // nostart / opened / overtime
  resumeId?: number
  fullname?: string // 简历姓名（列表回显）
  companyname?: string // 企业名（后台列表）
}

// 视频面试发起请求体
export interface VideoInterviewCreate {
  resumeId: number
  jobsId: number
  jobsName: string
  interviewTime: number
  contact: string
  telephone: string
}

// 房间码查询结果（公开，不含联系方式）
export interface VideoInterviewRoom {
  id: number
  jobsName: string
  interviewTime: number
  deadline: number
  roomStatus: string
  utype: number // 1=个人 2=企业
}

// 路由 meta 扩展
declare module 'vue-router' {
  interface RouteMeta {
    requiresAuth?: boolean
    utype?: number
  }
}

// 职位联系方式（09-jobs.md，ms_jobs_contact）
export interface JobsContact {
  contact: string
  qq: string
  telephone: string
  landlineTel: string
  address: string
  email: string
}

// 职位发布/编辑请求体（09-jobs.md #79/#82）
export interface JobsRequest {
  jobsName: string
  nature: number // 工作性质 code
  natureCn: string
  sex: number // 性别要求 3=不限
  amount: number // 招聘人数 0-99
  topclass: number // 一级分类
  category: number // 二级分类
  subclass: number // 三级分类
  categoryCn: string // 分类中文（随选随传）
  trade: number // 行业
  district: string // 地区 code
  districtCn: string
  tag: string // 标签（逗号分隔 code）
  education: number // 学历要求
  experience: number // 经验要求
  minwage: number // 薪资 元/月
  maxwage: number
  negotiable: number // 面议 0/1
  contents: string // 职位描述 ≤4000
  deadline: number // 有效期 unix 秒
  department: string
  mapX: number
  mapY: number
  mapZoom: number
  contact: JobsContact
  tags: number[] // 标签数组 uint32
}

// 职位列表项（09-jobs.md #80，jobs + jobs_tmp 合并）
export interface JobItem {
  id: number
  jobsId: number // tmp 场景：原 jobs id
  jobsName: string
  companyname: string
  companyId: number
  emergency: number
  stick: number
  nature: number
  natureCn: string
  sex: number
  amount: number
  topclass: number
  category: number
  subclass: number
  categoryCn: string
  trade: number
  district: string
  districtCn: string
  tag: string
  education: number
  experience: number
  minwage: number
  maxwage: number
  negotiable: number
  contents: string
  addtime: number
  deadline: number
  refreshtime: number
  audit: number // 0 草稿 1 通过 2 待审 3 不通过
  display: number // 1 展示 2 暂停
  click: number
  department: string
  mapX: number
  mapY: number
  mapZoom: number
  pending: boolean // 待审（tmp audit=2）
}

// 职位详情（09-jobs.md #81，编辑回显）
export interface JobDetail extends JobItem {
  contact: JobsContact
  tags: number[]
  reason: string // 不通过原因（audit=3）
}
