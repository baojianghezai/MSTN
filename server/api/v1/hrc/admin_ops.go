package hrc

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

// AdminOpsApi 后台运维接口（待办统计 + 推广投放审核）
type AdminOpsApi struct{}

// PendingCounts 后台待处理数量
// @Tags HrcAdminOps
// @Summary 后台待处理数量
// @Produce json
// @Success 200 {object} Response{data=hrcService.PendingCounts}
// @Router /api/v1/admin/pending-counts [get]
func (a *AdminOpsApi) PendingCounts(c *gin.Context) {
	counts, err := hrcService.ServiceGroupApp.PendingService.Counts(c.Request.Context())
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, counts)
}

// PromotionList 推广投放列表（audit=-1 全部）
// @Tags HrcAdminOps
// @Summary 推广投放列表
// @Produce json
// @Param audit query int false "审核状态：-1全部 0待审 1通过 3不通过"
// @Success 200 {object} Response{data=PageData}
// @Router /api/v1/admin/promotions [get]
func (a *AdminOpsApi) PromotionList(c *gin.Context) {
	audit, _ := strconv.Atoi(c.DefaultQuery("audit", "-1"))
	var page request.PageInfo
	_ = c.ShouldBindQuery(&page)
	list, total, err := hrcService.ServiceGroupApp.PromotionService.AdminList(c.Request.Context(), int8(audit), page)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, PageData{List: list, Total: total, Page: page.Page, PageSize: page.PageSize})
}

// AdminPromotionAuditRequest 推广审核请求
type AdminPromotionAuditRequest struct {
	Audit  int8   `json:"audit" binding:"required"` // 1=通过 3=不通过
	Reason string `json:"reason"`
}

// PromotionAudit 审核推广投放（通过后才在首页展示）
// @Tags HrcAdminOps
// @Summary 审核推广投放
// @Accept application/json
// @Produce json
// @Param id path int true "推广 id"
// @Param data body AdminPromotionAuditRequest true "审核结果"
// @Success 200 {object} Response
// @Router /api/v1/admin/promotions/{id}/audit [put]
func (a *AdminOpsApi) PromotionAudit(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	var req AdminPromotionAuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	if req.Audit == 3 && req.Reason == "" {
		Fail(c, CodeParamError, "不通过时必须填写原因")
		return
	}
	if err := hrcService.ServiceGroupApp.PromotionService.AuditPromotion(c.Request.Context(), id, req.Audit, req.Reason); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}
