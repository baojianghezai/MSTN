package hrc

import (
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
)

// 完善度计算（04 §2.2 字段权重：基本信息 20 + 教育 20 + 工作 25 + 期望 15 + 其他 20 = 100）
// #57 完善度详情与主表 complete_percent 同源：创建/编辑（#48/#51）与 #57 均调用 CalculateCompleteness。
//
// 字段级细则（04 §2.2 未逐字段拆分，本表为落地实现，docs/api/04-resume.md 同步）：
//
//	基本信息 20：姓名 5 / 性别 2 / 出生年 3 / 最高学历 5 / 专业 3 / 联系电话 2
//	教育经历 20：ms_resume_education 至少 1 条 → 20
//	工作     25：ms_resume_work 至少 1 条 → 15；ms_resume_project 至少 1 条 → 10
//	期望     15：期望地区 5 / 期望薪资 5 / 期望职位 5
//	其他     20：语言 5 / 培训 5 / 证书 5 / 自我评价 5
//	（「其他」类设计含「作品」项；ms_resume_img 后置，其分值暂由「自我评价」代位，作品上线后重拆）

// CompletenessCategory 单一完善度类别
type CompletenessCategory struct {
	Key   string `json:"key"`   // basic/education/work/intention/other
	Name  string `json:"name"`  // 类别中文名
	Max   int    `json:"max"`   // 类别满分
	Score int    `json:"score"` // 当前得分
}

// CompletenessResult 完善度计算结果（#57 返回体；Percent 与主表 complete_percent 同源）
type CompletenessResult struct {
	Percent    int                    `json:"percent"`
	Categories []CompletenessCategory `json:"categories"`
	Missing    []string               `json:"missing"` // 缺失项名称清单（中文，前端展示「还差什么」）
}

type completenessRule struct {
	catKey string
	label  string
	points int
	filled func(r *hrcModel.Resume, subs *ResumeSubTables) bool
}

var completenessRules = []completenessRule{
	{"basic", "姓名", 5, func(r *hrcModel.Resume, _ *ResumeSubTables) bool { return r.FullName != "" }},
	{"basic", "性别", 2, func(r *hrcModel.Resume, _ *ResumeSubTables) bool { return r.Sex != 0 }},
	{"basic", "出生年", 3, func(r *hrcModel.Resume, _ *ResumeSubTables) bool { return r.Birthdate != 0 }},
	{"basic", "最高学历", 5, func(r *hrcModel.Resume, _ *ResumeSubTables) bool { return r.Education != 0 }},
	{"basic", "专业", 3, func(r *hrcModel.Resume, _ *ResumeSubTables) bool { return r.Major != 0 }},
	{"basic", "联系电话", 2, func(r *hrcModel.Resume, _ *ResumeSubTables) bool { return r.Telephone != "" }},
	{"education", "教育经历", 20, func(_ *hrcModel.Resume, subs *ResumeSubTables) bool { return len(subs.Education) > 0 }},
	{"work", "工作经历", 15, func(_ *hrcModel.Resume, subs *ResumeSubTables) bool { return len(subs.Work) > 0 }},
	{"work", "项目经历", 10, func(_ *hrcModel.Resume, subs *ResumeSubTables) bool { return len(subs.Project) > 0 }},
	{"intention", "期望地区", 5, func(r *hrcModel.Resume, _ *ResumeSubTables) bool { return r.District != "" }},
	{"intention", "期望薪资", 5, func(r *hrcModel.Resume, _ *ResumeSubTables) bool { return r.Wage != 0 }},
	{"intention", "期望职位", 5, func(r *hrcModel.Resume, _ *ResumeSubTables) bool { return r.IntentionJobs != "" }},
	{"other", "语言能力", 5, func(_ *hrcModel.Resume, subs *ResumeSubTables) bool { return len(subs.Language) > 0 }},
	{"other", "培训经历", 5, func(_ *hrcModel.Resume, subs *ResumeSubTables) bool { return len(subs.Training) > 0 }},
	{"other", "资格证书", 5, func(_ *hrcModel.Resume, subs *ResumeSubTables) bool { return len(subs.Credent) > 0 }},
	{"other", "自我评价", 5, func(r *hrcModel.Resume, _ *ResumeSubTables) bool { return r.Specialty != "" }},
}

var completenessCategoryNames = []struct{ key, name string }{
	{"basic", "基本信息"},
	{"education", "教育经历"},
	{"work", "工作"},
	{"intention", "期望"},
	{"other", "其他"},
}

// CalculateCompleteness 计算完善度（缺一项扣对应分；子表类「至少 1 条」得满分）
func CalculateCompleteness(resume *hrcModel.Resume, subs *ResumeSubTables) CompletenessResult {
	subs = subs.safe()
	result := CompletenessResult{
		Categories: make([]CompletenessCategory, 0, len(completenessCategoryNames)),
		Missing:    []string{},
	}
	scores := make(map[string]int, len(completenessCategoryNames))
	maxes := make(map[string]int, len(completenessCategoryNames))
	for _, c := range completenessCategoryNames {
		scores[c.key] = 0
		maxes[c.key] = 0
	}
	for _, rule := range completenessRules {
		maxes[rule.catKey] += rule.points
		if rule.filled(resume, subs) {
			scores[rule.catKey] += rule.points
		} else {
			result.Missing = append(result.Missing, rule.label)
		}
	}
	for _, c := range completenessCategoryNames {
		result.Percent += scores[c.key]
		result.Categories = append(result.Categories, CompletenessCategory{
			Key:   c.key,
			Name:  c.name,
			Max:   maxes[c.key],
			Score: scores[c.key],
		})
	}
	return result
}

// buildSearchKeys 拼装搜索冗余字段 key_full/key_precise（04 §2.1：一期同步维护，二期迁 ES）
// 参照 v6 ResumeModel.check_resume：key_full=全字段文本拼装，key_precise=期望职位+各段职位
func buildSearchKeys(resume *hrcModel.Resume, subs *ResumeSubTables) (string, string) {
	subs = subs.safe()
	full := resume.IntentionJobs + resume.EducationCN + resume.Specialty
	for _, e := range subs.Education {
		full += e.School + e.Speciality
	}
	for _, w := range subs.Work {
		full += w.CompanyName + w.Jobs + w.Achievements
	}
	for _, t := range subs.Training {
		full += t.Agency + t.Course + t.Description
	}
	for _, l := range subs.Language {
		full += l.LanguageCN
	}
	for _, c := range subs.Credent {
		full += c.Name
	}
	precise := resume.IntentionJobs
	for _, w := range subs.Work {
		precise += w.Jobs
	}
	return full, precise
}
