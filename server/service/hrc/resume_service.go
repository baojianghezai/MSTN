package hrc

import (
	"context"
	"errors"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"gorm.io/gorm"
)

var (
	ErrResumeNotFound     = errors.New("简历不存在")
	ErrResumeProjectLimit = errors.New("项目经历最多 6 条")
	ErrResumeDisplayValue = errors.New("公开状态仅支持 1=公开 2=不公开")
)

// ResumeSubTables 简历子表批量结构（04 §3.1：随 #48 创建 / #51 编辑 主表同事务全量替换）
// 九张子表：education/work/language/training/credent/project/skill/portfolio/student_leader
// project 限 6 条，其余一期不限
type ResumeSubTables struct {
	Education     []hrcModel.ResumeEducation
	Work          []hrcModel.ResumeWork
	Language      []hrcModel.ResumeLanguage
	Training      []hrcModel.ResumeTraining
	Credent       []hrcModel.ResumeCredent
	Project       []hrcModel.ResumeProject
	Skill         []hrcModel.ResumeSkill
	Portfolio     []hrcModel.ResumePortfolio
	StudentLeader []hrcModel.ResumeStudentLeader
}

func (s *ResumeSubTables) safe() *ResumeSubTables {
	if s == nil {
		return &ResumeSubTables{}
	}
	return s
}

// ResumeService 简历服务（04 §3.1 创建/编辑全量替换；#49 列表 / #52 软删 / #54 公开隐藏 / #55 设默认 / #57 完善度）
type ResumeService struct{}

// CreateResume 创建简历（主表 + 6 子表同一事务；项目经历限 6 条；首份简历标记 def=1；
// 完善度与搜索索引按提交数据计算并同事务写入，04 §3.1 步骤 4）
func (s *ResumeService) CreateResume(ctx context.Context, uid uint64, resume *hrcModel.Resume, subs *ResumeSubTables) (uint64, error) {
	subs = subs.safe()
	if len(subs.Project) > hrcModel.ResumeProjectMax {
		return 0, ErrResumeProjectLimit
	}
	now := hrcModel.Now()
	resume.ID = 0
	resume.UID = uid
	resume.Display = 1     // 默认公开（#54 切换）
	resume.Audit = 1       // 简历默认通过（审核配置策略随 M3/C6 细化）
	resume.DisplayName = 1 // 默认显示姓名
	resume.DeletedAt = nil
	resume.AddTime = now
	resume.Refreshtime = &now
	resume.Click = 1

	var id uint64
	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// def 逻辑：本人首份简历标记默认（04 §2.1 def；投递默认取 def desc）
		var count int64
		if err := tx.Model(&hrcModel.Resume{}).Where("uid = ? AND deleted_at IS NULL", uid).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			resume.Def = 1
		} else {
			resume.Def = 0
		}
		if err := tx.Create(resume).Error; err != nil {
			return err
		}
		id = resume.ID
		if err := s.writeSubTables(tx, id, uid, subs); err != nil {
			return err
		}
		return s.fillDerivedFields(tx, id, resume, subs)
	})
	return id, err
}

// UpdateResume 编辑简历（归属校验 + 6 子表全量替换删旧插新 + 完善度/搜索索引重算，04 §3.1）
func (s *ResumeService) UpdateResume(ctx context.Context, uid uint64, id uint64, resume *hrcModel.Resume, subs *ResumeSubTables) error {
	subs = subs.safe()
	if len(subs.Project) > hrcModel.ResumeProjectMax {
		return ErrResumeProjectLimit
	}
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := getOwnedResume(tx, id, uid); err != nil {
			return err
		}
		updates := map[string]interface{}{
			"title":          resume.Title,
			"fullname":       resume.FullName,
			"sex":            resume.Sex,
			"sex_cn":         resume.SexCN,
			"birthdate":      resume.Birthdate,
			"residence":      resume.Residence,
			"education":      resume.Education,
			"education_cn":   resume.EducationCN,
			"major":          resume.Major,
			"major_cn":       resume.MajorCN,
			"experience":     resume.Experience,
			"experience_cn":  resume.ExperienceCN,
			"district":       resume.District,
			"district_cn":    resume.DistrictCN,
			"wage_min":       resume.WageMin,
			"wage_max":       resume.WageMax,
			"intention_jobs": resume.IntentionJobs,
			"specialty":      resume.Specialty,
			"telephone":      resume.Telephone,
			"email":          resume.Email,
			"display_name":   resume.DisplayName,
			"current":        resume.Current,
			"current_cn":     resume.CurrentCN,
			"mobile_audit":   resume.MobileAudit,
			"talent":         resume.Talent,
			"entrust":        resume.Entrust,
		}
		if err := tx.Model(&hrcModel.Resume{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		if err := s.replaceSubTables(tx, id, uid, subs); err != nil {
			return err
		}
		return s.fillDerivedFields(tx, id, resume, subs)
	})
}

