package hrc

import (
	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthApi struct{}

var authService = new(hrcService.AuthService)

// Captcha 获取图形验证码（一期 mock：返回占位，防刷在中间件层）
// @Tags HrcAuth
// @Summary 获取图形验证码
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/auth/captcha [post]
func (a *AuthApi) Captcha(c *gin.Context) {
	OKWithData(c, gin.H{"captchaId": "mock-id", "captchaImg": "data:image/png;base64,mock"})
}

// SmsCode 发送短信验证码
// @Tags HrcAuth
// @Summary 发送短信验证码
// @Accept application/json
// @Produce json
// @Param data body SmsCodeRequest true "发送参数"
// @Success 200 {object} Response
// @Router /api/v1/auth/sms-code [post]
func (a *AuthApi) SmsCode(c *gin.Context) {
	var req SmsCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	typ := req.Type
	if typ == "" {
		typ = "register"
	}
	if _, err := authService.SendSmsCodeFromIP(req.Mobile, typ, c.ClientIP()); err != nil {
		switch err {
		case hrcService.ErrMobileFormat, hrcService.ErrSmsTypeInvalid:
			Fail(c, CodeParamError, err.Error())
		case hrcService.ErrAccountNotFound:
			Fail(c, CodeAccountNotFound, err.Error())
		case hrcService.ErrSmsRateLimit:
			Fail(c, CodeSmsRateLimit, err.Error())
		default:
			Fail(c, CodeParamError, err.Error())
		}
		return
	}
	OKWithMessage(c, "验证码已发送")
}

// Register 会员注册（个人/企业）
// @Tags HrcAuth
// @Summary 会员注册（个人/企业）
// @Accept application/json
// @Produce json
// @Param data body RegisterRequest true "注册参数"
// @Success 200 {object} Response{data=AuthTokenData}
// @Router /api/v1/auth/register [post]
func (a *AuthApi) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	//检验个人用户还是企业用户
	if req.Utype != 1 && req.Utype != 2 {
		Fail(c, CodeParamError, "用户类型错误")
		return
	}
	member, err := authService.Register(req.Mobile, req.Code, req.Password, req.Utype)
	if err != nil {
		a.loginFailResponse(c, err)
		return
	}
	token, _ := middlewarehrc.DefaultMemberJWT().CreateToken(member.UID, member.Utype, 0)
	OKWithData(c, AuthTokenData{Token: token, UID: member.UID, Utype: member.Utype, Mobile: member.Mobile, PasswordSet: member.Password != ""})
}

// Login 密码登录（账号类型自动识别：mobile/email/username）
// @Tags HrcAuth
// @Summary 密码登录
// @Accept application/json
// @Produce json
// @Param data body LoginRequest true "登录参数"
// @Success 200 {object} Response{data=AuthTokenData}
// @Router /api/v1/auth/login [post]
func (a *AuthApi) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	member, err := authService.LoginByPassword(req.Account, req.Password)
	if err != nil {
		a.loginFailResponse(c, err)
		return
	}
	token, _ := middlewarehrc.DefaultMemberJWT().CreateToken(member.UID, member.Utype, 0)
	OKWithData(c, AuthTokenData{Token: token, UID: member.UID, Utype: member.Utype, Mobile: member.Mobile, PasswordSet: member.Password != ""})
}

// LoginSms 短信验证码登录（无账号自动注册）
// @Tags HrcAuth
// @Summary 短信验证码登录
// @Accept application/json
// @Produce json
// @Param data body LoginSmsRequest true "登录参数"
// @Success 200 {object} Response{data=AuthTokenData}
// @Router /api/v1/auth/login/sms [post]
func (a *AuthApi) LoginSms(c *gin.Context) {
	var req LoginSmsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	// utype 可选：1=个人 2=企业 0=未指定（已注册以库为准；未注册默认个人）
	if req.Utype != 0 && req.Utype != 1 && req.Utype != 2 {
		Fail(c, CodeParamError, "用户类型错误")
		return
	}
	member, err := authService.LoginBySms(req.Mobile, req.Code, req.Utype)
	if err != nil {
		a.loginFailResponse(c, err)
		return
	}
	token, _ := middlewarehrc.DefaultMemberJWT().CreateToken(member.UID, member.Utype, 0)
	OKWithData(c, AuthTokenData{Token: token, UID: member.UID, Utype: member.Utype, Mobile: member.Mobile, PasswordSet: member.Password != ""})
}

