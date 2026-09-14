package hrc

import "time"

// Order uses a small state machine: 1=pending payment, 2=paid, 3=closed.
// Payment channel is deliberately a business label in this first release;
// gateway-specific fields will be filled by a later payment adapter.
type Order struct {
	ID               uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OID              string     `gorm:"column:oid;size:40;uniqueIndex" json:"oid"`
	UID              uint64     `gorm:"column:uid;index" json:"uid"`
	OrderType        int8       `gorm:"column:order_type;default:1" json:"orderType"`
	SetmealID        uint64     `gorm:"column:setmeal_id;index" json:"setmealId"`
	SetmealName      string     `gorm:"column:setmeal_name;size:60" json:"setmealName"`
	Amount           int64      `gorm:"column:amount" json:"amount"`
	PayAmount        int64      `gorm:"column:pay_amount;default:0" json:"payAmount"`
	Payment          string     `gorm:"column:payment;size:30" json:"payment"`
	TransactionID    *string    `gorm:"column:transaction_id;size:100;uniqueIndex" json:"transactionId"`
	PaymentStartedAt *time.Time `gorm:"column:payment_started_at" json:"paymentStartedAt"`
	IsPaid           int8       `gorm:"column:is_paid;default:1;index" json:"isPaid"`
	CreatedAt        time.Time  `gorm:"column:created_at;index" json:"createdAt"`
	PaidAt           *time.Time `gorm:"column:paid_at" json:"paidAt"`
	ClosedAt         *time.Time `gorm:"column:closed_at" json:"closedAt"`
}

func (Order) TableName() string { return "ms_order" }
