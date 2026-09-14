package hrc

import (
	"context"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func resumeMain() *hrcModel.Resume {
	return &hrcModel.Resume{
		Title:        "后端开发工程师",
		FullName:     "张三",
		Sex:          1,
		SexCN:        "男",
		Education:    4,
		EducationCN:  "本科",
		Experience:   3,
		ExperienceCN: "3-5年",
	}
}

// resumeComplete 完善度满分简历（04 §2.2 全部 16 项命中）
func resumeComplete() *hrcModel.Resume {
	r := resumeMain()
	r.Birthdate = 1995
	r.Major = 7
	r.MajorCN = "软件工程"
	r.Telephone = "13800138000"
	r.District = "4401"
	r.Wage = 8
	r.IntentionJobs = "Go 开发"
	r.Specialty = "自我评价"
	return r
}

func projects(n int) []hrcModel.ResumeProject {
	out := make([]hrcModel.ResumeProject, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, hrcModel.ResumeProject{
			StartYear:   2023,
			StartMonth:  1,
			EndYear:     2024,
			EndMonth:    6,
			ToDate:      0,
			ProjectName: "人才网重构",
			Role:        "后端",
			Description: "负责简历模块",
		})
	}
	return out
}

func subsWithProjects(n int) *ResumeSubTables {
	return &ResumeSubTables{Project: projects(n)}
}

// subsFull 六子表各 1 条（子表类完善度项全命中）
func subsFull() *ResumeSubTables {
	return &ResumeSubTables{
		Education: []hrcModel.ResumeEducation{{StartYear: 2015, EndYear: 2019, School: "中山大学", Speciality: "软件工程", Education: 4, EducationCN: "本科"}},
		Work:      []hrcModel.ResumeWork{{StartYear: 2019, StartMonth: 7, ToDate: 1, CompanyName: "某某科技", Jobs: "Go 工程师", Achievements: "负责简历模块"}},
		Language:  []hrcModel.ResumeLanguage{{Language: 1, LanguageCN: "英语", Level: 4, LevelCN: "流利"}},
		Training:  []hrcModel.ResumeTraining{{StartYear: 2020, Agency: "某某机构", Course: "Go 高级"}},
		Credent:   []hrcModel.ResumeCredent{{Name: "软件设计师", Year: 2020, Month: 6}},
		Project:   projects(1),
	}
}

func resumeDB(t *testing.T) *gorm.DB {
	return testutil.NewMemoryDB(t,
		&hrcModel.Resume{},
		&hrcModel.ResumeProject{},
		&hrcModel.ResumeEducation{},
		&hrcModel.ResumeWork{},
		&hrcModel.ResumeLanguage{},
		&hrcModel.ResumeTraining{},
		&hrcModel.ResumeCredent{},
	)
}

func TestResumeCreateWithProjects(t *testing.T) {
	db := resumeDB(t)
	svc := &ResumeService{}

	id, err := svc.CreateResume(context.Background(), 1, resumeMain(), subsWithProjects(2))
	require.NoError(t, err)
	require.NotZero(t, id)

	var r hrcModel.Resume
	require.NoError(t, db.First(&r, id).Error)
	require.Equal(t, uint64(1), r.UID)
	require.Equal(t, int8(1), r.Audit)
	require.Equal(t, int8(1), r.Display)
	require.NotZero(t, r.AddTime)

	var list []hrcModel.ResumeProject
	require.NoError(t, db.Where("pid = ?", id).Order("id asc").Find(&list).Error)
	require.Len(t, list, 2)
	require.Equal(t, id, list[0].PID)
	require.Equal(t, uint64(1), list[0].UID)
	require.Equal(t, "人才网重构", list[0].ProjectName)
}

func TestResumeCreateProjectLimit(t *testing.T) {
	resumeDB(t)
	svc := &ResumeService{}

	_, err := svc.CreateResume(context.Background(), 1, resumeMain(), subsWithProjects(7))
	require.ErrorIs(t, err, ErrResumeProjectLimit)
}

func TestResumeCreateFirstIsDefault(t *testing.T) {
	db := resumeDB(t)
	svc := &ResumeService{}

	id1, err := svc.CreateResume(context.Background(), 1, resumeMain(), nil)
	require.NoError(t, err)
	id2, err := svc.CreateResume(context.Background(), 1, resumeMain(), nil)
	require.NoError(t, err)

	var r1, r2 hrcModel.Resume
	require.NoError(t, db.First(&r1, id1).Error)
	require.NoError(t, db.First(&r2, id2).Error)
	require.Equal(t, int8(1), r1.Def, "首份简历应标记默认")
	require.Equal(t, int8(0), r2.Def, "后续简历默认不标记")
}

