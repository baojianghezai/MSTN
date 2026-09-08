package hrc

import "github.com/gin-gonic/gin"

type AdminRouter struct{}

// InitAdminRouter 后台业务路由（挂 /api/v1/admin，受 GVA admin JWT 保护）
// 对应 09 §4.6.2：先实现申诉处理（#158/#159，含注销恢复）
func (r *AdminRouter) InitAdminRouter(Router *gin.RouterGroup) {
	Router.GET("setmeals", hrcSetmealApi.AdminList)
	Router.POST("setmeals", hrcSetmealApi.AdminCreate)
	Router.PUT("setmeals/:id", hrcSetmealApi.AdminUpdate)
	Router.GET("orders", hrcSetmealApi.AdminOrders)
	Router.POST("orders/:id/confirm", hrcSetmealApi.AdminConfirmOrder)

	appealAdmin := Router.Group("appeals")
	{
		appealAdmin.GET("", hrcAdminApi.AppealList)
		appealAdmin.PUT(":id", hrcAdminApi.ProcessAppeal)
	}

	// 系统配置（09 §4.6.2 编号 150）
	Router.GET("configs", hrcAdminApi.ConfigList)
	Router.PUT("configs", hrcAdminApi.ConfigSave)

	// 企业注销申请（09 §4.6.2 编号 161a-c）
	companyCancel := Router.Group("company/cancellations")
	{
		companyCancel.GET("", hrcAdminApi.CompanyCancellationList)
		companyCancel.POST(":id/handle", hrcAdminApi.CompanyCancellationHandle)
		companyCancel.DELETE(":id", hrcAdminApi.CompanyCancellationDelete)
	}

	// 企业资料审核（09 §4.6.2 编号 134/134a/135）
	companyProfiles := Router.Group("company-profiles")
	{
		companyProfiles.GET("", hrcAdminApi.CompanyProfileList)
		companyProfiles.GET(":id", hrcAdminApi.CompanyProfileDetail)
		companyProfiles.PUT(":id/audit", hrcAdminApi.CompanyProfileAudit)
	}

	// 视频面试（11 §四 P0#2）
	Router.GET("video-interviews", hrcVideoInterviewApi.AdminList)

	// 统计看板 + 报表（08 §4 A6，11 §四 P0#4：#119/#120 + 求职/企业分布）
	Router.GET("dashboard", hrcDashboardApi.Dashboard)
	Router.GET("dashboard/trend", hrcDashboardApi.DashboardTrend)
	Router.GET("statistics/resume", hrcDashboardApi.StatisticsResume)
	Router.GET("statistics/company", hrcDashboardApi.StatisticsCompany)

	// 后台导出（11 §四 P0#5：企业 + 职位）
	Router.POST("companies/export", hrcExportApi.ExportCompanies)
	Router.POST("jobs/export", hrcExportApi.ExportJobs)

	// 微信支付回调日志（11 §四 P0#7）
	Router.GET("wxpay-logs", hrcAdminApi.WxpayLogList)

	// 后台数据清理：固定目标预览、确认执行和审计记录
	Router.POST("data-cleanup/preview", hrcAdminApi.DataCleanupPreview)
	Router.POST("data-cleanup/execute", hrcAdminApi.DataCleanupExecute)
	Router.GET("data-cleanup/history", hrcAdminApi.DataCleanupHistory)

	// 职位管理/审核（03 模块，09 §4.6.2 #121/#122/#123）
	Router.GET("jobs", hrcAdminJobsApi.AdminListJobs)
	Router.GET("jobs/tmp", hrcAdminJobsApi.AdminListJobsTmp)
	Router.GET("jobs/:id", hrcAdminJobsApi.AdminGetJobDetail)
	Router.PUT("jobs/:id/audit", hrcAdminJobsApi.AuditJob)
}
