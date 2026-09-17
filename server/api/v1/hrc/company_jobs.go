package hrc

import (
	"errors"
	"strconv"
	"time"

	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type CompanyJobsApi struct{}

// CreateJob 发布职位（09 §4.3 #79）
// @Tags HrcCompanyJobs
// @Summary 发布职位
// @Accept application/json
// @Produce json
// @Param data body JobsRequest true "职位参数"
// @Success 200 {object} Response{data=map[string]uint64}
// @Router /api/v1/company/jobs [post]
func (a *CompanyJobsApi) CreateJob(c *gin.Context) {
	var req JobsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	job := jobsFromRequest(req)
	job.HrUID = middlewarehrc.GetMemberRealUID(c)
	id, err := hrcService.ServiceGroupApp.JobsService.CreateJob(c.Request.Context(), uid, job, jobsContactFromRequest(req), req.Tags)
	if err != nil {
		code := CodeParamError
		if errors.Is(err, hrcService.ErrSetmealJobLimit) {
			code = CodeSetmealLimit
		}
		Fail(c, code, err.Error())
		return
	}
	OKWithData(c, gin.H{"id": id})
}

// ListJobs 我的职位列表（09 §4.3 #80）
// @Tags HrcCompanyJobs
// @Summary 我的职位列表
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} Response{data=PageData}
// @Router /api/v1/company/jobs [get]
func (a *CompanyJobsApi) ListJobs(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	uid := middlewarehrc.GetMemberUID(c)
	list, total, err := hrcService.ServiceGroupApp.JobsService.ListJobs(c.Request.Context(), uid, pageInfo)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, PageData{List: list, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}

// GetJob 职位详情（编辑回显，09 §4.3 #81，含不通过原因）
// pending=1 表示按 jobs_tmp 表 id 查（列表项 pending=true 透传）
// @Tags HrcCompanyJobs
// @Summary 职位详情（编辑回显）
// @Produce json
// @Param id path int true "职位 id"
// @Param pending query bool false "是否查待审表（jobs_tmp）"
// @Success 200 {object} Response{data=hrcService.JobDetail}
// @Router /api/v1/company/jobs/{id} [get]
func (a *CompanyJobsApi) GetJob(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	pending := c.Query("pending") == "1"
	uid := middlewarehrc.GetMemberUID(c)
	detail, err := hrcService.ServiceGroupApp.JobsService.GetJob(c.Request.Context(), uid, id, pending)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, detail)
}

// UpdateJob 编辑职位（09 §4.3 #82，写 tmp 待审 / 直接更新）
// pending=1 表示编辑 jobs_tmp 行（列表项 pending=true 透传，tmp-only 职位：新发布未过审 / 被拒重提）；
// 缺省时 jobs 表未命中会自动回退 jobs_tmp（D1，08-21）
// @Tags HrcCompanyJobs
// @Summary 编辑职位
// @Accept application/json
// @Produce json
// @Param id path int true "职位 id"
// @Param pending query bool false "是否编辑待审表（jobs_tmp）行"
// @Param data body JobsRequest true "职位参数"
// @Success 200 {object} Response
// @Router /api/v1/company/jobs/{id} [put]
func (a *CompanyJobsApi) UpdateJob(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	pending := c.Query("pending") == "1"
	var req JobsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	job := jobsFromRequest(req)
	job.HrUID = middlewarehrc.GetMemberRealUID(c)
	if err := hrcService.ServiceGroupApp.JobsService.UpdateJob(c.Request.Context(), uid, id, pending, job, jobsContactFromRequest(req), req.Tags); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// DeleteJob 删除职位（逻辑删除，09 §4.3 #83）
// pending=1 表示删除 jobs_tmp 行（tmp-only 职位）；缺省时 jobs 表未命中会自动回退 jobs_tmp（D1，08-21）
// @Tags HrcCompanyJobs
// @Summary 删除职位
// @Produce json
// @Param id path int true "职位 id"
// @Param pending query bool false "是否删除待审表（jobs_tmp）行"
// @Success 200 {object} Response
// @Router /api/v1/company/jobs/{id} [delete]
func (a *CompanyJobsApi) DeleteJob(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	pending := c.Query("pending") == "1"
	uid := middlewarehrc.GetMemberUID(c)
	if err := hrcService.ServiceGroupApp.JobsService.DeleteJob(c.Request.Context(), uid, id, pending); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// PauseJob 暂停职位（09 §4.3 #84）
// @Tags HrcCompanyJobs
// @Summary 暂停职位
// @Produce json
// @Param id path int true "职位 id"
// @Success 200 {object} Response
// @Router /api/v1/company/jobs/{id}/pause [put]
func (a *CompanyJobsApi) PauseJob(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	if err := hrcService.ServiceGroupApp.JobsService.PauseJob(c.Request.Context(), uid, id); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// ResumeJob 恢复职位（09 §4.3 #85）
// @Tags HrcCompanyJobs
// @Summary 恢复职位
// @Produce json
// @Param id path int true "职位 id"
// @Success 200 {object} Response
// @Router /api/v1/company/jobs/{id}/resume [put]
func (a *CompanyJobsApi) ResumeJob(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	if err := hrcService.ServiceGroupApp.JobsService.ResumeJob(c.Request.Context(), uid, id); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// RefreshJob 刷新职位（09 §4.3 #86）
// @Tags HrcCompanyJobs
// @Summary 刷新职位
// @Produce json
// @Param id path int true "职位 id"
// @Success 200 {object} Response
// @Router /api/v1/company/jobs/{id}/refresh [put]
func (a *CompanyJobsApi) RefreshJob(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	if err := hrcService.ServiceGroupApp.JobsService.RefreshJob(c.Request.Context(), uid, id); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// jobsFromRequest 请求 DTO → 职位模型
func jobsFromRequest(req JobsRequest) *hrcModel.Jobs {
	return &hrcModel.Jobs{JobsBase: hrcModel.JobsBase{
		JobsName:   req.JobsName,
		Nature:     req.Nature,
		NatureCN:   req.NatureCN,
		Sex:        req.Sex,
		Amount:     req.Amount,
		TopClass:   req.TopClass,
		Category:   req.Category,
		SubClass:   req.SubClass,
		CategoryCN: req.CategoryCN,
		Trade:      req.Trade,
		District:   req.District,
		DistrictCN: req.DistrictCN,
		Tag:        req.Tag,
		Education:  req.Education,
		Experience: req.Experience,
		MinWage:    req.MinWage,
		MaxWage:    req.MaxWage,
		Negotiable: req.Negotiable,
		Contents:   req.Contents,
		Deadline:   int64ToTime(req.Deadline),
		Department: req.Department,
		MapX:       req.MapX,
		MapY:       req.MapY,
		MapZoom:    req.MapZoom,
	}}
}

// jobsContactFromRequest 请求 DTO → 联系方式模型
func jobsContactFromRequest(req JobsRequest) *hrcModel.JobsContact {
	return &hrcModel.JobsContact{
		Contact:     req.Contact.Contact,
		QQ:          req.Contact.QQ,
		Telephone:   req.Contact.Telephone,
		LandlineTel: req.Contact.LandlineTel,
		Address:     req.Contact.Address,
		Email:       req.Contact.Email,
	}
}

func int64ToTime(ts int64) time.Time {
	if ts <= 0 {
		return time.Time{}
	}
	return time.Unix(ts, 0)
}

func int64ToTimePtr(ts int64) *time.Time {
	if ts <= 0 {
		return nil
	}
	t := time.Unix(ts, 0)
	return &t
}
