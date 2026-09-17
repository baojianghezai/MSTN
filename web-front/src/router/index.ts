import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import PublicLayout from '@/layouts/PublicLayout.vue'
import PersonalLayout from '@/layouts/PersonalLayout.vue'

// 路由按「布局域」组织（design/10 §3.3）：
// PublicLayout 域（公开）+ 个人中心域（PersonalLayout，需登录 utype=1）
// + 企业中心域（CompanyLayout，需登录 utype=2）
const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: PublicLayout,
    children: [
      {
        path: '',
        name: 'Home',
        meta: { title: '首页' },
        component: () => import('@/views/home/index.vue')
      },
      {
        path: 'login',
        name: 'Login',
        meta: { title: '登录' },
        component: () => import('@/views/auth/login.vue')
      },
      {
        path: 'register',
        name: 'Register',
        meta: { title: '注册' },
        component: () => import('@/views/auth/register.vue')
      },
      {
        path: 'forgot-password',
        name: 'ForgotPassword',
        meta: { title: '重置密码' },
        component: () => import('@/views/auth/forgot-password.vue')
      },
      {
        path: 'appeal',
        name: 'Appeal',
        meta: { title: '账号申诉' },
        component: () => import('@/views/appeal/index.vue')
      },
      {
        path: 'jobs',
        name: 'Jobs',
        meta: { title: '职位列表' },
        component: () => import('@/views/jobs/list.vue')
      },
      {
        path: 'jobs/:id',
        name: 'JobDetail',
        meta: { title: '职位详情' },
        component: () => import('@/views/jobs/detail.vue')
      },
      {
        path: 'talents',
        name: 'Talents',
        meta: { title: '找人才' },
        component: () => import('@/views/talents/list.vue')
      },
      // 占位页：后台导航已挂出但后端接口尚未实现（企业/招聘会/资讯 M3/M6，help 内容接口未就绪）
      {
        path: 'companies',
        name: 'Companies',
        meta: { title: '找企业' },
        component: () => import('@/views/companies/list.vue')
      },
      {
        path: 'companies/:id',
        name: 'CompanyDetail',
        meta: { title: '企业详情' },
        component: () => import('@/views/companies/detail.vue')
      },
      {
        path: 'jobfairs',
        name: 'Jobfairs',
        meta: { title: '招聘会', contentType: 2, contentTitle: '招聘会', detailName: 'JobfairDetail' },
        component: () => import('@/views/content/list.vue')
      },
      {
        path: 'jobfairs/:id',
        name: 'JobfairDetail',
        meta: { title: '招聘会详情' },
        component: () => import('@/views/content/detail.vue')
      },
      {
        path: 'news',
        name: 'News',
        meta: { title: '资讯', contentType: 1, contentTitle: '资讯', detailName: 'NewsDetail' },
        component: () => import('@/views/content/list.vue')
      },
      {
        path: 'news/:id',
        name: 'NewsDetail',
        meta: { title: '资讯详情' },
        component: () => import('@/views/content/detail.vue')
      },
      {
        path: 'help',
        name: 'Help',
        meta: { title: '帮助', contentType: 3, contentTitle: '帮助中心', detailName: 'HelpDetail' },
        component: () => import('@/views/content/list.vue')
      },
      {
        path: 'help/:id',
        name: 'HelpDetail',
        meta: { title: '帮助详情' },
        component: () => import('@/views/content/detail.vue')
      }
    ]
  },
  {
    path: '/personal',
    component: PersonalLayout,
    meta: { title: '个人中心', requiresAuth: true, utype: 1 },
    children: [
      {
        path: '',
        name: 'Personal',
        meta: { title: '个人中心' },
        component: () => import('@/views/personal/index.vue')
      },
      {
        path: 'profile',
        name: 'PersonalProfile',
        meta: { title: '个人资料' },
        component: () => import('@/views/personal/profile.vue')
      },
      {
        path: 'resumes',
        name: 'PersonalResumes',
        meta: { title: '我的简历' },
        component: () => import('@/views/personal/resumes.vue')
      },
      {
        path: 'resumes/new',
        name: 'ResumeNew',
        meta: { title: '新建简历' },
        component: () => import('@/views/personal/resume-edit.vue')
      },
      {
        path: 'resumes/:id',
        name: 'ResumeEdit',
        meta: { title: '编辑简历' },
        component: () => import('@/views/personal/resume-edit.vue')
      },
      {
        path: 'video-interviews',
        name: 'PersonalVideoInterviews',
        meta: { title: '视频面试' },
        component: () => import('@/views/personal/video-interviews.vue')
      },
      {
        path: 'interviews',
        name: 'PersonalInterviews',
        meta: { title: '面试邀请' },
        component: () => import('@/views/personal/interviews.vue')
      },
      {
        path: 'messages',
        name: 'PersonalMessages',
        meta: { title: '站内信' },
        component: () => import('@/views/personal/messages.vue')
      },
      {
        path: 'chat',
        name: 'PersonalChat',
        meta: { title: '在线对话', chatScope: 'personal' },
        component: () => import('@/views/chat/index.vue')
      },
      {
        path: 'applies',
        name: 'PersonalApplies',
        meta: { title: '我的投递' },
        component: () => import('@/views/personal/applies.vue')
      }
    ]
  },
  {
    path: '/video-interviews/room/:code',
    name: 'VideoInterviewRoom',
    meta: { title: '视频面试房间' },
    component: () => import('@/views/video-interview-room.vue')
  },
  {
    path: '/company',
    component: () => import('@/layouts/CompanyLayout.vue'),
    meta: { requiresAuth: true, utype: 2 },
    children: [
      {
        path: '',
        name: 'Company',
        meta: { title: '企业中心' },
        component: () => import('@/views/company/index.vue')
      },
      {
        path: 'profile',
        name: 'CompanyProfile',
        meta: { title: '企业资料' },
        component: () => import('@/views/company/profile.vue')
      },
      {
        path: 'cancellation',
        name: 'CompanyCancellation',
        meta: { title: '企业注销' },
        component: () => import('@/views/company/cancellation.vue')
      },
      {
        path: 'video-interviews',
        name: 'CompanyVideoInterviews',
        meta: { title: '视频面试' },
        component: () => import('@/views/company/video-interviews.vue')
      },
      {
        path: 'interviews',
        name: 'CompanyInterviews',
        meta: { title: '面试邀请' },
        component: () => import('@/views/company/interviews.vue')
      },
      {
        path: 'messages',
        name: 'CompanyMessages',
        meta: { title: '站内信' },
        component: () => import('@/views/company/messages.vue')
      },
      {
        path: 'chat',
        name: 'CompanyChat',
        meta: { title: '在线对话', chatScope: 'company' },
        component: () => import('@/views/chat/index.vue')
      },
      {
        path: 'talents',
        name: 'CompanyTalents',
        meta: { title: '找人才' },
        component: () => import('@/views/company/talents.vue')
      },
      {
        path: 'talent-library',
        name: 'CompanyTalentLibrary',
        meta: { title: '人才库' },
        component: () => import('@/views/company/talent-library.vue')
      },
      {
        path: 'jobs',
        name: 'CompanyJobs',
        meta: { title: '职位管理' },
        component: () => import('@/views/company/jobs.vue')
      },
      {
        path: 'applies',
        name: 'CompanyApplies',
        meta: { title: '收到的简历' },
        component: () => import('@/views/company/applies.vue')
      },
      {
        path: 'plan',
        name: 'CompanyPlan',
        meta: { title: '套餐与订单' },
        component: () => import('@/views/company/plan.vue')
      },
      {
        path: 'promotions',
        name: 'CompanyPromotions',
        meta: { title: '首页推广' },
        component: () => import('@/views/company/promotions.vue')
      },
      {
        path: 'hrs',
        name: 'CompanyHRs',
        meta: { title: '员工账号' },
        component: () => import('@/views/company/hrs.vue')
      },
      {
        path: 'jobfairs',
        name: 'CompanyJobfairs',
        meta: { title: '举办招聘会' },
        component: () => import('@/views/company/jobfairs.vue')
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    meta: { title: '页面不存在' },
    component: () => import('@/views/error/404.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
