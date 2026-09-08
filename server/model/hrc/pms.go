package hrc

// Pms 会员站内信。消息内容为纯文本，msg_check：0 未读，1 已读。
type Pms struct {
	ID       uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	MsgFrom  uint64 `gorm:"column:msgfrom;default:0;index" json:"msgFrom"`
	MsgTouid uint64 `gorm:"column:msgtouid;index:idx_pms_recipient_status" json:"msgTouid"`
	Title    string `gorm:"column:title;size:120" json:"title"`
	Message  string `gorm:"column:message;type:text" json:"message"`
	Type     string `gorm:"column:type;size:30;index" json:"type"`
	Link     string `gorm:"column:link;size:255" json:"link"`
	MsgCheck int8   `gorm:"column:msg_check;default:0;index:idx_pms_recipient_status" json:"msgCheck"`
	AddTime  int64  `gorm:"column:addtime;index" json:"addtime"`
}

func (Pms) TableName() string { return "ms_pms" }
