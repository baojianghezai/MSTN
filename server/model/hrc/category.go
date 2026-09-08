package hrc

// CategoryGroup 分类分组（trade/district/education/experience/wage/major/jobs）
// 设计依据：01_数据库设计 §1.3（ms_category_group）
type CategoryGroup struct {
	ID    uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Alias string `gorm:"column:alias;size:30;uniqueIndex" json:"alias"` // 分组标识
	Name  string `gorm:"column:name;size:60" json:"name"`               // 分组名（中文）
	Sort  int    `gorm:"column:sort;default:0" json:"sort"`
}

func (CategoryGroup) TableName() string { return "ms_category_group" }

// Category 分类值（group_id 关联分组，parent_id 支持层级）
// 设计依据：01_数据库设计 §1.3（ms_category，一期简化为通用分类表，地区/职位分类也走此表）
type Category struct {
	ID       uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	GroupID  uint64 `gorm:"column:group_id;index" json:"groupId"`
	ParentID uint64 `gorm:"column:parent_id;default:0" json:"parentId"`
	Name     string `gorm:"column:name;size:60" json:"name"`
	Sort     int    `gorm:"column:sort;default:0" json:"sort"`
	Display  int8   `gorm:"column:display;default:1" json:"display"` // 1显示 2隐藏
}

func (Category) TableName() string { return "ms_category" }
