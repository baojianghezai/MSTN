package hrc

import "time"

// 内容类型（#17：资讯/招聘会/帮助 一期简版，统一一张表按 type 区分）
const (
	ArticleTypeNews    int8 = 1 // 资讯
	ArticleTypeJobfair int8 = 2 // 招聘会
	ArticleTypeHelp    int8 = 3 // 帮助
)

// Article 内容文章（资讯/招聘会/帮助）
// 设计依据：07 内容模块一期简版（列表 + 详情）；招聘会额外用 hold_time/address/organizer
type Article struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Type        int8      `gorm:"column:type;index" json:"type"`              // 1=资讯 2=招聘会 3=帮助
	CompanyUID  uint64    `gorm:"column:company_uid;index" json:"companyUid"` // 主办企业 uid（0=平台官方；招聘会由企业举办）
	Title       string    `gorm:"column:title;size:150" json:"title"`         // 标题
	Summary     string    `gorm:"column:summary;size:255" json:"summary"`     // 摘要
	Cover       string    `gorm:"column:cover;size:255" json:"cover"`         // 封面图
	Content     string    `gorm:"column:content;type:text" json:"content"`    // 正文（HTML/纯文本）
	Source      string    `gorm:"column:source;size:60" json:"source"`        // 来源/作者
	HoldTime    string    `gorm:"column:hold_time;size:60" json:"holdTime"`   // 招聘会举办时间
	Address     string    `gorm:"column:address;size:150" json:"address"`     // 招聘会地点
	Organizer   string    `gorm:"column:organizer;size:100" json:"organizer"` // 主办方
	Sort        int       `gorm:"column:sort;default:0" json:"sort"`
	Display     int8      `gorm:"column:display;default:1" json:"display"` // 1=展示 2=隐藏
	Click       uint      `gorm:"column:click;default:0" json:"click"`
	AddTime     time.Time `gorm:"column:addtime" json:"addtime"`
	UpdateTime  time.Time `gorm:"column:update_time" json:"updateTime"`
	SignupCount int       `gorm:"-" json:"signupCount"` // 招聘会报名人数（不落库，查询时聚合）
}

func (Article) TableName() string { return "ms_article" }

// JobfairSignup 招聘会报名（个人用户自行参加；同一招聘会同一人唯一）
type JobfairSignup struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	JobfairID   uint64    `gorm:"column:jobfair_id;index;uniqueIndex:uk_jobfair_personal" json:"jobfairId"` // ms_article.id（type=2）
	PersonalUID uint64    `gorm:"column:personal_uid;index;uniqueIndex:uk_jobfair_personal" json:"personalUid"`
	AddTime     time.Time `gorm:"column:addtime;index" json:"addtime"`
}

func (JobfairSignup) TableName() string { return "ms_jobfair_signup" }
