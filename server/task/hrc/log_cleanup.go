package hrc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/task"
)

// RegisterLogCleanup 注册日志清理任务
// 逻辑：清理 90 天前的操作日志（GVA sys_operation_records + 业务 ms_members_log）
func RegisterLogCleanup() {
	task.Register("HRC_LogCleanup", "每周清理90天前的操作日志（sys_operation_records / ms_members_log）", logCleanup)
}

func logCleanup(ctx context.Context, _ json.RawMessage) error {
	cutoff := time.Now().AddDate(0, 0, -90)

	// GVA 操作日志：使用 created_at (time.Time)
	if err := global.GVA_DB.WithContext(ctx).
		Exec("DELETE FROM sys_operation_records WHERE created_at < ?", cutoff).Error; err != nil {
		return err
	}

	// 会员操作日志：使用 log_addtime (time.Time)
	if err := global.GVA_DB.WithContext(ctx).
		Exec("DELETE FROM ms_members_log WHERE log_addtime > ? AND log_addtime < ?", time.Time{}, cutoff).Error; err != nil {
		return err
	}

	return nil
}
