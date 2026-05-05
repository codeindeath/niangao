# 年糕 App 产品需求文档（PRD）

> **版本**：v2.1 | **日期**：2026-05-05 | **状态**：创作者/价值度星级/平台经验/名人种子/无限滚动/经验删除 全部完成

---

## 一、产品概述

### 1.1 产品定位
年糕是一款**结构化经验分享 + AI 辅导社区**。用户发布 ≤100 字的精炼经验，平台通过「硬策略 + AI」双层审核确保内容质量，基于领域偏好进行个性化推荐，并提供 AI 对话辅导。

### 1.2 目标用户
中国 20-35 岁职场青年，有成长意愿。

### 1.3 核心差异化

| 维度 | 传统社区 | 年糕 |
|------|---------|------|
| 内容形式 | 长文 / 短视频 | **≤100 字结构化经验** |
| 质量保障 | 人工审核 | **硬策略 + AI 双层审核** |
| 分发机制 | 时间线 / 关注 | **领域偏好加权推荐** |
| AI 应用 | 辅助写作 | **准入分析 + 6 维打分 + 对话辅导** |

---

## 二、产品功能全景

```
年糕 App
│
├─ 🔐 登录体系
│   ├─ Apple Sign In（唯一生产登录方式）
│   ├─ Dev Login（开发环境一键登录）
│   ├─ JWT Token 管理（生成 / 刷新 / 过期）
│   └─ 用户信息持久化（AsyncStorage）
│
├─ 🏠 首页（经验流）
│   ├─ 推荐流 ─ 登录用户：GET /recommend（领域偏好加权）
│   ├─ 热门流 ─ 未登录用户：GET /experiences（热度排序）
│   ├─ 下拉刷新
│   ├─ 点赞 / 取消点赞（即时 UI + API）
│   ├─ 收藏 / 取消收藏（即时 UI + API）
│   └─ 点击进入详情页
│
├─ ✏️ 发布经验
│   ├─ 内容输入（10-100 字，实时字数统计）
│   ├─ 一级领域选择（5 个）
│   ├─ 二级领域选择（选择一级后展示，共 21 个）
│   ├─ AI 生成解读（可选，调用 /ai generate-interpretation）
│   ├─ 私密开关（默认公开）
│   ├─ 公开经验 → 硬策略检查 → AI 审核+打分 → 入库
│   └─ 私密经验 → 直接入库（跳过审核）
│
├─ 📋 经验详情页
│   ├─ 创作者信息（头像、名称）
│   ├─ 平台标签「官」+ 来源（平台生产经验）
│   ├─ 领域标签（优先显示二级领域）
│   ├─ 经验内容
│   ├─ 价值度 ★★★★★ + 点击弹出打分理由（≤15字）
│   ├─ 审核状态：未通过 → 灰色提示卡「这条经验未通过审核，仅你自己可见」
│   ├─ AI 解读（如有）
│   ├─ 点赞 / 收藏操作
│   └─ 删除（仅作者可见，已入池经验软删除，其他硬删除）
│
├─ 👤 个人中心
│   ├─ 用户头像 + 昵称
│   ├─ Tab: 我的经验（含私密 🔒 + 审核中 ⏳ + 已拒绝 ❕）
│   ├─ Tab: 我的收藏
│   ├─ 退出登录
│   └─ 未登录 → "去登录"引导按钮
│
├─ 🤖 AI 对话
│   ├─ 基于用户经验库的人本主义辅导
│   ├─ 多轮对话（上下文记忆）
│   ├─ 引用经验 ID
│   └─ 每日对话轮次限制（50 轮）
│
├─ 🔍 搜索
│   ├─ 关键词搜索经验
│   └─ 分页加载
│
└─ 🧠 推荐引擎（后端）
    ├─ 领域偏好计算（发布×2 + 收藏×1）
    ├─ 热度加权（like + bookmark + 1）
    ├─ 排除已读 + 私密 + 未审核
    └─ 新用户冷启动 → 回退纯热度
```

---

## 三、核心流程（流程图）

### 3.1 经验发布完整流程

