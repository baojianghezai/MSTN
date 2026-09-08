package hrc

import (
	"strconv"

	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type PersonalApplyApi struct{}

// Apply 投递职位（09 §4.2 #61）
// @Tags HrcApply
// @Summary 投递职位
// @Accept application/json
// @Produce json
// @Param data body ApplyRequest true "投递参数"
// @Success 200 {object} Response
// @Router /api/v1/personal/applies [post]
func (a *PersonalApplyApi) Apply(c *gin.Context) {
	var req ApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	if err := hrcService.ServiceGroupApp.ApplyService.Apply(c.Request.Context(), uid, req.JobsIDs, req.ResumeID, req.Notes); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// List 我的投递列表（09 §4.2 #62）
// @Tags HrcApply
// @Summary 我的投递列表
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param status query int false "状态：0全部 1未读 2已读"
// @Success 200 {object} Response{data=PageData}
// @Router /api/v1/personal/applies [get]
func (a *PersonalApplyApi) List(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))
	uid := middlewarehrc.GetMemberUID(c)
	list, total, err := hrcService.ServiceGroupApp.ApplyService.List(c.Request.Context(), uid, pageInfo, int8(status))
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, PageData{List: list, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}

// Delete 删除投递记录（09 §4.2 #63）
// @Tags HrcApply
// @Summary 删除投递记录
// @Produce json
// @Param did path int true "投递记录 id"
// @Success 200 {object} Response
// @Router /api/v1/personal/applies/{did} [delete]
func (a *PersonalApplyApi) Delete(c *gin.Context) {
	did, err := strconv.ParseUint(c.Param("did"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	if err := hrcService.ServiceGroupApp.ApplyService.Delete(c.Request.Context(), uid, did); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}
