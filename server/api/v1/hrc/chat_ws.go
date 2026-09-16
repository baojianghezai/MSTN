package hrc

import (
	"net/http"
	"strings"
	"time"

	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ChatWsApi 在线对话 WebSocket（会员 token 走 query，浏览器 WS 无法带自定义 Header）
type ChatWsApi struct{}

var chatUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 会员 token 已校验；跨域交给反向代理/CORS 管控，这里放开 Origin
	CheckOrigin: func(r *http.Request) bool { return true },
}

const (
	chatPongWait   = 60 * time.Second
	chatPingPeriod = 30 * time.Second
)

// Connect 建立在线对话长连接
// @Tags HrcChat
// @Summary 在线对话 WebSocket
// @Produce json
// @Param token query string true "会员 JWT"
// @Router /api/v1/ws/chat [get]
func (a *ChatWsApi) Connect(c *gin.Context) {
	token := strings.TrimSpace(c.Query("token"))
	if token == "" {
		token = strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
	}
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 1001, "message": "未登录"})
		return
	}
	claims, err := middlewarehrc.DefaultMemberJWT().ParseToken(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 1001, "message": "登录已失效，请重新登录"})
		return
	}
	conn, err := chatUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	client := hrcService.NewChatClient(claims.UID, conn)
	hrcService.ChatHubInstance.Register(client)
	defer func() {
		hrcService.ChatHubInstance.Unregister(client)
		_ = conn.Close()
	}()

	_ = client.WriteJSON(gin.H{"type": "ready", "data": gin.H{"uid": claims.UID}})

	conn.SetReadLimit(1024)
	_ = conn.SetReadDeadline(time.Now().Add(chatPongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(chatPongWait))
	})

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(chatPingPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := client.WriteJSON(gin.H{"type": "ping"}); err != nil {
					return
				}
			case <-done:
				return
			}
		}
	}()

	// 读循环仅用于感知断开与处理客户端心跳；收到的业务消息一律走 REST 落库
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
	close(done)
}
