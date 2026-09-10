# MSRC人才网 — 第一阶段待完成模块开发指南

> 基于 design/ 12份设计文档 + 代码库现状分析
> 生成日期：2026-09-10

## 总预估

**25 ~ 35 人天（约 5 ~ 7 周，1人全职）**

---

## 开发顺序（按优先级）

### Step 1：A8 定时任务 × 11（P0 阻塞型，3~4天）

> 没有定时任务 → 套餐/职位不过期、订单不关闭、推广不重置

**设计文档**：`design/08_后台RBAC系统统计模块设计.md` §6

**需要实现的 11 个任务**：

| # | 任务名 | 触发频率 | 逻辑 |
|---|--------|---------|------|
| 1 | 职位过期 | 每日 | `deadline < now AND display != 0` → display=0 |
| 2 | 套餐到期 | 每日 | `setmeal_deadline < now` → 清零套餐权益 |
| 3 | 套餐到期提醒 | 每日 | 到期前 7/3/1 天发站内信 |
| 4 | 自动刷新职位 | 每日 | 有 refresh_jobs_free 额度的套餐用户，自动刷新 |
| 5 | 订单超时关闭 | 每30分钟 | `is_paid=1 AND addtime < now-30min` → 原子关闭(1→3) |
| 6 | 推广过期 | 每日 | stick/emergency 到期重置 |
| 7 | 支付对账 | 每日凌晨 | 微信/支付宝账单 vs ms_order 核对 |
| 8 | 数据库备份 | 每日凌晨 | mysqldump + 保留 N 天 |
| 9 | 日志清理 | 每周 | 清理 90 天前的 ms_admin_log / ms_members_log |
| 10 | 统计日聚合 | 每日凌晨 | 汇总前一天注册/投递/下载/订单数据 |
| 11 | 支付回调日志清理 | 每月 | 清理 180 天前的 ms_wxpay_log |

**现有代码**：
- `server/task/registry.go` — 通用 task 注册器（已有）
- `server/task/clearTable.go` — GVA 系统清理（已有，可参考模式）
- `server/initialize/timed_task.go` — GVA 定时任务加载器（已有）
- **需要新建**：`server/task/hrc/` 目录，每个任务一个文件

**Go 实现要点**：
```go
// task/hrc/ 目录结构
task/hrc/
  job_expire.go          // #1 职位过期
  setmeal_expire.go      // #2 套餐到期
  setmeal_remind.go      // #3 套餐提醒
  auto_refresh.go        // #4 自动刷新
  order_timeout.go       // #5 订单超时关闭
  promotion_expire.go    // #6 推广过期
  payment_reconcile.go   // #7 支付对账
  db_backup.go           // #8 数据库备份
  log_cleanup.go         // #9 日志清理
  stats_daily.go         // #10 统计日聚合
  wxpay_log_cleanup.go   // #11 支付日志清理
```

**任务注册方式**：
```go
// initialize/timed_task.go 或 gorm.go 中注册
import hrcTask "github.com/flipped-aurora/gin-vue-admin/server/task/hrc"

// 在 GVA 定时任务系统中注册
hrcTask.RegisterAll(db)
```

**验收标准**：
- 11 个任务全部注册到 sys_timed_tasks 表
- 手动触发每个任务不报错
- 订单超时关闭有并发安全（atomic UPDATE is_paid=1→3）
- 套餐到期正确清零 jobs_meanway / download_resume 等配额

---

### Step 2：E6 帮助中心（P0 核心，1天）

> 最简单的 CMS 模块，先拿下

**设计文档**：`design/07_内容招聘会站内信模块设计.md` §2.5

**数据表**：
```sql
ms_help_category (id, name, sort, addtime)
ms_help (id, cid, title, contents, sort, addtime)
```

**需要实现**：
1. Model: `model/hrc/help.go` — HelpCategory + Help
2. Service: `service/hrc/help_service.go` — CRUD
3. API: `api/v1/hrc/help.go` — PublicGetHelpCategories, PublicGetHelpList, PublicGetHelpDetail + Admin CRUD
4. Router: `router/hrc/help.go`
5. Frontend: `views/help/index.vue` — 分类 + 文章列表 + 详情
6. Admin: 后台帮助管理页

**验收标准**：
- 前台 /help 展示分类和文章列表
- 后台可增删改查帮助分类和文章

---

### Step 3：E1 资讯/文章（P0 核心，2~3天）

**设计文档**：`design/07_内容招聘会站内信模块设计.md` §2.1

**数据表**：
```sql
ms_article_category (id, parentid, name, sort, addtime)
ms_article (id, cid, title, contents, cover, source, author, seo_title, seo_keywords, seo_description, clicks, is_display, is_focus, is_recommend, sort, addtime)
```

**需要实现**：
1. Model: `model/hrc/article.go`
2. Service: `service/hrc/article_service.go` — 列表(分页+分类筛选)、详情(点击+1)、Admin CRUD
3. API: `api/v1/hrc/cms.go`（已有基础，扩展 article 接口）
4. Router: 扩展 `router/hrc/cms.go`
5. Frontend: `views/news/list.vue` + `views/news/detail.vue`
6. Admin: 后台资讯管理页

---

### Step 4：E2 招聘会（P0 核心，4~5天）

> 最复杂的模块，5张关联表

**设计文档**：`design/07_内容招聘会站内信模块设计.md` §2.3

**数据表**：
```sql
ms_jobfair (id, title, content, start_time, end_time, address, area_id,预定_status, position_limit, img, is_display, addtime)
ms_jobfair_area (id, jobfair_id, area_name, area_price, area_max, area_sort)
ms_jobfair_position (id, jobfair_id, position_name, position_price, position_max, position_sort, img)
ms_jobfair_exhibitors (id, jobfair_id, uid, area_id, position_id, companyname, audit, audit_reason, addtime)
ms_jobfair_personal (id, jobfair_id, uid, realname, mobile, addtime)
```

**业务流程**：
1. 后台创建招聘会 → 设置展区/展位
2. 企业报名 → 选择展区/展位 → 审核 → 通过/拒绝
3. 前台展示招聘会列表 + 详情（含展区/展位平面图）
4. 个人预约参加

---

### Step 5：A3 敏感词 + 搜索优化 + 真实短信/验证码（P1，3~4天）

- 敏感词：`ms_badword` 表已有 model，需实现过滤 Service
- 搜索：Redis 热词 Top50，搜索次数统计
- 短信：接入阿里云/腾讯云 SMS API
- 验证码：接入图片验证码服务

---

### Step 6：微信登录（P1，2~3天）

- 公众号 OAuth2 授权回调
- 获取 openid + 用户信息
- 绑定/解绑流程

---

### Step 7：后台内容管理页 + 日志页（P1，3~4天）

- 资讯 CRUD 管理页
- 招聘会 CRUD 管理页
- 帮助 CRUD 管理页
- 操作日志 / 会员日志列表页

---

### Step 8：联调 + 部署（收尾，2~3天）

- 全链路 E2E 测试
- 支付沙箱回调验证
- 部署包打包（exe + build + SQL + 文档）

---

## 代码规范

- **分层**：Router → API → Service → Model
- **分页**：使用 `request.PageInfo`，返回 `{code, message, data: {list, total, page, pageSize}}`
- **事务**：涉及多表写入用 `db.Transaction()`
- **错误码**：参考 `design/09_API接口规范.md` §1.3
- **审计日志**：关键操作写 `ms_admin_log`
- **定时任务**：统一在 `task/hrc/` 下，通过 GVA 定时任务系统注册
