package hrc

import "time"

// Page 静态页（关于我们/收费标准/联系我们等，alias 唯一）
// 设计依据：07_内容招聘会站内信模块设计 §2.5 / 09 §4.1 编号 43
type Page struct {
	ID       uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Alias    string    `gorm:"column:alias;size:60;uniqueIndex" json:"alias"` // 别名：about/contact/fee
	Title    string    `gorm:"column:title;size:100" json:"title"`
	Contents string    `gorm:"column:contents;type:text" json:"contents"`
	AddTime  time.Time `gorm:"column:add_time" json:"addTime"`
}

func (Page) TableName() string { return "ms_page" }

// Navigation 前台导航（07 §2.5 / 09 §4.1 编号 44）
type Navigation struct {
	ID      uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Title   string    `gorm:"column:title;size:60" json:"title"`
	URL     string    `gorm:"column:url;size:255" json:"url"`
	Sort    int       `gorm:"column:sort;default:0" json:"sort"`
	Display int8      `gorm:"column:display;default:1" json:"display"` // 1显示 2隐藏
	AddTime time.Time `gorm:"column:add_time" json:"addTime"`
}

func (Navigation) TableName() string { return "ms_navigation" }
