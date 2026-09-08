package hrc

import (
	"errors"
	"net/http"
	"strconv"

	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
	"github.com/smartwalle/alipay/v3"
)

type PaymentApi struct{}

type StartPaymentRequest struct {
	Provider string `json:"provider" binding:"required"`
}

func (a *PaymentApi) Start(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	var req StartPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "请选择支付方式")
		return
	}
	result, err := hrcService.ServiceGroupApp.PaymentService.Start(c.Request.Context(), middlewarehrc.GetMemberUID(c), id, req.Provider)
	if err != nil {
		code := CodeParamError
		if errors.Is(err, hrcService.ErrOrderNotFound) {
			code = CodeOrderNotFound
		} else if errors.Is(err, hrcService.ErrOrderPaid) {
			code = CodeOrderPaid
		} else if errors.Is(err, hrcService.ErrOrderClosed) {
			code = CodeOrderClosed
		}
		Fail(c, code, err.Error())
		return
	}
	OKWithData(c, result)
}

func (a *PaymentApi) WechatNotify(c *gin.Context) {
	if err := hrcService.ServiceGroupApp.PaymentService.HandleWechatNotify(c.Request.Context(), c.Request); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "处理失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
}

func (a *PaymentApi) AlipayNotify(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		c.String(http.StatusBadRequest, "failure")
		return
	}
	if err := hrcService.ServiceGroupApp.PaymentService.HandleAlipayNotify(c.Request.Context(), c.Request.PostForm); err != nil {
		c.String(http.StatusInternalServerError, "failure")
		return
	}
	alipay.ACKNotification(c.Writer)
}
