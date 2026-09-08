# 12 Billing

All money values are cents (`int64`). A company can pay a pending plan order through WeChat Native payment or Alipay page payment. Entitlements are granted only after a signed provider callback is verified.

## Company APIs

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| GET | `/api/v1/setmeals` | public | Lists sellable plans only. |
| GET | `/api/v1/company/members/setmeal` | company member | Returns the active entitlement snapshot or `null`. |
| POST | `/api/v1/orders` | company member | Creates a plan order. Body: `{ "setmealId": 1 }`. |
| GET | `/api/v1/orders?page=1&pageSize=20` | company member | Lists the caller's orders. |
| POST | `/api/v1/orders/{id}/cancel` | company member | Cancels a pending order. |
| POST | `/api/v1/orders/{id}/pay` | company member | Starts a gateway payment. Body: `{ "provider": "wechat_native" }` or `{ "provider": "alipay_page" }`. |

`Setmeal` contains `id`, `name`, `price`, `durationDays`, `jobsMeanwhile`, `resumeDownloads`, `homePushSlots`, `homeAdSlots`, `enableVideo`, `display`, `sort`, and `description`.

An entitlement contains `setmealId`, `setmealName`, `expireAt`, `jobsMeanwhile`, `resumeDownloadsTotal`, `resumeDownloadsUsed`, `homePushSlots`, `homeAdSlots`, and `enableVideo`. It is a purchase snapshot, so later plan changes do not change purchased rights.

Order status `isPaid`: `1` pending payment, `2` active, `3` cancelled. WeChat payment returns `qrCodeUrl`; Alipay payment returns `redirectUrl`. A hidden or missing plan cannot create a new order. An active purchased plan limits new online job postings; limit exceeded returns business code `3001`. Existing companies without purchase history remain unblocked during this rollout.

## Provider Callbacks

| Method | Path | Response |
| --- | --- | --- |
| POST | `/api/v1/pay/notify/wechat` | WeChat JSON acknowledgement |
| POST | `/api/v1/pay/notify/alipay` | `success` or `failure` |

Callbacks are public endpoints because they are authenticated by the provider signature rather than a user JWT. The server validates the gateway signature, order number, paid amount, application/merchant identity, and gateway transaction ID. A replay with the same transaction ID is idempotent; a different transaction ID cannot change an already paid order.

## Admin APIs

All APIs below require an authenticated admin user.

| Method | Path | Description |
| --- | --- | --- |
| GET | `/api/v1/admin/setmeals` | Lists all plans, including hidden plans. |
| POST | `/api/v1/admin/setmeals` | Creates a plan. |
| PUT | `/api/v1/admin/setmeals/{id}` | Updates a plan. |
| GET | `/api/v1/admin/orders?status=1&page=1&pageSize=20` | Lists orders. `status` may be 1, 2, or 3. |
| POST | `/api/v1/admin/orders/{id}/confirm` | Confirms receipt and grants entitlement. |

Confirm request body:

```json
{ "payAmount": 9900, "payment": "manual" }
```

`payAmount=0` uses the order amount; otherwise it must equal the order amount. A successful confirmation changes the order to active, writes or updates the entitlement snapshot, and synchronizes the plan display fields on the company profile. Error codes: missing order `4001`, already active `4002`, cancelled `4005`.

Refunds, coupon accounting, and payment-provider reconciliation queries remain outside this release. See `server/docs/payment-setup.md` for secret and callback configuration.
