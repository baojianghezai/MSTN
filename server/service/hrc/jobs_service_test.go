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

// newJobsMemoryDB keeps the job test schema aligned with CreateJob's
// entitlement lookup.
func newJobsMemoryDB(t testing.TB, models ...any) *gorm.DB {
	models = append(models,
		&hrcModel.Setmeal{},
		&hrcModel.MembersSetmeal{},
		&hrcModel.Order{},
	)
	return testutil.NewMemoryDB(t, models...)
}

func validJob() *hrcModel.Jobs {
	return &hrcModel.Jobs{JobsBase: hrcModel.JobsBase{
		JobsName:   "Go 后端工程师",
		Nature:     1,
		NatureCN:   "全职",
		Sex:        3,
		Amount:     2,
		TopClass:   1,
		Category:   2,
		SubClass:   3,
		CategoryCN: "技术/后端",
		Trade:      1,
		District:   "4401",
		DistrictCN: "广州",
		Education:  4,
		Experience: 3,
		MinWage:    10000,
		MaxWage:    15000,
		Contents:   "负责后端服务开发",
		Deadline:   time.Now().Add(30 * 24 * time.Hour).Unix(),
	}}
}

func seedAuditedCompany(t *testing.T, db *gorm.DB, uid uint64) {
	seedAuditedCompanyNamed(t, db, uid, "甲企业")
}

// seedAuditedCompanyNamed 指定企业名（companyname 有 UNIQUE 约束，同库多企业需不同名）
func seedAuditedCompanyNamed(t *testing.T, db *gorm.DB, uid uint64, name string) {
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: uid, CompanyName: &name, Audit: 1}).Error)
}

func TestJobsCreateDisplayDirect(t *testing.T) {
	db := newJobsMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.JobsTmp{}, &hrcModel.JobsContact{}, &hrcModel.JobsTag{}, &hrcModel.CompanyProfile{}, &hrcModel.Config{})
	seedAuditedCompany(t, db, 1)
	// X2 后直接显示仅由显式配置 "2" 触发
	require.NoError(t, db.Create(&hrcModel.Config{CfgGroup: "jobs", Name: "mscms_jobs_display", Value: "2"}).Error)
	svc := &JobsService{}
	id, err := svc.CreateJob(context.Background(), 1, validJob(), &hrcModel.JobsContact{Contact: "李四", Telephone: "138"}, []uint32{1, 2})
	require.NoError(t, err)
	require.NotZero(t, id)

	var j hrcModel.Jobs
	require.NoError(t, db.First(&j, id).Error)
	require.Equal(t, int8(1), j.Audit)
	require.Equal(t, "甲企业", j.CompanyName)

	var contact hrcModel.JobsContact
	require.NoError(t, db.Where("pid = ?", id).First(&contact).Error)
	require.Equal(t, "李四", contact.Contact)
	var tagCount int64
	db.Model(&hrcModel.JobsTag{}).Where("pid = ?", id).Count(&tagCount)
	require.Equal(t, int64(2), tagCount)
}

func TestJobsCreateDisplayAudit(t *testing.T) {
	db := newJobsMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.JobsTmp{}, &hrcModel.JobsContact{}, &hrcModel.JobsTag{}, &hrcModel.CompanyProfile{}, &hrcModel.Config{})
	seedAuditedCompany(t, db, 1)
	require.NoError(t, db.Create(&hrcModel.Config{CfgGroup: "jobs", Name: "mscms_jobs_display", Value: "1"}).Error)
	svc := &JobsService{}
	id, err := svc.CreateJob(context.Background(), 1, validJob(), &hrcModel.JobsContact{Contact: "李四"}, nil)
	require.NoError(t, err)

	var tmp hrcModel.JobsTmp
	require.NoError(t, db.First(&tmp, id).Error)
	require.Equal(t, int8(2), tmp.Audit)
	require.Equal(t, uint64(0), tmp.JobsID)

	var jobCount int64
	db.Model(&hrcModel.Jobs{}).Count(&jobCount)
	require.Zero(t, jobCount)
}

