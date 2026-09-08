package hrc

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"gorm.io/gorm"
)

var (
	ErrVideoInterviewClosed     = errors.New("视频面试未开启")
	ErrVideoInterviewNotFound   = errors.New("视频面试不存在")
	ErrVideoInterviewDuplicated = errors.New("您已对该简历进行过面试邀请，不能重复邀请")
	ErrResumeNotAvailable       = errors.New("简历不存在")
)

// VideoInterviewInput 企业发起视频面试邀请参数
type VideoInterviewInput struct {
	ResumeID      uint64
	JobsID        uint64
	JobsName      string
	InterviewTime int64
	Contact       string
	Telephone     string
}

// VideoInterviewItem 视频面试列表/详情项（含关联展示字段）
type VideoInterviewItem struct {
	hrcModel.VideoInterview
	ResumeID    uint64 `json:"resumeId"`
	FullName    string `json:"fullname"`
	CompanyName string `json:"companyname"`
}

// VideoInterviewService 视频面试服务（11 §四 P0#2，v6 qs_video_interview 平移）
type VideoInterviewService struct{}

// CreateInterview 企业发起视频面试邀请（生成企业/个人房间码；套餐/职位审核/下载资格随 M3/M4 补）
func (s *VideoInterviewService) CreateInterview(ctx context.Context, companyUID uint64, in VideoInterviewInput) (*hrcModel.VideoInterview, error) {
	open, err := getConfigValue(ctx, "video_interview_open")
	if err != nil {
		return nil, err
	}
	if open != "1" {
		return nil, ErrVideoInterviewClosed
	}
	// 简历存在性 + 归属个人 uid
	var resume hrcModel.Resume
	if err := global.GVA_DB.WithContext(ctx).Where("id = ?", in.ResumeID).First(&resume).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrResumeNotAvailable
		}
		return nil, err
	}
	// 去重：同一企业 + 同一简历 + 同一职位 + 未过期
	var dup int64
	if err := global.GVA_DB.WithContext(ctx).Model(&hrcModel.VideoInterview{}).
		Where("company_uid = ? AND personal_uid = ? AND jobs_id = ? AND deadline > ?",
			companyUID, resume.UID, in.JobsID, time.Now().Unix()).Count(&dup).Error; err != nil {
		return nil, err
	}
	if dup > 0 {
		return nil, ErrVideoInterviewDuplicated
	}

	now := time.Now().Unix()
	record := &hrcModel.VideoInterview{
		CompanyUID:    companyUID,
		PersonalUID:   resume.UID,
		JobsID:        in.JobsID,
		JobsName:      in.JobsName,
		InterviewTime: in.InterviewTime,
		Deadline:      in.InterviewTime + hrcModel.VideoDeadlineDays*24*3600,
		Contact:       in.Contact,
		ContactTel:    in.Telephone,
		AddTime:       now,
	}
	companyCode, err := generateVideoCode(ctx, "company_code")
	if err != nil {
		return nil, err
	}
	personalCode, err := generateVideoCode(ctx, "personal_code")
	if err != nil {
		return nil, err
	}
	record.CompanyCode = companyCode
	record.PersonalCode = personalCode

	if err := global.GVA_DB.WithContext(ctx).Create(record).Error; err != nil {
		return nil, err
	}
	return record, nil
}

// CompanyList 企业端：我发出的视频面试列表（含简历姓名）
func (s *VideoInterviewService) CompanyList(ctx context.Context, companyUID uint64, info request.PageInfo) ([]VideoInterviewItem, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.VideoInterview{}).Where("company_uid = ?", companyUID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	var list []hrcModel.VideoInterview
	if err := db.Order("id desc").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return s.enrich(ctx, list, true, false), total, nil
}

// CompanyDetail 企业端：视频面试详情（归属校验）
func (s *VideoInterviewService) CompanyDetail(ctx context.Context, companyUID uint64, id uint64) (*hrcModel.VideoInterview, error) {
	var v hrcModel.VideoInterview
	err := global.GVA_DB.WithContext(ctx).Where("id = ? AND company_uid = ?", id, companyUID).First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrVideoInterviewNotFound
	}
	return &v, err
}

