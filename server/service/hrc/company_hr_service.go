package hrc

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrHRMobileExists = errors.New("该手机号已注册")
	ErrHRNotFound     = errors.New("HR 子账号不存在")
	ErrHRStatusValue  = errors.New("状态仅支持 1=启用 2=禁用")
)

// CompanyHRItem 企业 HR 子账号列表项
type CompanyHRItem struct {
	UID      uint64    `json:"uid"`
	Username string    `json:"username"`
	Mobile   string    `json:"mobile"`
	RealName string    `json:"realName"` // HR 姓名（在线对话展示接待人）
	Status   int8      `json:"status"`   // 1=启用 2=禁用
	RegTime  time.Time `json:"regTime"`
}

// CompanyHRService 企业多 HR 子账号（#21：主账号创建/管理，共享企业数据）
type CompanyHRService struct{}

// Create 创建 HR 子账号（归属企业主账号；手机号+验证码+密码，数据共享企业主体）
func (s *CompanyHRService) Create(ctx context.Context, ownerUID uint64, mobile, password, realname, code string) (*CompanyHRItem, error) {
	mobile = strings.TrimSpace(mobile)
	if err := (&AuthService{}).ValidateMobile(mobile); err != nil {
		return nil, err
	}
	// 手机号验证码校验（场景 hr；支持统一万能验证码）
	if err := (&AuthService{}).CheckSmsCode(mobile, "hr", strings.TrimSpace(code)); err != nil {
		return nil, err
	}
	if err := validateMemberPassword(password); err != nil {
		return nil, err
	}
	db := global.GVA_DB.WithContext(ctx)
	var count int64
	if err := db.Model(&hrcModel.Members{}).Where("mobile = ? AND deleted_at IS NULL", mobile).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrHRMobileExists
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	hr := &hrcModel.Members{
		Utype:          2,
		CompanyUID:     ownerUID,
		Username:       mobile, // 一期用户名=手机号
		Mobile:         mobile,
		MobileAudit:    1,
		Password:       string(hash),
		RegTime:        hrcModel.Now(),
		Status:         1,
		InvitationCode: randomInviteCode(),
	}
	if err := db.Create(hr).Error; err != nil {
		if isDuplicateKeyErr(err) {
			return nil, ErrHRMobileExists
		}
		return nil, err
	}
	// HR 姓名存 MembersInfo（在线对话中对端可显示接待人姓名）
	if name := strings.TrimSpace(realname); name != "" {
		if err := db.Create(&hrcModel.MembersInfo{UID: hr.UID, RealName: name}).Error; err != nil {
			return nil, err
		}
	}
	return &CompanyHRItem{UID: hr.UID, Username: hr.Username, Mobile: hr.Mobile, RealName: strings.TrimSpace(realname), Status: hr.Status, RegTime: hr.RegTime}, nil
}

// List 企业 HR 子账号列表
func (s *CompanyHRService) List(ctx context.Context, ownerUID uint64) ([]CompanyHRItem, error) {
	var members []hrcModel.Members
	if err := global.GVA_DB.WithContext(ctx).
		Where("company_uid = ? AND deleted_at IS NULL", ownerUID).
		Order("uid asc").Find(&members).Error; err != nil {
		return nil, err
	}
	items := make([]CompanyHRItem, 0, len(members))
	if len(members) == 0 {
		return items, nil
	}
	uids := make([]uint64, 0, len(members))
	for _, m := range members {
		uids = append(uids, m.UID)
	}
	var infos []hrcModel.MembersInfo
	_ = global.GVA_DB.WithContext(ctx).Select("uid", "realname").Where("uid IN ?", uids).Find(&infos).Error
	nameMap := make(map[uint64]string, len(infos))
	for _, info := range infos {
		nameMap[info.UID] = info.RealName
	}
	for _, m := range members {
		items = append(items, CompanyHRItem{UID: m.UID, Username: m.Username, Mobile: m.Mobile, RealName: nameMap[m.UID], Status: m.Status, RegTime: m.RegTime})
	}
	return items, nil
}

// SetStatus 启用/禁用 HR 子账号（禁用后其 token 失效）
func (s *CompanyHRService) SetStatus(ctx context.Context, ownerUID uint64, hrUID uint64, status int8) error {
	if status != 1 && status != 2 {
		return ErrHRStatusValue
	}
	res := global.GVA_DB.WithContext(ctx).Model(&hrcModel.Members{}).
		Where("uid = ? AND company_uid = ? AND deleted_at IS NULL", hrUID, ownerUID).
		Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrHRNotFound
	}
	if status == 2 {
		middlewarehrc.InvalidateSession(hrUID)
	}
	return nil
}

// ResetPassword 重置 HR 子账号密码（重置后踢出其会话）
func (s *CompanyHRService) ResetPassword(ctx context.Context, ownerUID uint64, hrUID uint64, password string) error {
	if err := validateMemberPassword(password); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	res := global.GVA_DB.WithContext(ctx).Model(&hrcModel.Members{}).
		Where("uid = ? AND company_uid = ? AND deleted_at IS NULL", hrUID, ownerUID).
		Update("password", string(hash))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrHRNotFound
	}
	middlewarehrc.InvalidateSession(hrUID)
	return nil
}

// Delete 移除 HR 子账号（真正删除会员与其资料，释放手机号/用户名唯一索引；历史会话保留）
func (s *CompanyHRService) Delete(ctx context.Context, ownerUID uint64, hrUID uint64) error {
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var hr hrcModel.Members
		if err := tx.Where("uid = ? AND company_uid = ?", hrUID, ownerUID).First(&hr).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrHRNotFound
			}
			return err
		}
		middlewarehrc.InvalidateSession(hrUID)
		if err := tx.Where("uid = ?", hrUID).Delete(&hrcModel.MembersInfo{}).Error; err != nil {
			return err
		}
		// Members 的 DeletedAt 非 gorm.DeletedAt，Delete 为物理删除；HR 子账号无需保留审计
		return tx.Where("uid = ?", hrUID).Delete(&hrcModel.Members{}).Error
	})
}
