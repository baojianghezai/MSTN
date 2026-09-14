package hrc

import "time"

// Config 系统配置 KV（业务配置，分组管理：site/security/sms/payment）
// 设计依据：01_数据库设计 §1.1 / 09 §4.6.2 编号 150
type Config struct {
	ID       uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CfgGroup string    `gorm:"column:cfg_group;size:30;index" json:"cfgGroup"` // 分组
	Name     string    `gorm:"column:name;size:60;uniqueIndex" json:"name"`    // 配置键
	Value    string    `gorm:"column:value;type:text" json:"value"`            // 值（JSON 或字符串）
	Remark   string    `gorm:"column:remark;size:255" json:"remark"`
	AddTime  time.Time `gorm:"column:add_time" json:"addTime"`
}

func (Config) TableName() string { return "ms_config" }