```
用户点击「发布经验」
        │
        ▼
┌─ 填写内容（10-100字）──┐
│  选择一级领域（必选）    │
│  选择二级领域（必选）    │
│  (可选) AI 生成解读      │
│  私密开关（默认公开）    │
└────────────────────────┘
        │
        ▼
   【私密？】─── 是 ──→ 直接入库
        │              review_status = "private"
        │              自动收藏
        │              返回创建成功
        否
        │
        ▼
   【硬策略检查】
   ├─ 字数 ≥ 10 字？
   ├─ 包含汉字/字母？
   ├─ 敏感词过滤？
   └─ 重复字符 ≤ 10？
        │
    不通过 ──→ 入库 (review_status="rejected")
        │      自动收藏
        │      返回拒绝原因 + 提示
        │
     通过
        │
        ▼
   【AI 审核】(DeepSeek API, 30s 超时)
   ├─ 合规性检查
   ├─ 内容类型判断（经验 vs 纯知识）
   └─ 6 维度打分
        │
   ┌────┼────┐
   ▼    ▼    ▼
 通过  拒绝  超时/失败
   │    │    │
   │    │    └─→ 入库 (review_status="pending")
   │    │         等待异步重试
   │    │
   │    └─→ 入库 (review_status="rejected")
   │         保存拒绝原因
   │         自动收藏
   │         返回提示："已保存但未进入平台池"
   │
   └─→ 入库 (review_status="approved")
        保存 6 维度评分
        自动收藏
        进入平台经验池
        返回成功 + 评分
```

### 3.2 首页加载流程

```
用户打开 App
        │
        ▼
   检查登录状态
   （AsyncStorage Token）
        │
   ┌────┴────┐
   ▼         ▼
 已登录    未登录
   │         │
   ▼         ▼
 GET        GET
 /recommend /experiences
   │         │
   ▼         ▼
 个性化推荐  热门精选
 (偏好加权)  (热度排序)
   │         │
   ▼         ▼
 标题:"为你推荐"  标题:"为你推荐"
 副标题:"基于你的  副标题:"热门经验精选"
        偏好·个性
        化推荐"
```

### 3.3 推荐引擎计算流程

```
用户请求推荐
        │
        ▼
 ┌─ 统计用户领域偏好 ─────────────────────┐
 │                                         │
 │  SELECT domain, COUNT(*)*2              │
 │  FROM experiences                       │
 │  WHERE author_id = $user_id             │
 │  GROUP BY domain                        │
 │  UNION ALL                              │
 │  SELECT e.domain, COUNT(*)              │
 │  FROM bookmarks b                       │
 │  JOIN experiences e ON e.id = b.exp_id  │
 │  WHERE b.user_id = $user_id             │
 │  GROUP BY e.domain                      │
 │                                         │
 │  无历史 → 所有权重 = 1（冷启动）         │
 └─────────────────────────────────────────┘
        │
        ▼
 ┌─ 查询候选经验池 ───────────────────────┐
 │                                         │
 │  WHERE status = 'published'             │
 │    AND review_status = 'approved'       │
 │    AND is_private = FALSE               │
 │    AND author_id != $user_id            │
 │    AND id NOT IN (用户已收藏)            │
 │                                         │
 │  排序: 领域权重 × (like+bkmk+1) DESC    │
 │  LIMIT 20                               │
 └─────────────────────────────────────────┘
        │
        ▼
   返回 Top 20
```

### 3.4 点赞 / 收藏交互流程

```
用户点击 ♥ / ★
        │
        ▼
  乐观更新 UI（立即切换状态）
        │
        ▼
  POST /api/v1/experiences/:id/like
  或
  POST /api/v1/experiences/:id/bookmark
        │
   ┌────┴────┐
   ▼         ▼
 成功      失败
   │         │
   ▼         ▼
 UI 保持   UI 回滚
           （恢复原状态）
```

### 3.5 登录流程

```
用户点击登录
        │
        ▼
  [生产环境]          [开发环境]
  Apple Sign In        Dev Login
  系统弹窗认证         输入昵称
        │                 │
        ▼                 ▼
  获取 identity_token   POST /auth/dev/login
        │                 │
        ▼                 ▼
  POST /auth/apple/login
        │
        ▼
  后端验证 Apple Token
  创建/查找用户
  生成 JWT + Refresh Token
        │
        ▼
  移动端存储:
  - auth_token (AsyncStorage)
  - user_info (AsyncStorage)
        │
        ▼
  跳转首页
```