// 六子表全量写入 + pid/uid 盖章 + 完善度/搜索索引同事务落主表
func TestResumeCreateWithAllSubTables(t *testing.T) {
	db := resumeDB(t)
	svc := &ResumeService{}

	id, err := svc.CreateResume(context.Background(), 1, resumeComplete(), subsFull())
	require.NoError(t, err)

	for _, model := range []interface{}{
		&hrcModel.ResumeEducation{}, &hrcModel.ResumeWork{}, &hrcModel.ResumeLanguage{},
		&hrcModel.ResumeTraining{}, &hrcModel.ResumeCredent{}, &hrcModel.ResumeProject{},
	} {
		var count int64
		require.NoError(t, db.Model(model).Where("pid = ? AND uid = ?", id, 1).Count(&count).Error)
		require.Equal(t, int64(1), count, "子表应写入且 pid/uid 盖章正确: %T", model)
	}

	var r hrcModel.Resume
	require.NoError(t, db.First(&r, id).Error)
	require.Equal(t, int8(100), r.CompletePercent, "满分简历完善度应为 100")
	require.Contains(t, r.KeyFull, "中山大学")
	require.Contains(t, r.KeyPrecise, "Go 工程师")
}

// 全量替换幂等：同参数编辑两次，子表条数不变
func TestResumeUpdateReplacesAllSubTablesIdempotent(t *testing.T) {
	db := resumeDB(t)
	svc := &ResumeService{}

	id, err := svc.CreateResume(context.Background(), 1, resumeComplete(), subsFull())
	require.NoError(t, err)

	// 第一轮：education 换成 3 条、work 清空、其余保留
	subs := subsFull()
	subs.Education = []hrcModel.ResumeEducation{
		{StartYear: 2010, EndYear: 2014, School: "A 大学"},
		{StartYear: 2014, EndYear: 2017, School: "B 大学"},
		{StartYear: 2017, EndYear: 2020, School: "C 大学", ToDate: 0},
	}
	subs.Work = nil
	require.NoError(t, svc.UpdateResume(context.Background(), 1, id, resumeMain(), subs))

	counts := func() (edu, work int64) {
		require.NoError(t, db.Model(&hrcModel.ResumeEducation{}).Where("pid = ?", id).Count(&edu).Error)
		require.NoError(t, db.Model(&hrcModel.ResumeWork{}).Where("pid = ?", id).Count(&work).Error)
		return
	}
	edu, work := counts()
	require.Equal(t, int64(3), edu, "旧教育经历应删净、新 3 条写入")
	require.Equal(t, int64(0), work, "提交空数组应清空工作")

	// 第二轮：同参数重放 → 幂等，条数不翻倍
	require.NoError(t, svc.UpdateResume(context.Background(), 1, id, resumeMain(), subs))
	edu2, work2 := counts()
	require.Equal(t, int64(3), edu2, "全量替换应幂等")
	require.Equal(t, int64(0), work2)
}

func TestResumeUpdateOwnership(t *testing.T) {
	resumeDB(t)
	svc := &ResumeService{}

	id, err := svc.CreateResume(context.Background(), 1, resumeMain(), nil)
	require.NoError(t, err)

	// 非归属会员编辑 → 简历不存在
	err = svc.UpdateResume(context.Background(), 2, id, resumeMain(), nil)
	require.ErrorIs(t, err, ErrResumeNotFound)
}

func TestResumeUpdateProjectLimit(t *testing.T) {
	resumeDB(t)
	svc := &ResumeService{}

	id, err := svc.CreateResume(context.Background(), 1, resumeMain(), nil)
	require.NoError(t, err)

	require.NoError(t, svc.UpdateResume(context.Background(), 1, id, resumeMain(), subsWithProjects(6))) // 恰好 6 条合法
	require.ErrorIs(t, svc.UpdateResume(context.Background(), 1, id, resumeMain(), subsWithProjects(7)), ErrResumeProjectLimit)
}

