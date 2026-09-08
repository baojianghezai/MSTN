package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/flipped-aurora/gin-vue-admin/server/router"
	"github.com/gin-gonic/gin"
)

// 占位方法，保证文件可以正确加载，避免go空变量检测报错，请勿删除。
func holder(routers ...*gin.RouterGroup) {
	_ = routers
	_ = router.RouterGroupApp
}

// initBizRouter 名硕人力业务域（hrc）路由挂载
// 设计依据：02_用户账号模块设计 §4.1（M24 决议）+ 09_API接口规范 §1.5
//
// 路径约定（与 GVA 原生解耦，独立 /api/v1 前缀）：
//   - GVA 原生接口：保持 GVA 自身 RouterPrefix（当前为空 → /base/login 等），前端 GVA web 沿用
//   - hrc 会员/公开接口：/api/v1/**（含 auth/前台/personal/company/orders/pay 回调），会员 JWT 自管
//   - hrc 后台业务接口：/api/v1/admin/**，挂 GVA admin JWT+Casbin 中间件（与 PrivateGroup 同链）
func initBizRouter(Router *gin.Engine, routers ...*gin.RouterGroup) {
	privateGroup := routers[0]
	publicGroup := routers[1]

	hrcRouter := router.RouterGroupApp.Hrc

	// hrc 会员业务域（独立 /api/v1 前缀，会员 JWT 在路由内部自行挂载）
	v1Group := Router.Group("/api/v1")
	hrcRouter.InitAuthRouter(v1Group)
	hrcRouter.InitAppealRouter(v1Group)
	hrcRouter.InitProfileRouter(v1Group)
	hrcRouter.InitUploadRouter(v1Group)
	hrcRouter.InitCmsRouter(v1Group)
	hrcRouter.InitCategoryRouter(v1Group)
	hrcRouter.InitResumeRouter(v1Group)
	hrcRouter.InitVideoInterviewRouter(v1Group)
	hrcRouter.InitInterviewRouter(v1Group)
	hrcRouter.InitMessageRouter(v1Group)
	hrcRouter.InitTalentRouter(v1Group)
	hrcRouter.InitJobsRouter(v1Group)
	hrcRouter.InitCompanyPublicRouter(v1Group)
	hrcRouter.InitApplyRouter(v1Group)
	hrcRouter.InitSetmealRouter(v1Group)
	hrcRouter.InitOrderRouter(v1Group)
	hrcRouter.InitPaymentRouter(v1Group)
	hrcRouter.InitPromotionRouter(v1Group)

	// hrc 后台业务域（/api/v1/admin）
	// 一期先挂 GVA admin JWT + 改密守卫 + 数据权限（跳过 Casbin）：
	// hrc 后台权限点尚未在 sys_api 登记、也未建角色绑定，挂 Casbin 会直接 403；
	// 待 M6 建后台页面/角色时再按 02 §4.1 M24 补 Casbin 接入。
	adminBizGroup := v1Group.Group("admin")
	adminBizGroup.Use(middleware.JWTAuth()).Use(middleware.MustChangePwdGuard()).Use(middleware.DataScope())
	hrcRouter.InitAdminRouter(adminBizGroup)

	// TODO: 后续模块路由在此注册
	// hrcRouter.InitApplyRouter(v1Group)     // 投递/下载/面试
	// hrcRouter.InitOrderRouter(v1Group)     // 订单/支付/回调
	// hrcRouter.InitCmsRouter(v1Group)       // 内容/招聘会/消息

	holder(publicGroup, privateGroup)
}
