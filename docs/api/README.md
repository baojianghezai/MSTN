# 名硕人力平台 API 文档

> 版本：V1.0 | 契约基准：设计文档 `09_API接口规范.md`
> 适用范围：PC 前台（web-front）、H5/小程序（mobile）、后台业务页（GVA web 的 hrc 部分）
> 本文档为**人工维护**的唯一权威接口文档；swagger.json 供前端自动生成类型使用，两者冲突时以本文档为准。

## 目录

- [总体规范](#总体规范)
- [错误码总表](#错误码总表)
- [01 认证模块（auth）](01-auth.md)
- [02 账号资料与申诉（profile/appeal）](02-account.md)
- [03 内容/配置/分类/上传（cms）](03-cms.md)
- [04 简历模块（resume）](04-resume.md)
- [05 视频面试（video-interview）](05-video-interview.md)
- [06 统计看板与报表（dashboard）](06-dashboard.md)
- [07 后台导出（export）](07-export.md)
- [08 微信支付回调日志（wxpay-log）](08-wxpay-log.md)
- [09 职位发布与管理（jobs）](09-jobs.md)
- [10 投递（apply）](10-apply.md)
- [11 企业收简历（company-apply）](11-company-apply.md)
- [12 套餐与订单（billing）](12-billing.md)
- [14 后台数据清理（data-cleanup）](14-data-cleanup.md)
- [15 标准线下面试邀请（interview）](15-interview.md)
- [16 站内信中心（message）](16-message.md)
- [17 企业找人才与人才库（talent）](17-talent.md)
- [18 公开找企业（company-search）](18-company-search.md)

## 总体规范

### 基础约定

| 项 | 约定 |
|---|---|
| 协议 | HTTPS + JSON，上传接口除外 |
| 前缀 | 所有 hrc 接口 `/api/v1` 开头（GVA 原生接口不含此前缀） |
| 时间格式 | 时间戳字段 int64 unix 秒；字符串时间 `yyyy-MM-dd HH:mm:ss` |
| 金额 | 订单/套餐金额单位为**分**（int64）；职位薪资 minwage/maxwage 单位为**元** |
| 分页参数 | `page`（默认1）、`pageSize`（默认20，最大100） |
| 字段命名 | 请求/响应字段小驼峰；枚举数字 + `xxxCn` 中文冗余字段 |

### 请求类型约定（每个接口文档必须标注）

| 请求类型 | 方法 | Content-Type | 参数位置 | 适用场景 |
|---|---|---|---|---|
| **JSON** | POST/PUT | `application/json` | 请求体（body） | 创建/更新/状态操作（默认） |
| **Query** | GET | — | URL 查询串 `?k=v&k2=v2` | 列表查询/详情/筛选 |
| **Form-data** | POST/PUT | `multipart/form-data` | 表单字段+文件 | 文件上传（图片/附件/简历） |
| **无参** | POST | 可不带 body | — | 退出/刷新等纯动作 |

**标注规则**：
- 每个接口开头标注 `请求类型：JSON / Query / Form-data / 无参`；
- JSON 请求：表格列出 body 字段；Query 请求：表格列出 query 参数；
- Form-data：文件字段标注 `type=file`，其他字段标注 `type=string` 等；
- 鉴权接口统一请求头：`Authorization: Bearer <access_token>`；
- 批量删除：`DELETE /resource` + JSON body `{"ids":[]}`（特例，文档单独标注）。

### 响应结构（hrc 域）

```json
// 成功
{ "code": 0, "message": "success", "data": { } }

// 失败（HTTP 200）
{ "code": 2002, "message": "密码错误", "data": null }

// 分页
{ "code": 0, "message": "success", "data": { "list": [], "total": 0, "page": 1, "pageSize": 20 } }
```

> 注意：GVA 原生接口是 `{code, data, msg}`（code=7 错误），与 hrc 双轨并存。前端两套 axios 实例分别处理。

### HTTP 状态码策略

| 情况 | HTTP | 业务 code |
|---|---|---|
| 成功 | 200 | 0 |
| 参数错误/业务失败 | 200 | 业务 code（2xxx-5xxx） |
| 未登录/token 失效 | **401** | 1001 |
| 无权限（utype 不符） | **403** | 1001 |
| 资源不存在 | 404 | 1001 |
| 限流 | 429 | 1001 |
| 服务器错误 | 500 | 1001 |

### 鉴权

- Header：`Authorization: Bearer <access_token>`（会员 JWT，独立于 GVA 后台 JWT）
- JWT Claims：`{uid, utype, ver}`；utype：1=个人 2=企业
- 角色分组：`/api/v1/personal/**`（utype=1）、`/api/v1/company/**`（utype=2）
- 接口清单中【鉴权】列：`无`=公开；`会员`=需登录；`个人`=需 utype=1；`企业`=需 utype=2；`admin`=GVA 后台登录态+Casbin

## 错误码总表

### 1xxx 通用

| code | 含义 | 说明 |
|---|---|---|
| 0 | 成功 | |
| 1001 | 通用错误 | 参数错误/未登录/无权限/服务器错误（具体 message 说明） |

### 2xxx 用户/账号

| code | 含义 | 触发场景 |
|---|---|---|
| 2001 | 账号不存在 | 密码登录：账号未注册 |
| 2002 | 密码错误 | 密码登录：密码不匹配 |
| 2003 | 验证码错误 | 图形验证码校验失败（预留） |
| 2004 | 账号被锁定/禁用 | 账号暂停/注销/登录失败锁定 |
| 2005 | 手机号已注册 | 注册：手机号唯一冲突 |
| 2006 | 短信验证码错误 | 验证码错误或已过期 |
| 2007 | 短信发送频繁 | 60 秒内重复发送 |

### 3xxx 职位/简历（预留）

| code | 含义 | 触发场景 |
|---|---|---|
| 3001 | 套餐限制 | 同时在线职位数超套餐 |
| 3002 | 审核不通过 | 职位/简历未过审 |
| 3003 | 已投递过 | 同一企业对同一简历重复投递 |
| 3004 | 企业未认证 | 企业资质未审核通过 |
| 3005 | 职位已过期 | 投递已过期职位 |

### 4xxx 订单/支付（预留）

| code | 含义 | 触发场景 |
|---|---|---|
| 4001 | 订单不存在 | 支付/查询订单号无效 |
| 4002 | 订单已支付 | 重复支付回调 |
| 4003 | 重复下单 | 同用户并发重复下单 |
| 4004 | 回调验签失败 | 支付回调签名错误 |
| 4005 | 订单已关闭 | 超时/取消后操作 |

### 5xxx 业务（预留）

| code | 含义 | 触发场景 |
|---|---|---|
| 5001 | 已下载 | 同一简历重复下载 |
| 5002 | 每日上限 | 下载简历达每日上限 |
| 5003 | 积分不足 | 积分付费通道余额不足 |
| 5004 | 无投递记录 | 已投递直接可见分支未命中 |
