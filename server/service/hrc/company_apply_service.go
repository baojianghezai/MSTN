package hrc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"gorm.io/gorm"
)

var (
	ErrReplyStatusInvalid        = errors.New("回复状态非法（0-4）")
	ErrResumeDownloadEntitlement = errors.New("当前套餐不含简历下载权益")
	ErrResumeDownloadLimit       = errors.New("简历下载次数已用完")
)

type ResumeDownloadFile struct {
	Filename string
	Content  []byte
}

// CompanyApplyService 企业收简历服务（05 §2.2；一期脱敏=快照字段，联系方式需下载 M4）
type CompanyApplyService struct{}

// List 收简历列表（按企业 uid；可选 jobs_id 职位筛选、status 已读状态筛选）
func (s *CompanyApplyService) List(ctx context.Context, companyUID uint64, info request.PageInfo, jobsID uint64, status int8) ([]hrcModel.PersonalJobsApply, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.PersonalJobsApply{}).Where("company_uid = ?", companyUID)
	if jobsID > 0 {
		db = db.Where("jobs_id = ?", jobsID)
	}
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

// Looked 标记已看（personal_look=2，归属校验）
func (s *CompanyApplyService) Looked(ctx context.Context, companyUID uint64, did uint64) error {
	res := global.GVA_DB.WithContext(ctx).Model(&hrcModel.PersonalJobsApply{}).
		Where("did = ? AND company_uid = ?", did, companyUID).
		Update("personal_look", 2)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrApplyNotFound
	}
	return nil
}

// Reply 回复状态（is_reply 0-4：0待反馈 1合适 2不合适 3待定 4未接通）
func (s *CompanyApplyService) Reply(ctx context.Context, companyUID uint64, did uint64, isReply int8) error {
	if isReply < 0 || isReply > 4 {
		return ErrReplyStatusInvalid
	}
	res := global.GVA_DB.WithContext(ctx).Model(&hrcModel.PersonalJobsApply{}).
		Where("did = ? AND company_uid = ?", did, companyUID).
		Updates(map[string]interface{}{"is_reply": isReply, "reply_time": hrcModel.Now()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrApplyNotFound
	}
	return nil
}

// DownloadResume unlocks and renders a submitted resume. The first download by
// a company consumes one active entitlement; repeat downloads are free.
func (s *CompanyApplyService) DownloadResume(ctx context.Context, companyUID uint64, did uint64) (*ResumeDownloadFile, error) {
	var resume hrcModel.Resume
	var subs ResumeSubTables

	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var apply hrcModel.PersonalJobsApply
		if err := tx.Where("did = ? AND company_uid = ?", did, companyUID).First(&apply).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrApplyNotFound
			}
			return err
		}
		if err := tx.Where("id = ? AND uid = ? AND deleted_at IS NULL", apply.ResumeID, apply.PersonalUID).First(&resume).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrResumeNotFound
			}
			return err
		}

		var downloaded hrcModel.ResumeDownload
		err := tx.Where("company_uid = ? AND resume_id = ?", companyUID, resume.ID).First(&downloaded).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := s.consumeResumeDownload(tx, companyUID); err != nil {
				return err
			}
			if err := tx.Create(&hrcModel.ResumeDownload{
				CompanyUID:   companyUID,
				ResumeID:     resume.ID,
				ApplyID:      apply.DID,
				DownloadedAt: time.Now(),
			}).Error; err != nil {
				return err
			}
			if err := ServiceGroupApp.MessageService.SendSystemNotice(tx, SystemNotice{
				FromUID: companyUID,
				ToUID:   apply.PersonalUID,
				Title:   "简历已被企业下载",
				Message: fmt.Sprintf("企业「%s」已下载您的简历《%s》。", apply.CompanyName, resume.Title),
				Type:    "resume_download",
				Link:    "/personal/applies",
			}); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		if err := tx.Model(&hrcModel.PersonalJobsApply{}).Where("did = ?", did).Update("personal_look", 2).Error; err != nil {
			return err
		}
		return findSubTables(tx, resume.ID, &subs)
	})
	if err != nil {
		return nil, err
	}

	return &ResumeDownloadFile{
		Filename: fmt.Sprintf("resume-%d.html", resume.ID),
		Content:  buildResumeHTML(&resume, &subs),
	}, nil
}

func (s *CompanyApplyService) consumeResumeDownload(tx *gorm.DB, companyUID uint64) error {
	var entitlement hrcModel.MembersSetmeal
	err := tx.Where("uid = ?", companyUID).First(&entitlement).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrResumeDownloadEntitlement
	}
	if err != nil {
		return err
	}
	if !entitlement.ExpireAt.After(time.Now()) || entitlement.ResumeDownloadsTotal <= 0 {
		return ErrResumeDownloadEntitlement
	}
	result := tx.Model(&hrcModel.MembersSetmeal{}).
		Where("id = ? AND expire_at > ? AND resume_downloads_used < resume_downloads_total", entitlement.ID, time.Now()).
		UpdateColumn("resume_downloads_used", gorm.Expr("resume_downloads_used + ?", 1))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrResumeDownloadLimit
	}
	return nil
}

