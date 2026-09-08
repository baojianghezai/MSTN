package hrc

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
)

type DataCleanupRequest struct {
	Target        string `json:"target" binding:"required"`
	RetentionDays int    `json:"retentionDays"`
}

type DataCleanupExecuteRequest struct {
	DataCleanupRequest
	Confirmation string `json:"confirmation" binding:"required"`
}

type DataCleanupHistoryItem = hrc.DataCleanupLog

// DataCleanupPreview previews one fixed cleanup target without modifying data.
// @Tags HrcAdmin
// @Summary 预览后台数据清理范围
// @Accept application/json
// @Produce json
// @Param data body DataCleanupRequest true "清理目标"
// @Success 200 {object} Response{data=hrcService.DataCleanupPreview}
// @Security ApiKeyAuth
// @Router /api/v1/admin/data-cleanup/preview [post]
func (a *AdminApi) DataCleanupPreview(c *gin.Context) {
	var req DataCleanupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	preview, err := hrcService.ServiceGroupApp.DataCleanupService.Preview(c.Request.Context(), req.Target, req.RetentionDays)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, preview)
}

// DataCleanupExecute executes a confirmed fixed cleanup target and creates an audit log.
// @Tags HrcAdmin
// @Summary 执行后台数据清理
// @Accept application/json
// @Produce json
// @Param data body DataCleanupExecuteRequest true "清理确认参数"
// @Success 200 {object} Response{data=hrcService.DataCleanupPreview}
// @Security ApiKeyAuth
// @Router /api/v1/admin/data-cleanup/execute [post]
func (a *AdminApi) DataCleanupExecute(c *gin.Context) {
	var req DataCleanupExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	preview, err := hrcService.ServiceGroupApp.DataCleanupService.Execute(c.Request.Context(), req.Target, req.RetentionDays, req.Confirmation, hrcService.DataCleanupOperator{
		ID:   utils.GetUserID(c),
		Name: utils.GetUserName(c),
	})
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, preview)
}

// DataCleanupHistory lists completed data cleanup audit records.
// @Tags HrcAdmin
// @Summary 后台数据清理记录
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} Response{data=PageData{list=[]DataCleanupHistoryItem}}
// @Security ApiKeyAuth
// @Router /api/v1/admin/data-cleanup/history [get]
func (a *AdminApi) DataCleanupHistory(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	list, total, err := hrcService.ServiceGroupApp.DataCleanupService.ListHistory(c.Request.Context(), pageInfo)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, PageData{List: list, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}
