package hrc

// PaymentNotifyLog records payment callback verification without retaining the
// callback body, which can contain payer and transaction information.
type PaymentNotifyLog struct {
	ID            uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Provider      string `gorm:"column:provider;size:20;index" json:"provider"`
	OrderID       uint64 `gorm:"column:order_id;index" json:"orderId"`
	OutTradeNo    string `gorm:"column:out_trade_no;size:40;index" json:"outTradeNo"`
	TransactionID string `gorm:"column:transaction_id;size:100;index" json:"transactionId"`
	PayloadSHA256 string `gorm:"column:payload_sha256;size:64" json:"payloadSha256"`
	Verified      bool   `gorm:"column:verified;index" json:"verified"`
	Message       string `gorm:"column:message;size:255" json:"message"`
	CreatedAt     int64  `gorm:"column:created_at;index" json:"createdAt"`
}

func (PaymentNotifyLog) TableName() string { return "ms_payment_notify_log" }
