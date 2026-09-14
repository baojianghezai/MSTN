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
	ErrAppealNotFound  = errors.New("申诉记录不存在")
	ErrRestoreNotFound = errors.New("未找到可恢复的注销账号（不在冷静期）")
)

// AdminService 后台业务服务（09 §4.6.2 hrc 后台接口）
type AdminService struct{}

// AppealList 申诉列表（分页 + 状态/手机号筛选）
func (s *AdminService) AppealList(ctx context.Context, info request.PageInfo, status int8, mobile string) (list []hrcModel.MembersAppeal, total int64, err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.MembersAppeal{})
	if status > 0 {
		db = db.Where("status = ?", status)
	}
	if mobile != "" {
		db = db.Where("mobile = ?", mobile)
	}
	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	err = db.Order("id desc").Limit(limit).Offset(offset).Find(&list).Error
	return list, total, err
}

// ProcessAppeal 处理申诉：更新状态；restore=true 且通过时恢复关联的注销账号
// 规则（02 §2.5）：恢复按申诉的 mobile 匹配 status=3 的账号 → status=1 + deleted_at=0
func (s *AdminService) ProcessAppeal(ctx context.Context, id uint64, status int8, restore bool) error {
	if status != 1 && status != 2 {
		return errors.New("处理状态非法（仅支持 1=已处理 2=已驳回）")
	}
	var appeal hrcModel.MembersAppeal
	if err := global.GVA_DB.WithContext(ctx).Where("id = ?", id).First(&appeal).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAppealNotFound
		}
		return err
	}
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&hrcModel.MembersAppeal{}).Where("id = ?", id).Update("status", status).Error; err != nil {
			return err
		}
		if restore && status == 1 {
			return restoreAccountByMobile(tx, appeal.Mobile)
		}
		return nil
	})
}

// restoreAccountByMobile 按手机号恢复注销账号（mobile 由 uk_mobile 唯一约束，冷静期内被注销账号占用）
func restoreAccountByMobile(tx *gorm.DB, mobile string) error {
	var member hrcModel.Members
	if err := tx.Where("mobile = ? AND status = 3", mobile).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRestoreNotFound
		}
		return err
	}
	return tx.Model(&hrcModel.Members{}).Where("uid = ?", member.UID).Updates(map[string]interface{}{
		"deleted_at": nil,
		"status":     1,
	}).Error
}
