package hrc

import (
	"context"
	"errors"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
)

var (
	ErrAppealParam     = errors.New("申诉参数不完整")
	ErrAppealRateLimit = errors.New("申诉提交过于频繁，请稍后再试")
)

// AppealService 账号申诉服务（02 §2.6）
type AppealService struct{}

// SubmitAppeal 提交申诉（限流保护：同手机号 3 次/小时）
// uid：已登录用户自动带入；未登录为 0（后台按手机号匹配，见 02 §2.6）
func (s *AppealService) SubmitAppeal(uid uint64, realname, mobile, email, description string) (uint64, error) {
	if !mobileRegex.MatchString(mobile) {
		return 0, ErrMobileFormat
	}
	if realname == "" || description == "" {
		return 0, ErrAppealParam
	}
	// 限流：同号 3 次/小时
	if global.GVA_REDIS != nil {
		ctx := context.Background()
		key := "hrc:appeal:rate:" + mobile
		count, _ := global.GVA_REDIS.Incr(ctx, key).Result()
		if count == 1 {
			_ = global.GVA_REDIS.Expire(ctx, key, time.Hour).Err()
		}
		if count > 3 {
			return 0, ErrAppealRateLimit
		}
	}
	appeal := hrcModel.MembersAppeal{
		UID:         uid,
		RealName:    realname,
		Mobile:      mobile,
		Email:       email,
		Description: description,
		AddTime:     hrcModel.Now(),
		Status:      0,
	}
	if err := global.GVA_DB.Create(&appeal).Error; err != nil {
		return 0, err
	}
	return appeal.ID, nil
}

// GetAppealStatus 按手机号查询申诉进度（最近 10 条，倒序）
func (s *AppealService) GetAppealStatus(mobile string) ([]hrcModel.MembersAppeal, error) {
	if !mobileRegex.MatchString(mobile) {
		return nil, ErrMobileFormat
	}
	var list []hrcModel.MembersAppeal
	err := global.GVA_DB.Where("mobile = ?", mobile).Order("id desc").Limit(10).Find(&list).Error
	return list, err
}