// X2（08-20）：缺省/非法配置 → 审核后显示（原「缺省=直显」用例语义翻转）
func TestJobsCreateDefaultGoesAudit(t *testing.T) {
	db := newJobsMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.JobsTmp{}, &hrcModel.JobsContact{}, &hrcModel.JobsTag{}, &hrcModel.CompanyProfile{}, &hrcModel.Config{})
	seedAuditedCompany(t, db, 1)
	// 未配置 → 默认走审核（tmp）路径
	svc := &JobsService{}
	id, err := svc.CreateJob(context.Background(), 1, validJob(), &hrcModel.JobsContact{Contact: "李四"}, nil)
	require.NoError(t, err)

	var tmp hrcModel.JobsTmp
	require.NoError(t, db.First(&tmp, id).Error)
	require.Equal(t, int8(2), tmp.Audit)

	var jobCount int64
	db.Model(&hrcModel.Jobs{}).Count(&jobCount)
	require.Zero(t, jobCount, "缺省配置下不应直接写 jobs 表")

	// 非法值同样走审核
	require.NoError(t, db.Create(&hrcModel.Config{CfgGroup: "jobs", Name: "mscms_jobs_display", Value: "9"}).Error)
	id2, err := svc.CreateJob(context.Background(), 1, validJob(), nil, nil)
	require.NoError(t, err)
	var tmp2 hrcModel.JobsTmp
	require.NoError(t, db.First(&tmp2, id2).Error)
	require.Equal(t, int8(2), tmp2.Audit)
}

func TestJobsCreateCompanyNotAudited(t *testing.T) {
	db := newJobsMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.JobsTmp{}, &hrcModel.CompanyProfile{}, &hrcModel.Config{})
	name := "甲企业"
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: 1, CompanyName: &name, Audit: 2}).Error)
	svc := &JobsService{}
	_, err := svc.CreateJob(context.Background(), 1, validJob(), nil, nil)
	require.ErrorIs(t, err, ErrCompanyAuditing)
}

func TestJobsValidate(t *testing.T) {
	newJobsMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.CompanyProfile{}, &hrcModel.Config{})
	svc := &JobsService{}

	j := validJob()
	j.JobsName = "A"
	require.ErrorIs(t, svc.validateJob(j), ErrJobNameInvalid)

	j = validJob()
	j.MinWage = 10000
	j.MaxWage = 10000
	require.ErrorIs(t, svc.validateJob(j), ErrJobWageInvalid)

	j = validJob()
	j.MinWage = 10000
	j.MaxWage = 30000
	require.ErrorIs(t, svc.validateJob(j), ErrJobWageInvalid)

	j = validJob()
	j.SubClass = 0
	require.ErrorIs(t, svc.validateJob(j), ErrJobCategoryRequired)

	j = validJob()
	j.Amount = 100
	require.ErrorIs(t, svc.validateJob(j), ErrJobAmountInvalid)

	require.NoError(t, svc.validateJob(validJob()))
}

func TestJobsAuditPassNew(t *testing.T) {
	db := newJobsMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.JobsTmp{}, &hrcModel.JobsContact{}, &hrcModel.JobsTag{}, &hrcModel.CompanyProfile{}, &hrcModel.Config{}, &hrcModel.AuditReason{})
	seedAuditedCompany(t, db, 1)
	require.NoError(t, db.Create(&hrcModel.Config{CfgGroup: "jobs", Name: "mscms_jobs_display", Value: "1"}).Error)
	svc := &JobsService{}
	id, err := svc.CreateJob(context.Background(), 1, validJob(), &hrcModel.JobsContact{Contact: "李四"}, []uint32{1})
	require.NoError(t, err)

	require.NoError(t, svc.AuditJob(context.Background(), id, 1, ""))

	var jobCount int64
	db.Model(&hrcModel.Jobs{}).Count(&jobCount)
	require.Equal(t, int64(1), jobCount)
	var tmpCount int64
	db.Model(&hrcModel.JobsTmp{}).Count(&tmpCount)
	require.Zero(t, tmpCount)

	var j hrcModel.Jobs
	require.NoError(t, db.Where("companyname = ?", "甲企业").First(&j).Error)
	require.Equal(t, int8(1), j.Audit)
	var contact hrcModel.JobsContact
	require.NoError(t, db.Where("pid = ?", j.ID).First(&contact).Error)
	require.Equal(t, "李四", contact.Contact)
}

