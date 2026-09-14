package hrc

import (
	"time"

	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
)

// ---- 认证模块 DTO（对应 09_API接口规范 §4.1）----

type SmsCodeRequest struct {
	Mobile string `json:"mobile" binding:"required"`              // 手机号
	Type   string `json:"type" enums:"register,login,reset,bind"` // 验证码类型
}

type RegisterRequest struct {
	Utype    int8   `json:"utype" enums:"1,2"` // 1=个人 2=企业
	Mobile   string `json:"mobile" binding:"required"`
	Code     string `json:"code" binding:"required"` // 短信验证码
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Account  string `json:"account" binding:"required"` // 手机号/邮箱/用户名
	Password string `json:"password" binding:"required"`
}

type LoginSmsRequest struct {
	Mobile string `json:"mobile" binding:"required"`
	Code   string `json:"code" binding:"required"`
	Utype  int8   `json:"utype" enums:"0,1,2"` // 0=未指定(以库为准/默认个人) 1=个人 2=企业
}

type LoginWechatRequest struct {
	Code    string `json:"code" binding:"required"`  // 微信 code
	Channel string `json:"channel" enums:"pc,h5,mp"` // 渠道：pc/h5/小程序
}

type WechatBindRequest struct {
	BindToken string `json:"bindToken" binding:"required"`
	Mobile    string `json:"mobile" binding:"required"`
	Code      string `json:"code" binding:"required"`
}

type WechatMpBindRequest struct {
	BindToken     string `json:"bindToken" binding:"required"`
	Code          string `json:"code" binding:"required"`
	EncryptedData string `json:"encryptedData" binding:"required"`
	IV            string `json:"iv" binding:"required"`
}

