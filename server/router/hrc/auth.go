package hrc

import (
	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/gin-gonic/gin"
)

type AuthRouter struct{}

// InitAuthRouter 认证路由（公开 + 需会员鉴权）
// 对应 09_API接口规范 §4.1（1-20）：
// 公开：captcha/sms-code/register/login/login-sms/login-wechat/wechat-bind/wechat-bind-exist/wechat-mp-bind/refresh/password-reset
// 鉴权：logout/logout-other/me/password/bind-mobile/unbind-mobile/cancel
func (r *AuthRouter) InitAuthRouter(Router *gin.RouterGroup) {
	authGroup := Router.Group("auth")

	// 公开接口
	{
		authGroup.POST("captcha", hrcAuthApi.Captcha)
		authGroup.POST("sms-code", hrcAuthApi.SmsCode)
		authGroup.POST("register", hrcAuthApi.Register)
		authGroup.POST("login", hrcAuthApi.Login)
		authGroup.POST("login/sms", hrcAuthApi.LoginSms)
		authGroup.POST("login/wechat", hrcAuthApi.LoginWechat)
		authGroup.POST("wechat/bind", hrcAuthApi.WechatBind)
		authGroup.POST("wechat/bind-exist", hrcAuthApi.WechatBindExist)
		authGroup.POST("wechat/mp-bind", hrcAuthApi.WechatMpBind)
		authGroup.POST("refresh", hrcAuthApi.Refresh)
	}

	// 需会员鉴权
	authMember := authGroup.Group("").Use(middlewarehrc.MemberAuth())
	{
		authMember.POST("logout", hrcAuthApi.Logout)
		authMember.POST("logout-other", hrcAuthApi.LogoutOther)
		authMember.GET("me", hrcAuthApi.Me)
		authMember.PUT("password", hrcAuthApi.ChangePassword)
		authMember.PUT("bind/mobile", hrcAuthApi.BindMobile)
		authMember.PUT("unbind/mobile", hrcAuthApi.UnbindMobile)
		authMember.POST("cancel", hrcAuthApi.Cancel)
	}

	// 公开但需验证码的重置（注销恢复已迁至后台 /api/v1/admin，见 02 §2.5 + #158/#159）
	{
		authGroup.POST("password/reset", hrcAuthApi.ResetPassword)
	}
}