// CompanyDelete 企业端：删除视频面试邀请（归属校验）
func (s *VideoInterviewService) CompanyDelete(ctx context.Context, companyUID uint64, id uint64) error {
	res := global.GVA_DB.WithContext(ctx).Where("id = ? AND company_uid = ?", id, companyUID).Delete(&hrcModel.VideoInterview{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrVideoInterviewNotFound
	}
	return nil
}

// PersonalList 个人端：我收到的视频面试列表
func (s *VideoInterviewService) PersonalList(ctx context.Context, personalUID uint64, info request.PageInfo) ([]hrcModel.VideoInterview, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.VideoInterview{}).Where("personal_uid = ?", personalUID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	var list []hrcModel.VideoInterview
	if err := db.Order("id desc").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// PersonalDetail 个人端：视频面试详情（归属校验）
func (s *VideoInterviewService) PersonalDetail(ctx context.Context, personalUID uint64, id uint64) (*hrcModel.VideoInterview, error) {
	var v hrcModel.VideoInterview
	err := global.GVA_DB.WithContext(ctx).Where("id = ? AND personal_uid = ?", id, personalUID).First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrVideoInterviewNotFound
	}
	return &v, err
}

// RoomByCode 房间码查面试信息（company_code → 企业端；personal_code → 个人端）
func (s *VideoInterviewService) RoomByCode(ctx context.Context, code string) (*hrcModel.VideoInterview, int8, error) {
	if len(code) != 6 {
		return nil, 0, ErrVideoInterviewNotFound
	}
	var v hrcModel.VideoInterview
	err := global.GVA_DB.WithContext(ctx).Where("company_code = ?", code).First(&v).Error
	if err == nil {
		return &v, 2, nil // 企业端
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, 0, err
	}
	err = global.GVA_DB.WithContext(ctx).Where("personal_code = ?", code).First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, 0, ErrVideoInterviewNotFound
	}
	if err != nil {
		return nil, 0, err
	}
	return &v, 1, nil // 个人端
}

// AdminList 后台：视频面试列表（关键字跨 职位名/公司名/简历姓名）
func (s *VideoInterviewService) AdminList(ctx context.Context, info request.PageInfo, keyword string) ([]VideoInterviewItem, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.VideoInterview{})
	if keyword != "" {
		like := "%" + keyword + "%"
		var companyUIDs []uint64
		global.GVA_DB.WithContext(ctx).Model(&hrcModel.CompanyProfile{}).
			Where("companyname LIKE ?", like).Pluck("uid", &companyUIDs)
		var personalUIDs []uint64
		global.GVA_DB.WithContext(ctx).Model(&hrcModel.Resume{}).
			Where("fullname LIKE ?", like).Pluck("uid", &personalUIDs)
		db = db.Where("jobs_name LIKE ? OR company_uid IN ? OR personal_uid IN ?", like, companyUIDs, personalUIDs)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	var list []hrcModel.VideoInterview
	if err := db.Order("id desc").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return s.enrich(ctx, list, true, true), total, nil
}

// enrich 批量补充简历姓名 / 企业名（避免 1:N join 产生重复行）
func (s *VideoInterviewService) enrich(ctx context.Context, list []hrcModel.VideoInterview, needResume, needCompany bool) []VideoInterviewItem {
	out := make([]VideoInterviewItem, 0, len(list))
	if len(list) == 0 {
		return out
	}
	resumeID := map[uint64]uint64{}    // uid -> resume id
	resumeFull := map[uint64]string{}  // uid -> fullname
	companyName := map[uint64]string{} // uid -> companyname
	if needResume {
		var uids []uint64
		for _, v := range list {
			uids = append(uids, v.PersonalUID)
		}
		var resumes []hrcModel.Resume
		global.GVA_DB.WithContext(ctx).Where("uid IN ?", uids).Find(&resumes)
		for _, r := range resumes {
			if _, ok := resumeID[r.UID]; !ok {
				resumeID[r.UID] = r.ID
				resumeFull[r.UID] = r.FullName
			}
		}
	}
	if needCompany {
		var uids []uint64
		for _, v := range list {
			uids = append(uids, v.CompanyUID)
		}
		var profiles []hrcModel.CompanyProfile
		global.GVA_DB.WithContext(ctx).Where("uid IN ?", uids).Find(&profiles)
		for _, p := range profiles {
			companyName[p.UID] = companyNameStrPtr(p.CompanyName)
		}
	}
	for _, v := range list {
		item := VideoInterviewItem{VideoInterview: v}
		item.ResumeID = resumeID[v.PersonalUID]
		item.FullName = resumeFull[v.PersonalUID]
		item.CompanyName = companyName[v.CompanyUID]
		out = append(out, item)
	}
	return out
}

// getConfigValue 读取系统配置值（不存在返回空串）
func getConfigValue(ctx context.Context, name string) (string, error) {
	var c hrcModel.Config
	err := global.GVA_DB.WithContext(ctx).Where("name = ?", name).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	return c.Value, err
}

// generateVideoCode 生成唯一 6 位字母数字房间码（按列查重，冲突重试）
func generateVideoCode(ctx context.Context, column string) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for i := 0; i < 10; i++ {
		code, err := randomCode(charset, 6)
		if err != nil {
			return "", err
		}
		var n int64
		if err := global.GVA_DB.WithContext(ctx).Model(&hrcModel.VideoInterview{}).
			Where(column+" = ?", code).Count(&n).Error; err != nil {
			return "", err
		}
		if n == 0 {
			return code, nil
		}
	}
	return "", errors.New("生成房间码失败，请重试")
}

func randomCode(charset string, n int) (string, error) {
	b := make([]byte, n)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[idx.Int64()]
	}
	return string(b), nil
}

// companyNameStrPtr *string → string（空指针转空串）
func companyNameStrPtr(name *string) string {
	if name == nil {
		return ""
	}
	return *name
}
