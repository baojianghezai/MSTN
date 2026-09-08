package hrc

import "github.com/gin-gonic/gin"

type AppealRouter struct{}

// InitAppealRouter 账号申诉路由（公开：09 §4.1 编号 45/46）
func (r *AppealRouter) InitAppealRouter(Router *gin.RouterGroup) {
	appealGroup := Router.Group("appeal")
	{
		appealGroup.POST("", hrcAppealApi.Submit)
		appealGroup.GET("status", hrcAppealApi.Status)
	}
}
