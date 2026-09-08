package initialize

import (
	"context"
	"errors"
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"
	"gorm.io/gorm"
)

// EnsureHrcMenus 幂等注册 hrc 业务菜单并绑定到超级管理员角色
//
// 为什么放在启动时而不是菜单 seed（server/source/system/menu.go）：
// GVA 的 seed 只在数据库首次初始化（InitDB）时执行，已初始化过的库不会再跑；
// 这里每次启动都执行，用「不存在才插入」保证幂等，新环境 / 已有环境都能自动补齐菜单。
// 以后 hrc 新增后台页面，照这个模式在下面追加二级菜单即可，无需人工点菜单管理界面。
func EnsureHrcMenus() {
	db := global.GVA_DB
	if db == nil {
		return
	}
	ctx := context.Background()

	// 1. 一级菜单「名硕人力」（含子菜单，组件用 routerHolder 承载二级路由）
	parent := system.SysBaseMenu{}
	err := db.WithContext(ctx).Where("name = ?", "hrc").First(&parent).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		parent = system.SysBaseMenu{
			MenuLevel: 0,
			ParentId:  0,
			Path:      "hrc",
			Name:      "hrc",
			Component: "view/routerHolder.vue",
			Sort:      50,
			Meta: system.Meta{
				Title: "名硕人力",
				Icon:  "service",
			},
		}
		if err = db.WithContext(ctx).Create(&parent).Error; err != nil {
			logger.Bg().Mod("hrc").ErrorDetail("create hrc parent menu failed", err)
			return
		}
		logger.Bg().Mod("hrc").Info("hrc parent menu created")
	}

	// 2. 二级菜单「申诉处理」
	// 组件路径必须带 .vue 后缀：前端 asyncRouter 用精确等值匹配
	// （glob key 去掉 ../ 后 === component），缺后缀永远匹配不上组件。
	const appealComponent = "view/hrc/appeal/list.vue"
	child := system.SysBaseMenu{}
	err = db.WithContext(ctx).Where("name = ?", "appeal").First(&child).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		child = system.SysBaseMenu{
			MenuLevel: 1,
			ParentId:  parent.ID,
			Path:      "appeal",
			Name:      "appeal",
			Component: appealComponent,
			Sort:      1,
			Meta: system.Meta{
				Title: "申诉处理",
				Icon:  "tickets",
			},
		}
		if err = db.WithContext(ctx).Create(&child).Error; err != nil {
			logger.Bg().Mod("hrc").ErrorDetail("create hrc appeal menu failed", err)
			return
		}
		logger.Bg().Mod("hrc").Info("hrc appeal menu created")
	} else if child.Component != appealComponent {
		// 已存在但组件路径不对（早期注册少了 .vue 后缀）：启动时自动修正
		if err = db.WithContext(ctx).Model(&child).Update("component", appealComponent).Error; err != nil {
			logger.Bg().Mod("hrc").ErrorDetail("fix appeal menu component failed", err)
		} else {
			logger.Bg().Mod("hrc").Info("appeal menu component fixed to " + appealComponent)
		}
	}

	// 2b. 二级菜单「企业注销申请」（后台列表/处理）
	const companyCancelComponent = "view/hrc/company/cancellation.vue"
	cancelMenu := system.SysBaseMenu{}
	err = db.WithContext(ctx).Where("name = ?", "companyCancellation").First(&cancelMenu).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		cancelMenu = system.SysBaseMenu{
			MenuLevel: 1,
			ParentId:  parent.ID,
			Path:      "company-cancellation",
			Name:      "companyCancellation",
			Component: companyCancelComponent,
			Sort:      2,
			Meta: system.Meta{
				Title: "企业注销申请",
				Icon:  "office-building",
			},
		}
		if err = db.WithContext(ctx).Create(&cancelMenu).Error; err != nil {
			logger.Bg().Mod("hrc").ErrorDetail("create company cancellation menu failed", err)
			return
		}
		logger.Bg().Mod("hrc").Info("company cancellation menu created")
	} else if cancelMenu.Component != companyCancelComponent {
		if err = db.WithContext(ctx).Model(&cancelMenu).Update("component", companyCancelComponent).Error; err != nil {
			logger.Bg().Mod("hrc").ErrorDetail("fix company cancellation menu component failed", err)
		}
	}

	// 2c. 二级菜单「企业审核」（后台企业资质审核）
	const companyAuditComponent = "view/hrc/company/audit.vue"
	auditMenu := system.SysBaseMenu{}
	err = db.WithContext(ctx).Where("name = ?", "companyAudit").First(&auditMenu).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		auditMenu = system.SysBaseMenu{
			MenuLevel: 1,
			ParentId:  parent.ID,
			Path:      "company-audit",
			Name:      "companyAudit",
			Component: companyAuditComponent,
			Sort:      3,
			Meta: system.Meta{
				Title: "企业审核",
				Icon:  "check",
			},
		}
		if err = db.WithContext(ctx).Create(&auditMenu).Error; err != nil {
			logger.Bg().Mod("hrc").ErrorDetail("create company audit menu failed", err)
			return
		}
		logger.Bg().Mod("hrc").Info("company audit menu created")
	} else if auditMenu.Component != companyAuditComponent {
		if err = db.WithContext(ctx).Model(&auditMenu).Update("component", companyAuditComponent).Error; err != nil {
			logger.Bg().Mod("hrc").ErrorDetail("fix company audit menu component failed", err)
		}
	}

	// 2d-2g. N2-N5 新增业务菜单（视频面试/看板/统计报表/支付日志）
	dashboardMenu := ensureChildMenu(ctx, db, parent.ID, "hrcDashboard", "dashboard", "view/hrc/dashboard/index.vue", 4, "看板", "data-line")
	statisticsMenu := ensureChildMenu(ctx, db, parent.ID, "hrcStatistics", "statistics", "view/hrc/statistics/index.vue", 5, "统计报表", "pie-chart")
	videoMenu := ensureChildMenu(ctx, db, parent.ID, "videoInterview", "video-interview", "view/hrc/videoInterview/list.vue", 6, "视频面试", "video-camera")
	wxpayMenu := ensureChildMenu(ctx, db, parent.ID, "wxpayLog", "wxpay-log", "view/hrc/wxpayLog/list.vue", 7, "支付日志", "money")

	// 2h. MVP 模块 A：职位管理（后台列表+审核，双 Tab 承载 #121/#122/#123）
	jobsMenu := ensureChildMenu(ctx, db, parent.ID, "jobsManage", "jobs-manage", "view/hrc/jobs/list.vue", 8, "职位管理", "suitcase")
	billingMenu := ensureChildMenu(ctx, db, parent.ID, "billingManage", "billing-manage", "view/hrc/billing/index.vue", 9, "套餐订单", "credit-card")
	dataCleanupMenu := ensureChildMenu(ctx, db, parent.ID, "hrcDataCleanup", "data-cleanup", "view/hrc/dataCleanup/index.vue", 10, "数据清理", "trash-2-gva")

	// 3. 绑定到超级管理员角色（默认 authority_id=888，找不到再按名称兜底）
	adminID := findAdminAuthorityID(ctx, db)
	if adminID == "" {
		logger.Bg().Mod("hrc").Warn("admin authority not found, skip menu binding")
		return
	}
	bindMenuToAuthority(ctx, db, adminID, parent.ID)
	bindMenuToAuthority(ctx, db, adminID, child.ID)
	bindMenuToAuthority(ctx, db, adminID, cancelMenu.ID)
	bindMenuToAuthority(ctx, db, adminID, auditMenu.ID)
	for _, m := range []system.SysBaseMenu{dashboardMenu, statisticsMenu, videoMenu, wxpayMenu, jobsMenu, billingMenu, dataCleanupMenu} {
		if m.ID != 0 {
			bindMenuToAuthority(ctx, db, adminID, m.ID)
		}
	}
}