---

## 四、功能详细描述

### 4.1 登录体系

| 功能点 | 详细说明 |
|--------|---------|
| 登录方式 | 生产环境仅 Apple Sign In；开发环境提供 Dev Login（输入昵称即可） |
| Token 管理 | JWT (HS256)，24h 过期；Refresh Token 支持刷新 |
| 存储 | AsyncStorage：`auth_token` + `user_info`（nickname, id 等） |
| 自动登录 | App 启动时读取 AsyncStorage，有 Token 则自动恢复登录态 |
| 登出 | 清除 AsyncStorage 中的 token 和 user_info，UI 回退到未登录状态 |
| 鉴权中间件 | Go 后端 `AuthMiddleware` 解析 JWT，注入 `user_id` 到 Gin Context |
| 强制认证 | `RequireAuth` 中间件：`user_id` 不存在 → 返回 401 `{"error":"请先登录"}` |

### 4.2 首页经验流

| 功能点 | 详细说明 |
|--------|---------|
| 数据源 | 登录 → `/recommend?limit=20&offset=N`；未登录 → `/experiences?page=N&page_size=20&sort=latest` |
| 无限滚动 | `onEndReached` 触发自动加载下一页，offset 递增 20，底部 loading 指示器 |
| 到底提示 | 数据全部加载完毕显示「— 已经到底了 —」 |
| Loading 态 | 居中 ActivityIndicator，深绿色 (`#4a7c59`) |
| 空状态 | "暂无推荐内容" + "发布经验后，推荐会更精准" |
| 错误态 | "加载失败，请检查网络连接" + 重试按钮 |
| 下拉刷新 | RefreshControl，重新请求推荐/列表（offset 重置为 0） |
| 点赞 | 乐观更新 UI（即点即变），API 失败则回滚 |
| 收藏 | 同上；已收藏显示"★ 已收藏"（金色 `#e8a850`） |
| 卡片信息 | 头像(首字母) + 创作者名称 + 领域标签 + 内容 + 价值度星级 + 互动按钮 |
| 创作者 | 优先 `creator_name`，其次 `author_name`，兜底「匿名」 |
| 领域标签 | 优先显示二级领域名（如"时间管理"），未知则回退一级（如"生活智慧"） |
| 平台标签 | `source_type='platform'` 时显示绿色圆形「官」标签 |
| 审核状态 | 审核通过 → 不展示任何标记；审核未通过 → 灰色 ❕ 感叹号 |
| 私密标记 | `is_private=true` 时显示 🔒 标记（仅"我的经验"中出现） |
| 价值度 | `quality_score` 映射 0-5 星（score/2 四舍五入），点击弹出 `<score_reason>`（≤15 字简短理由） |

### 4.3 发布经验

#### 4.3.1 表单字段

| 字段 | 类型 | 必填 | 限制 | 说明 |
|------|------|------|------|------|
| content | TextInput | ✅ | 10-100 字 | 经验内容，实时显示字数 X/100 |
| domain | 选择器 | ✅ | 5 选 1 | 一级领域：职场/关系/认知/生活/情感 |
| sub_domain | 选择器 | ✅ | 4-5 选 1 | 二级领域：选择一级后展示 |
| interpretation | TextInput | ❌ | ≤500 字 | AI 解读（可手动编辑或 AI 生成） |
| is_private | Switch | ❌ | 默认 false | 私密开关 |

#### 4.3.2 审核反馈

| 结果 | HTTP | 返回体 | 用户看到 |
|------|------|--------|---------|
| 通过 | 201 | `{experience, review:{status:"approved", score, message}}` | "发布成功" |
| 硬策略拒绝 | 201 | `{experience, review:{status:"rejected", reason, message}}` | "已保存但不符合准入规则" |
| AI 拒绝 | 201 | `{experience, review:{status:"rejected", reason, message}}` | "已保存但未通过 AI 审核" |
| AI 超时 | 201 | `{experience, review:{status:"pending", message}}` | "已保存，审核中" |
| 未登录 | 401 | `{error:"请先登录"}` | "请先登录后再发布经验" |
| 校验失败 | 400 | `{error:"..."}` | 具体字段错误提示 |

