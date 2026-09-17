package hrc

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"gorm.io/gorm"
)

var (
	ErrChatPeerInvalid     = errors.New("对方账号不存在或不可聊天")
	ErrChatSelf            = errors.New("不能和自己发起对话")
	ErrChatSessionNotFound = errors.New("会话不存在")
	ErrChatContentEmpty    = errors.New("消息内容不能为空")
	ErrChatContentTooLong  = errors.New("单条消息不能超过 1000 字")
)

// chatMessageMaxRunes 单条消息长度上限
const chatMessageMaxRunes = 1000

// ChatSessionItem 会话列表项（按会话方向返回对端信息）
type ChatSessionItem struct {
	ID           uint64     `json:"id"`
	PeerUID      uint64     `json:"peerUid"`
	PeerName     string     `json:"peerName"`
	PeerLogo     string     `json:"peerLogo"`
	JobsID       uint64     `json:"jobsId"`
	JobsName     string     `json:"jobsName"`
	LastContent  string     `json:"lastContent"`
	LastTime     *time.Time `json:"lastTime"`
	Unread       int        `json:"unread"`
	UpdateTime   time.Time  `json:"updateTime"`
	CompanyHRUID uint64     `json:"companyHrUid"` // 企业侧接待账号（个人端可用以区分同企业不同 HR）
}

// ChatMessageItem 消息项（Mine 便于前端左右气泡）
type ChatMessageItem struct {
	ID        uint64    `json:"id"`
	SessionID uint64    `json:"sessionId"`
	FromUID   uint64    `json:"fromUid"`
	ToUID     uint64    `json:"toUid"`
	Mine      bool      `json:"mine"`
	Content   string    `json:"content"`
	IsRead    int8      `json:"isRead"`
	AddTime   time.Time `json:"addtime"`
}

// ChatService 在线对话服务
// 会话归属：个人 + 企业 + 企业侧接待账号（realUID）。
// 企业主账号与各 HR 子账号各自独立会话；uid 为企业主账号（数据归属），realUID 为真实登录账号。
type ChatService struct{}

// OpenSession 打开/创建会话（幂等：同一「个人+企业+接待账号」仅一条；复聊时恢复己方删除标记）
func (s *ChatService) OpenSession(ctx context.Context, uid uint64, utype int8, realUID uint64, peerUID uint64, jobsID uint64, jobsName string) (*hrcModel.ImSession, error) {
	if peerUID == 0 || peerUID == uid {
		return nil, ErrChatSelf
	}
	db := global.GVA_DB.WithContext(ctx)
	var peer hrcModel.Members
	if err := db.Where("uid = ? AND deleted_at IS NULL AND status = 1", peerUID).First(&peer).Error; err != nil {
		return nil, ErrChatPeerInvalid
	}
	// 必须为个人 ↔ 企业，同类型账号不开放对话
	if peer.Utype == utype {
		return nil, ErrChatPeerInvalid
	}
	personalUID, companyUID, companyHRUID := uid, peerUID, peerUID
	if utype == 2 {
		personalUID, companyUID, companyHRUID = peerUID, uid, realUID
	} else if peer.CompanyUID > 0 {
		// 个人向 HR 子账号发起：数据归属企业主账号，接待账号为该 HR
		companyUID = peer.CompanyUID
	}

	now := hrcModel.Now()
	var session hrcModel.ImSession
	err := db.Where("personal_uid = ? AND company_uid = ? AND company_hr_uid = ?", personalUID, companyUID, companyHRUID).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		session = hrcModel.ImSession{
			PersonalUID:  personalUID,
			CompanyUID:   companyUID,
			CompanyHRUID: companyHRUID,
			JobsID:       jobsID,
			JobsName:     strings.TrimSpace(jobsName),
			PersonalName: s.personalName(db, personalUID),
			CompanyName:  s.companyName(db, companyUID),
			AddTime:      now,
			UpdateTime:   now,
		}
		if err := db.Create(&session).Error; err != nil {
			return nil, err
		}
		return &session, nil
	}
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{"update_time": now}
	if utype == 1 {
		updates["personal_deleted"] = 0
	} else {
		updates["company_deleted"] = 0
	}
	if jobsID > 0 && session.JobsID == 0 {
		updates["jobs_id"] = jobsID
		updates["jobs_name"] = strings.TrimSpace(jobsName)
	}
	if err := db.Model(&hrcModel.ImSession{}).Where("id = ?", session.ID).Updates(updates).Error; err != nil {
		return nil, err
	}
	if jobsID > 0 && session.JobsID == 0 {
		session.JobsID = jobsID
		session.JobsName = strings.TrimSpace(jobsName)
	}
	return &session, nil
}