// ensureChildMenu 幂等注册 hrc 二级菜单（不存在则创建，组件路径不符则修正），返回菜单记录
func ensureChildMenu(ctx context.Context, db *gorm.DB, parentID uint, name, path, component string, sort int, title, icon string) system.SysBaseMenu {
	menu := system.SysBaseMenu{}
	err := db.WithContext(ctx).Where("name = ?", name).First(&menu).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		menu = system.SysBaseMenu{
			MenuLevel: 1,
			ParentId:  parentID,
			Path:      path,
			Name:      name,
			Component: component,
			Sort:      sort,
			Meta: system.Meta{
				Title: title,
				Icon:  icon,
			},
		}
		if err = db.WithContext(ctx).Create(&menu).Error; err != nil {
			logger.Bg().Mod("hrc").ErrorDetail("create hrc menu failed", err)
			return system.SysBaseMenu{}
		}
		logger.Bg().Mod("hrc").Info("hrc menu created: " + name)
	} else if menu.Component != component {
		if err = db.WithContext(ctx).Model(&menu).Update("component", component).Error; err != nil {
			logger.Bg().Mod("hrc").ErrorDetail("fix hrc menu component failed", err)
		}
	}
	return menu
}

// findAdminAuthorityID 查找超级管理员角色 ID（字符串形式，与 sys_authority_menus 列类型一致）
func findAdminAuthorityID(ctx context.Context, db *gorm.DB) string {
	var auth system.SysAuthority
	if err := db.WithContext(ctx).Where("authority_id = ?", 888).First(&auth).Error; err == nil {
		return strconv.FormatUint(uint64(auth.AuthorityId), 10)
	}
	var byName system.SysAuthority
	if err := db.WithContext(ctx).Where("authority_name = ?", "超级管理员").First(&byName).Error; err == nil {
		return strconv.FormatUint(uint64(byName.AuthorityId), 10)
	}
	return ""
}

// bindMenuToAuthority 把菜单绑定到角色（幂等：已绑定则跳过）
func bindMenuToAuthority(ctx context.Context, db *gorm.DB, authorityID string, menuID uint) {
	var row system.SysAuthorityMenu
	err := db.WithContext(ctx).
		Where("sys_authority_authority_id = ? AND sys_base_menu_id = ?", authorityID, menuID).
		First(&row).Error
	if err == nil {
		return
	}
	link := system.SysAuthorityMenu{
		MenuId:      strconv.FormatUint(uint64(menuID), 10),
		AuthorityId: authorityID,
	}
	if err = db.WithContext(ctx).Create(&link).Error; err != nil {
		logger.Bg().Mod("hrc").ErrorDetail("bind menu to authority failed", err)
	}
}
