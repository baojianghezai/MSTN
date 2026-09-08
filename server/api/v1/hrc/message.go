package hrc

import (
	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type MessageApi struct{}

type MessageBatchRequest struct {
	IDs []uint64 `json:"ids" binding:"required"`
}

type PmsPageData struct {
	List     []hrcModel.Pms `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
}

type MessageUnreadData struct {
	Unread int64 `json:"unread"`
}

// List returns messages for the authenticated personal or company member.
// @Tags HrcMessage
// @Summary 站内信列表
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} Response{data=PmsPageData}
// @Security ApiKeyAuth
// @Router /api/v1/personal/messages [get]
// @Router /api/v1/company/messages [get]
func (a *MessageApi) List(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	if pageInfo.Page <= 0 {
		pageInfo.Page = 1
	}
	if pageInfo.PageSize <= 0 {
		pageInfo.PageSize = 20
	}
	list, total, err := hrcService.ServiceGroupApp.MessageService.List(c.Request.Context(), middlewarehrc.GetMemberUID(c), pageInfo)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, PmsPageData{List: list, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}

// MarkRead marks the authenticated member's messages as read.
// @Tags HrcMessage
// @Summary 批量标记站内信已读
// @Accept application/json
// @Produce json
// @Param data body MessageBatchRequest true "消息 id 列表，最多 100 条"
// @Success 200 {object} Response
// @Security ApiKeyAuth
// @Router /api/v1/personal/messages/read [put]
// @Router /api/v1/company/messages/read [put]
func (a *MessageApi) MarkRead(c *gin.Context) {
	var req MessageBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	if err := hrcService.ServiceGroupApp.MessageService.MarkRead(c.Request.Context(), middlewarehrc.GetMemberUID(c), req.IDs); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// Delete removes the authenticated member's messages.
// @Tags HrcMessage
// @Summary 批量删除站内信
// @Accept application/json
// @Produce json
// @Param data body MessageBatchRequest true "消息 id 列表，最多 100 条"
// @Success 200 {object} Response
// @Security ApiKeyAuth
// @Router /api/v1/personal/messages [delete]
// @Router /api/v1/company/messages [delete]
func (a *MessageApi) Delete(c *gin.Context) {
	var req MessageBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	if err := hrcService.ServiceGroupApp.MessageService.Delete(c.Request.Context(), middlewarehrc.GetMemberUID(c), req.IDs); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// UnreadCount returns the authenticated member's current unread message total.
// @Tags HrcMessage
// @Summary 站内信未读数
// @Produce json
// @Success 200 {object} Response{data=MessageUnreadData}
// @Security ApiKeyAuth
// @Router /api/v1/personal/messages/unread-count [get]
// @Router /api/v1/company/messages/unread-count [get]
func (a *MessageApi) UnreadCount(c *gin.Context) {
	unread, err := hrcService.ServiceGroupApp.MessageService.UnreadCount(c.Request.Context(), middlewarehrc.GetMemberUID(c))
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, MessageUnreadData{Unread: unread})
}
