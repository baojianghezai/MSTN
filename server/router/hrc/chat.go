package hrc

import (
	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/gin-gonic/gin"
)

// ChatRouter 在线对话路由（个人/企业双端共用；WebSocket 走 query token 校验）
type ChatRouter struct{}

// InitChatRouter 注册 REST + WebSocket 路由
func (r *ChatRouter) InitChatRouter(Router *gin.RouterGroup) {
	personalGroup := Router.Group("personal").Use(middlewarehrc.MemberAuth()).Use(middlewarehrc.UtypeAuth(1))
	companyGroup := Router.Group("company").Use(middlewarehrc.MemberAuth()).Use(middlewarehrc.UtypeAuth(2))
	registerChatRoutes(personalGroup)
	registerChatRoutes(companyGroup)

	// WebSocket：不能挂 MemberAuth（浏览器 WS 无法自定义 Header），token 由 handler 从 query 校验
	Router.GET("ws/chat", hrcChatWsApi.Connect)
}

func registerChatRoutes(group gin.IRoutes) {
	group.GET("chat/sessions", hrcChatApi.ListSessions)
	group.POST("chat/sessions", hrcChatApi.OpenSession)
	group.GET("chat/sessions/:id/messages", hrcChatApi.Messages)
	group.POST("chat/sessions/:id/messages", hrcChatApi.SendMessage)
	group.PUT("chat/sessions/:id/read", hrcChatApi.MarkRead)
	group.DELETE("chat/sessions/:id", hrcChatApi.DeleteSession)
	group.GET("chat/unread", hrcChatApi.Unread)
}
