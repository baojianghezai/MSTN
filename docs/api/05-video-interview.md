# 05 视频面试（video-interview）

> 对应设计：`11_v6差异分析与改进建议.md` §四 P0#2（v6 `qs_video_interview` 平移，腾讯云 TRTC）；`09_API接口规范.md` §4.2/§4.3/§4.6.2（接口编号待 PM 回填，本文件为临时契约）
> 实现状态：7 个会员/公开接口 + 1 个后台接口；配置复用 #150（group=video）
> 说明：前缀 `/api/v1`；`company/**` 需 utype=2，`personal/**` 需 utype=1；房间码查询公开（房间码即入房凭证）。

## 接口清单

| 方法 | 路径 | 请求类型 | 鉴权 | 说明 |
|---|---|---|---|---|
| POST | /api/v1/company/video-interviews | JSON | 企业 | 发起视频面试邀请 |
| GET | /api/v1/company/video-interviews | Query | 企业 | 我发出的视频面试列表 |
| GET | /api/v1/company/video-interviews/{id} | 无参 | 企业 | 视频面试详情 |
| DELETE | /api/v1/company/video-interviews/{id} | 无参 | 企业 | 删除视频面试邀请 |
| GET | /api/v1/personal/video-interviews | Query | 个人 | 我收到的视频面试列表 |
| GET | /api/v1/personal/video-interviews/{id} | 无参 | 个人 | 视频面试详情 |
| GET | /api/v1/video-interviews/room/{code} | 无参 | 无 | 房间码查询（TRTC 入房） |
| GET | /api/v1/admin/video-interviews | Query | admin | 视频面试列表（后台） |

---

## 数据模型（ms_video_interview，v6 qs_video_interview 平移）

| 字段 | 类型 | 说明 |
|---|---|---|
| company_uid | uint64 | 发起企业 uid |
| personal_uid | uint64 | 被邀个人 uid（简历所属） |
| jobs_id | uint64 | 关联职位 id |
| jobs_name | string(30) | 职位名快照 |
| interview_time | int64 | 面试时间（unix 秒） |
| deadline | int64 | 房间有效期（面试时间 +15 天） |
| contact / contact_tel | string(30) | 联系人 / 联系电话 |
| company_code / personal_code | string(6) | 企业端 / 个人端房间码（6 位字母数字） |

**房间状态机**（不落库，读取时按时间计算）：`nostart`（面试日未到）/ `opened`（面试日当天）/ `overtime`（deadline 已过）。

---

## POST /api/v1/company/video-interviews（发起邀请）

- **请求类型**：JSON；**鉴权**：企业（utype=2）
- 请求体：

```json
{
  "resumeId": 12, "jobsId": 100, "jobsName": "Go 开发工程师",
  "interviewTime": 1755129600, "contact": "李四", "telephone": "13800138000"
}
```

- 响应：`data` 为视频面试项（含 `companyCode`/`personalCode`/`roomStatus`）
- 校验：`video_interview_open=1`（配置）；简历存在；同一企业+同一简历+同一职位且未过期 → 拒绝重复
- 错误码：1001（视频面试未开启 / 简历不存在 / 已对该简历进行过面试邀请，不能重复邀请）

## GET /api/v1/company/video-interviews（企业列表）

- 分页 `{list,total,page,pageSize}`；list 项含 `fullname`（简历姓名）+ `roomStatus`

## GET /api/v1/personal/video-interviews（个人列表）

- 分页；list 项含 `jobsName` + `roomStatus`

## GET /api/v1/video-interviews/room/{code}（房间码查询，公开）

- 通过 `company_code`（→ utype=2）或 `personal_code`（→ utype=1）查面试
- 响应（不含联系方式）：

```json
{ "code": 0, "message": "success",
  "data": { "id": 12, "jobsName": "Go 开发工程师", "interviewTime": 1755129600,
            "deadline": 1756425600, "roomStatus": "opened", "utype": 1 } }
```

## GET /api/v1/admin/video-interviews（后台列表）

- Query：`page`、`pageSize`、`keyword`（跨职位名/公司名/简历姓名模糊搜索）
- list 项含 `fullname`（简历姓名）+ `companyname`（企业名）+ `roomStatus`

## 配置（复用 #150 GET/PUT /api/v1/admin/configs，group=video）

| name | 说明 |
|---|---|
| video_interview_open | 是否开启视频面试（"1" 开启，默认关闭） |
| trtc_appid | 腾讯云 TRTC AppID |
| trtc_appsecret | 腾讯云 TRTC AppSecret |

---

## 一期未实现（依赖 M3/M4 表，待补）

- 套餐权益校验：`ms_setmeal.enable_video`（M4 建套餐表后补）
- 职位审核校验：企业需有已过审职位（`ms_jobs`，M3）
- 下载/投递资格校验：需 `ms_company_down_resume`/`ms_personal_jobs_apply`（05 投递/下载，M3）
- 短信/站内信通知（M3/M6 消息模块）
