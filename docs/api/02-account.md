# 02 账号资料与申诉（profile / appeal）

> 对应设计：`02_用户账号模块设计.md` §2.2/§2.5/§2.6/§2.8/§3.1、`09_API接口规范.md` §4.1(45/46)、§4.2(76/77)、§4.3(107-110/110a/110b)、§4.6.2(134/134a/135/158/159/161a-c)
> 实现状态：9 个会员/公开接口 + 8 个后台接口
> 说明：前缀 `/api/v1`；`personal/**` 需 utype=1，`company/**` 需 utype=2，均走会员 JWT；申诉为公开接口（已登录自动带 uid）；`admin/**` 走 GVA admin JWT。

## 接口清单

| # | 方法 | 路径 | 请求类型 | 鉴权 | 说明 |
|---|---|---|---|---|---|
| 45 | POST | /api/v1/appeal | JSON | 无（已登录自动带 uid） | 提交账号申诉 |
| 46 | GET | /api/v1/appeal/status | Query | 无 | 申诉进度查询 |
| 76 | GET | /api/v1/personal/profile | 无参 | 个人 | 个人资料读取 |
| 77 | POST | /api/v1/personal/profile | JSON | 个人 | 个人资料维护 |
| 107 | GET | /api/v1/company/profile | 无参 | 企业 | 企业资料读取 |
| 108 | POST | /api/v1/company/profile | JSON | 企业 | 企业资料维护（提交即审核） |
| 109 | POST | /api/v1/company/profile/logo | Form-data | 企业 | 上传 Logo/证照 |
| 110 | GET | /api/v1/company/profile/audit-status | 无参 | 企业 | 企业审核状态 |
| 110a | POST | /api/v1/company/cancel | JSON | 企业 | 企业注销申请（短信二次确认） |
| 110b | GET | /api/v1/company/cancel | 无参 | 企业 | 查询企业注销申请状态 |
| 134 | GET | /api/v1/admin/company-profiles | Query | admin | 企业资料列表（审核） |
| 134a | GET | /api/v1/admin/company-profiles/{id} | 无参 | admin | 企业资料详情（含 Logo/证照） |
| 135 | PUT | /api/v1/admin/company-profiles/{id}/audit | JSON | admin | 企业资质审核（通过/不通过+原因） |
| 158 | GET | /api/v1/admin/appeals | Query | admin | 申诉列表 |
| 159 | PUT | /api/v1/admin/appeals/{id} | JSON | admin | 处理申诉（可恢复账号） |
| 161a | GET | /api/v1/admin/company/cancellations | Query | admin | 企业注销申请列表 |
| 161b | POST | /api/v1/admin/company/cancellations/{id}/handle | 无参 | admin | 处理企业注销（清企业业务数据） |
| 161c | DELETE | /api/v1/admin/company/cancellations/{id} | 无参 | admin | 硬删除注销申请记录 |

---

## 45. POST /api/v1/appeal（提交申诉）

- **请求类型**：JSON；**鉴权**：无（带 `Authorization: Bearer <token>` 时自动带 uid）
- 字段：`realname`(必填)、`mobile`(必填，须合法手机号)、`email`(选填)、`description`(必填)
- 限流：同手机号 **3 次/小时**（Redis `hrc:appeal:rate:<mobile>`）
- 响应成功：`{ "id": 申诉ID }`
- 错误码：1001（参数不完整 / 手机号格式错 / 过于频繁）

```json
// 请求
{ "realname": "张三", "mobile": "13800138000", "email": "z@a.com", "description": "收不到验证码" }
// 响应
{ "code": 0, "message": "success", "data": { "id": 12 } }
```

---

## 46. GET /api/v1/appeal/status（申诉进度）

- **请求类型**：Query；**鉴权**：无
- 参数：`mobile`(必填)
- 响应：按手机号返回最近 10 条申诉记录（倒序），`statusCn`：0 待处理 / 1 已处理 / 2 已驳回

```json
{ "code": 0, "message": "success",
  "data": [ { "id": 12, "uid": 0, "realname": "张三", "mobile": "13800138000",
              "status": 0, "statusCn": "待处理", "addtime": 1755000000 } ] }
```

---

## 76. GET /api/v1/personal/profile（个人资料）

- **请求类型**：无参；**鉴权**：个人（utype=1）
- 响应：`ms_members_info` 全字段（注册时已建空壳，恒有记录）