// ListSessions 会话列表（己方未删；按最后消息时间倒序）
// 企业侧仅返回当前登录账号（realUID）的会话，实现每个 HR 相互独立。
func (s *ChatService) ListSessions(ctx context.Context, uid uint64, utype int8, realUID uint64, info request.PageInfo) ([]ChatSessionItem, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.ImSession{})
	if utype == 1 {
		db = db.Where("personal_uid = ? AND personal_deleted = 0", uid)
	} else {
		db = db.Where("company_uid = ? AND company_hr_uid = ? AND company_deleted = 0", uid, realUID)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	var list []hrcModel.ImSession
	if err := db.Order("last_time desc, update_time desc, id desc").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	companyIDs := make([]uint64, 0, len(list))
	personalIDs := make([]uint64, 0, len(list))
	for _, it := range list {
		if utype == 1 {
			companyIDs = append(companyIDs, it.CompanyUID)
		} else {
			personalIDs = append(personalIDs, it.PersonalUID)
		}
	}
	logoMap := s.companyLogos(ctx, companyIDs)
	avatarMap := s.memberAvatars(ctx, personalIDs)
	// 名称回查用独立 session，避免复用上面带过滤条件的 builder（GORM 条件会叠加）
	lookupDB := global.GVA_DB.WithContext(ctx)

	items := make([]ChatSessionItem, 0, len(list))
	for _, it := range list {
		item := ChatSessionItem{
			ID: it.ID, JobsID: it.JobsID, JobsName: it.JobsName, LastContent: it.LastContent,
			LastTime: it.LastTime, UpdateTime: it.UpdateTime, CompanyHRUID: it.CompanyHRUID,
		}
		if utype == 1 {
			// 个人端：对端是企业侧接待账号（主账号或某 HR）
			item.PeerUID = it.CompanyHRUID
			item.PeerName = it.CompanyName
			if it.CompanyHRUID != it.CompanyUID {
				if hrName := s.memberDisplayName(lookupDB, it.CompanyHRUID); hrName != "" {
					item.PeerName = it.CompanyName + " · " + hrName
				}
			}
			item.PeerLogo = logoMap[it.CompanyUID]
			item.Unread = it.PersonalUnread
		} else {
			// 企业端：对端是求职者
			item.PeerUID = it.PersonalUID
			item.PeerName = it.PersonalName
			item.PeerLogo = avatarMap[it.PersonalUID]
			item.Unread = it.CompanyUnread
		}
		if item.PeerName == "" {
			item.PeerName = s.peerName(lookupDB, item.PeerUID, utype)
		}
		items = append(items, item)
	}
	return items, total, nil
}

