package hrc

import (
	"context"
	"errors"
	"mime/multipart"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/upload"
	"gorm.io/gorm"
)

var (
	ErrProfileNotFound   = errors.New("资料不存在")
	ErrCompanyNameEmpty  = errors.New("企业名称不能为空")
	ErrCompanyNameExists = errors.New("企业名称已被占用")
)

// ProfileService 个人/企业资料服务（02 §2.2 / §3.1）
type ProfileService struct{}

// GetPersonalProfile 个人资料读取（ms_members_info）
func (s *ProfileService) GetPersonalProfile(uid uint64) (*hrcModel.MembersInfo, error) {
	var info hrcModel.MembersInfo
	err := global.GVA_DB.Where("uid = ?", uid).First(&info).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProfileNotFound
	}
	return &info, err
}

// UpsertPersonalProfile 个人资料维护（存在更新，不存在创建）
func (s *ProfileService) UpsertPersonalProfile(uid uint64, info *hrcModel.MembersInfo) error {
	info.UID = uid
	var count int64
	global.GVA_DB.Model(&hrcModel.MembersInfo{}).Where("uid = ?", uid).Count(&count)
	if count == 0 {
		info.ID = 0
		return global.GVA_DB.Create(info).Error
	}
	return global.GVA_DB.Model(&hrcModel.MembersInfo{}).Where("uid = ?", uid).
		Select("realname", "sex", "sex_cn", "birthday", "residence", "education", "education_cn",
			"major", "major_cn", "experience", "experience_cn", "phone", "height", "marriage",
			"marriage_cn", "display_name", "qq", "weixin").
		Updates(info).Error
}

// GetCompanyProfile 企业资料读取（ms_company_profile）
func (s *ProfileService) GetCompanyProfile(uid uint64) (*hrcModel.CompanyProfile, error) {
	var p hrcModel.CompanyProfile
	err := global.GVA_DB.Where("uid = ?", uid).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProfileNotFound
	}
	return &p, err
}

// UpsertCompanyProfile 企业资料维护（提交即触发审核：audit=2 待审核；companyname 唯一）
func (s *ProfileService) UpsertCompanyProfile(uid uint64, p *hrcModel.CompanyProfile) error {
	if p.CompanyName == nil || *p.CompanyName == "" {
		return ErrCompanyNameEmpty
	}
	name := *p.CompanyName
	// 企业名唯一（排除自身）
	var dup int64
	global.GVA_DB.Model(&hrcModel.CompanyProfile{}).
		Where("companyname = ? AND uid != ?", name, uid).Count(&dup)
	if dup > 0 {
		return ErrCompanyNameExists
	}
	p.UID = uid
	p.Audit = 2 // 提交即进入待审核
	var count int64
	global.GVA_DB.Model(&hrcModel.CompanyProfile{}).Where("uid = ?", uid).Count(&count)
	if count == 0 {
		p.ID = 0
		p.AddTime = hrcModel.Now()
		p.Refreshtime = hrcModel.Now()
		p.Click = 1
		p.UserStatus = 1
		return global.GVA_DB.Create(p).Error
	}
	return global.GVA_DB.Model(&hrcModel.CompanyProfile{}).Where("uid = ?", uid).
		Select("companyname", "nature", "nature_cn", "trade", "trade_cn", "district", "district_cn",
			"scale", "scale_cn", "registered", "address", "contact", "telephone", "landline_tel",
			"email", "website", "certificate_img", "logo", "contents", "tag", "short_name", "short_desc",
			"audit", "refreshtime").
		Updates(map[string]interface{}{
			"companyname":     name,
			"nature":          p.Nature,
			"nature_cn":       p.NatureCN,
			"trade":           p.Trade,
			"trade_cn":        p.TradeCN,
			"district":        p.District,
			"district_cn":     p.DistrictCN,
			"scale":           p.Scale,
			"scale_cn":        p.ScaleCN,
			"registered":      p.Registered,
			"address":         p.Address,
			"contact":         p.Contact,
			"telephone":       p.Telephone,
			"landline_tel":    p.LandlineTel,
			"email":           p.Email,
			"website":         p.Website,
			"certificate_img": p.CertificateImg,
			"logo":            p.Logo,
			"contents":        p.Contents,
			"tag":             p.Tag,
			"short_name":      p.ShortName,
			"short_desc":      p.ShortDesc,
			"audit":           2,
			"refreshtime":     hrcModel.Now(),
		}).Error
}

// GetCompanyAudit 企业资质审核状态（未提交资料视为 0=草稿）
func (s *ProfileService) GetCompanyAudit(uid uint64) (int8, error) {
	var p hrcModel.CompanyProfile
	err := global.GVA_DB.Where("uid = ?", uid).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	return p.Audit, err
}

// UploadCompanyImage 上传企业 Logo/证照图片（返回可访问 URL 路径）
func (s *ProfileService) UploadCompanyImage(ctx context.Context, file *multipart.FileHeader) (string, error) {
	url, _, err := upload.NewOss().UploadFile(ctx, file)
	return url, err
}
