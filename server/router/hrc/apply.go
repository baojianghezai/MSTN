package hrc

import (
	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/gin-gonic/gin"
)

type ApplyRouter struct{}

// InitApplyRouter 投递路由（个人端 utype=1 + 企业端 utype=2；09 §4.2 #61-63 / §4.3 #90/#92-94）
func (r *ApplyRouter) InitApplyRouter(Router *gin.RouterGroup) {
	personalGroup := Router.Group("personal").Use(middlewarehrc.MemberAuth()).Use(middlewarehrc.UtypeAuth(1))
	{
		personalGroup.POST("applies", hrcPersonalApplyApi.Apply)
		personalGroup.GET("applies", hrcPersonalApplyApi.List)
		personalGroup.DELETE("applies/:did", hrcPersonalApplyApi.Delete)
	}

	companyGroup := Router.Group("company").Use(middlewarehrc.MemberAuth()).Use(middlewarehrc.UtypeAuth(2))
	{
		companyGroup.GET("applies", hrcCompanyApplyApi.List)
		companyGroup.PUT("applies/:did/looked", hrcCompanyApplyApi.Looked)
		companyGroup.PUT("applies/:did/reply", hrcCompanyApplyApi.Reply)
		companyGroup.GET("applies/:did/resume/download", hrcCompanyApplyApi.DownloadResume)
		companyGroup.GET("jobs/:id/applies", hrcCompanyApplyApi.JobApplies)
	}
}
