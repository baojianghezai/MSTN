package hrc

import (
	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/gin-gonic/gin"
)

// CompanyHRRouter 企业 HR 子账号路由（#21；仅企业主账号可管理）
type CompanyHRRouter struct{}

// InitCompanyHRRouter 注册 HR 子账号管理路由
func (r *CompanyHRRouter) InitCompanyHRRouter(Router *gin.RouterGroup) {
	group := Router.Group("company").Use(middlewarehrc.MemberAuth()).Use(middlewarehrc.UtypeAuth(2))
	group.GET("hrs", hrcCompanyHRApi.List)
	group.POST("hrs", hrcCompanyHRApi.Create)
	group.PUT("hrs/:uid/status", hrcCompanyHRApi.SetStatus)
	group.PUT("hrs/:uid/password", hrcCompanyHRApi.ResetPassword)
	group.DELETE("hrs/:uid", hrcCompanyHRApi.Delete)
}
