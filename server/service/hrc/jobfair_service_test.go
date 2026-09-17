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

// 企业举办招聘会（需权益）+ 个人报名/取消 + 报名计数
func TestJobfairCreateSignupFlow(t *testing.T) {
	db := testutil.NewMemoryDB(t,
		&hrcModel.Article{},
		&hrcModel.JobfairSignup{},
		&hrcModel.MembersSetmeal{},
		&hrcModel.CompanyProfile{},
		&hrcModel.MembersInfo{},
	)
	svc := &JobfairService{}
	ctx := context.Background()

	// 无权益 → 拒绝
	_, err := svc.Create(ctx, 100, JobfairInput{Title: "春季招聘会", HoldTime: "2026-03-15 09:00", Address: "人才市场"})
	require.ErrorIs(t, err, ErrJobfairEntitlement)

	// 授予举办招聘会权益
	require.NoError(t, db.Create(&hrcModel.MembersSetmeal{
		UID: 100, ExpireAt: hrcModel.Now().Add(time.Hour), EnableJobfair: true,
	}).Error)

	art, err := svc.Create(ctx, 100, JobfairInput{Title: "春季招聘会", HoldTime: "2026-03-15 09:00", Address: "人才市场", Content: "欢迎参加"})
	require.NoError(t, err)
	require.Equal(t, hrcModel.ArticleTypeJobfair, art.Type)
	require.Equal(t, uint64(100), art.CompanyUID)
	require.Equal(t, int8(1), art.Display)

	// 信息不完整
	_, err = svc.Create(ctx, 100, JobfairInput{Title: "缺时间地点"})
	require.ErrorIs(t, err, ErrJobfairInvalid)

	// 个人报名（幂等）
	require.NoError(t, svc.Signup(ctx, 1, art.ID))
	require.ErrorIs(t, svc.Signup(ctx, 1, art.ID), ErrJobfairAlreadySigned)

	ids, err := svc.MySignupIDs(ctx, 1)
	require.NoError(t, err)
	require.Len(t, ids, 1)

	// 我参加的招聘会含报名人数
	mine, total, err := svc.MyJobfairs(ctx, 1, request.PageInfo{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, mine, 1)
	require.Equal(t, 1, mine[0].SignupCount)

	// 企业列表归属
	list, err := svc.ListMine(ctx, 100)
	require.NoError(t, err)
	require.Len(t, list, 1)
	other, err := svc.ListMine(ctx, 200)
	require.NoError(t, err)
	require.Len(t, other, 0)

	// 取消报名
	require.NoError(t, svc.Cancel(ctx, 1, art.ID))
	require.ErrorIs(t, svc.Cancel(ctx, 1, art.ID), ErrJobfairNotSignup)

	// 跨企业删除 → 不存在
	require.ErrorIs(t, svc.Delete(ctx, 200, art.ID), ErrJobfairNotFound)
	require.NoError(t, svc.Delete(ctx, 100, art.ID))
}
