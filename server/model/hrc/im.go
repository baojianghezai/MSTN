package hrc

import "time"

// ImSession 在线对话会话（个人 ↔ 企业，唯一；设计依据：11 §1.1 IM 能力一期最小实现）
// 一期不做多客服/多 HR：同一「个人 + 企业」只有一条会话，职位仅作为发起上下文展示。
type ImSession struct {
	ID             uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PersonalUID    uint64     `gorm:"column:personal_uid;index;uniqueIndex:uk_im_personal_company" json:"personalUid"`
	CompanyUID     uint64     `gorm:"column:company_uid;index;uniqueIndex:uk_im_personal_company" json:"companyUid"`
	JobsID         uint64     `gorm:"column:jobs_id" json:"jobsId"`             // 发起上下文职位（可选）
	JobsName       string     `gorm:"column:jobs_name;size:60" json:"jobsName"` // 职位名快照
	PersonalName   string     `gorm:"column:personal_name;size:60" json:"personalName"`
	CompanyName    string     `gorm:"column:company_name;size:60" json:"companyName"`
	LastContent    string     `gorm:"column:last_content;size:500" json:"lastContent"`          // 最后一条消息预览
	LastTime       *time.Time `gorm:"column:last_time;index" json:"lastTime"`                   // 最后消息时间（列表排序）
	PersonalUnread int        `gorm:"column:personal_unread;default:0" json:"personalUnread"`   // 个人未读数
	CompanyUnread  int        `gorm:"column:company_unread;default:0" json:"companyUnread"`     // 企业未读数
	PersonalDelete int8       `gorm:"column:personal_deleted;default:0" json:"personalDeleted"` // 个人侧删除标记
	CompanyDelete  int8       `gorm:"column:company_deleted;default:0" json:"companyDeleted"`   // 企业侧删除标记
	AddTime        time.Time  `gorm:"column:addtime" json:"addtime"`
	UpdateTime     time.Time  `gorm:"column:update_time" json:"updateTime"`
}

func (ImSession) TableName() string { return "ms_im_session" }

// ImMessage 在线对话消息（设计依据：11 §1.1；一期为文本消息）
type ImMessage struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	SessionID uint64    `gorm:"column:session_id;index" json:"sessionId"`
	FromUID   uint64    `gorm:"column:from_uid;index" json:"fromUid"`
	ToUID     uint64    `gorm:"column:to_uid;index" json:"toUid"`
	Content   string    `gorm:"column:content;size:1000" json:"content"`
	IsRead    int8      `gorm:"column:is_read;default:0" json:"isRead"` // 0未读 1已读（收件方视角）
	AddTime   time.Time `gorm:"column:addtime;index" json:"addtime"`
}

func (ImMessage) TableName() string { return "ms_im_message" }