#### 4.3.3 经验删除

| 场景 | 行为 | 说明 |
|------|------|------|
| 审核通过(approved) + 作者删除 | **软删除** | 设置 `deleted_at`，用户个人列表不可见，平台经验池继续存在 |
| 未通过(rejected)/审核中(pending)/私密(private) + 作者删除 | **硬删除** | 从数据库物理删除 |
| 非作者删除 | **拒绝** | 返回 403 |

### 4.4 个人中心

| 功能点 | 详细说明 |
|--------|---------|
| 头像 | 圆形，首字母，深绿底色 |
| Tab 切换 | "我的经验 (N)" / "我的收藏 (M)"，绿色下划线激活态 |
| 我的经验 | 展示所有状态的经验（含 rejected/pending/private） |
| 我的收藏 | 仅展示 approved + private 的已收藏经验 |
| 空状态 | "暂无经验" / "暂无收藏" |
| 退出登录 | 红色文字按钮，点击后清除 token + user_info |
| 未登录 | 显示绿色"去登录"按钮 |

### 4.5 AI 对话

| 功能点 | 详细说明 |
|--------|---------|
| 触发 | 底部 Tab 独立入口 |
| 模型 | DeepSeek API |
| 上下文 | 用户已发布经验 + 历史对话（最近 5 条） |
| 辅导风格 | 人本主义（以人为本、非评判、启发式） |
| 引用 | AI 回复可引用具体经验 ID |
| 限制 | 每日 50 轮对话（可配置） |
| 用户 ID | 从 `getUserInfo()` 动态获取（非硬编码） |

### 4.6 搜索

| 功能点 | 详细说明 |
|--------|---------|
| 输入 | 文本输入框 |
| 接口 | `GET /experiences?search=关键词&page=N` |
| 匹配 | ILIKE 模糊匹配 content + interpretation |
| 分页 | 20 条/页，上拉加载更多 |

---

## 五、领域标签体系

### 5.1 结构

```
领域体系（2 级，5 × 4-5 = 21 个叶子节点）

career 职场成长
├── career-planning  职业规划
├── skill-building   技能提升
├── side-hustle      副业创业
└── workplace-comm   职场沟通

relationship 人际关系
├── intimate         亲密关系
├── family           家庭关系
├── social-skill     社交技巧
└── communication    沟通表达

cognition 认知升级
├── mental-model     思维模型
├── learning         学习方法
├── decision         决策判断
└── psychology       心理认知

life 生活智慧
├── finance          理财规划
├── health           健康养生
├── time-mgmt        时间管理
├── habits           习惯养成
└── digital-life     数字生活

emotion 情绪情感
├── regulation       情绪调节
├── self-growth      自我成长
├── happiness        幸福感
└── stress-mgmt      压力管理
```

### 5.2 UI 交互

```
创建经验时：
  1. 展示 5 个一级领域 Chip（如"职场成长""人际关系"...）
  2. 用户点击某个一级 → 下方展开该领域下的二级 Chip
  3. 用户点击二级 → 选中（深绿色高亮）
  4. 再次点击同一一级 → 取消选择（清空二级）

首页展示时：
  优先显示二级领域名（如"时间管理"）
  二级领域未知时回退一级（如"生活智慧"）
```

---

## 六、经验池准入体系

### 6.1 硬策略（第一层，Go 代码，毫秒级）

```
CheckHardPolicy(content)

Rule 1: 内容长度
  utf8.RuneCountInString ≥ 10
  不通过 → "内容过短，请至少输入10个字符"

Rule 2: 有意义内容
  正则 [\p{Han}a-zA-Z] 至少匹配一次
  不通过 → "内容不能全是数字或符号"

Rule 3: 敏感词过滤
  正则匹配 7 类关键词（赌博/色情/暴力/毒品/传销/高利贷/政治）
  不通过 → "内容包含不适当的关键词"

Rule 4: 重复字符
  hasExcessiveRepetition() 检测 >10 个连续相同字符
  不通过 → "内容包含过多重复字符"
```

