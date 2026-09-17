package hrc

import (
	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/gin-gonic/gin"
)

// JobfairRouter 招聘会路由（#2：企业举办 + 个人参加）
type JobfairRouter struct{}

// InitJobfairRouter 注册企业/个人招聘会路由
func (r *JobfairRouter) InitJobfairRouter(Router *gin.RouterGroup) {
	company := Router.Group("company").Use(middlewarehrc.MemberAuth()).Use(middlewarehrc.UtypeAuth(2))
	company.GET("jobfairs", hrcCompanyJobfairApi.List)
	company.POST("jobfairs", hrcCompanyJobfairApi.Create)
	company.PUT("jobfairs/:id", hrcCompanyJobfairApi.Update)
	company.DELETE("jobfairs/:id", hrcCompanyJobfairApi.Delete)

	personal := Router.Group("personal").Use(middlewarehrc.MemberAuth()).Use(middlewarehrc.UtypeAuth(1))
	personal.GET("jobfairs/signups", hrcPersonalJobfairApi.MySignupIDs)
	personal.GET("jobfairs/mine", hrcPersonalJobfairApi.Mine)
	personal.POST("jobfairs/:id/signup", hrcPersonalJobfairApi.Signup)
	personal.DELETE("jobfairs/:id/signup", hrcPersonalJobfairApi.Cancel)
}
