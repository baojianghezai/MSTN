package hrc

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"gorm.io/gorm"
)

var ErrCompanyNotFound = errors.New("企业不存在或暂不可见")

// CompanySearchFilter is the public company-directory filter set. It mirrors the
// legacy directory's keyword, nature, and trade filters and adds scale and district.
type CompanySearchFilter struct {
	Keyword  string
	Nature   uint16
	Trade    uint16
	Scale    uint16
	District string
}

// PublicCompanyItem deliberately excludes contacts, certificates, and member data.
type PublicCompanyItem struct {
	ID          uint64    `json:"id"`
	UID         uint64    `json:"uid"`
	CompanyName string    `json:"companyname"`
	Nature      uint16    `json:"nature"`
	NatureCN    string    `json:"natureCn"`
	Trade       uint16    `json:"trade"`
	TradeCN     string    `json:"tradeCn"`
	District    string    `json:"district"`
	DistrictCN  string    `json:"districtCn"`
	Scale       uint16    `json:"scale"`
	ScaleCN     string    `json:"scaleCn"`
	Logo        string    `json:"logo"`
	ShortName   string    `json:"shortName"`
	ShortDesc   string    `json:"shortDesc"`
	Tag         string    `json:"tag"`
	Refreshtime time.Time `json:"refreshtime"`
	JobsCount   int64     `json:"jobsCount"`
}

type PublicCompanyJob struct {
	ID          uint64    `json:"id"`
	JobsName    string    `json:"jobsName"`
	NatureCN    string    `json:"natureCn"`
	CategoryCN  string    `json:"categoryCn"`
	DistrictCN  string    `json:"districtCn"`
	Education   uint16    `json:"education"`
	Experience  uint16    `json:"experience"`
	MinWage     int       `json:"minwage"`
	MaxWage     int       `json:"maxwage"`
	Negotiable  int8      `json:"negotiable"`
	Amount      uint16    `json:"amount"`
	Emergency   int8      `json:"emergency"`
	Stick       int8      `json:"stick"`
	AddTime     time.Time `json:"addtime"`
	Refreshtime time.Time `json:"refreshtime"`
}

type PublicCompanyDetail struct {
	Company   PublicCompanyItem  `json:"company"`
	Contents  string             `json:"contents"`
	Address   string             `json:"address"`
	Website   string             `json:"website"`
	Jobs      []PublicCompanyJob `json:"jobs"`
	JobsTotal int64              `json:"jobsTotal"`
}

type CompanySearchService struct{}

func publicJobsQuery(ctx context.Context, companyID uint64) *gorm.DB {
	return global.GVA_DB.WithContext(ctx).Model(&hrcModel.Jobs{}).
		Where("company_id = ? AND display = 1 AND audit = 1 AND deleted_at IS NULL AND (deadline IS NULL OR deadline > ?)", companyID, time.Now())
}

func publicCompanyQuery(ctx context.Context) *gorm.DB {
	activeJobs := global.GVA_DB.WithContext(ctx).Model(&hrcModel.Jobs{}).
		Select("company_id, COUNT(*) AS jobs_count").
		Where("display = 1 AND audit = 1 AND deleted_at IS NULL AND (deadline IS NULL OR deadline > ?)", time.Now()).
		Group("company_id")

	return global.GVA_DB.WithContext(ctx).
		Table("ms_company_profile AS cp").
		Select("cp.id, cp.uid, cp.companyname, cp.nature, cp.nature_cn, cp.trade, cp.trade_cn, cp.district, cp.district_cn, cp.scale, cp.scale_cn, cp.logo, cp.short_name, cp.short_desc, cp.tag, cp.refreshtime, COALESCE(active_jobs.jobs_count, 0) AS jobs_count").
		Joins("LEFT JOIN (?) AS active_jobs ON active_jobs.company_id = cp.id", activeJobs).
		Where("cp.audit = 1 AND cp.user_status = 1 AND cp.companyname IS NOT NULL AND cp.companyname <> ''")
}

func (s *CompanySearchService) Search(ctx context.Context, f CompanySearchFilter, info request.PageInfo) ([]PublicCompanyItem, int64, error) {
	db := publicCompanyQuery(ctx)
	if keyword := strings.TrimSpace(f.Keyword); keyword != "" {
		db = db.Where("cp.companyname LIKE ? OR cp.short_name LIKE ? OR cp.short_desc LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if f.Nature > 0 {
		db = db.Where("cp.nature = ?", f.Nature)
	}
	if f.Trade > 0 {
		db = db.Where("cp.trade = ?", f.Trade)
	}
	if f.Scale > 0 {
		db = db.Where("cp.scale = ?", f.Scale)
	}
	if district := strings.TrimSpace(f.District); district != "" {
		db = db.Where("cp.district = ? OR cp.district_cn = ?", district, district)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	var list []PublicCompanyItem
	if err := db.Order("COALESCE(active_jobs.jobs_count, 0) DESC, cp.refreshtime DESC, cp.id DESC").Limit(limit).Offset(offset).Scan(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *CompanySearchService) findCompany(ctx context.Context, id uint64) (*PublicCompanyItem, error) {
	var company PublicCompanyItem
	result := publicCompanyQuery(ctx).Where("cp.id = ?", id).Limit(1).Scan(&company)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrCompanyNotFound
	}
	return &company, nil
}

func (s *CompanySearchService) Jobs(ctx context.Context, companyID uint64, info request.PageInfo) ([]PublicCompanyJob, int64, error) {
	if _, err := s.findCompany(ctx, companyID); err != nil {
		return nil, 0, err
	}
	db := publicJobsQuery(ctx, companyID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	var jobs []PublicCompanyJob
	if err := db.Select("id, jobs_name, nature_cn, category_cn, district_cn, education, experience, minwage, maxwage, negotiable, amount, emergency, stick, addtime, refreshtime").
		Order("stick DESC, emergency DESC, refreshtime DESC, id DESC").Limit(limit).Offset(offset).Scan(&jobs).Error; err != nil {
		return nil, 0, err
	}
	return jobs, total, nil
}

func (s *CompanySearchService) Detail(ctx context.Context, id uint64) (*PublicCompanyDetail, error) {
	company, err := s.findCompany(ctx, id)
	if err != nil {
		return nil, err
	}

	var profile hrcModel.CompanyProfile
	if err := global.GVA_DB.WithContext(ctx).Select("contents, address, website").Where("id = ?", id).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCompanyNotFound
		}
		return nil, err
	}
	if err := global.GVA_DB.WithContext(ctx).Model(&hrcModel.CompanyProfile{}).Where("id = ?", id).UpdateColumn("click", gorm.Expr("click + 1")).Error; err != nil {
		return nil, err
	}

	jobs, total, err := s.Jobs(ctx, id, request.PageInfo{Page: 1, PageSize: 4})
	if err != nil {
		return nil, err
	}
	return &PublicCompanyDetail{Company: *company, Contents: profile.Contents, Address: profile.Address, Website: profile.Website, Jobs: jobs, JobsTotal: total}, nil
}
