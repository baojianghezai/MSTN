package hrc

import (
	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/gin-gonic/gin"
)

type ProfileRouter struct{}

// InitProfileRouter 个人/企业资料路由（会员 JWT + utype 角色校验：09 §4.2/§4.3）
func (r *ProfileRouter) InitProfileRouter(Router *gin.RouterGroup) {
	personalGroup := Router.Group("personal").Use(middlewarehrc.MemberAuth()).Use(middlewarehrc.UtypeAuth(1))
	{
		personalGroup.GET("profile", hrcProfileApi.GetPersonalProfile)
		personalGroup.POST("profile", hrcProfileApi.UpdatePersonalProfile)
	}

	companyGroup := Router.Group("company").Use(middlewarehrc.MemberAuth()).Use(middlewarehrc.UtypeAuth(2))
	{
		companyGroup.GET("profile", hrcProfileApi.GetCompanyProfile)
		companyGroup.POST("profile", hrcProfileApi.UpdateCompanyProfile)
		companyGroup.POST("profile/logo", hrcProfileApi.UploadCompanyLogo)
		companyGroup.GET("profile/audit-status", hrcProfileApi.GetCompanyAuditStatus)
		// 企业注销（02 §2.8：#110a 申请 / #110b 状态查询）
		companyGroup.POST("cancel", hrcCompanyCancelApi.CompanyCancelApply)
		companyGroup.GET("cancel", hrcCompanyCancelApi.CompanyCancelStatus)
	}
}