// GetResume 简历详情（编辑回显：主表 + 6 子表；仅本人未软删）
func (s *ResumeService) GetResume(ctx context.Context, uid uint64, id uint64) (*hrcModel.Resume, *ResumeSubTables, error) {
	db := global.GVA_DB.WithContext(ctx)
	resume, err := getOwnedResume(db, id, uid)
	if err != nil {
		return nil, nil, err
	}
	subs := &ResumeSubTables{}
	if err := findSubTables(db, id, subs); err != nil {
		return nil, nil, err
	}
	return resume, subs, nil
}

// ListResumes 我的简历列表（#49；仅当前 uid 未软删；默认简历在前，按创建时间倒序）
// 仅取轻量主表字段（不返子表）
func (s *ResumeService) ListResumes(ctx context.Context, uid uint64, info request.PageInfo) ([]hrcModel.Resume, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.Resume{}).Where("uid = ? AND deleted_at IS NULL", uid)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	var list []hrcModel.Resume
	err := db.Select("id", "title", "complete_percent", "def", "display", "addtime", "refreshtime").
		Order("def desc, addtime desc").
		Limit(limit).Offset(offset).Find(&list).Error
	return list, total, err
}

// DeleteResume 删除简历（#52，04 §3.5：主表软删、子表保留、投递/下载记录保留；
// 删掉的若是默认简历，最近一份自动升为默认；软删行 def 同步清零，避免裸查库出现多行 def=1）
func (s *ResumeService) DeleteResume(ctx context.Context, uid uint64, id uint64) error {
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		r, err := getOwnedResume(tx, id, uid)
		if err != nil {
			return err
		}
		if err := tx.Model(&hrcModel.Resume{}).Where("id = ?", id).Updates(map[string]interface{}{
			"deleted_at": hrcModel.Now(),
			"def":        0,
		}).Error; err != nil {
			return err
		}
		if r.Def != 1 {
			return nil
		}
		var next hrcModel.Resume
		err = tx.Where("uid = ? AND id <> ? AND deleted_at IS NULL", uid, id).Order("addtime desc, id desc").First(&next).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		return tx.Model(&hrcModel.Resume{}).Where("id = ?", next.ID).Update("def", 1).Error
	})
}

// SetDisplay 公开/隐藏切换（#54；1=公开 2=不公开）
func (s *ResumeService) SetDisplay(ctx context.Context, uid uint64, id uint64, display int8) error {
	if display != 1 && display != 2 {
		return ErrResumeDisplayValue
	}
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := getOwnedResume(tx, id, uid); err != nil {
			return err
		}
		return tx.Model(&hrcModel.Resume{}).Where("id = ?", id).Update("display", display).Error
	})
}

// SetDefault 设为默认简历（#55；同 uid 内 def 互斥，事务内先清后设）
func (s *ResumeService) SetDefault(ctx context.Context, uid uint64, id uint64) error {
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := getOwnedResume(tx, id, uid); err != nil {
			return err
		}
		if err := tx.Model(&hrcModel.Resume{}).Where("uid = ? AND deleted_at IS NULL AND def = 1", uid).Update("def", 0).Error; err != nil {
			return err
		}
		return tx.Model(&hrcModel.Resume{}).Where("id = ?", id).Update("def", 1).Error
	})
}

// GetCompleteness 完善度详情（#57；与主表 complete_percent 同源计算，返回缺失字段清单）
func (s *ResumeService) GetCompleteness(ctx context.Context, uid uint64, id uint64) (*CompletenessResult, error) {
	db := global.GVA_DB.WithContext(ctx)
	resume, err := getOwnedResume(db, id, uid)
	if err != nil {
		return nil, err
	}
	subs := &ResumeSubTables{}
	if err := findSubTables(db, id, subs); err != nil {
		return nil, err
	}
	result := CalculateCompleteness(resume, subs)
	return &result, nil
}

