package hrc

// DataCleanupLog records every completed bulk cleanup run for administrator audit.
type DataCleanupLog struct {
	ID            uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Target        string `gorm:"column:target;size:40;index" json:"target"`
	RetentionDays int    `gorm:"column:retention_days;default:0" json:"retentionDays"`
	CutoffAt      int64  `gorm:"column:cutoff_at" json:"cutoffAt"`
	AffectedCount int64  `gorm:"column:affected_count" json:"affectedCount"`
	OperatorID    uint   `gorm:"column:operator_id;index" json:"operatorId"`
	OperatorName  string `gorm:"column:operator_name;size:60" json:"operatorName"`
	CreatedAt     int64  `gorm:"column:created_at;index" json:"createdAt"`
}

func (DataCleanupLog) TableName() string { return "ms_data_cleanup_log" }
