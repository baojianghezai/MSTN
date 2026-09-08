# 10 投递（apply）

> 对应设计：`05_投递下载面试模块设计.md` §2.1（投递规则）、`12_MVP可跑通链路.md` 模块 C、`01_数据库设计.md` §2.6、`09_API接口规范.md` §4.2(61-63)
> 实现状态：3 个个人端接口
> 说明：前缀 `/api/v1`；`personal/**` 需 utype=1（个人会员 JWT）。

## 接口清单

| # | 方法 | 路径 | 请求类型 | 说明 |
|---|---|---|---|---|
| 61 | POST | /api/v1/personal/applies | JSON | 投递职位 |
| 62 | GET | /api/v1/personal/applies | Query | 我的投递列表（状态筛选） |
| 63 | DELETE | /api/v1/personal/applies/{did} | 无参 | 删除投递记录 |

---

## 投递规则（05 §2.1）

- 每日上限：配置 `mscms_apply_jobs_max`（默认 10）
- 简历选择：`resumeId` 指定则校验归属+通过+未软删（不存在/已删/未过审 → 1001「简历不存在或已删除」）；未指定取默认简历（def desc 第一条，无 → 1001「请先填写简历」）
- 完善度门槛：所选简历 `completePercent` 须达到 40%，否则 → 1001「简历完善度不足，请先完善至 40% 再投递」
- 去重：**同一企业对同一份简历仅投一次**（`uk_uid_resume_company(personal_uid, resume_id, company_uid)` 唯一索引兜底）
- 职位过期/下架（display≠1 或 audit≠1 或已删或过期）→ 1001「职位不存在或已下架」

## 61. POST /api/v1/personal/applies（投递）

```json
// 请求
{ "jobsIds": [1, 2], "resumeId": 12, "notes": "随时到岗" }
// 响应
{ "code": 0, "message": "success", "data": {} }
```

- 错误：1001（请选择职位 / 请先填写简历 / 简历完善度不足，请先完善至 40% 再投递 / 今日投递次数已用完 / 您已向该公司投递过简历 / 职位不存在或已下架）

## 62. GET /api/v1/personal/applies（我的投递）

- Query：`page`、`pageSize`、`status`（0 全部 / 1 未读 / 2 已读）
- 响应：分页 list 项 `{did, resumeId, resumeName, jobsId, jobsName, companyId, companyName, companyUid, applyAddtime, personalLook, notes, isReply, replyTime}`

## 63. DELETE /api/v1/personal/applies/{did}（删除）

- 仅本人；不存在 → 1001「投递记录不存在」

---

## 一期未实现（05 §2.1 / 12 §四 后置）

- 企业屏蔽名单（personal_shield_company）
- 短信/站内信通知企业（M5 消息模块）
