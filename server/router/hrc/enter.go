package hrc

import (
	api "github.com/flipped-aurora/gin-vue-admin/server/api/v1"
)

// RouterGroup 名硕人力业务域路由组
type RouterGroup struct {
	AuthRouter
	AppealRouter
	ProfileRouter
	AdminRouter
	UploadRouter
	CmsRouter
	ResumeRouter
	VideoInterviewRouter
	JobsRouter
	CompanyPublicRouter
	ApplyRouter
	SetmealRouter
	OrderRouter
	PaymentRouter
	PromotionRouter
	InterviewRouter
	MessageRouter
	TalentRouter
	ChatRouter
	Category
}

var (
	hrcAuthApi           = api.ApiGroupApp.HrcApiGroup.AuthApi
	hrcAppealApi         = api.ApiGroupApp.HrcApiGroup.AppealApi
	hrcProfileApi        = api.ApiGroupApp.HrcApiGroup.ProfileApi
	hrcAdminApi          = api.ApiGroupApp.HrcApiGroup.AdminApi
	hrcUploadApi         = api.ApiGroupApp.HrcApiGroup.UploadApi
	hrcCmsApi            = api.ApiGroupApp.HrcApiGroup.CmsApi
	hrcCompanyCancelApi  = api.ApiGroupApp.HrcApiGroup.CompanyCancelApi
	hrcResumeApi         = api.ApiGroupApp.HrcApiGroup.ResumeApi
	hrcVideoInterviewApi = api.ApiGroupApp.HrcApiGroup.VideoInterviewApi
	hrcDashboardApi      = api.ApiGroupApp.HrcApiGroup.DashboardApi
	hrcExportApi         = api.ApiGroupApp.HrcApiGroup.ExportApi
	hrcCompanyJobsApi    = api.ApiGroupApp.HrcApiGroup.CompanyJobsApi
	hrcAdminJobsApi      = api.ApiGroupApp.HrcApiGroup.AdminJobsApi
	hrcJobsApi           = api.ApiGroupApp.HrcApiGroup.JobsApi
	hrcCompanyPublicApi  = api.ApiGroupApp.HrcApiGroup.CompanyPublicApi
	hrcPersonalApplyApi  = api.ApiGroupApp.HrcApiGroup.PersonalApplyApi
	hrcCompanyApplyApi   = api.ApiGroupApp.HrcApiGroup.CompanyApplyApi
	hrcSetmealApi        = api.ApiGroupApp.HrcApiGroup.SetmealApi
	hrcOrderApi          = api.ApiGroupApp.HrcApiGroup.OrderApi
	hrcPaymentApi        = api.ApiGroupApp.HrcApiGroup.PaymentApi
	hrcPromotionApi      = api.ApiGroupApp.HrcApiGroup.PromotionApi
	hrcInterviewApi      = api.ApiGroupApp.HrcApiGroup.InterviewApi
	hrcMessageApi        = api.ApiGroupApp.HrcApiGroup.MessageApi
	hrcTalentApi         = api.ApiGroupApp.HrcApiGroup.TalentApi
	hrcCategoryApi       = api.ApiGroupApp.HrcApiGroup.CategoryApi
	hrcChatApi           = api.ApiGroupApp.HrcApiGroup.ChatApi
	hrcChatWsApi         = api.ApiGroupApp.HrcApiGroup.ChatWsApi
)
