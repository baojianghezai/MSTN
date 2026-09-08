package hrc

// 审核类型（ms_audit_reason.type 区分审核对象）
const (
	AuditTypeJobs    int8 = 1 // 职位
	AuditTypeResume  int8 = 2 // 简历
	AuditTypeCompany int8 = 3 // 企业资质
	AuditTypePhoto   int8 = 4 // 简历照片
)

// AuditReason 审核不通过原因（通用表，按 type+type_id 关联）
// 设计依据：01_数据库设计 §1.4/§3.1（后台审核 reason 写入 ms_audit_reason，不通过原因供企业端回显）
type AuditReason struct {
	ID      uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Type    int8   `gorm:"column:type" json:"type"`            // 审核类型：1职位 2简历 3企业资质 4简历照片
	TypeID  uint64 `gorm:"column:type_id;index" json:"typeId"` // 关联记录 id
	Reason  string `gorm:"column:reason;size:255" json:"reason"`
	AddTime int64  `gorm:"column:addtime" json:"addtime"`
}

func (AuditReason) TableName() string { return "ms_audit_reason" }
