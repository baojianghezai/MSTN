# web-front 开发状态锚点（前端会话自用）

> 维护者：前端会话（session-b95ebb4f-ed51-41b9-9521-52e7bc22ec83，房间别名「前端」）
> 用途：上下文重置后，读本文件恢复现场，不必重读全部代码
> 纪律：每轮工作收尾时更新「当前任务」与「模块完成图」；细节历史进 `aiDoc/memory/business/` 与 git，本文件只留「现在状态」
> 最后更新：2026-08-21 11:16（本批全部闭环：B1 三档 + X1/X2/X3 + D1 均 PM 验收；部署包 `MSTN人才网-内网部署包-20260821.zip`（48MB）已出，待派活方内网更新）

## 一、重置后恢复三步

1. 读本文件
2. 查 `aiDoc/frontend-backend/handoff/` 最新文件——当前任务依赖后端时，通读对应 handoff（**不猜接口**）
3. 探活服务：8888 后端 / 8081 前端（见「环境速启」）——机器常重启，两个都默认按死的处理

## 二、当前任务（全文替换式更新）

- **B1 美化 P0/P1/P2：三档全部完成报单，待 PM 复核**
  - P0：uno.config.ts 色板（primary/accent）+ styles/index.css EP 7 档覆写 + 11 文件 blue-*/orange-*→primary-*/accent-* 全局换色
  - P1：presetIcons+lucide 图标系统（首页区块/九宫格 9/企业按钮 5）+ hero 点阵纹理&双色氛围光&双层投影 + JobsCard 公司首字 logo 色块&薪资 accent-500&hover 上浮 + danger 色阶 token（red-* 清零）+ 获取验证码按钮主色描边（5 处）+ el-card 头层级 + 安全卡等高（266/266、282/282）
  - P2：空态 el-empty×5 + 骨架屏×7（首载 el-skeleton/二次 v-loading）+ resume-edit 长表单（锚点导航 sticky top-14 白胶囊/各卡 scroll-mt-32/smooth scroll + sticky 保存栏 + 子表删除二次确认）
  - **P2 重大环境坑修复**：本环境 UnoCSS presetWind3 preflight **缺 `border: 0 solid` 全局重置** → border-style=none → 所有 UnoCSS border-* 工具类边框不渲染（width 计算值强制 0）。已在 styles/index.css 补 `*,::before,::after{border:0 solid #e5e7eb}`。此前 3 布局 footer/九宫格/子表行/验证码描边全部不可见，补后按设计渲染（视觉回归 9/10 未变脏）
  - 视觉自查：首页 hero 8.5 / 个人中心 9（border 修复后复测）/ 企业 7.5 / P2 编辑页锚点+保存栏 9（终检）；头像/注册时间未做（#10 /auth/me 一期仅返 uid/utype，无数据源）
- **修改批 X（派活方三处拍板，16:40 派单）：X1/X2/X3 前端全部完成报单，待 PM E2E**
  - X1 表单裸 0→占位（已报）：4 文件 43 个数字 el-select 加 `:empty-values="[null,undefined,0]"`（EP 2.13 原生 prop，选项 id 均非 0 才可用）+ 各 handleSave 提交前 toNum 归一 undefined→0（后端契约 0=空 不变）；`x1-verify.cjs` 13/13（含空态提交 payload 全 0 code=0）；视觉 10/10 + 8/10
  - X2 已报：jobs.vue 发布成功提示「职位已发布」→「已提交，待审核」（编辑仍「已提交保存」）；实测发布→toast 正确、列表出现审核中、前台搜索不可见（tmp 路径生效）
  - X3 已报：职位名称 el-input→el-select（filterable、`:allow-create="false"` 纯选择禁自由输入、default-first-option；options=categories.jobtitle #25 返回 50 岗；编辑回显历史名加 `historyTitleOption` 虚拟 option；空数组兜底提示保留，数据就位后已消失）；types/api.ts 加 `Categories.jobtitle?`；`x2x3-verify.cjs` 16/16 含级联三级冒烟（10 一级/81 行）
  - 薪资 2 倍校验（PM 08-21 裁决=做，已补）：jobs.vue handleSave `minwage>0 && maxwage>minwage*2 → 提示「最高薪资不能超过最低薪资的 2 倍」拦截`（口径略严于后端整数除法规则，用户永不撞后端错）；`x2x3-verify.cjs` 19/19（+X2-W1/W2/W3 拦截断言：文案/弹窗保留/未发 POST）
  - Issue A 后端结论=无漂移（14:0x 401 系探测踩坑，4 个坑已沉淀进关键记忆），已按更正调整认知；真实登录路径继续用（最稳）
