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

func TestApplySuccess(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.PersonalJobsApply{}, &hrcModel.Resume{}, &hrcModel.Jobs{}, &hrcModel.Config{}, &hrcModel.Pms{}, &hrcModel.MembersMsgtip{})
	require.NoError(t, db.Create(&hrcModel.Resume{UID: 1, FullName: "张三", Audit: 1, CompletePercent: MinApplyResumeCompletePercent}).Error)
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 100, JobsName: "Go 工程师", CompanyName: "甲企业", CompanyID: 1, Audit: 1, Display: 1, Deadline: time.Now().Add(24 * time.Hour)}}).Error)
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 200, JobsName: "前端工程师", CompanyName: "乙企业", CompanyID: 2, Audit: 1, Display: 1, Deadline: time.Now().Add(24 * time.Hour)}}).Error)

	var job1, job2 hrcModel.Jobs
	require.NoError(t, db.Where("jobs_name = ?", "Go 工程师").First(&job1).Error)
	require.NoError(t, db.Where("jobs_name = ?", "前端工程师").First(&job2).Error)

	svc := &ApplyService{}
	require.NoError(t, svc.Apply(context.Background(), 1, []uint64{job1.ID, job2.ID}, 1, ""))

	var count int64
	db.Model(&hrcModel.PersonalJobsApply{}).Count(&count)
	require.Equal(t, int64(2), count)
	db.Model(&hrcModel.Pms{}).Where("msgtouid = ?", 100).Count(&count)
	require.Equal(t, int64(1), count)
}

func TestApplyDuplicate(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.PersonalJobsApply{}, &hrcModel.Resume{}, &hrcModel.Jobs{}, &hrcModel.Config{}, &hrcModel.Pms{}, &hrcModel.MembersMsgtip{})
	require.NoError(t, db.Create(&hrcModel.Resume{UID: 1, FullName: "张三", Audit: 1, CompletePercent: MinApplyResumeCompletePercent}).Error)
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 100, JobsName: "Go 工程师", CompanyName: "甲企业", CompanyID: 1, Audit: 1, Display: 1, Deadline: time.Now().Add(24 * time.Hour)}}).Error)
	var job hrcModel.Jobs
	require.NoError(t, db.Where("jobs_name = ?", "Go 工程师").First(&job).Error)

	svc := &ApplyService{}
	require.NoError(t, svc.Apply(context.Background(), 1, []uint64{job.ID}, 1, ""))
	// 同企业同简历再投（即使换职位）→ 拒绝
	require.ErrorIs(t, svc.Apply(context.Background(), 1, []uint64{job.ID}, 1, ""), ErrAlreadyApplied)
}

func TestApplyNoResume(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.PersonalJobsApply{}, &hrcModel.Resume{}, &hrcModel.Jobs{}, &hrcModel.Config{}, &hrcModel.Pms{}, &hrcModel.MembersMsgtip{})
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 100, JobsName: "Go 工程师", CompanyName: "甲企业", CompanyID: 1, Audit: 1, Display: 1, Deadline: time.Now().Add(24 * time.Hour)}}).Error)
	var job hrcModel.Jobs
	require.NoError(t, db.Where("jobs_name = ?", "Go 工程师").First(&job).Error)

	svc := &ApplyService{}
	require.ErrorIs(t, svc.Apply(context.Background(), 1, []uint64{job.ID}, 0, ""), ErrResumeRequired)
}

func TestApplyIncompleteResume(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.PersonalJobsApply{}, &hrcModel.Resume{}, &hrcModel.Jobs{}, &hrcModel.Config{}, &hrcModel.Pms{}, &hrcModel.MembersMsgtip{})
	require.NoError(t, db.Create(&hrcModel.Resume{UID: 1, FullName: "张三", Audit: 1, CompletePercent: MinApplyResumeCompletePercent - 1}).Error)
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 100, JobsName: "Go 工程师", CompanyName: "甲企业", CompanyID: 1, Audit: 1, Display: 1, Deadline: time.Now().Add(24 * time.Hour)}}).Error)
	var job hrcModel.Jobs
	require.NoError(t, db.Where("jobs_name = ?", "Go 工程师").First(&job).Error)

	svc := &ApplyService{}
	require.ErrorIs(t, svc.Apply(context.Background(), 1, []uint64{job.ID}, 0, ""), ErrResumeIncomplete)
}