func TestJobsAuditRejectWritesReason(t *testing.T) {
	db := newJobsMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.JobsTmp{}, &hrcModel.JobsContact{}, &hrcModel.CompanyProfile{}, &hrcModel.Config{}, &hrcModel.AuditReason{})
	seedAuditedCompany(t, db, 1)
	require.NoError(t, db.Create(&hrcModel.Config{CfgGroup: "jobs", Name: "mscms_jobs_display", Value: "1"}).Error)
	svc := &JobsService{}
	id, err := svc.CreateJob(context.Background(), 1, validJob(), nil, nil)
	require.NoError(t, err)

	require.ErrorIs(t, svc.AuditJob(context.Background(), id, 3, ""), ErrJobAuditReasonRequired)

	require.NoError(t, svc.AuditJob(context.Background(), id, 3, "职位描述不符合要求"))
	var tmp hrcModel.JobsTmp
	require.NoError(t, db.First(&tmp, id).Error)
	require.Equal(t, int8(3), tmp.Audit)

	var reason hrcModel.AuditReason
	require.NoError(t, db.Where("type = ? AND type_id = ?", hrcModel.AuditTypeJobs, id).First(&reason).Error)
	require.Equal(t, "职位描述不符合要求", reason.Reason)
}

func TestJobsUpdateAuditFlow(t *testing.T) {
	db := newJobsMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.JobsTmp{}, &hrcModel.JobsContact{}, &hrcModel.JobsTag{}, &hrcModel.CompanyProfile{}, &hrcModel.Config{}, &hrcModel.AuditReason{})
	seedAuditedCompany(t, db, 1)
	require.NoError(t, db.Create(&hrcModel.Config{CfgGroup: "jobs", Name: "mscms_jobs_display", Value: "1"}).Error)
	svc := &JobsService{}
	id, err := svc.CreateJob(context.Background(), 1, validJob(), &hrcModel.JobsContact{Contact: "旧联系人"}, nil)
	require.NoError(t, err)
	require.NoError(t, svc.AuditJob(context.Background(), id, 1, ""))
	var old hrcModel.Jobs
	require.NoError(t, db.Where("companyname = ?", "甲企业").First(&old).Error)

	edit := validJob()
	edit.JobsName = "Go 高级工程师"
	require.NoError(t, svc.UpdateJob(context.Background(), 1, old.ID, false, edit, &hrcModel.JobsContact{Contact: "新联系人"}, nil))

	var afterEdit hrcModel.Jobs
	require.NoError(t, db.First(&afterEdit, old.ID).Error)
	require.Equal(t, int8(2), afterEdit.Audit)

	var tmp hrcModel.JobsTmp
	require.NoError(t, db.Where("jobs_id = ?", old.ID).First(&tmp).Error)
	require.Equal(t, int8(2), tmp.Audit)

	require.NoError(t, svc.AuditJob(context.Background(), tmp.ID, 1, ""))
	var passed hrcModel.Jobs
	require.NoError(t, db.First(&passed, old.ID).Error)
	require.Equal(t, "Go 高级工程师", passed.JobsName)
	require.Equal(t, int8(1), passed.Audit)

	var contact hrcModel.JobsContact
	require.NoError(t, db.Where("pid = ?", old.ID).First(&contact).Error)
	require.Equal(t, "新联系人", contact.Contact)
}

