package hrc

import (
	"context"
	"errors"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestMessageServiceReadDeleteAndOwnership(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.Pms{}, &hrcModel.MembersMsgtip{})
	svc := &MessageService{}
	require.NoError(t, svc.SendSystemNotice(db, SystemNotice{ToUID: 10, Title: "投递提醒", Message: "您收到新的投递", Type: "application"}))
	require.NoError(t, svc.SendSystemNotice(db, SystemNotice{ToUID: 10, Title: "面试提醒", Message: "您收到新的面试邀请", Type: "interview"}))
	require.NoError(t, svc.SendSystemNotice(db, SystemNotice{ToUID: 20, Title: "其他用户", Message: "不应被操作", Type: "application"}))

	list, total, err := svc.List(context.Background(), 10, request.PageInfo{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, list, 2)
	unread, err := svc.UnreadCount(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, int64(2), unread)

	require.NoError(t, svc.MarkRead(context.Background(), 10, []uint64{list[0].ID}))
	unread, err = svc.UnreadCount(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, int64(1), unread)
	require.ErrorIs(t, svc.MarkRead(context.Background(), 10, nil), ErrMessageIDsInvalid)

	var other hrcModel.Pms
	require.NoError(t, db.Where("msgtouid = ?", 20).First(&other).Error)
	require.NoError(t, svc.Delete(context.Background(), 10, []uint64{list[1].ID, other.ID}))
	var remaining int64
	require.NoError(t, db.Model(&hrcModel.Pms{}).Where("msgtouid = ?", 10).Count(&remaining).Error)
	require.Equal(t, int64(1), remaining)
	require.NoError(t, db.First(&other, other.ID).Error)

	var tip hrcModel.MembersMsgtip
	require.NoError(t, db.Where("uid = ? AND type = ?", 10, pmsMessageTipType).First(&tip).Error)
	require.Equal(t, 0, tip.Unread)
}

func TestMessageServiceRollsBackWithBusinessTransaction(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.Pms{}, &hrcModel.MembersMsgtip{})
	svc := &MessageService{}
	err := db.Transaction(func(tx *gorm.DB) error {
		require.NoError(t, svc.SendSystemNotice(tx, SystemNotice{ToUID: 10, Title: "测试", Message: "事务消息"}))
		return errors.New("rollback")
	})
	require.Error(t, err)
	var count int64
	require.NoError(t, db.Model(&hrcModel.Pms{}).Count(&count).Error)
	require.Zero(t, count)
}