// Messages 会话消息（倒序取最新一页后反转，前端按时间正序渲染）
func (s *ChatService) Messages(ctx context.Context, uid uint64, utype int8, realUID uint64, sessionID uint64, info request.PageInfo) ([]ChatMessageItem, int64, error) {
	db := global.GVA_DB.WithContext(ctx)
	if _, err := s.ownedSession(db, uid, utype, realUID, sessionID); err != nil {
		return nil, 0, err
	}
	var total int64
	if err := db.Model(&hrcModel.ImMessage{}).Where("session_id = ?", sessionID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	var msgs []hrcModel.ImMessage
	if err := db.Where("session_id = ?", sessionID).Order("id desc").Limit(limit).Offset(offset).Find(&msgs).Error; err != nil {
		return nil, 0, err
	}
	items := make([]ChatMessageItem, 0, len(msgs))
	for i := len(msgs) - 1; i >= 0; i-- {
		m := msgs[i]
		items = append(items, ChatMessageItem{
			ID: m.ID, SessionID: m.SessionID, FromUID: m.FromUID, ToUID: m.ToUID,
			Mine: m.FromUID == realUID, Content: m.Content, IsRead: m.IsRead, AddTime: m.AddTime,
		})
	}
	return items, total, nil
}

// SendMessage 发送消息（同一事务：写消息 + 更新会话预览/未读；返回对端真实 uid 供 WebSocket 推送）
func (s *ChatService) SendMessage(ctx context.Context, uid uint64, utype int8, realUID uint64, sessionID uint64, content string) (*ChatMessageItem, uint64, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, 0, ErrChatContentEmpty
	}
	if utf8.RuneCountInString(content) > chatMessageMaxRunes {
		return nil, 0, ErrChatContentTooLong
	}

	var msg *hrcModel.ImMessage
	var peerUID uint64
	now := hrcModel.Now()
	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		session, err := s.ownedSession(tx, uid, utype, realUID, sessionID)
		if err != nil {
			return err
		}
		peerUID = session.CompanyHRUID
		if utype == 2 {
			peerUID = session.PersonalUID
		}
		msg = &hrcModel.ImMessage{
			SessionID: sessionID, FromUID: realUID, ToUID: peerUID,
			Content: content, IsRead: 0, AddTime: now,
		}
		if err := tx.Create(msg).Error; err != nil {
			return err
		}
		updates := map[string]interface{}{
			"last_content": content,
			"last_time":    now,
			"update_time":  now,
		}
		if utype == 1 {
			updates["company_unread"] = gorm.Expr("company_unread + 1")
		} else {
			updates["personal_unread"] = gorm.Expr("personal_unread + 1")
		}
		return tx.Model(&hrcModel.ImSession{}).Where("id = ?", sessionID).Updates(updates).Error
	})
	if err != nil {
		return nil, 0, err
	}
	return &ChatMessageItem{
		ID: msg.ID, SessionID: msg.SessionID, FromUID: msg.FromUID, ToUID: msg.ToUID,
		Mine: true, Content: msg.Content, IsRead: msg.IsRead, AddTime: msg.AddTime,
	}, peerUID, nil
}

// MarkRead 标记会话内发给自己的消息已读，并清零己方未读
func (s *ChatService) MarkRead(ctx context.Context, uid uint64, utype int8, realUID uint64, sessionID uint64) error {
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := s.ownedSession(tx, uid, utype, realUID, sessionID); err != nil {
			return err
		}
		if err := tx.Model(&hrcModel.ImMessage{}).
			Where("session_id = ? AND to_uid = ? AND is_read = 0", sessionID, realUID).
			Update("is_read", 1).Error; err != nil {
			return err
		}
		column := "personal_unread"
		if utype == 2 {
			column = "company_unread"
		}
		return tx.Model(&hrcModel.ImSession{}).Where("id = ?", sessionID).
			Update(column, 0).Error
	})
}

// DeleteSession 删除会话（仅置己方删除标记，对方仍可见）
func (s *ChatService) DeleteSession(ctx context.Context, uid uint64, utype int8, realUID uint64, sessionID uint64) error {
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := s.ownedSession(tx, uid, utype, realUID, sessionID); err != nil {
			return err
		}
		column := "personal_deleted"
		if utype == 2 {
			column = "company_deleted"
		}
		return tx.Model(&hrcModel.ImSession{}).Where("id = ?", sessionID).Update(column, 1).Error
	})
}

