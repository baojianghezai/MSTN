package hrc

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type AdminJobsApi struct{}

// AdminListJobs 职位管理列表（后台，09 §4.6.2 #121）
// @Tags HrcAdminJobs
// @Summary 职位管理列表
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param audit query int false "审核状态：缺省全部，0草稿 1通过 2待审 3不通过"
// @Param keyword query string false "关键字（职位名/企业名）"
// @Success 200 {object} Response{data=PageData}
// @Router /api/v1/admin/jobs [get]
func (a *AdminJobsApi) AdminListJobs(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	audit := int8(-1)
	if v := c.Query("audit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 || n > 3 {
			Fail(c, CodeParamError, "审核状态非法")
			return
		}
		audit = int8(n)
	}
	keyword := c.Query("keyword")
	list, total, err := hrcService.ServiceGroupApp.JobsService.AdminListJobs(c.Request.Context(), pageInfo, audit, keyword)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, PageData{List: list, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}

// AdminListJobsTmp 待审职位列表（后台，09 §4.6.2 #122）
// @Tags HrcAdminJobs
// @Summary 待审职位列表
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} Response{data=PageData}
// @Router /api/v1/admin/jobs/tmp [get]
func (a *AdminJobsApi) AdminListJobsTmp(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	list, total, err := hrcService.ServiceGroupApp.JobsService.AdminListJobsTmp(c.Request.Context(), pageInfo)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, PageData{List: list, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}

// AdminGetJobDetail 职位审核详情（后台，含职位全量信息 + 联系方式/标签 + 关联企业资质）
// @Tags HrcAdminJobs
// @Summary 职位审核详情
// @Produce json
// @Param id path int true "职位 id（tmp 或 jobs）"
// @Success 200 {object} Response{data=AdminJobDetail}
// @Router /api/v1/admin/jobs/{id} [get]
func (a *AdminJobsApi) AdminGetJobDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	detail, err := hrcService.ServiceGroupApp.JobsService.AdminGetJobDetail(c.Request.Context(), id)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, detail)
}

// AuditJob 职位审核（后台，09 §4.6.2 #123）
// @Tags HrcAdminJobs
// @Summary 职位审核
// @Accept application/json
// @Produce json
// @Param id path int true "职位（tmp）id"
// @Param data body JobsAuditRequest true "审核参数"
// @Success 200 {object} Response
// @Router /api/v1/admin/jobs/{id}/audit [put]
func (a *AdminJobsApi) AuditJob(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	var req JobsAuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	if err := hrcService.ServiceGroupApp.JobsService.AuditJob(c.Request.Context(), id, req.Audit, req.Reason); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}
