package hrc

import (
	"strconv"

	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type TalentApi struct{}

type TalentSearchRequest struct {
	request.PageInfo
	Keyword    string `form:"keyword"`
	District   string `form:"district"`
	Education  uint16 `form:"education"`
	Experience uint16 `form:"experience"`
	WageMin    uint16 `form:"wageMin"`
	WageMax    uint16 `form:"wageMax"`
}

type PublicResumePageData struct {
	List     []hrcService.PublicResume `json:"list"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"pageSize"`
}

type CompanyTalentPageData struct {
	List     []hrcService.CompanyTalentItem `json:"list"`
	Total    int64                          `json:"total"`
	Page     int                            `json:"page"`
	PageSize int                            `json:"pageSize"`
}

type FavoriteTalentPageData struct {
	List     []hrcService.FavoriteTalentItem `json:"list"`
	Total    int64                           `json:"total"`
	Page     int                             `json:"page"`
	PageSize int                             `json:"pageSize"`
}

type TalentUnlockData struct {
	Detail        hrcService.TalentUnlockedDetail `json:"detail"`
	NewlyUnlocked bool                            `json:"newlyUnlocked"`
}

type TalentFollowUpRequest struct {
	FollowUp int8 `json:"followUp"`
}

type TalentFavoriteRequest struct {
	ResumeID uint64 `json:"resumeId" binding:"required"`
}

type TalentFavoriteData = hrcModel.CompanyFavorite

// Search returns publicly visible, contact-masked resumes.
// @Tags HrcTalent
// @Summary 公开简历列表
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param keyword query string false "关键字"
// @Param district query string false "地区"
// @Param education query int false "学历"
// @Param experience query int false "经验"
// @Param wage query int false "期望薪资"
// @Success 200 {object} Response{data=PublicResumePageData}
// @Router /api/v1/resumes [get]
func (a *TalentApi) Search(c *gin.Context) {
	a.listPublic(c, false)
}

// TalentList returns publicly visible premium-talent resumes.
// @Tags HrcTalent
// @Summary 高级人才专区
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param keyword query string false "关键字"
// @Param district query string false "地区"
// @Param education query int false "学历"
// @Param experience query int false "经验"
// @Param wage query int false "期望薪资"
// @Success 200 {object} Response{data=PublicResumePageData}
// @Router /api/v1/resumes/talents [get]
func (a *TalentApi) TalentList(c *gin.Context) {
	a.listPublic(c, true)
}

func (a *TalentApi) listPublic(c *gin.Context, talentOnly bool) {
	var req TalentSearchRequest
	_ = c.ShouldBindQuery(&req)
	normalizeTalentPageInfo(&req.PageInfo)
	list, total, err := hrcService.ServiceGroupApp.TalentService.Search(c.Request.Context(), req.PageInfo, hrcService.TalentSearch{
		Keyword: req.Keyword, District: req.District, Education: req.Education,
		Experience: req.Experience, WageMin: req.WageMin, WageMax: req.WageMax, TalentOnly: talentOnly,
	})
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, PublicResumePageData{List: list, Total: total, Page: req.Page, PageSize: req.PageSize})
}

// PublicDetail returns a publicly visible, contact-masked resume detail.
// @Tags HrcTalent
// @Summary 公开简历详情
// @Produce json
// @Param id path int true "简历 id"
// @Success 200 {object} Response{data=hrcService.PublicResumeDetail}
// @Router /api/v1/resumes/{id} [get]
func (a *TalentApi) PublicDetail(c *gin.Context) {
	resumeID, ok := parseTalentID(c, "id")
	if !ok {
		return
	}
	detail, err := hrcService.ServiceGroupApp.TalentService.PublicDetail(c.Request.Context(), resumeID)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, detail)
}

// Unlock lets a company unlock a public resume's full contact information.
// @Tags HrcTalent
// @Summary 企业解锁人才简历
// @Produce json
// @Param id path int true "简历 id"
// @Success 200 {object} Response{data=TalentUnlockData}
// @Security ApiKeyAuth
// @Router /api/v1/company/talents/{id}/unlock [post]
func (a *TalentApi) Unlock(c *gin.Context) {
	resumeID, ok := parseTalentID(c, "id")
	if !ok {
		return
	}
	detail, newlyUnlocked, err := hrcService.ServiceGroupApp.TalentService.Unlock(c.Request.Context(), middlewarehrc.GetMemberUID(c), resumeID)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, TalentUnlockData{Detail: *detail, NewlyUnlocked: newlyUnlocked})
}

