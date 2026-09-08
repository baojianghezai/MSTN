package hrc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MemberClaims 会员 JWT Claims（独立于 GVA 后台 admin JWT）
// 设计依据：02_用户账号模块设计 §3.2 / 09_API接口规范 §1.5
type MemberClaims struct {
	UID   uint64 `json:"uid"`   // 会员 uid
	Utype int8   `json:"utype"` // 1=个人 2=企业
	Ver   int64  `json:"ver"`   // 密码版本号（改密/踢出其他端时自增）
	jwt.RegisteredClaims
}

// MemberJWT 会员 JWT 签发器（复用 golang-jwt，独立 secret）
type MemberJWT struct {
	Secret     []byte
	Expires    time.Duration // access token 有效期
	RefreshExp time.Duration // refresh token 有效期
}

var (
	ErrTokenExpired = errors.New("token 已过期")
	ErrTokenInvalid = errors.New("token 无效")
)

// DefaultMemberJWT 从 GVA 配置构造默认会员 JWT（secret 使用 JWT.SigningKey 拼接 "member" 前缀，保证与 admin JWT 隔离）
func DefaultMemberJWT() *MemberJWT {
	accessExp, _ := time.ParseDuration(global.GVA_CONFIG.JWT.ExpiresTime)
	refreshExp, _ := time.ParseDuration(global.GVA_CONFIG.JWT.ExpiresTime) // 一期 access/refresh 同长
	if refreshExp <= 0 {
		refreshExp = 720 * time.Hour
	}
	return &MemberJWT{
		Secret:     []byte("member:" + global.GVA_CONFIG.JWT.SigningKey),
		Expires:    accessExp,
		RefreshExp: refreshExp,
	}
}

// CreateToken 签发 access token（单点登录：同 uid 新 token 顶掉旧 token）
func (m *MemberJWT) CreateToken(uid uint64, utype int8, ver int64) (string, error) {
	claims := MemberClaims{
		UID:   uid,
		Utype: utype,
		Ver:   ver,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.Expires)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "msrc",
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.Secret)
	if err != nil {
		return "", err
	}
	// 单点登录：记录当前有效 token（Redis 可用时）
	if global.GVA_REDIS != nil {
		ctx := context.Background()
		_ = global.GVA_REDIS.Set(ctx, memberSessionKey(uid), token, m.Expires).Err()
	}
	return token, nil
}

// InvalidateSession 使某 uid 的所有会话失效（登出/改密/踢人用）
func InvalidateSession(uid uint64) {
	if global.GVA_REDIS != nil {
		ctx := context.Background()
		_ = global.GVA_REDIS.Del(ctx, memberSessionKey(uid)).Err()
	}
}

func memberSessionKey(uid uint64) string {
	return fmt.Sprintf("hrc:member:session:%d", uid)
}

// ParseToken 解析并校验 token（不校验会话，供内部使用）
func (m *MemberJWT) ParseToken(tokenString string) (*MemberClaims, error) {
	claims := &MemberClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return m.Secret, nil
	})
	if err != nil {
		global.GVA_LOG.Error("member jwt parse failed", zap.Error(err))
		return nil, ErrTokenInvalid
	}
	if !token.Valid {
		return nil, ErrTokenExpired
	}
	return claims, nil
}

// MemberAuth 会员鉴权中间件（挂在 hrc 业务路由组，PublicGroup 下）
// 包含单点登录校验：Redis 会话与 token 不一致 → 视为已被新登录顶掉
func MemberAuth() gin.HandlerFunc {
	j := DefaultMemberJWT()
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"code": 1001, "message": "未登录", "data": nil})
			return
		}
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
		claims, err := j.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"code": 1001, "message": "登录已失效，请重新登录", "data": nil})
			return
		}
		// 单点登录校验：会话记录必须是当前 token（登出/改密/踢人后会话被清，旧 token 一律失效）
		if global.GVA_REDIS != nil {
			ctx := context.Background()
			current, _ := global.GVA_REDIS.Get(ctx, memberSessionKey(claims.UID)).Result()
			if current != token {
				c.AbortWithStatusJSON(401, gin.H{"code": 1001, "message": "账号已在其他设备登录，请重新登录", "data": nil})
				return
			}
		}
		// 加载用户并校验状态：不存在/已注销/已暂停的账号，其 token 一律失效（防伪造 token 与注销后复用）
		var member hrcModel.Members
		err = global.GVA_DB.Where("uid = ? AND deleted_at = 0", claims.UID).First(&member).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.AbortWithStatusJSON(401, gin.H{"code": 1001, "message": "登录已失效，请重新登录", "data": nil})
			return
		}
		if err != nil {
			global.GVA_LOG.Error("member auth load user failed", zap.Uint64("uid", claims.UID), zap.Error(err))
			c.AbortWithStatusJSON(401, gin.H{"code": 1001, "message": "登录已失效，请重新登录", "data": nil})
			return
		}
		if member.Status != 1 {
			c.AbortWithStatusJSON(401, gin.H{"code": 1001, "message": "账号状态异常，请联系管理员", "data": nil})
			return
		}
		c.Set("member_uid", claims.UID)
		c.Set("member_utype", claims.Utype)
		c.Set("member_claims", claims)
		c.Next()
	}
}

// UtypeAuth 角色校验中间件（utype: 1=个人 2=企业）
func UtypeAuth(utype int8) gin.HandlerFunc {
	return func(c *gin.Context) {
		ut, exists := c.Get("member_utype")
		if !exists || ut.(int8) != utype {
			c.AbortWithStatusJSON(403, gin.H{"code": 1001, "message": "无权限访问", "data": nil})
			return
		}
		c.Next()
	}
}

// GetMemberUID 从上下文取会员 uid
func GetMemberUID(c *gin.Context) uint64 {
	v, _ := c.Get("member_uid")
	if v == nil {
		return 0
	}
	return v.(uint64)
}

// GetMemberUtype 从上下文取会员类型
func GetMemberUtype(c *gin.Context) int8 {
	v, _ := c.Get("member_utype")
	if v == nil {
		return 0
	}
	return v.(int8)
}