func TestResumeGetDetail(t *testing.T) {
	resumeDB(t)
	svc := &ResumeService{}

	id, err := svc.CreateResume(context.Background(), 1, resumeComplete(), subsFull())
	require.NoError(t, err)

	resume, subs, err := svc.GetResume(context.Background(), 1, id)
	require.NoError(t, err)
	require.Equal(t, "张三", resume.FullName)
	require.Len(t, subs.Project, 1)
	require.Len(t, subs.Education, 1)
	require.Len(t, subs.Work, 1)
	require.Len(t, subs.Language, 1)
	require.Len(t, subs.Training, 1)
	require.Len(t, subs.Credent, 1)

	// 非归属会员读取 → 不存在
	_, _, err = svc.GetResume(context.Background(), 2, id)
	require.ErrorIs(t, err, ErrResumeNotFound)
}

// uid 隔离：A 建 2 份、B 建 1 份；列表互不可见；分页正常
func TestResumeListUidIsolation(t *testing.T) {
	db := resumeDB(t)
	svc := &ResumeService{}

	_, err := svc.CreateResume(context.Background(), 1, resumeMain(), nil)
	require.NoError(t, err)
	_, err = svc.CreateResume(context.Background(), 1, resumeMain(), nil)
	require.NoError(t, err)
	_, err = svc.CreateResume(context.Background(), 2, resumeMain(), nil)
	require.NoError(t, err)

	info := request.PageInfo{Page: 1, PageSize: 10}
	list, total, err := svc.ListResumes(context.Background(), 1, info)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, list, 2)
	// 列表只取轻量字段（不含 uid），归属隔离按库中主键回查验证
	for _, r := range list {
		var m hrcModel.Resume
		require.NoError(t, db.First(&m, r.ID).Error)
		require.Equal(t, uint64(1), m.UID)
	}

	list2, total2, err := svc.ListResumes(context.Background(), 2, info)
	require.NoError(t, err)
	require.Equal(t, int64(1), total2)
	require.Len(t, list2, 1)
}