func TestJobsPauseResumeDelete(t *testing.T) {
	// D1 后 DeleteJob 会回退查 jobs_tmp 表 → 需建表
	db := newJobsMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.JobsTmp{}, &hrcModel.CompanyProfile{}, &hrcModel.Config{})
	seedAuditedCompany(t, db, 1)
	// X2 后缺省走审核（tmp）路径，本用例操作 jobs 表行 → 显式直显
	require.NoError(t, db.Create(&hrcModel.Config{CfgGroup: "jobs", Name: "mscms_jobs_display", Value: "2"}).Error)
	svc := &JobsService{}
	id, err := svc.CreateJob(context.Background(), 1, validJob(), nil, nil)
	require.NoError(t, err)

	require.NoError(t, svc.PauseJob(context.Background(), 1, id))
	var j hrcModel.Jobs
	require.NoError(t, db.First(&j, id).Error)
	require.Equal(t, int8(2), j.Display)

	require.NoError(t, svc.ResumeJob(context.Background(), 1, id))
	require.NoError(t, db.First(&j, id).Error)
	require.Equal(t, int8(1), j.Display)

	require.NoError(t, svc.DeleteJob(context.Background(), 1, id, false))
	require.NoError(t, db.First(&j, id).Error)
	require.NotZero(t, j.DeletedAt)
	require.ErrorIs(t, svc.DeleteJob(context.Background(), 1, id, false), ErrJobNotFound)
	require.ErrorIs(t, svc.PauseJob(context.Background(), 2, id), ErrJobNotFound)
}

func TestJobsList(t *testing.T) {
	db := newJobsMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.JobsTmp{}, &hrcModel.CompanyProfile{}, &hrcModel.Config{})
	seedAuditedCompany(t, db, 1)
	require.NoError(t, db.Create(&hrcModel.Config{CfgGroup: "jobs", Name: "mscms_jobs_display", Value: "1"}).Error)
	svc := &JobsService{}
	_, err := svc.CreateJob(context.Background(), 1, validJob(), nil, nil)
	require.NoError(t, err)
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 1, JobsName: "直接发布的职位", Audit: 1, Display: 1}}).Error)

	list, total, err := svc.ListJobs(context.Background(), 1, request.PageInfo{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, list, 2)
}

// TestJobsGetJobIDOverlap jobs 与 jobs_tmp 两表 id 各自自增重叠：pending 参数区分表
func TestJobsGetJobIDOverlap(t *testing.T) {
	db := newJobsMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.JobsTmp{}, &hrcModel.JobsContact{}, &hrcModel.JobsTag{}, &hrcModel.AuditReason{}, &hrcModel.CompanyProfile{}, &hrcModel.Config{})
	seedAuditedCompany(t, db, 1)
	// jobs.id=1 与 tmp.id=1 并存（两表独立自增）
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 1, JobsName: "正式职位", Audit: 1, Display: 1}}).Error)
	require.NoError(t, db.Create(&hrcModel.JobsTmp{JobsBase: hrcModel.JobsBase{UID: 1, JobsName: "待审职位", Audit: 2}}).Error)
	require.NoError(t, db.Create(&hrcModel.JobsContact{PID: 1, Contact: "待审联系人"}).Error)

	svc := &JobsService{}
	// pending=false → 命中 jobs 表
	d1, err := svc.GetJob(context.Background(), 1, 1, false)
	require.NoError(t, err)
	require.Equal(t, "正式职位", d1.JobsName)
	require.False(t, d1.Pending)

	// pending=true → 命中 jobs_tmp 表（同一 id）
	d2, err := svc.GetJob(context.Background(), 1, 1, true)
	require.NoError(t, err)
	require.Equal(t, "待审职位", d2.JobsName)
	require.True(t, d2.Pending)
}

