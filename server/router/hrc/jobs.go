package hrc

import (
	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/gin-gonic/gin"
)

type JobsRouter struct{}

// InitJobsRouter 职位路由（企业端 utype=2 + 前台公开；09 §4.3 #79-86 / §4.1 #22-25）
func (r *JobsRouter) InitJobsRouter(Router *gin.RouterGroup) {
	companyGroup := Router.Group("company").Use(middlewarehrc.MemberAuth()).Use(middlewarehrc.UtypeAuth(2))
	{
		companyGroup.POST("jobs", hrcCompanyJobsApi.CreateJob)
		companyGroup.GET("jobs", hrcCompanyJobsApi.ListJobs)
		companyGroup.GET("jobs/:id", hrcCompanyJobsApi.GetJob)
		companyGroup.PUT("jobs/:id", hrcCompanyJobsApi.UpdateJob)
		companyGroup.DELETE("jobs/:id", hrcCompanyJobsApi.DeleteJob)
		companyGroup.PUT("jobs/:id/pause", hrcCompanyJobsApi.PauseJob)
		companyGroup.PUT("jobs/:id/resume", hrcCompanyJobsApi.ResumeJob)
		companyGroup.PUT("jobs/:id/refresh", hrcCompanyJobsApi.RefreshJob)
	}

	// 前台公开职位（09 §4.1 #22-25）
	Router.GET("jobs", hrcJobsApi.List)
	Router.GET("jobs/hot-words", hrcJobsApi.HotWords)
	Router.GET("jobs/filters", hrcJobsApi.Filters)
	Router.GET("jobs/:id", hrcJobsApi.Detail)
}
