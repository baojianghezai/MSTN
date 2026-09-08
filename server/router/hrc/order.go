package hrc

import (
	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/gin-gonic/gin"
)

type OrderRouter struct{}

func (r *OrderRouter) InitOrderRouter(router *gin.RouterGroup) {
	company := router.Group("").Use(middlewarehrc.MemberAuth()).Use(middlewarehrc.UtypeAuth(2))
	company.GET("company/members/setmeal", hrcSetmealApi.Current)
	company.POST("orders", hrcOrderApi.Create)
	company.GET("orders", hrcOrderApi.List)
	company.POST("orders/:id/cancel", hrcOrderApi.Cancel)
}
