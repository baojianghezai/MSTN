package hrc

// 简历五张结构化子表（M3 收尾批）
// 设计依据：04_简历模块设计 §2.4-2.8（v6 qs_resume_* 平移；education 的 campus_id 已剔除）
// 均无设计条数上限，一期不限（仅 project 限 6 条，见 resume.go ResumeProjectMax）
// 随简历 #48 创建 / #51 编辑 同事务全量替换（04 §3.1）

// ResumeEducation 教育经历
// 设计依据：04 §2.4（v6 qs_resume_education 平移，campus_id 剔除）
type ResumeEducation struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PID         uint64 `gorm:"column:pid;index" json:"pid"`                 // 简历主表 id（ms_resume.id）
	UID         uint64 `gorm:"column:uid;index" json:"uid"`                 // 简历所属会员 uid
	StartYear   uint16 `gorm:"column:startyear" json:"startyear"`           // 开始年
	StartMonth  uint8  `gorm:"column:startmonth" json:"startmonth"`         // 开始月
	EndYear     uint16 `gorm:"column:endyear" json:"endyear"`               // 结束年
	EndMonth    uint8  `gorm:"column:endmonth" json:"endmonth"`             // 结束月
	ToDate      int8   `gorm:"column:todate" json:"todate"`                 // 至今标记（0=已结束 1=至今）
	School      string `gorm:"column:school;size:50" json:"school"`         // 学校
	Speciality  string `gorm:"column:speciality;size:50" json:"speciality"` // 专业
	Education   uint16 `gorm:"column:education" json:"education"`           // 学历编码（枚举同主表 education）
	EducationCN string `gorm:"column:education_cn;size:30" json:"educationCn"`
}

func (ResumeEducation) TableName() string { return "ms_resume_education" }

// ResumeWork 工作经历
// 设计依据：04 §2.5（v6 qs_resume_work 平移）
type ResumeWork struct {
	ID           uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PID          uint64 `gorm:"column:pid;index" json:"pid"` // 简历主表 id
	UID          uint64 `gorm:"column:uid;index" json:"uid"` // 简历所属会员 uid
	StartYear    uint16 `gorm:"column:startyear" json:"startyear"`
	StartMonth   uint8  `gorm:"column:startmonth" json:"startmonth"`
	EndYear      uint16 `gorm:"column:endyear" json:"endyear"`
	EndMonth     uint8  `gorm:"column:endmonth" json:"endmonth"`
	ToDate       int8   `gorm:"column:todate" json:"todate"`                       // 至今标记（0=已结束 1=至今）
	WorkType     int8   `gorm:"column:work_type;default:1" json:"workType"`        // 1=工作 2=实习
	CompanyName  string `gorm:"column:companyname;size:50" json:"companyname"`     // 公司名称
	Jobs         string `gorm:"column:jobs;size:30" json:"jobs"`                   // 职位
	Achievements string `gorm:"column:achievements;size:1000" json:"achievements"` // 工作业绩
}

func (ResumeWork) TableName() string { return "ms_resume_work" }

// ResumeLanguage 语言能力
// 设计依据：04 §2.6（v6 qs_resume_language 平移）
type ResumeLanguage struct {
	ID         uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PID        uint64 `gorm:"column:pid;index" json:"pid"`     // 简历主表 id
	UID        uint64 `gorm:"column:uid;index" json:"uid"`     // 简历所属会员 uid
	Language   uint16 `gorm:"column:language" json:"language"` // 语言编码（枚举字典）
	LanguageCN string `gorm:"column:language_cn;size:50" json:"languageCn"`
	Level      uint16 `gorm:"column:level" json:"level"` // 等级编码（枚举字典）
	LevelCN    string `gorm:"column:level_cn;size:50" json:"levelCn"`
}

func (ResumeLanguage) TableName() string { return "ms_resume_language" }

