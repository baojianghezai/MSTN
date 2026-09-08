package hrc

import (
	"context"
	"errors"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"gorm.io/gorm"
)

var (
	ErrJobNotFound            = errors.New("职位不存在")
	ErrCompanyNotAudited      = errors.New("请先完成企业认证")
	ErrCompanyAuditing        = errors.New("认证审核中，请稍后再试")
	ErrJobNameInvalid         = errors.New("职位名称需 2-50 字符")
	ErrJobWageInvalid         = errors.New("薪资范围不合法")
	ErrJobCategoryRequired    = errors.New("请选择职位分类")
	ErrJobContentsTooLong     = errors.New("职位描述最多 4000 字符")
	ErrJobAmountInvalid       = errors.New("招聘人数需 0-99")
	ErrJobExpired             = errors.New("职位已过期，无法恢复")
	ErrJobAuditInvalid        = errors.New("审核状态非法（仅支持 1=通过 3=不通过）")
	ErrJobAuditReasonRequired = errors.New("不通过时必须填写原因")
	ErrJobAuditState          = errors.New("该职位不在待审核状态")
)

// JobsService 职位服务（03 模块：发布/编辑/删除/暂停/恢复/刷新 + 审核双表流）
// 一期未做：套餐 jobs_meanwhile 约束、敏感词过滤、置顶/紧急购买（M4/M5，见 12 §四）
type JobsService struct{}

// CompanyJobItem 企业端职位列表项（jobs / jobs_tmp 合并）
type CompanyJobItem struct {
	ID     uint64 `json:"id"`
	JobsID uint64 `json:"jobsId"` // tmp 场景：原 jobs id
	hrcModel.JobsBase
	Pending bool `json:"pending"` // 待审（tmp audit=2）
}

// JobDetail 职位详情（编辑回显，含联系方式/标签/不通过原因）
type JobDetail struct {
	hrcModel.JobsBase
	ID      uint64                `json:"id"`
	JobsID  uint64                `json:"jobsId"`
	Contact *hrcModel.JobsContact `json:"contact"`
	Tags    []uint32              `json:"tags"`
	Reason  string                `json:"reason"`
	Pending bool                  `json:"pending"`
}

// AdminJobDetail 后台职位审核详情（含职位全量字段 + 联系方式/标签 + 关联企业资质）
type AdminJobDetail struct {
	hrcModel.JobsBase
	ID      uint64                   `json:"id"`
	JobsID  uint64                   `json:"jobsId"` // tmp 场景：原 ms_jobs.id；纯正式职位为自身 id
	Contact *hrcModel.JobsContact    `json:"contact"`
	Tags    []uint32                 `json:"tags"`
	Reason  string                   `json:"reason"`
	Source  string                   `json:"source"`  // jobs | tmp
	Company *hrcModel.CompanyProfile `json:"company"` // 便于审核时对照企业资质
}

// CreateJob 发布职位（企业资质校验 + display 分支；display=1 写 tmp 待审，display=2 直接入 jobs）
func (s *JobsService) CreateJob(ctx context.Context, uid uint64, job *hrcModel.Jobs, contact *hrcModel.JobsContact, tags []uint32) (uint64, error) {
	if err := s.validateJob(job); err != nil {
		return 0, err
	}
	profile, err := s.checkCompanyAudit(ctx, uid)
	if err != nil {
		return 0, err
	}
	if err := ServiceGroupApp.SetmealService.ApplyJobEntitlement(ctx, uid, job); err != nil {
		return 0, err
	}
	mode, err := s.jobsDisplayMode(ctx)
	if err != nil {
		return 0, err
	}

	now := hrcModel.Now()
	job.UID = uid
	job.CompanyID = profile.ID
	job.CompanyName = companyNameStrPtr(profile.CompanyName)
	job.Audit = 1
	job.Display = 1
	job.Click = 1
	job.UserStatus = 1
	job.AddMode = 1
	job.AddTime = now
	job.Refreshtime = now
	job.DeletedAt = 0
	if job.Deadline == 0 {
		job.Deadline = now + 30*24*3600 // 默认 30 天有效期
	}

	var id uint64
	err = global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if mode == 1 {
			tmp := &hrcModel.JobsTmp{JobsID: 0, JobsBase: job.JobsBase}
			tmp.Audit = 2 // 待审
			if err := tx.Create(tmp).Error; err != nil {
				return err
			}
			id = tmp.ID
		} else {
			if err := tx.Create(job).Error; err != nil {
				return err
			}
			id = job.ID
		}
		return s.saveContactTags(tx, id, contact, tags, uid)
	})
	return id, err
}

