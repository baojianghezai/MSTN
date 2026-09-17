package middleware

import (
	"net/http"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/gin-gonic/gin"
)

// ClientKeyAuth 客户端密钥校验：开启后仅允许携带约定密钥的自家前端/网关调用。
//
// 设计说明：浏览器端密钥本质可被抓包获取，属于「提高门槛」的轻量防护；
// 真正硬边界应配合网关/Nginx 的 IP 白名单与 HTTPS（见部署说明）。
//
// 开关：system.client-auth-enable（默认关闭）；密钥：system.client-key。
// 放行：OPTIONS 预检、第三方支付回调（/api/v1/pay/notify/**）、密钥未配置时。
func ClientKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		expect := global.GVA_CONFIG.System.ClientKey
		if !global.GVA_CONFIG.System.ClientAuthEnable || expect == "" {
			c.Next()
			return
		}
		// 预检与第三方回调放行（回调方不会带我们的密钥）
		if c.Request.Method == http.MethodOptions || strings.HasPrefix(c.Request.URL.Path, "/api/v1/pay/notify") {
			c.Next()
			return
		}
		key := c.GetHeader("X-Client-Key")
		if key == "" {
			key = c.Query("clientKey") // WebSocket 无法带自定义 Header，走 query
		}
		if key != expect {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 1001, "message": "非法调用来源", "data": nil})
			return
		}
		c.Next()
	}
}
