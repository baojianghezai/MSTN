package hrc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"gorm.io/gorm"
)

var (
	ErrInterviewNotFound  = errors.New("面试邀请不存在")
	ErrInterviewDuplicate = errors.New("该候选人已收到此职位的面试邀请")
	ErrInterviewTime      = errors.New("面试时间必须晚于当前时间")
	ErrInterviewRequired  = errors.New("请填写面试地址、联系人和联系电话")
	ErrInterviewTooLong   = errors.New("面试邀请字段长度超出限制")
)

type InterviewCreateInput struct {
	ResumeID      uint64
	JobsID        uint64
	InterviewTime int64
	Address       string
	Contact       string
	Telephone     string
	Notes         string
}

// InterviewService 管理企业发出的标准线下面试邀请。
type InterviewService struct{}

func (s *InterviewService) Create(ctx context.Context, companyUID uint64, input InterviewCreateInput) (*hrcModel.CompanyInterview, error) {
	input.Address = strings.TrimSpace(input.Address)
	input.Contact = strings.TrimSpace(input.Contact)
	input.Telephone = strings.TrimSpace(input.Telephone)
	input.Notes = strings.TrimSpace(input.Notes)
	if input.InterviewTime <= time.Now().Unix() {
		return nil, ErrInterviewTime
	}
	if input.Address == "" || input.Contact == "" || input.Telephone == "" {
		return nil, ErrInterviewRequired
	}
	if len(input.Address) > 200 || len(input.Contact) > 30 || len(input.Telephone) > 30 || len(input.Notes) > 500 {
		return nil, ErrInterviewTooLong
	}

	var created hrcModel.CompanyInterview
	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var resume hrcModel.Resume
		if err := tx.Where("id = ? AND display = 1 AND audit = 1 AND deleted_at = 0", input.ResumeID).First(&resume).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrResumeUnavailable
			}
			return err
		}

		var job hrcModel.Jobs
		if err := tx.Where("id = ? AND uid = ? AND deleted_at = 0", input.JobsID, companyUID).First(&job).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrJobNotAvailable
			}
			return err
		}

		var duplicate int64
		if err := tx.Model(&hrcModel.CompanyInterview{}).
			Where("company_uid = ? AND resume_id = ? AND jobs_id = ?", companyUID, resume.ID, job.ID).
			Count(&duplicate).Error; err != nil {
			return err
		}
		if duplicate > 0 {
			return ErrInterviewDuplicate
		}

		created = hrcModel.CompanyInterview{
			ResumeID:         resume.ID,
			ResumeName:       resume.FullName,
			ResumeUID:        resume.UID,
			JobsID:           job.ID,
			JobsName:         job.JobsName,
			CompanyID:        job.CompanyID,
			CompanyName:      job.CompanyName,
			CompanyUID:       companyUID,
			InterviewTime:    input.InterviewTime,
			Address:          input.Address,
			Contact:          input.Contact,
			Telephone:        input.Telephone,
			Notes:            input.Notes,
			InterviewAddtime: time.Now().Unix(),
			PersonalLook:     1,
		}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}
		return ServiceGroupApp.MessageService.SendSystemNotice(tx, SystemNotice{
			FromUID: companyUID,
			ToUID:   resume.UID,
			Title:   "收到面试邀请",
			Message: fmt.Sprintf("企业「%s」邀请您参加「%s」的线下面试，时间：%s。", job.CompanyName, job.JobsName, time.Unix(input.InterviewTime, 0).Format("2006-01-02 15:04")),
			Type:    "interview",
			Link:    "/personal/interviews",
		})
	})
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (s *InterviewService) CompanyList(ctx context.Context, companyUID uint64, info request.PageInfo) ([]hrcModel.CompanyInterview, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.CompanyInterview{}).Where("company_uid = ?", companyUID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	var list []hrcModel.CompanyInterview
	if err := db.Order("did desc").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *InterviewService) Withdraw(ctx context.Context, companyUID uint64, did uint64) error {
	result := global.GVA_DB.WithContext(ctx).
		Where("did = ? AND company_uid = ?", did, companyUID).
		Delete(&hrcModel.CompanyInterview{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrInterviewNotFound
	}
	return nil
}

func (s *InterviewService) PersonalList(ctx context.Context, resumeUID uint64, info request.PageInfo) ([]hrcModel.CompanyInterview, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.CompanyInterview{}).Where("resume_uid = ?", resumeUID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	var list []hrcModel.CompanyInterview
	if err := db.Order("personal_look asc, interview_time asc, did desc").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *InterviewService) MarkRead(ctx context.Context, resumeUID uint64, did uint64) error {
	result := global.GVA_DB.WithContext(ctx).Model(&hrcModel.CompanyInterview{}).
		Where("did = ? AND resume_uid = ?", did, resumeUID).
		Update("personal_look", 2)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrInterviewNotFound
	}
	return nil
}