// UpdateJob 编辑职位
// M9：display=1 写 tmp + 原行 audit=2；display=2 直接改 jobs。
// D1（08-21）：pending=true（列表项 pending 透传，tmp 行 id）→ 直接原地重提 jobs_tmp；
// pending 缺省且 jobs 表未命中 → 回退 jobs_tmp 同 uid+id（tmp-only 职位：新发布未过审 / 被拒重提），
// 两表 id 各自自增会重叠，jobs 行存在时以 jobs 行为准（与 #81 GetJob 的 pending 语义一致）。
func (s *JobsService) UpdateJob(ctx context.Context, uid uint64, id uint64, pending bool, job *hrcModel.Jobs, contact *hrcModel.JobsContact, tags []uint32) error {
	if err := s.validateJob(job); err != nil {
		return err
	}
	if _, err := s.checkCompanyAudit(ctx, uid); err != nil {
		return err
	}
	mode, err := s.jobsDisplayMode(ctx)
	if err != nil {
		return err
	}

	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if pending {
			return s.updateTmpInPlace(tx, uid, id, job, contact, tags)
		}
		var old hrcModel.Jobs
		if err := tx.Where("id = ? AND uid = ? AND deleted_at = 0", id, uid).First(&old).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return s.updateTmpInPlace(tx, uid, id, job, contact, tags)
			}
			return err
		}
		if mode == 1 {
			// 待审：原行 audit=2（前台不可见），写 tmp
			if err := tx.Model(&hrcModel.Jobs{}).Where("id = ?", id).Update("audit", 2).Error; err != nil {
				return err
			}
			tmp := &hrcModel.JobsTmp{JobsID: id, JobsBase: job.JobsBase}
			tmp.UID = uid
			tmp.CompanyID = old.CompanyID
			tmp.CompanyName = old.CompanyName
			tmp.Audit = 2
			tmp.Display = old.Display
			tmp.Click = old.Click
			tmp.AddTime = old.AddTime
			tmp.DeletedAt = 0
			if err := tx.Create(tmp).Error; err != nil {
				return err
			}
			return s.saveContactTags(tx, tmp.ID, contact, tags, uid)
		}
		// 直接更新 jobs
		job.ID = id
		job.UID = uid
		job.CompanyID = old.CompanyID
		job.CompanyName = old.CompanyName
		job.Audit = 1
		job.Display = old.Display
		job.Click = old.Click
		job.AddTime = old.AddTime
		job.DeletedAt = 0
		job.Refreshtime = hrcModel.Now()
		if err := tx.Save(job).Error; err != nil {
			return err
		}
		if err := tx.Where("pid = ?", id).Delete(&hrcModel.JobsContact{}).Error; err != nil {
			return err
		}
		if err := tx.Where("pid = ?", id).Delete(&hrcModel.JobsTag{}).Error; err != nil {
			return err
		}
		return s.saveContactTags(tx, id, contact, tags, uid)
	})
}

// DeleteJob 逻辑删除（deleted_at 标记，保留投递记录）
// D1（08-21）：pending=true → 软删 jobs_tmp 行（tmp-only 职位）；jobs 表未命中时回退 jobs_tmp 同 uid+id
func (s *JobsService) DeleteJob(ctx context.Context, uid uint64, id uint64, pending bool) error {
	db := global.GVA_DB.WithContext(ctx)
	if !pending {
		res := db.Model(&hrcModel.Jobs{}).
			Where("id = ? AND uid = ? AND deleted_at = 0", id, uid).
			Update("deleted_at", hrcModel.Now())
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected > 0 {
			return nil
		}
	}
	// tmp 行（pending 项或 jobs 未命中回退）；contact/tag 随 pid 保留即可（职位已不可见）
	res := db.Model(&hrcModel.JobsTmp{}).
		Where("id = ? AND uid = ? AND deleted_at = 0", id, uid).
		Update("deleted_at", hrcModel.Now())
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrJobNotFound
	}
	return nil
}

