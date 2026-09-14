package hrc

import "time"

// JobsBase 职位公共字段（ms_jobs / ms_jobs_tmp 共用，01 §2.2 按 五点五 裁剪）
type JobsBase struct {
	UID             uint64     `gorm:"column:uid;index" json:"uid"` // 发布企业 uid
	JobsName        string     `gorm:"column:jobs_name;size:50;index" json:"jobsName"`
	CompanyName     string     `gorm:"column:companyname;size:50" json:"companyname"`
	CompanyID       uint64     `gorm:"column:company_id;index" json:"companyId"`
	Emergency       int8       `gorm:"column:emergency;default:0" json:"emergency"`
	Stick           int8       `gorm:"column:stick;default:0" json:"stick"`
	Nature          int        `gorm:"column:nature" json:"nature"` // 工作性质 code
	NatureCN        string     `gorm:"column:nature_cn;size:30" json:"natureCn"`
	Sex             int8       `gorm:"column:sex;default:3" json:"sex"`       // 性别要求 3不限
	Amount          uint16     `gorm:"column:amount" json:"amount"`           // 招聘人数
	TopClass        uint16     `gorm:"column:topclass" json:"topclass"`       // 一级分类
	Category        uint16     `gorm:"column:category;index" json:"category"` // 二级分类
	SubClass        uint16     `gorm:"column:subclass" json:"subclass"`       // 三级分类
	CategoryCN      string     `gorm:"column:category_cn;size:100" json:"categoryCn"`
	Trade           uint16     `gorm:"column:trade" json:"trade"`                // 行业
	District        string     `gorm:"column:district;size:100" json:"district"` // 地区 code
	DistrictCN      string     `gorm:"column:district_cn;size:100" json:"districtCn"`
	Tag             string     `gorm:"column:tag;size:50" json:"tag"`       // 标签（逗号分隔 code）
	Education       uint16     `gorm:"column:education" json:"education"`   // 学历要求
	Experience      uint16     `gorm:"column:experience" json:"experience"` // 经验要求
	MinWage         int        `gorm:"column:minwage" json:"minwage"`       // 薪资 元/月
	MaxWage         int        `gorm:"column:maxwage" json:"maxwage"`
	Negotiable      int8       `gorm:"column:negotiable;default:0" json:"negotiable"` // 面议
	Contents        string     `gorm:"column:contents;type:text" json:"contents"`     // 职位描述
	AddTime         time.Time  `gorm:"column:addtime" json:"addtime"`
	Deadline        time.Time  `gorm:"column:deadline" json:"deadline"` // 有效期
	Refreshtime     time.Time  `gorm:"column:refreshtime;index" json:"refreshtime"`
	SetmealDeadline time.Time  `gorm:"column:setmeal_deadline;default:0" json:"setmealDeadline"`
	SetmealID       uint16     `gorm:"column:setmeal_id" json:"setmealId"`
	SetmealName     string     `gorm:"column:setmeal_name;size:60" json:"setmealName"`
	Audit           int8       `gorm:"column:audit;default:0;index" json:"audit"` // 01 §3.1 权威枚举
	Display         int8       `gorm:"column:display;default:1" json:"display"`   // 1展示 2暂停
	Click           uint       `gorm:"column:click;default:1" json:"click"`
	UserStatus      int8       `gorm:"column:user_status;default:1" json:"userStatus"`
	AddMode         int8       `gorm:"column:add_mode;default:1" json:"addMode"`
	Department      string     `gorm:"column:department;size:60" json:"department"`
	MapX            float64    `gorm:"column:map_x;type:decimal(9,6)" json:"mapX"`
	MapY            float64    `gorm:"column:map_y;type:decimal(9,6)" json:"mapY"`
	MapZoom         int8       `gorm:"column:map_zoom" json:"mapZoom"`
	KeyPrecise      string     `gorm:"column:key_precise;type:text" json:"keyPrecise"`
	KeyFull         string     `gorm:"column:key_full;type:text" json:"keyFull"`
	DeletedAt       *time.Time `gorm:"column:deleted_at;index" json:"deletedAt"` // 逻辑删除（03 §3.1，保留投递记录）
	Logo            string     `gorm:"-" json:"logo"`                            // 来自 ms_company_profile，非持久化字段
}

// Jobs 职位主表
type Jobs struct {
	ID uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	JobsBase
}

func (Jobs) TableName() string { return "ms_jobs" }

// JobsTmp 职位编辑草稿（审核前，03 §2.1 双表；JobsID=0 表示新发布，>0 表示编辑原职位）
type JobsTmp struct {
	ID     uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	JobsID uint64 `gorm:"column:jobs_id;default:0" json:"jobsId"` // 原 ms_jobs.id（编辑场景）
	JobsBase
}

func (JobsTmp) TableName() string { return "ms_jobs_tmp" }

// JobsContact 职位联系方式（1:1，pid 关联职位 id，v6 qs_jobs_contact 平移）
type JobsContact struct {
	ID            uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PID           uint64 `gorm:"column:pid;index" json:"pid"`
	Contact       string `gorm:"column:contact;size:80" json:"contact"`
	QQ            string `gorm:"column:qq;size:20" json:"qq"`
	Telephone     string `gorm:"column:telephone;size:80" json:"telephone"`
	LandlineTel   string `gorm:"column:landline_tel;size:50" json:"landlineTel"`
	Address       string `gorm:"column:address;size:80" json:"address"`
	Email         string `gorm:"column:email;size:80" json:"email"`
	Notify        int8   `gorm:"column:notify" json:"notify"`
	NotifyMobile  int8   `gorm:"column:notify_mobile" json:"notifyMobile"`
	ContactShow   int8   `gorm:"column:contact_show;default:0" json:"contactShow"`
	TelephoneShow int8   `gorm:"column:telephone_show;default:0" json:"telephoneShow"`
	EmailShow     int8   `gorm:"column:email_show;default:0" json:"emailShow"`
	QQShow        int8   `gorm:"column:qq_show;default:0" json:"qqShow"`
	LandlineShow  int8   `gorm:"column:landline_tel_show" json:"landlineShow"`
}

func (JobsContact) TableName() string { return "ms_jobs_contact" }

// JobsTag 职位标签（N:1，pid 关联职位 id，v6 qs_jobs_tag 平移）
type JobsTag struct {
	ID  uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UID uint64 `gorm:"column:uid" json:"uid"`
	PID uint64 `gorm:"column:pid;index" json:"pid"`
	Tag uint32 `gorm:"column:tag;index" json:"tag"`
}

func (JobsTag) TableName() string { return "ms_jobs_tag" }
