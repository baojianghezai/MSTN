package hrc

import (
	"net/http"
	"time"

	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type ExportApi struct{}

// ExportCompanies 企业导出（CSV 下载，11 §四 P0#5）
// @Tags HrcExport
// @Summary 企业导出（CSV）
// @Accept application/json
// @Produce text/csv
// @Param data body ExportRequest true "企业资料 id 列表"
// @Success 200 {file} file
// @Router /api/v1/admin/companies/export [post]
func (a *ExportApi) ExportCompanies(c *gin.Context) {
	var req ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	data, err := hrcService.ServiceGroupApp.ExportService.ExportCompanies(c.Request.Context(), req.IDS)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	filename := "company-export-" + time.Now().Format("20060102") + ".csv"
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

// ExportJobs 职位导出（CSV 下载，11 §四 P0#5 后半）
// @Tags HrcExport
// @Summary 职位导出（CSV）
// @Accept application/json
// @Produce text/csv
// @Param data body ExportRequest true "职位 id 列表"
// @Success 200 {file} file
// @Router /api/v1/admin/jobs/export [post]
func (a *ExportApi) ExportJobs(c *gin.Context) {
	var req ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	data, err := hrcService.ServiceGroupApp.ExportService.ExportJobs(c.Request.Context(), req.IDS)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	filename := "jobs-export-" + time.Now().Format("20060102") + ".csv"
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}
