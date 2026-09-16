package hrc

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/stretchr/testify/require"
)

func TestCompanyApplyListLookedReply(t *testing.T) {
	db := testutil.NewMemoryDB(t, &hrcModel.PersonalJobsApply{})
	require.NoError(t, db.Create(&hrcModel.PersonalJobsApply{CompanyUID: 100, PersonalUID: 1, ResumeID: 1, ResumeName: "张三", JobsName: "Go 工程师", JobsID: 1, PersonalLook: 1}).Error)
	require.NoError(t, db.Create(&hrcModel.PersonalJobsApply{CompanyUID: 100, PersonalUID: 2, ResumeID: 2, ResumeName: "李四", JobsName: "前端工程师", JobsID: 2, PersonalLook: 2}).Error)
	require.NoError(t, db.Create(&hrcModel.PersonalJobsApply{CompanyUID: 200, PersonalUID: 3, ResumeID: 3, ResumeName: "王五", JobsName: "测试", JobsID: 3, PersonalLook: 1}).Error)

	svc := &CompanyApplyService{}
	info := request.PageInfo{Page: 1, PageSize: 10}

	// 全部（企业 100 有 2 条）
	list, total, err := svc.List(context.Background(), 100, info, 0, 0)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, list, 2)

	// 未读筛选
	list, total, err = svc.List(context.Background(), 100, info, 0, 1)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)

	// 单职位筛选
	list, total, err = svc.List(context.Background(), 100, info, 1, 0)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, "Go 工程师", list[0].JobsName)

	// 标记已看
	require.NoError(t, svc.Looked(context.Background(), 100, list[0].DID))
	var after hrcModel.PersonalJobsApply
	require.NoError(t, db.First(&after, list[0].DID).Error)
	require.Equal(t, int8(2), after.PersonalLook)

	// 回复状态
	require.NoError(t, svc.Reply(context.Background(), 100, list[0].DID, 1))
	require.NoError(t, db.First(&after, list[0].DID).Error)
	require.Equal(t, int8(1), after.IsReply)
	require.NotZero(t, after.ReplyTime)

	// 非法回复状态
	require.ErrorIs(t, svc.Reply(context.Background(), 100, list[0].DID, 9), ErrReplyStatusInvalid)

	// 非归属企业 → 不存在
	require.ErrorIs(t, svc.Looked(context.Background(), 999, list[0].DID), ErrApplyNotFound)
}

func TestCompanyApplyDownloadResumeConsumesQuotaOnce(t *testing.T) {
	// 强制字体缺失 → 回退 HTML 快照（PDF 生成另有专测）
	t.Setenv("RESUME_PDF_FONT", filepath.Join(t.TempDir(), "missing.ttf"))
	db := testutil.NewMemoryDB(t,
		&hrcModel.PersonalJobsApply{},
		&hrcModel.Resume{},
		&hrcModel.ResumeEducation{},
		&hrcModel.ResumeWork{},
		&hrcModel.ResumeLanguage{},
		&hrcModel.ResumeTraining{},
		&hrcModel.ResumeCredent{},
		&hrcModel.ResumeProject{},
		&hrcModel.ResumeSkill{},
		&hrcModel.ResumePortfolio{},
		&hrcModel.ResumeStudentLeader{},
		&hrcModel.MembersSetmeal{},
		&hrcModel.ResumeDownload{},
		&hrcModel.Pms{},
		&hrcModel.MembersMsgtip{},
	)
	resume := hrcModel.Resume{
		UID:       1,
		FullName:  "张三",
		Title:     "后端工程师",
		Telephone: "13800138000",
		Email:     "zhangsan@example.com",
		Specialty: "擅长 <Go>",
	}
	require.NoError(t, db.Create(&resume).Error)
	require.NoError(t, db.Create(&hrcModel.ResumeWork{
		PID: resume.ID, UID: 1, CompanyName: "示例公司", Jobs: "开发工程师",
	}).Error)
	apply := hrcModel.PersonalJobsApply{
		ResumeID: resume.ID, PersonalUID: 1, CompanyUID: 100, ResumeName: "张三", JobsName: "Go 工程师",
	}
	require.NoError(t, db.Create(&apply).Error)
	require.NoError(t, db.Create(&hrcModel.MembersSetmeal{
		UID: 100, ExpireAt: hrcModel.Now().Add(3600 * time.Second), ResumeDownloadsTotal: 1,
	}).Error)

	svc := &CompanyApplyService{}
	file, err := svc.DownloadResume(context.Background(), 100, apply.DID)
	require.NoError(t, err)
	require.Equal(t, "resume-1.html", file.Filename)
	require.Contains(t, string(file.Content), "13800138000")
	require.Contains(t, string(file.Content), "&lt;Go&gt;")

	var entitlement hrcModel.MembersSetmeal
	require.NoError(t, db.Where("uid = ?", 100).First(&entitlement).Error)
	require.Equal(t, 1, entitlement.ResumeDownloadsUsed)
	var count int64
	require.NoError(t, db.Model(&hrcModel.ResumeDownload{}).Count(&count).Error)
	require.Equal(t, int64(1), count)
	require.NoError(t, db.Model(&hrcModel.Pms{}).Where("msgtouid = ?", 1).Count(&count).Error)
	require.Equal(t, int64(1), count)

	_, err = svc.DownloadResume(context.Background(), 100, apply.DID)
	require.NoError(t, err)
	require.NoError(t, db.Where("uid = ?", 100).First(&entitlement).Error)
	require.Equal(t, 1, entitlement.ResumeDownloadsUsed)
	require.NoError(t, db.First(&apply, apply.DID).Error)
	require.Equal(t, int8(2), apply.PersonalLook)
	require.NoError(t, db.Model(&hrcModel.Pms{}).Where("msgtouid = ?", 1).Count(&count).Error)
	require.Equal(t, int64(1), count)

	_, err = svc.DownloadResume(context.Background(), 999, apply.DID)
	require.ErrorIs(t, err, ErrApplyNotFound)
}