// LoginWechat 微信登录（一期 mock：返回 bind_token，待微信开放平台参数就绪后实现）
// @Tags HrcAuth
// @Summary 微信登录（公众号/小程序）
// @Accept application/json
// @Produce json
// @Param data body LoginWechatRequest true "微信参数"
// @Success 200 {object} Response
// @Router /api/v1/auth/login/wechat [post]
func (a *AuthApi) LoginWechat(c *gin.Context) {
	var req LoginWechatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	OKWithData(c, gin.H{"bindToken": "mock-bind-token"})
}

// WechatBind 未登录微信扫码绑定（创建新账号）
// @Tags HrcAuth
// @Summary 微信扫码绑定新账号
// @Accept application/json
// @Produce json
// @Param data body WechatBindRequest true "绑定参数"
// @Success 200 {object} Response{data=AuthTokenData}
// @Router /api/v1/auth/wechat/bind [post]
func (a *AuthApi) WechatBind(c *gin.Context) {
	OKWithData(c, AuthTokenData{Token: "mock-token", UID: 1, Utype: 1})
}

// WechatBindExist 已登录账号绑定微信
// @Tags HrcAuth
// @Summary 已登录账号绑定微信
// @Accept application/json
// @Produce json
// @Param data body object true "{\"code\":\"wx-code\"}"
// @Success 200 {object} Response
// @Router /api/v1/auth/wechat/bind-exist [post]
func (a *AuthApi) WechatBindExist(c *gin.Context) {
	OK(c)
}

// WechatMpBind 小程序 getPhoneNumber 绑定
// @Tags HrcAuth
// @Summary 小程序 getPhoneNumber 绑定
// @Accept application/json
// @Produce json
// @Param data body WechatMpBindRequest true "小程序绑定参数"
// @Success 200 {object} Response{data=AuthTokenData}
// @Router /api/v1/auth/wechat/mp-bind [post]
func (a *AuthApi) WechatMpBind(c *gin.Context) {
	OKWithData(c, AuthTokenData{Token: "mock-token", UID: 1, Utype: 1})
}

// Refresh 刷新 token
// @Tags HrcAuth
// @Summary 刷新 token
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/auth/refresh [post]
func (a *AuthApi) Refresh(c *gin.Context) {
	OKWithData(c, gin.H{"token": "mock-new-token"})
}

// Logout 退出登录（清除当前会话）
// @Tags HrcAuth
// @Summary 退出登录
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/auth/logout [post]
func (a *AuthApi) Logout(c *gin.Context) {
	uid := middlewarehrc.GetMemberUID(c)
	middlewarehrc.InvalidateSession(uid)
	OK(c)
}

// LogoutOther 强制其他端下线（单点登录下语义=清除会话；多端扩展时按设备维度踢）
// @Tags HrcAuth
// @Summary 强制其他端下线
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/auth/logout-other [post]
func (a *AuthApi) LogoutOther(c *gin.Context) {
	uid := middlewarehrc.GetMemberUID(c)
	middlewarehrc.InvalidateSession(uid)
	OK(c)
}

// Me 当前用户信息
// @Tags HrcAuth
// @Summary 当前用户信息
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/auth/me [get]
func (a *AuthApi) Me(c *gin.Context) {
	uid := middlewarehrc.GetMemberUID(c)
	utype := middlewarehrc.GetMemberUtype(c)
	OKWithData(c, gin.H{"uid": uid, "utype": utype})
}

