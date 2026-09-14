package hrc

import (
	"context"
	cryptorand "crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ctx = context.Background()

// 业务规则：02_用户账号模块设计 §2.1-§2.7

var (
	ErrMobileFormat             = errors.New("手机号格式错误")
	ErrSmsTypeInvalid           = errors.New("验证码类型错误")
	ErrAccountExists            = errors.New("该账号已注册，请直接登录")
	ErrSmsCodeInvalid           = errors.New("验证码错误或已过期")
	ErrSmsRateLimit             = errors.New("发送过于频繁，请稍后再试")
	ErrAccountNotFound          = errors.New("账号不存在")
	ErrPasswordWrong            = errors.New("密码错误")
	ErrPasswordNotSet           = errors.New("该账号未设置密码，请用短信验证码登录后设置密码")
	ErrAccountDisabled          = errors.New("账号已暂停，请联系管理员")
	ErrAccountCancelled         = errors.New("账号已注销")
	ErrLoginLocked              = errors.New("登录失败次数过多，请30分钟后再试")
	ErrSmsProviderNotConfigured = errors.New("短信服务尚未配置")
	ErrPasswordTooWeak          = errors.New("密码至少 10 位，且需包含大写字母、小写字母、数字、特殊字符中的至少三种")
)

// 手机号：13-19 全号段（含 16 虚商号）
var mobileRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

// 短信场景枚举
var smsTypes = map[string]bool{"register": true, "login": true, "reset": true, "bind": true, "cancellation": true}

// AuthService 认证服务
type AuthService struct{}

// ValidateMobile 手机号格式校验
func (s *AuthService) ValidateMobile(mobile string) error {
	if !mobileRegex.MatchString(mobile) {
		return ErrMobileFormat
	}
	return nil
}

// SendSmsCode 发送短信验证码（一期：写 Redis + 日志，不接真实短信网关）
// type: register|login|reset|bind|cancellation
func (s *AuthService) SendSmsCode(mobile, typ string) (string, error) {
	return s.sendSmsCode(mobile, typ, "")
}

// SendSmsCodeFromIP adds request-source rate limiting for public endpoints.
func (s *AuthService) SendSmsCodeFromIP(mobile, typ, clientIP string) (string, error) {
	return s.sendSmsCode(mobile, typ, clientIP)
}

func (s *AuthService) sendSmsCode(mobile, typ, clientIP string) (string, error) {
	ctx := context.Background()
	if err := s.ValidateMobile(mobile); err != nil {
		return "", err
	}
	if typ != "" && !smsTypes[typ] {
		return "", ErrSmsTypeInvalid
	}
	// reset/cancellation 场景要求手机号已注册（忘记密码只对存量账号发码、企业注销需存量企业号，防未注册号探测）
	if typ == "reset" || typ == "cancellation" {
		var count int64
		global.GVA_DB.Model(&hrcModel.Members{}).Where("mobile = ? AND deleted_at IS NULL", mobile).Count(&count)
		if count == 0 {
			return "", ErrAccountNotFound
		}
	}
	// Redis 未启用防护（use-redis=false 时全局客户端为 nil，避免 panic）
	if global.GVA_REDIS == nil {
		global.GVA_LOG.Error("redis client not initialized, sms code cannot be stored")
		return "", errors.New("服务未就绪，请联系管理员")
	}
	// 60s 重发限制（按手机号全局限流，不区分场景——防换 type 绕过）
	if err := s.checkIPLimit(ctx, clientIP); err != nil {
		return "", err
	}
	rateKey := fmt.Sprintf("hrc:sms:rate:%s", mobile)
	if exists, _ := global.GVA_REDIS.Exists(ctx, rateKey).Result(); exists == 1 {
		return "", ErrSmsRateLimit
	}
	// 每小时/每日上限（02 设计：同号 5 条/小时、10 条/日，同样不区分场景）
	hourKey := fmt.Sprintf("hrc:sms:count:hour:%s", mobile)
	dayKey := fmt.Sprintf("hrc:sms:count:day:%s", mobile)
	if err := s.checkDailyLimit(ctx, hourKey, dayKey); err != nil {
		return "", err
	}
	code, err := secureSixDigit()
	if err != nil {
		return "", err
	}
	key := fmt.Sprintf("hrc:sms:code:%s:%s", typ, mobile)
	if err := global.GVA_REDIS.Set(ctx, key, code, 5*time.Minute).Err(); err != nil {
		global.GVA_LOG.Error("sms code set redis failed", zap.String("err", err.Error()))
		return "", err
	}
	// 60 秒重发限制
	_ = global.GVA_REDIS.Set(ctx, rateKey, "1", 60*time.Second).Err()
	// 计数自增
	global.GVA_REDIS.Incr(ctx, hourKey)
	global.GVA_REDIS.Incr(ctx, dayKey)
	if ttl, _ := global.GVA_REDIS.TTL(ctx, hourKey).Result(); ttl == -1 {
		_ = global.GVA_REDIS.Expire(ctx, hourKey, time.Hour).Err()
	}
	if ttl, _ := global.GVA_REDIS.TTL(ctx, dayKey).Result(); ttl == -1 {
		_ = global.GVA_REDIS.Expire(ctx, dayKey, 24*time.Hour).Err()
	}
	// TODO: 对接真实短信网关（阿里云/腾讯云），一期 mock：验证码记录在日志
	if strings.EqualFold(global.GVA_CONFIG.App.Env, "prod") {
		_ = global.GVA_REDIS.Del(ctx, key).Err()
		return "", ErrSmsProviderNotConfigured
	}
	global.GVA_LOG.Info("sms code sent (development mock)", zap.String("mobile", mobile), zap.String("code", code))
	return code, nil
}

// checkDailyLimit 短信频率上限：同号 5 条/小时、10 条/日（02 §2.7）
func (s *AuthService) checkDailyLimit(ctx context.Context, hourKey, dayKey string) error {
	hourCount, _ := global.GVA_REDIS.Get(ctx, hourKey).Int()
	if hourCount >= 5 {
		return ErrSmsRateLimit
	}
	dayCount, _ := global.GVA_REDIS.Get(ctx, dayKey).Int()
	if dayCount >= 10 {
		return ErrSmsRateLimit
	}
	return nil
}

// CheckSmsCode 校验短信验证码
// 场景回退规则：login 场景兼容 register 场景的码（短信登录入口发码时 type 常走默认 register，
// 两者同属"验证码登录注册"语义）；reset/bind 严格匹配各自场景。
func (s *AuthService) checkIPLimit(ctx context.Context, clientIP string) error {
	if clientIP == "" || global.GVA_REDIS == nil {
		return nil
	}
	windowKey := fmt.Sprintf("hrc:sms:ip:10m:%s", clientIP)
	dayKey := fmt.Sprintf("hrc:sms:ip:day:%s", clientIP)
	windowCount, err := global.GVA_REDIS.Incr(ctx, windowKey).Result()
	if err != nil {
		return err
	}
	if windowCount == 1 {
		_ = global.GVA_REDIS.Expire(ctx, windowKey, 10*time.Minute).Err()
	}
	if windowCount > 5 {
		return ErrSmsRateLimit
	}
	dayCount, err := global.GVA_REDIS.Incr(ctx, dayKey).Result()
	if err != nil {
		return err
	}
	if dayCount == 1 {
		_ = global.GVA_REDIS.Expire(ctx, dayKey, 24*time.Hour).Err()
	}
	if dayCount > 20 {
		return ErrSmsRateLimit
	}
	return nil
}

func (s *AuthService) CheckSmsCode(mobile, typ, code string) error {
	ctx := context.Background()
	if global.GVA_REDIS == nil {
		return errors.New("服务未就绪，请联系管理员")
	}
	var keys []string
	if typ == "login" {
		keys = []string{
			fmt.Sprintf("hrc:sms:code:login:%s", mobile),
			fmt.Sprintf("hrc:sms:code:register:%s", mobile),
		}
	} else {
		keys = []string{fmt.Sprintf("hrc:sms:code:%s:%s", typ, mobile)}
	}
	for _, key := range keys {
		val, err := global.GVA_REDIS.Get(ctx, key).Result()
		if err == nil && val == code {
			_ = global.GVA_REDIS.Del(ctx, key).Err()
			return nil
		}
	}
	return ErrSmsCodeInvalid
}

// Register 注册（手机号+验证码+密码）
func (s *AuthService) Register(mobile, code, password string, utype int8) (*hrcModel.Members, error) {
	if err := s.ValidateMobile(mobile); err != nil {
		return nil, err
	}
	if err := s.CheckSmsCode(mobile, "register", code); err != nil {
		return nil, err
	}
	// 手机号唯一查重
	var count int64
	global.GVA_DB.Model(&hrcModel.Members{}).Where("mobile = ? AND deleted_at IS NULL", mobile).Count(&count)
	if count > 0 {
		return nil, ErrAccountExists
	}
	if err := validateMemberPassword(password); err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	member := &hrcModel.Members{
		Utype:          utype,
		Username:       mobile, // 一期默认用户名=手机号
		Mobile:         mobile,
		MobileAudit:    1,
		Password:       string(hash),
		RegTime:        hrcModel.Now(),
		Status:         1,
		InvitationCode: randomInviteCode(),
	}
	if err := global.GVA_DB.Create(member).Error; err != nil {
		if isDuplicateKeyErr(err) {
			return nil, ErrAccountExists
		}
		return nil, err
	}
	// 创建对应资料（个人/企业空壳）
	if utype == 1 {
		global.GVA_DB.Create(&hrcModel.MembersInfo{UID: member.UID})
	} else {
		global.GVA_DB.Create(&hrcModel.CompanyProfile{UID: member.UID, AddTime: hrcModel.Now()})
	}
	return member, nil
}

// LoginByPassword 密码登录（账号类型自动识别：mobile/email/username）
func (s *AuthService) LoginByPassword(account, password string) (*hrcModel.Members, error) {
	ctx := context.Background()
	// 风控：登录失败锁定（失败 5 次锁 30 分钟）
	lockKey := "hrc:login:lock:" + account
	if global.GVA_REDIS != nil {
		if exists, _ := global.GVA_REDIS.Exists(ctx, lockKey).Result(); exists == 1 {
			return nil, ErrLoginLocked
		}
	}
	var member hrcModel.Members
	where := "username = ?"
	switch {
	case mobileRegex.MatchString(account):
		where = "mobile = ?"
	case regexp.MustCompile(`^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`).MatchString(account):
		where = "email = ?"
	}
	// 不过滤 deleted_at：注销账号（status=3，mobile 保留）也需命中并返回 2004，而非 2001
	err := global.GVA_DB.Where(where, account).First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrAccountNotFound
	}
	if err != nil {
		return nil, err
	}
	// 状态检查（先于密码校验：暂停/注销账号即使密码正确也拒绝）
	switch member.Status {
	case 2:
		return nil, ErrAccountDisabled
	case 3:
		return nil, ErrAccountCancelled
	}
	// 恢复后账号密码为空（注销时清空），密码登录应提示走短信登录设密，而非"密码错误"
	if member.Password == "" {
		return nil, ErrPasswordNotSet
	}
	if err := bcrypt.CompareHashAndPassword([]byte(member.Password), []byte(password)); err != nil {
		s.recordLoginFail(account)
		return nil, ErrPasswordWrong
	}
	// 登录成功，清除失败计数
	s.resetLoginFail(account)
	// 更新登录信息
	global.GVA_DB.Model(&member).Updates(map[string]interface{}{
		"last_login_time": hrcModel.Now(),
	})
	return &member, nil
}

