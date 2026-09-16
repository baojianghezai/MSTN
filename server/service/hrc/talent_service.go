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
	ErrTalentNotFound     = errors.New("简历不存在或暂不可见")
	ErrTalentNotUnlocked  = errors.New("请先解锁该候选人简历")
	ErrTalentFollowUp     = errors.New("跟进状态非法（0-4）")
	ErrTalentFavorite     = errors.New("该候选人已收藏")
	ErrTalentFavoriteGone = errors.New("收藏记录不存在")
)

type TalentSearch struct {
	Keyword    string
	District   string
	Education  uint16
	Experience uint16
	WageMin    uint16
	WageMax    uint16
	TalentOnly bool
}

// PublicResume is deliberately contact-free so it can safely be used by
// public search, public detail, and the company favorites list.
type PublicResume struct {
	ID              uint64    `json:"id"`
	Title           string    `json:"title"`
	FullName        string    `json:"fullname"`
	Sex             int8      `json:"sex"`
	SexCN           string    `json:"sexCn"`
	Birthdate       uint16    `json:"birthdate"`
	Education       uint16    `json:"education"`
	EducationCN     string    `json:"educationCn"`
	MajorCN         string    `json:"majorCn"`
	Experience      uint16    `json:"experience"`
	ExperienceCN    string    `json:"experienceCn"`
	District        string    `json:"district"`
	DistrictCN      string    `json:"districtCn"`
	WageMin         uint16    `json:"wageMin"`
	WageMax         uint16    `json:"wageMax"`
	WageCN          string    `json:"wageCn"`
	IntentionJobs   string    `json:"intentionJobs"`
	Specialty       string    `json:"specialty"`
	PhotoImg        string    `json:"photoImg"`
	CompletePercent int8      `json:"completePercent"`
	Talent          int8      `json:"talent"`
	Refreshtime     time.Time `json:"refreshtime"`
}

type PublicResumeDetail struct {
	Resume        PublicResume                   `json:"resume"`
	Projects      []hrcModel.ResumeProject       `json:"projects"`
	Educations    []hrcModel.ResumeEducation     `json:"educations"`
	Work          []hrcModel.ResumeWork          `json:"work"`
	Language      []hrcModel.ResumeLanguage      `json:"language"`
	Training      []hrcModel.ResumeTraining      `json:"training"`
	Credent       []hrcModel.ResumeCredent       `json:"credent"`
	Skill         []hrcModel.ResumeSkill         `json:"skill"`
	Portfolio     []hrcModel.ResumePortfolio     `json:"portfolio"`
	StudentLeader []hrcModel.ResumeStudentLeader `json:"studentLeader"`
}

type TalentUnlockedDetail struct {
	Resume        hrcModel.Resume                `json:"resume"`
	Projects      []hrcModel.ResumeProject       `json:"projects"`
	Educations    []hrcModel.ResumeEducation     `json:"educations"`
	Work          []hrcModel.ResumeWork          `json:"work"`
	Language      []hrcModel.ResumeLanguage      `json:"language"`
	Training      []hrcModel.ResumeTraining      `json:"training"`
	Credent       []hrcModel.ResumeCredent       `json:"credent"`
	Skill         []hrcModel.ResumeSkill         `json:"skill"`
	Portfolio     []hrcModel.ResumePortfolio     `json:"portfolio"`
	StudentLeader []hrcModel.ResumeStudentLeader `json:"studentLeader"`
}

type CompanyTalentItem struct {
	Download hrcModel.ResumeDownload `json:"download"`
	Resume   hrcModel.Resume         `json:"resume"`
}

type FavoriteTalentItem struct {
	Favorite hrcModel.CompanyFavorite `json:"favorite"`
	Resume   PublicResume             `json:"resume"`
}

// TalentService provides public candidate discovery and enterprise-only
// unlocking, favorites, and follow-up management.
type TalentService struct{}

func (s *TalentService) Search(ctx context.Context, info request.PageInfo, search TalentSearch) ([]PublicResume, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.Resume{}).
		Where("display = 1 AND audit = 1 AND deleted_at IS NULL")
	if search.TalentOnly {
		db = db.Where("talent = 1")
	}
	if search.District != "" {
		db = db.Where("district = ?", search.District)
	}
	if search.Education > 0 {
		db = db.Where("education = ?", search.Education)
	}
	if search.Experience > 0 {
		db = db.Where("experience = ?", search.Experience)
	}
	// 期望薪资区间与筛选区间有交集即命中（0 视为该端不限）
	if search.WageMin > 0 {
		db = db.Where("wage_max >= ? OR wage_max = 0", search.WageMin)
	}
	if search.WageMax > 0 {
		db = db.Where("wage_min <= ? OR wage_min = 0", search.WageMax)
	}
	if keyword := strings.TrimSpace(search.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("key_full LIKE ? OR title LIKE ? OR intention_jobs LIKE ?", like, like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	var resumes []hrcModel.Resume
	if err := db.Order("talent desc, refreshtime desc, id desc").Limit(limit).Offset(offset).Find(&resumes).Error; err != nil {
		return nil, 0, err
	}
	list := make([]PublicResume, 0, len(resumes))
	for _, resume := range resumes {
		list = append(list, publicResumeFromModel(resume))
	}
	return list, total, nil
}

func (s *TalentService) PublicDetail(ctx context.Context, resumeID uint64) (*PublicResumeDetail, error) {
	db := global.GVA_DB.WithContext(ctx)
	var resume hrcModel.Resume
	if err := db.Where("id = ? AND display = 1 AND audit = 1 AND deleted_at IS NULL", resumeID).First(&resume).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTalentNotFound
		}
		return nil, err
	}
	if err := db.Model(&hrcModel.Resume{}).Where("id = ?", resumeID).UpdateColumn("click", gorm.Expr("click + ?", 1)).Error; err != nil {
		return nil, err
	}
	subs := ResumeSubTables{}
	if err := findSubTables(db, resumeID, &subs); err != nil {
		return nil, err
	}
	return publicResumeDetailFromModel(resume, subs), nil
}

