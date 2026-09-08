package hrc

import (
	"context"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/stretchr/testify/require"
)

func talentDB(t *testing.T) {
	testutil.NewMemoryDB(t,
		&hrcModel.Resume{},
		&hrcModel.ResumeEducation{},
		&hrcModel.ResumeWork{},
		&hrcModel.ResumeLanguage{},
		&hrcModel.ResumeTraining{},
		&hrcModel.ResumeCredent{},
		&hrcModel.ResumeProject{},
		&hrcModel.MembersSetmeal{},
		&hrcModel.ResumeDownload{},
		&hrcModel.CompanyFavorite{},
		&hrcModel.Pms{},
		&hrcModel.MembersMsgtip{},
	)
}

func TestTalentSearchMasksAnonymousResume(t *testing.T) {
	talentDB(t)
	require.NoError(t, global.GVA_DB.Create(&hrcModel.Resume{
		UID: 1, Display: 1, Audit: 1, DisplayName: 2, FullName: "张三", Title: "资深后端工程师",
		Telephone: "13800138000", Email: "zhangsan@example.com", Photo: 1, PhotoAudit: 1, PhotoDisplay: 1,
		PhotoImg: "https://example.com/photo.png", KeyFull: "后端 Go", Talent: 1,
	}).Error)
	require.NoError(t, global.GVA_DB.Create(&hrcModel.Resume{
		UID: 2, Display: 2, Audit: 1, FullName: "隐藏简历", KeyFull: "后端",
	}).Error)

	svc := &TalentService{}
	list, total, err := svc.Search(context.Background(), request.PageInfo{Page: 1, PageSize: 10}, TalentSearch{Keyword: "后端"})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, list, 1)
	require.Equal(t, "求职者", list[0].FullName)
	require.Empty(t, list[0].PhotoImg)

	detail, err := svc.PublicDetail(context.Background(), list[0].ID)
	require.NoError(t, err)
	require.Equal(t, "求职者", detail.Resume.FullName)
	require.Empty(t, detail.Resume.PhotoImg)
}

func TestTalentUnlockConsumesQuotaAndNotifiesOnce(t *testing.T) {
	talentDB(t)
	resume := hrcModel.Resume{
		UID: 10, Display: 1, Audit: 1, FullName: "李四", Title: "Go 工程师",
		Telephone: "13900139000", Email: "lisi@example.com",
	}
	require.NoError(t, global.GVA_DB.Create(&resume).Error)
	require.NoError(t, global.GVA_DB.Create(&hrcModel.MembersSetmeal{
		UID: 20, ExpireAt: hrcModel.Now() + 3600, ResumeDownloadsTotal: 1,
	}).Error)

	svc := &TalentService{}
	detail, newlyUnlocked, err := svc.Unlock(context.Background(), 20, resume.ID)
	require.NoError(t, err)
	require.True(t, newlyUnlocked)
	require.Equal(t, resume.Telephone, detail.Resume.Telephone)

	_, newlyUnlocked, err = svc.Unlock(context.Background(), 20, resume.ID)
	require.NoError(t, err)
	require.False(t, newlyUnlocked)

	var entitlement hrcModel.MembersSetmeal
	require.NoError(t, global.GVA_DB.Where("uid = ?", 20).First(&entitlement).Error)
	require.Equal(t, 1, entitlement.ResumeDownloadsUsed)
	var downloadCount int64
	require.NoError(t, global.GVA_DB.Model(&hrcModel.ResumeDownload{}).Count(&downloadCount).Error)
	require.Equal(t, int64(1), downloadCount)
	var messageCount int64
	require.NoError(t, global.GVA_DB.Model(&hrcModel.Pms{}).Where("msgtouid = ?", 10).Count(&messageCount).Error)
	require.Equal(t, int64(1), messageCount)

	_, err = svc.GetUnlocked(context.Background(), 21, resume.ID)
	require.ErrorIs(t, err, ErrTalentNotUnlocked)
}

func TestTalentFavoritesAndFollowUpOwnership(t *testing.T) {
	talentDB(t)
	resume := hrcModel.Resume{UID: 10, Display: 1, Audit: 1, FullName: "王五", Title: "前端工程师"}
	require.NoError(t, global.GVA_DB.Create(&resume).Error)
	require.NoError(t, global.GVA_DB.Create(&hrcModel.ResumeDownload{CompanyUID: 20, ResumeID: resume.ID}).Error)

	svc := &TalentService{}
	require.NoError(t, svc.SetFollowUp(context.Background(), 20, resume.ID, 3))
	require.ErrorIs(t, svc.SetFollowUp(context.Background(), 21, resume.ID, 1), ErrTalentNotUnlocked)
	require.ErrorIs(t, svc.SetFollowUp(context.Background(), 20, resume.ID, 5), ErrTalentFollowUp)

	favorite, err := svc.Favorite(context.Background(), 20, resume.ID)
	require.NoError(t, err)
	require.ErrorIs(t, func() error {
		_, err := svc.Favorite(context.Background(), 20, resume.ID)
		return err
	}(), ErrTalentFavorite)

	list, total, err := svc.ListFavorites(context.Background(), 20, request.PageInfo{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, list, 1)
	require.Equal(t, resume.ID, list[0].Resume.ID)

	require.ErrorIs(t, svc.Unfavorite(context.Background(), 21, favorite.ID), ErrTalentFavoriteGone)
	require.NoError(t, svc.Unfavorite(context.Background(), 20, favorite.ID))

	talentList, total, err := svc.ListUnlocked(context.Background(), 20, request.PageInfo{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, talentList, 1)
	require.Equal(t, int8(3), talentList[0].Download.FollowUp)
}
