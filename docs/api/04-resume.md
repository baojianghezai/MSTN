# 04 简历模块（resume）

> 对应设计：`04_简历模块设计.md` §2.2-§2.8/§3.1/§3.5/§4.1、`01_数据库设计.md` §1.5/§2.3、`09_API接口规范.md` §4.2(48-52/54/55/57)
> 实现状态（M3 收尾批，2026-08-20）：8 个个人端接口（#48 创建 / #49 列表 / #50 详情 / #51 编辑 / #52 删除 / #54 公开隐藏 / #55 设默认 / #57 完善度详情）；6 子表（education/work/language/training/credent/project）随 #48/#51 同事务全量替换
> 后置（不在本批）：#53 刷新、#56 复制、#58 附件上传、#59 照片、#60 发邮箱、ms_resume_img / ms_resume_outward（04 §2.8 后注）
> 说明：前缀 `/api/v1`；`personal/**` 需 utype=1（个人会员 JWT）。

## 接口清单

| # | 方法 | 路径 | 请求类型 | 鉴权 | 说明 |
|---|---|---|---|---|---|
| 48 | POST | /api/v1/personal/resumes | JSON | 个人 | 创建简历（含 6 子表，项目经历限 6 条） |
| 49 | GET | /api/v1/personal/resumes | 分页参数 | 个人 | 我的简历列表（轻量字段，默认简历在前） |
| 50 | GET | /api/v1/personal/resumes/{id} | 无参 | 个人 | 简历详情（编辑回显，含 6 子表） |
| 51 | PUT | /api/v1/personal/resumes/{id} | JSON | 个人 | 编辑简历（6 子表全量替换） |
| 52 | DELETE | /api/v1/personal/resumes/{id} | 无参 | 个人 | 删除简历（软删除） |
| 54 | PUT | /api/v1/personal/resumes/{id}/display | JSON | 个人 | 公开/隐藏切换 |
| 55 | PUT | /api/v1/personal/resumes/{id}/default | 无参 | 个人 | 设为默认简历 |
| 57 | GET | /api/v1/personal/resumes/{id}/completeness | 无参 | 个人 | 完善度详情（缺失字段清单） |

---

## 子表字段（04 §2.3-2.8，v6 原表平移）

### 项目经历 ms_resume_project（§2.3，限 6 条）

| 字段 | 类型 | 说明 |
|---|---|---|
| startyear / startmonth | uint16 / uint8 | 开始年/月 |
| endyear / endmonth | uint16 / uint8 | 结束年/月 |
| todate | int8 | 至今标记（0=已结束 1=至今） |
| projectname | string(50) | 项目名称 |
| role | string(50) | 项目角色 |
| description | string(1000) | 项目描述 |

### 教育经历 ms_resume_education（§2.4，campus_id 已剔除，不限条数）

| 字段 | 类型 | 说明 |
|---|---|---|
| startyear / startmonth | uint16 / uint8 | 开始年/月 |
| endyear / endmonth | uint16 / uint8 | 结束年/月 |
| todate | int8 | 至今标记（0=已结束 1=至今） |
| school | string(50) | 学校 |
| speciality | string(50) | 专业 |
| education / educationCn | uint16 / string | 学历编码 / 中文 |

### 工作经历 ms_resume_work（§2.5，不限条数）

| 字段 | 类型 | 说明 |
|---|---|---|
| startyear / startmonth | uint16 / uint8 | 开始年/月 |
| endyear / endmonth | uint16 / uint8 | 结束年/月 |
| todate | int8 | 至今标记（0=已结束 1=至今） |
| companyname | string(50) | 公司名称 |
| jobs | string(30) | 职位 |
| achievements | string(1000) | 工作业绩 |

### 语言能力 ms_resume_language（§2.6，不限条数）

| 字段 | 类型 | 说明 |
|---|---|---|
| language / languageCn | uint16 / string | 语言编码 / 中文 |
| level / levelCn | uint16 / string | 等级编码 / 中文 |

### 培训经历 ms_resume_training（§2.7，不限条数）

| 字段 | 类型 | 说明 |
|---|---|---|
| startyear / startmonth | uint16 / uint8 | 开始年/月 |
| endyear / endmonth | uint16 / uint8 | 结束年/月 |
| todate | int8 | 至今标记（0=已结束 1=至今） |
| agency | string(50) | 培训机构 |
| course | string(50) | 课程名称 |
| description | string(1000) | 培训描述 |

### 证书 ms_resume_credent（§2.8，不限条数）

| 字段 | 类型 | 说明 |
|---|---|---|
| name | string(255) | 证书名称 |
| year / month | uint16 / uint8 | 获得年份/月份 |
| images | string(255) | 证书图片 URL（一期预留，前端不做上传 UI） |

> 提交时无需带 `id`/`pid`/`uid`（后端按登录会员 uid 与简历 id 盖章）；编辑为**全量替换**（删旧插新）。
> **⚠️ 键名约定**：子表数组键名为 `educations`（复数）——避免与主表学历编码 `education`（uint16）冲突；其余为 `work` / `language` / `training` / `credent` / `projects`。
> **上限**：同一简历项目经历（projects）**最多 6 条**，超出返回 1001「项目经历最多 6 条」；五张新子表一期不限条数。