// ResumeTraining 培训经历
// 设计依据：04 §2.7（v6 qs_resume_training 平移）
type ResumeTraining struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PID         uint64 `gorm:"column:pid;index" json:"pid"` // 简历主表 id
	UID         uint64 `gorm:"column:uid;index" json:"uid"` // 简历所属会员 uid
	StartYear   uint16 `gorm:"column:startyear" json:"startyear"`
	StartMonth  uint8  `gorm:"column:startmonth" json:"startmonth"`
	EndYear     uint16 `gorm:"column:endyear" json:"endyear"`
	EndMonth    uint8  `gorm:"column:endmonth" json:"endmonth"`
	ToDate      int8   `gorm:"column:todate" json:"todate"`                     // 至今标记（0=已结束 1=至今）
	Agency      string `gorm:"column:agency;size:50" json:"agency"`             // 培训机构
	Course      string `gorm:"column:course;size:50" json:"course"`             // 课程名称
	Description string `gorm:"column:description;size:1000" json:"description"` // 培训描述
}

func (ResumeTraining) TableName() string { return "ms_resume_training" }

// ResumeCredent 证书
// 设计依据：04 §2.8（v6 qs_resume_credent 平移；images 一期保留字段、前端不做上传 UI）
type ResumeCredent struct {
	ID     uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PID    uint64 `gorm:"column:pid;index" json:"pid"`          // 简历主表 id
	UID    uint64 `gorm:"column:uid;index" json:"uid"`          // 简历所属会员 uid
	Name   string `gorm:"column:name;size:255" json:"name"`     // 证书名称
	Year   uint16 `gorm:"column:year" json:"year"`              // 获得年份
	Month  uint8  `gorm:"column:month" json:"month"`            // 获得月份
	Images string `gorm:"column:images;size:255" json:"images"` // 证书图片 URL（一期预留，可空）
}

func (ResumeCredent) TableName() string { return "ms_resume_credent" }

// ResumeSkill 专业技能（可选额外项）
// 设计依据：11 新增——用户按需添加的技能标签
type ResumeSkill struct {
	ID    uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PID   uint64 `gorm:"column:pid;index" json:"pid"`     // 简历主表 id
	UID   uint64 `gorm:"column:uid;index" json:"uid"`     // 简历所属会员 uid
	Name  string `gorm:"column:name;size:50" json:"name"` // 技能名称
	Level uint8  `gorm:"column:level" json:"level"`       // 熟练度 1=入门 2=熟练 3=精通
}

func (ResumeSkill) TableName() string { return "ms_resume_skill" }

// ResumePortfolio 个人作品（可选额外项）
// 设计依据：11 新增——作品集链接/描述
type ResumePortfolio struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PID         uint64 `gorm:"column:pid;index" json:"pid"`                     // 简历主表 id
	UID         uint64 `gorm:"column:uid;index" json:"uid"`                     // 简历所属会员 uid
	Title       string `gorm:"column:title;size:100" json:"title"`              // 作品标题
	Description string `gorm:"column:description;size:1000" json:"description"` // 作品描述
	URL         string `gorm:"column:url;size:255" json:"url"`                  // 作品链接
}

func (ResumePortfolio) TableName() string { return "ms_resume_portfolio" }

// ResumeStudentLeader 学生干部经历（可选额外项）
// 设计依据：11 新增——学生组织/社团经历
type ResumeStudentLeader struct {
	ID           uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PID          uint64 `gorm:"column:pid;index" json:"pid"`                      // 简历主表 id
	UID          uint64 `gorm:"column:uid;index" json:"uid"`                      // 简历所属会员 uid
	Organization string `gorm:"column:organization;size:100" json:"organization"` // 组织名称
	Role         string `gorm:"column:role;size:50" json:"role"`                  // 担任职务
	StartYear    uint16 `gorm:"column:startyear" json:"startyear"`                // 开始年
	StartMonth   uint8  `gorm:"column:startmonth" json:"startmonth"`              // 开始月
	EndYear      uint16 `gorm:"column:endyear" json:"endyear"`                    // 结束年
	EndMonth     uint8  `gorm:"column:endmonth" json:"endmonth"`                  // 结束月
	ToDate       int8   `gorm:"column:todate" json:"todate"`                      // 至今标记（0=已结束 1=至今）
	Description  string `gorm:"column:description;size:1000" json:"description"`  // 经历描述
}

func (ResumeStudentLeader) TableName() string { return "ms_resume_student_leader" }