### 6.2 AI 审核（第二层，DeepSeek API，30s 超时）

```
callAIReview(content, domain, sub_domain)
  │
  ├─ 合规性检查
  │   是否符合中国法律法规？
  │   是否安全、正向、无不良引导？
  │
  └─ 内容类型判断
      是「可指导行动的经验」？
      还是「单纯的知识点」（公式/定义/事实罗列）？
      前者 approved，后者 rejected
```

### 6.3 审核状态机

```
                    ┌──────────┐
       私密经验 ──→  │ private  │  (跳过一切审核)
                    └──────────┘

                    ┌──────────┐
       硬策略失败 ─→ │ rejected │  (reason = 硬策略原因)
                    └──────────┘

                    ┌──────────┐
       AI 拒绝   ─→ │ rejected │  (reason = AI 原因)
                    └──────────┘
       公开经验 ─┐
                 ├→ 硬策略
                 │    │
                 │   通过
                 │    │
                 ├→ AI 审核 ─→ ┌──────────┐
                 │    │  通过  │ approved │ → 进入平台池
                 │    │        └──────────┘
                 │    │
                 │    ├── 拒绝 → rejected
                 │    │
                 │    └── 超时 → ┌─────────┐
                 │              │ pending  │ → 等待异步重试
                 │              └─────────┘
```

---

## 七、AI 质量打分

### 7.1 评分维度

```
          综合分 overall (加权)
         ┌──────────────────────┐
         │   value        25%   │  内容价值度：是否带来新认知/启发？
         │   actionable   25%   │  实操可执行度：读者能否直接应用？
         │   universal    20%   │  普适性：多少人能受益？
         │   original     15%   │  原创性：独到见解 vs 老生常谈？
         │   clarity      15%   │  清晰度：表达是否简洁明了？
         └──────────────────────┘
         每个维度 0-10 分，保留一位小数
```

### 7.2 与审核合并调用
AI 审核和打分在同一次 DeepSeek API 调用中完成，减少延迟。

---

## 八、推荐系统

### 8.1 算法公式

```
推荐分(R) = 领域偏好权重(W_domain) × 热度分(H)

W_domain = SUM(该领域发布次数 × 2) + SUM(该领域收藏次数 × 1)
（无任何历史 → 全领域 W_domain = 1）

H = like_count + bookmark_count + 1

排除条件:
  - 自己发布的 (author_id = 当前用户)
  - 已收藏的   (id IN 用户 bookmarks)
  - 未审核通过 (review_status != 'approved')
  - 私密的     (is_private = TRUE)

排序: R DESC, LIMIT 20
```

### 8.2 示例

```
用户 A:
  发布: career×1, life×1
  收藏: career×2, cognition×1

领域偏好:
  career    = 1×2 + 2×1 = 4
  life      = 1×2 + 0×1 = 2
  cognition = 0×2 + 1×1 = 1
  其他      = 1 (fallback)

候选经验:
  经验 X (career, likes=3, bkmk=2) → 4 × (3+2+1) = 24 ✅ 最高
  经验 Y (life, likes=1, bkmk=1)   → 2 × (1+1+1) = 6
  经验 Z (emotion, likes=5, bkmk=5) → 1 × (5+5+1) = 11
```

### 8.3 冷启动
新用户无发布、无收藏 → 所有领域 W=1 → 纯热度排序（等价于热门流）。

---

## 九、私密经验

```
┌──────────────────────────────────────────────┐
│              公开经验 vs 私密经验               │
├──────────────┬───────────────┬───────────────┤
│     维度      │    公  开     │    私  密     │
├──────────────┼───────────────┼───────────────┤
│ 硬策略检查    │      ✅       │      ❌       │
│ AI 审核       │      ✅       │      ❌       │
│ 质量打分      │      ✅       │      ❌       │
│ 进入平台池    │      ✅       │      ❌       │
│ 被推荐        │      ✅       │      ❌       │
│ 我的经验可见  │      ✅       │      ✅       │
│ 我的收藏可见  │      ✅       │      ✅       │
│ UI 标记       │      无       │   🔒 私密     │
│ review_status │ approved/     │   private     │
│              │ rejected/     │               │
│              │ pending       │               │
└──────────────┴───────────────┴───────────────┘
```

