# 16 站内信中心

接口前缀：`/api/v1`。所有接口均需要会员 JWT：`Authorization: Bearer <access_token>`。

个人端要求 `utype=1`，企业端要求 `utype=2`。两端使用同一套契约，且只能读取、标记或删除自己的消息。

## 消息对象

| 字段 | 类型 | 说明 |
|---|---|---|
| id | number | 消息 ID |
| msgFrom | number | 发送会员 UID；系统消息可为 0 |
| msgTouid | number | 接收会员 UID |
| title | string | 标题 |
| message | string | 纯文本正文，前端不得按 HTML 渲染 |
| type | string | `application`、`resume_download`、`interview` 等业务分类 |
| link | string | 站内跳转路径，可为空 |
| msgCheck | number | 0 未读，1 已读 |
| addtime | number | 创建时间，Unix 秒 |

## 接口

下表中的 `{scope}` 需要替换为 `personal` 或 `company`，并使用与端类型匹配的会员登录态。

| 方法 | 路径 | 请求类型 | 说明 |
|---|---|---|---|
| GET | `/{scope}/messages` | Query | 分页查询自己的站内信 |
| PUT | `/{scope}/messages/read` | JSON | 批量标记自己的消息已读 |
| DELETE | `/{scope}/messages` | JSON | 批量删除自己的消息 |
| GET | `/{scope}/messages/unread-count` | 无参 | 查询未读总数 |

### GET /{scope}/messages

请求类型：Query。

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

列表按 `addtime`、`id` 倒序排列。

### PUT /{scope}/messages/read

请求类型：JSON。

```json
{ "ids": [101, 102] }
```

`ids` 必须包含 1 至 100 个消息 ID。属于其他会员或不存在的 ID 不会被修改。

### DELETE /{scope}/messages

请求类型：JSON。

```json
{ "ids": [101, 102] }
```

`ids` 必须包含 1 至 100 个消息 ID。删除为物理删除，且只会删除当前会员所属的消息。

### GET /{scope}/messages/unread-count

请求类型：无参。

成功响应：

```json
{ "code": 0, "message": "success", "data": { "unread": 3 } }
```

## 自动通知

本期使用统一事务内通知入口；触发业务回滚时消息也会回滚。

| 事件 | 接收方 | type | 规则 |
|---|---|---|---|
| 个人投递职位 | 职位所属企业 | application | 每条成功投递生成一条消息 |
| 企业下载投递简历 | 简历所属个人 | resume_download | 仅首次成功解锁并扣减下载权益时生成；重复下载不重复通知 |
| 企业发起线下面试 | 被邀请个人 | interview | 线下面试邀请创建成功后生成 |

`ms_pms.msg_check` 是未读状态的真源；`ms_members_msgtip` 的 `pms` 项作为冗余计数同步维护。
