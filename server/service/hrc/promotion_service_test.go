package hrc

import (
	"context"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/stretchr/testify/require"
)

func TestPromotionCreateListAndDelete(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.CompanyProfile{}, &hrcModel.JobPromotion{}, &hrcModel.MembersSetmeal{})
	job := hrcModel.Jobs{JobsBase: hrcModel.JobsBase{
		UID: 100, JobsName: "Go 开发", CompanyName: "名硕科技", Display: 1, Audit: 1,
	}}
	require.NoError(t, db.Create(&job).Error)
	require.NoError(t, db.Create(&hrcModel.MembersSetmeal{
		UID: 100, ExpireAt: hrcModel.Now() + 3600, HomePushSlots: 1, HomeAdSlots: 1,
	}).Error)

	svc := &PromotionService{}
	push, err := svc.Create(context.Background(), 100, job.ID, hrcModel.JobPromotionTypePush, PromotionCreative{})
	require.NoError(t, err)
	ad, err := svc.Create(context.Background(), 100, job.ID, hrcModel.JobPromotionTypeAd, PromotionCreative{
		AdTitle:    "加入名硕科技",
		AdSubtitle: "寻找优秀工程师",
		AdImage:    "uploads/ad-banner.png",
	})
	require.NoError(t, err)

	home, err := svc.ListHome(context.Background())
	require.NoError(t, err)
	require.Len(t, home.Push, 1)
	require.Len(t, home.Ads, 1)
	require.Equal(t, job.ID, home.Push[0].ID)
	require.Equal(t, job.ID, home.Ads[0].ID)
	require.Equal(t, "uploads/ad-banner.png", home.Ads[0].AdImage)
	require.Equal(t, "加入名硕科技", home.Ads[0].AdTitle)

	mine, err := svc.ListMine(context.Background(), 100)
	require.NoError(t, err)
	require.Equal(t, 1, mine.HomePushSlots)
	require.Equal(t, int64(1), mine.PushUsed)
	require.Equal(t, int64(1), mine.AdUsed)

	require.NoError(t, svc.Delete(context.Background(), 100, push.ID))
	home, err = svc.ListHome(context.Background())
	require.NoError(t, err)
	require.Empty(t, home.Push)
	require.Len(t, home.Ads, 1)
	require.ErrorIs(t, svc.Delete(context.Background(), 999, ad.ID), ErrPromotionNotFound)
}

func TestPromotionRejectsMissingEntitlementAndInvalidJob(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.CompanyProfile{}, &hrcModel.JobPromotion{}, &hrcModel.MembersSetmeal{})
	job := hrcModel.Jobs{JobsBase: hrcModel.JobsBase{UID: 100, JobsName: "待审职位", Display: 1, Audit: 2}}
	require.NoError(t, db.Create(&job).Error)

	svc := &PromotionService{}
	_, err := svc.Create(context.Background(), 100, job.ID, hrcModel.JobPromotionTypePush, PromotionCreative{})
	require.ErrorIs(t, err, ErrPromotionEntitlement)

	require.NoError(t, db.Create(&hrcModel.MembersSetmeal{
		UID: 100, ExpireAt: hrcModel.Now() + 3600, HomePushSlots: 1,
	}).Error)
	_, err = svc.Create(context.Background(), 100, job.ID, hrcModel.JobPromotionTypePush, PromotionCreative{})
	require.ErrorIs(t, err, ErrPromotionJobInvalid)
	_, err = svc.Create(context.Background(), 100, job.ID, 9, PromotionCreative{})
	require.ErrorIs(t, err, ErrPromotionTypeInvalid)

	_, err = svc.Create(context.Background(), 100, job.ID, hrcModel.JobPromotionTypeAd, PromotionCreative{})
	require.ErrorIs(t, err, ErrPromotionAdImage)
}

func TestPromotionLegacyAdFallsBackToCompanyLogo(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.Jobs{}, &hrcModel.CompanyProfile{}, &hrcModel.JobPromotion{}, &hrcModel.MembersSetmeal{})
	companyName := "名硕科技"
	job := hrcModel.Jobs{JobsBase: hrcModel.JobsBase{
		UID: 200, JobsName: "产品经理", CompanyName: companyName, Display: 1, Audit: 1,
	}}
	require.NoError(t, db.Create(&job).Error)
	require.NoError(t, db.Create(&hrcModel.CompanyProfile{UID: 200, CompanyName: &companyName, Logo: "uploads/company-logo.png"}).Error)
	require.NoError(t, db.Create(&hrcModel.MembersSetmeal{UID: 200, ExpireAt: hrcModel.Now() + 3600, HomeAdSlots: 1}).Error)
	require.NoError(t, db.Create(&hrcModel.JobPromotion{UID: 200, JobID: job.ID, Type: hrcModel.JobPromotionTypeAd, CreatedAt: hrcModel.Now()}).Error)

	home, err := (&PromotionService{}).ListHome(context.Background())
	require.NoError(t, err)
	require.Len(t, home.Ads, 1)
	require.Equal(t, "uploads/company-logo.png", home.Ads[0].AdImage)
	require.Equal(t, "产品经理", home.Ads[0].AdTitle)
	require.Equal(t, companyName, home.Ads[0].AdSubtitle)
}
