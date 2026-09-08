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
	ErrCompanyCancellationPending   = errors.New("已有待处理的注销申请，请勿重复提交")
	ErrCompanyCancellationNotFound  = errors.New("注销申请不存在")
	ErrCompanyCancellationHandled   = errors.New("该申请已处理，请勿重复操作")
	ErrCompanyCancellationNoProfile = errors.New("请先完善企业资料再申请注销")
)

// CompanyCancellationService 企业注销服务（02 §2.8：申请→后台审批→清企业业务数据，账号保留）
type CompanyCancellationService struct{}

// ApplyCompanyCancellation 企业注销申请（短信二次确认；已有 status=0 未处理申请则拒绝）
func (s *CompanyCancellationService) ApplyCompanyCancellation(ctx context.Context, uid uint64, code string) error {
	var member hrcModel.Members
	if err := global.GVA_DB.WithContext(ctx).Where("uid = ? AND deleted_at = 0", uid).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAccountNotFound
		}
		return err
	}
	if member.Utype != 2 {
		return errors.New("仅企业账号可申请注销")
	}
	if member.Mobile == "" || len(member.Mobile) < 10 {
		return errors.New("当前账号未绑定手机号")
	}
	// 短信二次确认（cancellation 场景严格匹配，发码 type=cancellation）
	if err := (&AuthService{}).CheckSmsCode(member.Mobile, "cancellation", code); err != nil {
		return err
	}
	// 企业资料已填写（企业名非空）才可申请
	var profile hrcModel.CompanyProfile
	if err := global.GVA_DB.WithContext(ctx).Where("uid = ?", uid).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCompanyCancellationNoProfile
		}
		return err
	}
	if profile.CompanyName == nil || strings.TrimSpace(*profile.CompanyName) == "" {
		return ErrCompanyCancellationNoProfile
	}
	// 防重：已有未处理申请
	var count int64
	if err := global.GVA_DB.WithContext(ctx).Model(&hrcModel.CompanyCancellationApply{}).
		Where("uid = ? AND status = 0", uid).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrCompanyCancellationPending
	}
	return global.GVA_DB.WithContext(ctx).Create(&hrcModel.CompanyCancellationApply{
		UID:         uid,
		CompanyID:   profile.ID,
		CompanyName: *profile.CompanyName,
		AddTime:     hrcModel.Now(),
		Status:      0,
	}).Error
}

// GetCompanyCancellationStatus 查询企业注销申请状态（最近一条）
func (s *CompanyCancellationService) GetCompanyCancellationStatus(ctx context.Context, uid uint64) (*hrcModel.CompanyCancellationApply, error) {
	var apply hrcModel.CompanyCancellationApply
	err := global.GVA_DB.WithContext(ctx).Where("uid = ?", uid).Order("id desc").First(&apply).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCompanyCancellationNotFound
	}
	return &apply, err
}

// CompanyCancellationItem 后台列表项（join ms_members 补 username/mobile）
type CompanyCancellationItem struct {
	hrcModel.CompanyCancellationApply
	Username string `json:"username"`
	Mobile   string `json:"mobile"`
}

// CompanyCancellationList 后台注销申请列表（分页 + 状态筛选；status 0=全部 1=待处理 2=已处理）
func (s *CompanyCancellationService) CompanyCancellationList(ctx context.Context, info request.PageInfo, status int8) (list []CompanyCancellationItem, total int64, err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.CompanyCancellationApply{})
	if status == 1 {
		db = db.Where("status = 0")
	}
	if status == 2 {
		db = db.Where("status = 1")
	}
	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var applies []hrcModel.CompanyCancellationApply
	limit, offset := info.LimitOffset()
	if err = db.Order("id desc").Limit(limit).Offset(offset).Find(&applies).Error; err != nil {
		return nil, 0, err
	}
	if len(applies) == 0 {
		return []CompanyCancellationItem{}, total, nil
	}
	// join 等价：批量查 members 组装 username/mobile
	uids := make([]uint64, 0, len(applies))
	for _, a := range applies {
		uids = append(uids, a.UID)
	}
	var members []hrcModel.Members
	if err = global.GVA_DB.WithContext(ctx).Where("uid IN ?", uids).Find(&members).Error; err != nil {
		return nil, 0, err
	}
	memberMap := make(map[uint64]hrcModel.Members, len(members))
	for _, m := range members {
		memberMap[m.UID] = m
	}
	list = make([]CompanyCancellationItem, 0, len(applies))
	for _, a := range applies {
		item := CompanyCancellationItem{CompanyCancellationApply: a}
		if m, ok := memberMap[a.UID]; ok {
			item.Username = m.Username
			item.Mobile = m.Mobile
		}
		list = append(list, item)
	}
	return list, total, nil
}

// HandleCompanyCancellation 后台处理：事务内清企业业务数据 → status=1 + finishtime，账号保留
func (s *CompanyCancellationService) HandleCompanyCancellation(ctx context.Context, id uint64) error {
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var apply hrcModel.CompanyCancellationApply
		if err := tx.Where("id = ?", id).First(&apply).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCompanyCancellationNotFound
			}
			return err
		}
		if apply.Status == 1 {
			return ErrCompanyCancellationHandled
		}
		if err := clearCompanyBusinessData(tx, apply.UID); err != nil {
			return err
		}
		return tx.Model(&hrcModel.CompanyCancellationApply{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":     1,
			"finishtime": hrcModel.Now(),
		}).Error
	})
}

// DeleteCompanyCancellation 后台硬删除申请记录（任意状态）
func (s *CompanyCancellationService) DeleteCompanyCancellation(ctx context.Context, id uint64) error {
	result := global.GVA_DB.WithContext(ctx).Delete(&hrcModel.CompanyCancellationApply{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCompanyCancellationNotFound
	}
	return nil
}

// clearCompanyBusinessData 清企业业务数据（02 §2.8 清单，按一期实际建表裁剪：已建表删、未建表跳过）
// 企业 uid 列映射：ms_jobs 系 = uid；投递/下载/面试/收藏 = company_uid；套餐/相册 = uid
// 注：套餐重置（赠送默认套餐）与模板重置依赖 M4 套餐/模板体系，一期 ms_members_setmeal 未建，随 M4 实现时补充
func clearCompanyBusinessData(tx *gorm.DB, uid uint64) error {
	// 企业主体（一期已建）
	if err := tx.Where("uid = ?", uid).Delete(&hrcModel.CompanyProfile{}).Error; err != nil {
		return err
	}
	// uid 列表（已建才删）
	for _, t := range []string{"ms_jobs", "ms_jobs_tmp", "ms_jobs_contact", "ms_jobs_tag", "ms_members_setmeal", "ms_company_img"} {
		if !tx.Migrator().HasTable(t) {
			continue
		}
		if err := tx.Exec("DELETE FROM `"+t+"` WHERE uid = ?", uid).Error; err != nil {
			return err
		}
	}
	// company_uid 列表（投递/下载/面试/收藏，已建才删）
	for _, t := range []string{"ms_personal_jobs_apply", "ms_company_down_resume", "ms_company_interview", "ms_company_favorites"} {
		if !tx.Migrator().HasTable(t) {
			continue
		}
		if err := tx.Exec("DELETE FROM `"+t+"` WHERE company_uid = ?", uid).Error; err != nil {
			return err
		}
	}
	return nil
}
