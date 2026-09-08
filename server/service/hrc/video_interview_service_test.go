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

func TestVideoInterviewCreate(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.VideoInterview{}, &hrcModel.Resume{}, &hrcModel.Config{})
	require.NoError(t, db.Create(&hrcModel.Config{CfgGroup: "video", Name: "video_interview_open", Value: "1"}).Error)
	require.NoError(t, db.Create(&hrcModel.Resume{UID: 10, FullName: "张三"}).Error)

	svc := &VideoInterviewService{}
	interviewTime := time.Now().Add(24 * time.Hour).Unix()
	v, err := svc.CreateInterview(context.Background(), 20, VideoInterviewInput{
		ResumeID: 1, JobsID: 100, JobsName: "Go 开发", InterviewTime: interviewTime, Contact: "李四", Telephone: "13800138000",
	})
	require.NoError(t, err)
	require.NotZero(t, v.ID)
	require.Equal(t, uint64(20), v.CompanyUID)
	require.Equal(t, uint64(10), v.PersonalUID)
	require.Len(t, v.CompanyCode, 6)
	require.Len(t, v.PersonalCode, 6)
	require.NotEqual(t, v.CompanyCode, v.PersonalCode)
	require.Equal(t, interviewTime+15*24*3600, v.Deadline)
}

func TestVideoInterviewClosed(t *testing.T) {
	testutil.NewMemoryDB(t, &hrcModel.VideoInterview{}, &hrcModel.Resume{}, &hrcModel.Config{})
	svc := &VideoInterviewService{}
	// 未配置（默认关闭）
	_, err := svc.CreateInterview(context.Background(), 20, VideoInterviewInput{ResumeID: 1, JobsID: 1})
	require.ErrorIs(t, err, ErrVideoInterviewClosed)
}

func TestVideoInterviewResumeNotFound(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.VideoInterview{}, &hrcModel.Resume{}, &hrcModel.Config{})
	require.NoError(t, db.Create(&hrcModel.Config{CfgGroup: "video", Name: "video_interview_open", Value: "1"}).Error)
	svc := &VideoInterviewService{}
	_, err := svc.CreateInterview(context.Background(), 20, VideoInterviewInput{ResumeID: 999, JobsID: 1})
	require.ErrorIs(t, err, ErrResumeNotAvailable)
}

func TestVideoInterviewDuplicated(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.VideoInterview{}, &hrcModel.Resume{}, &hrcModel.Config{})
	require.NoError(t, db.Create(&hrcModel.Config{CfgGroup: "video", Name: "video_interview_open", Value: "1"}).Error)
	require.NoError(t, db.Create(&hrcModel.Resume{UID: 10, FullName: "张三"}).Error)
	svc := &VideoInterviewService{}
	in := VideoInterviewInput{ResumeID: 1, JobsID: 100, JobsName: "Go", InterviewTime: time.Now().Add(24 * time.Hour).Unix(), Contact: "李", Telephone: "1"}

	_, err := svc.CreateInterview(context.Background(), 20, in)
	require.NoError(t, err)
	// 同企业 + 同简历 + 同职位 + 未过期 → 拒绝重复
	_, err = svc.CreateInterview(context.Background(), 20, in)
	require.ErrorIs(t, err, ErrVideoInterviewDuplicated)
}

func TestVideoInterviewRoomByCode(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.VideoInterview{}, &hrcModel.Resume{}, &hrcModel.Config{})
	require.NoError(t, db.Create(&hrcModel.Config{CfgGroup: "video", Name: "video_interview_open", Value: "1"}).Error)
	require.NoError(t, db.Create(&hrcModel.Resume{UID: 10, FullName: "张三"}).Error)
	svc := &VideoInterviewService{}
	v, err := svc.CreateInterview(context.Background(), 20, VideoInterviewInput{
		ResumeID: 1, JobsID: 100, JobsName: "Go", InterviewTime: time.Now().Add(24 * time.Hour).Unix(), Contact: "李", Telephone: "1",
	})
	require.NoError(t, err)

	got, utype, err := svc.RoomByCode(context.Background(), v.CompanyCode)
	require.NoError(t, err)
	require.Equal(t, int8(2), utype)
	require.Equal(t, v.ID, got.ID)

	got, utype, err = svc.RoomByCode(context.Background(), v.PersonalCode)
	require.NoError(t, err)
	require.Equal(t, int8(1), utype)
	require.Equal(t, v.ID, got.ID)

	_, _, err = svc.RoomByCode(context.Background(), "ZZZZZZ")
	require.ErrorIs(t, err, ErrVideoInterviewNotFound)
}

func TestVideoInterviewCompanyDelete(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.VideoInterview{}, &hrcModel.Resume{}, &hrcModel.Config{})
	require.NoError(t, db.Create(&hrcModel.Config{CfgGroup: "video", Name: "video_interview_open", Value: "1"}).Error)
	require.NoError(t, db.Create(&hrcModel.Resume{UID: 10, FullName: "张三"}).Error)
	svc := &VideoInterviewService{}
	v, err := svc.CreateInterview(context.Background(), 20, VideoInterviewInput{
		ResumeID: 1, JobsID: 100, JobsName: "Go", InterviewTime: time.Now().Add(24 * time.Hour).Unix(), Contact: "李", Telephone: "1",
	})
	require.NoError(t, err)

	// 非归属企业删除 → 不存在
	require.ErrorIs(t, svc.CompanyDelete(context.Background(), 21, v.ID), ErrVideoInterviewNotFound)
	// 归属企业删除 → 成功
	require.NoError(t, svc.CompanyDelete(context.Background(), 20, v.ID))
	var n int64
	db.Model(&hrcModel.VideoInterview{}).Count(&n)
	require.Zero(t, n)
}

func TestVideoInterviewAdminList(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.VideoInterview{}, &hrcModel.Resume{}, &hrcModel.CompanyProfile{}, &hrcModel.Config{})
	require.NoError(t, db.Create(&hrcModel.Resume{UID: 10, FullName: "张三"}).Error)
	require.NoError(t, db.Create(&hrcModel.Resume{UID: 11, FullName: "王五"}).Error)
	name := "甲企业"
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: 20, CompanyName: &name}).Error)
	require.NoError(t, db.Create(&hrcModel.VideoInterview{CompanyUID: 20, PersonalUID: 10, JobsID: 1, JobsName: "后端", InterviewTime: time.Now().Unix()}).Error)
	require.NoError(t, db.Create(&hrcModel.VideoInterview{CompanyUID: 21, PersonalUID: 11, JobsID: 2, JobsName: "前端", InterviewTime: time.Now().Unix()}).Error)

	svc := &VideoInterviewService{}
	info := request.PageInfo{Page: 1, PageSize: 10}

	list, total, err := svc.AdminList(context.Background(), info, "")
	require.NoError(t, err)
	require.Equal(t, int64(2), total)

	// 按简历姓名关键字
	list, total, err = svc.AdminList(context.Background(), info, "张三")
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, "张三", list[0].FullName)

	// 按企业名关键字
	list, total, err = svc.AdminList(context.Background(), info, "甲企业")
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, "甲企业", list[0].CompanyName)

	// 按职位名关键字
	list, total, err = svc.AdminList(context.Background(), info, "前端")
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, "前端", list[0].JobsName)
}
