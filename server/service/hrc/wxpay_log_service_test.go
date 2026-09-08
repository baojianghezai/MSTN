package hrc

import (
	"context"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/stretchr/testify/require"
)

func TestWxpayLogRecordAndList(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.WxpayLog{})
	svc := &WxpayLogService{}

	require.NoError(t, svc.Record(context.Background(), &hrcModel.WxpayLog{
		OpenID: "oXyZ", TradeNo: "T20260817001", Amount: "100", Status: 1,
	}))
	require.NoError(t, svc.Record(context.Background(), &hrcModel.WxpayLog{
		OpenID: "oXyZ2", TradeNo: "T20260817002", Amount: "50", Status: 0, FailReason: "余额不足",
	}))

	info := request.PageInfo{Page: 1, PageSize: 10}
	// 全部
	list, total, err := svc.List(context.Background(), info, -1)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, list, 2)
	require.NotZero(t, list[0].AddTime)

	// 成功筛选
	list, total, err = svc.List(context.Background(), info, 1)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, int8(1), list[0].Status)

	// 失败筛选
	list, total, err = svc.List(context.Background(), info, 0)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, "余额不足", list[0].FailReason)

	_ = db
}
