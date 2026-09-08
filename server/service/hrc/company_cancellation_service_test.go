package hrc

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func randomMobile(prefix string) string {
	return fmt.Sprintf("%s%08d", prefix, rand.Intn(100000000))
}

// setupCancelTest 内存库 + nop logger + Redis（无 Redis 自动 skip）
func setupCancelTest(t *testing.T) *gorm.DB {
	t.Helper()
	db := testutil.NewMemoryDB(t, &hrcModel.Members{}, &hrcModel.CompanyProfile{}, &hrcModel.CompanyCancellationApply{})
	testutil.InitNopLogger()
	testutil.NewRedisOrSkip(t, testutil.WithAssignGlobal())
	return db
}

// newCompanyForCancel 建企业会员（utype=2）+ 已填写企业资料
func newCompanyForCancel(t *testing.T, db *gorm.DB, uid uint64, mobile string) {
	t.Helper()
	name := "测试企业"
	require.NoError(t, db.Create(&hrcModel.Members{UID: uid, Utype: 2, Username: mobile, Mobile: mobile, Status: 1}).Error)
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: uid, CompanyName: &name, Audit: 2}).Error)
}

// sendCancelCode 发企业注销短信码并返回（cancellation 场景）
func sendCancelCode(t *testing.T, mobile string) string {
	t.Helper()
	code, err := (&AuthService{}).SendSmsCode(mobile, "cancellation")
	require.NoError(t, err)
	require.NotEmpty(t, code)
	return code
}

func TestApplyCompanyCancellation(t *testing.T) {
	db := setupCancelTest(t)
	uid := uint64(1001)
	mobile := randomMobile("139")
	newCompanyForCancel(t, db, uid, mobile)

	code := sendCancelCode(t, mobile)
	require.NoError(t, (&CompanyCancellationService{}).ApplyCompanyCancellation(context.Background(), uid, code))

	var apply hrcModel.CompanyCancellationApply
	require.NoError(t, db.Where("uid = ?", uid).First(&apply).Error)
	require.Equal(t, int8(0), apply.Status)
	require.Equal(t, "测试企业", apply.CompanyName)
	require.NotZero(t, apply.CompanyID)
	require.NotZero(t, apply.AddTime)
}

func TestApplyCompanyCancellationRejectsPending(t *testing.T) {
	db := setupCancelTest(t)
	uid := uint64(1002)
	mobile := randomMobile("138")
	newCompanyForCancel(t, db, uid, mobile)
	require.NoError(t, db.Create(&hrcModel.CompanyCancellationApply{UID: uid, CompanyID: 1, CompanyName: "测试企业", Status: 0}).Error)

	code := sendCancelCode(t, mobile)
	err := (&CompanyCancellationService{}).ApplyCompanyCancellation(context.Background(), uid, code)
	require.ErrorIs(t, err, ErrCompanyCancellationPending)
}

func TestApplyCompanyCancellationRequiresProfile(t *testing.T) {
	db := setupCancelTest(t)
	uid := uint64(1003)
	mobile := randomMobile("137")
	// 空壳企业资料（companyname 为 NULL）不可申请
	require.NoError(t, db.Create(&hrcModel.Members{UID: uid, Utype: 2, Username: mobile, Mobile: mobile, Status: 1}).Error)
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: uid, Audit: 2}).Error)

	code := sendCancelCode(t, mobile)
	err := (&CompanyCancellationService{}).ApplyCompanyCancellation(context.Background(), uid, code)
	require.ErrorIs(t, err, ErrCompanyCancellationNoProfile)
}

func TestCompanyCancellationList(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.Members{}, &hrcModel.CompanyProfile{}, &hrcModel.CompanyCancellationApply{})
	require.NoError(t, db.Create(&hrcModel.Members{UID: 2001, Utype: 2, Username: "u2001", Mobile: "13800002001", Status: 1}).Error)
	require.NoError(t, db.Create(&hrcModel.Members{UID: 2002, Utype: 2, Username: "u2002", Mobile: "13800002002", Status: 1}).Error)
	require.NoError(t, db.Create(&hrcModel.CompanyCancellationApply{UID: 2001, CompanyID: 1, CompanyName: "A公司", Status: 0}).Error)
	require.NoError(t, db.Create(&hrcModel.CompanyCancellationApply{UID: 2002, CompanyID: 2, CompanyName: "B公司", Status: 1, FinishTime: 123}).Error)

	svc := &CompanyCancellationService{}
	info := request.PageInfo{Page: 1, PageSize: 10}

	list, total, err := svc.CompanyCancellationList(context.Background(), info, 0)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, list, 2)
	// id desc：2002 在前
	require.Equal(t, "B公司", list[0].CompanyName)
	require.Equal(t, "u2002", list[0].Username)
	require.Equal(t, "13800002002", list[0].Mobile)
	require.Equal(t, "A公司", list[1].CompanyName)

	// 状态筛选：待处理（存储 0）
	list, total, err = svc.CompanyCancellationList(context.Background(), info, 1)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, uint64(2001), list[0].UID)
}

func TestHandleCompanyCancellation(t *testing.T) {
	db := setupCancelTest(t)
	uid := uint64(3001)
	mobile := randomMobile("136")
	newCompanyForCancel(t, db, uid, mobile)
	require.NoError(t, db.Create(&hrcModel.CompanyCancellationApply{UID: uid, CompanyID: 1, CompanyName: "测试企业", Status: 0}).Error)
	var apply hrcModel.CompanyCancellationApply
	require.NoError(t, db.Where("uid = ?", uid).First(&apply).Error)

	svc := &CompanyCancellationService{}
	require.NoError(t, svc.HandleCompanyCancellation(context.Background(), apply.ID))

	var got hrcModel.CompanyCancellationApply
	require.NoError(t, db.First(&got, apply.ID).Error)
	require.Equal(t, int8(1), got.Status)
	require.NotZero(t, got.FinishTime)
	// 企业主体数据已清
	var pCount int64
	require.NoError(t, db.Model(&hrcModel.CompanyProfile{}).Where("uid = ?", uid).Count(&pCount).Error)
	require.Zero(t, pCount)

	// 已处理申请拒绝重复操作
	require.ErrorIs(t, svc.HandleCompanyCancellation(context.Background(), apply.ID), ErrCompanyCancellationHandled)
}

func TestDeleteCompanyCancellation(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.CompanyCancellationApply{})
	require.NoError(t, db.Create(&hrcModel.CompanyCancellationApply{UID: 4001, CompanyID: 1, CompanyName: "A公司", Status: 0}).Error)
	var apply hrcModel.CompanyCancellationApply
	require.NoError(t, db.First(&apply).Error)

	svc := &CompanyCancellationService{}
	require.NoError(t, svc.DeleteCompanyCancellation(context.Background(), apply.ID))
	require.ErrorIs(t, svc.DeleteCompanyCancellation(context.Background(), apply.ID), ErrCompanyCancellationNotFound)
}
