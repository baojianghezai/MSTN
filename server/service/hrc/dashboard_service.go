package hrc

import (
	"context"
	"errors"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
)

var errInvalidTrendMetric = errors.New("趋势指标非法（支持 register/resume/company/job/application）")

// DashboardService 统计看板/报表服务（08 §4 A6，11 §四 P0#4）
// 一期仅聚合已建表：会员/简历/企业/视频面试/申诉/注销；职位/投递/下载/订单收入随 M3/M4 补
type DashboardService struct{}

// DayMetrics 单日核心指标
type DayMetrics struct {
	PersonalUsers   int64 `json:"personalUsers"`   // 个人注册
	CompanyUsers    int64 `json:"companyUsers"`    // 企业注册
	Resumes         int64 `json:"resumes"`         // 新增简历
	Companies       int64 `json:"companies"`       // 新增企业资料
	Jobs            int64 `json:"jobs"`            // 新发布职位
	Applications    int64 `json:"applications"`    // 新增投递
	VideoInterviews int64 `json:"videoInterviews"` // 新增视频面试邀请
}

// TodoMetrics 待办提醒
type TodoMetrics struct {
	CompanyAudit        int64 `json:"companyAudit"`        // 待审企业（audit=2）
	ResumeAudit         int64 `json:"resumeAudit"`         // 待审简历（audit=2）
	JobAudit            int64 `json:"jobAudit"`            // 待审职位（jobs_tmp.audit=2）
	Appeal              int64 `json:"appeal"`              // 待处理申诉（status=0）
	CompanyCancellation int64 `json:"companyCancellation"` // 待处理企业注销（status=0）
}

// IncomeMetrics 收入（订单金额，M4 订单表建后补，现恒 0）
type IncomeMetrics struct {
	Today int64 `json:"today"`
	Month int64 `json:"month"`
}

// DashboardData 看板指标（#119）
type DashboardData struct {
	Today     DayMetrics    `json:"today"`
	Yesterday DayMetrics    `json:"yesterday"`
	Todo      TodoMetrics   `json:"todo"`
	Income    IncomeMetrics `json:"income"`
}

// TrendPoint 趋势点（register 用 personal/company，其余用 count）
type TrendPoint struct {
	Date     string `json:"date"`     // yyyy-MM-dd
	Personal int64  `json:"personal"` // register：个人注册
	Company  int64  `json:"company"`  // register：企业注册
	Count    int64  `json:"count"`    // resume/company：总数
}

// DistributionItem 分布项（code=分类值，cn=中文，count=数量）
type DistributionItem struct {
	Code  int64  `gorm:"column:code" json:"code"`
	CN    string `gorm:"column:cn" json:"cn"`
	Count int64  `gorm:"column:cnt" json:"count"`
}

// ResumeDistribution 求职者分布（性别/学历/经验）
type ResumeDistribution struct {
	Sex        []DistributionItem `json:"sex"`
	Education  []DistributionItem `json:"education"`
	Experience []DistributionItem `json:"experience"`
}

// CompanyDistribution 企业分布（性质/规模）
type CompanyDistribution struct {
	Nature []DistributionItem `json:"nature"`
	Scale  []DistributionItem `json:"scale"`
}

// Dashboard 看板指标（今日/昨日/待办/收入）
func (s *DashboardService) Dashboard(ctx context.Context) (*DashboardData, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	yesterdayStart := todayStart - 24*3600

	today, err := s.dayMetrics(ctx, todayStart, now.Unix()+1)
	if err != nil {
		return nil, err
	}
	yesterday, err := s.dayMetrics(ctx, yesterdayStart, todayStart)
	if err != nil {
		return nil, err
	}
	todo, err := s.todoMetrics(ctx)
	if err != nil {
		return nil, err
	}
	// 收入：ms_order 未建，恒 0（M4 补）
	return &DashboardData{Today: today, Yesterday: yesterday, Todo: todo, Income: IncomeMetrics{}}, nil
}

// Trend 趋势图（近 days 日按日聚合；metric: register/resume/company）
func (s *DashboardService) Trend(ctx context.Context, days int, metric string) ([]TrendPoint, error) {
	if days <= 0 || days > 90 {
		days = 30
	}
	now := time.Now()
	startDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(days - 1))
	start := startDay.Unix()

	points := make([]TrendPoint, days)
	for i := range points {
		points[i].Date = startDay.AddDate(0, 0, i).Format("2006-01-02")
	}

	db := global.GVA_DB.WithContext(ctx)
	switch metric {
	case "register":
		var members []hrcModel.Members
		if err := db.Select("utype", "reg_time").Where("reg_time >= ?", start).Find(&members).Error; err != nil {
			return nil, err
		}
		for _, m := range members {
			if idx := dayIndex(m.RegTime, startDay); idx >= 0 && idx < days {
				if m.Utype == 1 {
					points[idx].Personal++
				} else if m.Utype == 2 {
					points[idx].Company++
				}
			}
		}
	case "resume":
		var resumes []hrcModel.Resume
		if err := db.Select("addtime").Where("addtime >= ?", start).Find(&resumes).Error; err != nil {
			return nil, err
		}
		for _, r := range resumes {
			if idx := dayIndex(r.AddTime, startDay); idx >= 0 && idx < days {
				points[idx].Count++
			}
		}
	case "company":
		var profiles []hrcModel.CompanyProfile
		if err := db.Select("addtime").Where("addtime >= ?", start).Find(&profiles).Error; err != nil {
			return nil, err
		}
		for _, p := range profiles {
			if idx := dayIndex(p.AddTime, startDay); idx >= 0 && idx < days {
				points[idx].Count++
			}
		}
	case "job":
		var jobs []hrcModel.Jobs
		if err := db.Select("addtime").Where("addtime >= ? AND deleted_at = 0", start).Find(&jobs).Error; err != nil {
			return nil, err
		}
		for _, job := range jobs {
			if idx := dayIndex(job.AddTime, startDay); idx >= 0 && idx < days {
				points[idx].Count++
			}
		}
	case "application":
		var applications []hrcModel.PersonalJobsApply
		if err := db.Select("apply_addtime").Where("apply_addtime >= ?", start).Find(&applications).Error; err != nil {
			return nil, err
		}
		for _, application := range applications {
			if idx := dayIndex(application.ApplyAddtime, startDay); idx >= 0 && idx < days {
				points[idx].Count++
			}
		}
	default:
		return nil, errInvalidTrendMetric
	}
	return points, nil
}

