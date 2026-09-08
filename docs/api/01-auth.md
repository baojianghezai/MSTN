# 01 认证模块（auth）

> 对应设计：`02_用户账号模块设计.md`、`09_API接口规范.md` §4.1（编号 1-20）
> 实现状态：13 个真实实现 + 6 个 mock（captcha/refresh + 4 个微信）+ 2 个未实现（bind/wechat、upload）
> 说明：本模块全部接口前缀 `/api/v1/auth`；请求类型标注见各接口头部。

## 接口清单

| # | 方法 | 路径 | 请求类型 | 鉴权 | 说明 | 状态 |
|---|---|---|---|---|---|---|
| 1 | POST | /api/v1/auth/captcha | 无参 | 无 | 图形验证码 | ✅ mock（占位图） |
| 2 | POST | /api/v1/auth/sms-code | JSON | 无 | 发送短信验证码 | ✅ 真实（mock 网关） |
| 3 | POST | /api/v1/auth/register | JSON | 无 | 会员注册 | ✅ 真实 |
| 4 | POST | /api/v1/auth/login | JSON | 无 | 密码登录（单点） | ✅ 真实 |
| 5 | POST | /api/v1/auth/login/sms | JSON | 无 | 短信登录（自动注册） | ✅ 真实 |
| 6 | POST | /api/v1/auth/login/wechat | JSON | 无 | 微信登录 | ⏸ mock（待微信参数） |
| 7 | POST | /api/v1/auth/wechat/bind | JSON | 无 | 未登录微信绑定新账号 | ⏸ mock |
| 7a | POST | /api/v1/auth/wechat/mp-bind | JSON | 无 | 小程序手机号绑定 | ⏸ mock |
| 8 | POST | /api/v1/auth/wechat/bind-exist | JSON | 会员 | 已登录绑定微信 | ⏸ mock |
| 9 | POST | /api/v1/auth/refresh | 无参 | 无 | 刷新 token | ✅ mock（二期 refresh 机制） |
| 10 | GET | /api/v1/auth/me | 无参（Header） | 会员 | 当前用户信息 | ✅ 真实 |
| 11 | POST | /api/v1/auth/logout | 无参 | 会员 | 退出登录（清会话） | ✅ 真实 |
| 12 | PUT | /api/v1/auth/password | JSON | 会员 | 修改密码 | ✅ 真实 |
| 13 | POST | /api/v1/auth/password/reset | JSON | 无 | 忘记密码重置 | ✅ 真实 |
| 14 | PUT | /api/v1/auth/bind/wechat | JSON | 会员 | 绑定微信 | ❌ 未实现（待微信参数） |
| 15 | PUT | /api/v1/auth/bind/mobile | JSON | 会员 | 绑定手机 | ✅ 真实 |
| 16 | PUT | /api/v1/auth/unbind/mobile | 无参 | 会员 | 解绑手机 | ✅ 真实 |
| 17 | POST | /api/v1/auth/logout-other | 无参 | 会员 | 强制其他端下线 | ✅ 真实 |
| 18 | POST | /api/v1/auth/cancel | JSON | 会员 | 账号注销（两阶段匿名化） | ✅ 真实 |
| 19 | POST | /api/v1/auth/cancel/restore | JSON | 无 | 注销恢复（冷静期） | 🔀 已迁后台（见 02-account.md #158/#159） |
| 20 | POST | /api/v1/upload | Form-data | 会员 | 通用文件上传 | ❌ 未实现（M3 随上传模块） |

> 状态图例：✅ 可用 / ⏸ mock 待对接 / ❌ 未实现

---

## 1. POST /api/v1/auth/captcha

- **请求类型**：无参（POST 空 body 或省略 body）
- **鉴权**：无
- **Content-Type**：无

获取图形验证码（防机器人刷短信/登录）。

**响应**：

| 字段 | 类型 | 说明 |
|---|---|---|
| captchaId | string | 验证码 ID（后续接口回传） |
| captchaImg | string | base64 图片（`data:image/png;base64,...`） |

```json
{ "code": 0, "message": "success",
  "data": { "captchaId": "8f3a2c", "captchaImg": "data:image/png;base64,iVBOR..." } }
```

---

