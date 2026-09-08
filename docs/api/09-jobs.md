# 09 职位发布与管理（jobs）

> 对应设计：`03_职位模块设计.md`（发布/审核双表/暂停恢复刷新/搜索）、`12_MVP可跑通链路.md` 模块 A/B、`01_数据库设计.md` §1.4/§2.2、`09_API接口规范.md` §4.1(22-25)/§4.3(79-86)/§4.6.2(121-123)
> 实现状态：8 个企业端接口 + 4 个公开接口 + 3 个后台接口
> 说明：前缀 `/api/v1`；`company/**` 需 utype=2，`admin/**` 走 GVA admin JWT，`/jobs` 公开。

## 接口清单

### 前台公开（无鉴权）

| # | 方法 | 路径 | 请求类型 | 说明 |
|---|---|---|---|---|
| 22 | GET | /api/v1/jobs | Query | 职位列表（筛选+排序+分页） |
| 23 | GET | /api/v1/jobs/{id} | 无参 | 职位详情（click 自增） |
| 24 | GET | /api/v1/jobs/hot-words | 无参 | 热门搜索词 |
| 25 | GET | /api/v1/jobs/filters | 无参 | 筛选元数据（分类聚合，复用 #21） |

### 企业端（会员 JWT，utype=2）

| # | 方法 | 路径 | 请求类型 | 说明 |
|---|---|---|---|---|
| 79 | POST | /api/v1/company/jobs | JSON | 发布职位 |
| 80 | GET | /api/v1/company/jobs | Query | 我的职位列表（含待审） |
| 81 | GET | /api/v1/company/jobs/{id}?pending=0\|1 | Query | 职位详情（编辑回显，含 reason；pending=1 查 jobs_tmp） |
| 82 | PUT | /api/v1/company/jobs/{id}?pending=0\|1 | JSON | 编辑职位（写 tmp 待审 / 直接更新；**D1**：pending=1 原地重提 tmp 行；缺省时 jobs 表未命中自动回退 tmp） |
| 83 | DELETE | /api/v1/company/jobs/{id}?pending=0\|1 | 无参 | 删除职位（逻辑删除；**D1**：pending=1 软删 tmp 行；缺省时 jobs 表未命中自动回退 tmp） |
| 84 | PUT | /api/v1/company/jobs/{id}/pause | 无参 | 暂停 |
| 85 | PUT | /api/v1/company/jobs/{id}/resume | 无参 | 恢复 |
| 86 | PUT | /api/v1/company/jobs/{id}/refresh | 无参 | 刷新 |

### 后台（admin JWT + x-token）

| # | 方法 | 路径 | 请求类型 | 说明 |
|---|---|---|---|---|
| 121 | GET | /api/v1/admin/jobs | Query | 职位管理列表 |
| 122 | GET | /api/v1/admin/jobs/tmp | Query | 待审职位列表 |
| 123 | PUT | /api/v1/admin/jobs/{id}/audit | JSON | 职位审核（1=通过 3=不通过 + reason） |

---

## 审核双表机制（03 §2.1，配置 `mscms_jobs_display`）

> **修改批 X2（2026-08-20 拍板）**：职位发布后必须后台审核——**缺省/非法配置值 = `1`（审核后显示）**；仅显式 `2` 才直接显示。种子值已置 `1`，存量直显职位不受影响。

- `display=1`（**默认，审核后显示**）：新职位写 `ms_jobs_tmp`(audit=2) → 后台审核 #123 → 通过同步 `ms_jobs`(audit=1) 单事务；不通过 tmp audit=3 + `ms_audit_reason`
- `display=2`（直接显示，需显式配置）：新职位/编辑直接写 `ms_jobs`（audit=1 可见）
- **编辑（display=1，M9）**：编辑后原 `ms_jobs` 行 audit=2（前台不可见「审核中」），新内容写 tmp；审核通过覆盖原行（保留 id/click）；不通过恢复旧版 audit=1 + reason
- **D1（2026-08-21）：tmp-only 职位的编辑/删除**：display=1 默认下，新职位审核通过前只存在于 `ms_jobs_tmp`（tmp-only）。#82 编辑重提：`pending=1`（列表项 pending 透传）→ tmp 行原地更新（JobsBase 全量 + contact/tags 重挂 tmp.id + audit 置回 2，仅 audit∈{2,3} 可编辑）；不传 pending 且 jobs 表未命中 → 自动回退同一 tmp 路径（两表 id 各自自增会重叠，jobs 行存在时以 jobs 行为准，与 #81 语义一致）。#83 删除：`pending=1` → 软删 tmp 行；jobs 未命中 → 回退软删 tmp 行（contact/tag 随 pid 保留，职位已不可见）。#80 列表对 tmp 行同样过滤 `deleted_at=0`

