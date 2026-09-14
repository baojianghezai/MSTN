package hrc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/flipped-aurora/gin-vue-admin/server/task"
)

// RegisterSetmealExpire 注册套餐到期任务
// 1. 套餐到期 → 清零套餐权益
// 2. 到期前 7/3/1 天发站内信提醒
func RegisterSetmealExpire() {
	task.Register("HRC_SetmealExpire", "套餐到期清零权益 + 到期前7/3/1天发站内信提醒", setmealExpire)
}

func setmealExpire(ctx context.Context, _ json.RawMessage) error {
	now := time.Now()

	// 1. 到期 → 清零权益
	result := global.GVA_DB.WithContext(ctx).
		Model(&hrcModel.MembersSetmeal{}).
		Where("expire_at > ? AND expire_at < ?", time.Time{}, now).
		Updates(map[string]interface{}{
			"jobs_meanway":           0,
			"resume_downloads_total": 0,
			"home_push_slots":        0,
			"home_ad_slots":          0,
		})
	if result.Error != nil {
		return result.Error
	}

	// 2. 到期提醒：7/3/1 天前发站内信
	days := []int{7, 3, 1}
	for _, d := range days {
		startOfDay := now.AddDate(0, 0, d)
		startOfDay = time.Date(startOfDay.Year(), startOfDay.Month(), startOfDay.Day(), 0, 0, 0, 0, time.Local)
		endOfDay := startOfDay.Add(24 * time.Hour)

		var expiring []hrcModel.MembersSetmeal
		if err := global.GVA_DB.WithContext(ctx).
			Where("expire_at > ? AND expire_at < ?", startOfDay, endOfDay).
			Find(&expiring).Error; err != nil {
			return err
		}

		for _, m := range expiring {
			msg := hrcModel.Pms{
				MsgTouid: m.UID,
				Title:    "套餐到期提醒",
				Message:  "您的套餐将在 " + m.ExpireAt.Format("2006-01-02") + " 到期，请及时续费。",
				Type:     "system",
				AddTime:  now,
			}
			global.GVA_DB.WithContext(ctx).Create(&msg)
		}
	}

	return nil
}