// GetUnlocked returns an already unlocked candidate with full contact details.
// @Tags HrcTalent
// @Summary 企业人才库简历详情
// @Produce json
// @Param id path int true "简历 id"
// @Success 200 {object} Response{data=hrcService.TalentUnlockedDetail}
// @Security ApiKeyAuth
// @Router /api/v1/company/talents/{id} [get]
func (a *TalentApi) GetUnlocked(c *gin.Context) {
	resumeID, ok := parseTalentID(c, "id")
	if !ok {
		return
	}
	detail, err := hrcService.ServiceGroupApp.TalentService.GetUnlocked(c.Request.Context(), middlewarehrc.GetMemberUID(c), resumeID)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, detail)
}

// ListUnlocked returns the authenticated company's talent library.
// @Tags HrcTalent
// @Summary 企业人才库
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} Response{data=CompanyTalentPageData}
// @Security ApiKeyAuth
// @Router /api/v1/company/talents [get]
func (a *TalentApi) ListUnlocked(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	normalizeTalentPageInfo(&pageInfo)
	list, total, err := hrcService.ServiceGroupApp.TalentService.ListUnlocked(c.Request.Context(), middlewarehrc.GetMemberUID(c), pageInfo)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, CompanyTalentPageData{List: list, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}

// SetFollowUp saves a follow-up state for an unlocked candidate.
// @Tags HrcTalent
// @Summary 更新人才库跟进状态
// @Accept application/json
// @Produce json
// @Param id path int true "简历 id"
// @Param data body TalentFollowUpRequest true "跟进状态"
// @Success 200 {object} Response
// @Security ApiKeyAuth
// @Router /api/v1/company/talents/{id}/follow-up [put]
func (a *TalentApi) SetFollowUp(c *gin.Context) {
	resumeID, ok := parseTalentID(c, "id")
	if !ok {
		return
	}
	var req TalentFollowUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	if err := hrcService.ServiceGroupApp.TalentService.SetFollowUp(c.Request.Context(), middlewarehrc.GetMemberUID(c), resumeID, req.FollowUp); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// ListFavorites returns the authenticated company's favorites without contact details.
// @Tags HrcTalent
// @Summary 企业收藏简历列表
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} Response{data=FavoriteTalentPageData}
// @Security ApiKeyAuth
// @Router /api/v1/company/favorites [get]
func (a *TalentApi) ListFavorites(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	normalizeTalentPageInfo(&pageInfo)
	list, total, err := hrcService.ServiceGroupApp.TalentService.ListFavorites(c.Request.Context(), middlewarehrc.GetMemberUID(c), pageInfo)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, FavoriteTalentPageData{List: list, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}

// Favorite adds a public candidate to the authenticated company's favorites.
// @Tags HrcTalent
// @Summary 收藏人才简历
// @Accept application/json
// @Produce json
// @Param data body TalentFavoriteRequest true "简历 id"
// @Success 200 {object} Response{data=TalentFavoriteData}
// @Security ApiKeyAuth
// @Router /api/v1/company/favorites [post]
func (a *TalentApi) Favorite(c *gin.Context) {
	var req TalentFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	favorite, err := hrcService.ServiceGroupApp.TalentService.Favorite(c.Request.Context(), middlewarehrc.GetMemberUID(c), req.ResumeID)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, favorite)
}

// Unfavorite removes one of the authenticated company's favorite records.
// @Tags HrcTalent
// @Summary 取消收藏人才简历
// @Produce json
// @Param id path int true "收藏 id"
// @Success 200 {object} Response
// @Security ApiKeyAuth
// @Router /api/v1/company/favorites/{id} [delete]
func (a *TalentApi) Unfavorite(c *gin.Context) {
	favoriteID, ok := parseTalentID(c, "id")
	if !ok {
		return
	}
	if err := hrcService.ServiceGroupApp.TalentService.Unfavorite(c.Request.Context(), middlewarehrc.GetMemberUID(c), favoriteID); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

func parseTalentID(c *gin.Context, key string) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil || id == 0 {
		Fail(c, CodeParamError, "参数错误")
		return 0, false
	}
	return id, true
}

func normalizeTalentPageInfo(pageInfo *request.PageInfo) {
	if pageInfo.Page <= 0 {
		pageInfo.Page = 1
	}
	if pageInfo.PageSize <= 0 {
		pageInfo.PageSize = 20
	}
}