// PauseJob 暂停（display=2）
func (s *JobsService) PauseJob(ctx context.Context, uid uint64, id uint64) error {
	return s.setJobDisplay(ctx, uid, id, 2)
}

// ResumeJob 恢复（display=1，需未过期）
func (s *JobsService) ResumeJob(ctx context.Context, uid uint64, id uint64) error {
	var j hrcModel.Jobs
	if err := global.GVA_DB.WithContext(ctx).Where("id = ? AND uid = ? AND deleted_at = 0", id, uid).First(&j).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrJobNotFound
		}
		return err
	}
	if j.Deadline > 0 && j.Deadline < hrcModel.Now() {
		return ErrJobExpired
	}
	return s.setJobDisplay(ctx, uid, id, 1)
}

// RefreshJob 刷新（重置 refreshtime）
func (s *JobsService) RefreshJob(ctx context.Context, uid uint64, id uint64) error {
	res := global.GVA_DB.WithContext(ctx).Model(&hrcModel.Jobs{}).
		Where("id = ? AND uid = ? AND deleted_at = 0", id, uid).
		Update("refreshtime", hrcModel.Now())
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrJobNotFound
	}
	return nil
}

// ListJobs 企业端职位列表（jobs + jobs_tmp 合并，按 id desc）
// 编辑中（jobs audit=2 且存在对应 tmp）的 jobs 行由 tmp 行代表，避免重复展示
func (s *JobsService) ListJobs(ctx context.Context, uid uint64, info request.PageInfo) ([]CompanyJobItem, int64, error) {
	var jobs []hrcModel.Jobs
	if err := global.GVA_DB.WithContext(ctx).Where("uid = ? AND deleted_at = 0", uid).Find(&jobs).Error; err != nil {
		return nil, 0, err
	}
	// D1 补充（08-21）：tmp 行也过滤软删（#83 软删后列表不应再显示）
	var tmps []hrcModel.JobsTmp
	if err := global.GVA_DB.WithContext(ctx).Where("uid = ? AND deleted_at = 0", uid).Find(&tmps).Error; err != nil {
		return nil, 0, err
	}

	// 编辑中的 jobs 行（有对应 tmp）由 tmp 行代表
	editingJobsIDs := map[uint64]bool{}
	for _, t := range tmps {
		if t.JobsID > 0 {
			editingJobsIDs[t.JobsID] = true
		}
	}

	items := make([]CompanyJobItem, 0, len(jobs)+len(tmps))
	for _, j := range jobs {
		if j.Audit == 2 && editingJobsIDs[j.ID] {
			continue // 编辑中，由 tmp 行代表
		}
		items = append(items, CompanyJobItem{ID: j.ID, JobsBase: j.JobsBase, Pending: false})
	}
	for _, t := range tmps {
		items = append(items, CompanyJobItem{ID: t.ID, JobsID: t.JobsID, JobsBase: t.JobsBase, Pending: true})
	}
	// 按 id desc 排序（tmp 与 jobs 共用排序）
	sortCompanyJobItems(items)
	total := int64(len(items))
	limit, offset := info.LimitOffset()
	if int64(offset) >= total {
		return []CompanyJobItem{}, total, nil
	}
	end := offset + limit
	if int64(end) > total {
		end = int(total)
	}
	return items[offset:end], total, nil
}

