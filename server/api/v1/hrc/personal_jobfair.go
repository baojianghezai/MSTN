package hrc

import (
	"strconv"

	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

// PersonalJobfairApi 个人参加招聘会（#2）
type PersonalJobfairApi struct{}

func parseJobfairID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		Fail(c, CodeParamError, "参数错误")
		return 0, false
	}
	return id, true
}

// MySignupIDs 我已报名的招聘会 id 列表
// @Tags HrcPersonalJobfair
// @Summary 我报名的招聘会 id 列表
// @Produce json
// @Success 200 {object} Response{data=[]uint64}
// @Router /api/v1/personal/jobfairs/signups [get]
func (a *PersonalJobfairApi) MySignupIDs(c *gin.Context) {
	uid := middlewarehrc.GetMemberUID(c)
	ids, err := hrcService.ServiceGroupApp.JobfairService.MySignupIDs(c.Request.Context(), uid)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, ids)
}

// Mine 我参加的招聘会（分页）
// @Tags HrcPersonalJobfair
// @Summary 我参加的招聘会
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} Response{data=PageData}
// @Router /api/v1/personal/jobfairs/mine [get]
func (a *PersonalJobfairApi) Mine(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	uid := middlewarehrc.GetMemberUID(c)
	list, total, err := hrcService.ServiceGroupApp.JobfairService.MyJobfairs(c.Request.Context(), uid, pageInfo)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, PageData{List: list, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}

// Signup 报名参加招聘会
// @Tags HrcPersonalJobfair
// @Summary 报名参加招聘会
// @Produce json
// @Param id path int true "招聘会 id"
// @Success 200 {object} Response
// @Router /api/v1/personal/jobfairs/{id}/signup [post]
func (a *PersonalJobfairApi) Signup(c *gin.Context) {
	id, ok := parseJobfairID(c)
	if !ok {
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	if err := hrcService.ServiceGroupApp.JobfairService.Signup(c.Request.Context(), uid, id); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// Cancel 取消报名
// @Tags HrcPersonalJobfair
// @Summary 取消报名
// @Produce json
// @Param id path int true "招聘会 id"
// @Success 200 {object} Response
// @Router /api/v1/personal/jobfairs/{id}/signup [delete]
func (a *PersonalJobfairApi) Cancel(c *gin.Context) {
	id, ok := parseJobfairID(c)
	if !ok {
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	if err := hrcService.ServiceGroupApp.JobfairService.Cancel(c.Request.Context(), uid, id); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}
