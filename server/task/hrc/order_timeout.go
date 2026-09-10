package hrc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/flipped-aurora/gin-vue-admin/server/task"
)

const orderExpireKeyPrefix = "hrc:order:expire:"

// RegisterOrderTimeout 注册订单超时关闭任务
// 逻辑：is_paid=1 AND created_at < now-30min → is_paid=3（原子关闭）
func RegisterOrderTimeout() {
	task.Register("HRC_OrderTimeout", "未支付订单超时关闭（is_paid=1 且超过30分钟 → is_paid=3）", orderTimeout)
}

func orderTimeout(ctx context.Context, _ json.RawMessage) error {
	timeoutAt := time.Now().Add(-30 * time.Minute).Unix()

	// 先查找需要关闭的订单 ID
	var expiredIDs []uint64
	if err := global.GVA_DB.WithContext(ctx).
		Model(&hrcModel.Order{}).
		Where("is_paid = 1 AND created_at > 0 AND created_at < ?", timeoutAt).
		Pluck("id", &expiredIDs).Error; err != nil {
		return err
	}

	if len(expiredIDs) == 0 {
		return nil
	}

	// 批量关闭订单
	if err := global.GVA_DB.WithContext(ctx).
		Model(&hrcModel.Order{}).
		Where("id IN ?", expiredIDs).
		Updates(map[string]interface{}{
			"is_paid":   3,
			"closed_at": time.Now().Unix(),
		}).Error; err != nil {
		return err
	}

	// 清理 Redis 中的过期时间
	for _, id := range expiredIDs {
		_ = global.GVA_REDIS.Del(ctx, orderExpireKeyPrefix+fmt.Sprintf("%d", id)).Err()
	}

	return nil
}
