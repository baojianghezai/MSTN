package hrc

import "time"

// PersonalJobsApply 投递记录（01 §2.6，S1 决议：同一企业对同一份简历去重）
type PersonalJobsApply struct {
	DID          uint64     `gorm:"column:did;primaryKey;autoIncrement" json:"did"`
	ResumeID     uint64     `gorm:"column:resume_id;index;uniqueIndex:uk_uid_resume_company" json:"resumeId"`
	ResumeName   string     `gorm:"column:resume_name;size:60" json:"resumeName"`
	PersonalUID  uint64     `gorm:"column:personal_uid;index;uniqueIndex:uk_uid_resume_company" json:"personalUid"` // 求职者 uid
	JobsID       uint64     `gorm:"column:jobs_id;index" json:"jobsId"`
	JobsName     string     `gorm:"column:jobs_name;size:60" json:"jobsName"`
	CompanyID    uint64     `gorm:"column:company_id" json:"companyId"`
	CompanyName  string     `gorm:"column:company_name;size:60" json:"companyName"`
	CompanyUID   uint64     `gorm:"column:company_uid;index;uniqueIndex:uk_uid_resume_company" json:"companyUid"`
	ApplyAddtime time.Time  `gorm:"column:apply_addtime" json:"applyAddtime"`
	PersonalLook int8       `gorm:"column:personal_look;default:1" json:"personalLook"` // 1未读 2已读
	Notes        string     `gorm:"column:notes;size:200" json:"notes"`                 // 求职附言
	IsReply      int8       `gorm:"column:is_reply;default:0" json:"isReply"`           // 回复状态 0-4
	ReplyTime    *time.Time `gorm:"column:reply_time" json:"replyTime"`
}

func (PersonalJobsApply) TableName() string { return "ms_personal_jobs_apply" }
