package hrc

import (
	"context"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/stretchr/testify/require"
)

// 主账号创建/管理 HR 子账号；归属与隔离
func TestCompanyHRCreateListAndOwnership(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.Members{}, &hrcModel.MembersInfo{})
	svc := &CompanyHRService{}
	ctx := context.Background()

	item, err := svc.Create(ctx, 100, "13900000001", "Abc1234567!", "张经理", smsMasterKey)
	require.NoError(t, err)
	require.NotZero(t, item.UID)
	require.Equal(t, "张经理", item.RealName)

	var member hrcModel.Members
	require.NoError(t, db.Where("uid = ?", item.UID).First(&member).Error)
	require.Equal(t, uint64(100), member.CompanyUID, "HR 归属企业主账号")
	require.Equal(t, int8(2), member.Utype)

	// 手机号重复
	_, err = svc.Create(ctx, 100, "13900000001", "Abc1234567!", "", smsMasterKey)
	require.ErrorIs(t, err, ErrHRMobileExists)

	// 列表仅本人企业可见
	list, err := svc.List(ctx, 100)
	require.NoError(t, err)
	require.Len(t, list, 1)
	other, err := svc.List(ctx, 200)
	require.NoError(t, err)
	require.Len(t, other, 0)

	// 跨企业操作 → 不存在
	require.ErrorIs(t, svc.SetStatus(ctx, 200, item.UID, 2), ErrHRNotFound)

	// 启用/禁用 + 重置密码
	require.NoError(t, svc.SetStatus(ctx, 100, item.UID, 2))
	require.NoError(t, svc.ResetPassword(ctx, 100, item.UID, "Xyz9876543!"))

	// 移除后列表为空，数据库物理删除，且手机号可复用
	require.NoError(t, svc.Delete(ctx, 100, item.UID))
	list, err = svc.List(ctx, 100)
	require.NoError(t, err)
	require.Len(t, list, 0)

	var remain int64
	require.NoError(t, db.Model(&hrcModel.Members{}).Where("uid = ?", item.UID).Count(&remain).Error)
	require.Equal(t, int64(0), remain, "HR 记录应物理删除")
	var infoCount int64
	require.NoError(t, db.Model(&hrcModel.MembersInfo{}).Where("uid = ?", item.UID).Count(&infoCount).Error)
	require.Equal(t, int64(0), infoCount, "HR 资料应一并删除")

	recreated, err := svc.Create(ctx, 100, "13900000001", "Abc1234567!", "", smsMasterKey)
	require.NoError(t, err, "删除后同一手机号应可重新创建")
	require.NotEqual(t, item.UID, recreated.UID)

	// 弱密码拒绝
	_, err = svc.Create(ctx, 100, "13900000002", "short", "", smsMasterKey)
	require.Error(t, err)
}
