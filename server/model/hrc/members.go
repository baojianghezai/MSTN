package hrc

import "time"

// Members 会员主表（utype: 1=个人 2=企业）
// 设计依据：01_数据库设计 §2.1（密码 bcrypt，新增 deleted_at）
type Members struct {
	UID            uint64 `gorm:"column:uid;primaryKey;autoIncrement"`
	Utype          int8   `gorm:"column:utype;default:1"`              // 1个人 2企业
	Username       string `gorm:"column:username;size:60;uniqueIndex"` // 唯一
	Email          string `gorm:"column:email;size:80"`
	EmailAudit     int8   `gorm:"column:email_audit;default:0"`
	Mobile         string `gorm:"column:mobile;size:11;uniqueIndex"` // 唯一；注销时置 NULL（§六.2）
	MobileAudit    int8   `gorm:"column:mobile_audit;default:0"`
	Password       string `gorm:"column:password;size:100"` // bcrypt hash
	RegTime        int64  `gorm:"column:reg_time"`
	RegIP          string `gorm:"column:reg_ip;size:15"`
	RegAddress     string `gorm:"column:reg_address;size:30"`
	LastLoginTime  int64  `gorm:"column:last_login_time"`
	LastLoginIP    string `gorm:"column:last_login_ip;size:15"`
	Status         int8   `gorm:"column:status;default:1"` // 1正常 2暂停 3注销
	Avatars        string `gorm:"column:avatars;size:255"`
	Consultant     uint16 `gorm:"column:consultant;default:0"`
	SmsNum         int    `gorm:"column:sms_num;default:0"`
	RegType        int8   `gorm:"column:reg_type;default:0"`
	InvitationCode string `gorm:"column:invitation_code;size:8"`
	DeletedAt      int64  `gorm:"column:deleted_at;default:0"` // 软删除（注销合规）
}

func (Members) TableName() string { return "ms_members" }

// MembersInfo 会员详情（个人）
// 设计依据：01_数据库设计 §1.2
type MembersInfo struct {
	ID           uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UID          uint64 `gorm:"column:uid;index" json:"uid"`
	RealName     string `gorm:"column:realname;size:60" json:"realname"`
	Sex          int8   `gorm:"column:sex" json:"sex"`
	SexCN        string `gorm:"column:sex_cn;size:30" json:"sexCn"`
	Birthday     int64  `gorm:"column:birthday" json:"birthday"`
	Residence    string `gorm:"column:residence;size:60" json:"residence"`
	Education    uint16 `gorm:"column:education" json:"education"`
	EducationCN  string `gorm:"column:education_cn;size:30" json:"educationCn"`
	Major        uint16 `gorm:"column:major" json:"major"`
	MajorCN      string `gorm:"column:major_cn;size:30" json:"majorCn"`
	Experience   uint16 `gorm:"column:experience" json:"experience"`
	ExperienceCN string `gorm:"column:experience_cn;size:30" json:"experienceCn"`
	Phone        string `gorm:"column:phone;size:20" json:"phone"`
	Height       string `gorm:"column:height;size:5" json:"height"`
	Marriage     int8   `gorm:"column:marriage" json:"marriage"`
	MarriageCN   string `gorm:"column:marriage_cn;size:5" json:"marriageCn"`
	DisplayName  int8   `gorm:"column:display_name;default:1" json:"displayName"`
	QQ           string `gorm:"column:qq;size:30" json:"qq"`
	Weixin       string `gorm:"column:weixin;size:30" json:"weixin"`
}

func (MembersInfo) TableName() string { return "ms_members_info" }

// CompanyProfile 企业资料
// 设计依据：01_数据库设计 §1.2；companyname 唯一（uk_companyname，M12 决议）
type CompanyProfile struct {
	ID             uint64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UID            uint64  `gorm:"column:uid;index" json:"uid"`
	CompanyName    *string `gorm:"column:companyname;size:60;uniqueIndex" json:"companyname"` // 唯一；空壳为 NULL（唯一索引允许多 NULL）
	Nature         uint16  `gorm:"column:nature" json:"nature"`
	NatureCN       string  `gorm:"column:nature_cn;size:30" json:"natureCn"`
	Trade          uint16  `gorm:"column:trade" json:"trade"`
	TradeCN        string  `gorm:"column:trade_cn;size:30" json:"tradeCn"`
	District       string  `gorm:"column:district;size:100" json:"district"`
	DistrictCN     string  `gorm:"column:district_cn;size:100" json:"districtCn"`
	Scale          uint16  `gorm:"column:scale" json:"scale"`
	ScaleCN        string  `gorm:"column:scale_cn;size:30" json:"scaleCn"`
	Registered     string  `gorm:"column:registered;size:150" json:"registered"` // 注册资金
	Address        string  `gorm:"column:address;size:250" json:"address"`
	Contact        string  `gorm:"column:contact;size:100" json:"contact"`
	Telephone      string  `gorm:"column:telephone;size:130" json:"telephone"`
	LandlineTel    string  `gorm:"column:landline_tel;size:50" json:"landlineTel"`
	Email          string  `gorm:"column:email;size:100" json:"email"`
	Website        string  `gorm:"column:website;size:100" json:"website"`
	CertificateImg string  `gorm:"column:certificate_img;size:255" json:"certificateImg"` // 营业执照
	Logo           string  `gorm:"column:logo;size:255" json:"logo"`
	Contents       string  `gorm:"column:contents;type:text" json:"contents"` // 企业介绍
	SetmealID      uint16  `gorm:"column:setmeal_id" json:"setmealId"`
	SetmealName    string  `gorm:"column:setmeal_name;size:30" json:"setmealName"`
	Audit          int8    `gorm:"column:audit;default:0;index" json:"audit"` // 资质审核（权威枚举 01 §3.1）
	AddTime        int64   `gorm:"column:addtime" json:"addtime"`
	Refreshtime    int64   `gorm:"column:refreshtime" json:"refreshtime"`
	Click          uint    `gorm:"column:click;default:1" json:"click"`
	UserStatus     int8    `gorm:"column:user_status;default:1" json:"userStatus"`
	Tag            string  `gorm:"column:tag;size:60" json:"tag"`
	ShortName      string  `gorm:"column:short_name;size:60" json:"shortName"`
	ShortDesc      string  `gorm:"column:short_desc;size:255" json:"shortDesc"`
}

