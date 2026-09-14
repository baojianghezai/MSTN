package hrc

import "time"

// CompanyInterview 标准线下面试邀请。
// 简历、职位和企业名称均为创建时的快照，避免后续资料变更影响历史邀请。
type CompanyInterview struct {
	DID              uint64    `gorm:"column:did;primaryKey;autoIncrement" json:"did"`
	ResumeID         uint64    `gorm:"column:resume_id;index;uniqueIndex:uk_company_resume_job" json:"resumeId"`
	ResumeName       string    `gorm:"column:resume_name;size:60" json:"resumeName"`
	ResumeUID        uint64    `gorm:"column:resume_uid;index" json:"resumeUid"`
	JobsID           uint64    `gorm:"column:jobs_id;index;uniqueIndex:uk_company_resume_job" json:"jobsId"`
	JobsName         string    `gorm:"column:jobs_name;size:60" json:"jobsName"`
	CompanyID        uint64    `gorm:"column:company_id" json:"companyId"`
	CompanyName      string    `gorm:"column:company_name;size:60" json:"companyName"`
	CompanyUID       uint64    `gorm:"column:company_uid;index;uniqueIndex:uk_company_resume_job" json:"companyUid"`
	InterviewTime    time.Time `gorm:"column:interview_time;index" json:"interviewTime"`
	Address          string    `gorm:"column:address;size:200" json:"address"`
	Contact          string    `gorm:"column:contact;size:30" json:"contact"`
	Telephone        string    `gorm:"column:telephone;size:30" json:"telephone"`
	Notes            string    `gorm:"column:notes;size:500" json:"notes"`
	InterviewAddtime time.Time `gorm:"column:interview_addtime;index" json:"interviewAddtime"`
	PersonalLook     int8      `gorm:"column:personal_look;default:1" json:"personalLook"` // 1 未读，2 已读
}

func (CompanyInterview) TableName() string { return "ms_company_interview" }
