package hrc

import (
	"context"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
)

func TestAdminService_ProcessAppealRestore(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.Members{}, &hrcModel.MembersAppeal{})

	// 已注销账号（mobile 保留）
	deletedAt := time.Unix(1700000000, 0)
	member := hrcModel.Members{
		UID: 1, Utype: 1, Username: "u_匿名_1", Mobile: "13800138000",
		Status: 3, DeletedAt: &deletedAt,
	}
	if err := db.Create(&member).Error; err != nil {
		t.Fatalf("创建注销账号失败: %v", err)
	}
	appeal := hrcModel.MembersAppeal{
		RealName: "张三", Mobile: "13800138000", Description: "请求恢复账号", Status: 0,
	}
	if err := db.Create(&appeal).Error; err != nil {
		t.Fatalf("创建申诉失败: %v", err)
	}

	svc := &AdminService{}
	if err := svc.ProcessAppeal(context.Background(), appeal.ID, 1, true); err != nil {
		t.Fatalf("ProcessAppeal 失败: %v", err)
	}

	var got hrcModel.Members
	if err := db.First(&got, member.UID).Error; err != nil {
		t.Fatalf("查询会员失败: %v", err)
	}
	if got.Status != 1 || got.DeletedAt != nil {
		t.Fatalf("账号未恢复: status=%d deleted_at=%v", got.Status, got.DeletedAt)
	}

	var gotAppeal hrcModel.MembersAppeal
	if err := db.First(&gotAppeal, appeal.ID).Error; err != nil {
		t.Fatalf("查询申诉失败: %v", err)
	}
	if gotAppeal.Status != 1 {
		t.Fatalf("申诉状态未更新: status=%d", gotAppeal.Status)
	}
}

func TestAdminService_ProcessAppealRestoreNotFound(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.Members{}, &hrcModel.MembersAppeal{})

	// 申诉的手机号没有对应的注销账号
	appeal := hrcModel.MembersAppeal{
		RealName: "李四", Mobile: "13900000000", Description: "恢复", Status: 0,
	}
	if err := db.Create(&appeal).Error; err != nil {
		t.Fatalf("创建申诉失败: %v", err)
	}

	svc := &AdminService{}
	err := svc.ProcessAppeal(context.Background(), appeal.ID, 1, true)
	if err == nil {
		t.Fatal("期望恢复失败，但返回了 nil")
	}
}
