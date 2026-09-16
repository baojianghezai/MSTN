package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
)

// bizModel 名硕人力业务表（ms_*）注册
// 设计依据：01_数据库设计（一期 71 张对照，ms_ 前缀与 GVA sys_* 共存）
// 注意：业务表不在此批量 AutoMigrate 全部 71 张，按里程碑增量注册，
// 避免一次性建表过多；后续模块模型追加到 hrcModels 切片即可。
func bizModel() error {
	db := global.GVA_DB

	hrcModels := []interface{}{
		// M2 账号模块
		&hrcModel.Members{},
		&hrcModel.MembersInfo{},
		&hrcModel.CompanyProfile{},
		&hrcModel.MembersBind{},
		&hrcModel.MembersAppeal{},
		&hrcModel.MembersLog{},
		&hrcModel.MembersMsgtip{},
		&hrcModel.Pms{},
		&hrcModel.Oauth{},
		&hrcModel.UnbindMobile{},
		&hrcModel.CompanyCancellationApply{},
		&hrcModel.AuditReason{},

		// N1（项目经历 P0 补入）：简历主表 + 项目经历子表
		&hrcModel.Resume{},
		&hrcModel.ResumeProject{},

		// N2（视频面试 P0 补入）
		&hrcModel.VideoInterview{},

		// N5（支付日志 P0 补入）
		&hrcModel.WxpayLog{},

		// MVP 模块 A：职位发布与管理（03 模块）
		&hrcModel.Jobs{},
		&hrcModel.JobsTmp{},
		&hrcModel.JobsContact{},
		&hrcModel.JobsTag{},
		&hrcModel.JobPromotion{},

		// MVP 模块 C：投递（05 模块）
		&hrcModel.PersonalJobsApply{},
		&hrcModel.CompanyInterview{},
		&hrcModel.CompanyFavorite{},

		// 在线对话（IM：个人 ↔ 企业，WebSocket 实时）
		&hrcModel.ImSession{},
		&hrcModel.ImMessage{},

		// M1 基建：内容/配置/分类
		&hrcModel.Config{},
		&hrcModel.Page{},
		&hrcModel.Navigation{},
		&hrcModel.CategoryGroup{},
		&hrcModel.Category{},
		&hrcModel.CategoryJobs{},
		// 内容一期简版（资讯/招聘会/帮助）
		&hrcModel.Article{},

		// M3 收尾批：简历五张结构化子表（04 §2.4-2.8）
		&hrcModel.ResumeEducation{},
		&hrcModel.ResumeWork{},
		&hrcModel.ResumeLanguage{},
		&hrcModel.ResumeTraining{},
		&hrcModel.ResumeCredent{},
		// 简历可选加分项子表（专业技能/个人作品/学生干部经历）
		&hrcModel.ResumeSkill{},
		&hrcModel.ResumePortfolio{},
		&hrcModel.ResumeStudentLeader{},
		&hrcModel.Setmeal{},
		&hrcModel.MembersSetmeal{},
		&hrcModel.ResumeDownload{},
		&hrcModel.Order{},
		&hrcModel.PaymentNotifyLog{},
		&hrcModel.DataCleanupLog{},
	}

	for _, m := range hrcModels {
		if err := db.AutoMigrate(m); err != nil {
			return err
		}
	}

	// 投递去重由「企业级」放宽为「职位级」：AutoMigrate 不会删除旧唯一索引，显式清理
	if db.Migrator().HasIndex(&hrcModel.PersonalJobsApply{}, "uk_uid_resume_company") {
		if err := db.Migrator().DropIndex(&hrcModel.PersonalJobsApply{}, "uk_uid_resume_company"); err != nil {
			return err
		}
	}

	// M4 商业化最小闭环：套餐、企业权益与订单。
	return seedBiz(db)
}