| 字段 | 类型 | 说明 |
|---|---|---|
| realname | string | 真实姓名 |
| sex / sexCn | int8 / string | 性别 code + 中文 |
| birthday | int64 | 生日（unix 秒） |
| residence | string | 现居地 |
| education / educationCn | uint16 / string | 学历 |
| major / majorCn | uint16 / string | 专业 |
| experience / experienceCn | uint16 / string | 工作年限 |
| phone | string | 联系电话 |
| height | string | 身高 |
| marriage / marriageCn | int8 / string | 婚姻 |
| displayName | int8 | 姓名显示方式 |
| qq / weixin | string | 联系方式 |

> 枚举下拉数据源：`sex`/`marriage`/`education`/`experience`/`major` 的选项来自 #21 分类聚合（`GET /api/v1/categories`），对应分组 `sex`/`marriage`/`education`/`experience`/`major`。提交时 `xxx` 传分类 `id`、`xxxCn` 传分类 `name`（前端选中后随 code 一起提交）。

---

## 77. POST /api/v1/personal/profile（个人资料维护）

- **请求类型**：JSON；**鉴权**：个人（utype=1）
- 请求体：同 76 响应字段（除 `id`/`uid`），整包覆盖写（缺省字段置零/空）
- 响应：`code=0`

---

## 107. GET /api/v1/company/profile（企业资料）

- **请求类型**：无参；**鉴权**：企业（utype=2）
- 响应：`ms_company_profile` 全字段（注册时已建空壳，`audit=0`）

| 字段 | 类型 | 说明 |
|---|---|---|
| companyname | string | 企业名（唯一） |
| nature / natureCn | uint16 / string | 企业性质 |
| trade / tradeCn | uint16 / string | 行业 |
| district / districtCn | string | 地区 |
| scale / scaleCn | uint16 / string | 规模 |
| registered | string | 注册资金 |
| address | string | 地址 |
| contact / telephone / landlineTel | string | 联系人/电话/座机 |
| email / website | string | 邮箱/官网 |
| certificateImg | string | 营业执照 URL |
| logo | string | Logo URL |
| contents | string | 企业介绍 |
| shortName / shortDesc / tag | string | 简称/简介/标签 |
| audit | int8 | 0 未提交 1 通过 2 审核中 3 不通过 |

> 枚举下拉数据源：`nature`/`scale`/`trade`/`district` 的选项来自 #21 分类聚合（`GET /api/v1/categories`），对应分组 `nature`/`scale`/`trade`/`district`。提交时 `xxx` 传分类 `id`、`xxxCn` 传分类 `name`（前端选中后随 code 一起提交）。

---

## 108. POST /api/v1/company/profile（企业资料维护）

- **请求类型**：JSON；**鉴权**：企业（utype=2）
- 请求体：107 响应中可编辑字段（除 `id`/`uid`/`setmealId`/`setmealName`/`audit`/`addtime`/`refreshtime`/`click`/`userStatus`）
- **审核触发**：每次提交将 `audit` 置为 **2（审核中）**，等待后台资质审核（#135）
- 企业名查重：与其他企业重名返回 1001「企业名称已被占用」；空名称返回 1001「企业名称不能为空」
- 响应：`code=0`

---

## 109. POST /api/v1/company/profile/logo（上传 Logo/证照）

- **请求类型**：Form-data；**鉴权**：企业（utype=2）
- 表单字段：`file`（type=file，图片）
- 响应：`{ "url": "uploads/file/xxx.png" }`（前端将其写入 108 的 `logo` 或 `certificateImg`）

---

## 110. GET /api/v1/company/profile/audit-status（企业审核状态）

- **请求类型**：无参；**鉴权**：企业（utype=2）
- 响应：`{ "audit": 2, "auditCn": "审核中" }`；`auditCn`：0 未提交 / 1 已通过 / 2 审核中 / 3 未通过

---

## 134. GET /api/v1/admin/company-profiles（企业资料列表，审核）

- **请求类型**：Query；**鉴权**：admin（GVA 后台 JWT，Header `x-token`）
- 参数：`page`(默认1)、`pageSize`(默认10，最大100)、`audit`(缺省=全部；0=草稿 1=通过 2=待审 3=不通过)、`keyword`(企业名关键字)
- 响应：分页 `{list, total, page, pageSize}`；list 项：`{id, uid, companyname, logo, audit, auditCn, addtime, refreshtime}`（companyname 空指针转空串）

## 134a. GET /api/v1/admin/company-profiles/{id}（企业资料详情，审核）

- **请求类型**：无参；**鉴权**：admin
- 响应：`ms_company_profile` 全字段 + `auditCn`（含 `logo`、`certificateImg` 营业执照，审核预览用）
- 错误：1001（企业资料不存在）

## 135. PUT /api/v1/admin/company-profiles/{id}/audit（企业资质审核）

