package hrc

import (
	"context"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/stretchr/testify/require"
)

func TestDashboardMetrics(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.Members{}, &hrcModel.Resume{}, &hrcModel.CompanyProfile{}, &hrcModel.Jobs{}, &hrcModel.JobsTmp{}, &hrcModel.PersonalJobsApply{}, &hrcModel.VideoInterview{}, &hrcModel.MembersAppeal{}, &hrcModel.CompanyCancellationApply{})
	now := time.Now().Unix()
	yesterday := time.Now().Add(-24 * time.Hour).Unix()

	// 今日：2 个人 + 1 企业注册、1 简历、1 企业、1 视频面试
	require.NoError(t, db.Create(&hrcModel.Members{Utype: 1, Username: "u1", Mobile: "13800000001", RegTime: now}).Error)
	require.NoError(t, db.Create(&hrcModel.Members{Utype: 1, Username: "u2", Mobile: "13800000002", RegTime: now}).Error)
	require.NoError(t, db.Create(&hrcModel.Members{Utype: 2, Username: "c1", Mobile: "13800000003", RegTime: now}).Error)
	// 昨日：1 个人注册
	require.NoError(t, db.Create(&hrcModel.Members{Utype: 1, Username: "u3", Mobile: "13800000004", RegTime: yesterday}).Error)
	require.NoError(t, db.Create(&hrcModel.Resume{UID: 1, AddTime: now}).Error)
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: 2, AddTime: now}).Error)
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 2, AddTime: now}}).Error)
	require.NoError(t, db.Create(&hrcModel.PersonalJobsApply{PersonalUID: 1, ResumeID: 1, CompanyUID: 2, ApplyAddtime: now}).Error)
	require.NoError(t, db.Create(&hrcModel.VideoInterview{CompanyUID: 2, PersonalUID: 1, AddTime: now}).Error)
	// 待办（audit=2，老数据不参与今日新增计数）
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: 3, Audit: 2, AddTime: now - 3*24*3600}).Error)
	require.NoError(t, db.Create(&hrcModel.JobsTmp{JobsBase: hrcModel.JobsBase{UID: 2, Audit: 2}}).Error)
	require.NoError(t, db.Create(&hrcModel.MembersAppeal{UID: 1, Status: 0}).Error)
	require.NoError(t, db.Create(&hrcModel.CompanyCancellationApply{UID: 2, Status: 0}).Error)

	svc := &DashboardService{}
	data, err := svc.Dashboard(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(2), data.Today.PersonalUsers)
	require.Equal(t, int64(1), data.Today.CompanyUsers)
	require.Equal(t, int64(1), data.Today.Resumes)
	require.Equal(t, int64(1), data.Today.Companies)
	require.Equal(t, int64(1), data.Today.Jobs)
	require.Equal(t, int64(1), data.Today.Applications)
	require.Equal(t, int64(1), data.Today.VideoInterviews)
	require.Equal(t, int64(1), data.Yesterday.PersonalUsers)
	require.Equal(t, int64(1), data.Todo.CompanyAudit)
	require.Equal(t, int64(1), data.Todo.JobAudit)
	require.Equal(t, int64(1), data.Todo.Appeal)
	require.Equal(t, int64(1), data.Todo.CompanyCancellation)
}

func TestDashboardTrendRegister(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.Members{})
	now := time.Now()
	require.NoError(t, db.Create(&hrcModel.Members{Utype: 1, Username: "u1", Mobile: "13800000001", RegTime: now.Unix()}).Error)
	require.NoError(t, db.Create(&hrcModel.Members{Utype: 2, Username: "c1", Mobile: "13800000002", RegTime: now.Unix()}).Error)
	// 3 天前注册 1 个个人
	threeDaysAgo := now.AddDate(0, 0, -3)
	require.NoError(t, db.Create(&hrcModel.Members{Utype: 1, Username: "u3", Mobile: "13800000003", RegTime: threeDaysAgo.Unix()}).Error)

	svc := &DashboardService{}
	points, err := svc.Trend(context.Background(), 7, "register")
	require.NoError(t, err)
	require.Len(t, points, 7)

	var todayP, todayC, threeDayP int64
	for _, p := range points {
		if p.Date == now.Format("2006-01-02") {
			todayP = p.Personal
			todayC = p.Company
		}
		if p.Date == threeDaysAgo.Format("2006-01-02") {
			threeDayP = p.Personal
		}
	}
	require.Equal(t, int64(1), todayP)
	require.Equal(t, int64(1), todayC)
	require.Equal(t, int64(1), threeDayP)
}

