package hrc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/task"
)

// RegisterWxpayLogCleanup 注册微信支付日志清理任务
// 逻辑：清理 180 天前的 ms_wxpay_log
func RegisterWxpayLogCleanup() {
	task.Register("HRC_WxpayLogCleanup", "每月清理180天前的微信支付回调日志（ms_wxpay_log）", wxpayLogCleanup)
}

func wxpayLogCleanup(ctx context.Context, _ json.RawMessage) error {
	cutoff := time.Now().AddDate(0, 0, -180).Unix()
	return global.GVA_DB.WithContext(ctx).
		Exec("DELETE FROM ms_wxpay_log WHERE addtime > 0 AND addtime < ?", cutoff).Error
}
