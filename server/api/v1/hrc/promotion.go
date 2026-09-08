package hrc

import (
	"errors"
	"strconv"

	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type PromotionApi struct{}

type CreatePromotionRequest struct {
	JobID      uint64 `json:"jobId" binding:"required"`
	Type       int8   `json:"type" binding:"required"` // 1=推流 2=广告位
	AdTitle    string `json:"adTitle"`
	AdSubtitle string `json:"adSubtitle"`
	AdImage    string `json:"adImage"`
}

// PromotionResponse is kept as an API-owned schema so Swagger can resolve the
// promotion model without relying on an import alias in an annotation.
type PromotionResponse hrcModel.JobPromotion

// ListHome 首页有效推广（公开）
// @Tags HrcPromotion
// @Summary 首页推广位
// @Produce json
// @Success 200 {object} Response{data=hrcService.HomePromotionData}
// @Router /api/v1/home/promotions [get]
func (a *PromotionApi) ListHome(c *gin.Context) {
	data, err := hrcService.ServiceGroupApp.PromotionService.ListHome(c.Request.Context())
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, data)
}

// ListMine 企业当前投放及可用名额
// @Tags HrcPromotion
// @Summary 企业首页推广列表
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} Response{data=hrcService.CompanyPromotionData}
// @Router /api/v1/company/promotions [get]
func (a *PromotionApi) ListMine(c *gin.Context) {
	data, err := hrcService.ServiceGroupApp.PromotionService.ListMine(c.Request.Context(), middlewarehrc.GetMemberUID(c))
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, data)
}

// Create 企业投放首页推广
// @Tags HrcPromotion
// @Summary 创建首页推广
// @Accept application/json
// @Produce json
// @Param data body CreatePromotionRequest true "职位、推广类型与广告创意"
// @Security ApiKeyAuth
// @Success 200 {object} Response{data=PromotionResponse}
// @Router /api/v1/company/promotions [post]
func (a *PromotionApi) Create(c *gin.Context) {
	var req CreatePromotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	promotion, err := hrcService.ServiceGroupApp.PromotionService.Create(c.Request.Context(), middlewarehrc.GetMemberUID(c), req.JobID, req.Type, hrcService.PromotionCreative{
		AdTitle:    req.AdTitle,
		AdSubtitle: req.AdSubtitle,
		AdImage:    req.AdImage,
	})
	if err != nil {
		code := CodeParamError
		if errors.Is(err, hrcService.ErrPromotionEntitlement) || errors.Is(err, hrcService.ErrPromotionSlotLimit) {
			code = CodeSetmealLimit
		}
		Fail(c, code, err.Error())
		return
	}
	OKWithData(c, PromotionResponse(*promotion))
}

// Delete 企业撤下首页推广
// @Tags HrcPromotion
// @Summary 删除首页推广
// @Produce json
// @Param id path int true "推广记录 id"
// @Security ApiKeyAuth
// @Success 200 {object} Response
// @Router /api/v1/company/promotions/{id} [delete]
func (a *PromotionApi) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	if err := hrcService.ServiceGroupApp.PromotionService.Delete(c.Request.Context(), middlewarehrc.GetMemberUID(c), id); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}