// 软删：主表软删、子表保留；列表/详情/操作均不可见；删默认简历升级最近一份
func TestResumeDeleteSoft(t *testing.T) {
	db := resumeDB(t)
	svc := &ResumeService{}

	id1, err := svc.CreateResume(context.Background(), 1, resumeComplete(), subsFull())
	require.NoError(t, err)
	id2, err := svc.CreateResume(context.Background(), 1, resumeMain(), subsWithProjects(1))
	require.NoError(t, err)

	require.NoError(t, svc.DeleteResume(context.Background(), 1, id1)) // 删的是默认简历（def=1）

	var r1 hrcModel.Resume
	require.NoError(t, db.First(&r1, id1).Error)
	require.NotNil(t, r1.DeletedAt, "主表应软删")
	require.Equal(t, int8(0), r1.Def, "软删行 def 应清零（Issue B）")

	// 子表保留（合规审计）
	var eduCount int64
	require.NoError(t, db.Model(&hrcModel.ResumeEducation{}).Where("pid = ?", id1).Count(&eduCount).Error)
	require.Equal(t, int64(1), eduCount, "子表应保留")

	// 列表只剩 1 条，且 id2 升为默认
	list, total, err := svc.ListResumes(context.Background(), 1, request.PageInfo{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, id2, list[0].ID)
	var r2 hrcModel.Resume
	require.NoError(t, db.First(&r2, id2).Error)
	require.Equal(t, int8(1), r2.Def, "删默认简历后最近一份应升为默认")

	// 详情/编辑/删除软删简历 → 不存在
	_, _, err = svc.GetResume(context.Background(), 1, id1)
	require.ErrorIs(t, err, ErrResumeNotFound)
	require.ErrorIs(t, svc.UpdateResume(context.Background(), 1, id1, resumeMain(), nil), ErrResumeNotFound)
	require.ErrorIs(t, svc.DeleteResume(context.Background(), 1, id1), ErrResumeNotFound)

	// 非本人删除 → 不存在
	require.ErrorIs(t, svc.DeleteResume(context.Background(), 2, id2), ErrResumeNotFound)
}

// def 互斥：两份简历先后设默认，始终仅一份 def=1
func TestResumeSetDefaultMutualExclusion(t *testing.T) {
	db := resumeDB(t)
	svc := &ResumeService{}

	id1, err := svc.CreateResume(context.Background(), 1, resumeMain(), nil)
	require.NoError(t, err)
	id2, err := svc.CreateResume(context.Background(), 1, resumeMain(), nil)
	require.NoError(t, err)

	require.NoError(t, svc.SetDefault(context.Background(), 1, id2))
	var defCount int64
	require.NoError(t, db.Model(&hrcModel.Resume{}).Where("uid = ? AND deleted_at IS NULL AND def = 1", 1).Count(&defCount).Error)
	require.Equal(t, int64(1), defCount, "同 uid 内 def 应互斥")

	var r2 hrcModel.Resume
	require.NoError(t, db.First(&r2, id2).Error)
	require.Equal(t, int8(1), r2.Def)

	// 非本人设置 → 不存在
	require.ErrorIs(t, svc.SetDefault(context.Background(), 2, id1), ErrResumeNotFound)
}

// #54 公开/隐藏：1/2 合法、其他值拒绝
func TestResumeSetDisplay(t *testing.T) {
	db := resumeDB(t)
	svc := &ResumeService{}

	id, err := svc.CreateResume(context.Background(), 1, resumeMain(), nil)
	require.NoError(t, err)

	require.NoError(t, svc.SetDisplay(context.Background(), 1, id, 2))
	var r hrcModel.Resume
	require.NoError(t, db.First(&r, id).Error)
	require.Equal(t, int8(2), r.Display)

	require.NoError(t, svc.SetDisplay(context.Background(), 1, id, 1))
	require.NoError(t, db.First(&r, id).Error)
	require.Equal(t, int8(1), r.Display)

	require.ErrorIs(t, svc.SetDisplay(context.Background(), 1, id, 3), ErrResumeDisplayValue)
	require.ErrorIs(t, svc.SetDisplay(context.Background(), 2, id, 2), ErrResumeNotFound)
}

// #57 完善度：空简历 0 分且缺失 16 项；满分简历 100；缺项按权重扣分；与主表 complete_percent 同源
func TestResumeCompleteness(t *testing.T) {
	// 空简历 → 0 分，16 项全缺
	empty := CalculateCompleteness(&hrcModel.Resume{}, nil)
	require.Equal(t, 0, empty.Percent)
	require.Len(t, empty.Missing, 16)
	requireCompletenessWeights(t, empty)

	// 满分简历（主表 10 项 + 6 子表类全命中）
	full := CalculateCompleteness(resumeComplete(), subsFull())
	require.Equal(t, 100, full.Percent)
	require.Empty(t, full.Missing)
	requireCompletenessWeights(t, full)

	// 只填姓名 + 教育经历 → 5 + 20 = 25；缺项 15 个
	partial := CalculateCompleteness(&hrcModel.Resume{FullName: "张三"}, &ResumeSubTables{Education: []hrcModel.ResumeEducation{{School: "X"}}})
	require.Equal(t, 25, partial.Percent)
	require.Len(t, partial.Missing, 14)
	require.NotContains(t, partial.Missing, "姓名")
	require.NotContains(t, partial.Missing, "教育经历")
}

// 权重表与 04 §2.2 一致：基本信息 20 / 教育 20 / 工作 25 / 期望 15 / 其他 20
func requireCompletenessWeights(t *testing.T, r CompletenessResult) {
	want := map[string]int{"basic": 20, "education": 20, "work": 25, "intention": 15, "other": 20}
	require.Len(t, r.Categories, len(want))
	for _, c := range r.Categories {
		require.Equal(t, want[c.Key], c.Max, "类别 %s 满分应等于 04 §2.2 权重", c.Key)
		require.LessOrEqual(t, c.Score, c.Max)
	}
}

// #48 创建落库的 complete_percent 与 #57 同源：编辑后主表分值随子表变化
func TestResumeCompletenessSyncWithMainTable(t *testing.T) {
	db := resumeDB(t)
	svc := &ResumeService{}

	// 只带主表部分字段创建
	id, err := svc.CreateResume(context.Background(), 1, &hrcModel.Resume{FullName: "张三"}, nil)
	require.NoError(t, err)
	var r hrcModel.Resume
	require.NoError(t, db.First(&r, id).Error)
	require.Equal(t, int8(5), r.CompletePercent, "仅姓名应为 5 分")

	// 补全子表后编辑 → 主表分值同步上升
	require.NoError(t, svc.UpdateResume(context.Background(), 1, id, resumeComplete(), subsFull()))
	require.NoError(t, db.First(&r, id).Error)
	require.Equal(t, int8(100), r.CompletePercent)

	// #57 与主表同源
	result, err := svc.GetCompleteness(context.Background(), 1, id)
	require.NoError(t, err)
	require.Equal(t, int(r.CompletePercent), result.Percent)

	// 非本人查完善度 → 不存在
	_, err = svc.GetCompleteness(context.Background(), 2, id)
	require.ErrorIs(t, err, ErrResumeNotFound)
}
