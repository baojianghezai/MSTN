package hrc

import (
	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/gin-gonic/gin"
)

type PromotionRouter struct{}

func (r *PromotionRouter) InitPromotionRouter(router *gin.RouterGroup) {
	router.GET("home/promotions", hrcPromotionApi.ListHome)
	company := router.Group("company/promotions").Use(middlewarehrc.MemberAuth()).Use(middlewarehrc.UtypeAuth(2))
	{
		company.GET("", hrcPromotionApi.ListMine)
		company.POST("", hrcPromotionApi.Create)
		company.DELETE(":id", hrcPromotionApi.Delete)
	}
}
