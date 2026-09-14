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

func companyName(value string) *string { return &value }

func TestCompanySearchOnlyReturnsPublicAuditedProfiles(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.CompanyProfile{}, &hrcModel.Jobs{})
	now := time.Now()
	refreshTime := now
	visible := &hrcModel.CompanyProfile{UID: 1, CompanyName: companyName("名硕科技"), Audit: 1, UserStatus: 1, Nature: 1, Trade: 2, Scale: 3, District: "北京", DistrictCN: "北京市", NatureCN: "民营企业", TradeCN: "互联网", ScaleCN: "100-499人", ShortDesc: "招聘 Go 工程师", Refreshtime: &refreshTime}
	require.NoError(t, db.Create(visible).Error)
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: 2, CompanyName: companyName("待审企业"), Audit: 2, UserStatus: 1}).Error)
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: 3, CompanyName: companyName("已停用企业"), Audit: 1, UserStatus: 2}).Error)
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{CompanyID: visible.ID, JobsName: "Go 工程师", Audit: 1, Display: 1, Deadline: now.Add(3600 * time.Second), Refreshtime: now}}).Error)
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{CompanyID: visible.ID, JobsName: "已过期", Audit: 1, Display: 1, Deadline: now.Add(-time.Second)}}).Error)

	svc := &CompanySearchService{}
	list, total, err := svc.Search(context.Background(), CompanySearchFilter{Keyword: "名硕", Trade: 2, District: "北京"}, request.PageInfo{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, list, 1)
	require.Equal(t, visible.ID, list[0].ID)
	require.Equal(t, int64(1), list[0].JobsCount)
}

func TestCompanySearchDetailAndJobsStayPublic(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.CompanyProfile{}, &hrcModel.Jobs{})
	now := time.Now()
	profile := &hrcModel.CompanyProfile{UID: 1, CompanyName: companyName("名硕科技"), Audit: 1, UserStatus: 1, Contents: "企业介绍", Address: "北京市朝阳区", Website: "https://example.com", Telephone: "13800138000", Email: "secret@example.com", CertificateImg: "secret.png", Click: 10}
	require.NoError(t, db.Create(profile).Error)
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{CompanyID: profile.ID, JobsName: "后端工程师", Audit: 1, Display: 1, Deadline: now.Add(3600 * time.Second), Refreshtime: now}}).Error)
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{CompanyID: profile.ID, JobsName: "暂停职位", Audit: 1, Display: 2, Deadline: now.Add(3600 * time.Second)}}).Error)

	svc := &CompanySearchService{}
	detail, err := svc.Detail(context.Background(), profile.ID)
	require.NoError(t, err)
	require.Equal(t, "企业介绍", detail.Contents)
	require.Len(t, detail.Jobs, 1)

	jobs, total, err := svc.Jobs(context.Background(), profile.ID, request.PageInfo{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, jobs, 1)

	var after hrcModel.CompanyProfile
	require.NoError(t, db.First(&after, profile.ID).Error)
	require.Equal(t, uint(11), after.Click)
	_, err = svc.Detail(context.Background(), 999)
	require.ErrorIs(t, err, ErrCompanyNotFound)
}
