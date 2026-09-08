package hrc

import (
	"strconv"

	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type DashboardApi struct{}

// Dashboard 看板指标（今日/昨日/待办/收入，09 §4.6.2 #119）
// @Tags HrcDashboard
// @Summary 看板指标
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} Response{data=hrcService.DashboardData}
// @Router /api/v1/admin/dashboard [get]
func (a *DashboardApi) Dashboard(c *gin.Context) {
	data, err := hrcService.ServiceGroupApp.DashboardService.Dashboard(c.Request.Context())
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, data)
}

// DashboardTrend 趋势图数据（09 §4.6.2 #120）
// @Tags HrcDashboard
// @Summary 趋势图数据
// @Produce json
// @Param days query int false "天数（默认30，上限90）"
// @Param metric query string false "指标：register/resume/company/job/application"
// @Security ApiKeyAuth
// @Success 200 {object} Response{data=[]hrcService.TrendPoint}
// @Router /api/v1/admin/dashboard/trend [get]
func (a *DashboardApi) DashboardTrend(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	metric := c.DefaultQuery("metric", "register")
	data, err := hrcService.ServiceGroupApp.DashboardService.Trend(c.Request.Context(), days, metric)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, data)
}

// StatisticsResume 求职者分布（性别/学历/经验，08 §4 统计报表）
// @Tags HrcDashboard
// @Summary 求职者分布
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} Response{data=hrcService.ResumeDistribution}
// @Router /api/v1/admin/statistics/resume [get]
func (a *DashboardApi) StatisticsResume(c *gin.Context) {
	data, err := hrcService.ServiceGroupApp.DashboardService.ResumeDistribution(c.Request.Context())
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, data)
}

// StatisticsCompany 企业分布（性质/规模，08 §4 统计报表）
// @Tags HrcDashboard
// @Summary 企业分布
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} Response{data=hrcService.CompanyDistribution}
// @Router /api/v1/admin/statistics/company [get]
func (a *DashboardApi) StatisticsCompany(c *gin.Context) {
	data, err := hrcService.ServiceGroupApp.DashboardService.CompanyDistribution(c.Request.Context())
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, data)
}
