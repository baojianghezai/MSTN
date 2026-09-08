package hrc

import (
	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/gin-gonic/gin"
)

type InterviewRouter struct{}

// InitInterviewRouter 标准线下面试邀请路由。
func (r *InterviewRouter) InitInterviewRouter(Router *gin.RouterGroup) {
	companyGroup := Router.Group("company").Use(middlewarehrc.MemberAuth()).Use(middlewarehrc.UtypeAuth(2))
	{
		companyGroup.POST("interviews", hrcInterviewApi.Create)
		companyGroup.GET("interviews", hrcInterviewApi.CompanyList)
		companyGroup.DELETE("interviews/:did", hrcInterviewApi.Withdraw)
	}

	personalGroup := Router.Group("personal").Use(middlewarehrc.MemberAuth()).Use(middlewarehrc.UtypeAuth(1))
	{
		personalGroup.GET("interviews", hrcInterviewApi.PersonalList)
		personalGroup.PUT("interviews/:did/read", hrcInterviewApi.MarkRead)
	}
}