// 附件简历（PDF）优先下发：企业下载拿到 PDF 内容与文件名，而非 HTML 快照
func TestCompanyApplyDownloadResumePrefersAttachedPDF(t *testing.T) {
	storeDir := t.TempDir()
	prevStore := global.GVA_CONFIG.Local.StorePath
	global.GVA_CONFIG.Local.StorePath = storeDir
	t.Cleanup(func() { global.GVA_CONFIG.Local.StorePath = prevStore })

	pdfContent := []byte("%PDF-1.4 candidate resume")
	require.NoError(t, os.WriteFile(filepath.Join(storeDir, "candidate.pdf"), pdfContent, 0o644))

	db := testutil.NewMemoryDB(t,
		&hrcModel.PersonalJobsApply{},
		&hrcModel.Resume{},
		&hrcModel.ResumeEducation{},
		&hrcModel.ResumeWork{},
		&hrcModel.ResumeLanguage{},
		&hrcModel.ResumeTraining{},
		&hrcModel.ResumeCredent{},
		&hrcModel.ResumeProject{},
		&hrcModel.ResumeSkill{},
		&hrcModel.ResumePortfolio{},
		&hrcModel.ResumeStudentLeader{},
		&hrcModel.MembersSetmeal{},
		&hrcModel.ResumeDownload{},
		&hrcModel.Pms{},
		&hrcModel.MembersMsgtip{},
	)
	resume := hrcModel.Resume{UID: 1, FullName: "张三", WordResume: "uploads/file/candidate.pdf", WordResumeTitle: "张三的简历.pdf"}
	require.NoError(t, db.Create(&resume).Error)
	apply := hrcModel.PersonalJobsApply{ResumeID: resume.ID, PersonalUID: 1, CompanyUID: 100, ResumeName: "张三"}
	require.NoError(t, db.Create(&apply).Error)
	require.NoError(t, db.Create(&hrcModel.MembersSetmeal{
		UID: 100, ExpireAt: hrcModel.Now().Add(3600 * time.Second), ResumeDownloadsTotal: 1,
	}).Error)

	svc := &CompanyApplyService{}
	file, err := svc.DownloadResume(context.Background(), 100, apply.DID)
	require.NoError(t, err)
	require.Equal(t, "张三的简历.pdf", file.Filename)
	require.Equal(t, "application/pdf", file.ContentType)
	require.Equal(t, pdfContent, file.Content)
}

// 在线简历按所选模板渲染 PDF（需系统可用中文 TTF，否则跳过）
func TestRenderResumePDF(t *testing.T) {
	if _, err := FindResumePDFFont(); err != nil {
		t.Skip("未找到可用于 PDF 的中文 TTF 字体，跳过：", err)
	}
	resume := &hrcModel.Resume{
		FullName:      "张三",
		EducationCN:   "本科",
		IntentionJobs: "Go 开发",
		Telephone:     "13800138000",
		Email:         "zhangsan@example.com",
		Specialty:     "擅长后端开发与系统设计",
		WageMin:       8000,
		WageMax:       15000,
	}
	subs := &ResumeSubTables{
		Work: []hrcModel.ResumeWork{{CompanyName: "某某科技", Jobs: "Go 工程师", Achievements: "负责简历模块"}},
	}
	for _, tpl := range []int8{1, 2, 3} {
		out, err := RenderResumePDF(resume, subs, tpl)
		require.NoError(t, err)
		require.True(t, bytes.HasPrefix(out, []byte("%PDF-")), "模板 %d 应输出 PDF", tpl)
		require.Greater(t, len(out), 1000)
	}
}

func TestCompanyApplyDownloadResumeRejectsExhaustedOrMissingEntitlement(t *testing.T) {
	db := testutil.NewMemoryDB(t,
		&hrcModel.PersonalJobsApply{},
		&hrcModel.Resume{},
		&hrcModel.ResumeEducation{},
		&hrcModel.ResumeWork{},
		&hrcModel.ResumeLanguage{},
		&hrcModel.ResumeTraining{},
		&hrcModel.ResumeCredent{},
		&hrcModel.ResumeProject{},
		&hrcModel.ResumeSkill{},
		&hrcModel.ResumePortfolio{},
		&hrcModel.ResumeStudentLeader{},
		&hrcModel.MembersSetmeal{},
		&hrcModel.ResumeDownload{},
		&hrcModel.Pms{},
		&hrcModel.MembersMsgtip{},
	)
	resume := hrcModel.Resume{UID: 1, FullName: "李四"}
	require.NoError(t, db.Create(&resume).Error)
	apply := hrcModel.PersonalJobsApply{ResumeID: resume.ID, PersonalUID: 1, CompanyUID: 200}
	require.NoError(t, db.Create(&apply).Error)

	svc := &CompanyApplyService{}
	_, err := svc.DownloadResume(context.Background(), 200, apply.DID)
	require.ErrorIs(t, err, ErrResumeDownloadEntitlement)

	require.NoError(t, db.Create(&hrcModel.MembersSetmeal{
		UID: 200, ExpireAt: hrcModel.Now().Add(3600 * time.Second), ResumeDownloadsTotal: 1, ResumeDownloadsUsed: 1,
	}).Error)
	_, err = svc.DownloadResume(context.Background(), 200, apply.DID)
	require.ErrorIs(t, err, ErrResumeDownloadLimit)
}
