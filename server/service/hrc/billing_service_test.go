package hrc

import (
	"context"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newBillingMemoryDB(t testing.TB) *gorm.DB {
	return testutil.NewMemoryDB(t,
		&hrcModel.Setmeal{},
		&hrcModel.MembersSetmeal{},
		&hrcModel.Order{},
		&hrcModel.CompanyProfile{},
		&hrcModel.Jobs{},
		&hrcModel.JobsTmp{},
	)
}

func seedBillingPlan(t testing.TB, db *gorm.DB, display bool) hrcModel.Setmeal {
	t.Helper()
	plan := hrcModel.Setmeal{
		Name:            "Professional",
		Price:           29900,
		DurationDays:    90,
		JobsMeanwhile:   2,
		ResumeDownloads: 80,
		HomePushSlots:   2,
		HomeAdSlots:     1,
		EnableVideo:     true,
		Display:         display,
	}
	require.NoError(t, db.Create(&plan).Error)
	return plan
}

func TestCreateSetmealOrderRejectsHiddenPlan(t *testing.T) {
	db := newBillingMemoryDB(t)
	plan := seedBillingPlan(t, db, false)

	_, err := (&OrderService{}).CreateSetmealOrder(context.Background(), 101, plan.ID)
	require.ErrorIs(t, err, ErrSetmealNotFound)
}

func TestConfirmPaidGrantsEntitlementOnce(t *testing.T) {
	db := newBillingMemoryDB(t)
	plan := seedBillingPlan(t, db, true)
	companyName := "Acme"
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: 101, CompanyName: &companyName}).Error)

	svc := &OrderService{}
	order, err := svc.CreateSetmealOrder(context.Background(), 101, plan.ID)
	require.NoError(t, err)
	require.NoError(t, svc.ConfirmPaid(context.Background(), order.ID, plan.Price, "manual"))

	var paid hrcModel.Order
	require.NoError(t, db.First(&paid, order.ID).Error)
	require.Equal(t, int8(2), paid.IsPaid)
	require.Equal(t, plan.Price, paid.PayAmount)
	require.Equal(t, "manual", paid.Payment)
	require.NotZero(t, paid.PaidAt)

	var entitlement hrcModel.MembersSetmeal
	require.NoError(t, db.Where("uid = ?", 101).First(&entitlement).Error)
	require.Equal(t, plan.ID, entitlement.SetmealID)
	require.Equal(t, plan.JobsMeanwhile, entitlement.JobsMeanwhile)
	require.Equal(t, plan.ResumeDownloads, entitlement.ResumeDownloadsTotal)
	require.Equal(t, plan.HomePushSlots, entitlement.HomePushSlots)
	require.Equal(t, plan.HomeAdSlots, entitlement.HomeAdSlots)
	require.True(t, entitlement.EnableVideo)
	require.Greater(t, entitlement.ExpireAt, time.Now().Unix())

	var company hrcModel.CompanyProfile
	require.NoError(t, db.Where("uid = ?", 101).First(&company).Error)
	require.Equal(t, uint16(plan.ID), company.SetmealID)
	require.Equal(t, plan.Name, company.SetmealName)

	require.ErrorIs(t, svc.ConfirmPaid(context.Background(), order.ID, plan.Price, "manual"), ErrOrderPaid)
	var entitlementCount int64
	require.NoError(t, db.Model(&hrcModel.MembersSetmeal{}).Where("uid = ?", 101).Count(&entitlementCount).Error)
	require.Equal(t, int64(1), entitlementCount)
}

func TestClosedOrderCannotBeConfirmed(t *testing.T) {
	db := newBillingMemoryDB(t)
	plan := seedBillingPlan(t, db, true)
	svc := &OrderService{}
	order, err := svc.CreateSetmealOrder(context.Background(), 101, plan.ID)
	require.NoError(t, err)

	require.NoError(t, svc.Close(context.Background(), 101, order.ID))
	require.ErrorIs(t, svc.ConfirmPaid(context.Background(), order.ID, plan.Price, "manual"), ErrOrderClosed)

	var closed hrcModel.Order
	require.NoError(t, db.First(&closed, order.ID).Error)
	require.Equal(t, int8(3), closed.IsPaid)
}

func TestConfirmGatewayPaidIsIdempotentForSameTransaction(t *testing.T) {
	db := newBillingMemoryDB(t)
	plan := seedBillingPlan(t, db, true)
	companyName := "Acme"
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: 101, CompanyName: &companyName}).Error)

	svc := &OrderService{}
	order, err := svc.CreateSetmealOrder(context.Background(), 101, plan.ID)
	require.NoError(t, err)
	require.NoError(t, svc.ConfirmGatewayPaid(context.Background(), order.ID, plan.Price, PaymentProviderWechatNative, "wx-transaction-1"))
	require.NoError(t, svc.ConfirmGatewayPaid(context.Background(), order.ID, plan.Price, PaymentProviderWechatNative, "wx-transaction-1"))
	require.ErrorIs(t, svc.ConfirmGatewayPaid(context.Background(), order.ID, plan.Price, PaymentProviderWechatNative, "wx-transaction-2"), ErrOrderPaid)

	var paid hrcModel.Order
	require.NoError(t, db.First(&paid, order.ID).Error)
	require.NotNil(t, paid.TransactionID)
	require.Equal(t, "wx-transaction-1", *paid.TransactionID)

	var entitlementCount int64
	require.NoError(t, db.Model(&hrcModel.MembersSetmeal{}).Where("uid = ?", 101).Count(&entitlementCount).Error)
	require.Equal(t, int64(1), entitlementCount)
}

func TestConfirmGatewayPaidRejectsIncorrectAmountAndClosedOrder(t *testing.T) {
	db := newBillingMemoryDB(t)
	plan := seedBillingPlan(t, db, true)
	svc := &OrderService{}
	order, err := svc.CreateSetmealOrder(context.Background(), 101, plan.ID)
	require.NoError(t, err)
	require.Error(t, svc.ConfirmGatewayPaid(context.Background(), order.ID, plan.Price-1, PaymentProviderAlipayPage, "alipay-transaction-1"))

	require.NoError(t, svc.Close(context.Background(), 101, order.ID))
	require.ErrorIs(t, svc.ConfirmGatewayPaid(context.Background(), order.ID, plan.Price, PaymentProviderAlipayPage, "alipay-transaction-1"), ErrOrderClosed)
}

func TestApplyJobEntitlementEnforcesOnlineLimit(t *testing.T) {
	db := newBillingMemoryDB(t)
	now := time.Now().Unix()
	require.NoError(t, db.Create(&hrcModel.MembersSetmeal{
		UID:           101,
		SetmealID:     1,
		SetmealName:   "Starter",
		ExpireAt:      now + 86400,
		JobsMeanwhile: 1,
	}).Error)
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{
		UID:       101,
		Display:   1,
		DeletedAt: 0,
	}}).Error)

	err := (&SetmealService{}).ApplyJobEntitlement(context.Background(), 101, &hrcModel.Jobs{})
	require.ErrorIs(t, err, ErrSetmealJobLimit)
}
