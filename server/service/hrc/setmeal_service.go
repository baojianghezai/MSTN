package hrc

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"gorm.io/gorm"
)

var (
	ErrSetmealNotFound = errors.New("套餐不存在或已下架")
	ErrSetmealInvalid  = errors.New("套餐参数不合法")
	ErrSetmealJobLimit = errors.New("当前套餐的在线职位数已用完")
)

type SetmealService struct{}

func (s *SetmealService) PublicList(ctx context.Context) ([]hrcModel.Setmeal, error) {
	var list []hrcModel.Setmeal
	err := global.GVA_DB.WithContext(ctx).Where("display = ?", true).Order("sort desc, id asc").Find(&list).Error
	return list, err
}

func (s *SetmealService) AdminList(ctx context.Context) ([]hrcModel.Setmeal, error) {
	var list []hrcModel.Setmeal
	err := global.GVA_DB.WithContext(ctx).Order("sort desc, id asc").Find(&list).Error
	return list, err
}

func (s *SetmealService) Save(ctx context.Context, plan *hrcModel.Setmeal) error {
	plan.Name = strings.TrimSpace(plan.Name)
	if plan.Name == "" || plan.Price < 0 || plan.DurationDays <= 0 || plan.JobsMeanwhile < 0 || plan.ResumeDownloads < 0 || plan.HomePushSlots < 0 || plan.HomeAdSlots < 0 {
		return ErrSetmealInvalid
	}
	if plan.ID == 0 {
		return global.GVA_DB.WithContext(ctx).Create(plan).Error
	}
	return global.GVA_DB.WithContext(ctx).Save(plan).Error
}

func (s *SetmealService) Current(ctx context.Context, uid uint64) (*hrcModel.MembersSetmeal, error) {
	var current hrcModel.MembersSetmeal
	err := global.GVA_DB.WithContext(ctx).Where("uid = ?", uid).First(&current).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && current.ExpireAt <= time.Now().Unix()) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &current, nil
}

// ApplyJobEntitlement checks an active paid entitlement and stamps its snapshot
// onto a newly created job. Legacy companies without a purchase record remain
// available during the commercial rollout and are not silently blocked.
func (s *SetmealService) ApplyJobEntitlement(ctx context.Context, uid uint64, job *hrcModel.Jobs) error {
	current, err := s.Current(ctx, uid)
	if err != nil || current == nil {
		return err
	}
	if current.JobsMeanwhile > 0 {
		var jobsCount, pendingCount int64
		if err := global.GVA_DB.WithContext(ctx).Model(&hrcModel.Jobs{}).
			Where("uid = ? AND deleted_at = 0 AND display = ?", uid, 1).Count(&jobsCount).Error; err != nil {
			return err
		}
		if err := global.GVA_DB.WithContext(ctx).Model(&hrcModel.JobsTmp{}).
			Where("uid = ? AND deleted_at = 0 AND audit = ?", uid, 2).Count(&pendingCount).Error; err != nil {
			return err
		}
		if jobsCount+pendingCount >= int64(current.JobsMeanwhile) {
			return ErrSetmealJobLimit
		}
	}
	job.SetmealID = uint16(current.SetmealID)
	job.SetmealName = current.SetmealName
	job.SetmealDeadline = current.ExpireAt
	return nil
}