---

## 48. POST /api/v1/personal/resumes（创建简历）

- **请求类型**：JSON；**鉴权**：个人（utype=1）
- 请求体：简历主表可编辑字段 + 6 子表数组（可空）：`educations` / `work` / `language` / `training` / `credent` / `projects`
- 响应成功：`{ "id": 简历ID }`

```json
// 请求（节选：主表 + 教育经历子表 + 工作经历子表）
{
  "title": "后端开发工程师", "fullname": "张三", "sex": 1, "sexCn": "男",
  "education": 4, "educationCn": "本科", "major": 7, "majorCn": "软件工程",
  "birthdate": 1995, "district": "4401", "districtCn": "广州",
  "wage": 8, "wageCn": "8-10K", "intentionJobs": "Go 开发",
  "specialty": "自我评价", "telephone": "13800138000", "email": "z@a.com",
  "educations": [
    { "startyear": 2015, "startmonth": 9, "endyear": 2019, "endmonth": 6, "todate": 0,
      "school": "中山大学", "speciality": "软件工程", "education": 4, "educationCn": "本科" }
  ],
  "work": [
    { "startyear": 2019, "startmonth": 7, "endyear": 2026, "endmonth": 0, "todate": 1,
      "companyname": "某某科技", "jobs": "Go 工程师", "achievements": "负责简历模块" }
  ],
  "projects": []
}
// 响应
{ "code": 0, "message": "success", "data": { "id": 12 } }
```

- 服务端默认：`display=1`（公开）、`audit=1`（通过）、`displayName=1`（显示姓名）、`click=1`、`addtime`/`refreshtime` 为当前时间戳
- **def 逻辑**：本人**首份**简历自动 `def=1`（默认简历），后续简历 `def=0`
- 完善度（complete_percent）与搜索索引（key_full/key_precise）按提交数据计算并同事务写入（04 §3.1 步骤 4，与 #57 同源）

---

## 49. GET /api/v1/personal/resumes（我的简历列表）

- **请求类型**：GET 分页参数（`page` / `pageSize`，pageSize 上限 100）；**鉴权**：个人（utype=1，仅本人）
- 行为：仅本人未软删简历；**默认简历在前，按创建时间倒序**
- 响应：分页 `{ list, total, page, pageSize }`；list 项为**轻量主表字段**（不返子表）：

```json
{ "code": 0, "message": "success",
  "data": {
    "list": [
      { "id": 12, "title": "后端开发工程师", "completePercent": 85, "def": 1,
        "display": 1, "addtime": 1755000000, "refreshtime": 1755000100 }
    ],
    "total": 2, "page": 1, "pageSize": 10 } }
```

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint64 | 简历 id |
| title | string | 简历标题 |
| completePercent | int8 | 完善度 0-100（与 #57 同源） |
| def | int8 | 默认简历标记（1=默认） |
| display | int8 | 1=公开 2=不公开 |
| addtime | int64 | 创建时间戳 |
| refreshtime | int64 | 刷新时间戳 |

---

## 50. GET /api/v1/personal/resumes/{id}（简历详情，编辑回显）

- **请求类型**：无参；**鉴权**：个人（utype=1，仅本人）
- 响应：主表字段平铺 + 6 子表数组（均按 id 升序；教育经历子表键名 `educations`）

```json
{ "code": 0, "message": "success",
  "data": {
    "id": 12, "uid": 9, "title": "后端开发工程师", "fullname": "张三",
    "education": 4, "educationCn": "本科", "addtime": 1755000000,
    "educations": [
      { "id": 1, "pid": 12, "uid": 9, "startyear": 2015, "startmonth": 9,
        "endyear": 2019, "endmonth": 6, "todate": 0,
        "school": "中山大学", "speciality": "软件工程", "education": 4, "educationCn": "本科" }
    ],
    "work": [], "language": [], "training": [], "credent": [],
    "projects": [] } }
```

- 错误码：1001（简历不存在 / 非本人简历 / 已软删）

---

## 51. PUT /api/v1/personal/resumes/{id}（编辑简历）

- **请求类型**：JSON；**鉴权**：个人（utype=1，仅本人）
- 请求体：同 #48（主表可编辑字段 + 6 子表数组）
- 行为：主表可编辑字段更新；6 子表**全量替换**（删除该简历下旧数据、按本次提交重建；`projects` 限 6 条，空数组=清空）；完善度与搜索索引重算
- 响应成功：`code=0`（data 空）

- 错误码：1001（简历不存在 / 非本人简历 / 已软删 / 项目经历超过 6 条）

---

## 52. DELETE /api/v1/personal/resumes/{id}（删除简历）