func (s *TalentService) Unlock(ctx context.Context, companyUID uint64, resumeID uint64) (*TalentUnlockedDetail, bool, error) {
	var resume hrcModel.Resume
	var subs ResumeSubTables
	newlyUnlocked := false
	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND display = 1 AND audit = 1 AND deleted_at IS NULL", resumeID).First(&resume).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTalentNotFound
			}
			return err
		}
		var downloaded hrcModel.ResumeDownload
		err := tx.Where("company_uid = ? AND resume_id = ?", companyUID, resumeID).First(&downloaded).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := (&CompanyApplyService{}).consumeResumeDownload(tx, companyUID); err != nil {
				return err
			}
			if err := tx.Create(&hrcModel.ResumeDownload{
				CompanyUID:   companyUID,
				ResumeID:     resumeID,
				DownloadedAt: time.Now(),
			}).Error; err != nil {
				return err
			}
			if err := ServiceGroupApp.MessageService.SendSystemNotice(tx, SystemNotice{
				FromUID: companyUID,
				ToUID:   resume.UID,
				Title:   "简历已被企业解锁",
				Message: fmt.Sprintf("有企业已从人才库解锁您的简历《%s》。", resume.Title),
				Type:    "resume_download",
				Link:    "/personal/messages",
			}); err != nil {
				return err
			}
			newlyUnlocked = true
		} else if err != nil {
			return err
		}
		return findSubTables(tx, resumeID, &subs)
	})
	if err != nil {
		return nil, false, err
	}
	return unlockedResumeDetailFromModel(resume, subs), newlyUnlocked, nil
}

func (s *TalentService) GetUnlocked(ctx context.Context, companyUID uint64, resumeID uint64) (*TalentUnlockedDetail, error) {
	db := global.GVA_DB.WithContext(ctx)
	var download hrcModel.ResumeDownload
	if err := db.Where("company_uid = ? AND resume_id = ?", companyUID, resumeID).First(&download).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTalentNotUnlocked
		}
		return nil, err
	}
	var resume hrcModel.Resume
	if err := db.Where("id = ? AND deleted_at IS NULL", resumeID).First(&resume).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTalentNotFound
		}
		return nil, err
	}
	subs := ResumeSubTables{}
	if err := findSubTables(db, resumeID, &subs); err != nil {
		return nil, err
	}
	return unlockedResumeDetailFromModel(resume, subs), nil
}

func (s *TalentService) ListUnlocked(ctx context.Context, companyUID uint64, info request.PageInfo) ([]CompanyTalentItem, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.ResumeDownload{}).Where("company_uid = ?", companyUID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	var downloads []hrcModel.ResumeDownload
	if err := db.Order("downloaded_at desc, id desc").Limit(limit).Offset(offset).Find(&downloads).Error; err != nil {
		return nil, 0, err
	}
	resumeIDs := make([]uint64, 0, len(downloads))
	for _, download := range downloads {
		resumeIDs = append(resumeIDs, download.ResumeID)
	}
	resumeMap := make(map[uint64]hrcModel.Resume, len(resumeIDs))
	if len(resumeIDs) > 0 {
		var resumes []hrcModel.Resume
		if err := global.GVA_DB.WithContext(ctx).Where("id IN ? AND deleted_at IS NULL", resumeIDs).Find(&resumes).Error; err != nil {
			return nil, 0, err
		}
		for _, resume := range resumes {
			resume.WageCN = formatWageRange(resume.WageMin, resume.WageMax)
			resumeMap[resume.ID] = resume
		}
	}
	list := make([]CompanyTalentItem, 0, len(downloads))
	for _, download := range downloads {
		if resume, ok := resumeMap[download.ResumeID]; ok {
			list = append(list, CompanyTalentItem{Download: download, Resume: resume})
		}
	}
	return list, total, nil
}

func (s *TalentService) SetFollowUp(ctx context.Context, companyUID uint64, resumeID uint64, followUp int8) error {
	if followUp < 0 || followUp > 4 {
		return ErrTalentFollowUp
	}
	result := global.GVA_DB.WithContext(ctx).Model(&hrcModel.ResumeDownload{}).
		Where("company_uid = ? AND resume_id = ?", companyUID, resumeID).
		Update("follow_up", followUp)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTalentNotUnlocked
	}
	return nil
}

