package hrc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	orderExpireKeyPrefix = "hrc:order:expire:"
	orderExpireDuration  = 30 * time.Minute
)

var (
	ErrOrderNotFound   = errors.New("订单不存在")
	ErrOrderPaid       = errors.New("订单已支付")
	ErrOrderClosed     = errors.New("订单已关闭")
	ErrPaymentConflict = errors.New("支付交易号已被其他订单使用")
)

type OrderService struct{}

func (s *OrderService) CreateSetmealOrder(ctx context.Context, uid, setmealID uint64) (*hrcModel.Order, error) {
	var plan hrcModel.Setmeal
	if err := global.GVA_DB.WithContext(ctx).Where("id = ? AND display = ?", setmealID, true).First(&plan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSetmealNotFound
		}
		return nil, err
	}
	now := time.Now()
	order := &hrcModel.Order{
		OID:         newOrderID(),
		UID:         uid,
		OrderType:   1,
		SetmealID:   plan.ID,
		SetmealName: plan.Name,
		Amount:      plan.Price,
		IsPaid:      1,
		CreatedAt:   now,
	}
	if err := global.GVA_DB.WithContext(ctx).Create(order).Error; err != nil {
		return nil, err
	}
	// Redis 存储订单过期时间，供前端倒计时使用
	expireAt := now.Add(orderExpireDuration)
	if global.GVA_REDIS != nil {
		_ = global.GVA_REDIS.Set(ctx, orderExpireKeyPrefix+fmt.Sprintf("%d", order.ID), expireAt.Unix(), orderExpireDuration).Err()
	}
	return order, nil
}

// OrderWithExpire 扩展订单信息，包含 Redis 中的过期时间
type OrderWithExpire struct {
	hrcModel.Order
	ExpireAt int64 `json:"expireAt"` // 过期时间戳（unix），0 表示无倒计时
}

func (s *OrderService) ListMine(ctx context.Context, uid uint64, page request.PageInfo) ([]OrderWithExpire, int64, error) {
	limit, offset := page.LimitOffset()
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.Order{}).Where("uid = ?", uid)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []hrcModel.Order
	if err := db.Order("id desc").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	// 从 Redis 批量获取过期时间
	result := make([]OrderWithExpire, len(list))
	for i, order := range list {
		result[i].Order = order
		if order.IsPaid == 1 && global.GVA_REDIS != nil {
			val, err := global.GVA_REDIS.Get(ctx, orderExpireKeyPrefix+fmt.Sprintf("%d", order.ID)).Result()
			if err == nil {
				if v, e := strconv.ParseInt(val, 10, 64); e == nil {
					result[i].ExpireAt = v
				}
			}
		}
	}
	return result, total, nil
}

func (s *OrderService) AdminList(ctx context.Context, page request.PageInfo, status int8) ([]hrcModel.Order, int64, error) {
	limit, offset := page.LimitOffset()
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.Order{})
	if status != 0 {
		db = db.Where("is_paid = ?", status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []hrcModel.Order
	err := db.Order("id desc").Limit(limit).Offset(offset).Find(&list).Error
	return list, total, err
}

// ConfirmPaid is the manual-payment handoff. The order transition and
// entitlement write happen in one transaction, so retries cannot grant twice.
func (s *OrderService) ConfirmPaid(ctx context.Context, id uint64, payAmount int64, payment string) error {
	if payAmount < 0 {
		return ErrSetmealInvalid
	}
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order hrcModel.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&order, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrOrderNotFound
			}
			return err
		}
		if order.IsPaid == 2 {
			return ErrOrderPaid
		}
		if order.IsPaid == 3 {
			return ErrOrderClosed
		}
		if payAmount == 0 {
			payAmount = order.Amount
		}
		if payAmount != order.Amount {
			return errors.New("实收金额与订单金额不一致")
		}
		now := time.Now()
		if err := tx.Model(&hrcModel.Order{}).Where("id = ? AND is_paid = ?", order.ID, 1).Updates(map[string]interface{}{
			"is_paid": 2, "pay_amount": payAmount, "payment": payment, "paid_at": now,
		}).Error; err != nil {
			return err
		}
		// 支付成功，删除 Redis 过期时间
		if global.GVA_REDIS != nil {
			_ = global.GVA_REDIS.Del(ctx, orderExpireKeyPrefix+fmt.Sprintf("%d", order.ID)).Err()
		}
		return grantOrderEntitlement(tx, &order, now)
	})
}

