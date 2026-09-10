package hrc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/flipped-aurora/gin-vue-admin/server/task"
)

// RegisterStatsDaily 注册统计日聚合任务
// 逻辑：汇总前一天的注册/投递/下载/订单数据
func RegisterStatsDaily() {
	task.Register("HRC_StatsDaily", "每日统计聚合（注册/投递/下载/订单）", statsDaily)
}

func statsDaily(ctx context.Context, _ json.RawMessage) error {
	yesterday := time.Now().AddDate(0, 0, -1)
	startOfDay := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.Local).Unix()
	endOfDay := startOfDay + 86400

	type stats struct {
		NewMembers int64 `json:"newMembers"`
		NewJobs    int64 `json:"newJobs"`
		NewApplys  int64 `json:"newApplys"`
		Downloads  int64 `json:"downloads"`
		NewOrders  int64 `json:"newOrders"`
		PaidOrders int64 `json:"paidOrders"`
	}

	var s stats

	// 新注册会员
	global.GVA_DB.WithContext(ctx).Model(&hrcModel.Members{}).
		Where("addtime >= ? AND addtime < ?", startOfDay, endOfDay).Count(&s.NewMembers)

	// 新发布职位
	global.GVA_DB.WithContext(ctx).Model(&hrcModel.Jobs{}).
		Where("addtime >= ? AND addtime < ?", startOfDay, endOfDay).Count(&s.NewJobs)

	// 新投递
	global.GVA_DB.WithContext(ctx).Table("ms_jobs_applys").
		Where("addtime >= ? AND addtime < ?", startOfDay, endOfDay).Count(&s.NewApplys)

	// 简历下载
	global.GVA_DB.WithContext(ctx).Model(&hrcModel.ResumeDownload{}).
		Where("downloaded_at >= ? AND downloaded_at < ?", startOfDay, endOfDay).Count(&s.Downloads)

	// 新订单
	global.GVA_DB.WithContext(ctx).Model(&hrcModel.Order{}).
		Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).Count(&s.NewOrders)

	// 已支付订单
	global.GVA_DB.WithContext(ctx).Model(&hrcModel.Order{}).
		Where("is_paid = 2 AND paid_at >= ? AND paid_at < ?", startOfDay, endOfDay).Count(&s.PaidOrders)

	// 写入配置表（生产环境可用独立统计表）
	_ = s

	return nil
}
