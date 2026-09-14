package hrc

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"gorm.io/gorm"
)

var (
	ErrPromotionTypeInvalid = errors.New("推广类型不合法")
	ErrPromotionJobInvalid  = errors.New("职位不存在、未通过审核或不属于当前企业")
	ErrPromotionEntitlement = errors.New("当前套餐不含首页推广权益")
	ErrPromotionSlotLimit   = errors.New("当前套餐的首页推广位已用完")
	ErrPromotionAdImage     = errors.New("广告位请先上传横幅图片")
	ErrPromotionNotFound    = errors.New("推广记录不存在")
)

// HomePromotionData is the public homepage payload. Push jobs appear in the
// promoted stream while ad jobs are rendered in the dedicated ad placement.
type HomePromotionData struct {
	Push []hrcModel.Jobs `json:"push"`
	Ads  []HomeAd        `json:"ads"`
}

// HomeAd is an image-backed homepage placement associated with an active job.
// Jobs is embedded to retain the existing public job fields for ad consumers.
type HomeAd struct {
	hrcModel.Jobs
	PromotionID uint64 `json:"promotionId"`
	AdTitle     string `json:"adTitle"`
	AdSubtitle  string `json:"adSubtitle"`
	AdImage     string `json:"adImage"`
}

// PromotionCreative is configured only when a company purchases an ad placement.
type PromotionCreative struct {
	AdTitle    string
	AdSubtitle string
	AdImage    string
}

type CompanyPromotionItem struct {
	hrcModel.JobPromotion
	JobsName    string `json:"jobsName"`
	CompanyName string `json:"companyname"`
}

type CompanyPromotionData struct {
	List          []CompanyPromotionItem `json:"list"`
	HomePushSlots int                    `json:"homePushSlots"`
	HomeAdSlots   int                    `json:"homeAdSlots"`
	PushUsed      int64                  `json:"pushUsed"`
	AdUsed        int64                  `json:"adUsed"`
}

type PromotionService struct{}

// ListHome returns only paid, active, and publicly visible job promotions.
func (s *PromotionService) ListHome(ctx context.Context) (*HomePromotionData, error) {
	data := &HomePromotionData{Push: []hrcModel.Jobs{}, Ads: []HomeAd{}}
	rows, err := s.activeQuery(global.GVA_DB.WithContext(ctx)).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var row promotionJob
		if err := global.GVA_DB.ScanRows(rows, &row); err != nil {
			return nil, err
		}
		if row.Type == hrcModel.JobPromotionTypePush {
			data.Push = append(data.Push, row.Jobs)
		} else {
			adImage := row.AdImage
			if adImage == "" {
				adImage = row.CompanyLogo
			}
			adTitle := row.AdTitle
			if adTitle == "" {
				adTitle = row.JobsName
			}
			adSubtitle := row.AdSubtitle
			if adSubtitle == "" {
				adSubtitle = row.CompanyName
			}
			data.Ads = append(data.Ads, HomeAd{
				Jobs:        row.Jobs,
				PromotionID: row.PromotionID,
				AdTitle:     adTitle,
				AdSubtitle:  adSubtitle,
				AdImage:     adImage,
			})
		}
	}
	return data, rows.Err()
}

// ListMine exposes the caller's slot allocation and current promoted jobs.
func (s *PromotionService) ListMine(ctx context.Context, uid uint64) (*CompanyPromotionData, error) {
	data := &CompanyPromotionData{List: []CompanyPromotionItem{}}
	current, err := (&SetmealService{}).Current(ctx, uid)
	if err != nil {
		return nil, err
	}
	if current != nil {
		data.HomePushSlots = current.HomePushSlots
		data.HomeAdSlots = current.HomeAdSlots
	}
	if err := global.GVA_DB.WithContext(ctx).Table("ms_job_promotion AS promotion").
		Select("promotion.*, jobs.jobs_name, jobs.companyname").
		Joins("JOIN ms_jobs AS jobs ON jobs.id = promotion.job_id").
		Where("promotion.uid = ?", uid).
		Order("promotion.type asc, promotion.created_at desc").
		Scan(&data.List).Error; err != nil {
		return nil, err
	}
	for _, item := range data.List {
		if item.Type == hrcModel.JobPromotionTypePush {
			data.PushUsed++
		} else if item.Type == hrcModel.JobPromotionTypeAd {
			data.AdUsed++
		}
	}
	return data, nil
}