// TestJobsListEditDedup 编辑中（jobs audit=2 + 对应 tmp）只由 tmp 行代表，不重复展示
func TestJobsListEditDedup(t *testing.T) {
	db := newJobsMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.JobsTmp{}, &hrcModel.CompanyProfile{}, &hrcModel.Config{})
	seedAuditedCompany(t, db, 1)
	require.NoError(t, db.Create(&hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 1, JobsName: "编辑中职位", Audit: 2, Display: 1}}).Error)
	var j hrcModel.Jobs
	require.NoError(t, db.Where("jobs_name = ?", "编辑中职位").First(&j).Error)
	require.NoError(t, db.Create(&hrcModel.JobsTmp{JobsID: j.ID, JobsBase: hrcModel.JobsBase{UID: 1, JobsName: "编辑中职位新内容", Audit: 2}}).Error)

	svc := &JobsService{}
	list, total, err := svc.ListJobs(context.Background(), 1, request.PageInfo{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.True(t, list[0].Pending)
	require.Equal(t, "编辑中职位新内容", list[0].JobsName)
}

// D1（08-21）：tmp-only 职位全链路——发布(display=1)→拒→编辑重提(jobs 未命中回退 tmp)→通过→jobs 表出现新行
func TestJobsTmpOnlyResubmitAfterReject(t *testing.T) {
	db := newJobsMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.JobsTmp{}, &hrcModel.JobsContact{}, &hrcModel.JobsTag{}, &hrcModel.CompanyProfile{}, &hrcModel.Config{}, &hrcModel.AuditReason{})
	seedAuditedCompany(t, db, 1)
	require.NoError(t, db.Create(&hrcModel.Config{CfgGroup: "jobs", Name: "mscms_jobs_display", Value: "1"}).Error)
	svc := &JobsService{}

	// 发布 → tmp-only（jobs 表无行）
	id, err := svc.CreateJob(context.Background(), 1, validJob(), &hrcModel.JobsContact{Contact: "旧联系人"}, []uint32{9})
	require.NoError(t, err)
	var jobCount int64
	db.Model(&hrcModel.Jobs{}).Count(&jobCount)
	require.Zero(t, jobCount)

	// 审核不通过
	require.NoError(t, svc.AuditJob(context.Background(), id, 3, "不符合要求"))
	var tmp hrcModel.JobsTmp
	require.NoError(t, db.First(&tmp, id).Error)
	require.Equal(t, int8(3), tmp.Audit)

	// 编辑重提（不传 pending → jobs 未命中 → 回退 tmp 原地更新）
	edit := validJob()
	edit.JobsName = "Go 高级工程师"
	require.NoError(t, svc.UpdateJob(context.Background(), 1, id, false, edit, &hrcModel.JobsContact{Contact: "新联系人"}, []uint32{7, 8}))
	require.NoError(t, db.First(&tmp, id).Error)
	require.Equal(t, int8(2), tmp.Audit, "重提后应回到待审")
	require.Equal(t, "Go 高级工程师", tmp.JobsName)
	require.Equal(t, uint64(0), tmp.JobsID, "tmp-only 职位应保持 JobsID=0")

	// contact/tags 重挂 tmp.id（旧行删除、新行写入）
	var contact hrcModel.JobsContact
	require.NoError(t, db.Where("pid = ?", id).First(&contact).Error)
	require.Equal(t, "新联系人", contact.Contact)
	var tagCount int64
	db.Model(&hrcModel.JobsTag{}).Where("pid = ?", id).Count(&tagCount)
	require.Equal(t, int64(2), tagCount)

	// 审核通过 → jobs 表出现新行（新 id），tmp 删除，contact 重键到 jobs id
	require.NoError(t, svc.AuditJob(context.Background(), id, 1, ""))
	db.Model(&hrcModel.Jobs{}).Count(&jobCount)
	require.Equal(t, int64(1), jobCount)
	var tmpCount int64
	db.Model(&hrcModel.JobsTmp{}).Where("deleted_at = 0").Count(&tmpCount)
	require.Zero(t, tmpCount)
	var j hrcModel.Jobs
	require.NoError(t, db.Where("jobs_name = ?", "Go 高级工程师").First(&j).Error)
	require.Equal(t, int8(1), j.Audit)
	var jc hrcModel.JobsContact
	require.NoError(t, db.Where("pid = ?", j.ID).First(&jc).Error)
	require.Equal(t, "新联系人", jc.Contact)
	var jtc int64
	db.Model(&hrcModel.JobsTag{}).Where("pid = ?", j.ID).Count(&jtc)
	require.Equal(t, int64(2), jtc)
}

