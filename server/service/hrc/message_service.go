package hrc

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const pmsMessageTipType = "pms"

var ErrMessageIDsInvalid = errors.New("请选择 1-100 条消息")

type SystemNotice struct {
	FromUID uint64
	ToUID   uint64
	Title   string
	Message string
	Type    string
	Link    string
}

// MessageService 提供会员站内信读写和通用系统通知入口。
type MessageService struct{}

// SendSystemNotice writes a notification with the caller's transaction so the
// notification and the triggering business record commit or roll back together.
func (s *MessageService) SendSystemNotice(tx *gorm.DB, notice SystemNotice) error {
	notice.Title = strings.TrimSpace(notice.Title)
	notice.Message = strings.TrimSpace(notice.Message)
	notice.Type = strings.TrimSpace(notice.Type)
	if notice.ToUID == 0 || notice.Title == "" || notice.Message == "" {
		return errors.New("站内信参数不完整")
	}
	if notice.Type == "" {
		notice.Type = pmsMessageTipType
	}
	if len(notice.Title) > 120 || len(notice.Type) > 30 || len(notice.Link) > 255 {
		return errors.New("站内信字段长度超出限制")
	}

	now := time.Now()
	if err := tx.Create(&hrcModel.Pms{
		MsgFrom:  notice.FromUID,
		MsgTouid: notice.ToUID,
		Title:    notice.Title,
		Message:  notice.Message,
		Type:     notice.Type,
		Link:     notice.Link,
		MsgCheck: 0,
		AddTime:  now,
	}).Error; err != nil {
		return err
	}

	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "uid"}, {Name: "type"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"unread":      gorm.Expr("unread + ?", 1),
			"update_time": now,
		}),
	}).Create(&hrcModel.MembersMsgtip{
		UID:        notice.ToUID,
		Type:       pmsMessageTipType,
		Unread:     1,
		UpdateTime: now,
	}).Error
}

func (s *MessageService) List(ctx context.Context, uid uint64, info request.PageInfo) ([]hrcModel.Pms, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.Pms{}).Where("msgtouid = ?", uid)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	var list []hrcModel.Pms
	if err := db.Order("addtime desc, id desc").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *MessageService) UnreadCount(ctx context.Context, uid uint64) (int64, error) {
	var total int64
	err := global.GVA_DB.WithContext(ctx).Model(&hrcModel.Pms{}).
		Where("msgtouid = ? AND msg_check = 0", uid).
		Count(&total).Error
	return total, err
}

func (s *MessageService) MarkRead(ctx context.Context, uid uint64, ids []uint64) error {
	if !validMessageIDs(ids) {
		return ErrMessageIDsInvalid
	}
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&hrcModel.Pms{}).
			Where("msgtouid = ? AND id IN ? AND msg_check = 0", uid, ids).
			Update("msg_check", 1).Error; err != nil {
			return err
		}
		return s.syncUnreadTip(tx, uid)
	})
}

func (s *MessageService) Delete(ctx context.Context, uid uint64, ids []uint64) error {
	if !validMessageIDs(ids) {
		return ErrMessageIDsInvalid
	}
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("msgtouid = ? AND id IN ?", uid, ids).Delete(&hrcModel.Pms{}).Error; err != nil {
			return err
		}
		return s.syncUnreadTip(tx, uid)
	})
}

func (s *MessageService) syncUnreadTip(tx *gorm.DB, uid uint64) error {
	var unread int64
	if err := tx.Model(&hrcModel.Pms{}).Where("msgtouid = ? AND msg_check = 0", uid).Count(&unread).Error; err != nil {
		return err
	}
	now := time.Now()
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "uid"}, {Name: "type"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"unread":      int(unread),
			"update_time": now,
		}),
	}).Create(&hrcModel.MembersMsgtip{
		UID:        uid,
		Type:       pmsMessageTipType,
		Unread:     int(unread),
		UpdateTime: now,
	}).Error
}

func validMessageIDs(ids []uint64) bool {
	return len(ids) > 0 && len(ids) <= 100
}
