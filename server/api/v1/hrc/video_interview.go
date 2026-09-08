package hrc

import (
	"strconv"
	"time"

	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type VideoInterviewApi struct{}

// CreateInterview 企业发起视频面试邀请（11 §四 P0#2）
// @Tags HrcVideoInterview
// @Summary 发起视频面试邀请
// @Accept application/json
// @Produce json
// @Param data body VideoInterviewCreateRequest true "邀请参数"
// @Success 200 {object} Response{data=VideoInterviewItem}
// @Router /api/v1/company/video-interviews [post]
func (a *VideoInterviewApi) CreateInterview(c *gin.Context) {
	var req VideoInterviewCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	v, err := hrcService.ServiceGroupApp.VideoInterviewService.CreateInterview(c.Request.Context(), uid, hrcService.VideoInterviewInput{
		ResumeID:      req.ResumeID,
		JobsID:        req.JobsID,
		JobsName:      req.JobsName,
		InterviewTime: req.InterviewTime,
		Contact:       req.Contact,
		Telephone:     req.Telephone,
	})
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, videoInterviewItemFromModel(v, 0, "", ""))
}

// CompanyList 企业端：我发出的视频面试列表
// @Tags HrcVideoInterview
// @Summary 我发出的视频面试列表
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} Response{data=PageData}
// @Router /api/v1/company/video-interviews [get]
func (a *VideoInterviewApi) CompanyList(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	uid := middlewarehrc.GetMemberUID(c)
	list, total, err := hrcService.ServiceGroupApp.VideoInterviewService.CompanyList(c.Request.Context(), uid, pageInfo)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	items := make([]VideoInterviewItem, 0, len(list))
	for _, it := range list {
		items = append(items, videoInterviewItemFromService(it))
	}
	OKWithData(c, PageData{List: items, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}

// CompanyDetail 企业端：视频面试详情
// @Tags HrcVideoInterview
// @Summary 视频面试详情（企业端）
// @Produce json
// @Param id path int true "面试 id"
// @Success 200 {object} Response{data=VideoInterviewItem}
// @Router /api/v1/company/video-interviews/{id} [get]
func (a *VideoInterviewApi) CompanyDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	v, err := hrcService.ServiceGroupApp.VideoInterviewService.CompanyDetail(c.Request.Context(), uid, id)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, videoInterviewItemFromModel(v, 0, "", ""))
}

// CompanyDelete 企业端：删除视频面试邀请
// @Tags HrcVideoInterview
// @Summary 删除视频面试邀请
// @Produce json
// @Param id path int true "面试 id"
// @Success 200 {object} Response
// @Router /api/v1/company/video-interviews/{id} [delete]
func (a *VideoInterviewApi) CompanyDelete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	if err := hrcService.ServiceGroupApp.VideoInterviewService.CompanyDelete(c.Request.Context(), uid, id); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// PersonalList 个人端：我收到的视频面试列表
// @Tags HrcVideoInterview
// @Summary 我收到的视频面试列表
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} Response{data=PageData}
// @Router /api/v1/personal/video-interviews [get]
func (a *VideoInterviewApi) PersonalList(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	uid := middlewarehrc.GetMemberUID(c)
	list, total, err := hrcService.ServiceGroupApp.VideoInterviewService.PersonalList(c.Request.Context(), uid, pageInfo)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	items := make([]VideoInterviewItem, 0, len(list))
	for _, v := range list {
		items = append(items, videoInterviewItemFromModel(&v, 0, "", ""))
	}
	OKWithData(c, PageData{List: items, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}

// PersonalDetail 个人端：视频面试详情
// @Tags HrcVideoInterview
// @Summary 视频面试详情（个人端）
// @Produce json
// @Param id path int true "面试 id"
// @Success 200 {object} Response{data=VideoInterviewItem}
// @Router /api/v1/personal/video-interviews/{id} [get]
func (a *VideoInterviewApi) PersonalDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	v, err := hrcService.ServiceGroupApp.VideoInterviewService.PersonalDetail(c.Request.Context(), uid, id)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, videoInterviewItemFromModel(v, 0, "", ""))
}

// RoomByCode 房间码查面试信息（TRTC 入房，公开）
// @Tags HrcVideoInterview
// @Summary 房间码查询（TRTC 入房）
// @Produce json
// @Param code path string true "房间码（6 位字母数字）"
// @Success 200 {object} Response{data=VideoInterviewRoomData}
// @Router /api/v1/video-interviews/room/{code} [get]
func (a *VideoInterviewApi) RoomByCode(c *gin.Context) {
	code := c.Param("code")
	v, utype, err := hrcService.ServiceGroupApp.VideoInterviewService.RoomByCode(c.Request.Context(), code)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, VideoInterviewRoomData{
		ID:            v.ID,
		JobsName:      v.JobsName,
		InterviewTime: v.InterviewTime,
		Deadline:      v.Deadline,
		RoomStatus:    videoRoomStatusCN(v.InterviewTime, v.Deadline),
		Utype:         utype,
	})
}

// AdminList 后台：视频面试列表（关键字跨职位名/公司名/简历姓名）
// @Tags HrcVideoInterview
// @Summary 视频面试列表（后台）
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param keyword query string false "关键字（职位名/公司名/简历姓名）"
// @Success 200 {object} Response{data=PageData}
// @Router /api/v1/admin/video-interviews [get]
func (a *VideoInterviewApi) AdminList(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	keyword := c.Query("keyword")
	list, total, err := hrcService.ServiceGroupApp.VideoInterviewService.AdminList(c.Request.Context(), pageInfo, keyword)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	items := make([]VideoInterviewItem, 0, len(list))
	for _, it := range list {
		items = append(items, videoInterviewItemFromService(it))
	}
	OKWithData(c, PageData{List: items, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}

// videoInterviewItemFromModel 模型 → 列表项（关联字段由调用方补充）
func videoInterviewItemFromModel(v *hrcModel.VideoInterview, resumeID uint64, fullName, companyName string) VideoInterviewItem {
	return VideoInterviewItem{
		ID:            v.ID,
		CompanyUID:    v.CompanyUID,
		PersonalUID:   v.PersonalUID,
		JobsID:        v.JobsID,
		JobsName:      v.JobsName,
		InterviewTime: v.InterviewTime,
		Deadline:      v.Deadline,
		Contact:       v.Contact,
		ContactTel:    v.ContactTel,
		AddTime:       v.AddTime,
		CompanyCode:   v.CompanyCode,
		PersonalCode:  v.PersonalCode,
		RoomStatus:    videoRoomStatusCN(v.InterviewTime, v.Deadline),
		ResumeID:      resumeID,
		FullName:      fullName,
		CompanyName:   companyName,
	}
}

// videoInterviewItemFromService service 列表项 → 列表项
func videoInterviewItemFromService(it hrcService.VideoInterviewItem) VideoInterviewItem {
	return videoInterviewItemFromModel(&it.VideoInterview, it.ResumeID, it.FullName, it.CompanyName)
}

// videoRoomStatusCN 房间状态机（不落库，按时间计算）：nostart/opened/overtime
func videoRoomStatusCN(interviewTime, deadline int64) string {
	now := time.Now().Unix()
	if deadline < now {
		return "overtime"
	}
	day := time.Unix(interviewTime, 0)
	midnight := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location()).Unix()
	if now < midnight {
		return "nostart"
	}
	return "opened"
}