// ChangePassword 修改密码
// @Tags HrcAuth
// @Summary 修改密码（原密码校验）
// @Accept application/json
// @Produce json
// @Param data body ChangePasswordRequest true "改密参数"
// @Success 200 {object} Response
// @Router /api/v1/auth/password [put]
func (a *AuthApi) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	if err := authService.ChangePassword(uid, req.OldPassword, req.NewPassword); err != nil {
		a.loginFailResponse(c, err)
		return
	}
	// 改密后踢出所有会话（旧 token 失效）
	middlewarehrc.InvalidateSession(uid)
	OK(c)
}

// ResetPassword 忘记密码重置
// @Tags HrcAuth
// @Summary 忘记密码重置
// @Accept application/json
// @Produce json
// @Param data body ResetPasswordRequest true "重置参数"
// @Success 200 {object} Response
// @Router /api/v1/auth/password/reset [post]
func (a *AuthApi) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	if err := authService.ResetPassword(req.Mobile, req.Code, req.NewPassword); err != nil {
		a.loginFailResponse(c, err)
		return
	}
	OK(c)
}

// BindMobile 绑定手机
// @Tags HrcAuth
// @Summary 绑定手机
// @Accept application/json
// @Produce json
// @Param data body BindMobileRequest true "绑定参数"
// @Success 200 {object} Response
// @Router /api/v1/auth/bind/mobile [put]
func (a *AuthApi) BindMobile(c *gin.Context) {
	var req BindMobileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	if err := authService.BindMobile(uid, req.Mobile, req.Code); err != nil {
		a.loginFailResponse(c, err)
		return
	}
	OK(c)
}

// UnbindMobile 解绑手机
// @Tags HrcAuth
// @Summary 解绑手机
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/auth/unbind/mobile [put]
func (a *AuthApi) UnbindMobile(c *gin.Context) {
	uid := middlewarehrc.GetMemberUID(c)
	if err := authService.UnbindMobile(uid); err != nil {
		a.loginFailResponse(c, err)
		return
	}
	OK(c)
}

// Cancel 账号注销
// @Tags HrcAuth
// @Summary 账号注销
// @Accept application/json
// @Produce json
// @Param data body CancelRequest true "注销参数"
// @Success 200 {object} Response
// @Router /api/v1/auth/cancel [post]
func (a *AuthApi) Cancel(c *gin.Context) {
	var req CancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	if err := authService.CancelAccount(uid, req.Code); err != nil {
		a.loginFailResponse(c, err)
		return
	}
	middlewarehrc.InvalidateSession(uid)
	OK(c)
}

// loginFailResponse 登录失败统一响应（错误码映射）
func (a *AuthApi) loginFailResponse(c *gin.Context, err error) {
	switch err {
	case hrcService.ErrMobileFormat:
		Fail(c, CodeParamError, err.Error())
	case hrcService.ErrAccountExists:
		Fail(c, CodeMobileRegistered, err.Error())
	case hrcService.ErrSmsCodeInvalid:
		Fail(c, CodeSmsCodeError, err.Error())
	case hrcService.ErrAccountNotFound:
		Fail(c, CodeAccountNotFound, err.Error())
	case hrcService.ErrPasswordWrong, hrcService.ErrPasswordNotSet:
		Fail(c, CodePasswordError, err.Error())
	case hrcService.ErrAccountDisabled, hrcService.ErrAccountCancelled:
		Fail(c, CodeAccountLocked, err.Error())
	case hrcService.ErrLoginLocked:
		Fail(c, CodeAccountLocked, err.Error())
	case hrcService.ErrOldPasswordWrong:
		Fail(c, CodePasswordError, "原密码错误")
	case hrcService.ErrSamePassword:
		Fail(c, CodePasswordError, err.Error())
	case hrcService.ErrPasswordTooShort, hrcService.ErrPasswordTooWeak:
		Fail(c, CodePasswordError, err.Error())
	case hrcService.ErrMobileBound:
		Fail(c, CodeMobileRegistered, err.Error())
	default:
		Fail(c, CodeParamError, err.Error())
	}
	zap.L().Warn("login failed", zap.Error(err))
}
