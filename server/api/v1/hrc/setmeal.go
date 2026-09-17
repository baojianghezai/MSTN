package hrc

import (
	"errors"
	"strconv"

	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type SetmealApi struct{}

type SetmealSaveRequest struct {
	Name            string `json:"name"`
	Price           int64  `json:"price"`
	DurationDays    int    `json:"durationDays"`
	JobsMeanwhile   int    `json:"jobsMeanwhile"`
	ResumeDownloads int    `json:"resumeDownloads"`
	HomePushSlots   int    `json:"homePushSlots"`
	HomeAdSlots     int    `json:"homeAdSlots"`
	EnableVideo     bool   `json:"enableVideo"`
	EnableJobfair   bool   `json:"enableJobfair"`
	Display         bool   `json:"display"`
	Sort            int    `json:"sort"`
	Description     string `json:"description"`
}

func (a *SetmealApi) PublicList(c *gin.Context) {
	list, err := hrcService.ServiceGroupApp.SetmealService.PublicList(c.Request.Context())
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, list)
}

func (a *SetmealApi) Current(c *gin.Context) {
	current, err := hrcService.ServiceGroupApp.SetmealService.Current(c.Request.Context(), middlewarehrc.GetMemberUID(c))
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, current)
}

func (a *SetmealApi) AdminList(c *gin.Context) {
	list, err := hrcService.ServiceGroupApp.SetmealService.AdminList(c.Request.Context())
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, list)
}

func (a *SetmealApi) AdminCreate(c *gin.Context) {
	var req SetmealSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	plan := hrcModel.Setmeal{
		Name: req.Name, Price: req.Price, DurationDays: req.DurationDays,
		JobsMeanwhile: req.JobsMeanwhile, ResumeDownloads: req.ResumeDownloads, HomePushSlots: req.HomePushSlots, HomeAdSlots: req.HomeAdSlots,
		EnableVideo: req.EnableVideo, EnableJobfair: req.EnableJobfair, Display: req.Display, Sort: req.Sort, Description: req.Description,
	}
	if err := hrcService.ServiceGroupApp.SetmealService.Save(c.Request.Context(), &plan); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, plan)
}

func (a *SetmealApi) AdminUpdate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	var req SetmealSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	plan := hrcModel.Setmeal{
		ID: id, Name: req.Name, Price: req.Price, DurationDays: req.DurationDays,
		JobsMeanwhile: req.JobsMeanwhile, ResumeDownloads: req.ResumeDownloads, HomePushSlots: req.HomePushSlots, HomeAdSlots: req.HomeAdSlots,
		EnableVideo: req.EnableVideo, EnableJobfair: req.EnableJobfair, Display: req.Display, Sort: req.Sort, Description: req.Description,
	}
	if err := hrcService.ServiceGroupApp.SetmealService.Save(c.Request.Context(), &plan); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, plan)
}

func (a *SetmealApi) AdminOrders(c *gin.Context) {
	var page request.PageInfo
	_ = c.ShouldBindQuery(&page)
	status, _ := strconv.ParseInt(c.DefaultQuery("status", "0"), 10, 8)
	list, total, err := hrcService.ServiceGroupApp.OrderService.AdminList(c.Request.Context(), page, int8(status))
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, PageData{List: list, Total: total, Page: page.Page, PageSize: page.PageSize})
}

type ConfirmOrderRequest struct {
	PayAmount int64  `json:"payAmount"`
	Payment   string `json:"payment"`
}

func (a *SetmealApi) AdminConfirmOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	var req ConfirmOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	err = hrcService.ServiceGroupApp.OrderService.ConfirmPaid(c.Request.Context(), id, req.PayAmount, req.Payment)
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
	OK(c)
}
