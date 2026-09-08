package hrc

import (
	"context"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/stretchr/testify/require"
)

func companyNamePtr(name string) *string { return &name }

func TestCompanyAuditList(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.CompanyProfile{}, &hrcModel.AuditReason{})
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: 1, CompanyName: companyNamePtr("甲企业"), Audit: 2}).Error)
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: 2, CompanyName: companyNamePtr("乙企业"), Audit: 1}).Error)
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: 3, CompanyName: companyNamePtr("丙企业"), Audit: 2}).Error)

	svc := &CompanyAuditService{}
	info := request.PageInfo{Page: 1, PageSize: 10}

	// 全部（audit<0）
	list, total, err := svc.CompanyList(context.Background(), info, -1, "")
	require.NoError(t, err)
	require.Equal(t, int64(3), total)
	require.Len(t, list, 3)

	// audit=2 筛选
	list, total, err = svc.CompanyList(context.Background(), info, 2, "")
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Equal(t, uint64(3), list[0].UID) // id desc

	// audit=0 草稿筛选
	list, total, err = svc.CompanyList(context.Background(), info, 0, "")
	require.NoError(t, err)
	require.Equal(t, int64(0), total)

	// 关键字筛选
	list, total, err = svc.CompanyList(context.Background(), info, -1, "乙")
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, "乙企业", *list[0].CompanyName)
}

func TestCompanyAuditDetail(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.CompanyProfile{}, &hrcModel.AuditReason{})
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: 1, CompanyName: companyNamePtr("甲企业"), Audit: 2, Logo: "uploads/logo.png"}).Error)

	svc := &CompanyAuditService{}
	p, err := svc.CompanyDetail(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, uint64(1), p.UID)
	require.Equal(t, "uploads/logo.png", p.Logo)

	_, err = svc.CompanyDetail(context.Background(), 999)
	require.ErrorIs(t, err, ErrCompanyProfileNotFound)
}

func TestCompanyAuditPass(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.CompanyProfile{}, &hrcModel.AuditReason{})
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: 1, CompanyName: companyNamePtr("甲企业"), Audit: 2}).Error)

	svc := &CompanyAuditService{}
	require.NoError(t, svc.CompanyAudit(context.Background(), 1, 1, ""))

	var p hrcModel.CompanyProfile
	require.NoError(t, db.First(&p, 1).Error)
	require.Equal(t, int8(1), p.Audit)
	// 通过不写 reason
	var reasonCount int64
	require.NoError(t, db.Model(&hrcModel.AuditReason{}).Count(&reasonCount).Error)
	require.Zero(t, reasonCount)
}

func TestCompanyAuditRejectWritesReason(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.CompanyProfile{}, &hrcModel.AuditReason{})
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: 1, CompanyName: companyNamePtr("甲企业"), Audit: 2}).Error)

	svc := &CompanyAuditService{}
	require.NoError(t, svc.CompanyAudit(context.Background(), 1, 3, "营业执照不清晰"))

	var p hrcModel.CompanyProfile
	require.NoError(t, db.First(&p, 1).Error)
	require.Equal(t, int8(3), p.Audit)

	var reason hrcModel.AuditReason
	require.NoError(t, db.Where("type = ? AND type_id = ?", hrcModel.AuditTypeCompany, 1).First(&reason).Error)
	require.Equal(t, "营业执照不清晰", reason.Reason)
	require.NotZero(t, reason.AddTime)
}

func TestCompanyAuditRejectRequiresReason(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.CompanyProfile{}, &hrcModel.AuditReason{})
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: 1, CompanyName: companyNamePtr("甲企业"), Audit: 2}).Error)

	svc := &CompanyAuditService{}
	// 不通过但未填原因 → 拒绝（防绕过前端直调）
	require.Error(t, svc.CompanyAudit(context.Background(), 1, 3, ""))

	var p hrcModel.CompanyProfile
	require.NoError(t, db.First(&p, 1).Error)
	require.Equal(t, int8(2), p.Audit) // audit 未被修改
}

func TestCompanyAuditInvalid(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.CompanyProfile{}, &hrcModel.AuditReason{})
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: 1, CompanyName: companyNamePtr("甲企业"), Audit: 2}).Error)

	svc := &CompanyAuditService{}
	require.Error(t, svc.CompanyAudit(context.Background(), 1, 2, ""))
	require.ErrorIs(t, svc.CompanyAudit(context.Background(), 999, 1, ""), ErrCompanyProfileNotFound)
}