// getOwnedResume 归属校验（本人 + 未软删）
func getOwnedResume(db *gorm.DB, id uint64, uid uint64) (*hrcModel.Resume, error) {
	var r hrcModel.Resume
	err := db.Where("id = ? AND uid = ? AND deleted_at IS NULL", id, uid).First(&r).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrResumeNotFound
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// findSubTables 读取一份简历的 9 张子表（按 id 升序）
func findSubTables(db *gorm.DB, id uint64, subs *ResumeSubTables) error {
	subs = subs.safe()
	if err := db.Where("pid = ?", id).Order("id asc").Find(&subs.Education).Error; err != nil {
		return err
	}
	if err := db.Where("pid = ?", id).Order("id asc").Find(&subs.Work).Error; err != nil {
		return err
	}
	if err := db.Where("pid = ?", id).Order("id asc").Find(&subs.Language).Error; err != nil {
		return err
	}
	if err := db.Where("pid = ?", id).Order("id asc").Find(&subs.Training).Error; err != nil {
		return err
	}
	if err := db.Where("pid = ?", id).Order("id asc").Find(&subs.Credent).Error; err != nil {
		return err
	}
	if err := db.Where("pid = ?", id).Order("id asc").Find(&subs.Project).Error; err != nil {
		return err
	}
	if err := db.Where("pid = ?", id).Order("id asc").Find(&subs.Skill).Error; err != nil {
		return err
	}
	if err := db.Where("pid = ?", id).Order("id asc").Find(&subs.Portfolio).Error; err != nil {
		return err
	}
	if err := db.Where("pid = ?", id).Order("id asc").Find(&subs.StudentLeader).Error; err != nil {
		return err
	}
	return nil
}

// writeSubTables 批量写入子表（盖章 pid/uid）
func (s *ResumeService) writeSubTables(tx *gorm.DB, id uint64, uid uint64, subs *ResumeSubTables) error {
	stampSubTables(id, uid, subs)
	if len(subs.Education) > 0 {
		if err := tx.Create(&subs.Education).Error; err != nil {
			return err
		}
	}
	if len(subs.Work) > 0 {
		if err := tx.Create(&subs.Work).Error; err != nil {
			return err
		}
	}
	if len(subs.Language) > 0 {
		if err := tx.Create(&subs.Language).Error; err != nil {
			return err
		}
	}
	if len(subs.Training) > 0 {
		if err := tx.Create(&subs.Training).Error; err != nil {
			return err
		}
	}
	if len(subs.Credent) > 0 {
		if err := tx.Create(&subs.Credent).Error; err != nil {
			return err
		}
	}
	if len(subs.Project) > 0 {
		if err := tx.Create(&subs.Project).Error; err != nil {
			return err
		}
	}
	if len(subs.Skill) > 0 {
		if err := tx.Create(&subs.Skill).Error; err != nil {
			return err
		}
	}
	if len(subs.Portfolio) > 0 {
		if err := tx.Create(&subs.Portfolio).Error; err != nil {
			return err
		}
	}
	if len(subs.StudentLeader) > 0 {
		if err := tx.Create(&subs.StudentLeader).Error; err != nil {
			return err
		}
	}
	return nil
}

// replaceSubTables 子表全量替换（04 §3.1：删旧插新）
func (s *ResumeService) replaceSubTables(tx *gorm.DB, id uint64, uid uint64, subs *ResumeSubTables) error {
	for _, m := range []interface{}{
		&hrcModel.ResumeEducation{}, &hrcModel.ResumeWork{}, &hrcModel.ResumeLanguage{},
		&hrcModel.ResumeTraining{}, &hrcModel.ResumeCredent{}, &hrcModel.ResumeProject{},
		&hrcModel.ResumeSkill{}, &hrcModel.ResumePortfolio{}, &hrcModel.ResumeStudentLeader{},
	} {
		if err := tx.Where("pid = ?", id).Delete(m).Error; err != nil {
			return err
		}
	}
	return s.writeSubTables(tx, id, uid, subs)
}

// stampSubTables 子表盖章 pid/uid（提交数据不信任调用方传入的 id/pid/uid）
func stampSubTables(id uint64, uid uint64, subs *ResumeSubTables) {
	for i := range subs.Education {
		subs.Education[i].ID = 0
		subs.Education[i].PID = id
		subs.Education[i].UID = uid
	}
	for i := range subs.Work {
		subs.Work[i].ID = 0
		subs.Work[i].PID = id
		subs.Work[i].UID = uid
	}
	for i := range subs.Language {
		subs.Language[i].ID = 0
		subs.Language[i].PID = id
		subs.Language[i].UID = uid
	}
	for i := range subs.Training {
		subs.Training[i].ID = 0
		subs.Training[i].PID = id
		subs.Training[i].UID = uid
	}
	for i := range subs.Credent {
		subs.Credent[i].ID = 0
		subs.Credent[i].PID = id
		subs.Credent[i].UID = uid
	}
	for i := range subs.Project {
		subs.Project[i].ID = 0
		subs.Project[i].PID = id
		subs.Project[i].UID = uid
	}
	for i := range subs.Skill {
		subs.Skill[i].ID = 0
		subs.Skill[i].PID = id
		subs.Skill[i].UID = uid
	}
	for i := range subs.Portfolio {
		subs.Portfolio[i].ID = 0
		subs.Portfolio[i].PID = id
		subs.Portfolio[i].UID = uid
	}
	for i := range subs.StudentLeader {
		subs.StudentLeader[i].ID = 0
		subs.StudentLeader[i].PID = id
		subs.StudentLeader[i].UID = uid
	}
}

// fillDerivedFields 计算完善度 + 搜索索引并写主表（04 §3.1 步骤 4；与 #57 同源）
func (s *ResumeService) fillDerivedFields(tx *gorm.DB, id uint64, resume *hrcModel.Resume, subs *ResumeSubTables) error {
	result := CalculateCompleteness(resume, subs)
	keyFull, keyPrecise := buildSearchKeys(resume, subs)
	return tx.Model(&hrcModel.Resume{}).Where("id = ?", id).Updates(map[string]interface{}{
		"complete_percent": int8(result.Percent),
		"key_full":         keyFull,
		"key_precise":      keyPrecise,
	}).Error
}
