package hrc

import "time"

// WxpayLog 微信支付回调日志（对账/审计）
// 设计依据：11_v6差异分析 §四 P0#7（v6 qs_wxpay_log 平移；06 §3.3 对账配套）
// status：0=失败 1=成功；amount 沿用 v6 字符串结构（M4 支付回调写入）
type WxpayLog struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OpenID     string    `gorm:"column:openid;size:50" json:"openid"`           // 支付用户 openid
	TradeNo    string    `gorm:"column:trade_no;size:100;index" json:"tradeNo"` // 商户/微信交易号
	Amount     string    `gorm:"column:amount;size:30" json:"amount"`           // 金额（字符串，v6 原结构）
	AddTime    time.Time `gorm:"column:addtime" json:"addtime"`
	Status     int8      `gorm:"column:status" json:"status"` // 0=失败 1=成功
	FailReason string    `gorm:"column:fail_reason;size:255" json:"failReason"`
}

func (WxpayLog) TableName() string { return "ms_wxpay_log" }
