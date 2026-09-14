package hrc

import (
	"context"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/stretchr/testify/require"
)

func TestDataCleanupExpiredJobsAndHistory(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.JobsTmp{}, &hrcModel.JobPromotion{}, &hrcModel.MembersSetmeal{}, &hrcModel.WxpayLog{}, &hrcModel.PaymentNotifyLog{}, &hrcModel.DataCleanupLog{})
	now := time.Now()
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 1, JobsName: "过期职位", Deadline: now.Add(-60 * time.Second)}}).Error)
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 1, JobsName: "有效职位", Deadline: now.Add(60 * time.Second)}}).Error)

	svc := &DataCleanupService{}
	preview, err := svc.Preview(context.Background(), DataCleanupTargetExpiredJobs, 0)
	require.NoError(t, err)
	require.Equal(t, int64(1), preview.AffectedCount)
	require.True(t, preview.SoftDelete)

	result, err := svc.Execute(context.Background(), DataCleanupTargetExpiredJobs, 0, "CONFIRM", DataCleanupOperator{ID: 9, Name: "admin"})
	require.NoError(t, err)
	require.Equal(t, int64(1), result.AffectedCount)

	var expired hrcModel.Jobs
	require.NoError(t, db.Where("jobs_name = ?", "过期职位").First(&expired).Error)
	require.NotNil(t, expired.DeletedAt)
	var active hrcModel.Jobs
	require.NoError(t, db.Where("jobs_name = ?", "有效职位").First(&active).Error)
	require.Nil(t, active.DeletedAt)

	history, total, err := svc.ListHistory(context.Background(), request.PageInfo{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, history, 1)
	require.Equal(t, uint(9), history[0].OperatorID)
	require.Equal(t, DataCleanupTargetExpiredJobs, history[0].Target)
}

func TestDataCleanupDeletesOnlyAllowedHistoricalRecords(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.JobsTmp{}, &hrcModel.JobPromotion{}, &hrcModel.MembersSetmeal{}, &hrcModel.WxpayLog{}, &hrcModel.PaymentNotifyLog{}, &hrcModel.DataCleanupLog{})
	now := time.Now()
	require.NoError(t, db.Create(&hrcModel.JobsTmp{JobsBase: hrcModel.JobsBase{JobsName: "驳回草稿", Audit: 3}}).Error)
	require.NoError(t, db.Create(&hrcModel.JobsTmp{JobsBase: hrcModel.JobsBase{JobsName: "待审草稿", Audit: 2}}).Error)

	invalidJob := hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 3, JobsName: "下线职位", Display: 2, Audit: 1}}
	validJob := hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 4, JobsName: "在线职位", Display: 1, Audit: 1, Deadline: now.Add(3600 * time.Second)}}
	noEntitlementJob := hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 5, JobsName: "无套餐职位", Display: 1, Audit: 1, Deadline: now.Add(3600 * time.Second)}}
	require.NoError(t, db.Create(&invalidJob).Error)
	require.NoError(t, db.Create(&validJob).Error)
	require.NoError(t, db.Create(&noEntitlementJob).Error)
	require.NoError(t, db.Create(&hrcModel.MembersSetmeal{UID: 3, ExpireAt: now.Add(3600 * time.Second)}).Error)
	require.NoError(t, db.Create(&hrcModel.MembersSetmeal{UID: 4, ExpireAt: now.Add(3600 * time.Second)}).Error)
	require.NoError(t, db.Create(&hrcModel.JobPromotion{UID: 3, JobID: invalidJob.ID, Type: hrcModel.JobPromotionTypePush}).Error)
	require.NoError(t, db.Create(&hrcModel.JobPromotion{UID: 4, JobID: validJob.ID, Type: hrcModel.JobPromotionTypePush}).Error)
	require.NoError(t, db.Create(&hrcModel.JobPromotion{UID: 5, JobID: noEntitlementJob.ID, Type: hrcModel.JobPromotionTypePush}).Error)
	require.NoError(t, db.Create(&hrcModel.WxpayLog{TradeNo: "old", AddTime: now.Add(-90 * 24 * 3600 * time.Second)}).Error)
	require.NoError(t, db.Create(&hrcModel.WxpayLog{TradeNo: "new", AddTime: now.Add(-2 * 24 * 3600 * time.Second)}).Error)
	require.NoError(t, db.Create(&hrcModel.PaymentNotifyLog{OutTradeNo: "old", CreatedAt: now.Add(-90 * 24 * 3600 * time.Second)}).Error)
	require.NoError(t, db.Create(&hrcModel.PaymentNotifyLog{OutTradeNo: "new", CreatedAt: now.Add(-2 * 24 * 3600 * time.Second)}).Error)

	svc := &DataCleanupService{}
	for _, target := range []string{DataCleanupTargetRejectedJobDrafts, DataCleanupTargetExpiredPromotions, DataCleanupTargetOldWxpayLogs, DataCleanupTargetOldPaymentNotifyLogs} {
		_, err := svc.Execute(context.Background(), target, 30, "CONFIRM", DataCleanupOperator{})
		require.NoError(t, err)
	}

	var drafts, promotions, wxpayLogs, notifyLogs int64
	require.NoError(t, db.Model(&hrcModel.JobsTmp{}).Count(&drafts).Error)
	require.NoError(t, db.Model(&hrcModel.JobPromotion{}).Count(&promotions).Error)
	require.NoError(t, db.Model(&hrcModel.WxpayLog{}).Count(&wxpayLogs).Error)
	require.NoError(t, db.Model(&hrcModel.PaymentNotifyLog{}).Count(&notifyLogs).Error)
	require.Equal(t, int64(1), drafts)
	require.Equal(t, int64(1), promotions)
	require.Equal(t, int64(1), wxpayLogs)
	require.Equal(t, int64(1), notifyLogs)
}

func TestDataCleanupRejectsInvalidConfirmationAndRetention(t *testing.T) {
	testutil.NewMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.JobsTmp{}, &hrcModel.JobPromotion{}, &hrcModel.MembersSetmeal{}, &hrcModel.WxpayLog{}, &hrcModel.PaymentNotifyLog{}, &hrcModel.DataCleanupLog{})
	svc := &DataCleanupService{}
	_, err := svc.Preview(context.Background(), DataCleanupTargetOldWxpayLogs, 1)
	require.ErrorIs(t, err, ErrDataCleanupRetention)
	_, err = svc.Execute(context.Background(), DataCleanupTargetExpiredJobs, 0, "confirm", DataCleanupOperator{})
	require.ErrorIs(t, err, ErrDataCleanupConfirmation)
	_, err = svc.Preview(context.Background(), "members", 0)
	require.ErrorIs(t, err, ErrDataCleanupTargetInvalid)
}