// GetJob 职位详情（编辑回显）
// pending=false：按 jobs 表 id 查；pending=true：按 jobs_tmp 表 id 查。
// 两表 id 各自自增会重叠，必须由 pending 区分表（列表项 pending 字段透传）
func (s *JobsService) GetJob(ctx context.Context, uid uint64, id uint64, pending bool) (*JobDetail, error) {
	db := global.GVA_DB.WithContext(ctx)
	detail := &JobDetail{}
	if pending {
		var t hrcModel.JobsTmp
		if err := db.Where("id = ? AND uid = ?", id, uid).First(&t).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrJobNotFound
			}
			return nil, err
		}
		detail.ID = t.ID
		detail.JobsID = t.JobsID
		detail.JobsBase = t.JobsBase
		detail.Pending = true
	} else {
		var j hrcModel.Jobs
		if err := db.Where("id = ? AND uid = ? AND deleted_at = 0", id, uid).First(&j).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrJobNotFound
			}
			return nil, err
		}
		detail.ID = j.ID
		detail.JobsBase = j.JobsBase
		detail.Pending = false
	}

	// 联系方式/标签：pid 关联当前查询行的 id（tmp.id 或 jobs.id）
	var contact hrcModel.JobsContact
	if err := db.Where("pid = ?", id).First(&contact).Error; err == nil {
		detail.Contact = &contact
	}
	var tags []hrcModel.JobsTag
	if err := db.Where("pid = ?", id).Find(&tags).Error; err != nil {
		return nil, err
	}
	detail.Tags = make([]uint32, 0, len(tags))
	for _, t := range tags {
		detail.Tags = append(detail.Tags, t.Tag)
	}
	// 不通过原因（audit=3 行，type_id 关联当前行 id）
	if detail.Audit == 3 {
		var reason hrcModel.AuditReason
		if err := db.Where("type = ? AND type_id = ?", hrcModel.AuditTypeJobs, id).Order("id desc").First(&reason).Error; err == nil {
			detail.Reason = reason.Reason
		}
	}
	return detail, nil
}

// AuditJob 审核（通过=tmp→jobs 事务同步；不通过=tmp audit=3 + reason，编辑场景恢复原 jobs）
func (s *JobsService) AuditJob(ctx context.Context, id uint64, audit int8, reason string) error {
	if audit != 1 && audit != 3 {
		return ErrJobAuditInvalid
	}
	if audit == 3 && strings.TrimSpace(reason) == "" {
		return ErrJobAuditReasonRequired
	}
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var tmp hrcModel.JobsTmp
		if err := tx.Where("id = ?", id).First(&tmp).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrJobNotFound
			}
			return err
		}
		if tmp.Audit != 2 {
			return ErrJobAuditState
		}
		if audit == 3 {
			if err := tx.Model(&hrcModel.JobsTmp{}).Where("id = ?", id).Update("audit", 3).Error; err != nil {
				return err
			}
			if err := tx.Create(&hrcModel.AuditReason{Type: hrcModel.AuditTypeJobs, TypeID: id, Reason: reason, AddTime: hrcModel.Now()}).Error; err != nil {
				return err
			}
			// 编辑场景：恢复原 jobs 为通过（旧内容不变）
			if tmp.JobsID > 0 {
				return tx.Model(&hrcModel.Jobs{}).Where("id = ?", tmp.JobsID).Update("audit", 1).Error
			}
			return nil
		}
		// 通过
		if tmp.JobsID == 0 {
			job := &hrcModel.Jobs{JobsBase: tmp.JobsBase}
			job.Audit = 1
			job.Display = 1
			job.DeletedAt = 0
			if err := tx.Create(job).Error; err != nil {
				return err
			}
			if err := s.rekeyContactTags(tx, id, job.ID); err != nil {
				return err
			}
		} else {
			var old hrcModel.Jobs
			if err := tx.Where("id = ?", tmp.JobsID).First(&old).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrJobNotFound
				}
				return err
			}
			job := &hrcModel.Jobs{ID: old.ID, JobsBase: tmp.JobsBase}
			job.Audit = 1
			job.Display = old.Display
			job.Click = old.Click
			job.AddTime = old.AddTime
			job.DeletedAt = 0
			if err := tx.Save(job).Error; err != nil {
				return err
			}
			if err := s.rekeyContactTags(tx, id, tmp.JobsID); err != nil {
				return err
			}
		}
		return tx.Where("id = ?", id).Delete(&hrcModel.JobsTmp{}).Error
	})
}

