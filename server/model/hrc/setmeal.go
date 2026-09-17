package hrc

import "time"

// Setmeal defines a sellable company membership plan. Amounts are stored in cents.
type Setmeal struct {
	ID              uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name            string `gorm:"column:name;size:60;uniqueIndex" json:"name"`
	Price           int64  `gorm:"column:price;default:0" json:"price"`
	DurationDays    int    `gorm:"column:duration_days;default:30" json:"durationDays"`
	JobsMeanwhile   int    `gorm:"column:jobs_meanwhile;default:0" json:"jobsMeanwhile"`
	ResumeDownloads int    `gorm:"column:resume_downloads;default:0" json:"resumeDownloads"`
	HomePushSlots   int    `gorm:"column:home_push_slots;default:0" json:"homePushSlots"`
	HomeAdSlots     int    `gorm:"column:home_ad_slots;default:0" json:"homeAdSlots"`
	EnableVideo     bool   `gorm:"column:enable_video;default:false" json:"enableVideo"`
	EnableJobfair   bool   `gorm:"column:enable_jobfair;default:false" json:"enableJobfair"` // 是否可举办招聘会
	Display         bool   `gorm:"column:display;index" json:"display"`
	Sort            int    `gorm:"column:sort;default:0" json:"sort"`
	Description     string `gorm:"column:description;type:text" json:"description"`
}

func (Setmeal) TableName() string { return "ms_setmeal" }

// MembersSetmeal is the active entitlement snapshot. It deliberately keeps plan
// values so future plan edits never change an already purchased entitlement.
type MembersSetmeal struct {
	ID                   uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UID                  uint64    `gorm:"column:uid;uniqueIndex" json:"uid"`
	SetmealID            uint64    `gorm:"column:setmeal_id;index" json:"setmealId"`
	SetmealName          string    `gorm:"column:setmeal_name;size:60" json:"setmealName"`
	ExpireAt             time.Time `gorm:"column:expire_at;index" json:"expireAt"`
	JobsMeanwhile        int       `gorm:"column:jobs_meanwhile" json:"jobsMeanwhile"`
	ResumeDownloadsTotal int       `gorm:"column:resume_downloads_total" json:"resumeDownloadsTotal"`
	ResumeDownloadsUsed  int       `gorm:"column:resume_downloads_used;default:0" json:"resumeDownloadsUsed"`
	HomePushSlots        int       `gorm:"column:home_push_slots;default:0" json:"homePushSlots"`
	HomeAdSlots          int       `gorm:"column:home_ad_slots;default:0" json:"homeAdSlots"`
	EnableVideo          bool      `gorm:"column:enable_video;default:false" json:"enableVideo"`
	EnableJobfair        bool      `gorm:"column:enable_jobfair;default:false" json:"enableJobfair"`
	UpdatedAt            time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (MembersSetmeal) TableName() string { return "ms_members_setmeal" }

// ResumeDownload records the first time a company unlocks a submitted resume.
// A company can download the same resume again without consuming another quota.
type ResumeDownload struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CompanyUID   uint64    `gorm:"column:company_uid;uniqueIndex:uk_company_resume;index" json:"companyUid"`
	ResumeID     uint64    `gorm:"column:resume_id;uniqueIndex:uk_company_resume;index" json:"resumeId"`
	ApplyID      uint64    `gorm:"column:apply_id;index" json:"applyId"`
	FollowUp     int8      `gorm:"column:follow_up;default:0" json:"followUp"` // 0待跟进 1合适 2不合适 3待定 4未接通
	DownloadedAt time.Time `gorm:"column:downloaded_at;index" json:"downloadedAt"`
}

func (ResumeDownload) TableName() string { return "ms_resume_download" }
