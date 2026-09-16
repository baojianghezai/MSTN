package hrc

import "time"

// ResumeProjectMax 项目经历上限（个人端同一简历限 6 条，04 §2.3）
const ResumeProjectMax = 6

// Resume 简历主表
// 设计依据：01_数据库设计 §2.3
type Resume struct {
	ID                uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UID               uint64     `gorm:"column:uid;index" json:"uid"`
	Display           int8       `gorm:"column:display;default:1" json:"display"`   // 1=公开 2=不公开
	Audit             int8       `gorm:"column:audit;default:1;index" json:"audit"` // 审核状态（01 §3.1 权威枚举）
	Title             string     `gorm:"column:title;size:80" json:"title"`         // 简历标题
	FullName          string     `gorm:"column:fullname;size:15" json:"fullname"`   // 姓名
	Sex               int8       `gorm:"column:sex" json:"sex"`                     // 性别
	SexCN             string     `gorm:"column:sex_cn;size:3" json:"sexCn"`
	Birthdate         uint16     `gorm:"column:birthdate" json:"birthdate"`         // 出生年
	Residence         string     `gorm:"column:residence;size:30" json:"residence"` // 籍贯（省/市/区，斜杠分隔）
	Education         uint16     `gorm:"column:education" json:"education"`         // 最高学历
	EducationCN       string     `gorm:"column:education_cn;size:30" json:"educationCn"`
	Major             uint16     `gorm:"column:major" json:"major"` // 专业
	MajorCN           string     `gorm:"column:major_cn;size:50" json:"majorCn"`
	Experience        uint16     `gorm:"column:experience" json:"experience"` // 工作年限
	ExperienceCN      string     `gorm:"column:experience_cn;size:30" json:"experienceCn"`
	District          string     `gorm:"column:district;size:100" json:"district"` // 期望地区（省/市/区，斜杠分隔，含"不限"）
	DistrictCN        string     `gorm:"column:district_cn;size:255" json:"districtCn"`
	WageMin           uint16     `gorm:"column:wage_min" json:"wageMin"`                      // 期望薪资下限（元/月）
	WageMax           uint16     `gorm:"column:wage_max" json:"wageMax"`                      // 期望薪资上限（元/月）
	WageCN            string     `gorm:"-" json:"wageCn"`                                     // 期望薪资展示文案（不落库，按上下限计算）
	IntentionJobs     string     `gorm:"column:intention_jobs;size:255" json:"intentionJobs"` // 期望职位
	Specialty         string     `gorm:"column:specialty;size:1000" json:"specialty"`         // 自我评价
	Telephone         string     `gorm:"column:telephone;size:50" json:"telephone"`           // 联系电话
	Email             string     `gorm:"column:email;size:60" json:"email"`
	AddTime           time.Time  `gorm:"column:addtime" json:"addtime"`
	Refreshtime       *time.Time `gorm:"column:refreshtime;index" json:"refreshtime"`              // 刷新时间
	CompletePercent   int8       `gorm:"column:complete_percent;default:0" json:"completePercent"` // 完善度
	Talent            int8       `gorm:"column:talent;default:0" json:"talent"`                    // 高级人才
	Entrust           int8       `gorm:"column:entrust;default:0" json:"entrust"`                  // 委托
	DisplayName       int8       `gorm:"column:display_name;default:1" json:"displayName"`         // 1=显示姓名 2=匿名
	Def               int8       `gorm:"column:def;default:0" json:"def"`                          // 默认简历标记
	Photo             int8       `gorm:"column:photo;default:0" json:"photo"`                      // 是否有照片
	PhotoImg          string     `gorm:"column:photo_img;size:255" json:"photoImg"`                // 照片URL
	PhotoAudit        int8       `gorm:"column:photo_audit;default:1" json:"photoAudit"`           // 照片审核 0待审 1通过 3不通过
	PhotoDisplay      int8       `gorm:"column:photo_display;default:1" json:"photoDisplay"`       // 照片是否展示
	WordResume        string     `gorm:"column:word_resume;size:255" json:"wordResume"`            // 附件简历URL
	WordResumeTitle   string     `gorm:"column:word_resume_title;size:255" json:"wordResumeTitle"`
	WordResumeAddtime *time.Time `gorm:"column:word_resume_addtime" json:"wordResumeAddtime"`
	KeyFull           string     `gorm:"column:key_full;type:text" json:"keyFull"`       // 搜索全量索引
	KeyPrecise        string     `gorm:"column:key_precise;type:text" json:"keyPrecise"` // 搜索精确索引
	Click             uint       `gorm:"column:click;default:1" json:"click"`            // 点击量
	Current           uint16     `gorm:"column:current" json:"current"`                  // 目前状态
	CurrentCN         string     `gorm:"column:current_cn;size:50" json:"currentCn"`
	MobileAudit       int8       `gorm:"column:mobile_audit" json:"mobileAudit"` // 手机认证
	DeletedAt         *time.Time `gorm:"column:deleted_at" json:"deletedAt"`     // 软删除（0=未删；04 §3.5 主表软删、子表保留）
}

func (Resume) TableName() string { return "ms_resume" }

// ResumeProject 项目经历（简历子表，个人端同一简历限 6 条）
// 设计依据：04_简历模块设计 §2.3、01_数据库设计 §1.5（v6 qs_resume_project 平移，11 号评审 P0#1 补入）
type ResumeProject struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PID         uint64 `gorm:"column:pid;index" json:"pid"`                     // 简历主表 id（ms_resume.id）
	UID         uint64 `gorm:"column:uid;index" json:"uid"`                     // 简历所属会员 uid
	StartYear   uint16 `gorm:"column:startyear" json:"startyear"`               // 开始年
	StartMonth  uint8  `gorm:"column:startmonth" json:"startmonth"`             // 开始月
	EndYear     uint16 `gorm:"column:endyear" json:"endyear"`                   // 结束年
	EndMonth    uint8  `gorm:"column:endmonth" json:"endmonth"`                 // 结束月
	ToDate      int8   `gorm:"column:todate" json:"todate"`                     // 至今标记（0=已结束 1=至今）
	ProjectName string `gorm:"column:projectname;size:50" json:"projectname"`   // 项目名称
	Role        string `gorm:"column:role;size:50" json:"role"`                 // 项目角色
	Description string `gorm:"column:description;size:1000" json:"description"` // 项目描述
}

func (ResumeProject) TableName() string { return "ms_resume_project" }