// 指定已软删简历投递 → 「简历不存在或已删除」（修改批 X 文案修复）
func TestApplyDeletedResume(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.PersonalJobsApply{}, &hrcModel.Resume{}, &hrcModel.Jobs{}, &hrcModel.Config{}, &hrcModel.Pms{}, &hrcModel.MembersMsgtip{})
	now := hrcModel.Now()
	require.NoError(t, db.Create(&hrcModel.Resume{UID: 1, FullName: "张三", Audit: 1, DeletedAt: &now}).Error)
	var r hrcModel.Resume
	require.NoError(t, db.First(&r).Error)
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 100, JobsName: "Go 工程师", CompanyName: "甲企业", CompanyID: 1, Audit: 1, Display: 1, Deadline: time.Now().Add(24 * time.Hour)}}).Error)
	var job hrcModel.Jobs
	require.NoError(t, db.Where("jobs_name = ?", "Go 工程师").First(&job).Error)

	svc := &ApplyService{}
	require.ErrorIs(t, svc.Apply(context.Background(), 1, []uint64{job.ID}, r.ID, ""), ErrResumeUnavailable)
}

func TestApplyJobNotAvailable(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.PersonalJobsApply{}, &hrcModel.Resume{}, &hrcModel.Jobs{}, &hrcModel.Config{}, &hrcModel.Pms{}, &hrcModel.MembersMsgtip{})
	require.NoError(t, db.Create(&hrcModel.Resume{UID: 1, FullName: "张三", Audit: 1, CompletePercent: MinApplyResumeCompletePercent}).Error)
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 100, JobsName: "已下架", CompanyName: "甲企业", CompanyID: 1, Audit: 1, Display: 2, Deadline: time.Now().Add(24 * time.Hour)}}).Error)
	var job hrcModel.Jobs
	require.NoError(t, db.Where("jobs_name = ?", "已下架").First(&job).Error)

	svc := &ApplyService{}
	require.ErrorIs(t, svc.Apply(context.Background(), 1, []uint64{job.ID}, 1, ""), ErrJobNotAvailable)
}

func TestApplyDailyLimit(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.PersonalJobsApply{}, &hrcModel.Resume{}, &hrcModel.Jobs{}, &hrcModel.Config{}, &hrcModel.Pms{}, &hrcModel.MembersMsgtip{})
	require.NoError(t, db.Create(&hrcModel.Config{CfgGroup: "jobs", Name: "mscms_apply_jobs_max", Value: "1"}).Error)
	require.NoError(t, db.Create(&hrcModel.Resume{UID: 1, FullName: "张三", Audit: 1, CompletePercent: MinApplyResumeCompletePercent}).Error)
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 100, JobsName: "Go 工程师", CompanyName: "甲企业", CompanyID: 1, Audit: 1, Display: 1, Deadline: time.Now().Add(24 * time.Hour)}}).Error)
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 200, JobsName: "前端工程师", CompanyName: "乙企业", CompanyID: 2, Audit: 1, Display: 1, Deadline: time.Now().Add(24 * time.Hour)}}).Error)
	var job1, job2 hrcModel.Jobs
	require.NoError(t, db.Where("jobs_name = ?", "Go 工程师").First(&job1).Error)
	require.NoError(t, db.Where("jobs_name = ?", "前端工程师").First(&job2).Error)

	svc := &ApplyService{}
	// 上限 1，投 2 个 → 拒绝
	require.ErrorIs(t, svc.Apply(context.Background(), 1, []uint64{job1.ID, job2.ID}, 1, ""), ErrApplyDailyLimit)
}

func TestApplyListDelete(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.PersonalJobsApply{}, &hrcModel.Resume{}, &hrcModel.Jobs{}, &hrcModel.Config{}, &hrcModel.Pms{}, &hrcModel.MembersMsgtip{})
	require.NoError(t, db.Create(&hrcModel.Resume{UID: 1, FullName: "张三", Audit: 1, CompletePercent: MinApplyResumeCompletePercent}).Error)
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 100, JobsName: "Go 工程师", CompanyName: "甲企业", CompanyID: 1, Audit: 1, Display: 1, Deadline: time.Now().Add(24 * time.Hour)}}).Error)
	var job hrcModel.Jobs
	require.NoError(t, db.Where("jobs_name = ?", "Go 工程师").First(&job).Error)

	svc := &ApplyService{}
	require.NoError(t, svc.Apply(context.Background(), 1, []uint64{job.ID}, 1, ""))

	list, total, err := svc.List(context.Background(), 1, request.PageInfo{Page: 1, PageSize: 10}, 0)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, list, 1)

	require.NoError(t, svc.Delete(context.Background(), 1, list[0].DID))
	require.ErrorIs(t, svc.Delete(context.Background(), 2, list[0].DID), ErrApplyNotFound)
}
