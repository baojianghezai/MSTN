package hrc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/flipped-aurora/gin-vue-admin/server/task"
)

// RegisterJobExpire 注册职位过期任务
// 逻辑：deadline < now AND display != 0 → display=0（暂停展示）
func RegisterJobExpire() {
	task.Register("HRC_JobExpire", "职位过期自动下架（deadline < now → display=0）", jobExpire)
}

func jobExpire(ctx context.Context, _ json.RawMessage) error {
	now := time.Now().Unix()
	result := global.GVA_DB.WithContext(ctx).
		Model(&hrcModel.Jobs{}).
		Where("deadline > 0 AND deadline < ? AND display != 0", now).
		Update("display", 0)
	return result.Error
}
