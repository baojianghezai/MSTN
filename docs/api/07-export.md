# 07 后台导出（export）

> 对应设计：`11_v6差异分析与改进建议.md` §四 P0#5（后台导出接口）
> 实现状态：企业导出 + 职位导出（CSV）
> 说明：前缀 `/api/v1`；`admin/**` 走 GVA admin JWT（`x-token`）；导出接口返回 CSV 文件流（非 JSON）。

## 接口清单

| 方法 | 路径 | 请求类型 | 鉴权 | 说明 |
|---|---|---|---|---|
| POST | /api/v1/admin/companies/export | JSON | admin | 企业导出（CSV 下载） |
| POST | /api/v1/admin/jobs/export | JSON | admin | 职位导出（CSV 下载） |

---

## POST /api/v1/admin/companies/export（企业导出）

- **请求类型**：JSON；**鉴权**：admin
- 请求体：`{ "ids": [1, 2, 3] }`（企业资料 id 列表，来自 #134 列表勾选）

```json
{ "ids": [1, 2, 3] }
```

- 响应：CSV 文件流（`Content-Type: text/csv; charset=utf-8`，`Content-Disposition: attachment`，含 UTF-8 BOM）
- 列头（14 列）：企业名称、企业性质、所属行业、企业规模、所在地区、注册资金、企业网址、企业福利、企业简介、联系人、联系电话、座机、联系邮箱、联系地址
- 错误：`ids` 为空或无匹配数据 → 1001「没有符合条件的数据」（JSON 包裹）

> 注：v6 导出列含「币种/QQ」，一期模型未建此两字段，故略去。

---

## POST /api/v1/admin/jobs/export（职位导出）

- **请求类型**：JSON；**鉴权**：admin
- 请求体：`{ "ids": [1, 2, 3] }`（`ms_jobs.id` 列表，来自后台职位管理列表勾选）
- 仅导出 `deleted_at=0` 的职位；软删行即使 id 在列表中也会被跳过
- 响应：CSV 文件流（同企业导出，文件名 `jobs-export-YYYYMMDD.csv`）
- 列头（16 列）：职位ID、职位名称、企业名称、企业UID、工作性质、职位分类、地区、学历要求码、经验要求码、薪资、招聘人数、审核状态、展示状态、发布时间、刷新时间、点击量
- 薪资：面议写「面议」，否则 `min-max`
- 错误：`ids` 为空或无匹配未删除数据 → 1001「没有符合条件的数据」