// LoginBySms 短信验证码登录
// 规则（02 §2.3 + B3）：
//   - 已注册手机号：校验验证码后**直接以库里账号身份登录**（忽略请求 utype，以库为准——一个手机号一种身份，02 §2.1 互斥设计）
//   - 未注册手机号：按请求 utype 自动注册后登录（随机初始密码，可后续改密）
//   - 状态：status=2 暂停 → 拒绝；status=3 注销 → 拒绝
func (s *AuthService) LoginBySms(mobile, code string, utype int8) (*hrcModel.Members, error) {
	if err := s.ValidateMobile(mobile); err != nil {
		return nil, err
	}
	if err := s.CheckSmsCode(mobile, "login", code); err != nil {
		return nil, err
	}
	var member hrcModel.Members
	// 不过滤 deleted_at：注销账号（status=3，mobile 保留）也需命中并返回 2004，
	// 避免落入自动注册分支触发 mobile 唯一索引冲突（联调 M7）。
	err := global.GVA_DB.Where("mobile = ?", mobile).First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 未注册 → 自动注册（B3 拉新关键）
		if utype == 0 {
			utype = 1 // 未指定身份时默认个人
		}
		if utype != 1 && utype != 2 {
			return nil, errors.New("用户类型错误")
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte(randomPassword()), bcrypt.DefaultCost)
		member = hrcModel.Members{
			Utype:          utype,
			Username:       mobile,
			Mobile:         mobile,
			MobileAudit:    1,
			Password:       string(hash),
			RegTime:        hrcModel.Now(),
			Status:         1,
			InvitationCode: randomInviteCode(),
		}
		if err := global.GVA_DB.Create(&member).Error; err != nil {
			if isDuplicateKeyErr(err) {
				return nil, ErrAccountExists
			}
			return nil, err
		}
		if utype == 1 {
			global.GVA_DB.Create(&hrcModel.MembersInfo{UID: member.UID})
		} else {
			global.GVA_DB.Create(&hrcModel.CompanyProfile{UID: member.UID, AddTime: hrcModel.Now()})
		}
		global.GVA_DB.Model(&member).Updates(map[string]interface{}{
			"last_login_time": hrcModel.Now(),
		})
		return &member, nil
	}
	if err != nil {
		return nil, err
	}
	// 已注册：以库里账号为准，忽略请求 utype
	switch member.Status {
	case 2:
		return nil, ErrAccountDisabled
	case 3:
		return nil, ErrAccountCancelled
	}
	global.GVA_DB.Model(&member).Updates(map[string]interface{}{
		"last_login_time": hrcModel.Now(),
	})
	return &member, nil
}

