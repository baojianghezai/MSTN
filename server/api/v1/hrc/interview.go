package hrc

import (
	"strconv"

	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type InterviewApi struct{}

type InterviewCreateRequest struct {
	ResumeID      uint64 `json:"resumeId" binding:"required"`
	JobsID        uint64 `json:"jobsId" binding:"required"`
	InterviewTime int64  `json:"interviewTime" binding:"required"`
	Address       string `json:"address" binding:"required"`
	Contact       string `json:"contact" binding:"required"`
	Telephone     string `json:"telephone" binding:"required"`
	Notes         string `json:"notes"`
}

type CompanyInterviewPageData struct {
	List     []hrcModel.CompanyInterview `json:"list"`
	Total    int64                       `json:"total"`
	Page     int                         `json:"page"`
	PageSize int                         `json:"pageSize"`
}

// Create 企业创建标准线下面试邀请。
// @Tags HrcInterview
// @Summary 发起线下面试邀请
// @Accept application/json
// @Produce json
// @Param data body InterviewCreateRequest true "邀请参数"
// @Success 200 {object} Response{data=hrcModel.CompanyInterview}
// @Security ApiKeyAuth
// @Router /api/v1/company/interviews [post]
func (a *InterviewApi) Create(c *gin.Context) {
	var req InterviewCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	interview, err := hrcService.ServiceGroupApp.InterviewService.Create(c.Request.Context(), middlewarehrc.GetMemberUID(c), hrcService.InterviewCreateInput{
		ResumeID:      req.ResumeID,
		JobsID:        req.JobsID,
		InterviewTime: int64ToTime(req.InterviewTime),
		Address:       req.Address,
		Contact:       req.Contact,
		Telephone:     req.Telephone,
		Notes:         req.Notes,
	})
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, interview)
}

// CompanyList 企业查看已发出的线下面试邀请。
// @Tags HrcInterview
// @Summary 我发出的线下面试邀请
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} Response{data=CompanyInterviewPageData}
// @Security ApiKeyAuth
// @Router /api/v1/company/interviews [get]
func (a *InterviewApi) CompanyList(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	list, total, err := hrcService.ServiceGroupApp.InterviewService.CompanyList(c.Request.Context(), middlewarehrc.GetMemberUID(c), pageInfo)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, CompanyInterviewPageData{List: list, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}

// Withdraw 企业撤回线下面试邀请。
// @Tags HrcInterview
// @Summary 撤回线下面试邀请
// @Produce json
// @Param did path int true "邀请 id"
// @Success 200 {object} Response
// @Security ApiKeyAuth
// @Router /api/v1/company/interviews/{did} [delete]
func (a *InterviewApi) Withdraw(c *gin.Context) {
	did, err := strconv.ParseUint(c.Param("did"), 10, 64)
	if err != nil || did == 0 {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	if err := hrcService.ServiceGroupApp.InterviewService.Withdraw(c.Request.Context(), middlewarehrc.GetMemberUID(c), did); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// PersonalList 个人查看收到的线下面试邀请。
// @Tags HrcInterview
// @Summary 我的线下面试邀请
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} Response{data=CompanyInterviewPageData}
// @Security ApiKeyAuth
// @Router /api/v1/personal/interviews [get]
func (a *InterviewApi) PersonalList(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	list, total, err := hrcService.ServiceGroupApp.InterviewService.PersonalList(c.Request.Context(), middlewarehrc.GetMemberUID(c), pageInfo)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, CompanyInterviewPageData{List: list, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}

// MarkRead 个人将线下面试邀请标记为已读。
// @Tags HrcInterview
// @Summary 标记线下面试邀请已读
// @Produce json
// @Param did path int true "邀请 id"
// @Success 200 {object} Response
// @Security ApiKeyAuth
// @Router /api/v1/personal/interviews/{did}/read [put]
func (a *InterviewApi) MarkRead(c *gin.Context) {
	did, err := strconv.ParseUint(c.Param("did"), 10, 64)
	if err != nil || did == 0 {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	if err := hrcService.ServiceGroupApp.InterviewService.MarkRead(c.Request.Context(), middlewarehrc.GetMemberUID(c), did); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}
