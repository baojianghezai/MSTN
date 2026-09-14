package hrc

import (
	"context"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func seedVisibleJob(t *testing.T, db *gorm.DB, name string, uid uint64, opts ...func(*hrcModel.Jobs)) {
	j := &hrcModel.Jobs{JobsBase: hrcModel.JobsBase{
		UID:         uid,
		JobsName:    name,
		CompanyName: "甲企业",
		Audit:       1,
		Display:     1,
		Refreshtime: time.Now(),
		Deadline:    time.Now().Add(24 * time.Hour),
	}}
	for _, o := range opts {
		o(j)
	}
	_ = db.Create(j)
}

func TestJobsSearchList(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.JobsContact{}, &hrcModel.JobsTag{}, &hrcModel.Config{})
	seedVisibleJob(t, db, "Go 后端工程师", 1, func(j *hrcModel.Jobs) {
		j.Trade = 1
		j.Category = 2
		j.District = "4401"
		j.MinWage = 10000
		j.MaxWage = 15000
	})
	seedVisibleJob(t, db, "前端工程师", 2, func(j *hrcModel.Jobs) { j.Trade = 1; j.Education = 4 })
	seedVisibleJob(t, db, "已暂停职位", 1, func(j *hrcModel.Jobs) { j.Display = 2 })
	seedVisibleJob(t, db, "未过审职位", 1, func(j *hrcModel.Jobs) { j.Audit = 2 })
	seedVisibleJob(t, db, "已过期职位", 1, func(j *hrcModel.Jobs) { j.Deadline = time.Now().Add(-time.Hour) })

	svc := &JobsSearchService{}
	info := request.PageInfo{Page: 1, PageSize: 10}

	// 全部可见（仅 2 条 display=1&audit=1&未过期）
	list, total, err := svc.Search(context.Background(), JobsSearchFilter{}, info)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, list, 2)

	// 关键字筛选
	list, total, err = svc.Search(context.Background(), JobsSearchFilter{Keyword: "前端"}, info)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, "前端工程师", list[0].JobsName)

	// 薪资下限筛选（Go 后端工程师 minwage=10000 ≥ 8000；前端工程师 minwage=0 不匹配）
	list, total, err = svc.Search(context.Background(), JobsSearchFilter{MinWage: 8000}, info)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, "Go 后端工程师", list[0].JobsName)

	// 学历筛选
	list, total, err = svc.Search(context.Background(), JobsSearchFilter{Education: 4}, info)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
}

func TestJobsSearchDetail(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.JobsContact{}, &hrcModel.JobsTag{}, &hrcModel.Config{})
	seedVisibleJob(t, db, "Go 后端工程师", 1, func(j *hrcModel.Jobs) { j.Click = 10 })
	var j hrcModel.Jobs
	require.NoError(t, db.Where("jobs_name = ?", "Go 后端工程师").First(&j).Error)
	require.NoError(t, db.Create(&hrcModel.JobsContact{PID: j.ID, Contact: "李四", Telephone: "138"}).Error)
	require.NoError(t, db.Create(&hrcModel.JobsTag{UID: 1, PID: j.ID, Tag: 1}).Error)

	svc := &JobsSearchService{}
	detail, err := svc.Detail(context.Background(), j.ID)
	require.NoError(t, err)
	require.Equal(t, "Go 后端工程师", detail.JobsName)
	require.Equal(t, "李四", detail.Contact.Contact)
	require.Len(t, detail.Tags, 1)

	// click 自增
	var after hrcModel.Jobs
	require.NoError(t, db.First(&after, j.ID).Error)
	require.Equal(t, uint(11), after.Click)

	// 不存在
	_, err = svc.Detail(context.Background(), 999)
	require.ErrorIs(t, err, ErrJobNotFound)
}

func TestJobsSearchHotWords(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.Config{})
	require.NoError(t, db.Create(&hrcModel.Config{CfgGroup: "jobs", Name: "mscms_hot_words", Value: "Java,前端,销售"}).Error)
	svc := &JobsSearchService{}
	words, err := svc.HotWords(context.Background())
	require.NoError(t, err)
	require.Equal(t, []string{"Java", "前端", "销售"}, words)
}
