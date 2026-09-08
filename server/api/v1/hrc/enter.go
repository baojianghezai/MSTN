package hrc

// ApiGroup 名硕人力业务域 API 组（与 GVA system/example/media 平级）
type ApiGroup struct {
	AuthApi           AuthApi
	AppealApi         AppealApi
	ProfileApi        ProfileApi
	AdminApi          AdminApi
	UploadApi         UploadApi
	CmsApi            CmsApi
	CompanyCancelApi  CompanyCancelApi
	ResumeApi         ResumeApi
	VideoInterviewApi VideoInterviewApi
	DashboardApi      DashboardApi
	ExportApi         ExportApi
	CompanyJobsApi    CompanyJobsApi
	AdminJobsApi      AdminJobsApi
	JobsApi           JobsApi
	CompanyPublicApi  CompanyPublicApi
	PersonalApplyApi  PersonalApplyApi
	CompanyApplyApi   CompanyApplyApi
	SetmealApi        SetmealApi
	OrderApi          OrderApi
	PaymentApi        PaymentApi
	PromotionApi      PromotionApi
	InterviewApi      InterviewApi
	MessageApi        MessageApi
	TalentApi         TalentApi
	CategoryApi       CategoryApi
}

var ApiGroupApp = new(ApiGroup)
