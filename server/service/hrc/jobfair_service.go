package hrc

import (
	"context"
	"errors"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"gorm.io/gorm/clause"
)

var (
	ErrJobfairEntitlement   = errors.New("当前套餐不含举办招聘会权益")
	ErrJobfairNotFound      = errors.New("招聘会不存在")
	ErrJobfairInvalid       = errors.New("招聘会信息不完整（标题/举办时间/地点必填）")
	ErrJobfairAlreadySigned = errors.New("您已报名该招聘会")
	ErrJobfairNotSignup     = errors.New("报名记录不存在")
)

// JobfairInput 企业举办招聘会入参
type JobfairInput struct {
	Title     string
	Summary   string
	Cover     string
	Content   string
	HoldTime  string
	Address   string
	Organizer string
}

// JobfairService 招聘会服务（#2：企业有权益可举办，个人可自行参加）
type JobfairService struct{}

func (s *JobfairService) ensureEntitlement(ctx context.Context, uid uint64) error {
	current, err := (&SetmealService{}).Current(ctx, uid)
	if err != nil {
		return err
	}
	if current == nil || !current.EnableJobfair {
		return ErrJobfairEntitlement
	}
	return nil
}

// Create 企业举办招聘会（写入 ms_article type=2）
func (s *JobfairService) Create(ctx context.Context, uid uint64, in JobfairInput) (*hrcModel.Article, error) {
	if err := s.ensureEntitlement(ctx, uid); err != nil {
		return nil, err
	}
	if err := validateJobfairInput(in); err != nil {
		return nil, err
	}
	organizer := strings.TrimSpace(in.Organizer)
	if organizer == "" {
		organizer = (&ChatService{}).companyName(global.GVA_DB.WithContext(ctx), uid)
		if organizer == "企业" {
			organizer = ""
		}
	}
	now := hrcModel.Now()
	article := &hrcModel.Article{
		Type:       hrcModel.ArticleTypeJobfair,
		CompanyUID: uid,
		Title:      strings.TrimSpace(in.Title),
		Summary:    strings.TrimSpace(in.Summary),
		Cover:      strings.TrimSpace(in.Cover),
		Content:    strings.TrimSpace(in.Content),
		HoldTime:   strings.TrimSpace(in.HoldTime),
		Address:    strings.TrimSpace(in.Address),
		Organizer:  organizer,
		Display:    1,
		AddTime:    now,
		UpdateTime: now,
	}
	if err := global.GVA_DB.WithContext(ctx).Create(article).Error; err != nil {
		return nil, err
	}
	return article, nil
}

// Update 编辑本企业举办的招聘会
func (s *JobfairService) Update(ctx context.Context, uid uint64, id uint64, in JobfairInput) error {
	if err := validateJobfairInput(in); err != nil {
		return err
	}
	res := global.GVA_DB.WithContext(ctx).Model(&hrcModel.Article{}).
		Where("id = ? AND company_uid = ? AND type = ?", id, uid, hrcModel.ArticleTypeJobfair).
		Updates(map[string]interface{}{
			"title":       strings.TrimSpace(in.Title),
			"summary":     strings.TrimSpace(in.Summary),
			"cover":       strings.TrimSpace(in.Cover),
			"content":     strings.TrimSpace(in.Content),
			"hold_time":   strings.TrimSpace(in.HoldTime),
			"address":     strings.TrimSpace(in.Address),
			"update_time": hrcModel.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrJobfairNotFound
	}
	return nil
}

// Delete 下架本企业举办的招聘会（display=2）
func (s *JobfairService) Delete(ctx context.Context, uid uint64, id uint64) error {
	res := global.GVA_DB.WithContext(ctx).Model(&hrcModel.Article{}).
		Where("id = ? AND company_uid = ? AND type = ?", id, uid, hrcModel.ArticleTypeJobfair).
		Update("display", 2)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrJobfairNotFound
	}
	return nil
}

// ListMine 本企业举办的招聘会（含已下架）
func (s *JobfairService) ListMine(ctx context.Context, uid uint64) ([]hrcModel.Article, error) {
	var list []hrcModel.Article
	if err := global.GVA_DB.WithContext(ctx).
		Where("company_uid = ? AND type = ?", uid, hrcModel.ArticleTypeJobfair).
		Order("addtime desc, id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	s.fillSignupCounts(ctx, list)
	return list, nil
}

// Signup 个人报名参加招聘会（幂等：重复报名返回 ErrJobfairAlreadySigned）
func (s *JobfairService) Signup(ctx context.Context, uid uint64, jobfairID uint64) error {
	db := global.GVA_DB.WithContext(ctx)
	var count int64
	if err := db.Model(&hrcModel.Article{}).
		Where("id = ? AND type = ? AND display = 1", jobfairID, hrcModel.ArticleTypeJobfair).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return ErrJobfairNotFound
	}
	res := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&hrcModel.JobfairSignup{
		JobfairID:   jobfairID,
		PersonalUID: uid,
		AddTime:     hrcModel.Now(),
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrJobfairAlreadySigned
	}
	return nil
}

// Cancel 取消报名
func (s *JobfairService) Cancel(ctx context.Context, uid uint64, jobfairID uint64) error {
	res := global.GVA_DB.WithContext(ctx).
		Where("jobfair_id = ? AND personal_uid = ?", jobfairID, uid).
		Delete(&hrcModel.JobfairSignup{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrJobfairNotSignup
	}
	return nil
}

// MySignupIDs 我报名的招聘会 id 列表（列表页标记「已报名」）
func (s *JobfairService) MySignupIDs(ctx context.Context, uid uint64) ([]uint64, error) {
	ids := make([]uint64, 0)
	if err := global.GVA_DB.WithContext(ctx).Model(&hrcModel.JobfairSignup{}).
		Where("personal_uid = ?", uid).Pluck("jobfair_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// MyJobfairs 我参加的招聘会（按报名时间倒序分页）
func (s *JobfairService) MyJobfairs(ctx context.Context, uid uint64, info request.PageInfo) ([]hrcModel.Article, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Table("ms_jobfair_signup AS signup").
		Joins("JOIN ms_article AS article ON article.id = signup.jobfair_id").
		Where("signup.personal_uid = ? AND article.display = 1", uid)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	var list []hrcModel.Article
	if err := db.Select("article.*").Order("signup.id desc").Limit(limit).Offset(offset).Scan(&list).Error; err != nil {
		return nil, 0, err
	}
	s.fillSignupCounts(ctx, list)
	return list, total, nil
}

func (s *JobfairService) fillSignupCounts(ctx context.Context, list []hrcModel.Article) {
	if len(list) == 0 {
		return
	}
	ids := make([]uint64, 0, len(list))
	for _, a := range list {
		ids = append(ids, a.ID)
	}
	type countRow struct {
		JobfairID uint64
		Total     int
	}
	var rows []countRow
	_ = global.GVA_DB.WithContext(ctx).Model(&hrcModel.JobfairSignup{}).
		Select("jobfair_id, COUNT(*) AS total").
		Where("jobfair_id IN ?", ids).
		Group("jobfair_id").Scan(&rows).Error
	countMap := make(map[uint64]int, len(rows))
	for _, r := range rows {
		countMap[r.JobfairID] = r.Total
	}
	for i := range list {
		list[i].SignupCount = countMap[list[i].ID]
	}
}

func validateJobfairInput(in JobfairInput) error {
	if strings.TrimSpace(in.Title) == "" || strings.TrimSpace(in.HoldTime) == "" || strings.TrimSpace(in.Address) == "" {
		return ErrJobfairInvalid
	}
	return nil
}
