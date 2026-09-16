package hrc

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type CmsApi struct{}

// Articles 内容列表（资讯/招聘会/帮助，type: 1/2/3）
// @Tags HrcCms
// @Summary 内容列表
// @Produce json
// @Param type query int false "类型：1资讯 2招聘会 3帮助"
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} Response{data=PageData}
// @Router /api/v1/articles [get]
func (a *CmsApi) Articles(c *gin.Context) {
	typ, _ := strconv.Atoi(c.DefaultQuery("type", "0"))
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	list, total, err := hrcService.ServiceGroupApp.CmsService.ListArticles(c.Request.Context(), int8(typ), pageInfo)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, PageData{List: list, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}

// Article 内容详情（资讯/招聘会/帮助）
// @Tags HrcCms
// @Summary 内容详情
// @Produce json
// @Param id path int true "内容 id"
// @Success 200 {object} Response
// @Router /api/v1/articles/{id} [get]
func (a *CmsApi) Article(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	article, err := hrcService.ServiceGroupApp.CmsService.GetArticle(c.Request.Context(), id)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, article)
}

// Page 静态页（按 alias）
// @Tags HrcCms
// @Summary 静态页
// @Produce json
// @Param alias path string true "别名"
// @Success 200 {object} Response
// @Router /api/v1/pages/{alias} [get]
func (a *CmsApi) Page(c *gin.Context) {
	alias := c.Param("alias")
	page, err := hrcService.ServiceGroupApp.CmsService.GetPage(c.Request.Context(), alias)
	if err != nil {
		if err == hrcService.ErrPageNotFound {
			Fail(c, CodeParamError, err.Error())
			return
		}
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, page)
}

// Navigations 前台导航
// @Tags HrcCms
// @Summary 前台导航
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/navigations [get]
func (a *CmsApi) Navigations(c *gin.Context) {
	list, err := hrcService.ServiceGroupApp.CmsService.GetNavigations(c.Request.Context())
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, list)
}

// Categories 分类聚合（行业/地区/学历/经验/薪资等，按 group alias 分组）
// @Tags HrcCms
// @Summary 分类聚合
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/categories [get]
func (a *CmsApi) Categories(c *gin.Context) {
	data, err := hrcService.ServiceGroupApp.CmsService.GetCategories(c.Request.Context())
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, data)
}

// Districts 地区树单层查询
// @Tags HrcCms
// @Summary 按父级读取地区
// @Produce json
// @Param parentId query integer false "父级地区 ID，缺省为顶级地区"
// @Success 200 {object} Response "地区分类列表"
// @Router /api/v1/categories/districts [get]
func (a *CmsApi) Districts(c *gin.Context) {
	parentID := uint64(0)
	if raw := c.Query("parentId"); raw != "" {
		parsed, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			Fail(c, CodeParamError, "parentId 参数不合法")
			return
		}
		parentID = parsed
	}
	list, err := hrcService.ServiceGroupApp.CmsService.GetDistricts(c.Request.Context(), parentID)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, list)
}