// AuthTokenData 登录/注册返回
type AuthTokenData struct {
	Token       string `json:"token"` // access token
	UID         uint64 `json:"uid"`
	Utype       int8   `json:"utype"`
	Mobile      string `json:"mobile"`      // 当前登录手机号（解绑后为 unbound_<uid> 占位）
	PasswordSet bool   `json:"passwordSet"` // 是否已设置密码（false 时前端引导走 #13 忘记密码重置设密）
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type ResetPasswordRequest struct {
	Mobile      string `json:"mobile" binding:"required"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type BindMobileRequest struct {
	Mobile string `json:"mobile" binding:"required"`
	Code   string `json:"code" binding:"required"`
}

type CancelRequest struct {
	Code string `json:"code" binding:"required"`
}

// ---- 账号申诉 DTO（09 §4.1 编号 45/46）----

type AppealSubmitRequest struct {
	RealName    string `json:"realname" binding:"required"`
	Mobile      string `json:"mobile" binding:"required"`
	Email       string `json:"email"`
	Description string `json:"description" binding:"required"`
}

type AppealProcessRequest struct {
	Status  int8 `json:"status"`  // 1=已处理 2=已驳回
	Restore bool `json:"restore"` // 是否恢复账号（仅 status=1 时生效）
}

// AppealStatusItem 申诉进度项（statusCn 由后端补充中文）
type AppealStatusItem struct {
	ID          uint64    `json:"id"`
	UID         uint64    `json:"uid"`
	RealName    string    `json:"realname"`
	Mobile      string    `json:"mobile"`
	Email       string    `json:"email"`
	Description string    `json:"description"`
	AddTime     time.Time `json:"addtime"`
	Status      int8      `json:"status"`
	StatusCN    string    `json:"statusCn"`
}

// ---- 个人/企业资料 DTO（09 §4.1 编号 76/77、107-110）----

type PersonalProfileRequest struct {
	RealName     string `json:"realname"`
	Sex          int8   `json:"sex"`
	SexCN        string `json:"sexCn"`
	Birthday     int64  `json:"birthday"`
	Residence    string `json:"residence"`
	Education    uint16 `json:"education"`
	EducationCN  string `json:"educationCn"`
	Major        uint16 `json:"major"`
	MajorCN      string `json:"majorCn"`
	Experience   uint16 `json:"experience"`
	ExperienceCN string `json:"experienceCn"`
	Phone        string `json:"phone"`
	Height       string `json:"height"`
	Marriage     int8   `json:"marriage"`
	MarriageCN   string `json:"marriageCn"`
	DisplayName  int8   `json:"displayName"`
	QQ           string `json:"qq"`
	Weixin       string `json:"weixin"`
}

type CompanyProfileRequest struct {
	CompanyName    string `json:"companyname"`
	Nature         uint16 `json:"nature"`
	NatureCN       string `json:"natureCn"`
	Trade          uint16 `json:"trade"`
	TradeCN        string `json:"tradeCn"`
	District       string `json:"district"`
	DistrictCN     string `json:"districtCn"`
	Scale          uint16 `json:"scale"`
	ScaleCN        string `json:"scaleCn"`
	Registered     string `json:"registered"`
	Address        string `json:"address"`
	Contact        string `json:"contact"`
	Telephone      string `json:"telephone"`
	LandlineTel    string `json:"landlineTel"`
	Email          string `json:"email"`
	Website        string `json:"website"`
	CertificateImg string `json:"certificateImg"`
	Logo           string `json:"logo"`
	Contents       string `json:"contents"`
	Tag            string `json:"tag"`
	ShortName      string `json:"shortName"`
	ShortDesc      string `json:"shortDesc"`
}

// CompanyAuditData 企业审核状态返回
type CompanyAuditData struct {
	Audit   int8   `json:"audit"`
	AuditCN string `json:"auditCn"`
}

// ---- 企业注销 DTO（02 §2.8：110a/110b、161a-c）----

// CompanyCancelRequest 企业注销申请（短信二次确认，type=cancellation）
type CompanyCancelRequest struct {
	Code string `json:"code" binding:"required"`
}

// CompanyCancellationStatusData 企业注销申请状态（企业端 110b）
type CompanyCancellationStatusData struct {
	ID          uint64 `json:"id"`
	CompanyName string `json:"companyname"`
	AddTime     int64  `json:"addtime"`
	Status      int8   `json:"status"`   // 0=待处理 1=已处理
	StatusCN    string `json:"statusCn"` // 待处理/已处理
	FinishTime  int64  `json:"finishtime"`
}

// CompanyCancellationAdminItem 后台注销申请列表项（161a，含会员 username/mobile）
type CompanyCancellationAdminItem struct {
	ID          uint64 `json:"id"`
	UID         uint64 `json:"uid"`
	CompanyID   uint64 `json:"companyId"`
	CompanyName string `json:"companyname"`
	Username    string `json:"username"`
	Mobile      string `json:"mobile"`
	AddTime     int64  `json:"addtime"`
	Status      int8   `json:"status"`
	StatusCN    string `json:"statusCn"`
	FinishTime  int64  `json:"finishtime"`
}

// ---- 企业资质审核 DTO（09 §4.6.2 #134/#134a/#135）----

// CompanyProfileAdminItem 后台企业资料列表项
type CompanyProfileAdminItem struct {
	ID          uint64    `json:"id"`
	UID         uint64    `json:"uid"`
	CompanyName string    `json:"companyname"`
	Logo        string    `json:"logo"`
	Audit       int8      `json:"audit"`   // 0=未提交 1=通过 2=待审 3=不通过
	AuditCN     string    `json:"auditCn"` // 未提交/已通过/审核中/未通过
	AddTime     time.Time `json:"addtime"`
	Refreshtime time.Time `json:"refreshtime"`
}

// CompanyProfileAdminDetail 后台企业资料详情（含 Logo/营业执照证照，审核用）
type CompanyProfileAdminDetail struct {
	hrcModel.CompanyProfile
	AuditCN string `json:"auditCn"`
}

// CompanyProfileAuditRequest 企业资质审核请求
type CompanyProfileAuditRequest struct {
	Audit  int8   `json:"audit"`  // 1=通过 3=不通过
	Reason string `json:"reason"` // 不通过原因（audit=3 时写入 ms_audit_reason）
}

// ---- 内容/配置 DTO（09 §4.1 编号 43/44、§4.6.2 编号 150）----

type ConfigItem struct {
	Name   string `json:"name" binding:"required"`
	Value  string `json:"value"`
	Remark string `json:"remark"`
}

type SaveConfigsRequest struct {
	Group string       `json:"group"`
	Items []ConfigItem `json:"items" binding:"required"`
}

// ---- 简历 DTO（09 §4.2 编号 48/50/51）----

// ResumeProjectRequest 项目经历子表项（随简历创建/编辑批量提交）
type ResumeProjectRequest struct {
	StartYear   uint16 `json:"startyear"`   // 开始年
	StartMonth  uint8  `json:"startmonth"`  // 开始月
	EndYear     uint16 `json:"endyear"`     // 结束年
	EndMonth    uint8  `json:"endmonth"`    // 结束月
	ToDate      int8   `json:"todate"`      // 至今标记（0=已结束 1=至今）
	ProjectName string `json:"projectname"` // 项目名称
	Role        string `json:"role"`        // 项目角色
	Description string `json:"description"` // 项目描述
}

// ResumeEducationRequest 教育经历子表项（随简历创建/编辑批量提交；04 §2.4）
type ResumeEducationRequest struct {
	StartYear   uint16 `json:"startyear"`
	StartMonth  uint8  `json:"startmonth"`
	EndYear     uint16 `json:"endyear"`
	EndMonth    uint8  `json:"endmonth"`
	ToDate      int8   `json:"todate"` // 至今标记（0=已结束 1=至今）
	School      string `json:"school"`
	Speciality  string `json:"speciality"`
	Education   uint16 `json:"education"`
	EducationCN string `json:"educationCn"`
}

// ResumeWorkRequest 工作经历子表项（04 §2.5）
type ResumeWorkRequest struct {
	StartYear    uint16 `json:"startyear"`
	StartMonth   uint8  `json:"startmonth"`
	EndYear      uint16 `json:"endyear"`
	EndMonth     uint8  `json:"endmonth"`
	ToDate       int8   `json:"todate"` // 至今标记（0=已结束 1=至今）
	CompanyName  string `json:"companyname"`
	Jobs         string `json:"jobs"`
	Achievements string `json:"achievements"`
}

// ResumeLanguageRequest 语言能力子表项（04 §2.6）
type ResumeLanguageRequest struct {
	Language   uint16 `json:"language"`
	LanguageCN string `json:"languageCn"`
	Level      uint16 `json:"level"`
	LevelCN    string `json:"levelCn"`
}

// ResumeTrainingRequest 培训经历子表项（04 §2.7）
type ResumeTrainingRequest struct {
	StartYear   uint16 `json:"startyear"`
	StartMonth  uint8  `json:"startmonth"`
	EndYear     uint16 `json:"endyear"`
	EndMonth    uint8  `json:"endmonth"`
	ToDate      int8   `json:"todate"` // 至今标记（0=已结束 1=至今）
	Agency      string `json:"agency"`
	Course      string `json:"course"`
	Description string `json:"description"`
}

// ResumeCredentRequest 证书子表项（04 §2.8；images 一期预留、前端不做上传 UI）
type ResumeCredentRequest struct {
	Name   string `json:"name"`
	Year   uint16 `json:"year"`
	Month  uint8  `json:"month"`
	Images string `json:"images"`
}

// ResumeRequest 创建/编辑简历请求（主表可编辑字段 + 6 子表：项目经历限 6 条，其余一期不限）
type ResumeRequest struct {
	Title         string                   `json:"title"`
	FullName      string                   `json:"fullname"`
	Sex           int8                     `json:"sex"`
	SexCN         string                   `json:"sexCn"`
	Birthdate     uint16                   `json:"birthdate"`
	Residence     string                   `json:"residence"`
	Education     uint16                   `json:"education"`
	EducationCN   string                   `json:"educationCn"`
	Major         uint16                   `json:"major"`
	MajorCN       string                   `json:"majorCn"`
	Experience    uint16                   `json:"experience"`
	ExperienceCN  string                   `json:"experienceCn"`
	District      string                   `json:"district"`
	DistrictCN    string                   `json:"districtCn"`
	Wage          uint16                   `json:"wage"`
	WageCN        string                   `json:"wageCn"`
	IntentionJobs string                   `json:"intentionJobs"`
	Specialty     string                   `json:"specialty"`
	Telephone     string                   `json:"telephone"`
	Email         string                   `json:"email"`
	DisplayName   int8                     `json:"displayName"`
	Current       uint16                   `json:"current"`
	CurrentCN     string                   `json:"currentCn"`
	MobileAudit   int8                     `json:"mobileAudit"`
	Talent        int8                     `json:"talent"`
	Entrust       int8                     `json:"entrust"`
	Projects      []ResumeProjectRequest   `json:"projects"`   // 项目经历子表（限 6 条）
	Educations    []ResumeEducationRequest `json:"educations"` // 教育经历子表（键名复数，避免与主表学历编码 education 冲突）
	Work          []ResumeWorkRequest      `json:"work"`       // 工作经历子表
	Language      []ResumeLanguageRequest  `json:"language"`   // 语言能力子表
	Training      []ResumeTrainingRequest  `json:"training"`   // 培训经历子表
	Credent       []ResumeCredentRequest   `json:"credent"`    // 证书子表
}

// ResumeDetailData 简历详情返回（编辑回显：主表字段平铺 + 6 子表数组，子表按 id 升序）
// 注意：主表学历编码仍是 education（uint16）；教育经历子表数组键名为 educations（复数，避免冲突）
type ResumeDetailData struct {
	hrcModel.Resume
	Projects   []hrcModel.ResumeProject   `json:"projects"`
	Educations []hrcModel.ResumeEducation `json:"educations"`
	Work       []hrcModel.ResumeWork      `json:"work"`
	Language   []hrcModel.ResumeLanguage  `json:"language"`
	Training   []hrcModel.ResumeTraining  `json:"training"`
	Credent    []hrcModel.ResumeCredent   `json:"credent"`
}

// ResumeLite 我的简历列表项（#49 轻量主表字段，不返子表）
type ResumeLite struct {
	ID              uint64    `json:"id"`
	Title           string    `json:"title"`
	CompletePercent int8      `json:"completePercent"`
	Def             int8      `json:"def"`
	Display         int8      `json:"display"`
	AddTime         time.Time `json:"addtime"`
	Refreshtime     time.Time `json:"refreshtime"`
}

// ResumeDisplayRequest 公开/隐藏切换请求（#54）
type ResumeDisplayRequest struct {
	Display int8 `json:"display" binding:"required,oneof=1 2"` // 1=公开 2=不公开
}

// ---- 视频面试 DTO（11 §四 P0#2，v6 qs_video_interview 平移）----

// VideoInterviewCreateRequest 企业发起视频面试邀请
type VideoInterviewCreateRequest struct {
	ResumeID      uint64 `json:"resumeId" binding:"required"`      // 简历 id
	JobsID        uint64 `json:"jobsId" binding:"required"`        // 职位 id
	JobsName      string `json:"jobsName" binding:"required"`      // 职位名快照（ms_jobs 未建，前端传入）
	InterviewTime int64  `json:"interviewTime" binding:"required"` // 面试时间（unix 秒）
	Contact       string `json:"contact" binding:"required"`       // 联系人
	Telephone     string `json:"telephone" binding:"required"`     // 联系电话
}

// VideoInterviewItem 视频面试列表/详情项
type VideoInterviewItem struct {
	ID            uint64 `json:"id"`
	CompanyUID    uint64 `json:"companyUid"`
	PersonalUID   uint64 `json:"personalUid"`
	JobsID        uint64 `json:"jobsId"`
	JobsName      string `json:"jobsName"`
	InterviewTime int64  `json:"interviewTime"`
	Deadline      int64  `json:"deadline"`
	Contact       string `json:"contact"`
	ContactTel    string `json:"contactTel"`
	AddTime       int64  `json:"addtime"`
	CompanyCode   string `json:"companyCode"`
	PersonalCode  string `json:"personalCode"`
	RoomStatus    string `json:"roomStatus"` // nostart / opened / overtime
	ResumeID      uint64 `json:"resumeId"`
	FullName      string `json:"fullname"`    // 简历姓名（公司端/后台）
	CompanyName   string `json:"companyname"` // 企业名（后台）
}

// VideoInterviewRoomData 房间码查询返回（TRTC 入房用，不含联系方式）
type VideoInterviewRoomData struct {
	ID            uint64 `json:"id"`
	JobsName      string `json:"jobsName"`
	InterviewTime int64  `json:"interviewTime"`
	Deadline      int64  `json:"deadline"`
	RoomStatus    string `json:"roomStatus"` // nostart / opened / overtime
	Utype         int8   `json:"utype"`      // 1=个人端 2=企业端
}

// ---- 后台导出 DTO（11 §四 P0#5）----

// ExportRequest 导出请求（按 id 列表导出）
type ExportRequest struct {
	IDS []uint64 `json:"ids" binding:"required"`
}

// ---- 职位 DTO（03 模块，09 §4.3 编号 79-86 / §4.6.2 编号 121-123）----

// JobsContactRequest 职位联系方式
type JobsContactRequest struct {
	Contact     string `json:"contact"`
	QQ          string `json:"qq"`
	Telephone   string `json:"telephone"`
	LandlineTel string `json:"landlineTel"`
	Address     string `json:"address"`
	Email       string `json:"email"`
}

// JobsRequest 发布/编辑职位请求
type JobsRequest struct {
	JobsName   string             `json:"jobsName"`
	Nature     int                `json:"nature"`
	NatureCN   string             `json:"natureCn"`
	Sex        int8               `json:"sex"`
	Amount     uint16             `json:"amount"`
	TopClass   uint16             `json:"topclass"`
	Category   uint16             `json:"category"`
	SubClass   uint16             `json:"subclass"`
	CategoryCN string             `json:"categoryCn"`
	Trade      uint16             `json:"trade"`
	District   string             `json:"district"`
	DistrictCN string             `json:"districtCn"`
	Tag        string             `json:"tag"`
	Education  uint16             `json:"education"`
	Experience uint16             `json:"experience"`
	MinWage    int                `json:"minwage"`
	MaxWage    int                `json:"maxwage"`
	Negotiable int8               `json:"negotiable"`
	Contents   string             `json:"contents"`
	Deadline   int64              `json:"deadline"`
	Department string             `json:"department"`
	MapX       float64            `json:"mapX"`
	MapY       float64            `json:"mapY"`
	MapZoom    int8               `json:"mapZoom"`
	Contact    JobsContactRequest `json:"contact"`
	Tags       []uint32           `json:"tags"`
}

// JobsAuditRequest 职位审核请求（#123）
type JobsAuditRequest struct {
	Audit  int8   `json:"audit"`  // 1=通过 3=不通过
	Reason string `json:"reason"` // 不通过原因
}

// ---- 投递 DTO（05 模块，09 §4.2 编号 61-63）----

// ApplyRequest 投递职位请求（#61）
type ApplyRequest struct {
	JobsIDs  []uint64 `json:"jobsIds" binding:"required"` // 职位 id 数组
	ResumeID uint64   `json:"resumeId"`                   // 简历 id（0=默认简历）
	Notes    string   `json:"notes"`                      // 求职附言
}

// CompanyApplyReplyRequest 企业回复状态请求（#94）
type CompanyApplyReplyRequest struct {
	IsReply int8 `json:"isReply"` // 0待反馈 1合适 2不合适 3待定 4未接通
}
