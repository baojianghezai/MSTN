package hrc

import (
	"strconv"

	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

// CompanyJobfairApi 企业举办招聘会（#2；需套餐含「举办招聘会」权益）
type CompanyJobfairApi struct{}

type CompanyJobfairRequest struct {
	Title     string `json:"title" binding:"required"`
	Summary   string `json:"summary"`
	Cover     string `json:"cover"`
	Content   string `json:"content"`
	HoldTime  string `json:"holdTime" binding:"required"`
	Address   string `json:"address" binding:"required"`
	Organizer string `json:"organizer"`
}

func jobfairInputFromRequest(req CompanyJobfairRequest) hrcService.JobfairInput {
	return hrcService.JobfairInput{
		Title: req.Title, Summary: req.Summary, Cover: req.Cover, Content: req.Content,
		HoldTime: req.HoldTime, Address: req.Address, Organizer: req.Organizer,
	}
}

// List 我举办的招聘会
// @Tags HrcCompanyJobfair
// @Summary 我举办的招聘会
// @Produce json
// @Success 200 {object} Response{data=[]hrcModel.Article}
// @Router /api/v1/company/jobfairs [get]
func (a *CompanyJobfairApi) List(c *gin.Context) {
	uid := middlewarehrc.GetMemberUID(c)
	list, err := hrcService.ServiceGroupApp.JobfairService.ListMine(c.Request.Context(), uid)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, list)
}

// Create 举办招聘会（需套餐权益）
// @Tags HrcCompanyJobfair
// @Summary 举办招聘会
// @Accept application/json
// @Produce json
// @Param data body CompanyJobfairRequest true "招聘会信息"
// @Success 200 {object} Response{data=hrcModel.Article}
// @Router /api/v1/company/jobfairs [post]
func (a *CompanyJobfairApi) Create(c *gin.Context) {
	var req CompanyJobfairRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	article, err := hrcService.ServiceGroupApp.JobfairService.Create(c.Request.Context(), uid, jobfairInputFromRequest(req))
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, article)
}

// Update 编辑招聘会
// @Tags HrcCompanyJobfair
// @Summary 编辑招聘会
// @Accept application/json
// @Produce json
// @Param id path int true "招聘会 id"
// @Param data body CompanyJobfairRequest true "招聘会信息"
// @Success 200 {object} Response
// @Router /api/v1/company/jobfairs/{id} [put]
func (a *CompanyJobfairApi) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	var req CompanyJobfairRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	if err := hrcService.ServiceGroupApp.JobfairService.Update(c.Request.Context(), uid, id, jobfairInputFromRequest(req)); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// Delete 下架招聘会
// @Tags HrcCompanyJobfair
// @Summary 下架招聘会
// @Produce json
// @Param id path int true "招聘会 id"
// @Success 200 {object} Response
// @Router /api/v1/company/jobfairs/{id} [delete]
func (a *CompanyJobfairApi) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	if err := hrcService.ServiceGroupApp.JobfairService.Delete(c.Request.Context(), uid, id); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}