## 发布校验（03 §2.3）

- 企业资质 `ms_company_profile.audit=1` 才可发布（未过审 → 1001「请先完成企业认证」，待审 → 1001「认证审核中」）
- `jobsName` 2-50 字符；`contents` ≤4000；`topclass/category/subclass` 三级必填；`amount` 0-99
- 薪资：`negotiable=0` 时 `minwage/maxwage` 0-999999 且 max>min、min>0 时 max/min≤2

## 职位字段（小驼峰，01 §2.2 裁剪后）

`jobsName / nature / natureCn / sex / amount / topclass / category / subclass / categoryCn / trade / district / districtCn / tag / education / experience / minwage / maxwage / negotiable / contents / deadline / department / mapX / mapY / mapZoom`

> sex/education/experience/trade 为编码，前端按分类数据映射中文；nature_cn/category_cn/district_cn 为冗余中文（前端随表单传入）。
> 联系方式在 `contact` 子对象（`contact/qq/telephone/landlineTel/address/email`）；标签在 `tags`（uint32 数组）。

## 发布请求示例

```json
{
  "jobsName": "Go 后端工程师", "nature": 1, "natureCn": "全职", "sex": 3, "amount": 2,
  "topclass": 1, "category": 2, "subclass": 3, "categoryCn": "技术/后端",
  "trade": 1, "district": "4401", "districtCn": "广州",
  "education": 4, "experience": 3, "minwage": 10000, "maxwage": 15000,
  "negotiable": 0, "contents": "负责后端服务开发", "deadline": 1756425600,
  "contact": { "contact": "李四", "telephone": "13800138000" },
  "tags": [1, 2]
}
```

## 列表/详情返回

- #80 列表项：`id / jobsId` + 职位字段 + `pending`（待审 true）；`jobs` 与 `jobs_tmp` 合并，按 id desc；**编辑中 jobs 行由 tmp 行代表，不重复出现**
- #81 详情：职位字段 + `contact`（联系方式对象）+ `tags`（数组）+ `reason`（不通过原因）+ `pending`；**必须带 `pending` 参数**（两表 id 会重叠）：列表项 `pending=true` → `GET /company/jobs/{id}?pending=1`（查 jobs_tmp），`pending=false` → 不带 pending（查 jobs）

## 前台搜索（03 §2.5）

**#22 GET /api/v1/jobs** —— Query：`keyword`(职位名/企业名 LIKE)、`trade`、`category`(二级)、`district`、`education`、`experience`、`minwage`、`maxwage`、`order`(last 默认/addtime/salary/stick)、`page`、`pageSize`
- 仅返回 `display=1 & audit=1 & 未删除 & 未过期(deadline=0 或 >now)` 的职位

**#23 GET /api/v1/jobs/{id}** —— 职位详情（click 自增），返回职位字段 + `contact` + `tags`

**#24 GET /api/v1/jobs/hot-words** —— 热门词（一期读配置 `mscms_hot_words` 逗号分隔，Redis Top50 二期）

**#25 GET /api/v1/jobs/filters** —— 筛选元数据（分类聚合 `{alias: [category...]}`，复用 #21；含 `jobcategory` 职位三级分类分组（parent_id 层级）+ `jobtitle` 职位名称扁平列表（无层级，修改批 X3 新增））

## 职位三级分类数据源

发布表单「三级分类级联」数据源为 `#21/#25` 的 `jobcategory` 分组（三级层级 parent_id）：一级 topclass → 二级 category → 三级 subclass；已 seed **10 个一级行业**（技术/产品/设计/运营/市场/销售/职能/电商/金融/教育培训，X3 由 7 扩到 10 并补二/三级骨架），`categoryCn` 由前端按选中的三级分类拼接后随表单传入。

## 职位名称数据源（修改批 X3）

发布表单「职位名称」由自由输入改为**列表选择**（el-select filterable，**不允许自由输入**）：数据源为 `#21/#25` 的 `jobtitle` 分组（扁平列表 50 个常用岗位，无层级）；`jobs_name` 仍存所选名称字符串（varchar(50)，#79 契约不变）。编辑回显：已存名字若不在列表内需显示该名字（前端加虚拟 option，不丢数据）。「后期往上添加」= 插 `ms_category` 行（group=jobtitle）或 M5 后台 CRUD。

## 一期未实现（12 §四 后置）

- 套餐 `jobs_meanwhile` 同时在线数约束（M4 ms_setmeal 未建）
- 敏感词过滤（M5）
- 置顶/紧急推广购买（M4 走订单）
