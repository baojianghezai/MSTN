package hrc

import (
	"context"
	"errors"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"gorm.io/gorm"
)

var (
	ErrPageNotFound    = errors.New("静态页不存在")
	ErrArticleNotFound = errors.New("内容不存在")
)

// CmsService 内容/配置服务（静态页、导航、分类、系统配置）
type CmsService struct{}

// GetPage 按 alias 读取静态页
func (s *CmsService) GetPage(ctx context.Context, alias string) (*hrcModel.Page, error) {
	var page hrcModel.Page
	err := global.GVA_DB.WithContext(ctx).Where("alias = ?", alias).First(&page).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPageNotFound
	}
	return &page, err
}

// ListArticles 内容列表（资讯/招聘会/帮助；仅展示中，sort 降序 + 时间倒序）
func (s *CmsService) ListArticles(ctx context.Context, typ int8, info request.PageInfo) ([]hrcModel.Article, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.Article{}).Where("display = 1")
	if typ > 0 {
		db = db.Where("type = ?", typ)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	var list []hrcModel.Article
	if err := db.Order("sort desc, addtime desc, id desc").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetArticle 内容详情（仅展示中；click 自增）
func (s *CmsService) GetArticle(ctx context.Context, id uint64) (*hrcModel.Article, error) {
	db := global.GVA_DB.WithContext(ctx)
	var article hrcModel.Article
	if err := db.Where("id = ? AND display = 1", id).First(&article).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrArticleNotFound
		}
		return nil, err
	}
	if err := db.Model(&hrcModel.Article{}).Where("id = ?", id).UpdateColumn("click", gorm.Expr("click + 1")).Error; err != nil {
		return nil, err
	}
	return &article, nil
}

// GetNavigations 前台导航（显示中的，按 sort 升序）
func (s *CmsService) GetNavigations(ctx context.Context) ([]hrcModel.Navigation, error) {
	var list []hrcModel.Navigation
	err := global.GVA_DB.WithContext(ctx).Where("display = 1").Order("sort asc").Find(&list).Error
	return list, err
}

// GetCategories 分类聚合：返回 group.alias -> []Category。地区目录单独按父级读取，
// 避免公共页面为 3241 条地区数据承担不必要的响应和渲染成本。
func (s *CmsService) GetCategories(ctx context.Context) (map[string][]hrcModel.Category, error) {
	var groups []hrcModel.CategoryGroup
	if err := global.GVA_DB.WithContext(ctx).Order("sort asc").Find(&groups).Error; err != nil {
		return nil, err
	}
	groupIndex := make(map[uint64]string, len(groups))
	categoryGroupIDs := make([]uint64, 0, len(groups))
	result := make(map[string][]hrcModel.Category, len(groups))
	for _, g := range groups {
		groupIndex[g.ID] = g.Alias
		result[g.Alias] = []hrcModel.Category{}
		if g.Alias != "district" {
			categoryGroupIDs = append(categoryGroupIDs, g.ID)
		}
	}
	if len(categoryGroupIDs) == 0 {
		return result, nil
	}

	var cats []hrcModel.Category
	if err := global.GVA_DB.WithContext(ctx).
		Where("display = 1 AND group_id IN ?", categoryGroupIDs).
		Order("sort asc").Find(&cats).Error; err != nil {
		return nil, err
	}
	for _, c := range cats {
		if alias, ok := groupIndex[c.GroupID]; ok {
			result[alias] = append(result[alias], c)
		}
	}
	return result, nil
}

// GetDistricts returns one district-tree level. parentID=0 yields provinces,
// municipalities, and autonomous regions; callers load children on demand.
func (s *CmsService) GetDistricts(ctx context.Context, parentID uint64) ([]hrcModel.Category, error) {
	var group hrcModel.CategoryGroup
	if err := global.GVA_DB.WithContext(ctx).Where("alias = ?", "district").First(&group).Error; err != nil {
		return nil, err
	}
	list := make([]hrcModel.Category, 0)
	err := global.GVA_DB.WithContext(ctx).
		Where("group_id = ? AND parent_id = ? AND display = 1", group.ID, parentID).
		Order("sort asc").Find(&list).Error
	return list, err
}

// GetConfigs 系统配置（按分组读取，group 为空返回全部）
func (s *CmsService) GetConfigs(ctx context.Context, group string) ([]hrcModel.Config, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.Config{})
	if group != "" {
		db = db.Where("cfg_group = ?", group)
	}
	var list []hrcModel.Config
	err := db.Order("id asc").Find(&list).Error
	return list, err
}

// ConfigKV 配置项键值（保存用）
type ConfigKV struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Remark string `json:"remark"`
}

// SaveConfigs 系统配置保存（按 name 幂等 upsert）
func (s *CmsService) SaveConfigs(ctx context.Context, group string, items []ConfigKV) error {
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, it := range items {
			if it.Name == "" {
				continue
			}
			var c hrcModel.Config
			err := tx.Where("name = ?", it.Name).First(&c).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(&hrcModel.Config{
					CfgGroup: group, Name: it.Name, Value: it.Value, Remark: it.Remark, AddTime: hrcModel.Now(),
				}).Error; err != nil {
					return err
				}
				continue
			}
			if err != nil {
				return err
			}
			if err := tx.Model(&hrcModel.Config{}).Where("name = ?", it.Name).
				Updates(map[string]interface{}{"value": it.Value, "remark": it.Remark}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
