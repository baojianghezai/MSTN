package hrc

import (
	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type CompanyCancelApi struct{}

// CompanyCancelApply 企业注销申请（短信二次确认，02 §2.8）
// @Tags HrcCompany
// @Summary 企业注销申请
// @Accept application/json
// @Produce json
// @Param data body CompanyCancelRequest true "注销参数"
// @Success 200 {object} Response
// @Router /api/v1/company/cancel [post]
func (a *CompanyCancelApi) CompanyCancelApply(c *gin.Context) {
	var req CompanyCancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	if err := hrcService.ServiceGroupApp.CompanyCancellationService.ApplyCompanyCancellation(c.Request.Context(), uid, req.Code); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// CompanyCancelStatus 查询企业注销申请状态
// @Tags HrcCompany
// @Summary 查询企业注销申请状态
// @Produce json
// @Success 200 {object} Response{data=CompanyCancellationStatusData}
// @Router /api/v1/company/cancel [get]
func (a *CompanyCancelApi) CompanyCancelStatus(c *gin.Context) {
	uid := middlewarehrc.GetMemberUID(c)
	apply, err := hrcService.ServiceGroupApp.CompanyCancellationService.GetCompanyCancellationStatus(c.Request.Context(), uid)
	if err != nil {
		if err == hrcService.ErrCompanyCancellationNotFound {
			// 无申请：返回成功 + 空 data，前端按空对象兜底（避免误导性报错）
			OKWithData(c, nil)
			return
		}
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, CompanyCancellationStatusData{
		ID:          apply.ID,
		CompanyName: apply.CompanyName,
		AddTime:     apply.AddTime.Unix(),
		Status:      apply.Status,
		StatusCN:    companyCancelStatusCN(apply.Status),
		FinishTime:  timeOrZeroUnix(apply.FinishTime),
	})
}

// companyCancelStatusCN 注销申请状态中文（0=待处理 1=已处理）
func companyCancelStatusCN(status int8) string {
	if status == 1 {
		return "已处理"
	}
	return "待处理"
}