- **请求类型**：无参；**鉴权**：个人（utype=1，仅本人）
- 行为（04 §3.5 软删除）：主表 `deleted_at` 置当前时间戳；**子表保留**（合规审计）、投递/下载记录保留；
  若删掉的是默认简历（def=1），本人最近一份未删简历自动升为默认（无剩余则无默认，投递取最新）
- 删除后：#49 列表 / #50 详情 / #51 编辑 / #54 / #55 / #57 对该 id 均返回 1001；投递选简历（#61）不可再选中
- 响应成功：`code=0`（data 空）

- 错误码：1001（简历不存在 / 非本人简历 / 已软删）

---

## 54. PUT /api/v1/personal/resumes/{id}/display（公开/隐藏切换）

- **请求类型**：JSON；**鉴权**：个人（utype=1，仅本人）
- 请求体：`{ "display": 1|2 }`（1=公开 2=不公开，必填）

```json
{ "display": 2 }
```

- 响应成功：`code=0`（data 空）
- 错误码：1001（display 非法 / 简历不存在 / 非本人简历 / 已软删）

---

## 55. PUT /api/v1/personal/resumes/{id}/default（设为默认简历）

- **请求类型**：无参；**鉴权**：个人（utype=1，仅本人）
- 行为：同 uid 内 def **互斥事务**——先清本人所有 def=1，再置目标 def=1；投递（#61 resumeId=0）取 def=1 优先
- 响应成功：`code=0`（data 空）

- 错误码：1001（简历不存在 / 非本人简历 / 已软删）

---

## 57. GET /api/v1/personal/resumes/{id}/completeness（完善度详情）

- **请求类型**：无参；**鉴权**：个人（utype=1，仅本人）
- 行为：按 04 §2.2 权重计算（与主表 `completePercent` **同源**），返回分类得分 + 缺失字段清单

```json
{ "code": 0, "message": "success",
  "data": {
    "percent": 60,
    "categories": [
      { "key": "basic",     "name": "基本信息", "max": 20, "score": 10 },
      { "key": "education", "name": "教育经历", "max": 20, "score": 20 },
      { "key": "work",      "name": "工作",     "max": 25, "score": 15 },
      { "key": "intention", "name": "期望",     "max": 15, "score": 0  },
      { "key": "other",     "name": "其他",     "max": 20, "score": 15 }
    ],
    "missing": ["性别", "出生年", "期望地区", "期望薪资", "期望职位", "自我评价"] } }
```

- 错误码：1001（简历不存在 / 非本人简历 / 已软删）

### 完善度字段级权重（04 §2.2 落地细则，缺一项扣对应分）

| 类别 | 满分 | 明细（字段 → 分值） |
|---|---|---|
| basic 基本信息 | 20 | 姓名 5 / 性别 2 / 出生年 3 / 最高学历 5 / 专业 3 / 联系电话 2 |
| education 教育经历 | 20 | 教育经历子表 ≥1 条 → 20 |
| work 工作 | 25 | 工作经历子表 ≥1 条 → 15；项目经历子表 ≥1 条 → 10 |
| intention 期望 | 15 | 期望地区 5 / 期望薪资 5 / 期望职位 5 |
| other 其他 | 20 | 语言 5 / 培训 5 / 证书 5 / 自我评价 5 |

> 「其他」类设计含「作品」项；ms_resume_img 后置，其分值暂由「自我评价」代位，作品上线后重拆（总分保持 20）。
> `missing` 为未命中项的中文名称清单（全空简历 16 项）。

---

## 简历主表字段（ms_resume，01 §2.3，可编辑子集）

创建/编辑请求可携带以下主表字段（其余为服务端管理或随 #53-60 维护）：

| 字段 | 类型 | 说明 |
|---|---|---|
| title | string(80) | 简历标题 |
| fullname | string(15) | 姓名 |
| sex / sexCn | int8 / string | 性别 / 中文 |
| birthdate | uint16 | 出生年 |
| residence | string(30) | 籍贯 |
| education / educationCn | uint16 / string | 最高学历 / 中文 |
| major / majorCn | uint16 / string | 专业 / 中文 |
| experience / experienceCn | uint16 / string | 工作年限 / 中文 |
| district / districtCn | string | 期望地区 / 中文 |
| wage / wageCn | uint16 / string | 期望薪资 / 中文 |
| intentionJobs | string(255) | 期望职位 |
| specialty | string(1000) | 自我评价 |
| telephone | string(50) | 联系电话 |
| email | string(60) | 邮箱 |
| displayName | int8 | 1=显示姓名 2=匿名 |
| current / currentCn | uint16 / string | 目前状态 / 中文 |
| mobileAudit | int8 | 手机认证 |
| talent | int8 | 高级人才（默认 0） |
| entrust | int8 | 委托（默认 0） |

> 服务端管理字段：`complete_percent`（完善度，#48/#51 自动计算）、`key_full`/`key_precise`（搜索索引，#48/#51 自动拼装；一期文本拼装、二期迁 ES）、`def`（默认简历：#48 首份自动置 1 / #55 互斥切换 / #52 删默认自动升级）、`deleted_at`（#52 软删）、`display`（#54 切换）。
