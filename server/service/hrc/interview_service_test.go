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

func interviewTestInput(resumeID, jobsID uint64) InterviewCreateInput {
	return InterviewCreateInput{
		ResumeID:      resumeID,
		JobsID:        jobsID,
		InterviewTime: time.Now().Add(time.Hour),
		Address:       "上海市浦东新区示例路 1 号",
		Contact:       "王经理",
		Telephone:     "13800138000",
		Notes:         "请携带作品集",
	}
}

func TestInterviewCreateAndDuplicate(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.CompanyInterview{}, &hrcModel.Resume{}, &hrcModel.Jobs{}, &hrcModel.Pms{}, &hrcModel.MembersMsgtip{})
	resume := hrcModel.Resume{UID: 10, FullName: "张三", Display: 1, Audit: 1}
	require.NoError(t, db.Create(&resume).Error)
	job := hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 20, JobsName: "Go 工程师", CompanyName: "示例企业", CompanyID: 3}}
	require.NoError(t, db.Create(&job).Error)

	svc := &InterviewService{}
	created, err := svc.Create(context.Background(), 20, interviewTestInput(resume.ID, job.ID))
	require.NoError(t, err)
	require.Equal(t, resume.FullName, created.ResumeName)
	require.Equal(t, job.JobsName, created.JobsName)
	require.Equal(t, int8(1), created.PersonalLook)
	var messageCount int64
	require.NoError(t, db.Model(&hrcModel.Pms{}).Where("msgtouid = ?", resume.UID).Count(&messageCount).Error)
	require.Equal(t, int64(1), messageCount)
	require.ErrorIs(t, func() error {
		_, err := svc.Create(context.Background(), 20, interviewTestInput(resume.ID, job.ID))
		return err
	}(), ErrInterviewDuplicate)
}

func TestInterviewCreateValidatesResources(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.CompanyInterview{}, &hrcModel.Resume{}, &hrcModel.Jobs{}, &hrcModel.Pms{}, &hrcModel.MembersMsgtip{})
	hidden := hrcModel.Resume{UID: 10, FullName: "隐藏简历", Display: 2, Audit: 1}
	require.NoError(t, db.Create(&hidden).Error)
	job := hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 20, JobsName: "Go 工程师"}}
	require.NoError(t, db.Create(&job).Error)

	svc := &InterviewService{}
	require.ErrorIs(t, func() error {
		_, err := svc.Create(context.Background(), 20, interviewTestInput(hidden.ID, job.ID))
		return err
	}(), ErrResumeUnavailable)
	require.ErrorIs(t, func() error {
		_, err := svc.Create(context.Background(), 21, interviewTestInput(hidden.ID, job.ID))
		return err
	}(), ErrResumeUnavailable)

	publicResume := hrcModel.Resume{UID: 10, FullName: "公开简历", Display: 1, Audit: 1}
	require.NoError(t, db.Create(&publicResume).Error)
	require.ErrorIs(t, func() error {
		_, err := svc.Create(context.Background(), 21, interviewTestInput(publicResume.ID, job.ID))
		return err
	}(), ErrJobNotAvailable)
}

func TestInterviewCompanyAndPersonalOwnership(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.CompanyInterview{}, &hrcModel.Resume{}, &hrcModel.Jobs{}, &hrcModel.Pms{}, &hrcModel.MembersMsgtip{})
	resume := hrcModel.Resume{UID: 10, FullName: "张三", Display: 1, Audit: 1}
	require.NoError(t, db.Create(&resume).Error)
	job := hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 20, JobsName: "Go 工程师", CompanyName: "示例企业", CompanyID: 3}}
	require.NoError(t, db.Create(&job).Error)

	svc := &InterviewService{}
	created, err := svc.Create(context.Background(), 20, interviewTestInput(resume.ID, job.ID))
	require.NoError(t, err)

	companyList, companyTotal, err := svc.CompanyList(context.Background(), 20, request.PageInfo{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(1), companyTotal)
	require.Len(t, companyList, 1)

	personalList, personalTotal, err := svc.PersonalList(context.Background(), 10, request.PageInfo{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(1), personalTotal)
	require.Len(t, personalList, 1)
	require.ErrorIs(t, svc.MarkRead(context.Background(), 11, created.DID), ErrInterviewNotFound)
	require.NoError(t, svc.MarkRead(context.Background(), 10, created.DID))

	var after hrcModel.CompanyInterview
	require.NoError(t, db.First(&after, created.DID).Error)
	require.Equal(t, int8(2), after.PersonalLook)
	require.ErrorIs(t, svc.Withdraw(context.Background(), 21, created.DID), ErrInterviewNotFound)
	require.NoError(t, svc.Withdraw(context.Background(), 20, created.DID))
	require.ErrorIs(t, svc.Withdraw(context.Background(), 20, created.DID), ErrInterviewNotFound)
}