// AdminListJobs 后台职位列表（关键字 + audit 筛选）
func (s *JobsService) AdminListJobs(ctx context.Context, info request.PageInfo, audit int8, keyword string) ([]hrcModel.Jobs, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.Jobs{}).Where("deleted_at = 0")
	if audit >= 0 {
		db = db.Where("audit = ?", audit)
	}
	if keyword != "" {
		db = db.Where("jobs_name LIKE ? OR companyname LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	var list []hrcModel.Jobs
	if err := db.Order("id desc").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// AdminListJobsTmp 后台待审职位列表（jobs_tmp）
func (s *JobsService) AdminListJobsTmp(ctx context.Context, info request.PageInfo) ([]hrcModel.JobsTmp, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.JobsTmp{}).Where("audit = 2")
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	var list []hrcModel.JobsTmp
	if err := db.Order("id desc").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// AdminGetJobDetail 后台职位审核详情（jobs 与 jobs_tmp 双表查找，含联系方式/标签/不通过原因/关联企业资质）
func (s *JobsService) AdminGetJobDetail(ctx context.Context, id uint64) (*AdminJobDetail, error) {
	db := global.GVA_DB.WithContext(ctx)
	detail := &AdminJobDetail{ID: id, JobsID: id, Tags: []uint32{}}

	// 优先查 tmp（待审/不通过都在此表），未命中则查正式 jobs；两表 id 自增会重叠
	var tmp hrcModel.JobsTmp
	tmpErr := db.Where("id = ?", id).First(&tmp).Error
	if tmpErr == nil {
		detail.JobsBase = tmp.JobsBase
		detail.JobsID = tmp.JobsID
		if detail.JobsID == 0 {
			detail.JobsID = id
		}
		detail.Source = "tmp"
	} else if errors.Is(tmpErr, gorm.ErrRecordNotFound) {
		var job hrcModel.Jobs
		if err := db.Where("id = ? AND deleted_at = 0", id).First(&job).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrJobNotFound
			}
			return nil, err
		}
		detail.JobsBase = job.JobsBase
		detail.JobsID = job.ID
		detail.Source = "jobs"
	} else {
		return nil, tmpErr
	}

	// 联系方式：pid 关联当前查询行的 id（tmp.id 或 jobs.id）
	var contact hrcModel.JobsContact
	if err := db.Where("pid = ?", id).First(&contact).Error; err == nil {
		detail.Contact = &contact
	}

	// 标签
	var tags []hrcModel.JobsTag
	if err := db.Where("pid = ?", id).Find(&tags).Error; err != nil {
		return nil, err
	}
	for _, t := range tags {
		detail.Tags = append(detail.Tags, t.Tag)
	}

	// 不通过原因（audit=3 行）
	if detail.Audit == 3 {
		var reason hrcModel.AuditReason
		if err := db.Where("type = ? AND type_id = ?", hrcModel.AuditTypeJobs, id).Order("id desc").First(&reason).Error; err == nil {
			detail.Reason = reason.Reason
		}
	}

	// 关联企业资质（便于审核对照；缺省不阻断）
	if detail.CompanyID > 0 {
		var company hrcModel.CompanyProfile
		if err := db.Where("id = ?", detail.CompanyID).First(&company).Error; err == nil {
			detail.Company = &company
		}
	}

	return detail, nil
}

// ---- 内部辅助 ----

func (s *JobsService) validateJob(job *hrcModel.Jobs) error {
	if n := len([]rune(job.JobsName)); n < 2 || n > 50 {
		return ErrJobNameInvalid
	}
	if len([]rune(job.Contents)) > 4000 {
		return ErrJobContentsTooLong
	}
	if job.TopClass == 0 || job.Category == 0 || job.SubClass == 0 {
		return ErrJobCategoryRequired
	}
	if job.Amount > 99 {
		return ErrJobAmountInvalid
	}
	if job.Negotiable == 0 {
		if job.MinWage < 0 || job.MaxWage < 0 || job.MaxWage > 999999 || job.MinWage > 999999 {
			return ErrJobWageInvalid
		}
		if job.MaxWage <= job.MinWage {
			return ErrJobWageInvalid
		}
		if job.MinWage > 0 && job.MaxWage/job.MinWage > 2 {
			return ErrJobWageInvalid
		}
	}
	return nil
}

// checkCompanyAudit 企业资质校验（S10：audit=1 才可发布）
func (s *JobsService) checkCompanyAudit(ctx context.Context, uid uint64) (*hrcModel.CompanyProfile, error) {
	var p hrcModel.CompanyProfile
	err := global.GVA_DB.WithContext(ctx).Where("uid = ?", uid).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCompanyNotAudited
	}
	if err != nil {
		return nil, err
	}
	if p.Audit == 2 {
		return nil, ErrCompanyAuditing
	}
	if p.Audit != 1 {
		return nil, ErrCompanyNotAudited
	}
	return &p, nil
}

