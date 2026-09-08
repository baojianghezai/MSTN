package hrc

import (
	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/gin-gonic/gin"
)

type TalentRouter struct{}

func (r *TalentRouter) InitTalentRouter(Router *gin.RouterGroup) {
	Router.GET("resumes", hrcTalentApi.Search)
	Router.GET("resumes/talents", hrcTalentApi.TalentList)
	Router.GET("resumes/:id", hrcTalentApi.PublicDetail)

	companyGroup := Router.Group("company").Use(middlewarehrc.MemberAuth()).Use(middlewarehrc.UtypeAuth(2))
	{
		companyGroup.GET("talents", hrcTalentApi.ListUnlocked)
		companyGroup.GET("talents/:id", hrcTalentApi.GetUnlocked)
		companyGroup.POST("talents/:id/unlock", hrcTalentApi.Unlock)
		companyGroup.PUT("talents/:id/follow-up", hrcTalentApi.SetFollowUp)
		companyGroup.GET("favorites", hrcTalentApi.ListFavorites)
		companyGroup.POST("favorites", hrcTalentApi.Favorite)
		companyGroup.DELETE("favorites/:id", hrcTalentApi.Unfavorite)
	}
}
