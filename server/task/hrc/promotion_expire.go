package hrc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/flipped-aurora/gin-vue-admin/server/task"
)

// RegisterPromotionExpire 注册推广过期任务
// 1. 删除套餐已过期的推广记录
// 2. 重置 stick/emergency 标记（如有）
func RegisterPromotionExpire() {
	task.Register("HRC_PromotionExpire", "推广过期清理（套餐到期 → 删除推广记录 + 复位stick/emergency）", promotionExpire)
}

func promotionExpire(ctx context.Context, _ json.RawMessage) error {
	now := time.Now().Unix()

	// 1. 删除套餐已过期的推广记录
	if err := global.GVA_DB.WithContext(ctx).
		Exec("DELETE FROM ms_job_promotion WHERE uid IN (SELECT uid FROM ms_members_setmeal WHERE expire_at > 0 AND expire_at < ?)", now).Error; err != nil {
		return err
	}

	// 2. 重置 stick/emergency：将已过期职位的 stick/emergency 复位
	if err := global.GVA_DB.WithContext(ctx).
		Model(&hrcModel.Jobs{}).
		Where("(deadline > 0 AND deadline < ?) AND (stick != 0 OR emergency != 0)", now).
		Updates(map[string]interface{}{
			"stick":     0,
			"emergency": 0,
		}).Error; err != nil {
		return err
	}

	return nil
}