func (s *TalentService) ListFavorites(ctx context.Context, companyUID uint64, info request.PageInfo) ([]FavoriteTalentItem, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.CompanyFavorite{}).Where("company_uid = ?", companyUID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	var favorites []hrcModel.CompanyFavorite
	if err := db.Order("addtime desc, id desc").Limit(limit).Offset(offset).Find(&favorites).Error; err != nil {
		return nil, 0, err
	}
	resumeIDs := make([]uint64, 0, len(favorites))
	for _, favorite := range favorites {
		resumeIDs = append(resumeIDs, favorite.ResumeID)
	}
	resumeMap := make(map[uint64]hrcModel.Resume, len(resumeIDs))
	if len(resumeIDs) > 0 {
		var resumes []hrcModel.Resume
		if err := global.GVA_DB.WithContext(ctx).Where("id IN ? AND display = 1 AND audit = 1 AND deleted_at IS NULL", resumeIDs).Find(&resumes).Error; err != nil {
			return nil, 0, err
		}
		for _, resume := range resumes {
			resumeMap[resume.ID] = resume
		}
	}
	list := make([]FavoriteTalentItem, 0, len(favorites))
	for _, favorite := range favorites {
		if resume, ok := resumeMap[favorite.ResumeID]; ok {
			list = append(list, FavoriteTalentItem{Favorite: favorite, Resume: publicResumeFromModel(resume)})
		}
	}
	return list, total, nil
}

func (s *TalentService) Favorite(ctx context.Context, companyUID uint64, resumeID uint64) (*hrcModel.CompanyFavorite, error) {
	var favorite hrcModel.CompanyFavorite
	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var resume hrcModel.Resume
		if err := tx.Where("id = ? AND display = 1 AND audit = 1 AND deleted_at IS NULL", resumeID).First(&resume).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTalentNotFound
			}
			return err
		}
		err := tx.Where("company_uid = ? AND resume_id = ?", companyUID, resumeID).First(&favorite).Error
		if err == nil {
			return ErrTalentFavorite
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		favorite = hrcModel.CompanyFavorite{CompanyUID: companyUID, ResumeID: resumeID, AddTime: time.Now()}
		return tx.Create(&favorite).Error
	})
	if err != nil {
		return nil, err
	}
	return &favorite, nil
}

func (s *TalentService) Unfavorite(ctx context.Context, companyUID uint64, favoriteID uint64) error {
	result := global.GVA_DB.WithContext(ctx).Where("id = ? AND company_uid = ?", favoriteID, companyUID).Delete(&hrcModel.CompanyFavorite{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTalentFavoriteGone
	}
	return nil
}

func publicResumeFromModel(resume hrcModel.Resume) PublicResume {
	fullName := resume.FullName
	photoImg := ""
	if resume.DisplayName == 2 {
		fullName = "求职者"
	} else if resume.Photo == 1 && resume.PhotoAudit == 1 && resume.PhotoDisplay == 1 {
		photoImg = resume.PhotoImg
	}
	var refreshTime time.Time
	if resume.Refreshtime != nil {
		refreshTime = *resume.Refreshtime
	}
	return PublicResume{
		ID: resume.ID, Title: resume.Title, FullName: fullName, Sex: resume.Sex, SexCN: resume.SexCN,
		Birthdate: resume.Birthdate, Education: resume.Education, EducationCN: resume.EducationCN,
		MajorCN: resume.MajorCN, Experience: resume.Experience, ExperienceCN: resume.ExperienceCN,
		District: resume.District, DistrictCN: resume.DistrictCN,
		WageMin: resume.WageMin, WageMax: resume.WageMax, WageCN: formatWageRange(resume.WageMin, resume.WageMax),
		IntentionJobs: resume.IntentionJobs, Specialty: resume.Specialty, PhotoImg: photoImg,
		CompletePercent: resume.CompletePercent, Talent: resume.Talent, Refreshtime: refreshTime,
	}
}

func publicResumeDetailFromModel(resume hrcModel.Resume, subs ResumeSubTables) *PublicResumeDetail {
	return &PublicResumeDetail{
		Resume: publicResumeFromModel(resume), Projects: subs.Project, Educations: subs.Education,
		Work: subs.Work, Language: subs.Language, Training: subs.Training, Credent: subs.Credent,
		Skill: subs.Skill, Portfolio: subs.Portfolio, StudentLeader: subs.StudentLeader,
	}
}

func unlockedResumeDetailFromModel(resume hrcModel.Resume, subs ResumeSubTables) *TalentUnlockedDetail {
	resume.WageCN = formatWageRange(resume.WageMin, resume.WageMax)
	return &TalentUnlockedDetail{
		Resume: resume, Projects: subs.Project, Educations: subs.Education,
		Work: subs.Work, Language: subs.Language, Training: subs.Training, Credent: subs.Credent,
		Skill: subs.Skill, Portfolio: subs.Portfolio, StudentLeader: subs.StudentLeader,
	}
}
