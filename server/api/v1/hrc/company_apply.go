package hrc

import (
	"errors"
	"net/http"
	"strconv"

	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type CompanyApplyApi struct{}

// List 收简历列表（09 §4.3 #92）
// @Tags HrcCompanyApply
// @Summary 收简历列表
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param status query int false "已读状态：0全部 1未读 2已读"
// @Success 200 {object} Response{data=PageData}
// @Router /api/v1/company/applies [get]
func (a *CompanyApplyApi) List(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))
	uid := middlewarehrc.GetMemberUID(c)
	list, total, err := hrcService.ServiceGroupApp.CompanyApplyService.List(c.Request.Context(), uid, pageInfo, 0, int8(status))
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, PageData{List: list, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}

// JobApplies 单职位收到的投递（09 §4.3 #90）
// @Tags HrcCompanyApply
// @Summary 单职位收到的投递
// @Produce json
// @Param id path int true "职位 id"
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} Response{data=PageData}
// @Router /api/v1/company/jobs/{id}/applies [get]
func (a *CompanyApplyApi) JobApplies(c *gin.Context) {
	jobsID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	uid := middlewarehrc.GetMemberUID(c)
	list, total, err := hrcService.ServiceGroupApp.CompanyApplyService.List(c.Request.Context(), uid, pageInfo, jobsID, 0)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, PageData{List: list, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}

// Looked 标记已看（09 §4.3 #93）
// @Tags HrcCompanyApply
// @Summary 标记已看
// @Produce json
// @Param did path int true "投递记录 id"
// @Success 200 {object} Response
// @Router /api/v1/company/applies/{did}/looked [put]
func (a *CompanyApplyApi) Looked(c *gin.Context) {
	did, err := strconv.ParseUint(c.Param("did"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	if err := hrcService.ServiceGroupApp.CompanyApplyService.Looked(c.Request.Context(), uid, did); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// Reply 回复状态（09 §4.3 #94）
// @Tags HrcCompanyApply
// @Summary 回复状态
// @Accept application/json
// @Produce json
// @Param did path int true "投递记录 id"
// @Param data body CompanyApplyReplyRequest true "回复参数"
// @Success 200 {object} Response
// @Router /api/v1/company/applies/{did}/reply [put]
func (a *CompanyApplyApi) Reply(c *gin.Context) {
	did, err := strconv.ParseUint(c.Param("did"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	var req CompanyApplyReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	if err := hrcService.ServiceGroupApp.CompanyApplyService.Reply(c.Request.Context(), uid, did, req.IsReply); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// DownloadResume 下载完整简历（#95；首次下载扣套餐额度，重复下载不扣）
// @Tags HrcCompanyApply
// @Summary 下载完整简历
// @Produce text/html
// @Param did path int true "投递记录 id"
// @Success 200 {file} file
// @Router /api/v1/company/applies/{did}/resume/download [get]
func (a *CompanyApplyApi) DownloadResume(c *gin.Context) {
	did, err := strconv.ParseUint(c.Param("did"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	file, err := hrcService.ServiceGroupApp.CompanyApplyService.DownloadResume(c.Request.Context(), uid, did)
	if err != nil {
		code := CodeParamError
		if errors.Is(err, hrcService.ErrResumeDownloadEntitlement) || errors.Is(err, hrcService.ErrResumeDownloadLimit) {
			code = CodePointsNotEnough
		}
		if errors.Is(err, hrcService.ErrApplyNotFound) {
			code = CodeNoApplyRecord
		}
		Fail(c, code, err.Error())
		return
	}
	c.Header("Content-Disposition", `attachment; filename="`+file.Filename+`"`)
	c.Data(http.StatusOK, "text/html; charset=utf-8", file.Content)
}
