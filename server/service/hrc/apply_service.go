package hrc

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"gorm.io/gorm"
)

var (
	ErrResumeRequired    = errors.New("请先填写简历")
	ErrResumeUnavailable = errors.New("简历不存在或已删除")
	ErrResumeIncomplete  = errors.New("简历完善度不足，请先完善至 40% 再投递")
	ErrApplyDailyLimit   = errors.New("今日投递次数已用完")
	ErrAlreadyApplied    = errors.New("您已向该公司投递过简历")
	ErrJobNotAvailable   = errors.New("职位不存在或已下架")
	ErrApplyNotFound     = errors.New("投递记录不存在")
)

// MinApplyResumeCompletePercent 投递所需的最低简历完善度，避免空白或信息过少的简历进入企业端。
const MinApplyResumeCompletePercent int8 = 40

// ApplyService 投递服务（05 §2.1；一期未做：企业屏蔽名单、短信通知）
type ApplyService struct{}

// Apply 投递职位（每日上限 + 默认简历 + 去重「同一企业对同一份简历仅投一次」）
func (s *ApplyService) Apply(ctx context.Context, uid uint64, jobsIDs []uint64, resumeID uint64, notes string) error {
	if len(jobsIDs) == 0 {
		return errors.New("请选择职位")
	}
	resume, err := s.selectResume(ctx, uid, resumeID)
	if err != nil {
		return err
	}
	if resume.CompletePercent < MinApplyResumeCompletePercent {
		return ErrResumeIncomplete
	}
	// 每日上限（配置 mscms_apply_jobs_max，默认 10）
	max := 10
	if v, err := getConfigValue(ctx, "mscms_apply_jobs_max"); err == nil && v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			max = n
		}
	}
	todayStart := time.Now()
	todayStart = time.Date(todayStart.Year(), todayStart.Month(), todayStart.Day(), 0, 0, 0, 0, todayStart.Location())
	var todayCount int64
	if err := global.GVA_DB.WithContext(ctx).Model(&hrcModel.PersonalJobsApply{}).
		Where("personal_uid = ? AND apply_addtime >= ?", uid, todayStart.Unix()).Count(&todayCount).Error; err != nil {
		return err
	}
	if todayCount+int64(len(jobsIDs)) > int64(max) {
		return ErrApplyDailyLimit
	}

	now := time.Now().Unix()
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, jobsID := range jobsIDs {
			var job hrcModel.Jobs
			err := tx.Where("id = ? AND display = 1 AND audit = 1 AND deleted_at = 0 AND (deadline = 0 OR deadline > ?)", jobsID, now).First(&job).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrJobNotAvailable
			}
			if err != nil {
				return err
			}
			// 去重：同一企业对同一份简历（uk_uid_resume_company 兜底）
			var dup int64
			if err := tx.Model(&hrcModel.PersonalJobsApply{}).
				Where("personal_uid = ? AND resume_id = ? AND company_uid = ?", uid, resume.ID, job.UID).Count(&dup).Error; err != nil {
				return err
			}
			if dup > 0 {
				return ErrAlreadyApplied
			}
			apply := &hrcModel.PersonalJobsApply{
				ResumeID:     resume.ID,
				ResumeName:   resume.FullName,
				PersonalUID:  uid,
				JobsID:       job.ID,
				JobsName:     job.JobsName,
				CompanyID:    job.CompanyID,
				CompanyName:  job.CompanyName,
				CompanyUID:   job.UID,
				ApplyAddtime: now,
				PersonalLook: 1,
				Notes:        notes,
				IsReply:      0,
			}
			if err := tx.Create(apply).Error; err != nil {
				return err
			}
			if err := ServiceGroupApp.MessageService.SendSystemNotice(tx, SystemNotice{
				FromUID: uid,
				ToUID:   job.UID,
				Title:   "收到新的职位投递",
				Message: fmt.Sprintf("%s 向职位「%s」投递了简历，请及时处理。", resume.FullName, job.JobsName),
				Type:    "application",
				Link:    "/company/applies",
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

// List 我的投递列表（状态筛选：personal_look 1未读 2已读；is_reply 0-4）
func (s *ApplyService) List(ctx context.Context, uid uint64, info request.PageInfo, status int8) ([]hrcModel.PersonalJobsApply, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.PersonalJobsApply{}).Where("personal_uid = ?", uid)
	if status > 0 {
		db = db.Where("personal_look = ?", status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	var list []hrcModel.PersonalJobsApply
	if err := db.Order("did desc").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Delete 删除投递记录（仅本人）
func (s *ApplyService) Delete(ctx context.Context, uid uint64, did uint64) error {
	res := global.GVA_DB.WithContext(ctx).Where("did = ? AND personal_uid = ?", did, uid).Delete(&hrcModel.PersonalJobsApply{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrApplyNotFound
	}
	return nil
}

// selectResume 简历选择：resumeID 指定则校验归属+通过+未软删；否则取默认简历（def desc, id desc 第一条）
func (s *ApplyService) selectResume(ctx context.Context, uid uint64, resumeID uint64) (*hrcModel.Resume, error) {
	db := global.GVA_DB.WithContext(ctx)
	if resumeID > 0 {
		var resume hrcModel.Resume
		err := db.Where("id = ? AND uid = ? AND audit = 1 AND deleted_at = 0", resumeID, uid).First(&resume).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrResumeUnavailable // 指定简历不存在/已删/未过审
		}
		if err != nil {
			return nil, err
		}
		return &resume, nil
	}
	var resume hrcModel.Resume
	err := db.Where("uid = ? AND audit = 1 AND deleted_at = 0", uid).Order("def desc, id desc").First(&resume).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrResumeRequired
	}
	return &resume, err
}