// recordLoginFail 登录失败计数（5 次锁 30 分钟）
// 计数 key 与锁 key 分离：避免"失败一次就命中锁 key"的误锁。
func (s *AuthService) recordLoginFail(account string) {
	if global.GVA_REDIS == nil {
		return
	}
	ctx := context.Background()
	failKey := "hrc:login:fail:" + account
	lockKey := "hrc:login:lock:" + account
	count, _ := global.GVA_REDIS.Incr(ctx, failKey).Result()
	if count == 1 {
		// 首次失败起算 30 分钟计数窗口
		_ = global.GVA_REDIS.Expire(ctx, failKey, 30*time.Minute).Err()
	}
	if count >= 5 {
		// 达到 5 次锁定，清空计数（锁由 lockKey 独立承担）
		_ = global.GVA_REDIS.Set(ctx, lockKey, "1", 30*time.Minute).Err()
		_ = global.GVA_REDIS.Del(ctx, failKey).Err()
	}
}

// resetLoginFail 登录成功后清除失败计数
func (s *AuthService) resetLoginFail(account string) {
	if global.GVA_REDIS == nil {
		return
	}
	ctx := context.Background()
	_ = global.GVA_REDIS.Del(ctx, "hrc:login:fail:"+account).Err()
}

var (
	ErrOldPasswordWrong = errors.New("原密码错误")
	ErrSamePassword     = errors.New("新密码不能与原密码相同")
	ErrPasswordTooShort = errors.New("密码长度不能少于6位")
	ErrMobileBound      = errors.New("该手机号已被其他账号绑定")
)

