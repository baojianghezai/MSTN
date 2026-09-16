package hrc

import "github.com/gin-gonic/gin"

type CmsRouter struct{}

// InitCmsRouter 内容/配置公开路由（静态页/导航/分类）
// 对应 09 §4.1 编号 43/44、21
func (r *CmsRouter) InitCmsRouter(Router *gin.RouterGroup) {
	Router.GET("pages/:alias", hrcCmsApi.Page)
	Router.GET("navigations", hrcCmsApi.Navigations)
	Router.GET("categories", hrcCmsApi.Categories)
	Router.GET("categories/districts", hrcCmsApi.Districts)
	Router.GET("articles", hrcCmsApi.Articles)
	Router.GET("articles/:id", hrcCmsApi.Article)
}
