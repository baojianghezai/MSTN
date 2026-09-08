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

var ErrCompanyProfileNotFound = errors.New("企业资料不存在")

// CompanyAuditService 企业资质审核后台服务（09 §4.6.2 #134/#134a/#135）
type CompanyAuditService struct{}

// CompanyList 企业资料列表（分页 + audit 状态筛选 + 企业名关键字）
// audit：<0 全部；0=草稿 1=通过 2=待审 3=不通过
func (s *CompanyAuditService) CompanyList(ctx context.Context, info request.PageInfo, audit int8, keyword string) (list []hrcModel.CompanyProfile, total int64, err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.CompanyProfile{})
	if audit >= 0 {
		db = db.Where("audit = ?", audit)
	}
	if keyword != "" {
		db = db.Where("companyname LIKE ?", "%"+keyword+"%")
	}
	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	err = db.Order("id desc").Limit(limit).Offset(offset).Find(&list).Error
	return list, total, err
}

// CompanyDetail 企业资料详情（含 Logo/营业执照证照，审核用）
func (s *CompanyAuditService) CompanyDetail(ctx context.Context, id uint64) (*hrcModel.CompanyProfile, error) {
	var p hrcModel.CompanyProfile
	err := global.GVA_DB.WithContext(ctx).Where("id = ?", id).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCompanyProfileNotFound
	}
	return &p, err
}

// CompanyAudit 企业资质审核（通过 audit=1 / 不通过 audit=3）
// 不通过时 reason 必填并写入 ms_audit_reason（Type=AuditTypeCompany，01 §3.1）
func (s *CompanyAuditService) CompanyAudit(ctx context.Context, id uint64, audit int8, reason string) error {
	if audit != 1 && audit != 3 {
		return errors.New("审核状态非法（仅支持 1=通过 3=不通过）")
	}
	if audit == 3 && strings.TrimSpace(reason) == "" {
		return errors.New("不通过时必须填写原因")
	}
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var p hrcModel.CompanyProfile
		if err := tx.Where("id = ?", id).First(&p).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCompanyProfileNotFound
			}
			return err
		}
		if err := tx.Model(&hrcModel.CompanyProfile{}).Where("id = ?", id).Update("audit", audit).Error; err != nil {
			return err
		}
		// 不通过：写审核原因（供企业端回显）
		if audit == 3 {
			return tx.Create(&hrcModel.AuditReason{
				Type:    hrcModel.AuditTypeCompany,
				TypeID:  p.ID,
				Reason:  reason,
				AddTime: hrcModel.Now(),
			}).Error
		}
		return nil
	})
}
