package hrc

import (
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
)

func TestUnbindMobileReleasesUsername(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.Members{}, &hrcModel.UnbindMobile{})

	member := hrcModel.Members{
		UID: 1, Utype: 1, Username: "13800138000", Mobile: "13800138000", Status: 1,
	}
	if err := db.Create(&member).Error; err != nil {
		t.Fatalf("创建会员失败: %v", err)
	}

	svc := &AuthService{}
	if err := svc.UnbindMobile(member.UID); err != nil {
		t.Fatalf("UnbindMobile 失败: %v", err)
	}

	var got hrcModel.Members
	if err := db.First(&got, member.UID).Error; err != nil {
		t.Fatalf("查询会员失败: %v", err)
	}
	const placeholder = "unbound_1"
	if got.Mobile != placeholder || got.Username != placeholder {
		t.Fatalf("解绑未同时释放 username: mobile=%q username=%q", got.Mobile, got.Username)
	}
}
