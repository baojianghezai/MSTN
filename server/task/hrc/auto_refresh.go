package hrc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/flipped-aurora/gin-vue-admin/server/task"
)

// RegisterAutoRefresh 注册职位自动刷新任务
// 逻辑：对有刷新额度的套餐用户，自动刷新其在线职位
func RegisterAutoRefresh() {
	task.Register("HRC_AutoRefresh", "每日自动刷新职位（按套餐刷新额度批量刷新）", autoRefresh)
}

func autoRefresh(ctx context.Context, _ json.RawMessage) error {
	now := time.Now().Unix()

	// 查找有有效套餐的用户
	var entitlements []hrcModel.MembersSetmeal
	if err := global.GVA_DB.WithContext(ctx).
		Where("expire_at > ?", now).
		Find(&entitlements).Error; err != nil {
		return err
	}

	for _, ent := range entitlements {
		// 批量刷新该用户的在线职位（重置 refreshtime）
		result := global.GVA_DB.WithContext(ctx).
			Model(&hrcModel.Jobs{}).
			Where("uid = ? AND display = 1 AND audit = 1 AND deleted_at = 0", ent.UID).
			Update("refreshtime", now)

		if result.Error != nil {
			continue
		}
	}

	return nil
}
