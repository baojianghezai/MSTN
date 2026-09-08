# 06 统计看板与报表（dashboard / statistics）

> 对应设计：`08_后台RBAC系统统计模块设计.md` §4（A6 统计看板）、`11_v6差异分析与改进建议.md` §四 P0#4（看板扩展为独立统计报表）、`09_API接口规范.md` §4.6.2（#119/#120）
> 实现状态：4 个后台接口（看板/趋势/求职分布/企业分布）
> 说明：前缀 `/api/v1`；`admin/**` 走 GVA admin JWT（`x-token`）。看板已聚合会员、简历、企业、职位、投递、视频面试、申诉与注销；收入仍等待订单金额口径收口。

## 接口清单

| # | 方法 | 路径 | 请求类型 | 鉴权 | 说明 |
|---|---|---|---|---|---|
| 119 | GET | /api/v1/admin/dashboard | Query | admin | 看板指标（今日/昨日/待办/收入） |
| 120 | GET | /api/v1/admin/dashboard/trend | Query | admin | 趋势图数据 {days, metric} |
| — | GET | /api/v1/admin/statistics/resume | 无参 | admin | 求职者分布（性别/学历/经验） |
| — | GET | /api/v1/admin/statistics/company | 无参 | admin | 企业分布（性质/规模） |

> 后两个统计报表接口编号待 PM 回填 design/09。

---

## 119. GET /api/v1/admin/dashboard（看板指标）

- **请求类型**：无参
- 响应：

```json
{ "code": 0, "message": "success",
  "data": {
    "today":     { "personalUsers": 2, "companyUsers": 1, "resumes": 1, "companies": 1, "jobs": 3, "applications": 5, "videoInterviews": 1 },
    "yesterday": { "personalUsers": 1, "companyUsers": 0, "resumes": 0, "companies": 0, "jobs": 1, "applications": 2, "videoInterviews": 0 },
    "todo":      { "companyAudit": 1, "resumeAudit": 0, "jobAudit": 2, "appeal": 1, "companyCancellation": 1 },
    "income":    { "today": 0, "month": 0 }
  } }
```

- `jobs` 统计 `ms_jobs.addtime` 且排除逻辑删除职位；`applications` 统计 `ms_personal_jobs_apply.apply_addtime`。
- `jobAudit` 统计 `ms_jobs_tmp.audit=2` 且排除逻辑删除草稿，后台可跳转至“待审职位”处理。
- `income` 当前恒 0，待订单金额口径确认后再接入。

## 120. GET /api/v1/admin/dashboard/trend（趋势图）

- Query：`days`（默认 30，上限 90）、`metric`（`register`/`resume`/`company`/`job`/`application`）
- 响应：`[]`，每项 `{ date, personal, company, count }`；`register` 用 `personal`/`company`，其余指标用 `count`

```json
{ "code": 0, "data": [
  { "date": "2026-08-17", "personal": 2, "company": 1, "count": 0 },
  { "date": "2026-08-16", "personal": 0, "company": 0, "count": 0 } ] }
```

## GET /api/v1/admin/statistics/resume（求职者分布）

- 响应：`{ sex: [...], education: [...], experience: [...] }`，每项 `{ code, cn, count }`（`code` 分类值、`cn` 中文、`count` 数量，按 code 升序，排除 code=0）

## GET /api/v1/admin/statistics/company（企业分布）

- 响应：`{ nature: [...], scale: [...] }`，每项 `{ code, cn, count }`

---

## 待后续补充

- 岗位分析（职位学历/经验要求分布）。
- 下载、刷新等行为指标，以及待审简历图统计。
- 收入（今日/本月订单金额）的确认支付口径。
