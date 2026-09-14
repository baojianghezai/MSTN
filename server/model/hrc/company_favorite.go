package hrc

import "time"

// CompanyFavorite 企业收藏的候选人。收藏不解锁联系方式，也不消耗套餐次数。
type CompanyFavorite struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CompanyUID uint64    `gorm:"column:company_uid;index;uniqueIndex:uk_company_resume_favorite" json:"companyUid"`
	ResumeID   uint64    `gorm:"column:resume_id;index;uniqueIndex:uk_company_resume_favorite" json:"resumeId"`
	AddTime    time.Time `gorm:"column:addtime;index" json:"addtime"`
}

func (CompanyFavorite) TableName() string { return "ms_company_favorites" }
