package hrc

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type AdminApi struct{}

// AppealList 申诉列表（后台客服）
// @Tags HrcAdmin
// @Summary 申诉列表
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param status query int false "状态：0全部 1已处理 2已驳回"
// @Param mobile query string false "手机号"
// @Success 200 {object} Response{data=PageData}
// @Router /api/v1/admin/appeals [get]
func (a *AdminApi) AppealList(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))
	mobile := c.Query("mobile")

	list, total, err := hrcService.ServiceGroupApp.AdminService.AppealList(c.Request.Context(), pageInfo, int8(status), mobile)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	items := make([]AppealStatusItem, 0, len(list))
	for _, v := range list {
		items = append(items, AppealStatusItem{
			ID:          v.ID,
			UID:         v.UID,
			RealName:    v.RealName,
			Mobile:      v.Mobile,
			Email:       v.Email,
			Description: v.Description,
			AddTime:     v.AddTime,
			Status:      v.Status,
			StatusCN:    appealStatusCN(v.Status),
		})
	}
	OKWithData(c, PageData{List: items, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}

// ConfigList 系统配置（分组读取）
// @Tags HrcAdmin
// @Summary 系统配置读取
// @Produce json
// @Param group query string false "分组：site/security/sms/payment"
// @Success 200 {object} Response
// @Router /api/v1/admin/configs [get]
func (a *AdminApi) ConfigList(c *gin.Context) {
	group := c.Query("group")
	list, err := hrcService.ServiceGroupApp.CmsService.GetConfigs(c.Request.Context(), group)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, list)
}

// ConfigSave 系统配置保存（按 name 幂等 upsert）
// @Tags HrcAdmin
// @Summary 系统配置保存
// @Accept application/json
// @Produce json
// @Param data body SaveConfigsRequest true "配置项"
// @Success 200 {object} Response
// @Router /api/v1/admin/configs [put]
func (a *AdminApi) ConfigSave(c *gin.Context) {
	var req SaveConfigsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	items := make([]hrcService.ConfigKV, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, hrcService.ConfigKV{Name: it.Name, Value: it.Value, Remark: it.Remark})
	}
	if err := hrcService.ServiceGroupApp.CmsService.SaveConfigs(c.Request.Context(), req.Group, items); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// ProcessAppeal 处理申诉（可恢复账号）
// @Tags HrcAdmin
// @Summary 处理申诉
// @Accept application/json
// @Produce json
// @Param id path int true "申诉ID"
// @Param data body AppealProcessRequest true "处理参数"
// @Success 200 {object} Response
// @Router /api/v1/admin/appeals/{id} [put]
func (a *AdminApi) ProcessAppeal(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	var req AppealProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	if err := hrcService.ServiceGroupApp.AdminService.ProcessAppeal(c.Request.Context(), id, req.Status, req.Restore); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// CompanyCancellationList 企业注销申请列表（后台，161a）
// @Tags HrcAdmin
// @Summary 企业注销申请列表
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param status query int false "状态：0全部 1待处理 2已处理"
// @Success 200 {object} Response{data=PageData}
// @Router /api/v1/admin/company/cancellations [get]
func (a *AdminApi) CompanyCancellationList(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))

	list, total, err := hrcService.ServiceGroupApp.CompanyCancellationService.CompanyCancellationList(c.Request.Context(), pageInfo, int8(status))
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	items := make([]CompanyCancellationAdminItem, 0, len(list))
	for _, v := range list {
		items = append(items, CompanyCancellationAdminItem{
			ID:          v.ID,
			UID:         v.UID,
			CompanyID:   v.CompanyID,
			CompanyName: v.CompanyName,
			Username:    v.Username,
			Mobile:      v.Mobile,
			AddTime:     v.AddTime,
			Status:      v.Status,
			StatusCN:    companyCancelStatusCN(v.Status),
			FinishTime:  v.FinishTime,
		})
	}
	OKWithData(c, PageData{List: items, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}

// CompanyCancellationHandle 处理企业注销（清企业业务数据，161b）
// @Tags HrcAdmin
// @Summary 处理企业注销
// @Produce json
// @Param id path int true "申请ID"
// @Success 200 {object} Response
// @Router /api/v1/admin/company/cancellations/{id}/handle [post]
func (a *AdminApi) CompanyCancellationHandle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	if err := hrcService.ServiceGroupApp.CompanyCancellationService.HandleCompanyCancellation(c.Request.Context(), id); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// CompanyCancellationDelete 硬删除注销申请记录（161c）
// @Tags HrcAdmin
// @Summary 删除注销申请记录
// @Produce json
// @Param id path int true "申请ID"
// @Success 200 {object} Response
// @Router /api/v1/admin/company/cancellations/{id} [delete]
func (a *AdminApi) CompanyCancellationDelete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	if err := hrcService.ServiceGroupApp.CompanyCancellationService.DeleteCompanyCancellation(c.Request.Context(), id); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// CompanyProfileList 企业资料列表（后台审核，134）
// @Tags HrcAdmin
// @Summary 企业资料列表
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param audit query int false "审核状态：缺省全部，0草稿 1通过 2待审 3不通过"
// @Param keyword query string false "企业名关键字"
// @Success 200 {object} Response{data=PageData}
// @Router /api/v1/admin/company-profiles [get]
func (a *AdminApi) CompanyProfileList(c *gin.Context) {
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

	list, total, err := hrcService.ServiceGroupApp.CompanyAuditService.CompanyList(c.Request.Context(), pageInfo, audit, keyword)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	items := make([]CompanyProfileAdminItem, 0, len(list))
	for _, p := range list {
		items = append(items, CompanyProfileAdminItem{
			ID:          p.ID,
			UID:         p.UID,
			CompanyName: companyNameStr(p.CompanyName),
			Logo:        p.Logo,
			Audit:       p.Audit,
			AuditCN:     companyAuditCN(p.Audit),
			AddTime:     p.AddTime,
			Refreshtime: p.Refreshtime,
		})
	}
	OKWithData(c, PageData{List: items, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}

// CompanyProfileDetail 企业资料详情（后台审核，134a）
// @Tags HrcAdmin
// @Summary 企业资料详情
// @Produce json
// @Param id path int true "企业资料ID"
// @Success 200 {object} Response{data=CompanyProfileAdminDetail}
// @Router /api/v1/admin/company-profiles/{id} [get]
func (a *AdminApi) CompanyProfileDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	p, err := hrcService.ServiceGroupApp.CompanyAuditService.CompanyDetail(c.Request.Context(), id)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, CompanyProfileAdminDetail{CompanyProfile: *p, AuditCN: companyAuditCN(p.Audit)})
}

// CompanyProfileAudit 企业资质审核（135）
// @Tags HrcAdmin
// @Summary 企业资质审核
// @Accept application/json
// @Produce json
// @Param id path int true "企业资料ID"
// @Param data body CompanyProfileAuditRequest true "审核参数"
// @Success 200 {object} Response
// @Router /api/v1/admin/company-profiles/{id}/audit [put]
func (a *AdminApi) CompanyProfileAudit(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	var req CompanyProfileAuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	if err := hrcService.ServiceGroupApp.CompanyAuditService.CompanyAudit(c.Request.Context(), id, req.Audit, req.Reason); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// companyNameStr *string → string（空指针转空串，前端友好）
func companyNameStr(name *string) string {
	if name == nil {
		return ""
	}
	return *name
}
