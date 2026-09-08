# 17 企业找人才与人才库

接口前缀：`/api/v1`。公开简历接口不要求登录；企业接口要求会员 JWT 且 `utype=2`：`Authorization: Bearer <access_token>`。

## 隐私规则

- 公开搜索、公开详情和企业收藏列表始终不返回电话、邮箱和附件简历地址。
- `displayName=2` 的匿名简历统一显示为“求职者”，不返回照片。
- 企业仅在成功解锁后，才能从人才库读取该候选人的完整简历、电话、邮箱和附件地址。
- 新解锁仅允许 `display=1`、`audit=1`、未删除的公开简历；历史已解锁简历在仍未删除时可继续从人才库查看。

## 公开简历

### GET /resumes

请求类型：Query。公开简历搜索。

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| page / pageSize | number | 否 | 默认 1 / 20，最大 100 |
| keyword | string | 否 | 职位、技能、简历标题关键字 |
| district | string | 否 | 期望地区编码 |
| education / experience / wage | number | 否 | 学历、经验、期望薪资编码 |

响应 `data`：`{ list: PublicResume[], total, page, pageSize }`，按高级人才、刷新时间、ID 排序。

### GET /resumes/talents

请求类型：Query。参数与 `/resumes` 相同，仅返回 `talent=1` 的高级人才。

### GET /resumes/{id}

请求类型：无参。返回公开脱敏简历主信息及教育、工作、项目、语言、培训、证书子表，不返回联系方式。

`PublicResume` 主要字段：`id`、`title`、`fullname`、`educationCn`、`experienceCn`、`districtCn`、`wageCn`、`intentionJobs`、`specialty`、`photoImg`、`talent`、`refreshtime`。

## 企业主动解锁与人才库

### POST /company/talents/{id}/unlock

请求类型：无参。首次解锁消耗当前有效套餐的一次简历下载权益，创建 `ms_resume_download` 记录并通知候选人；已解锁或已通过投递下载同一简历时不再扣费或通知。

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": { "detail": { "resume": {} }, "newlyUnlocked": true }
}
```

`detail.resume` 为完整简历，包含 `telephone`、`email` 和附件简历字段。无有效下载权益时返回业务失败。

### GET /company/talents

请求类型：Query。分页返回当前企业已解锁的人才库。

`list` 元素：`{ download, resume }`；`download` 包含 `resumeId`、`followUp`、`downloadedAt`，`resume` 包含完整联系方式。

### GET /company/talents/{id}

请求类型：无参。仅已解锁企业可访问，返回完整简历和全部子表。

### PUT /company/talents/{id}/follow-up

请求类型：JSON。

```json
{ "followUp": 3 }
```

跟进状态：0 待跟进，1 合适，2 不合适，3 待定，4 未接通。仅当前企业已解锁记录可修改。

## 企业收藏

### GET /company/favorites

请求类型：Query。分页返回当前企业的收藏，简历仍为公开脱敏字段。

### POST /company/favorites

请求类型：JSON。

```json
{ "resumeId": 101 }
```

收藏不消耗下载次数；同一企业不得重复收藏同一简历。

### DELETE /company/favorites/{id}

请求类型：无参。`id` 为收藏记录 ID，只能删除当前企业自己的收藏。

## 自动通知

企业首次从人才库成功解锁简历时，会在同一事务内向简历所属个人写入 `resume_download` 站内信。套餐扣次、解锁记录和通知任一失败都会整体回滚。