func buildResumeHTML(resume *hrcModel.Resume, subs *ResumeSubTables) []byte {
	var body bytes.Buffer
	writeResumeSection(&body, "基本信息", []string{
		"姓名：" + resume.FullName,
		"性别：" + resume.SexCN,
		"出生年份：" + uintText(resume.Birthdate),
		"最高学历：" + resume.EducationCN,
		"专业：" + resume.MajorCN,
		"工作经验：" + resume.ExperienceCN,
		"联系电话：" + resume.Telephone,
		"邮箱：" + resume.Email,
	})
	writeResumeSection(&body, "求职意向", []string{
		"期望职位：" + resume.IntentionJobs,
		"期望地区：" + resume.DistrictCN,
		"期望薪资：" + resume.WageCN,
		"当前状态：" + resume.CurrentCN,
	})
	writeResumeEntries(&body, "教育经历", len(subs.Education), func(i int) string {
		item := subs.Education[i]
		return fmt.Sprintf("%s %s%s", formatResumePeriod(item.StartYear, item.StartMonth, item.EndYear, item.EndMonth, item.ToDate), item.School, joinResumeDetails(item.Speciality, item.EducationCN))
	})
	writeResumeEntries(&body, "工作经历", len(subs.Work), func(i int) string {
		item := subs.Work[i]
		return fmt.Sprintf("%s %s%s%s", formatResumePeriod(item.StartYear, item.StartMonth, item.EndYear, item.EndMonth, item.ToDate), item.CompanyName, joinResumeDetails(item.Jobs), joinResumeDetails(item.Achievements))
	})
	writeResumeEntries(&body, "项目经历", len(subs.Project), func(i int) string {
		item := subs.Project[i]
		return fmt.Sprintf("%s %s%s%s", formatResumePeriod(item.StartYear, item.StartMonth, item.EndYear, item.EndMonth, item.ToDate), item.ProjectName, joinResumeDetails(item.Role), joinResumeDetails(item.Description))
	})
	writeResumeEntries(&body, "语言能力", len(subs.Language), func(i int) string {
		item := subs.Language[i]
		return joinResumeDetails(item.LanguageCN, item.LevelCN)
	})
	writeResumeEntries(&body, "培训经历", len(subs.Training), func(i int) string {
		item := subs.Training[i]
		return fmt.Sprintf("%s %s%s%s", formatResumePeriod(item.StartYear, item.StartMonth, item.EndYear, item.EndMonth, item.ToDate), item.Agency, joinResumeDetails(item.Course), joinResumeDetails(item.Description))
	})
	writeResumeEntries(&body, "证书", len(subs.Credent), func(i int) string {
		item := subs.Credent[i]
		return joinResumeDetails(fmt.Sprintf("%d-%02d", item.Year, item.Month), item.Name)
	})
	writeResumeSection(&body, "自我评价", []string{resume.Specialty})

	title := resume.Title
	if title == "" {
		title = resume.FullName + "的简历"
	}
	return []byte("<!doctype html><html lang=\"zh-CN\"><head><meta charset=\"utf-8\"><title>" + html.EscapeString(title) + "</title><style>body{max-width:800px;margin:36px auto;font:14px/1.7 Arial,sans-serif;color:#1f2937;padding:0 24px}h1{font-size:28px;margin:0 0 24px}h2{font-size:18px;border-bottom:1px solid #d1d5db;padding-bottom:6px;margin:28px 0 10px}.row{margin:4px 0}.entry{margin:8px 0;white-space:pre-wrap}</style></head><body><h1>" + html.EscapeString(title) + "</h1>" + body.String() + "</body></html>")
}

func writeResumeSection(body *bytes.Buffer, heading string, rows []string) {
	body.WriteString("<section><h2>" + html.EscapeString(heading) + "</h2>")
	for _, row := range rows {
		if strings.TrimSpace(row) != "" {
			body.WriteString("<div class=\"row\">" + html.EscapeString(row) + "</div>")
		}
	}
	body.WriteString("</section>")
}

func writeResumeEntries(body *bytes.Buffer, heading string, count int, entry func(int) string) {
	if count == 0 {
		return
	}
	body.WriteString("<section><h2>" + html.EscapeString(heading) + "</h2>")
	for i := 0; i < count; i++ {
		body.WriteString("<div class=\"entry\">" + html.EscapeString(entry(i)) + "</div>")
	}
	body.WriteString("</section>")
}

func formatResumePeriod(startYear uint16, startMonth uint8, endYear uint16, endMonth uint8, toDate int8) string {
	start := fmt.Sprintf("%d-%02d", startYear, startMonth)
	if toDate == 1 {
		return start + " 至今"
	}
	return fmt.Sprintf("%s 至 %d-%02d", start, endYear, endMonth)
}

func joinResumeDetails(values ...string) string {
	items := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			items = append(items, value)
		}
	}
	if len(items) == 0 {
		return ""
	}
	return " | " + strings.Join(items, " | ")
}

func uintText(value uint16) string {
	if value == 0 {
		return ""
	}
	return fmt.Sprintf("%d", value)
}
