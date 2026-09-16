package hrc

import (
	"strconv"

	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

// ChatApi 在线对话接口（个人/企业双端共用，utype 由路由中间件注入）
type ChatApi struct{}

// ChatOpenRequest 打开会话请求（peerUid=对方会员 uid）
type ChatOpenRequest struct {
	PeerUID  uint64 `json:"peerUid" binding:"required"` // 对方 uid
	JobsID   uint64 `json:"jobsId"`                     // 上下文职位（可选）
	JobsName string `json:"jobsName"`                   // 职位名快照（随 job 传入）
}

// ChatSendRequest 发送消息请求
type ChatSendRequest struct {
	Content string `json:"content" binding:"required"`
}

// ListSessions 会话列表（09 §4.2 扩展；按最后消息时间倒序）
// @Tags HrcChat
// @Summary 在线对话会话列表
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} Response{data=response.PageResult{list=[]hrcService.ChatSessionItem}}
// @Router /api/v1/personal/chat/sessions [get]
func (a *ChatApi) ListSessions(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	uid := middlewarehrc.GetMemberUID(c)
	utype := middlewarehrc.GetMemberUtype(c)
	list, total, err := hrcService.ServiceGroupApp.ChatService.ListSessions(c.Request.Context(), uid, utype, pageInfo)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, response.PageResult{List: list, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}

// OpenSession 打开/创建会话（幂等）
// @Tags HrcChat
// @Summary 打开或创建在线对话
// @Accept application/json
// @Produce json
// @Param data body ChatOpenRequest true "对方 uid + 职位上下文"
// @Success 200 {object} Response{data=hrcService.ChatSessionItem}
// @Router /api/v1/personal/chat/sessions [post]
func (a *ChatApi) OpenSession(c *gin.Context) {
	var req ChatOpenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	utype := middlewarehrc.GetMemberUtype(c)
	session, err := hrcService.ServiceGroupApp.ChatService.OpenSession(c.Request.Context(), uid, utype, req.PeerUID, req.JobsID, req.JobsName)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, session)
}

// Messages 会话消息（分页，返回按时间正序）
// @Tags HrcChat
// @Summary 会话消息列表
// @Produce json
// @Param id path int true "会话 id"
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} Response{data=response.PageResult{list=[]hrcService.ChatMessageItem}}
// @Router /api/v1/personal/chat/sessions/{id}/messages [get]
func (a *ChatApi) Messages(c *gin.Context) {
	id, ok := parseChatID(c)
	if !ok {
		return
	}
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	uid := middlewarehrc.GetMemberUID(c)
	utype := middlewarehrc.GetMemberUtype(c)
	list, total, err := hrcService.ServiceGroupApp.ChatService.Messages(c.Request.Context(), uid, utype, id, pageInfo)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, response.PageResult{List: list, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}

// SendMessage 发送消息（落库后经 WebSocket 实时推送给对端）
// @Tags HrcChat
// @Summary 发送消息
// @Accept application/json
// @Produce json
// @Param id path int true "会话 id"
// @Param data body ChatSendRequest true "消息内容"
// @Success 200 {object} Response{data=hrcService.ChatMessageItem}
// @Router /api/v1/personal/chat/sessions/{id}/messages [post]
func (a *ChatApi) SendMessage(c *gin.Context) {
	id, ok := parseChatID(c)
	if !ok {
		return
	}
	var req ChatSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	utype := middlewarehrc.GetMemberUtype(c)
	item, peerUID, err := hrcService.ServiceGroupApp.ChatService.SendMessage(c.Request.Context(), uid, utype, id, req.Content)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	hrcService.ChatHubInstance.SendTo(peerUID, gin.H{"type": "message", "data": item})
	OKWithData(c, item)
}

// MarkRead 标记会话已读（清零己方未读）
// @Tags HrcChat
// @Summary 会话标记已读
// @Produce json
// @Param id path int true "会话 id"
// @Success 200 {object} Response
// @Router /api/v1/personal/chat/sessions/{id}/read [put]
func (a *ChatApi) MarkRead(c *gin.Context) {
	id, ok := parseChatID(c)
	if !ok {
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	utype := middlewarehrc.GetMemberUtype(c)
	if err := hrcService.ServiceGroupApp.ChatService.MarkRead(c.Request.Context(), uid, utype, id); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// DeleteSession 删除会话（仅己方隐藏）
// @Tags HrcChat
// @Summary 删除会话（己方）
// @Produce json
// @Param id path int true "会话 id"
// @Success 200 {object} Response
// @Router /api/v1/personal/chat/sessions/{id} [delete]
func (a *ChatApi) DeleteSession(c *gin.Context) {
	id, ok := parseChatID(c)
	if !ok {
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	utype := middlewarehrc.GetMemberUtype(c)
	if err := hrcService.ServiceGroupApp.ChatService.DeleteSession(c.Request.Context(), uid, utype, id); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// Unread 己方在线对话未读总数（顶栏红点）
// @Tags HrcChat
// @Summary 在线对话未读总数
// @Produce json
// @Success 200 {object} Response{data=map[string]int64}
// @Router /api/v1/personal/chat/unread [get]
func (a *ChatApi) Unread(c *gin.Context) {
	uid := middlewarehrc.GetMemberUID(c)
	utype := middlewarehrc.GetMemberUtype(c)
	total, err := hrcService.ServiceGroupApp.ChatService.UnreadTotal(c.Request.Context(), uid, utype)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, gin.H{"unread": total})
}

func parseChatID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		Fail(c, CodeParamError, "参数错误")
		return 0, false
	}
	return id, true
}
