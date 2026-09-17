package initialize

import (
	"errors"
	"strconv"

	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"gorm.io/gorm"
)

// seedBiz 业务表最小种子数据（仅新增缺失项，幂等）。
func seedBiz(db *gorm.DB) error {
	if err := seedNavigations(db); err != nil {
		return err
	}
	if err := seedPages(db); err != nil {
		return err
	}
	if err := seedCategories(db); err != nil {
		return err
	}
	if err := seedJobCategories(db); err != nil {
		return err
	}
	if err := seedJobsDisplayConfig(db); err != nil {
		return err
	}
	if err := seedArticles(db); err != nil {
		return err
	}
	return seedSetmeals(db)
}

// seedArticles 内容一期简版种子（资讯/招聘会/帮助，仅首次建库时插入）
func seedArticles(db *gorm.DB) error {
	var count int64
	if err := db.Model(&hrcModel.Article{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	now := hrcModel.Now()
	articles := []hrcModel.Article{
		{Type: hrcModel.ArticleTypeNews, Title: "名硕人才网正式上线，助力企业高效招聘", Summary: "汇聚真实职位与优质人才，让求职招聘更简单。", Content: "名硕人才网正式上线，平台聚焦真实职位直投、企业招聘与人才服务。\n\n企业可发布职位、下载简历、发起面试；求职者可在线制作简历、投递职位并实时沟通。", Source: "名硕人才网", Sort: 10, Display: 1, AddTime: now, UpdateTime: now},
		{Type: hrcModel.ArticleTypeNews, Title: "简历制作小技巧：让 HR 一眼看中你", Summary: "填好期望薪资、突出项目经历，完善度越高越容易被下载。", Content: "1. 完善基本信息与求职意向；\n2. 量化工作业绩与项目成果；\n3. 按需添加技能、证书等加分项；\n4. 上传 PDF 附件简历，企业下载更专业。", Source: "名硕人才网", Sort: 9, Display: 1, AddTime: now, UpdateTime: now},
		{Type: hrcModel.ArticleTypeNews, Title: "春季招聘会预告：百余家企业现场纳才", Summary: "线上线下同步，覆盖互联网、制造、医疗等行业。", Content: "春季招聘会即将启动，百余家企业现场纳才，欢迎求职者到场应聘。", Source: "名硕人才网", Sort: 8, Display: 1, AddTime: now, UpdateTime: now},
		{Type: hrcModel.ArticleTypeJobfair, Title: "名硕人才网春季大型综合招聘会", Summary: "互联网/制造/医疗/教育多行业专场。", Content: "现场设企业展位、简历诊断、面试洽谈区，欢迎求职者携简历参加。", HoldTime: "2026-03-15 09:00-16:00", Address: "市人才市场一楼大厅", Organizer: "名硕人才网", Sort: 10, Display: 1, AddTime: now, UpdateTime: now},
		{Type: hrcModel.ArticleTypeJobfair, Title: "互联网专场招聘会", Summary: "聚焦研发、产品、设计、运营岗位。", Content: "面向互联网从业者的专场招聘会，覆盖研发、产品、设计、运营等岗位。", HoldTime: "2026-04-20 13:30-17:30", Address: "高新区人力资源产业园", Organizer: "名硕人才网", Sort: 9, Display: 1, AddTime: now, UpdateTime: now},
		{Type: hrcModel.ArticleTypeHelp, Title: "如何发布职位？", Summary: "企业中心 → 职位管理 → 发布职位。", Content: "登录企业账号后，进入「职位管理」点击「发布职位」，填写职位信息与联系方式，提交后等待审核，审核通过即展示。", Sort: 10, Display: 1, AddTime: now, UpdateTime: now},
		{Type: hrcModel.ArticleTypeHelp, Title: "如何下载简历？", Summary: "收到简历 → 下载简历，消耗套餐下载权益。", Content: "企业进入「收到简历」，点击「下载简历」即可下载；有 PDF 附件简历时优先下发附件，否则按求职者所选模板生成 PDF。首次下载消耗套餐下载次数，重复下载不扣。", Sort: 9, Display: 1, AddTime: now, UpdateTime: now},
		{Type: hrcModel.ArticleTypeHelp, Title: "如何发起在线沟通？", Summary: "职位详情/收到简历页可发起在线对话。", Content: "求职者可在职位详情页点击「在线沟通」，企业可在「收到简历」列表点击「在线沟通」，双方实时对话。", Sort: 8, Display: 1, AddTime: now, UpdateTime: now},
	}
	return db.Create(&articles).Error
}

func seedSetmeals(db *gorm.DB) error {
	var count int64
	if err := db.Model(&hrcModel.Setmeal{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return seedDefaultPromotionSlots(db)
	}
	if err := db.Create(&[]hrcModel.Setmeal{
		{Name: "基础版", Price: 9900, DurationDays: 30, JobsMeanwhile: 3, ResumeDownloads: 10, HomePushSlots: 0, HomeAdSlots: 0, EnableVideo: false, EnableJobfair: false, Display: true, Sort: 1, Description: "适合初次招聘的企业"},
		{Name: "专业版", Price: 29900, DurationDays: 90, JobsMeanwhile: 10, ResumeDownloads: 80, HomePushSlots: 1, HomeAdSlots: 0, EnableVideo: true, EnableJobfair: true, Display: true, Sort: 2, Description: "适合稳定招聘的企业"},
		{Name: "旗舰版", Price: 69900, DurationDays: 365, JobsMeanwhile: 30, ResumeDownloads: 500, HomePushSlots: 3, HomeAdSlots: 1, EnableVideo: true, EnableJobfair: true, Display: true, Sort: 3, Description: "适合全年招聘需求"},
	}).Error; err != nil {
		return err
	}
	return nil
}

// Existing default plans predate homepage exposure. Backfill only the known
// seeded plans still carrying their original descriptions and zero new quotas.
func seedDefaultPromotionSlots(db *gorm.DB) error {
	defaults := []struct {
		name, description string
		push, ad          int
	}{
		{"基础版", "适合初次招聘的企业", 0, 0},
		{"专业版", "适合稳定招聘的企业", 1, 0},
		{"旗舰版", "适合全年招聘需求", 3, 1},
	}
	for _, item := range defaults {
		if item.push == 0 && item.ad == 0 {
			continue
		}
		if err := db.Model(&hrcModel.Setmeal{}).
			Where("name = ? AND description = ? AND home_push_slots = 0 AND home_ad_slots = 0", item.name, item.description).
			Updates(map[string]interface{}{"home_push_slots": item.push, "home_ad_slots": item.ad}).Error; err != nil {
			return err
		}
	}
	// 举办招聘会权益回填（专业版/旗舰版；仅补默认未开启的行）
	for _, name := range []string{"专业版", "旗舰版"} {
		if err := db.Model(&hrcModel.Setmeal{}).
			Where("name = ? AND enable_jobfair = ?", name, false).
			Update("enable_jobfair", true).Error; err != nil {
			return err
		}
	}
	return nil
}

// seedJobsDisplayConfig 职位显示方式默认配置（修改批 X2，08-20 派活方拍板：发布后必须后台审核）
// 仅当配置不存在时插入 "1"（审核后显示）；已存在的库不覆盖（本地库由运维 UPDATE 同步）
func seedJobsDisplayConfig(db *gorm.DB) error {
	var count int64
	if err := db.Model(&hrcModel.Config{}).Where("name = ?", "mscms_jobs_display").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return db.Create(&hrcModel.Config{
		CfgGroup: "jobs",
		Name:     "mscms_jobs_display",
		Value:    "1",
		Remark:   "职位显示方式：1=审核后显示（默认，X2）2=直接显示",
		AddTime:  hrcModel.Now(),
	}).Error
}

func seedNavigations(db *gorm.DB) error {
	var count int64
	db.Model(&hrcModel.Navigation{}).Count(&count)
	if count > 0 {
		return nil
	}
	now := hrcModel.Now()
	navs := []hrcModel.Navigation{
		{Title: "首页", URL: "/", Sort: 1, Display: 1, AddTime: now},
		{Title: "找工作", URL: "/jobs", Sort: 2, Display: 1, AddTime: now},
		{Title: "找企业", URL: "/companies", Sort: 3, Display: 1, AddTime: now},
		{Title: "招聘会", URL: "/jobfairs", Sort: 4, Display: 1, AddTime: now},
		{Title: "资讯", URL: "/news", Sort: 5, Display: 1, AddTime: now},
		{Title: "帮助", URL: "/help", Sort: 6, Display: 1, AddTime: now},
	}
	return db.Create(&navs).Error
}

func seedPages(db *gorm.DB) error {
	var count int64
	db.Model(&hrcModel.Page{}).Count(&count)
	if count > 0 {
		return nil
	}
	pages := []hrcModel.Page{
		{Alias: "about", Title: "关于我们", Contents: "<p>关于我们（待编辑）</p>", AddTime: hrcModel.Now()},
	}
	return db.Create(&pages).Error
}

// seedGroup 分类分组定义 + 该分组下的基础分类值。
type seedGroup struct {
	alias string
	name  string
	sort  int
	items []string
}

func seedCategories(db *gorm.DB) error {
	groups := []seedGroup{
		{alias: "education", name: "学历", sort: 1, items: []string{"初中及以下", "高中", "中专", "大专", "本科", "硕士", "博士"}},
		{alias: "experience", name: "经验", sort: 2, items: []string{"在校生", "应届生", "1年以内", "1-3年", "3-5年", "5-10年", "10年以上"}},
		{alias: "wage", name: "薪资", sort: 3, items: []string{"3K以下", "3-5K", "5-10K", "10-15K", "15-20K", "20-30K", "30K以上", "面议"}},
		{alias: "trade", name: "行业", sort: 4, items: []string{"互联网/IT", "电子商务", "金融", "教育培训", "医疗健康", "制造业", "房地产", "物流运输", "餐饮服务", "其他"}},
		// 地区（district）暂不 seed：省市县 ~2800 条，后续从原 74CMS sql_category_district.sql 导入
		{alias: "district", name: "地区", sort: 5, items: nil},
		{alias: "major", name: "专业", sort: 6, items: []string{
			"计算机科学与技术", "软件工程", "电子信息工程", "通信工程", "自动化",
			"机械设计制造及其自动化", "土木工程", "电气工程及其自动化", "金融学", "会计学",
			"财务管理", "市场营销", "工商管理", "人力资源管理", "国际经济与贸易",
			"法学", "汉语言文学", "英语", "临床医学", "护理学",
			"教育学", "学前教育", "其他",
		}},
		{alias: "sex", name: "性别", sort: 7, items: []string{"男", "女"}},
		{alias: "marriage", name: "婚姻", sort: 8, items: []string{"未婚", "已婚", "保密"}},
		{alias: "nature", name: "企业性质", sort: 9, items: []string{"国有企业", "民营企业", "外资企业", "合资企业", "事业单位", "政府机关", "其他"}},
		{alias: "scale", name: "企业规模", sort: 10, items: []string{"少于50人", "50-99人", "100-499人", "500-999人", "1000人以上"}},
		{alias: "jobnature", name: "职位性质", sort: 13, items: nil},
		{alias: "jobtag", name: "职位福利", sort: 14, items: nil},
		{alias: "resumetag", name: "简历标签", sort: 15, items: nil},
		{alias: "language", name: "语言", sort: 16, items: nil},
		{alias: "languagelevel", name: "语言熟练度", sort: 17, items: nil},
		{alias: "current", name: "当前求职状态", sort: 18, items: nil},
		{alias: "age", name: "年龄段", sort: 19, items: nil},
		// 保留现有的常用职位；完整目录在 legacyJobCatalogData 中增量补齐。
		{alias: "jobtitle", name: "职位名称", sort: 12, items: []string{
			"Java开发", "Go开发", "前端开发", "Python开发", "C++开发", "全栈开发",
			"移动端开发", "游戏开发", "嵌入式开发", "小程序开发",
			"测试工程师", "自动化测试", "数据分析师", "数据工程师", "算法工程师",
			"大数据工程师", "运维工程师", "DBA", "安全工程师", "架构师",
			"项目经理", "产品经理", "高级产品经理", "UI设计师", "视觉设计师", "交互设计师",
			"运营专员", "用户运营", "内容运营", "电商运营", "新媒体运营",
			"市场专员", "品牌策划", "销售代表", "销售经理", "渠道销售", "客服专员",
			"HR专员", "会计", "财务分析师", "行政专员", "律师", "翻译", "编辑",
			"采购专员", "质检员", "物流专员", "教师", "医生", "护士",
		}},
	}

	groupIDs := make(map[string]uint64, len(groups))
	// 幂等：分组和运营已有分类都保留，仅补齐缺失的基础分类。
	for _, g := range groups {
		var group hrcModel.CategoryGroup
		err := db.Where("alias = ?", g.alias).First(&group).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			group = hrcModel.CategoryGroup{Alias: g.alias, Name: g.name, Sort: g.sort}
			if err := db.Create(&group).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		if err := seedCategoryNames(db, group.ID, g.items); err != nil {
			return err
		}
		groupIDs[g.alias] = group.ID
	}
	for alias, names := range legacyFlatCategories() {
		if err := seedCategoryNames(db, groupIDs[alias], names); err != nil {
			return err
		}
	}
	if err := seedCategoryNames(db, groupIDs["jobtitle"], legacyJobTitles()); err != nil {
		return err
	}
	return seedDistrictCategories(db, groupIDs["district"], legacyDistrictCatalog())
}

func seedCategoryNames(db *gorm.DB, groupID uint64, names []string) error {
	if groupID == 0 || len(names) == 0 {
		return nil
	}

	var existing []string
	if err := db.Model(&hrcModel.Category{}).
		Where("group_id = ? AND parent_id = ?", groupID, 0).
		Pluck("name", &existing).Error; err != nil {
		return err
	}
	known := make(map[string]struct{}, len(existing))
	for _, name := range existing {
		known[name] = struct{}{}
	}

	var maxSort int
	if err := db.Model(&hrcModel.Category{}).
		Where("group_id = ? AND parent_id = ?", groupID, 0).
		Select("COALESCE(MAX(sort), 0)").Scan(&maxSort).Error; err != nil {
		return err
	}
	missing := make([]hrcModel.Category, 0, len(names))
	for _, name := range names {
		if name == "" {
			continue
		}
		if _, ok := known[name]; ok {
			continue
		}
		maxSort++
		known[name] = struct{}{}
		missing = append(missing, hrcModel.Category{GroupID: groupID, Name: name, Sort: maxSort, Display: 1})
	}
	if len(missing) == 0 {
		return nil
	}
	return db.Create(&missing).Error
}

func seedDistrictCategories(db *gorm.DB, groupID uint64, entries []legacyDistrictEntry) error {
	if groupID == 0 || len(entries) == 0 {
		return nil
	}

	var existing []hrcModel.Category
	if err := db.Where("group_id = ?", groupID).Find(&existing).Error; err != nil {
		return err
	}
	categoryIDs := make(map[string]uint64, len(existing))
	maxSort := make(map[uint64]int)
	for _, category := range existing {
		categoryIDs[categoryKey(category.ParentID, category.Name)] = category.ID
		if category.Sort > maxSort[category.ParentID] {
			maxSort[category.ParentID] = category.Sort
		}
	}

	legacyIDs := make(map[int]uint64, len(entries))
	for _, entry := range entries {
		parentID := uint64(0)
		if entry.ParentID != 0 {
			var ok bool
			parentID, ok = legacyIDs[entry.ParentID]
			if !ok {
				return errors.New("legacy district catalogue has an invalid parent reference")
			}
		}
		key := categoryKey(parentID, entry.Name)
		if categoryID, ok := categoryIDs[key]; ok {
			legacyIDs[entry.ID] = categoryID
			continue
		}
		maxSort[parentID]++
		category := hrcModel.Category{GroupID: groupID, ParentID: parentID, Name: entry.Name, Sort: maxSort[parentID], Display: 1}
		if err := db.Create(&category).Error; err != nil {
			return err
		}
		categoryIDs[key] = category.ID
		legacyIDs[entry.ID] = category.ID
	}
	return nil
}

func categoryKey(parentID uint64, name string) string {
	return strconv.FormatUint(parentID, 10) + "\x00" + name
}

// seedJobCategories 增量导入职位三级分类。已有运营分类不会被删除或覆盖。
func seedJobCategories(db *gorm.DB) error {
	var group hrcModel.CategoryGroup
	err := db.Where("alias = ?", "jobcategory").First(&group).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		group = hrcModel.CategoryGroup{Alias: "jobcategory", Name: "职位分类", Sort: 11}
		if err := db.Create(&group).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	for _, entry := range legacyJobCatalog() {
		topClassID, err := seedJobCategoryNode(db, group.ID, 0, entry.TopClass)
		if err != nil {
			return err
		}
		categoryID, err := seedJobCategoryNode(db, group.ID, topClassID, entry.Category)
		if err != nil {
			return err
		}
		if _, err := seedJobCategoryNode(db, group.ID, categoryID, entry.Title); err != nil {
			return err
		}
	}
	return nil
}

func seedJobCategoryNode(db *gorm.DB, groupID, parentID uint64, name string) (uint64, error) {
	var category hrcModel.Category
	err := db.Where("group_id = ? AND parent_id = ? AND name = ?", groupID, parentID, name).First(&category).Error
	if err == nil {
		return category.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	var maxSort int
	if err := db.Model(&hrcModel.Category{}).
		Where("group_id = ? AND parent_id = ?", groupID, parentID).
		Select("COALESCE(MAX(sort), 0)").Scan(&maxSort).Error; err != nil {
		return 0, err
	}
	category = hrcModel.Category{GroupID: groupID, ParentID: parentID, Name: name, Sort: maxSort + 1, Display: 1}
	if err := db.Create(&category).Error; err != nil {
		return 0, err
	}
	return category.ID, nil
}
