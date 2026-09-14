package hrc

import "time"

// VideoInterview 视频面试邀请（腾讯云 TRTC 房间）
// 设计依据：11_v6差异分析 P0#2（v6 qs_video_interview 平移）
// 房间状态机（不落库，读取时按时间计算）：nostart（面试日未到）/ opened（面试日当天）/ overtime（deadline 已过）
type VideoInterview struct {
	ID            uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CompanyUID    uint64    `gorm:"column:company_uid;index" json:"companyUid"`   // 发起企业 uid
	PersonalUID   uint64    `gorm:"column:personal_uid;index" json:"personalUid"` // 被邀个人 uid（简历所属）
	JobsID        uint64    `gorm:"column:jobs_id" json:"jobsId"`                 // 关联职位 id
	JobsName      string    `gorm:"column:jobs_name;size:30" json:"jobsName"`     // 职位名快照
	InterviewTime time.Time `gorm:"column:interview_time" json:"interviewTime"`   // 面试时间
	Deadline      time.Time `gorm:"column:deadline" json:"deadline"`              // 房间有效期（面试时间 +15 天）
	Contact       string    `gorm:"column:contact;size:30" json:"contact"`        // 联系人
	ContactTel    string    `gorm:"column:contact_tel;size:30" json:"contactTel"` // 联系电话
	AddTime       time.Time `gorm:"column:addtime" json:"addtime"`
	CompanyCode   string    `gorm:"column:company_code;size:6;index" json:"companyCode"`   // 企业端房间码
	PersonalCode  string    `gorm:"column:personal_code;size:6;index" json:"personalCode"` // 个人端房间码
}

func (VideoInterview) TableName() string { return "ms_video_interview" }

// VideoDeadlineDays 房间有效期天数（面试时间之后的第 15 天过期，v6 逻辑）
const VideoDeadlineDays = 15
