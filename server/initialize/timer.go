// server/initialize/timer.go
package initialize

import (
	"context"
	"encoding/json"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	mediaService "github.com/flipped-aurora/gin-vue-admin/server/service/media"
	"github.com/flipped-aurora/gin-vue-admin/server/task"
	hrcTask "github.com/flipped-aurora/gin-vue-admin/server/task/hrc"
)

// Timer 注册可供定时任务面板选用的命名方法。
// 不再硬编码调度: 调度由 sys_timed_tasks 表驱动(种子见 source/system/timed_task.go),
// 启动加载见 LoadTimedTasks。ctx 由统一 Runner 注入(已带 datascope.WithSystem 与超时)。
func Timer() {
	// GVA 系统任务
	task.Register("ClearDB", "清理数据库过期日志(操作记录/JWT黑名单/定时任务执行日志)", func(ctx context.Context, _ json.RawMessage) error {
		return task.ClearTable(global.GVA_DB.WithContext(ctx))
	})
	task.Register("CleanStaleUploads", "清理过期大文件上传会话", func(ctx context.Context, _ json.RawMessage) error {
		svc := mediaService.MediaUploadService{}
		return svc.CleanupStale(ctx, global.GVA_CONFIG.Media.SessionTTL)
	})

	// HRC 业务定时任务（11个）
	hrcTask.RegisterJobExpire()        // 职位过期下架
	hrcTask.RegisterSetmealExpire()    // 套餐到期清零 + 到期提醒
	hrcTask.RegisterOrderTimeout()     // 订单超时关闭
	hrcTask.RegisterPromotionExpire()  // 推广过期清理
	hrcTask.RegisterPaymentReconcile() // 支付对账
	hrcTask.RegisterDBBackup()         // 数据库备份
	hrcTask.RegisterLogCleanup()       // 日志清理
	hrcTask.RegisterStatsDaily()       // 统计日聚合
	hrcTask.RegisterWxpayLogCleanup()  // 微信支付日志清理
	hrcTask.RegisterAutoRefresh()      // 职位自动刷新
}
