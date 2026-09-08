package hrc

// CompanyCancellationApply 企业注销申请
// 设计依据：02_用户账号模块设计 §2.8、01_数据库设计 §1.2（v6 qs_company_cancellation_apply 平移）
// 语义：企业注销=清企业业务数据、会员账号（ms_members）保留可复用；区别于个人注销（账号级两阶段匿名化）
type CompanyCancellationApply struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UID         uint64 `gorm:"column:uid;index" json:"uid"`                    // 企业会员 uid
	CompanyID   uint64 `gorm:"column:company_id" json:"companyId"`             // 企业资料 id（ms_company_profile.id）
	CompanyName string `gorm:"column:companyname;size:100" json:"companyname"` // 企业名快照
	AddTime     int64  `gorm:"column:addtime" json:"addtime"`                  // 申请时间
	Status      int8   `gorm:"column:status;default:0" json:"status"`          // 0=待处理 1=已处理
	FinishTime  int64  `gorm:"column:finishtime;default:0" json:"finishtime"`  // 处理完成时间
}

func (CompanyCancellationApply) TableName() string { return "ms_company_cancellation_apply" }
