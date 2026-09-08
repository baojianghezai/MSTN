package hrc

import (
	"strconv"

	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type CmsApi struct{}

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
