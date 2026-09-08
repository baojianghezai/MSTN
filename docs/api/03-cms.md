# 03 内容/配置/分类/上传（cms）

> 对应设计：`09_API接口规范.md` §4.1(20/21/43/44)、§4.6.2(150)、`07_内容招聘会站内信模块设计.md` §2.5
> 实现状态：5 个真实实现
> 说明：前缀 `/api/v1`；`upload` 需会员 JWT；`admin/configs` 走 GVA admin JWT。

## 接口清单

| # | 方法 | 路径 | 请求类型 | 鉴权 | 说明 |
|---|---|---|---|---|---|
| 20 | POST | /api/v1/upload | Form-data | 会员 | 通用文件上传（返回 url） |
| 21 | GET | /api/v1/categories | Query | 无 | 分类聚合（学历/经验/薪资/行业/地区/专业/性别/婚姻/企业性质/企业规模） |
| 43 | GET | /api/v1/pages/{alias} | 无参 | 无 | 静态页（关于我们等） |
| 44 | GET | /api/v1/navigations | 无参 | 无 | 前台导航 |
| 150 | GET/PUT | /api/v1/admin/configs | Query/JSON | admin | 系统配置（分组读取/保存） |

---

## 20. POST /api/v1/upload（通用上传）

- **请求类型**：Form-data；**鉴权**：会员（`Authorization: Bearer <token>`）
- 表单字段：`file`（type=file，单文件；图片/附件/富文本插图）
- 存储：由 `system.oss-type` 决定（一期 local → `uploads/file/`）
- 响应：`{ "url": "uploads/file/xxx.ext" }`

## 21. GET /api/v1/categories（分类聚合）

- **请求类型**：Query（无参）；**鉴权**：无
- 响应：按分组别名聚合，`{ "<alias>": [{id, groupId, parentId, name, sort, display}] }`
- 分类项：`{id, groupId, parentId, name, sort, display}`；下拉选项约定 **value=分类 `id`、label=分类 `name`**（对应资料字段的 `xxx` code 与 `xxxCn` 中文冗余）
- 一期已 seed 以下分组（`sort` 为排序）：

| alias | 分组名 | 一期分类值 |
|---|---|---|
| education | 学历 | 初中及以下 / 高中 / 中专 / 大专 / 本科 / 硕士 / 博士 |
| experience | 经验 | 在校生 / 应届生 / 1年以内 / 1-3年 / 3-5年 / 5-10年 / 10年以上 |
| wage | 薪资 | 3K以下 / 3-5K / 5-10K / 10-15K / 15-20K / 20-30K / 30K以上 / 面议 |
| trade | 行业 | 互联网/IT / 电子商务 / 金融 / 教育培训 / 医疗健康 / 制造业 / 房地产 / 物流运输 / 餐饮服务 / 其他 |
| district | 地区 | 暂为空（省市县 ~2800 条待从原 74CMS `sql_category_district.sql` 导入） |
| major | 专业 | 计算机科学与技术 / 软件工程 / 电子信息工程 / 通信工程 / 自动化 / 机械设计制造及其自动化 / 土木工程 / 电气工程及其自动化 / 金融学 / 会计学 / 财务管理 / 市场营销 / 工商管理 / 人力资源管理 / 国际经济与贸易 / 法学 / 汉语言文学 / 英语 / 临床医学 / 护理学 / 教育学 / 学前教育 / 其他 |
| sex | 性别 | 男 / 女 |
| marriage | 婚姻 | 未婚 / 已婚 / 保密 |
| nature | 企业性质 | 国有企业 / 民营企业 / 外资企业 / 合资企业 / 事业单位 / 政府机关 / 其他 |
| scale | 企业规模 | 少于50人 / 50-99人 / 100-499人 / 500-999人 / 1000人以上 |

- 资料枚举字段 → 分组对应：个人 `sex`→sex、`marriage`→marriage、`major`→major、`education`→education、`experience`→experience；企业 `nature`→nature、`scale`→scale、`trade`→trade、`district`→district
- 说明：seed 幂等按分组补齐（分组已存在不覆盖）；新增分组随服务重启自动补种

## 21a. GET /api/v1/categories/districts（地区单层查询）

- **请求类型**：Query；**鉴权**：无
- 参数：`parentId` 可选，缺省或 `0` 返回 34 个顶级地区；传入地区分类 `id` 返回其直接下级。
- 响应：`[{id, groupId, parentId, name, sort, display}]`。
- 性能约束：`/categories` 只保留 `district: []` 占位，不返回完整地区树；前端继续展开时调用本接口，避免一次传输 3241 条地区记录。

旧系统目录补全后，`/categories` 还提供 `jobnature`、`jobtag`、`resumetag`、`language`、`languagelevel`、`current` 和 `age` 分组。职位发布的福利标签使用 `jobtag` 的分类 ID。

## 43. GET /api/v1/pages/{alias}（静态页）

- **请求类型**：无参；**鉴权**：无
- 参数：`alias`（路径参数，如 `about`/`contact`/`fee`）
- 响应：`{id, alias, title, contents, addTime}`；不存在返回 1001

## 44. GET /api/v1/navigations（前台导航）

- **请求类型**：无参；**鉴权**：无
- 响应：`[{id, title, url, sort, display, addTime}]`（display=1，按 sort 升序）

## 150. GET/PUT /api/v1/admin/configs（系统配置）

- **请求类型**：Query / JSON；**鉴权**：admin（GVA 后台 JWT，Header `x-token`）
- GET：`?group=site`（group 为空返回全部）
- PUT body：`{ "group": "site", "items": [{ "name": "site_name", "value": "xxx", "remark": "..." }] }`
- 说明：按 `name` 幂等 upsert（`ms_config` 表，分组 site/security/sms/payment）

---

## 错误码

沿用 `README.md` 错误码总表：业务失败 1001；未登录 401 + code 1001（会员）/ 401 + code 7（admin）。