// D1：显式 pending=true 走 tmp 原地更新 + uid 隔离 + 非 2/3 状态防御
func TestJobsTmpUpdateExplicitPending(t *testing.T) {
	db := newJobsMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.JobsTmp{}, &hrcModel.JobsContact{}, &hrcModel.JobsTag{}, &hrcModel.CompanyProfile{}, &hrcModel.Config{})
	seedAuditedCompany(t, db, 1)
	seedAuditedCompanyNamed(t, db, 2, "乙企业")
	require.NoError(t, db.Create(&hrcModel.Config{CfgGroup: "jobs", Name: "mscms_jobs_display", Value: "1"}).Error)
	svc := &JobsService{}
	id, err := svc.CreateJob(context.Background(), 1, validJob(), nil, nil)
	require.NoError(t, err)

	// 显式 pending=true → tmp 原地更新
	edit := validJob()
	edit.JobsName = "显式指定 tmp 行"
	require.NoError(t, svc.UpdateJob(context.Background(), 1, id, true, edit, nil, nil))
	var tmp hrcModel.JobsTmp
	require.NoError(t, db.First(&tmp, id).Error)
	require.Equal(t, int8(2), tmp.Audit)
	require.Equal(t, "显式指定 tmp 行", tmp.JobsName)

	// uid 隔离：另一企业（含回退路径）均命中不到
	require.ErrorIs(t, svc.UpdateJob(context.Background(), 2, id, true, validJob(), nil, nil), ErrJobNotFound)
	require.ErrorIs(t, svc.UpdateJob(context.Background(), 2, id, false, validJob(), nil, nil), ErrJobNotFound)

	// 防御：audit 非 2/3 的 tmp 行不可编辑（正常流程不存在，1=通过防御）
	require.NoError(t, db.Model(&hrcModel.JobsTmp{}).Where("id = ?", id).Update("audit", 1).Error)
	require.ErrorIs(t, svc.UpdateJob(context.Background(), 1, id, true, validJob(), nil, nil), ErrJobAuditState)
}

// D1：删除 tmp-only 职位（回退路径 + 显式 pending + 重复删/隔离）
func TestJobsTmpOnlyDelete(t *testing.T) {
	db := newJobsMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.JobsTmp{}, &hrcModel.CompanyProfile{}, &hrcModel.Config{})
	seedAuditedCompany(t, db, 1)
	seedAuditedCompanyNamed(t, db, 2, "乙企业")
	require.NoError(t, db.Create(&hrcModel.Config{CfgGroup: "jobs", Name: "mscms_jobs_display", Value: "1"}).Error)
	svc := &JobsService{}

	id, err := svc.CreateJob(context.Background(), 1, validJob(), nil, nil)
	require.NoError(t, err)
	id2, err := svc.CreateJob(context.Background(), 1, validJob(), nil, nil)
	require.NoError(t, err)

	// 回退路径：jobs 未命中 → tmp 软删
	require.NoError(t, svc.DeleteJob(context.Background(), 1, id, false))
	var tmp hrcModel.JobsTmp
	require.NoError(t, db.First(&tmp, id).Error)
	require.NotZero(t, tmp.DeletedAt)

	// 显式 pending 路径（tmp 主键已被上一查询置位，必须换新变量，否则 gorm 附加旧主键条件）
	require.NoError(t, svc.DeleteJob(context.Background(), 1, id2, true))
	var tmp2 hrcModel.JobsTmp
	require.NoError(t, db.First(&tmp2, id2).Error)
	require.NotZero(t, tmp2.DeletedAt)

	// 已删/他人 → ErrJobNotFound（两路径都隔离）
	require.ErrorIs(t, svc.DeleteJob(context.Background(), 1, id, false), ErrJobNotFound)
	require.ErrorIs(t, svc.DeleteJob(context.Background(), 1, id, true), ErrJobNotFound)
	require.ErrorIs(t, svc.DeleteJob(context.Background(), 2, id, false), ErrJobNotFound)

	// D1 补充：软删后的 tmp 行不再出现在企业列表
	list, total, err := svc.ListJobs(context.Background(), 1, request.PageInfo{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Zero(t, total, "两个 tmp-only 职位均已软删，列表应为空")
	require.Len(t, list, 0)
}