func (CompanyProfile) TableName() string { return "ms_company_profile" }

// MembersBind 会员绑定（微信等）
type MembersBind struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	UID         uint64 `gorm:"column:uid;index"`
	Type        string `gorm:"column:type;size:20"`   // wechat
	Keyid       string `gorm:"column:keyid;size:100"` // unionid/openid
	Info        string `gorm:"column:info;type:text"`
	Bindingtime int64  `gorm:"column:bindingtime"`
}

func (MembersBind) TableName() string { return "ms_members_bind" }

// MembersAppeal 账号申诉（M11 决议：新增 uid 可空字段）
type MembersAppeal struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UID         uint64 `gorm:"column:uid;default:0" json:"uid"` // 可空：已登录自动带，未登录按手机号匹配
	RealName    string `gorm:"column:realname;size:60" json:"realname"`
	Mobile      string `gorm:"column:mobile;size:11" json:"mobile"`
	Email       string `gorm:"column:email;size:80" json:"email"`
	Description string `gorm:"column:description;type:text" json:"description"`
	AddTime     int64  `gorm:"column:addtime" json:"addtime"`
	Status      int8   `gorm:"column:status;default:0" json:"status"` // 0待处理 1已处理 2已驳回
}

func (MembersAppeal) TableName() string { return "ms_members_appeal" }

// MembersLog 会员操作日志
type MembersLog struct {
	LogID       uint64 `gorm:"column:log_id;primaryKey;autoIncrement"`
	LogUID      uint64 `gorm:"column:log_uid;index"`
	LogUsername string `gorm:"column:log_username;size:60"`
	LogAddtime  int64  `gorm:"column:log_addtime"`
	LogValue    string `gorm:"column:log_value;type:text"`
	LogIP       string `gorm:"column:log_ip;size:15"`
	LogAddress  string `gorm:"column:log_address;size:30"`
	LogUtype    int8   `gorm:"column:log_utype;default:1"`
	LogType     int8   `gorm:"column:log_type;default:1"`
}

func (MembersLog) TableName() string { return "ms_members_log" }

// MembersMsgtip 消息提醒计数
type MembersMsgtip struct {
	UID        uint64 `gorm:"column:uid;primaryKey"`
	Type       string `gorm:"column:type;size:20;primaryKey"`
	UpdateTime int64  `gorm:"column:update_time"`
	Unread     int    `gorm:"column:unread;default:0"`
}

func (MembersMsgtip) TableName() string { return "ms_members_msgtip" }

// Oauth 第三方登录配置
type Oauth struct {
	ID         uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	Alias      string `gorm:"column:alias;size:30"`
	Name       string `gorm:"column:name;size:60"`
	Info       string `gorm:"column:info;type:text"`
	Config     string `gorm:"column:config;type:text"`
	Apply      int8   `gorm:"column:apply;default:0"`
	CreateTime int64  `gorm:"column:create_time"`
	OrdID      int    `gorm:"column:ordid;default:0"`
	Status     int8   `gorm:"column:status;default:1"`
}

func (Oauth) TableName() string { return "ms_oauth" }

// UnbindMobile 解绑手机记录
type UnbindMobile struct {
	ID       uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	UID      uint64 `gorm:"column:uid;index"`
	Utype    int8   `gorm:"column:utype;default:1"`
	Username string `gorm:"column:username;size:60"`
	Mobile   string `gorm:"column:mobile;size:11"`
	AddTime  int64  `gorm:"column:add_time"`
	Remark   string `gorm:"column:remark;size:255"`
}

func (UnbindMobile) TableName() string { return "ms_unbind_mobile" }

// Now 统一时间戳工具
func Now() int64 {
	return time.Now().Unix()
}
