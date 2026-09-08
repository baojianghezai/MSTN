package hrc

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
)

// WxpayLogService 微信支付回调日志服务（11 §四 P0#7，对账/审计）
// Record 供 M4 支付回调（pay/notify/wxpay）调用；一期先建表 + 后台查询
type WxpayLogService struct{}

// Record 记录一条支付回调日志（M4 支付回调调用）
func (s *WxpayLogService) Record(ctx context.Context, log *hrcModel.WxpayLog) error {
	if log.AddTime == 0 {
		log.AddTime = hrcModel.Now()
	}
	return global.GVA_DB.WithContext(ctx).Create(log).Error
}

// List 支付日志列表（status：-1 全部 / 0 失败 / 1 成功）
func (s *WxpayLogService) List(ctx context.Context, info request.PageInfo, status int8) ([]hrcModel.WxpayLog, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.WxpayLog{})
	if status >= 0 {
		db = db.Where("status = ?", status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	var list []hrcModel.WxpayLog
	if err := db.Order("id desc").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