- **请求类型**：JSON；**鉴权**：admin
- 请求体：`{ "audit": 1|3, "reason": "营业执照不清晰" }`
  - `audit=1` 通过、`audit=3` 不通过（非法值拒绝）
  - `audit=3` 时 `reason` 写入 `ms_audit_reason`（Type=企业资质，供企业端回显）
- 响应：`code=0`；错误：1001（企业资料不存在 / 审核状态非法）
- 说明：审核结果企业端 #110 同步可见；`audit=1` 后企业方可发布职位（03 设计 S10，职位侧校验随 M3 实现）

---

## 158. GET /api/v1/admin/appeals（申诉列表）

- **请求类型**：Query；**鉴权**：admin（GVA 后台 JWT，Header `x-token`）
- 参数：`page`(默认1)、`pageSize`(默认10，最大100)、`status`(0全部/1已处理/2已驳回)、`mobile`
- 响应：分页 `{list, total, page, pageSize}`，list 项含 `statusCn`（待处理/已处理/已驳回）

## 159. PUT /api/v1/admin/appeals/{id}（处理申诉）

- **请求类型**：JSON；**鉴权**：admin
- 请求体：`{ "status": 1|2, "restore": true|false }`
  - `status=1` 已处理、`status=2` 已驳回
  - `restore=true`（且 `status=1`）时，按申诉的 `mobile` 匹配 `status=3` 的注销账号并恢复（`status=1`、`deleted_at=0`）
- 响应：`code=0`；错误：1001（申诉不存在 / 处理状态非法 / 无冷静期账号）
- 说明：注销恢复已从公开的 `/auth/cancel/restore` 迁到此处，用户经申诉→客服处理闭环恢复账号。恢复后密码仍为空（注销时清空），短信登录返回 `passwordSet=false`，前端引导走 #13 忘记密码重置设密

---

## 110a. POST /api/v1/company/cancel（企业注销申请）

- **请求类型**：JSON；**鉴权**：企业（utype=2）
- 请求体：`{ "code": "123456" }`（短信验证码，**type=cancellation**）
- 前置条件：手机号已注册、企业资料已填写（企业名非空）；已有 **status=0 未处理申请**时拒绝重复提交
- 响应：`code=0`；错误：1001（验证码错误 / 未完善企业资料 / 已有待处理申请 / 未绑定手机号）
- 说明：企业注销=清企业业务数据、**会员账号保留可复用**（区别于个人注销 #18 的两阶段匿名化）

## 110b. GET /api/v1/company/cancel（查询注销申请状态）

- **请求类型**：无参；**鉴权**：企业（utype=2）
- 响应：最近一条申请 `{id, companyname, addtime, status, statusCn, finishtime}`；`statusCn`：0 待处理 / 1 已处理

```json
{ "code": 0, "message": "success",
  "data": { "id": 3, "companyname": "名硕科技", "addtime": 1755000000,
            "status": 0, "statusCn": "待处理", "finishtime": 0 } }
```

## 161a. GET /api/v1/admin/company/cancellations（企业注销申请列表）

- **请求类型**：Query；**鉴权**：admin（GVA 后台 JWT，Header `x-token`）
- 参数：`page`(默认1)、`pageSize`(默认10，最大100)、`status`(0全部/1待处理/2已处理)
- 响应：分页 `{list, total, page, pageSize}`，list 项含 `companyname/username/mobile/status/statusCn/finishtime`（username/mobile 来自 ms_members）

## 161b. POST /api/v1/admin/company/cancellations/{id}/handle（处理企业注销）

- **请求类型**：无参；**鉴权**：admin
- 响应：`code=0`；错误：1001（申请不存在 / 已处理）
- 说明：事务内清除该企业业务数据（ms_company_profile、ms_jobs/ms_jobs_tmp/ms_jobs_contact/ms_jobs_tag、ms_personal_jobs_apply(企业收到的投递)、ms_company_down_resume、ms_company_interview、ms_company_favorites、ms_members_setmeal、ms_company_img，**按一期实际建表裁剪**），然后 `status=1` + `finishtime`；套餐重置（赠送默认套餐）随 M4 套餐体系实现时补充

## 161c. DELETE /api/v1/admin/company/cancellations/{id}（删除申请记录）

- **请求类型**：无参；**鉴权**：admin
- 响应：`code=0`；错误：1001（申请不存在）
- 说明：硬删除（任意状态）；不影响已处理结果

---

## 错误码

沿用 `README.md` 错误码总表：参数/业务失败返回 1001（具体 message 说明）；未登录 401 + code 1001；utype 不符 403 + code 1001。
