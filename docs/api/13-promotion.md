# 13 首页推广（promotion）

首页推广由企业已支付套餐中的展示位权益驱动。套餐有效期内，企业可将已审核、在线的职位投放到首页推流或广告位；广告位须绑定横幅创意，首页以顶部轮播横幅展示。套餐失效、职位下线或审核状态变化后，首页自动不再展示。

## 权益字段

`Setmeal` 与 `MembersSetmeal` 新增：

- `homePushSlots`：同时占用的首页推流位数量。
- `homeAdSlots`：同时占用的首页广告位数量。

默认套餐：基础版 0/0，专业版 1/0，旗舰版 3/1。企业购买套餐并支付成功后，权益快照随订单写入；自定义套餐可由后台套餐 API 配置两个字段。

## 接口

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| GET | `/api/v1/home/promotions` | 公开 | 首页有效推流与广告位 |
| GET | `/api/v1/company/promotions` | 企业 | 当前企业投放及套餐额度 |
| POST | `/api/v1/company/promotions` | 企业 | 创建投放 |
| DELETE | `/api/v1/company/promotions/{id}` | 企业 | 撤下投放 |

### GET /api/v1/home/promotions

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "push": ["职位列表项"],
    "ads": [{
      "id": 100,
      "promotionId": 8,
      "jobsName": "Go 开发工程师",
      "companyname": "名硕科技",
      "adTitle": "加入名硕科技",
      "adSubtitle": "寻找优秀工程师",
      "adImage": "uploads/promotion/banner.png"
    }]
  }
}
```

仅返回关联企业套餐未过期，且职位 `display=1`、`audit=1`、未删除、未过期的记录。`ads` 保留职位列表字段，新增 `promotionId` 与广告创意字段；历史广告记录没有 `adImage` 时，会回退企业 Logo 和职位/企业文案。

### GET /api/v1/company/promotions

响应包含 `list`、`homePushSlots`、`homeAdSlots`、`pushUsed` 与 `adUsed`。无有效套餐时额度均为 0。

### POST /api/v1/company/promotions

请求：

```json
{
  "jobId": 100,
  "type": 2,
  "adTitle": "加入名硕科技",
  "adSubtitle": "寻找优秀工程师",
  "adImage": "uploads/promotion/banner.png"
}
```

- `type=1` 为首页推流，`type=2` 为首页广告位。
- 仅能投放当前企业已审核、在线且未过期的职位。
- 每个职位同一类型只能投放一次；超过套餐名额或无有效推广权益时返回 `3001`。
- 广告位（`type=2`）必须提供 `adImage`，建议使用 1200 × 280 横幅图；`adTitle`（最多 60 字符）和 `adSubtitle`（最多 120 字符）可选，空值时分别回退职位名和企业名。
- 企业端可通过现有 `POST /api/v1/upload` 上传横幅，携带会员 `Authorization: Bearer <token>` 后取响应中的 `data.url` 作为 `adImage`。

### DELETE /api/v1/company/promotions/{id}

仅允许删除本企业投放；删除后立即释放对应展示位。
