# 14 后台数据清理（data-cleanup）

后台数据清理工具只开放给已登录且拥有 Casbin 权限的 GVA 管理员。它不接受表名、SQL 或自定义过滤条件，只允许执行下列固定目标；所有执行均写入 `ms_data_cleanup_log`。

请求头：`x-token: <GVA admin JWT>`。

| 方法 | 路径 | 请求类型 | 说明 |
|---|---|---|---|
| POST | `/api/v1/admin/data-cleanup/preview` | JSON | 预览清理范围，不修改数据 |
| POST | `/api/v1/admin/data-cleanup/execute` | JSON | 以确认词执行当前范围清理 |
| GET | `/api/v1/admin/data-cleanup/history` | Query | 查询已完成的清理审计记录 |

## 固定清理目标

| target | 行为 | retentionDays |
|---|---|---|
| `expired_jobs` | `deadline` 早于当前时间且未删除的职位写入 `deleted_at`，不删除投递 | 不传/0 |
| `rejected_job_drafts` | 物理删除 `ms_jobs_tmp.audit=3` 的职位草稿 | 不传/0 |
| `expired_promotions` | 删除关联职位下线、过期、未通过审核，或没有有效套餐权益的首页推广 | 不传/0 |
| `old_payment_notify_logs` | 物理删除截止时间之前的支付通知日志 | 7 至 3650 |
| `old_wxpay_logs` | 物理删除截止时间之前的微信支付回调日志 | 7 至 3650 |

不会清理会员、企业资料、简历、投递、订单、套餐权益、支付订单或上传文件。职位草稿的联系人和标签使用共享 `pid`，因此本工具不会级联删除这些记录。

## POST /api/v1/admin/data-cleanup/preview

请求体：

```json
{ "target": "old_wxpay_logs", "retentionDays": 30 }
```

响应 `data`：

```json
{
  "target": "old_wxpay_logs",
  "title": "历史微信支付日志",
  "description": "物理删除保留期限之前的微信支付回调日志，不影响订单和套餐权益。",
  "affectedCount": 18,
  "retentionDays": 30,
  "cutoffAt": 1784976000,
  "softDelete": false
}
```

- `affectedCount` 是调用时的预估数量；执行时按当前数据重新计算。
- `cutoffAt` 是 Unix 秒。非日志目标为当前执行时刻，日志目标为当前时刻减去 `retentionDays`。
- `softDelete=true` 仅出现在 `expired_jobs`。

## POST /api/v1/admin/data-cleanup/execute

请求体在预览参数上增加确认词：

```json
{ "target": "old_wxpay_logs", "retentionDays": 30, "confirmation": "CONFIRM" }
```

- `confirmation` 必须精确为 `CONFIRM`。
- 成功响应与 preview 相同，但 `affectedCount` 是事务内实际处理数量。
- 清理动作和审计记录在同一事务中完成；审计包含目标、保留天数、截止时间、影响数量、后台操作人和执行时间。

## GET /api/v1/admin/data-cleanup/history

Query：`page`、`pageSize`。

响应为分页结构，`list` 项字段：`id`、`target`、`retentionDays`、`cutoffAt`、`affectedCount`、`operatorId`、`operatorName`、`createdAt`。
