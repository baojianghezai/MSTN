package hrc

import (
	"context"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func chatDB(t *testing.T) *gorm.DB {
	return testutil.NewMemoryDB(t,
		&hrcModel.ImSession{},
		&hrcModel.ImMessage{},
		&hrcModel.Members{},
		&hrcModel.MembersInfo{},
		&hrcModel.CompanyProfile{},
	)
}

func seedChatUsers(t *testing.T, db *gorm.DB) {
	require.NoError(t, db.Create(&hrcModel.Members{UID: 1, Utype: 1, Username: "u1", Mobile: "13800000001", Status: 1}).Error)
	require.NoError(t, db.Create(&hrcModel.Members{UID: 2, Utype: 2, Username: "c1", Mobile: "13800000002", Status: 1}).Error)
	require.NoError(t, db.Create(&hrcModel.MembersInfo{UID: 1, RealName: "张三"}).Error)
	name := "甲企业"
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: 2, CompanyName: &name}).Error)
}

// 打开会话（幂等）→ 发送 → 未读 → 列表 → 已读 → 消息 → 己方删除
func TestChatOpenSendReadFlow(t *testing.T) {
	db := chatDB(t)
	seedChatUsers(t, db)
	svc := &ChatService{}
	ctx := context.Background()

	session, err := svc.OpenSession(ctx, 1, 1, 2, 0, "")
	require.NoError(t, err)
	require.Equal(t, uint64(1), session.PersonalUID)
	require.Equal(t, uint64(2), session.CompanyUID)
	require.Equal(t, "张三", session.PersonalName)
	require.Equal(t, "甲企业", session.CompanyName)

	// 企业反向打开 → 幂等复用同一会话
	session2, err := svc.OpenSession(ctx, 2, 2, 1, 0, "")
	require.NoError(t, err)
	require.Equal(t, session.ID, session2.ID)

	// 个人发消息 → 返回对端 uid 供 WebSocket 推送
	item, peerUID, err := svc.SendMessage(ctx, 1, 1, session.ID, "你好，想了解这个岗位")
	require.NoError(t, err)
	require.Equal(t, uint64(2), peerUID)
	require.True(t, item.Mine)

	// 企业侧未读 +1
	total, err := svc.UnreadTotal(ctx, 2, 2)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)

	// 企业会话列表：对端是求职者
	list, count, err := svc.ListSessions(ctx, 2, 2, request.PageInfo{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
	require.Equal(t, 1, list[0].Unread)
	require.Equal(t, "张三", list[0].PeerName)
	require.Equal(t, "你好，想了解这个岗位", list[0].LastContent)

	// 企业标记已读 → 未读清零
	require.NoError(t, svc.MarkRead(ctx, 2, 2, session.ID))
	total, err = svc.UnreadTotal(ctx, 2, 2)
	require.NoError(t, err)
	require.Equal(t, int64(0), total)

	// 企业视角消息（Mine=false）
	msgs, msgTotal, err := svc.Messages(ctx, 2, 2, session.ID, request.PageInfo{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(1), msgTotal)
	require.Len(t, msgs, 1)
	require.False(t, msgs[0].Mine)
	require.Equal(t, "你好，想了解这个岗位", msgs[0].Content)

	// 企业删除己方会话 → 企业列表空，个人仍可见
	require.NoError(t, svc.DeleteSession(ctx, 2, 2, session.ID))
	_, count, err = svc.ListSessions(ctx, 2, 2, request.PageInfo{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(0), count)
	_, count, err = svc.ListSessions(ctx, 1, 1, request.PageInfo{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}

// 会话归属隔离 + 非本人不可读写
func TestChatOwnership(t *testing.T) {
	db := chatDB(t)
	seedChatUsers(t, db)
	svc := &ChatService{}
	ctx := context.Background()

	session, err := svc.OpenSession(ctx, 1, 1, 2, 0, "")
	require.NoError(t, err)

	// 无关账号 uid=99 读写 → 会话不存在
	_, _, err = svc.SendMessage(ctx, 99, 1, session.ID, "hi")
	require.ErrorIs(t, err, ErrChatSessionNotFound)
	_, _, err = svc.Messages(ctx, 99, 1, session.ID, request.PageInfo{Page: 1, PageSize: 10})
	require.ErrorIs(t, err, ErrChatSessionNotFound)
	require.ErrorIs(t, svc.MarkRead(ctx, 99, 1, session.ID), ErrChatSessionNotFound)
}

// 不能和自己 / 同类型账号发起对话
func TestChatOpenValidation(t *testing.T) {
	db := chatDB(t)
	seedChatUsers(t, db)
	require.NoError(t, db.Create(&hrcModel.Members{UID: 3, Utype: 1, Username: "u3", Mobile: "13800000003", Status: 1}).Error)
	svc := &ChatService{}
	ctx := context.Background()

	_, err := svc.OpenSession(ctx, 1, 1, 1, 0, "")
	require.ErrorIs(t, err, ErrChatSelf)

	// uid=3 同为企业？不，这里 uid=3 是个人，与 uid=1 同为个人 → 拒绝
	_, err = svc.OpenSession(ctx, 1, 1, 3, 0, "")
	require.ErrorIs(t, err, ErrChatPeerInvalid)

	_, err = svc.OpenSession(ctx, 1, 1, 999, 0, "")
	require.ErrorIs(t, err, ErrChatPeerInvalid)
}