func TestDashboardTrendInvalidMetric(t *testing.T) {
	testutil.NewMemoryDB(t, &hrcModel.Members{})
	svc := &DashboardService{}
	_, err := svc.Trend(context.Background(), 7, "bogus")
	require.Error(t, err)
}

func TestDashboardTrendJobAndApplication(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.PersonalJobsApply{})
	now := time.Now()
	twoDaysAgo := now.AddDate(0, 0, -2)
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 1, AddTime: now.Unix()}}).Error)
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 1, AddTime: twoDaysAgo.Unix()}}).Error)
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 1, AddTime: now.Unix(), DeletedAt: 1}}).Error)
	require.NoError(t, db.Create(&hrcModel.PersonalJobsApply{PersonalUID: 1, ResumeID: 1, CompanyUID: 1, ApplyAddtime: now.Unix()}).Error)
	require.NoError(t, db.Create(&hrcModel.PersonalJobsApply{PersonalUID: 2, ResumeID: 2, CompanyUID: 2, ApplyAddtime: twoDaysAgo.Unix()}).Error)

	svc := &DashboardService{}
	jobPoints, err := svc.Trend(context.Background(), 7, "job")
	require.NoError(t, err)
	applicationPoints, err := svc.Trend(context.Background(), 7, "application")
	require.NoError(t, err)

	countsByDate := func(points []TrendPoint) map[string]int64 {
		result := make(map[string]int64, len(points))
		for _, point := range points {
			result[point.Date] = point.Count
		}
		return result
	}
	jobCounts := countsByDate(jobPoints)
	applicationCounts := countsByDate(applicationPoints)
	require.Equal(t, int64(1), jobCounts[now.Format("2006-01-02")])
	require.Equal(t, int64(1), jobCounts[twoDaysAgo.Format("2006-01-02")])
	require.Equal(t, int64(1), applicationCounts[now.Format("2006-01-02")])
	require.Equal(t, int64(1), applicationCounts[twoDaysAgo.Format("2006-01-02")])
}

func TestDashboardResumeDistribution(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.Resume{})
	require.NoError(t, db.Create(&hrcModel.Resume{UID: 1, Sex: 1, SexCN: "男", Education: 4, EducationCN: "本科", Experience: 3}).Error)
	require.NoError(t, db.Create(&hrcModel.Resume{UID: 2, Sex: 1, SexCN: "男", Education: 2, EducationCN: "高中", Experience: 3}).Error)
	require.NoError(t, db.Create(&hrcModel.Resume{UID: 3, Sex: 2, SexCN: "女", Education: 4, EducationCN: "本科", Experience: 1}).Error)

	svc := &DashboardService{}
	data, err := svc.ResumeDistribution(context.Background())
	require.NoError(t, err)
	require.Len(t, data.Sex, 2)
	require.Equal(t, int64(1), data.Sex[0].Code) // 男
	require.Equal(t, int64(2), data.Sex[0].Count)
	require.Len(t, data.Education, 2)
	require.Equal(t, int64(4), data.Education[1].Code) // 本科（排序 2,4）
	require.Equal(t, int64(2), data.Education[1].Count)
}

func TestDashboardCompanyDistribution(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.CompanyProfile{})
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: 1, Nature: 1, NatureCN: "民营", Scale: 2, ScaleCN: "50-100人"}).Error)
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: 2, Nature: 1, NatureCN: "民营", Scale: 3, ScaleCN: "100-500人"}).Error)

	svc := &DashboardService{}
	data, err := svc.CompanyDistribution(context.Background())
	require.NoError(t, err)
	require.Len(t, data.Nature, 1)
	require.Equal(t, int64(1), data.Nature[0].Code)
	require.Equal(t, int64(2), data.Nature[0].Count)
	require.Len(t, data.Scale, 2)
}
