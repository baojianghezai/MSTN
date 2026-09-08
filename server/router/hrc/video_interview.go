package hrc

import (
	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/gin-gonic/gin"
)

type VideoInterviewRouter struct{}

// InitVideoInterviewRouter 视频面试路由（11 §四 P0#2）
// 企业端/个人端走会员 JWT；房间码查询公开（房间码即入房凭证）
func (r *VideoInterviewRouter) InitVideoInterviewRouter(Router *gin.RouterGroup) {
	companyGroup := Router.Group("company").Use(middlewarehrc.MemberAuth()).Use(middlewarehrc.UtypeAuth(2))
	{
		companyGroup.POST("video-interviews", hrcVideoInterviewApi.CreateInterview)
		companyGroup.GET("video-interviews", hrcVideoInterviewApi.CompanyList)
		companyGroup.GET("video-interviews/:id", hrcVideoInterviewApi.CompanyDetail)
		companyGroup.DELETE("video-interviews/:id", hrcVideoInterviewApi.CompanyDelete)
	}

	personalGroup := Router.Group("personal").Use(middlewarehrc.MemberAuth()).Use(middlewarehrc.UtypeAuth(1))
	{
		personalGroup.GET("video-interviews", hrcVideoInterviewApi.PersonalList)
		personalGroup.GET("video-interviews/:id", hrcVideoInterviewApi.PersonalDetail)
	}

	// 房间码查询（公开，TRTC 入房）
	Router.GET("video-interviews/room/:code", hrcVideoInterviewApi.RoomByCode)
}
