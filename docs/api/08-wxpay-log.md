# 08 微信支付回调日志（wxpay-log）

> 对应设计：`11_v6差异分析与改进建议.md` §四 P0#7（v6 `qs_wxpay_log` 平移）、`06_商业化模块设计.md` §3.3（对账配套）
> 实现状态：1 个后台查询接口 + 记录服务（M4 支付回调调用）
> 说明：前缀 `/api/v1`；`admin/**` 走 GVA admin JWT（`x-token`）。

## 接口清单

| 方法 | 路径 | 请求类型 | 鉴权 | 说明 |
|---|---|---|---|---|
| GET | /api/v1/admin/wxpay-logs | Query | admin | 微信支付回调日志列表 |

---

## 数据模型（ms_wxpay_log，v6 qs_wxpay_log 平移）

| 字段 | 类型 | 说明 |
|---|---|---|
| openid | string(50) | 支付用户 openid |
| trade_no | string(100) | 商户/微信交易号 |
| amount | string(30) | 金额（字符串，v6 原结构） |
| addtime | int64 | 时间（unix 秒） |
| status | int8 | 0=失败 1=成功 |
| fail_reason | string(255) | 失败原因 |

> `Record()` 服务方法供 M4 支付回调（`POST /api/v1/pay/notify/wxpay`）调用；一期回调链路未建（M4），先建表 + 后台查询。

## GET /api/v1/admin/wxpay-logs（列表）

- Query：`page`、`pageSize`、`status`（缺省全部，0=失败 1=成功）
- 响应：分页 `{list,total,page,pageSize}`，list 项为日志字段（`id/openid/tradeNo/amount/addtime/status/failReason`）

```json
{ "code": 0, "message": "success",
  "data": { "list": [ { "id": 1, "openid": "oXyZ", "tradeNo": "T20260817001",
            "amount": "100", "addtime": 1755000000, "status": 1, "failReason": "" } ],
            "total": 1, "page": 1, "pageSize": 20 } }
```

---

## 一期未实现（依赖 M4，待补）

- 支付回调链路（`pay/notify/wxpay` + 验签 + 幂等发货）：M4 订单/支付模块
- 对账 Cron（本地已支付 vs 第三方订单查询比对）：M4
