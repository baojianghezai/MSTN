package hrc

import "time"

const (
	JobPromotionTypePush int8 = 1 // 首页职位推流
	JobPromotionTypeAd   int8 = 2 // 首页广告位
)

// JobPromotion represents a company job occupying a paid homepage exposure slot.
// One job can occupy at most one slot of each type for the owning company.
type JobPromotion struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UID        uint64    `gorm:"column:uid;index" json:"uid"`
	JobID      uint64    `gorm:"column:job_id;uniqueIndex:uk_job_promotion_type;index" json:"jobId"`
	Type       int8      `gorm:"column:type;uniqueIndex:uk_job_promotion_type;index" json:"type"`
	AdTitle    string    `gorm:"column:ad_title;size:60" json:"adTitle"`
	AdSubtitle string    `gorm:"column:ad_subtitle;size:120" json:"adSubtitle"`
	AdImage    string    `gorm:"column:ad_image;size:255" json:"adImage"`
	Audit      int8      `gorm:"column:audit;default:0;index" json:"audit"` // 0待审 1通过 3不通过（投放需后台审核）
	Reason     string    `gorm:"column:reason;size:255" json:"reason"`      // 不通过原因
	Sort       int       `gorm:"column:sort;default:0" json:"sort"`
	CreatedAt  time.Time `gorm:"column:created_at;index" json:"createdAt"`
}

func (JobPromotion) TableName() string { return "ms_job_promotion" }