// UnreadTotal 己方未读消息总数（企业侧按当前登录账号 realUID 统计）
func (s *ChatService) UnreadTotal(ctx context.Context, uid uint64, utype int8, realUID uint64) (int64, error) {
	column := "personal_unread"
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.ImSession{}).Where("personal_uid = ? AND personal_deleted = 0", uid)
	if utype == 2 {
		column = "company_unread"
		db = global.GVA_DB.WithContext(ctx).Model(&hrcModel.ImSession{}).
			Where("company_uid = ? AND company_hr_uid = ? AND company_deleted = 0", uid, realUID)
	}
	var total int64
	err := db.Select("COALESCE(SUM(" + column + "), 0)").Scan(&total).Error
	return total, err
}

// ownedSession 归属校验：个人校验 personal_uid；企业校验 company_uid + 接待账号 company_hr_uid
func (s *ChatService) ownedSession(db *gorm.DB, uid uint64, utype int8, realUID uint64, sessionID uint64) (*hrcModel.ImSession, error) {
	var session hrcModel.ImSession
	var err error
	if utype == 1 {
		err = db.Where("id = ? AND personal_uid = ?", sessionID, uid).First(&session).Error
	} else {
		err = db.Where("id = ? AND company_uid = ? AND company_hr_uid = ?", sessionID, uid, realUID).First(&session).Error
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChatSessionNotFound
		}
		return nil, err
	}
	return &session, nil
}

func (s *ChatService) peerName(db *gorm.DB, peerUID uint64, selfUtype int8) string {
	if selfUtype == 1 {
		return s.companyName(db, peerUID)
	}
	return s.personalName(db, peerUID)
}

// memberDisplayName 真实账号显示名（HR 被移除后返回空，避免出现「求职者」等兜底文案）
func (s *ChatService) memberDisplayName(db *gorm.DB, uid uint64) string {
	var info hrcModel.MembersInfo
	if err := db.Where("uid = ?", uid).First(&info).Error; err == nil && strings.TrimSpace(info.RealName) != "" {
		return info.RealName
	}
	var m hrcModel.Members
	if err := db.Where("uid = ?", uid).First(&m).Error; err == nil {
		if strings.TrimSpace(m.Username) != "" {
			return m.Username
		}
		return strings.TrimSpace(m.Mobile)
	}
	return ""
}

func (s *ChatService) personalName(db *gorm.DB, uid uint64) string {
	var info hrcModel.MembersInfo
	if err := db.Where("uid = ?", uid).First(&info).Error; err == nil && strings.TrimSpace(info.RealName) != "" {
		return info.RealName
	}
	var m hrcModel.Members
	if err := db.Where("uid = ?", uid).First(&m).Error; err == nil {
		if strings.TrimSpace(m.Username) != "" {
			return m.Username
		}
		return m.Mobile
	}
	return "求职者"
}

func (s *ChatService) companyName(db *gorm.DB, uid uint64) string {
	var p hrcModel.CompanyProfile
	if err := db.Where("uid = ?", uid).First(&p).Error; err == nil && p.CompanyName != nil && strings.TrimSpace(*p.CompanyName) != "" {
		return *p.CompanyName
	}
	return "企业"
}

func (s *ChatService) companyLogos(ctx context.Context, uids []uint64) map[uint64]string {
	out := map[uint64]string{}
	if len(uids) == 0 {
		return out
	}
	var profiles []hrcModel.CompanyProfile
	if err := global.GVA_DB.WithContext(ctx).Select("uid, logo").Where("uid IN ?", uids).Find(&profiles).Error; err == nil {
		for _, p := range profiles {
			out[p.UID] = p.Logo
		}
	}
	return out
}

func (s *ChatService) memberAvatars(ctx context.Context, uids []uint64) map[uint64]string {
	out := map[uint64]string{}
	if len(uids) == 0 {
		return out
	}
	var members []hrcModel.Members
	if err := global.GVA_DB.WithContext(ctx).Select("uid, avatars").Where("uid IN ?", uids).Find(&members).Error; err == nil {
		for _, m := range members {
			out[m.UID] = m.Avatars
		}
	}
	return out
}