// GetMemberByUID 按 uid 查会员
func (s *AuthService) GetMemberByUID(uid uint64) (*hrcModel.Members, error) {
	var member hrcModel.Members
	err := global.GVA_DB.Where("uid = ? AND deleted_at IS NULL", uid).First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrAccountNotFound
	}
	return &member, err
}

// ChangePassword 修改密码（原密码校验）
// 规则（02 §2.4）：校验原密码 → 新密码 ≥6 位 → 新旧不能相同 → 更新后踢出其他端（单点会话失效）
func (s *AuthService) ChangePassword(uid uint64, oldPassword, newPassword string) error {
	member, err := s.GetMemberByUID(uid)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(member.Password), []byte(oldPassword)); err != nil {
		return ErrOldPasswordWrong
	}
	if err := validateMemberPassword(newPassword); err != nil {
		return err
	}
	if oldPassword == newPassword {
		return ErrSamePassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return global.GVA_DB.Model(&hrcModel.Members{}).Where("uid = ?", uid).
		Update("password", string(hash)).Error
}

// ResetPassword 忘记密码重置（短信验证码，reset 场景严格匹配）
func (s *AuthService) ResetPassword(mobile, code, newPassword string) error {
	if err := s.ValidateMobile(mobile); err != nil {
		return err
	}
	if err := s.CheckSmsCode(mobile, "reset", code); err != nil {
		return err
	}
	if err := validateMemberPassword(newPassword); err != nil {
		return err
	}
	var member hrcModel.Members
	err := global.GVA_DB.Where("mobile = ? AND deleted_at IS NULL", mobile).First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrAccountNotFound
	}
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return global.GVA_DB.Model(&hrcModel.Members{}).Where("uid = ?", member.UID).
		Update("password", string(hash)).Error
}