// ResumeDistribution 求职者分布（性别/学历/经验）
func (s *DashboardService) ResumeDistribution(ctx context.Context) (*ResumeDistribution, error) {
	sex, err := s.distribution(ctx, "ms_resume", "sex", "sex_cn")
	if err != nil {
		return nil, err
	}
	edu, err := s.distribution(ctx, "ms_resume", "education", "education_cn")
	if err != nil {
		return nil, err
	}
	exp, err := s.distribution(ctx, "ms_resume", "experience", "experience_cn")
	if err != nil {
		return nil, err
	}
	return &ResumeDistribution{Sex: sex, Education: edu, Experience: exp}, nil
}

// CompanyDistribution 企业分布（性质/规模）
func (s *DashboardService) CompanyDistribution(ctx context.Context) (*CompanyDistribution, error) {
	nature, err := s.distribution(ctx, "ms_company_profile", "nature", "nature_cn")
	if err != nil {
		return nil, err
	}
	scale, err := s.distribution(ctx, "ms_company_profile", "scale", "scale_cn")
	if err != nil {
		return nil, err
	}
	return &CompanyDistribution{Nature: nature, Scale: scale}, nil
}

// dayMetrics 单日新增指标（[start, end)）
func (s *DashboardService) dayMetrics(ctx context.Context, start, end int64) (DayMetrics, error) {
	var m DayMetrics
	db := global.GVA_DB.WithContext(ctx)
	if err := db.Model(&hrcModel.Members{}).Where("utype = 1 AND reg_time >= ? AND reg_time < ?", start, end).Count(&m.PersonalUsers).Error; err != nil {
		return m, err
	}
	if err := db.Model(&hrcModel.Members{}).Where("utype = 2 AND reg_time >= ? AND reg_time < ?", start, end).Count(&m.CompanyUsers).Error; err != nil {
		return m, err
	}
	if err := db.Model(&hrcModel.Resume{}).Where("addtime >= ? AND addtime < ?", start, end).Count(&m.Resumes).Error; err != nil {
		return m, err
	}
	if err := db.Model(&hrcModel.CompanyProfile{}).Where("addtime >= ? AND addtime < ?", start, end).Count(&m.Companies).Error; err != nil {
		return m, err
	}
	if err := db.Model(&hrcModel.Jobs{}).Where("addtime >= ? AND addtime < ? AND deleted_at = 0", start, end).Count(&m.Jobs).Error; err != nil {
		return m, err
	}
	if err := db.Model(&hrcModel.PersonalJobsApply{}).Where("apply_addtime >= ? AND apply_addtime < ?", start, end).Count(&m.Applications).Error; err != nil {
		return m, err
	}
	if err := db.Model(&hrcModel.VideoInterview{}).Where("addtime >= ? AND addtime < ?", start, end).Count(&m.VideoInterviews).Error; err != nil {
		return m, err
	}
	return m, nil
}

// todoMetrics 待办提醒计数
func (s *DashboardService) todoMetrics(ctx context.Context) (TodoMetrics, error) {
	var t TodoMetrics
	db := global.GVA_DB.WithContext(ctx)
	if err := db.Model(&hrcModel.CompanyProfile{}).Where("audit = 2").Count(&t.CompanyAudit).Error; err != nil {
		return t, err
	}
	if err := db.Model(&hrcModel.Resume{}).Where("audit = 2").Count(&t.ResumeAudit).Error; err != nil {
		return t, err
	}
	if err := db.Model(&hrcModel.JobsTmp{}).Where("audit = 2 AND deleted_at = 0").Count(&t.JobAudit).Error; err != nil {
		return t, err
	}
	if err := db.Model(&hrcModel.MembersAppeal{}).Where("status = 0").Count(&t.Appeal).Error; err != nil {
		return t, err
	}
	if err := db.Model(&hrcModel.CompanyCancellationApply{}).Where("status = 0").Count(&t.CompanyCancellation).Error; err != nil {
		return t, err
	}
	return t, nil
}

// distribution 分类分布聚合（GROUP BY 分类码，取代表中文）
func (s *DashboardService) distribution(ctx context.Context, table, codeCol, cnCol string) ([]DistributionItem, error) {
	var items []DistributionItem
	err := global.GVA_DB.WithContext(ctx).Table(table).
		Select(codeCol + " AS code, MAX(" + cnCol + ") AS cn, COUNT(*) AS cnt").
		Where(codeCol + " > 0").
		Group(codeCol).
		Order(codeCol + " ASC").
		Scan(&items).Error
	if items == nil {
		items = []DistributionItem{}
	}
	return items, err
}

// dayIndex 时间戳 → 相对 startDay 的天偏移（越界返回 -1）
func dayIndex(ts int64, startDay time.Time) int {
	t := time.Unix(ts, 0)
	d := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return int(d.Sub(startDay).Hours() / 24)
}