## 2. POST /api/v1/auth/sms-code

- **请求类型**：JSON（body）
- **鉴权**：无
- **Content-Type**：`application/json`

发送短信验证码。

**请求 body**：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| mobile | string | 是 | 手机号（11 位，`^(13\|14\|15\|17\|18\|19)\d{9}$`） |
| type | string | 否 | 场景枚举：`register`（默认）/`login`/`reset`/`bind`/`cancellation` |

```json
{ "mobile": "13800138000", "type": "register" }
```

**响应**：成功 `code=0`，message="验证码已发送"。

**规则**：
- 验证码 6 位数字，Redis 存储 **5 分钟有效**
- 手机号 13-19 全号段（含 16 虚商号）
- type 必须是 `register/login/reset/bind/cancellation` 之一，否则 1001
- type=reset/**cancellation** 时手机号**须已注册**（未注册返回 2001，防未注册号探测）；register/login/bind 不查库
- **限流均按手机号全局，不区分场景（防换 type 绕过）**：60 秒重发限制；同号 5 条/小时、10 条/日（超限返回 2007）
- 一期不接真实短信网关，验证码打印在服务端日志（mock）
- 错误码：2007（频繁）、1001（手机号格式/type 非法）、2001（reset/cancellation 场景手机号未注册）

---

## 3. POST /api/v1/auth/register

- **请求类型**：JSON（body）
- **鉴权**：无
- **Content-Type**：`application/json`

会员注册（个人/企业）。

**请求 body**：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| utype | int8 | 是 | 用户类型：1=个人 2=企业 |
| mobile | string | 是 | 手机号 |
| code | string | 是 | 短信验证码（type=register 发送的） |
| password | string | 是 | 密码（≥6 位，bcrypt 存储） |

```json
{ "utype": 1, "mobile": "13800138000", "code": "123456", "password": "abc12345" }
```

**响应**：

| 字段 | 类型 | 说明 |
|---|---|---|
| token | string | 会员 JWT access token |
| uid | uint64 | 会员 ID |
| utype | int8 | 用户类型 |
| mobile | string | 当前登录手机号（解绑后为 `unbound_<uid>` 占位） |
| passwordSet | bool | 是否已设置密码（false 时前端引导走 #13 忘记密码重置设密） |

```json
{ "code": 0, "message": "success",
  "data": { "token": "eyJhbGci...", "uid": 12, "utype": 1, "mobile": "13800138000", "passwordSet": true } }
```

**业务规则**：
- 手机号唯一（uk_mobile），已注册返回 **2005**
- 注册成功自动创建资料壳：个人→ms_members_info；企业→ms_company_profile（空壳待补全）
- 企业注册后需补全资料并过资质审核（audit=1）才能发布职位

**错误码**：2005/2006/1001（手机号格式/密码过短）

---

## 4. POST /api/v1/auth/login

- **请求类型**：JSON（body）
- **鉴权**：无
- **Content-Type**：`application/json`

密码登录。账号类型**自动识别**：手机号→mobile 字段、邮箱→email、其他→username。

**请求 body**：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| account | string | 是 | 手机号/邮箱/用户名 |
| password | string | 是 | 密码 |

```json
{ "account": "13800138000", "password": "abc12345" }
```

**响应**：同 register（token/uid/utype/mobile/passwordSet）。

**业务规则**：
- 密码 bcrypt 校验，错误返回 **2002**
- 恢复后未设密码的账号（password 为空）密码登录 → **2002「该账号未设置密码，请用短信验证码登录后设置密码」**
- 账号不存在返回 **2001**；账号暂停（status=2）或注销（status=3）返回 **2004**
- **风控**：同一 account 密码错误 5 次 → 锁定 30 分钟（Redis），期间返回 2004
- **单点登录**：同一 uid 新登录顶掉旧 token；旧 token 再访问返回 401 "账号已在其他设备登录"

**错误码**：2001/2002/2004

---

## 5. POST /api/v1/auth/login/sms

- **请求类型**：JSON（body）
- **鉴权**：无
- **Content-Type**：`application/json`

短信验证码登录。**无账号自动注册**（拉新关键）。

**请求 body**：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| mobile | string | 是 | 手机号 |
| code | string | 是 | 短信验证码（type=login 发送的） |
| utype | int8 | 否 | 0=未指定（默认）/1=个人/2=企业。**仅自动注册时生效**，已注册账号一律以库内身份登录 |

```json
{ "mobile": "13800138000", "code": "123456", "utype": 0 }
```

**响应**：同 register（token/uid/utype/mobile/passwordSet）。恢复后未设密账号 `passwordSet=false`，前端据此引导设密。

**业务规则（含已注册手机号处理）**：

| 场景 | 处理 |
|---|---|
| 未注册手机号 | 自动注册（utype 缺省按 0→个人），随机初始密码，返回新账号 token |
| **已注册手机号** | **忽略请求 utype，以库内身份直接登录**（一个手机号一种身份，02 设计互斥）；不新建账号 |
| 已注册但 status=2（暂停） | 拒绝，2004 "账号已暂停" |
| 已注册但 status=3（注销） | 拒绝，2004 "账号已注销" |
| 验证码错误/过期 | 2006 |

**验证码兼容规则**：本接口接受 **login 或 register 两种场景**发送的验证码（登录入口发码时 type 常走默认 register，两者语义等价）；`reset`/`bind` 场景的码不可用于登录。

**错误码**：2006/2004/1001

---

## 6. POST /api/v1/auth/login/wechat

- **请求类型**：JSON（body）
- **鉴权**：无
- **Content-Type**：`application/json`

微信登录（公众号/小程序）。**一期 mock**：不接微信开放平台，直接返回 bindToken。

**请求 body**：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| code | string | 是 | 微信登录 code |
| channel | string | 否 | 渠道：pc/h5/mp（小程序） |

```json
{ "code": "071xxx", "channel": "h5" }
```

**响应**：

| 字段 | 类型 | 说明 |
|---|---|---|
| bindToken | string | 临时绑定凭证（30 分钟有效），用于后续 bind 接口 |

**二期实现规划**：code→openid/unionid 换取→查 ms_members_bind→命中直接发 token；未命中返回 bindToken。

---

## 7. POST /api/v1/auth/wechat/bind

- **请求类型**：JSON（body）
- **鉴权**：无
- **Content-Type**：`application/json`

未登录微信扫码后绑定**新账号**（创建）。

**请求 body**：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| bindToken | string | 是 | login/wechat 返回的临时凭证 |
| mobile | string | 是 | 手机号 |
| code | string | 是 | 短信验证码（type=bind） |

**响应**：同 register。

---

## 7a. POST /api/v1/auth/wechat/mp-bind

- **请求类型**：JSON（body）
- **鉴权**：无
- **Content-Type**：`application/json`

小程序 `getPhoneNumber` 手机号绑定。

**请求 body**：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| bindToken | string | 是 | 临时凭证 |
| code | string | 是 | wx.login 的 code |
| encryptedData | string | 是 | getPhoneNumber 返回的加密数据 |
| iv | string | 是 | 加密向量 |

**响应**：同 register。

---

## 8. POST /api/v1/auth/wechat/bind-exist

- **请求类型**：JSON（body）
- **鉴权**：会员（`Authorization: Bearer <token>`）
- **Content-Type**：`application/json`

**已登录会员**把微信绑定到当前账号。

**请求 body**：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| code | string | 是 | 微信登录 code |

**响应**：`code=0`。

---

## 9. POST /api/v1/auth/refresh

- **请求类型**：无参（POST 空 body）
- **鉴权**：无（一期 mock；二期需带 refresh_token）
- **Content-Type**：无

刷新 token（一期 mock，返回占位 token）。

**响应**：

| 字段 | 类型 | 说明 |
|---|---|---|
| token | string | 新 access token |

---

## 10. GET /api/v1/auth/me

- **请求类型**：无参（GET，无 query 参数）
- **鉴权**：会员（`Authorization: Bearer <token>`）
- **Content-Type**：无

当前登录用户信息。

**请求头**：

| Header | 值 | 说明 |
|---|---|---|
| Authorization | `Bearer <access_token>` | 会员 JWT |

**响应**：

| 字段 | 类型 | 说明 |
|---|---|---|
| uid | uint64 | 会员 ID |
| utype | int8 | 1=个人 2=企业 |

> 二期将扩展：username/mobile（脱敏）/avatar/资料完整度/企业审核状态等。

---

## 11. POST /api/v1/auth/logout

- **请求类型**：无参
- **鉴权**：会员
- **Content-Type**：无

退出登录（删除 Redis 会话记录）。**响应**：`code=0`。

## 17. POST /api/v1/auth/logout-other

- **请求类型**：无参
- **鉴权**：会员
- **Content-Type**：无

强制其他端下线（删除 Redis 会话记录，所有已签发 token 失效）。**响应**：`code=0`。

---

## 12. PUT /api/v1/auth/password

- **请求类型**：JSON（body）
- **鉴权**：会员
- **Content-Type**：`application/json`

修改密码（原密码校验）。

**请求 body**：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| old_password | string | 是 | 原密码 |
| new_password | string | 是 | 新密码（≥6 位，不能与原密码相同） |

```json
{ "old_password": "abc12345", "new_password": "xyz98765" }
```

**响应**：`code=0`。

**业务规则**：
- 原密码错误 → 2002 "原密码错误"
- 新密码 <6 位 → 2002；新旧相同 → 2002
- **改密成功踢出所有会话**（旧 token 全部失效，需重新登录）

**错误码**：2002

---

## 13. POST /api/v1/auth/password/reset

- **请求类型**：JSON（body）
- **鉴权**：无
- **Content-Type**：`application/json`

忘记密码重置（短信验证码，reset 场景严格匹配）。

**请求 body**：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| mobile | string | 是 | 手机号（须已注册） |
| code | string | 是 | 短信验证码（**必须 type=reset 发送**） |
| new_password | string | 是 | 新密码（≥6 位） |

**响应**：`code=0`。

**错误码**：2001（未注册）/2006（码错误或用错场景）/2002（密码过短）

---

## 15. PUT /api/v1/auth/bind/mobile

- **请求类型**：JSON（body）
- **鉴权**：会员
- **Content-Type**：`application/json`

绑定手机（换绑）。

**请求 body**：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| mobile | string | 是 | 新手机号 |
| code | string | 是 | 短信验证码（**必须 type=bind 发送**） |

**响应**：`code=0`。

**业务规则**：
- 新手机号已被其他账号绑定 → 2005
- bind 场景严格匹配（login/register 码不可用）→ 2006
- 成功后 ms_members.mobile 更新 + mobile_audit=1

---

## 16. PUT /api/v1/auth/unbind/mobile

- **请求类型**：无参
- **鉴权**：会员
- **Content-Type**：无

解绑手机。

**响应**：`code=0`。

**业务规则**：
- mobile 置 `unbound_{uid}` 占位（列非 NULL 的掩码方案，01 设计 §六.2）
- 记录 ms_unbind_mobile（原手机号留痕，合规审计）
- 已解绑再调 → 1001 "未绑定手机号"

---

## 18. POST /api/v1/auth/cancel

- **请求类型**：JSON（body）
- **鉴权**：会员
- **Content-Type**：`application/json`

账号注销（两阶段匿名化，02 设计 §2.5）。

**请求 body**：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| code | string | 是 | 短信验证码（type=bind 发送） |

**响应**：`code=0`。

**业务规则（阶段一，注销时立即执行）**：
- status=3 + deleted_at=now
- username → `u_匿名_{uid}`、email 清空、password 清空
- **mobile 保留原值**（冷静期恢复用；账号已 status=3 无法登录，不构成占用）
- 会话失效（token 全踢）
- **阶段二（30 天期满，Cron 执行）**：mobile 置掩码释放号码（一期 Cron 随 M3 定时任务实现）

**错误码**：2006

---

## 19. POST /api/v1/auth/cancel/restore（🔀 已迁移到后台）

> 该公开接口已下线，避免被恶意调用。恢复能力改为后台客服通过申诉处理流程执行：
> **用户提交申诉（#45）→ 后台客服处理申诉（#159，`restore=true`）→ 恢复账号**。
> 详见 `02-account.md` 的 #158/#159。

原业务规则保留：按 mobile 匹配 status=3 的账号 → 恢复 status=1、deleted_at=0；恢复后 password 为空（注销时已清空），需走忘记密码重置。

**错误码**：1001