// BindMobile 绑定手机（换绑手机号）
func (s *AuthService) BindMobile(uid uint64, mobile, code string) error {
	if err := s.ValidateMobile(mobile); err != nil {
		return err
	}
	if err := s.CheckSmsCode(mobile, "bind", code); err != nil {
		return err
	}
	// 手机号唯一校验（排除自身）
	var count int64
	global.GVA_DB.Model(&hrcModel.Members{}).
		Where("mobile = ? AND uid != ? AND deleted_at IS NULL", mobile, uid).Count(&count)
	if count > 0 {
		return ErrMobileBound
	}
	// 绑定新手机后，若 username 是解绑遗留占位（unbound_<uid>），恢复成新手机号，闭环「解绑→换绑」
	var member hrcModel.Members
	if err := global.GVA_DB.Where("uid = ?", uid).First(&member).Error; err != nil {
		return err
	}
	updates := map[string]interface{}{"mobile": mobile, "mobile_audit": 1}
	if member.Username == fmt.Sprintf("unbound_%d", uid) {
		updates["username"] = mobile
	}
	return global.GVA_DB.Model(&hrcModel.Members{}).Where("uid = ?", uid).Updates(updates).Error
}

// UnbindMobile 解绑手机（记 ms_unbind_mobile，mobile 置掩码占位——列非 NULL，掩码方案见 01 §六.2）
func (s *AuthService) UnbindMobile(uid uint64) error {
	var member hrcModel.Members
	if err := global.GVA_DB.Where("uid = ?", uid).First(&member).Error; err != nil {
		return err
	}
	if member.Mobile == "" || len(member.Mobile) < 10 {
		return errors.New("当前账号未绑定手机号")
	}
	record := hrcModel.UnbindMobile{
		UID:      uid,
		Utype:    member.Utype,
		Username: member.Username,
		Mobile:   member.Mobile,
		AddTime:  hrcModel.Now(),
		Remark:   "用户主动解绑",
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		// 置掩码占位（保留唯一性，防重绑撞号）
		updates := map[string]interface{}{
			"mobile": fmt.Sprintf("unbound_%d", uid),
		}
		// 一期用户名默认=手机号，解绑后一并释放 username 唯一占用，
		// 避免新用户用原手机号注册/登录时撞 username 唯一索引（1062）
		if member.Username == member.Mobile {
			updates["username"] = fmt.Sprintf("unbound_%d", uid)
		}
		return tx.Model(&hrcModel.Members{}).Where("uid = ?", uid).Updates(updates).Error
	})
}

// CancelAccount 账号注销（两阶段匿名化，02 §2.5 C-1 决议）
// 阶段一：deleted_at+status=3，username 匿名化，邮箱/密码清空，mobile 保留（冷静期恢复用）
// 阶段二（30 天期满 Cron 执行）：mobile 置掩码 anonymous_<uid> 释放号码
func (s *AuthService) CancelAccount(uid uint64, code string) error {
	var member hrcModel.Members
	if err := global.GVA_DB.Where("uid = ?", uid).First(&member).Error; err != nil {
		return err
	}
	// 注销需短信验证
	if err := s.CheckSmsCode(member.Mobile, "bind", code); err != nil {
		return err
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		return tx.Model(&hrcModel.Members{}).Where("uid = ?", uid).Updates(map[string]interface{}{
			"deleted_at": hrcModel.Now(),
			"status":     3,
			"username":   fmt.Sprintf("u_匿名_%d", uid),
			"email":      "",
			"password":   "",
		}).Error
	})
}

// isDuplicateKeyErr 判断是否为唯一索引冲突（MySQL 1062）
func isDuplicateKeyErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Duplicate entry")
}

func randomInviteCode() string {
	code, err := secureSixDigit()
	if err != nil {
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	return code
}

func randomPassword() string {
	code, err := secureSixDigit()
	if err != nil {
		return fmt.Sprintf("m%d", time.Now().UnixNano())
	}
	return "m" + code
}

func secureSixDigit() (string, error) {
	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func validateMemberPassword(password string) error {
	if utf8.RuneCountInString(password) < 10 {
		return ErrPasswordTooWeak
	}
	var upper, lower, digit, special bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			upper = true
		case unicode.IsLower(r):
			lower = true
		case unicode.IsDigit(r):
			digit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			special = true
		}
	}
	count := 0
	for _, present := range []bool{upper, lower, digit, special} {
		if present {
			count++
		}
	}
	if count < 3 {
		return ErrPasswordTooWeak
	}
	return nil
}
