package hrc

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"gorm.io/gorm"
)

// JobsSearchService 职位前台搜索（03 §2.5，一期 MySQL LIKE + 索引）
type JobsSearchService struct{}

// JobsSearchFilter 职位筛选条件
type JobsSearchFilter struct {
	Keyword    string
	Trade      uint16
	Category   uint16
	District   string
	Education  uint16
	Experience uint16
	MinWage    int
	MaxWage    int
	Order      string // last/addtime/salary/stick
}

// JobPublicDetail 职位公开详情（前台，联系方式按可见性规则脱敏——一期直接返回）
type JobPublicDetail struct {
	hrcModel.JobsBase
	ID      uint64                `json:"id"`
	Contact *hrcModel.JobsContact `json:"contact"`
	Tags    []uint32              `json:"tags"`
}

// Search 职位列表（仅 display=1 & audit=1 & 未删除 & 未过期）
func (s *JobsSearchService) Search(ctx context.Context, f JobsSearchFilter, info request.PageInfo) ([]hrcModel.Jobs, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.Jobs{}).
		Where("display = 1 AND audit = 1 AND deleted_at = 0 AND (deadline = 0 OR deadline > ?)", time.Now().Unix())
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		db = db.Where("jobs_name LIKE ? OR companyname LIKE ?", kw, kw)
	}
	if f.Trade > 0 {
		db = db.Where("trade = ?", f.Trade)
	}
	if f.Category > 0 {
		db = db.Where("category = ?", f.Category)
	}
	if f.District != "" {
		db = db.Where("district = ?", f.District)
	}
	if f.Education > 0 {
		db = db.Where("education = ?", f.Education)
	}
	if f.Experience > 0 {
		db = db.Where("experience = ?", f.Experience)
	}
	if f.MinWage > 0 {
		db = db.Where("minwage >= ?", f.MinWage)
	}
	if f.MaxWage > 0 {
		db = db.Where("maxwage <= ?", f.MaxWage)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	switch f.Order {
	case "addtime":
		db = db.Order("addtime desc")
	case "salary":
		db = db.Order("maxwage desc, refreshtime desc")
	case "stick":
		db = db.Order("stick desc, refreshtime desc")
	default: // last
		db = db.Order("refreshtime desc")
	}
	var list []hrcModel.Jobs
	if err := db.Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Detail 职位详情（前台，click 自增）
func (s *JobsSearchService) Detail(ctx context.Context, id uint64) (*JobPublicDetail, error) {
	db := global.GVA_DB.WithContext(ctx)
	var j hrcModel.Jobs
	err := db.Where("id = ? AND display = 1 AND audit = 1 AND deleted_at = 0", id).First(&j).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrJobNotFound
	}
	if err != nil {
		return nil, err
	}
	// click 自增（异步批量回写一期简化为同步自增）
	if err := db.Model(&hrcModel.Jobs{}).Where("id = ?", id).UpdateColumn("click", gorm.Expr("click + 1")).Error; err != nil {
		return nil, err
	}

	detail := &JobPublicDetail{ID: j.ID, JobsBase: j.JobsBase}
	var contact hrcModel.JobsContact
	if err := db.Where("pid = ?", id).First(&contact).Error; err == nil {
		detail.Contact = &contact
	}
	var tags []hrcModel.JobsTag
	if err := db.Where("pid = ?", id).Find(&tags).Error; err != nil {
		return nil, err
	}
	detail.Tags = make([]uint32, 0, len(tags))
	for _, t := range tags {
		detail.Tags = append(detail.Tags, t.Tag)
	}
	return detail, nil
}

// HotWords 热门搜索词（一期读配置 mscms_hot_words，Redis Top50 二期）
func (s *JobsSearchService) HotWords(ctx context.Context) ([]string, error) {
	v, err := getConfigValue(ctx, "mscms_hot_words")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(v) == "" {
		return []string{}, nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out, nil
}
