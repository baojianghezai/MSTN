package hrc

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
)

// PendingCounts 后台待处理数量（#3：职位审核 / 企业审核 / 待处理订单 / 待审推广）
type PendingCounts struct {
	JobAudit       int64 `json:"jobAudit"`
	CompanyAudit   int64 `json:"companyAudit"`
	OrderPending   int64 `json:"orderPending"`
	PromotionAudit int64 `json:"promotionAudit"`
	Total          int64 `json:"total"`
}

// PendingService 后台待办统计
type PendingService struct{}

// Counts 统计各后台模块待处理数量
func (s *PendingService) Counts(ctx context.Context) (*PendingCounts, error) {
	db := global.GVA_DB.WithContext(ctx)
	var out PendingCounts
	if err := db.Model(&hrcModel.JobsTmp{}).Where("audit = 2").Count(&out.JobAudit).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&hrcModel.CompanyProfile{}).Where("audit = 2").Count(&out.CompanyAudit).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&hrcModel.Order{}).Where("is_paid = 1").Count(&out.OrderPending).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&hrcModel.JobPromotion{}).Where("audit = 0").Count(&out.PromotionAudit).Error; err != nil {
		return nil, err
	}
	out.Total = out.JobAudit + out.CompanyAudit + out.OrderPending + out.PromotionAudit
	return &out, nil
}
