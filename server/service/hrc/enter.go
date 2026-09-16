package hrc

// ServiceGroup 名硕人力业务域服务组（通过 service.ServiceGroupApp.HrcServiceGroup 访问）
type ServiceGroup struct {
	AuthService                AuthService
	AppealService              AppealService
	ProfileService             ProfileService
	AdminService               AdminService
	UploadService              UploadService
	CmsService                 CmsService
	CompanyCancellationService CompanyCancellationService
	CompanyAuditService        CompanyAuditService
	ResumeService              ResumeService
	VideoInterviewService      VideoInterviewService
	DashboardService           DashboardService
	ExportService              ExportService
	WxpayLogService            WxpayLogService
	JobsService                JobsService
	JobsSearchService          JobsSearchService
	CompanySearchService       CompanySearchService
	ApplyService               ApplyService
	CompanyApplyService        CompanyApplyService
	SetmealService             SetmealService
	OrderService               OrderService
	PaymentService             PaymentService
	PromotionService           PromotionService
	DataCleanupService         DataCleanupService
	InterviewService           InterviewService
	MessageService             MessageService
	TalentService              TalentService
	CategoryService            CategoryService
	ChatService                ChatService
}

var ServiceGroupApp = new(ServiceGroup)