// Create allocates a currently active package slot to an approved job.
func (s *PromotionService) Create(ctx context.Context, uid, jobID uint64, promotionType int8, creative PromotionCreative) (*hrcModel.JobPromotion, error) {
	if promotionType != hrcModel.JobPromotionTypePush && promotionType != hrcModel.JobPromotionTypeAd {
		return nil, ErrPromotionTypeInvalid
	}
	creative.AdTitle = strings.TrimSpace(creative.AdTitle)
	creative.AdSubtitle = strings.TrimSpace(creative.AdSubtitle)
	creative.AdImage = strings.TrimSpace(creative.AdImage)
	if promotionType == hrcModel.JobPromotionTypeAd && creative.AdImage == "" {
		return nil, ErrPromotionAdImage
	}
	promotion := &hrcModel.JobPromotion{}
	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current hrcModel.MembersSetmeal
		if err := tx.Where("uid = ?", uid).First(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrPromotionEntitlement
			}
			return err
		}
		if current.ExpireAt.Before(time.Now()) || (promotionType == hrcModel.JobPromotionTypePush && current.HomePushSlots == 0) || (promotionType == hrcModel.JobPromotionTypeAd && current.HomeAdSlots == 0) {
			return ErrPromotionEntitlement
		}
		var job hrcModel.Jobs
		if err := tx.Where("id = ? AND uid = ? AND display = 1 AND audit = 1 AND deleted_at IS NULL AND (deadline IS NULL OR deadline = ? OR deadline > ?)", jobID, uid, time.Time{}, time.Now()).First(&job).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrPromotionJobInvalid
			}
			return err
		}
		var used int64
		if err := tx.Model(&hrcModel.JobPromotion{}).Where("uid = ? AND type = ?", uid, promotionType).Count(&used).Error; err != nil {
			return err
		}
		limit := current.HomePushSlots
		if promotionType == hrcModel.JobPromotionTypeAd {
			limit = current.HomeAdSlots
		}
		if used >= int64(limit) {
			return ErrPromotionSlotLimit
		}
		promotion.UID = uid
		promotion.JobID = jobID
		promotion.Type = promotionType
		if promotionType == hrcModel.JobPromotionTypeAd {
			promotion.AdTitle = creative.AdTitle
			promotion.AdSubtitle = creative.AdSubtitle
			promotion.AdImage = creative.AdImage
		}
		promotion.CreatedAt = time.Now()
		if err := tx.Create(promotion).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return promotion, nil
}

func (s *PromotionService) Delete(ctx context.Context, uid, id uint64) error {
	result := global.GVA_DB.WithContext(ctx).Where("id = ? AND uid = ?", id, uid).Delete(&hrcModel.JobPromotion{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPromotionNotFound
	}
	return nil
}

type promotionJob struct {
	PromotionID uint64 `gorm:"column:promotion_id"`
	Type        int8   `gorm:"column:promotion_type"`
	AdTitle     string `gorm:"column:ad_title"`
	AdSubtitle  string `gorm:"column:ad_subtitle"`
	AdImage     string `gorm:"column:ad_image"`
	CompanyLogo string `gorm:"column:company_logo"`
	hrcModel.Jobs
}

func (s *PromotionService) activeQuery(db *gorm.DB) *gorm.DB {
	now := time.Now()
	return db.Table("ms_job_promotion AS promotion").
		Select("promotion.id AS promotion_id, promotion.type AS promotion_type, promotion.ad_title, promotion.ad_subtitle, promotion.ad_image, jobs.*, company.logo AS company_logo").
		Joins("JOIN ms_jobs AS jobs ON jobs.id = promotion.job_id").
		Joins("LEFT JOIN ms_company_profile AS company ON company.uid = jobs.uid").
		Joins("JOIN ms_members_setmeal AS entitlement ON entitlement.uid = promotion.uid").
		Where("entitlement.expire_at > ? AND jobs.display = 1 AND jobs.audit = 1 AND jobs.deleted_at IS NULL AND (jobs.deadline IS NULL OR jobs.deadline > ?)", now, now).
		Order("promotion.type asc, promotion.sort desc, promotion.created_at desc")
}
