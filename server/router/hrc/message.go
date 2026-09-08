package hrc

import (
	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/gin-gonic/gin"
)

type MessageRouter struct{}

// InitMessageRouter registers the shared personal and company message-center routes.
func (r *MessageRouter) InitMessageRouter(Router *gin.RouterGroup) {
	personalGroup := Router.Group("personal").Use(middlewarehrc.MemberAuth()).Use(middlewarehrc.UtypeAuth(1))
	companyGroup := Router.Group("company").Use(middlewarehrc.MemberAuth()).Use(middlewarehrc.UtypeAuth(2))
	registerMessageRoutes(personalGroup)
	registerMessageRoutes(companyGroup)
}

func registerMessageRoutes(group gin.IRoutes) {
	group.GET("messages", hrcMessageApi.List)
	group.PUT("messages/read", hrcMessageApi.MarkRead)
	group.DELETE("messages", hrcMessageApi.Delete)
	group.GET("messages/unread-count", hrcMessageApi.UnreadCount)
}
