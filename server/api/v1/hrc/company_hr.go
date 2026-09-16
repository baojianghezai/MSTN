package hrc

import (
	"strconv"

	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

// CompanyHRApi 企业 HR 子账号管理（#21；仅企业主账号可操作）
type CompanyHRApi struct{}

type CompanyHRCreateRequest struct {
	Mobile   string `json:"mobile" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CompanyHRStatusRequest struct {
	Status int8 `json:"status" binding:"required"`
}

type CompanyHRPasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

// List 企业 HR 子账号列表
// @Tags HrcCompanyHR
// @Summary HR 子账号列表
// @Produce json
// @Success 200 {object} Response{data=[]hrcService.CompanyHRItem}
// @Router /api/v1/company/hrs [get]
func (a *CompanyHRApi) List(c *gin.Context) {
	if !requireCompanyOwner(c) {
		return
	}
	ownerUID := middlewarehrc.GetMemberUID(c)
	list, err := hrcService.ServiceGroupApp.CompanyHRService.List(c.Request.Context(), ownerUID)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, list)
}

// Create 新增 HR 子账号（手机号+密码登录，数据共享企业主体）
// @Tags HrcCompanyHR
// @Summary 新增 HR 子账号
// @Accept application/json
// @Produce json
// @Param data body CompanyHRCreateRequest true "手机号 + 密码"
// @Success 200 {object} Response{data=hrcService.CompanyHRItem}
// @Router /api/v1/company/hrs [post]
func (a *CompanyHRApi) Create(c *gin.Context) {
	if !requireCompanyOwner(c) {
		return
	}
	var req CompanyHRCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	ownerUID := middlewarehrc.GetMemberUID(c)
	item, err := hrcService.ServiceGroupApp.CompanyHRService.Create(c.Request.Context(), ownerUID, req.Mobile, req.Password)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, item)
}

// SetStatus 启用/禁用 HR 子账号
// @Tags HrcCompanyHR
// @Summary 启用/禁用 HR 子账号
// @Accept application/json
// @Produce json
// @Param uid path int true "HR uid"
// @Param data body CompanyHRStatusRequest true "1=启用 2=禁用"
// @Success 200 {object} Response
// @Router /api/v1/company/hrs/{uid}/status [put]
func (a *CompanyHRApi) SetStatus(c *gin.Context) {
	if !requireCompanyOwner(c) {
		return
	}
	hrUID, err := strconv.ParseUint(c.Param("uid"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	var req CompanyHRStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	ownerUID := middlewarehrc.GetMemberUID(c)
	if err := hrcService.ServiceGroupApp.CompanyHRService.SetStatus(c.Request.Context(), ownerUID, hrUID, req.Status); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// ResetPassword 重置 HR 子账号密码
// @Tags HrcCompanyHR
// @Summary 重置 HR 子账号密码
// @Accept application/json
// @Produce json
// @Param uid path int true "HR uid"
// @Param data body CompanyHRPasswordRequest true "新密码"
// @Success 200 {object} Response
// @Router /api/v1/company/hrs/{uid}/password [put]
func (a *CompanyHRApi) ResetPassword(c *gin.Context) {
	if !requireCompanyOwner(c) {
		return
	}
	hrUID, err := strconv.ParseUint(c.Param("uid"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	var req CompanyHRPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	ownerUID := middlewarehrc.GetMemberUID(c)
	if err := hrcService.ServiceGroupApp.CompanyHRService.ResetPassword(c.Request.Context(), ownerUID, hrUID, req.Password); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// Delete 移除 HR 子账号
// @Tags HrcCompanyHR
// @Summary 移除 HR 子账号
// @Produce json
// @Param uid path int true "HR uid"
// @Success 200 {object} Response
// @Router /api/v1/company/hrs/{uid} [delete]
func (a *CompanyHRApi) Delete(c *gin.Context) {
	if !requireCompanyOwner(c) {
		return
	}
	hrUID, err := strconv.ParseUint(c.Param("uid"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	ownerUID := middlewarehrc.GetMemberUID(c)
	if err := hrcService.ServiceGroupApp.CompanyHRService.Delete(c.Request.Context(), ownerUID, hrUID); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

func requireCompanyOwner(c *gin.Context) bool {
	if !middlewarehrc.IsMemberOwner(c) {
		Fail(c, CodeParamError, "仅企业主账号可管理 HR 子账号")
		return false
	}
	return true
}