// ConfirmGatewayPaid transitions an order after a payment provider has verified
// its callback. Replayed callbacks for the same gateway transaction are safe.
func (s *OrderService) ConfirmGatewayPaid(ctx context.Context, id uint64, payAmount int64, provider, transactionID string) error {
	if payAmount <= 0 || provider == "" || transactionID == "" {
		return ErrSetmealInvalid
	}
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order hrcModel.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&order, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrOrderNotFound
			}
			return err
		}
		if order.IsPaid == 2 {
			if order.Payment == provider && order.TransactionID != nil && *order.TransactionID == transactionID {
				return nil
			}
			return ErrOrderPaid
		}
		if order.IsPaid == 3 {
			return ErrOrderClosed
		}
		if payAmount != order.Amount {
			return errors.New("实收金额与订单金额不一致")
		}

		var duplicate hrcModel.Order
		err := tx.Where("transaction_id = ? AND id <> ?", transactionID, order.ID).First(&duplicate).Error
		if err == nil {
			return ErrPaymentConflict
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		now := time.Now()
		if err := tx.Model(&hrcModel.Order{}).Where("id = ? AND is_paid = ?", order.ID, 1).Updates(map[string]interface{}{
			"is_paid": 2, "pay_amount": payAmount, "payment": provider, "transaction_id": transactionID, "paid_at": now,
		}).Error; err != nil {
			return err
		}
		// 支付成功，删除 Redis 过期时间
		if global.GVA_REDIS != nil {
			_ = global.GVA_REDIS.Del(ctx, orderExpireKeyPrefix+fmt.Sprintf("%d", order.ID)).Err()
		}
		return grantOrderEntitlement(tx, &order, now)
	})
}

func grantOrderEntitlement(tx *gorm.DB, order *hrcModel.Order, now time.Time) error {
	var plan hrcModel.Setmeal
	if err := tx.First(&plan, order.SetmealID).Error; err != nil {
		return ErrSetmealNotFound
	}
	var current hrcModel.MembersSetmeal
	err := tx.Where("uid = ?", order.UID).First(&current).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	base := now
	if err == nil && current.ExpireAt.After(now) {
		base = current.ExpireAt
	}
	entitlement := hrcModel.MembersSetmeal{
		UID:                  order.UID,
		SetmealID:            plan.ID,
		SetmealName:          plan.Name,
		ExpireAt:             base.AddDate(0, 0, plan.DurationDays),
		JobsMeanwhile:        plan.JobsMeanwhile,
		ResumeDownloadsTotal: plan.ResumeDownloads,
		HomePushSlots:        plan.HomePushSlots,
		HomeAdSlots:          plan.HomeAdSlots,
		EnableVideo:          plan.EnableVideo,
		EnableJobfair:        plan.EnableJobfair,
		UpdatedAt:            now,
	}
	if err == nil {
		entitlement.ID = current.ID
		entitlement.ResumeDownloadsUsed = 0
		if err := tx.Save(&entitlement).Error; err != nil {
			return err
		}
	} else if err := tx.Create(&entitlement).Error; err != nil {
		return err
	}
	return tx.Model(&hrcModel.CompanyProfile{}).Where("uid = ?", order.UID).Updates(map[string]interface{}{
		"setmeal_id": plan.ID, "setmeal_name": plan.Name,
	}).Error
}

func (s *OrderService) Close(ctx context.Context, uid, id uint64) error {
	result := global.GVA_DB.WithContext(ctx).Model(&hrcModel.Order{}).Where("id = ? AND uid = ? AND is_paid = ?", id, uid, 1).Updates(map[string]interface{}{
		"is_paid": 3, "closed_at": time.Now(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrOrderClosed
	}
	// 删除 Redis 中的过期时间
	if global.GVA_REDIS != nil {
		_ = global.GVA_REDIS.Del(ctx, orderExpireKeyPrefix+fmt.Sprintf("%d", id)).Err()
	}
	return nil
}

func newOrderID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("MSTN%s%s", time.Now().Format("20060102150405"), hex.EncodeToString(b))
}
