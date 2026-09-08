package hrc

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 业务错误码区间（对应设计文档 09_API接口规范 §1.3）
const (
	CodeSuccess = 0

	// 1xxx 通用
	CodeParamError = 1001

	// 2xxx 用户/账号
	CodeAccountNotFound  = 2001
	CodePasswordError    = 2002
	CodeCaptchaError     = 2003
	CodeAccountLocked    = 2004
	CodeMobileRegistered = 2005
	CodeSmsCodeError     = 2006
	CodeSmsRateLimit     = 2007

	// 3xxx 职位/简历
	CodeSetmealLimit    = 3001
	CodeAuditNotPassed  = 3002
	CodeAlreadyApplied  = 3003
	CodeCompanyNotAudit = 3004
	CodeJobsExpired     = 3005

	// 4xxx 订单/支付
	CodeOrderNotFound   = 4001
	CodeOrderPaid       = 4002
	CodeOrderDuplicated = 4003
	CodeNotifySignFail  = 4004
	CodeOrderClosed     = 4005

	// 5xxx 业务
	CodeAlreadyDownloaded = 5001
	CodeDailyLimit        = 5002
	CodePointsNotEnough   = 5003
	CodeNoApplyRecord     = 5004
)

// Response hrc 业务域统一响应结构 {code, message, data}（与 GVA 原生 {code, data, msg} 双轨并存）
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// PageData 分页响应
type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

func OK(c *gin.Context) {
	OKWithData(c, map[string]interface{}{})
}

func OKWithData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: CodeSuccess, Message: "success", Data: data})
}

func OKWithMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{Code: CodeSuccess, Message: message, Data: map[string]interface{}{}})
}

func Fail(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{Code: code, Message: message, Data: nil})
}

func FailWithHttp(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response{Code: code, Message: message, Data: nil})
}
