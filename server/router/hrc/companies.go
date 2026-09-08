package hrc

import "github.com/gin-gonic/gin"

type CompanyPublicRouter struct{}

func (r *CompanyPublicRouter) InitCompanyPublicRouter(Router *gin.RouterGroup) {
	Router.GET("companies", hrcCompanyPublicApi.List)
	Router.GET("companies/:id/jobs", hrcCompanyPublicApi.Jobs)
	Router.GET("companies/:id", hrcCompanyPublicApi.Detail)
}
