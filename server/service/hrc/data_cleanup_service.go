package hrc

import (
	"context"
	"errors"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"gorm.io/gorm"
)

const (
	DataCleanupTargetExpiredJobs          = "expired_jobs"
	DataCleanupTargetRejectedJobDrafts    = "rejected_job_drafts"
	DataCleanupTargetExpiredPromotions    = "expired_promotions"
	DataCleanupTargetOldPaymentNotifyLogs = "old_payment_notify_logs"
	DataCleanupTargetOldWxpayLogs         = "old_wxpay_logs"

	dataCleanupConfirmation = "CONFIRM"
	minRetentionDays        = 7
	maxRetentionDays        = 3650
)

var (
	ErrDataCleanupTargetInvalid = errors.New("清理目标非法")
	ErrDataCleanupRetention     = errors.New("保留天数需在 7 至 3650 天之间")
	ErrDataCleanupConfirmation  = errors.New("确认词错误")
)

// DataCleanupService keeps destructive operations to a small allowlist.
type DataCleanupService struct{}

type DataCleanupPreview struct {
	Target        string `json:"target"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	AffectedCount int64  `json:"affectedCount"`
	RetentionDays int    `json:"retentionDays"`
	CutoffAt      int64  `json:"cutoffAt"`
	SoftDelete    bool   `json:"softDelete"`
}

type DataCleanupOperator struct {
	ID   uint
	Name string
}

func (s *DataCleanupService) Preview(ctx context.Context, target string, retentionDays int) (DataCleanupPreview, error) {
	preview, err := buildDataCleanupPreview(target, retentionDays)
	if err != nil {
		return DataCleanupPreview{}, err
	}

	db := global.GVA_DB.WithContext(ctx)
	var count int64
	switch target {
	case DataCleanupTargetExpiredJobs:
		err = db.Model(&hrcModel.Jobs{}).
			Where("deleted_at = 0 AND deadline > 0 AND deadline < ?", preview.CutoffAt).
			Count(&count).Error
	case DataCleanupTargetRejectedJobDrafts:
		err = db.Model(&hrcModel.JobsTmp{}).Where("audit = ?", 3).Count(&count).Error
	case DataCleanupTargetExpiredPromotions:
		err = invalidPromotionsQuery(db, preview.CutoffAt).Count(&count).Error
	case DataCleanupTargetOldPaymentNotifyLogs:
		err = db.Model(&hrcModel.PaymentNotifyLog{}).Where("created_at < ?", preview.CutoffAt).Count(&count).Error
	case DataCleanupTargetOldWxpayLogs:
		err = db.Model(&hrcModel.WxpayLog{}).Where("addtime < ?", preview.CutoffAt).Count(&count).Error
	}
	if err != nil {
		return DataCleanupPreview{}, err
	}
	preview.AffectedCount = count
	return preview, nil
}

func (s *DataCleanupService) Execute(ctx context.Context, target string, retentionDays int, confirmation string, operator DataCleanupOperator) (DataCleanupPreview, error) {
	if confirmation != dataCleanupConfirmation {
		return DataCleanupPreview{}, ErrDataCleanupConfirmation
	}
	preview, err := buildDataCleanupPreview(target, retentionDays)
	if err != nil {
		return DataCleanupPreview{}, err
	}

	err = global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var result *gorm.DB
		switch target {
		case DataCleanupTargetExpiredJobs:
			result = tx.Model(&hrcModel.Jobs{}).
				Where("deleted_at = 0 AND deadline > 0 AND deadline < ?", preview.CutoffAt).
				Update("deleted_at", time.Now().Unix())
		case DataCleanupTargetRejectedJobDrafts:
			result = tx.Where("audit = ?", 3).Delete(&hrcModel.JobsTmp{})
		case DataCleanupTargetExpiredPromotions:
			result = invalidPromotionsQuery(tx, preview.CutoffAt).Delete(&hrcModel.JobPromotion{})
		case DataCleanupTargetOldPaymentNotifyLogs:
			result = tx.Where("created_at < ?", preview.CutoffAt).Delete(&hrcModel.PaymentNotifyLog{})
		case DataCleanupTargetOldWxpayLogs:
			result = tx.Where("addtime < ?", preview.CutoffAt).Delete(&hrcModel.WxpayLog{})
		}
		if result.Error != nil {
			return result.Error
		}
		preview.AffectedCount = result.RowsAffected
		return tx.Create(&hrcModel.DataCleanupLog{
			Target:        target,
			RetentionDays: preview.RetentionDays,
			CutoffAt:      preview.CutoffAt,
			AffectedCount: preview.AffectedCount,
			OperatorID:    operator.ID,
			OperatorName:  operator.Name,
			CreatedAt:     time.Now().Unix(),
		}).Error
	})
	if err != nil {
		return DataCleanupPreview{}, err
	}
	return preview, nil
}

func (s *DataCleanupService) ListHistory(ctx context.Context, info request.PageInfo) ([]hrcModel.DataCleanupLog, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.DataCleanupLog{})
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	var list []hrcModel.DataCleanupLog
	if err := db.Order("id desc").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func buildDataCleanupPreview(target string, retentionDays int) (DataCleanupPreview, error) {
	now := time.Now().Unix()
	preview := DataCleanupPreview{Target: target}
	switch target {
	case DataCleanupTargetExpiredJobs:
		preview.Title = "过期职位"
		preview.Description = "将截止时间早于当前时间且尚未删除的职位逻辑下线，不删除投递记录。"
		preview.CutoffAt = now
		preview.SoftDelete = true
	case DataCleanupTargetRejectedJobDrafts:
		preview.Title = "已驳回职位草稿"
		preview.Description = "物理删除审核状态为不通过的职位草稿；不处理共享 pid 的联系人和标签记录。"
		preview.CutoffAt = now
	case DataCleanupTargetExpiredPromotions:
		preview.Title = "失效首页推广"
		preview.Description = "删除关联职位已下线、已过期、未通过审核或企业套餐已失效的推广记录。"
		preview.CutoffAt = now
	case DataCleanupTargetOldPaymentNotifyLogs:
		if err := validateRetentionDays(retentionDays); err != nil {
			return DataCleanupPreview{}, err
		}
		preview.Title = "历史支付通知日志"
		preview.Description = "物理删除保留期限之前的支付通知审计日志，不影响订单和套餐权益。"
		preview.RetentionDays = retentionDays
		preview.CutoffAt = time.Now().AddDate(0, 0, -retentionDays).Unix()
	case DataCleanupTargetOldWxpayLogs:
		if err := validateRetentionDays(retentionDays); err != nil {
			return DataCleanupPreview{}, err
		}
		preview.Title = "历史微信支付日志"
		preview.Description = "物理删除保留期限之前的微信支付回调日志，不影响订单和套餐权益。"
		preview.RetentionDays = retentionDays
		preview.CutoffAt = time.Now().AddDate(0, 0, -retentionDays).Unix()
	default:
		return DataCleanupPreview{}, ErrDataCleanupTargetInvalid
	}
	return preview, nil
}

func validateRetentionDays(days int) error {
	if days < minRetentionDays || days > maxRetentionDays {
		return ErrDataCleanupRetention
	}
	return nil
}

func invalidPromotionsQuery(db *gorm.DB, now int64) *gorm.DB {
	invalidJobs := db.Session(&gorm.Session{NewDB: true}).Model(&hrcModel.Jobs{}).
		Select("id").Where("deleted_at <> 0 OR display <> 1 OR audit <> 1 OR (deadline > 0 AND deadline < ?)", now)
	validEntitlements := db.Session(&gorm.Session{NewDB: true}).Model(&hrcModel.MembersSetmeal{}).
		Select("uid").Where("expire_at > ?", now)
	return db.Model(&hrcModel.JobPromotion{}).Where("job_id IN (?) OR uid NOT IN (?)", invalidJobs, validEntitlements)
}
