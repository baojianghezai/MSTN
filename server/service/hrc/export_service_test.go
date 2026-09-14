package hrc

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/stretchr/testify/require"
)

func TestExportCompanies(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.CompanyProfile{})
	name := "甲企业"
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: 1, CompanyName: &name, NatureCN: "民营", ScaleCN: "50-100人", Contact: "李四"}).Error)
	name2 := "乙企业"
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: 2, CompanyName: &name2, NatureCN: "国企", ScaleCN: "100-500人", Contact: "王五"}).Error)

	svc := &ExportService{}
	data, err := svc.ExportCompanies(context.Background(), []uint64{1, 2})
	require.NoError(t, err)

	csvStr := string(data)
	// UTF-8 BOM 前缀
	require.True(t, strings.HasPrefix(csvStr, "\xEF\xBB\xBF"))
	require.Contains(t, csvStr, "企业名称")
	require.Contains(t, csvStr, "甲企业")
	require.Contains(t, csvStr, "乙企业")
	require.Contains(t, csvStr, "民营")
}

func TestExportCompaniesEmpty(t *testing.T) {
	testutil.NewMemoryDB(t, &hrcModel.CompanyProfile{})
	svc := &ExportService{}
	// 空 id 列表
	_, err := svc.ExportCompanies(context.Background(), nil)
	require.ErrorIs(t, err, ErrExportEmpty)
	// 无匹配数据
	_, err = svc.ExportCompanies(context.Background(), []uint64{999})
	require.ErrorIs(t, err, ErrExportEmpty)
}

func TestExportJobs(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.Jobs{})
	ts := time.Unix(1700000000, 0)
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{
		UID: 10, JobsName: "后端工程师", CompanyName: "甲企业", NatureCN: "全职",
		CategoryCN: "互联网", DistrictCN: "上海", MinWage: 10000, MaxWage: 20000,
		Amount: 2, Audit: 1, Display: 1, AddTime: ts, Refreshtime: ts, Click: 5,
	}}).Error)
	deletedAt := ts
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{
		UID: 11, JobsName: "已删职位", CompanyName: "乙企业", DeletedAt: &deletedAt,
	}}).Error)

	svc := &ExportService{}
	data, err := svc.ExportJobs(context.Background(), []uint64{1, 2})
	require.NoError(t, err)
	csvStr := string(data)
	require.True(t, strings.HasPrefix(csvStr, "\xEF\xBB\xBF"))
	require.Contains(t, csvStr, "职位名称")
	require.Contains(t, csvStr, "后端工程师")
	require.Contains(t, csvStr, "10000-20000")
	require.NotContains(t, csvStr, "已删职位")
}

func TestExportJobsEmpty(t *testing.T) {
	testutil.NewMemoryDB(t, &hrcModel.Jobs{})
	svc := &ExportService{}
	_, err := svc.ExportJobs(context.Background(), nil)
	require.ErrorIs(t, err, ErrExportEmpty)
	_, err = svc.ExportJobs(context.Background(), []uint64{999})
	require.ErrorIs(t, err, ErrExportEmpty)
}