---

## 十、推送通知（规划中）

| 场景 | 触发条件 | 推送内容 |
|------|---------|---------|
| 审核完成 | pending → approved/rejected | "你的经验已通过审核" / "审核未通过，点击查看原因" |
| 获得互动 | 经验被点赞/收藏 | "有 N 人觉得你的经验有用" |
| 新推荐 | 每日定时 | "今日为你推荐 N 条新经验" |

---

## 十一、技术架构

### 11.1 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                      iOS App                            │
│              React Native + Xcode 原生                   │
│    AsyncStorage (token, user_info)                      │
└──────────────┬──────────────────────┬───────────────────┘
               │                      │
          Nginx :80                    │
               │                      │
       ┌───────┴───────┐      ┌──────┴──────┐
       │  /api/*        │      │  :8000 直连  │
       ▼                │      ▼              │
┌──────────────┐        │ ┌──────────┐       │
│ Go Backend   │        │ │ Python AI│       │
│ Gin + pgx    │        │ │ FastAPI  │       │
│ :8080        │        │ │ :8000    │       │
└──────┬───────┘        │ └────┬─────┘       │
       │                │      │             │
       ▼                │      ▼             │
┌──────────────┐        │ ┌──────────┐       │
│ PostgreSQL   │        │ │ DeepSeek │       │
│ 14           │        │ │ API      │       │
└──────────────┘        │ └──────────┘       │
                        │                    │
  ECS: 115.190.177.146  │  2核8GB, Ubuntu 22.04
```

### 11.2 技术栈

| 层 | 技术 | 说明 |
|----|------|------|
| 移动端 | React Native 0.76 + TypeScript | Xcode 原生构建 |
| 后端 | Go 1.22 + Gin + pgx/v5 | REST API |
| AI 服务 | Python 3.10 + FastAPI | 对话 + 审核 + 打分 |
| 数据库 | PostgreSQL 14 | JSONB + ENUM + 触发器 |
| 部署 | systemd + Nginx | Go/Python 双服务 |
| 鉴权 | JWT (HS256) + Apple Sign In | 24h 过期 + Refresh |

### 11.3 完整 API 参考

| 方法 | 路径 | 功能 | 认证 | 请求体/参数 |
|------|------|------|------|-----------|
| POST | `/api/v1/auth/apple/login` | Apple 登录 | ❌ | `{identity_token, full_name?}` |
| POST | `/api/v1/auth/dev/login` | 开发登录 | ❌ | `{nickname?}` |
| POST | `/api/v1/auth/refresh` | 刷新 Token | ❌ | `{refresh_token}` |
| GET | `/api/v1/experiences` | 经验列表 | ❌ | `?domain=&sub_domain=&sort=&page=&page_size=&search=` |
| GET | `/api/v1/experiences/:id` | 经验详情 | ❌ | — |
| POST | `/api/v1/experiences` | 创建经验 | ✅ | `{content, domain, sub_domain, interpretation?, is_private?}` |
| PUT | `/api/v1/experiences/:id` | 编辑经验 | ✅ | `{content, domain, interpretation?}` |
| DELETE | `/api/v1/experiences/:id` | 删除经验 | ✅ | 已入池(approved) → 软删除(set deleted_at)；其他 → 硬删除 |
| POST | `/api/v1/experiences/:id/like` | 点赞/取消 | ✅ | — |
| POST | `/api/v1/experiences/:id/bookmark` | 收藏/取消 | ✅ | — |
| GET | `/api/v1/experiences/recommend` | 个性推荐 | ✅ | `?limit=20&offset=0`（offset 分页） |
| GET | `/api/v1/me/experiences` | 我的经验 | ✅ | `?page=&page_size=` |
| GET | `/api/v1/me/bookmarks` | 我的收藏 | ✅ | `?page=&page_size=` |
| POST | `/api/v1/review` | AI 审核+打分 | ❌ | `{content, domain, sub_domain}` |
| POST | `/api/v1/chat/send` | AI 对话 | ❌ | `{message, user_id, history?, conversation_id?}` |
| POST | `/api/v1/chat/generate-interpretation` | AI 生成解读 | ❌ | `{content, domain}` |
| GET | `/health` | 健康检查 | ❌ | — |

### 11.4 数据库触发器

```
likes 表:
  AFTER INSERT → experiences.like_count + 1 (update_like_count)
  AFTER DELETE → experiences.like_count - 1

bookmarks 表:
  AFTER INSERT → experiences.bookmark_count + 1 (update_bookmark_count)
  AFTER DELETE → experiences.bookmark_count - 1
  AFTER INSERT → users.bookmark_count + 1     (update_user_bookmark_count)
  AFTER DELETE → users.bookmark_count - 1

experiences 表:
  AFTER INSERT → users.experience_count + 1   (update_experience_count)
  AFTER DELETE → users.experience_count - 1

practiced（bookmarks.practiced 变更）:
  AFTER UPDATE/INSERT → users.practiced_count ± 1 (update_practiced_count)
```

---

## 十二、完整数据模型

### 12.1 users 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID PK | — |
| nickname | VARCHAR(30) | 昵称 |
| avatar_url | TEXT | 头像 URL |
| bio | VARCHAR(200) | 个人简介 |
| apple_user_id | VARCHAR(255) UNIQUE | Apple 用户标识 |
| wechat_openid | VARCHAR(128) NULLABLE | 已废弃（保留列） |
| experience_count | INT DEFAULT 0 | 经验数（触发器） |
| bookmark_count | INT DEFAULT 0 | 收藏数（触发器） |
| practiced_count | INT DEFAULT 0 | 实践数（触发器） |
| created_at | TIMESTAMPTZ | — |
| updated_at | TIMESTAMPTZ | — |

### 12.2 experiences 表

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| id | UUID PK | uuid_generate_v4() | — |
| author_id | UUID FK | — | → users.id |
| content | VARCHAR(100) | — | 经验内容 |
| interpretation | TEXT | NULL | AI 解读，≤500 字 |
| domain | ENUM domain_type | — | 一级领域 |
| sub_domain | VARCHAR(50) | NULL | 二级领域 |
| is_official | BOOLEAN | FALSE | 官方种子 |
| is_private | BOOLEAN | FALSE | 私密经验 |
| source_type | VARCHAR(20) | 'user' | platform（平台生产） / user（用户原创） |
| creator_name | VARCHAR(100) | NULL | 原内容创作者（书籍作者、名人等） |
| source_label | VARCHAR(100) | NULL | 来源标注 |
| review_status | ENUM review_status | 'pending' | pending/approved/rejected/private |
| review_reason | TEXT | NULL | 审核理由 |
| quality_score | NUMERIC(3,1) | NULL | 0-10 综合分（前端 ÷2 展示为 0-5 星价值度） |
| score_reason | VARCHAR(100) | NULL | 打分简短理由，≤15 字 |
| score_details | JSONB | NULL | 6 维度明细 |
| like_count | INT | 0 | 触发器维护 |
| bookmark_count | INT | 0 | 触发器维护 |
| interpretation_generated | BOOLEAN | FALSE | AI 生成标记 |
| status | VARCHAR(20) | 'published' | published/hidden/flagged |
| deleted_at | TIMESTAMPTZ | NULL | 软删除时间（approved 经验用户删除时设置） |
| created_at | TIMESTAMPTZ | NOW() | — |
| updated_at | TIMESTAMPTZ | NOW() | — |

#### 12.2.1 平台生产经验

平台通过 AI 批量生产的经验，与用户经验区别：

| 属性 | 平台生产 (`source_type='platform'`) | 用户原创 (`source_type='user'`) |
|------|-----------------------------------|-------------------------------|
| is_official | TRUE | FALSE |
| 创作者 | `creator_name`（原内容作者） | `author_name`（发布用户昵称） |
| 标签 | 绿色圆形「官」+ hover「这条经验是年糕app根据公开内容生产的」 | 无 |
| 解读 | 必填（AI 生成，≤500 字） | 可选 |
| 价值度 | 必填（0-10 综合分 + 简短理由） | AI 审核后自动生成 |
| 删除 | 用户删除后从个人列表消失，平台池继续可见 | 已入池→软删除；其他→硬删除 |

批量评分脚本：`backend/scripts/batch_score.py`，调用 AI Service 的 `/api/v1/review` + `/api/v1/chat/generate-interpretation`，解释文本在句子边界智能截断以符合 500 字约束。

#### 12.2.2 名人经验种子（迁移 009）

已收录 9 位名人的经验，含完整解读和价值度：

| # | 名人 | 经验核心 | 价值度 |
|---|------|---------|--------|
| 1 | Elon Musk | 第一性原理思考 | 9.5 |
| 2 | 李小龙 | Be water — 像水一样适应 | 9.3 |
| 3 | 张小龙 | 好的产品用完即走 | 9.0 |
| 4 | 雷军 | 站在风口，识别真正的势 | 8.8 |
| 5 | Steve Jobs | Stay hungry, stay foolish | 9.6 |
| 6 | 张一鸣 | 延迟满足感是底层能力 | 9.2 |
| 7 | 段永平 | 做对的事情，再把它做对 | 9.1 |
| 8 | Kevin Kelly | 1000个铁粉理论 | 8.7 |
| 9 | 毛泽东 | 没有调查就没有发言权 | 9.4 |

### 12.3 likes / bookmarks / conversations / messages

```
likes (user_id, experience_id) PK
  └─ 触发器: update_like_count()

bookmarks (user_id, experience_id) PK
  ├─ practiced BOOLEAN, practiced_at TIMESTAMPTZ
  ├─ 触发器: update_bookmark_count()
  ├─ 触发器: update_user_bookmark_count()
  └─ 触发器: update_practiced_count()

conversations (id, user_id, title, created_at, updated_at)
  └─ messages (id, conversation_id, role, content, referenced_experience_ids, created_at)
```

---

## 十三、错误处理规范

### 13.1 HTTP 状态码

| 状态码 | 场景 | 返回体 |
|--------|------|--------|
| 200 | 成功 | 正常数据 |
| 201 | 创建成功 | 经验对象 + review 结果 |
| 400 | 参数校验失败 | `{error: "具体原因"}` |
| 401 | 未登录 | `{error: "请先登录"}` |
| 404 | 资源不存在 | `{error: "experience not found"}` |
| 500 | 服务器错误 | `{error: "failed to ..."}` |

### 13.2 移动端错误处理

```typescript
// config.ts — 统一错误类
class ApiError extends Error {
  status: number;
}

// 使用
if (e instanceof ApiError && e.status === 401) {
  Alert.alert('未登录', '请先登录后再发布经验');
} else {
  Alert.alert('发布失败', e?.message || String(e));
}
```

---

## 十四、测试与质量

### 14.1 测试覆盖

| 层 | 包/模块 | 测试文件 | 测试数 | 行覆盖 |
|----|---------|---------|--------|--------|
| Go backend | 5 | 6 | 50+ | ~49% |
| Python AI | 1 | 1 | 14 | ~32% |
| React Native | 0 | 0 | 0 | 0% |

### 14.2 已知限制

| 限制 | 影响 | 计划 |
|------|------|------|
| AI 服务直连 :8000 | 绕过 Nginx 统一入口 | 待改路由为 /ai/* 前缀 |
| scanExperiences 未含新字段 | 列表不显示 sub_domain/review_status | P1 |
| 移动端零测试 | 无回归保障 | P2 |
| 无 CI/CD | 手动部署 | P3 |
| repository 包无测试 | 需 DB mock | P2 |

---

## 十五、种子数据

- **105 条** approved 官方种子经验
- 21 二级领域 × 5 条/领域
- 每条 60-99 字，高质量中文经验
- author_id: `00000000-0000-0000-0000-000000000000`（官方账号）

---

## 十六、变更记录

| 版本 | 日期 | 变更 |
|------|------|------|
| v2.0 | 2026-05-05 | **重大更新**：新增完整流程图（发布/首页/推荐/登录/点赞）、屏幕详细描述、完整 API 参考、数据模型、错误处理规范、触发器文档 |
| v1.2 | 2026-05-05 | 首页接入推荐流 |
| v1.1 | 2026-05-05 | 审计修复 |
| v1.0 | 2026-05-05 | 初始 PRD |
