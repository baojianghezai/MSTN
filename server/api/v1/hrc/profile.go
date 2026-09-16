package hrc

import (
	"time"

	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type ProfileApi struct{}

// GetPersonalProfile 个人资料读取
// @Tags HrcProfile
// @Summary 个人资料读取
// @Produce json
// @Success 200 {object} Response{data=hrcModel.MembersInfo}
// @Router /api/v1/personal/profile [get]
func (a *ProfileApi) GetPersonalProfile(c *gin.Context) {
	uid := middlewarehrc.GetMemberUID(c)
	info, err := hrcService.ServiceGroupApp.ProfileService.GetPersonalProfile(uid)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, info)
}

// UpdatePersonalProfile 个人资料维护
// @Tags HrcProfile
// @Summary 个人资料维护
// @Accept application/json
// @Produce json
// @Param data body PersonalProfileRequest true "个人资料"
// @Success 200 {object} Response
// @Router /api/v1/personal/profile [post]
func (a *ProfileApi) UpdatePersonalProfile(c *gin.Context) {
	var req PersonalProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	info := &hrcModel.MembersInfo{
		RealName:     req.RealName,
		Sex:          req.Sex,
		SexCN:        req.SexCN,
		Birthday:     parseISOTimePtr(req.Birthday),
		Residence:    req.Residence,
		Education:    req.Education,
		EducationCN:  req.EducationCN,
		Major:        req.Major,
		MajorCN:      req.MajorCN,
		Experience:   req.Experience,
		ExperienceCN: req.ExperienceCN,
		Phone:        req.Phone,
		Height:       req.Height,
		Marriage:     req.Marriage,
		MarriageCN:   req.MarriageCN,
		DisplayName:  req.DisplayName,
		QQ:           req.QQ,
		Weixin:       req.Weixin,
	}
	if err := hrcService.ServiceGroupApp.ProfileService.UpsertPersonalProfile(uid, info); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// GetCompanyProfile 企业资料读取
// @Tags HrcProfile
// @Summary 企业资料读取
// @Produce json
// @Success 200 {object} Response{data=hrcModel.CompanyProfile}
// @Router /api/v1/company/profile [get]
func (a *ProfileApi) GetCompanyProfile(c *gin.Context) {
	uid := middlewarehrc.GetMemberUID(c)
	p, err := hrcService.ServiceGroupApp.ProfileService.GetCompanyProfile(uid)
	if err != nil {
		if err == hrcService.ErrProfileNotFound {
			OKWithData(c, gin.H{})
			return
		}
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, p)
}

// UpdateCompanyProfile 企业资料维护（提交即触发审核）
// @Tags HrcProfile
// @Summary 企业资料维护
// @Accept application/json
// @Produce json
// @Param data body CompanyProfileRequest true "企业资料"
// @Success 200 {object} Response
// @Router /api/v1/company/profile [post]
func (a *ProfileApi) UpdateCompanyProfile(c *gin.Context) {
	var req CompanyProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	p := &hrcModel.CompanyProfile{
		CompanyName:    &req.CompanyName,
		Nature:         req.Nature,
		NatureCN:       req.NatureCN,
		Trade:          req.Trade,
		TradeCN:        req.TradeCN,
		District:       req.District,
		DistrictCN:     req.DistrictCN,
		Scale:          req.Scale,
		ScaleCN:        req.ScaleCN,
		Registered:     req.Registered,
		Address:        req.Address,
		Contact:        req.Contact,
		Telephone:      req.Telephone,
		LandlineTel:    req.LandlineTel,
		Email:          req.Email,
		Website:        req.Website,
		CertificateImg: req.CertificateImg,
		Logo:           req.Logo,
		Contents:       req.Contents,
		Tag:            req.Tag,
		ShortName:      req.ShortName,
		ShortDesc:      req.ShortDesc,
	}
	if err := hrcService.ServiceGroupApp.ProfileService.UpsertCompanyProfile(uid, p); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// UploadCompanyLogo 上传企业 Logo/证照（返回 URL）
// @Tags HrcProfile
// @Summary 上传企业 Logo/证照
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "图片文件"
// @Success 200 {object} Response
// @Router /api/v1/company/profile/logo [post]
func (a *ProfileApi) UploadCompanyLogo(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		Fail(c, CodeParamError, "接收文件失败")
		return
	}
	url, err := hrcService.ServiceGroupApp.ProfileService.UploadCompanyImage(c.Request.Context(), file)
	if err != nil {
		Fail(c, CodeParamError, "上传失败")
		return
	}
	OKWithData(c, gin.H{"url": url})
}

// GetCompanyAuditStatus 企业审核状态
// @Tags HrcProfile
// @Summary 企业审核状态
// @Produce json
// @Success 200 {object} Response{data=CompanyAuditData}
// @Router /api/v1/company/profile/audit-status [get]
func (a *ProfileApi) GetCompanyAuditStatus(c *gin.Context) {
	uid := middlewarehrc.GetMemberUID(c)
	audit, err := hrcService.ServiceGroupApp.ProfileService.GetCompanyAudit(uid)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, CompanyAuditData{Audit: audit, AuditCN: companyAuditCN(audit)})
}

func companyAuditCN(audit int8) string {
	switch audit {
	case 0:
		return "未提交"
	case 1:
		return "已通过"
	case 2:
		return "审核中"
	case 3:
		return "未通过"
	default:
		return ""
	}
}

// parseISOTimePtr 解析 ISO8601 日期字符串为 *time.Time
func parseISOTimePtr(s string) *time.Time {
	if s == "" {
		return nil
	}
	layouts := []string{"2006-01-02", time.RFC3339}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}