- **当前状态：本批全部闭环（08-21 11:15 PM 收官），前端无在途**。验收总账：M3 简历子表 ✅ / B1 P0/P1/P2 ✅（含 UnoCSS border 坑修复）/ X1/X2/X3 ✅ / D1 抓到→修复→复验 18/18 闭环 ✅；部署包 `C:\Users\18089\Desktop\MSTN\MSTN人才网-内网部署包-20260821.zip`（48MB，后端 D1 版 exe + 双前端新构建 + gva.sql + 部署说明含 5 条验证清单）已出，待派活方内网更新。遗留观察项（不阻塞、后续批次）：九宫格「换绑/绑定手机」文案待派活方拍板；B1 建议级微调（锚点激活态/空态 action）
  - **D1（PM E2E 15/18 抓到，阻塞部署包）已闭环**：后端 08-21 10:47 报单修复（#82/#83 认 jobs_tmp：pending=1 显式走 tmp 重提/软删，缺省 jobs 未命中自动回退，连带 #80 补 tmp deleted_at=0 过滤；8888 换新构建）。前端 UI 复验 `.local/d1-verify.cjs` **10/10 全绿**（建 tmp 职位→列表审核中→编辑重提原 1001 现通过→回显新名→删除原 1001 现通过→列表消失→前台不可见）
  - 前端 D1 配合改动（后端建议）：api/jobs.ts `updateJob(id,data,pending?)` / `deleteJob(id,pending?)` 显式 `?pending=1`；jobs.vue 新增 `editPending` ref（openEdit 置 row.pending / openCreate 复位），保存/删除均显式透传（getJob 原已透传）；vue-tsc 0 错误
  - 测试数据清零：复验 tmp 行 id=13（uid20）+ contact 孤儿已硬删，ms_jobs_tmp 空，ms_jobs 仅留 fixture 1-3
  - PM 侧配套（后端已代修）：pm-x-accept.cjs X2.10 竞态（T0 小数 vs addtime 整秒）改 `Math.floor(Date.now()/1000)-1`
  - 自查修复通报（已闭环）：P2 时误删 jobs.vue `<template #empty>` 开标签（孤儿闭标签；vue-tsc 0 但 vite:vue 编译 500）。教训：**模板改动必须真实加载页面验证**
- ⚠️ **服务状态（08-21 机器重启后）**：8888 由 **GoLand 调试实例**（`___go_build_gin_vue_admin_server.exe`，pid 12016，9:14 起）接管——后端/用户从 GoLand 起的；8081 vite = job pwsh-2；**验证 token 一律测试账号密码登录**（派活方 08-21 指令，PM msg-mt2bytoe-6）：账号 `C:\Users\18089\Desktop\MSTN\测试账号.txt`（个人 13300001111=uid23 / 企业 13300009999=uid20，密码 123456），`POST /api/v1/auth/login {account,password}` 拿 7 天 token，`.local/m3-auth.cjs` 已切此方式（旧号按 utype 自动映射，缓存 m3-tokens.json，探活 #10 /auth/me）；**SMS mock/伪造 token 弃用**；测试账号是演示号——不改其密码/资料，测试数据收尾清理
- 等待节奏：pwsh 后台 `Start-Sleep -Seconds 1800` + `job_output wait`；不高频轮询；不挂大 goal

## 三、模块完成图（web-front）

