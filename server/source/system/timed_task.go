package system

import (
	"context"

	sysModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const initOrderTimedTask = initOrderCasbin + 1

type initTimedTask struct{}

// auto run
func init() {
	system.RegisterInit(initOrderTimedTask, &initTimedTask{})
}

func (i *initTimedTask) InitializerName() string {
	return sysModel.SysTimedTask{}.TableName()
}

func (i *initTimedTask) MigrateTable(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	return ctx, db.AutoMigrate(&sysModel.SysTimedTask{}, &sysModel.SysTimedTaskLog{})
}

func (i *initTimedTask) TableCreated(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	m := db.Migrator()
	return m.HasTable(&sysModel.SysTimedTask{}) && m.HasTable(&sysModel.SysTimedTaskLog{})
}

// InitializeData 吃狗粮: 原 initialize/timer.go 两条硬编码任务迁为种子, 由面板接管
func (i *initTimedTask) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	entities := []sysModel.SysTimedTask{
		// GVA 系统任务
		{Name: "ClearDB", Description: "定时清理数据库过期日志(操作记录/JWT黑名单/定时任务执行日志)", Spec: "@daily", ExecutorType: sysModel.TimedTaskExecutorMethod, MethodName: "ClearDB", Enabled: true},
		{Name: "CleanStaleUploads", Description: "定时清理过期大文件上传会话", Spec: "@hourly", ExecutorType: sysModel.TimedTaskExecutorMethod, MethodName: "CleanStaleUploads", Enabled: true},
		// HRC 业务定时任务（11个）
		{Name: "HRC_JobExpire", Description: "职位过期自动下架（deadline < now → display=0）", Spec: "@daily", ExecutorType: sysModel.TimedTaskExecutorMethod, MethodName: "HRC_JobExpire", Enabled: true},
		{Name: "HRC_SetmealExpire", Description: "套餐到期清零权益 + 到期前7/3/1天发站内信提醒", Spec: "@daily", ExecutorType: sysModel.TimedTaskExecutorMethod, MethodName: "HRC_SetmealExpire", Enabled: true},
		{Name: "HRC_OrderTimeout", Description: "未支付订单超时关闭（is_paid=1 且超过30分钟 → is_paid=3）", Spec: "*/30 * * * *", ExecutorType: sysModel.TimedTaskExecutorMethod, MethodName: "HRC_OrderTimeout", Enabled: true},
		{Name: "HRC_PromotionExpire", Description: "推广过期清理（套餐到期 → 删除推广记录 + 复位stick/emergency）", Spec: "@daily", ExecutorType: sysModel.TimedTaskExecutorMethod, MethodName: "HRC_PromotionExpire", Enabled: true},
		{Name: "HRC_PaymentReconcile", Description: "每日支付对账（本地已支付订单 vs 第三方回调日志核对）", Spec: "@daily", ExecutorType: sysModel.TimedTaskExecutorMethod, MethodName: "HRC_PaymentReconcile", Enabled: true},
		{Name: "HRC_DBBackup", Description: "每日数据库备份（mysqldump，保留最近7天）", Spec: "@daily", ExecutorType: sysModel.TimedTaskExecutorMethod, MethodName: "HRC_DBBackup", Enabled: false},
		{Name: "HRC_LogCleanup", Description: "每周清理90天前的操作日志（sys_operation_records / ms_members_log）", Spec: "@weekly", ExecutorType: sysModel.TimedTaskExecutorMethod, MethodName: "HRC_LogCleanup", Enabled: true},
		{Name: "HRC_StatsDaily", Description: "每日统计聚合（注册/投递/下载/订单）", Spec: "@daily", ExecutorType: sysModel.TimedTaskExecutorMethod, MethodName: "HRC_StatsDaily", Enabled: true},
		{Name: "HRC_WxpayLogCleanup", Description: "每月清理180天前的微信支付回调日志（ms_wxpay_log）", Spec: "@monthly", ExecutorType: sysModel.TimedTaskExecutorMethod, MethodName: "HRC_WxpayLogCleanup", Enabled: true},
		{Name: "HRC_AutoRefresh", Description: "每日自动刷新职位（按套餐刷新额度批量刷新）", Spec: "@daily", ExecutorType: sysModel.TimedTaskExecutorMethod, MethodName: "HRC_AutoRefresh", Enabled: true},
	}
	if err := db.Create(&entities).Error; err != nil {
		return ctx, errors.Wrap(err, sysModel.SysTimedTask{}.TableName()+"表数据初始化失败!")
	}
	next := context.WithValue(ctx, i.InitializerName(), entities)
	return next, nil
}

func (i *initTimedTask) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	if errors.Is(db.Where("name = ?", "ClearDB").First(&sysModel.SysTimedTask{}).Error, gorm.ErrRecordNotFound) {
		return false
	}
	return true
}
