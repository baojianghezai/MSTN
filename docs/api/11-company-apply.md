# 11 企业收简历（company-apply）

> 对应设计：`05_投递下载面试模块设计.md` §2.1 企业侧操作、`12_MVP可跑通链路.md` 模块 D、`09_API接口规范.md` §4.3(90/92-94)
> 实现状态：4 个企业端接口
> 说明：前缀 `/api/v1`；`company/**` 需 utype=2（企业会员 JWT）。

## 接口清单

| # | 方法 | 路径 | 请求类型 | 说明 |
|---|---|---|---|---|
| 92 | GET | /api/v1/company/applies | Query | 收简历列表（脱敏快照） |
| 90 | GET | /api/v1/company/jobs/{id}/applies | Query | 单职位收到的投递 |
| 93 | PUT | /api/v1/company/applies/{did}/looked | 无参 | 标记已看 |
| 94 | PUT | /api/v1/company/applies/{did}/reply | JSON | 回复状态 |
| 95 | GET | /api/v1/company/applies/{did}/resume/download | 无参 | 下载完整简历 |

---

## 92. GET /api/v1/company/applies（收简历列表）

- Query：`page`、`pageSize`、`status`（0 全部 / 1 未读 / 2 已读）
- 响应：分页 list 项 `{did, resumeId, resumeName, jobsId, jobsName, companyId, companyName, applyAddtime, personalLook, notes, isReply, replyTime}`
- 脱敏：快照字段（简历名/职位名），联系方式不在投递记录里，查看完整简历走下载（M4）

## 90. GET /api/v1/company/jobs/{id}/applies（单职位投递）

- 同 #92，按 `jobs_id` 过滤

## 93. PUT /api/v1/company/applies/{did}/looked（标记已看）

- 将 `personal_look` 置 2（已读）；不存在/非本企业 → 1001「投递记录不存在」

## 94. PUT /api/v1/company/applies/{did}/reply（回复状态）

```json
{ "isReply": 1 }
```

- `isReply`：0 待反馈 / 1 合适 / 2 不合适 / 3 待定 / 4 未接通；非法（<0 或 >4）→ 1001「回复状态非法（0-4）」
- 同时写 `reply_time`

---

## 95. GET /api/v1/company/applies/{did}/resume/download（下载完整简历）

- 鉴权：企业会员，仅能下载本企业收到的投递；非归属或投递不存在返回 `5004`。
- 响应：HTML 文件流，`Content-Type: text/html; charset=utf-8`，文件名 `resume-{resumeId}.html`；包含简历主表、教育/工作/项目/语言/培训/证书经历及完整联系方式。
- 首次下载：要求存在未过期套餐且剩余下载次数大于 0；事务内创建下载记录、`resume_downloads_used + 1` 并标记投递为已读。
- 重复下载同一份简历：不再次扣减额度。
- 错误：无有效下载权益或次数用尽返回 `5003`；简历已删除返回 `1001`。

---

## 已完成的后置项

- 简历下载扣费（#95，套餐联动）
- 简历完整联系方式可见性（下载文件内提供）