| 模块 | 状态 | 落点 |
|---|---|---|
| A 职位发布/管理（企业） | ✅ PM 已验收 | views/company/jobs.vue（jobs/jobs_tmp 双表，getJob(id, pending) 消歧） |
| B 职位浏览（公开列表/详情） | ✅ PM 已验收 | views/jobs/list.vue、detail.vue |
| C 投递（个人） | ✅ PM 已验收 | detail.vue 投递弹窗（resumeId=0 默认简历）+ personal/applies.vue |
| D 收简历（企业） | ✅ PM 已验收 | views/company/applies.vue |
| 认证（登录/注册/重置/SMS） | ✅ | views/auth/* |
| 个人资料 / 申诉 | ✅ | personal/profile.vue、appeal/index.vue |
| 视频面试（N2） | ✅ | personal/company video-interviews + room 页 |
| 简历创建/编辑（含项目经历 N1，限 6 条） | ✅ | personal/resume-edit.vue |
| 骨架完善（08-19） | ✅ 已验收 | 首页/404/4 占位页/PersonalLayout/企业工作台数据卡/JobsCard/列表城市筛选+深链/注册按钮修复 |
| 简历子表（M3 收尾） | ✅ PM 已验收（前后端全过） | personal/resumes.vue（#49 列表+操作）+ resume-edit.vue 6 子表表单 + 路由/导航 |
| B1 美化 P0/P1/P2 | ✅ PM 已验收（18/18 内含） | P0 色板+EP 同步+全局换色；P1 图标系统/hero 质感/JobsCard logo/danger token/验证码按钮/卡片层级/安全卡等高；P2 空态×5/骨架×7/长表单锚点+sticky 保存栏+删除确认/**border preflight 缺 border:0 solid 修复** |
| 修改批 X1（表单裸 0→占位） | ✅ PM 已验收（X1 零值契约 code=0） | 4 文件 43 个数字 el-select `:empty-values` + 提交 toNum 归一 |
| 修改批 X2/X3 前端 | ✅ PM 已验收（X2 全链路+X3 50/81 行） | X2 发布提示「已提交，待审核」；X3 职位名 el-select（jobtitle 50 岗、纯选择、历史值虚拟 option）；x2x3-verify 19/19（含薪资 2 倍校验） |
| D1 复验（#82/#83 认 jobs_tmp） | ✅ PM 已验收（18/18 内含 D1 修复点） | jobs.vue+api/jobs.ts pending 显式透传（updateJob/deleteJob/editPending）；d1-verify.cjs 10/10；测试数据清零 |

后置（明确不在当前批）：
- 简历：#53 刷新 / #56 复制 / #58 附件 / #59 照片 / #60 发邮箱 / 作品照片 ms_resume_img → 下一批
- 公开域：找企业/招聘会/资讯/帮助 = coming-soon 占位页（等 M3/M6 后端）
- 商业化：套餐购买/收银台（M5）、下载扣费 #95、敏感词/置顶购买

## 四、环境速启

| 项 | 命令/位置 | 备注 |
|---|---|---|
| 前端 dev | `cd msrc-admin/web-front; npm run dev`（pwsh run_in_background） | 8081；vite proxy：/api/v1、/uploads → 127.0.0.1:8888（保留前缀，不能 rewrite） |
| 后端 | `Start-Process msrc-admin\server\server.exe -WorkingDirectory msrc-admin\server` | 8888；config.yaml 在 server/ 下，必须带 WorkingDirectory |
| Redis | 127.0.0.1:6379 无密码 | member 会话键 `hrc:member:session:<uid>` |
| 测试 token | **测试账号密码登录** `.local/m3-auth.cjs`（`getRealToken(accountOrMobile, utype)`，缓存 m3-tokens.json，探活 #10 /auth/me） | 派活方 08-21 指令：个人 13300001111=uid23 / 企业 13300009999=uid20（密码 123456），`POST /api/v1/auth/login {account,password}` 7 天 token；旧号（13900001111 等）按 utype 自动映射到双演示号；**SMS mock/伪造 token 弃用**；演示号不改密码/资料、测试数据收尾清理 |
| 前端登录态 | localStorage：ms_token / ms_utype / ms_uid / ms_mobile | 与 GVA 后台 token 隔离（design/10 §1.3） |

## 五、验证约定

- 类型：`npm run type-check`（vue-tsc --noEmit）必过
- 点触冒烟：`.local/skeleton-smoke.cjs`（17 断言：导航/深链/404/占位/两中心/九宫格）+ `.local/layout-check.cjs`（8 页横向溢出）
- Playwright 坑：`npx playwright@1.62.1` 默认要 chromium-1234，机器装的是 **chromium-1228**——必须 `chromium.launch({ executablePath: 'C:\\Users\\18089\\AppData\\Local\\ms-playwright\\chromium-1228\\chrome-win64\\chrome.exe' })`
- 视觉：**本会话自查**（派活方 08-20 定调，助理退出视觉验收）——本模型无图像输入（read_image/describe_image 不可用），用 `.local/vision-check.ps1 -Image <png> -Prompt "..."` 直连 seetacloud 视觉通道（enable_thinking:false，ASCII-only 脚本）；截图随报单附件发，PM 验收照旧
- 业务记忆：`aiDoc/memory/business/{active,done}/` 一个功能点一个文件 + `demand-index.md` 登记（AGENTS.md 规则）；本文件不重复细节

## 六、协作协议

- 房间：`room-mswo3ooq-1`（msrc-admin-三人协作）；de_broadcast recipients 传 `room:room-mswo3ooq-1`
- 角色：PM session-fe18…（派单/验收，无视觉）/ 后端 session-2d8d…（handoff 作者）/ 派活方=用户（拍板，传话筒）
- 流程：后端 handoff → 前端读单做页（不猜接口）→ 报单带 handoff 文件名+截图 → PM E2E 验收
- 节奏：被唤醒照单连续干→报单→停；等待用后台 sleep+wait；不挂大 goal（空转烧 token）

## 七、文件地图（web-front/src/）

- 布局：PublicLayout（公开：导航来自 /navigations 有兜底、顶栏搜索）/ PersonalLayout（个人：顶栏 5 导航）/ CompanyLayout（企业：工作台域、返回首页）
- `router/index.ts`：按布局域组织；`router/guards.ts`：登录+utype 守卫（meta 父级继承）+ document.title 同步
- `utils/request.ts`：hrc `{code,message,data}`，401 清 token 跳登录；分页 `{page,pageSize,total,list}`
- `types/api.ts` 全部类型；`api/*.ts` 按模块封装
- `views/`：home / jobs / auth / appeal / personal / company / placeholder（coming-soon）/ error（404）
- `components/JobsCard.vue`：首页+列表共用职位卡片
- 样式：UnoCSS 原子类优先（AGENTS.md 前端样式规则）
