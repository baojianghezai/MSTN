package hrc

import (
	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/gin-gonic/gin"
)

type PaymentRouter struct{}

func (r *PaymentRouter) InitPaymentRouter(router *gin.RouterGroup) {
	company := router.Group("").Use(middlewarehrc.MemberAuth()).Use(middlewarehrc.UtypeAuth(2))
	company.POST("orders/:id/pay", hrcPaymentApi.Start)

	notify := router.Group("pay/notify")
	notify.POST("wechat", hrcPaymentApi.WechatNotify)
	notify.POST("alipay", hrcPaymentApi.AlipayNotify)
}