// jobsDisplayMode 读配置 mscms_jobs_display（"2"=直接显示；缺省/其他=审核后显示）
// 修改批 X2（08-20 派活方拍板）：职位发布后必须后台审核，默认值从「直接显示」翻转为「审核」
func (s *JobsService) jobsDisplayMode(ctx context.Context) (int8, error) {
	v, err := getConfigValue(ctx, "mscms_jobs_display")
	if err != nil {
		return 0, err
	}
	if v == "2" {
		return 2, nil
	}
	return 1, nil
}

func (s *JobsService) setJobDisplay(ctx context.Context, uid uint64, id uint64, display int8) error {
	res := global.GVA_DB.WithContext(ctx).Model(&hrcModel.Jobs{}).
		Where("id = ? AND uid = ? AND deleted_at = 0", id, uid).
		Update("display", display)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrJobNotFound
	}
	return nil
}

// updateTmpInPlace tmp-only 职位原地重提（D1，08-21）：
// audit∈{2 待审, 3 不通过} 允许编辑，JobsBase 全量更新 + contact/tags 重挂 tmp.id + audit 置回 2；
// 其他状态（如 1 通过——正常流程不存在，防御）报「该职位状态不可编辑」。
func (s *JobsService) updateTmpInPlace(tx *gorm.DB, uid, id uint64, job *hrcModel.Jobs, contact *hrcModel.JobsContact, tags []uint32) error {
	var old hrcModel.JobsTmp
	if err := tx.Where("id = ? AND uid = ? AND deleted_at = 0", id, uid).First(&old).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrJobNotFound
		}
		return err
	}
	if old.Audit != 2 && old.Audit != 3 {
		return ErrJobAuditState
	}
	job.UID = uid
	job.CompanyID = old.CompanyID
	job.CompanyName = old.CompanyName
	job.Audit = 2
	job.Display = old.Display
	job.Click = old.Click
	job.AddTime = old.AddTime
	job.DeletedAt = 0
	job.Refreshtime = hrcModel.Now()
	tmp := &hrcModel.JobsTmp{ID: old.ID, JobsID: old.JobsID, JobsBase: job.JobsBase}
	if err := tx.Save(tmp).Error; err != nil {
		return err
	}
	if err := tx.Where("pid = ?", id).Delete(&hrcModel.JobsContact{}).Error; err != nil {
		return err
	}
	if err := tx.Where("pid = ?", id).Delete(&hrcModel.JobsTag{}).Error; err != nil {
		return err
	}
	return s.saveContactTags(tx, id, contact, tags, uid)
}

// saveContactTags 保存联系方式 + 标签（pid 关联职位 id）
func (s *JobsService) saveContactTags(tx *gorm.DB, pid uint64, contact *hrcModel.JobsContact, tags []uint32, uid uint64) error {
	if contact != nil {
		contact.ID = 0
		contact.PID = pid
		if err := tx.Create(contact).Error; err != nil {
			return err
		}
	}
	if len(tags) > 0 {
		rows := make([]hrcModel.JobsTag, 0, len(tags))
		for _, t := range tags {
			rows = append(rows, hrcModel.JobsTag{UID: uid, PID: pid, Tag: t})
		}
		if err := tx.Create(&rows).Error; err != nil {
			return err
		}
	}
	return nil
}

// rekeyContactTags 审核通过后把联系方式/标签从 tmp id 重键到正式 jobs id（先删旧后重键）
func (s *JobsService) rekeyContactTags(tx *gorm.DB, fromPID, toPID uint64) error {
	if fromPID == toPID {
		return nil
	}
	if err := tx.Where("pid = ?", toPID).Delete(&hrcModel.JobsContact{}).Error; err != nil {
		return err
	}
	if err := tx.Where("pid = ?", toPID).Delete(&hrcModel.JobsTag{}).Error; err != nil {
		return err
	}
	if err := tx.Model(&hrcModel.JobsContact{}).Where("pid = ?", fromPID).Update("pid", toPID).Error; err != nil {
		return err
	}
	return tx.Model(&hrcModel.JobsTag{}).Where("pid = ?", fromPID).Update("pid", toPID).Error
}

// sortCompanyJobItems 按 id desc 排序（jobs 与 tmp 合并后）
func sortCompanyJobItems(items []CompanyJobItem) {
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j].ID > items[j-1].ID; j-- {
			items[j], items[j-1] = items[j-1], items[j]
		}
	}
}
