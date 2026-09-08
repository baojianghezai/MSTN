package hrc

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

// WxpayLogList 微信支付回调日志列表（后台，11 §四 P0#7）
// @Tags HrcAdmin
// @Summary 微信支付回调日志列表
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param status query int false "状态：缺省全部，0失败 1成功"
// @Success 200 {object} Response{data=PageData}
// @Router /api/v1/admin/wxpay-logs [get]
func (a *AdminApi) WxpayLogList(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	status := int8(-1)
	if v := c.Query("status"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 || n > 1 {
			Fail(c, CodeParamError, "状态非法")
			return
		}
		status = int8(n)
	}
	list, total, err := hrcService.ServiceGroupApp.WxpayLogService.List(c.Request.Context(), pageInfo, status)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, PageData{List: list, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}
