package hrc

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type CompanyPublicApi struct{}

func publicPage(c *gin.Context) request.PageInfo {
	var page request.PageInfo
	_ = c.ShouldBindQuery(&page)
	if page.Page <= 0 {
		page.Page = 1
	}
	if page.PageSize <= 0 {
		page.PageSize = 10
	} else if page.PageSize > request.MaxPageSize {
		page.PageSize = request.MaxPageSize
	}
	return page
}

// List is the public company directory.
// @Tags HrcCompany
// @Summary 企业列表
// @Produce json
// @Param keyword query string false "企业名称或简介"
// @Param nature query int false "企业性质"
// @Param trade query int false "行业"
// @Param scale query int false "企业规模"
// @Param district query string false "地区"
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} Response{data=PageData}
// @Router /api/v1/companies [get]
func (a *CompanyPublicApi) List(c *gin.Context) {
	page := publicPage(c)
	filter := hrcService.CompanySearchFilter{
		Keyword:  c.Query("keyword"),
		District: c.Query("district"),
		Nature:   queryUint16(c, "nature"),
		Trade:    queryUint16(c, "trade"),
		Scale:    queryUint16(c, "scale"),
	}
	list, total, err := hrcService.ServiceGroupApp.CompanySearchService.Search(c.Request.Context(), filter, page)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, PageData{List: list, Total: total, Page: page.Page, PageSize: page.PageSize})
}

// Detail returns safe public company information and four active jobs.
// @Tags HrcCompany
// @Summary 企业详情
// @Produce json
// @Param id path int true "企业资料 ID"
// @Success 200 {object} Response{data=hrcService.PublicCompanyDetail}
// @Router /api/v1/companies/{id} [get]
func (a *CompanyPublicApi) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	detail, err := hrcService.ServiceGroupApp.CompanySearchService.Detail(c.Request.Context(), id)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, detail)
}

// Jobs returns the remaining active jobs for a public company.
// @Tags HrcCompany
// @Summary 企业在招职位
// @Produce json
// @Param id path int true "企业资料 ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} Response{data=PageData}
// @Router /api/v1/companies/{id}/jobs [get]
func (a *CompanyPublicApi) Jobs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	page := publicPage(c)
	jobs, total, err := hrcService.ServiceGroupApp.CompanySearchService.Jobs(c.Request.Context(), id, page)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, PageData{List: jobs, Total: total, Page: page.Page, PageSize: page.PageSize})
}
