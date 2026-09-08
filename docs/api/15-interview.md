# 15 标准线下面试邀请

接口前缀：`/api/v1`。所有接口均需要会员 JWT：`Authorization: Bearer <access_token>`。

线下面试邀请与保留中的视频面试为两条独立业务线。本模块不创建视频房间、不发送站内信或短信。

## 数据结构

邀请对象字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| did | number | 邀请记录 ID |
| resumeId / resumeName / resumeUid | number / string / number | 创建时的简历快照 |
| jobsId / jobsName | number / string | 创建时的职位快照 |
| companyId / companyName / companyUid | number / string / number | 创建时的企业快照 |
| interviewTime | number | 面试时间，Unix 秒 |
| address | string | 面试详细地址 |
| contact | string | 联系人 |
| telephone | string | 联系电话 |
| notes | string | 备注，可为空 |
| interviewAddtime | number | 邀请创建时间，Unix 秒 |
| personalLook | number | 1 未读，2 已读 |

## 企业端

企业端接口要求 `utype=2`。

### POST /company/interviews

请求类型：JSON。创建线下面试邀请。

```json
{
  "resumeId": 101,
  "jobsId": 20,
  "interviewTime": 1787652000,
  "address": "上海市浦东新区示例路 1 号",
  "contact": "王经理",
  "telephone": "13800138000",
  "notes": "请携带作品集"
}
```

- `interviewTime` 必须晚于当前时间。
- `address`、`contact`、`telephone` 必填，最大长度分别为 200、30、30；`notes` 最大 500。
- 简历必须公开、审核通过且未删除；职位必须归属于当前企业且未删除。
- 相同 `companyUid + resumeId + jobsId` 只能保留一条邀请。

成功响应 `data` 为完整邀请对象。

### GET /company/interviews

请求类型：Query。查询当前企业发出的邀请。

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| page | number | 否 | 页码，默认 1 |
| pageSize | number | 否 | 每页数量，默认 20，最大 100 |

成功响应：

```json
{
  "code": 0,
  "message": "success",
  "data": { "list": [], "total": 0, "page": 1, "pageSize": 20 }
}
```

`list` 元素为完整邀请对象，按 `did` 倒序。

### DELETE /company/interviews/{did}

请求类型：无参。撤回当前企业发出的邀请。

- `did` 为邀请记录 ID。
- 记录不存在或不属于当前企业时返回业务失败响应。

成功响应 `data` 为空对象。

## 个人端

个人端接口要求 `utype=1`。

### GET /personal/interviews

请求类型：Query。查询当前个人收到的邀请。

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| page | number | 否 | 页码，默认 1 |
| pageSize | number | 否 | 每页数量，默认 20，最大 100 |

成功响应分页结构同企业列表。`list` 按未读优先、面试时间升序排列。

### PUT /personal/interviews/{did}/read

请求类型：无参。将当前个人的邀请标记为已读。

- `did` 为邀请记录 ID。
- 记录不存在或不属于当前个人时返回业务失败响应。

成功响应 `data` 为空对象。

## 业务失败

所有参数校验、资源不可用、重复邀请及越权操作都返回 HTTP 200、`code=1001`，并在 `message` 给出原因。认证失败由中间件返回 HTTP 401/403。
