package hrc

import (
	"errors"
	"strconv"

	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type OrderApi struct{}

type CreateOrderRequest struct {
	SetmealID uint64 `json:"setmealId" binding:"required"`
}

func (a *OrderApi) Create(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "请选择套餐")
		return
	}
	order, err := hrcService.ServiceGroupApp.OrderService.CreateSetmealOrder(c.Request.Context(), middlewarehrc.GetMemberUID(c), req.SetmealID)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, order)
}

func (a *OrderApi) List(c *gin.Context) {
	var page request.PageInfo
	_ = c.ShouldBindQuery(&page)
	list, total, err := hrcService.ServiceGroupApp.OrderService.ListMine(c.Request.Context(), middlewarehrc.GetMemberUID(c), page)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, PageData{List: list, Total: total, Page: page.Page, PageSize: page.PageSize})
}

func (a *OrderApi) Cancel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	err = hrcService.ServiceGroupApp.OrderService.Close(c.Request.Context(), middlewarehrc.GetMemberUID(c), id)
	if err != nil {
		code := CodeParamError
		if errors.Is(err, hrcService.ErrOrderClosed) {
			code = CodeOrderClosed
		}
		Fail(c, code, err.Error())
		return
	}
	OK(c)
}
