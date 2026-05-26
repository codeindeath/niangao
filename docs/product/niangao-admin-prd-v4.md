# 年糕管理后台 PRD v4.5

日期：2026-05-24
状态：一致性修订后的功能实现方案基准
范围：年糕管理后台，不包含 iOS App 用户侧页面
依赖：`niangao-user-prd-v4.md`、`niangao-technical-architecture-v4.md`

## 1. 后台定位

管理后台的核心定位是支持年糕 App 的运营、内容供给、质量治理、用户反馈、AI 成本和系统稳定性管理。

它不是单纯的经验生产后台。平台经验生产只是后台能力之一，不能压过用户产品运营和数据治理。

一级导航固定为：

1. 运营总览
2. 数据看板
3. 经验管理
4. 内容生产
5. 用户与反馈
6. AI 与系统

后台必须能管理：

- 平台精选经验。
- 用户公开原创经验。
- 用户私密原创经验。
- 用户私密议题和聊天。
- 用户反馈。
- 内容生产素材、批次、候选和任务。
- DeepSeek 各功能调用、key alias、限流、预算和失败。

## 2. 角色与权限

角色：

- Owner：全部权限，包括管理员管理、AI 配置、敏感数据查看、删除。
- Admin：日常运营权限，包括内容、用户、反馈、批次和任务管理。
- Viewer：只读权限，不可查看私密内容明文，不可执行写操作。

权限规则：

- 所有后台接口统一走 `/api/v1/admin/*`。
- 所有写操作必须校验角色。
- 私密经验、私密议题、聊天明文、用户轻画像默认折叠。
- 展开私密内容必须写入 `admin_private_access_logs`。
- Viewer 只能看聚合数据和脱敏信息。
- 后台不展示任何真实 DeepSeek key，只展示 key alias 和启用状态。

审计字段：

- admin_user_id。
- action。
- target_type。
- target_id。
- before_snapshot。
- after_snapshot。
- reason。
- ip。
- user_agent。
- created_at。

## 3. 通用页面规则

### 3.1 列表

所有列表页必须支持：

- 搜索。
- 筛选。
- 排序。
- 分页。
- 空状态。
- 加载状态。
- 错误重试。
- 权限不足状态。

分页规则：

- 后端 cursor 或 page 都可，但同一模块内部必须统一。
- 大数据列表优先 cursor。
- 管理操作列表可用 page + page_size。

### 3.2 详情

详情页固定结构：

1. 基础信息。
2. 当前状态。
3. 关键数据。
4. 关联对象。
5. 操作区。
6. 操作日志。

敏感区域：

- 默认折叠。
- 按钮文案为「查看私密内容」。
- 展开前写审计。
- 展开后展示本次查看人和查看时间。

### 3.3 批量操作

批量操作必须：

- 显示选中数量。
- 明确影响范围。
- 二次确认。
- 后端生成 operation_batch_id。
- 写入操作日志。
- 返回成功、失败、跳过数量。

失败项：

- 可下载失败明细。
- 不因部分失败回滚全部，除非该操作明确要求原子性。

## 4. 运营总览

### 4.1 页面目标

让运营者打开后台后快速判断：

- App 是否正常。
- 今日用户和内容是否异常。
- 公开经验供给是否足够。
- AI 和队列是否有风险。
- 是否存在需要立即处理的运营动作。

接口：

- `GET /api/v1/admin/overview`

### 4.2 指标卡

展示：

- 今日新增用户。
- DAU。
- 今日记下数。
- 今日聊聊议题数。
- 今日收藏数。
- 今日有启发数。
- 公开经验总数。
- 推荐可用经验数。
- 平台精选总数。
- 原创公开总数。
- 原创私密总数。
- AI 调用成功率。
- 今日 AI 成本估算。
- 队列积压数。

实现逻辑：

- 实时性要求低的指标从 daily stats 表读取。
- AI 成本和队列积压可实时读取。
- 如果聚合表缺失当天数据，允许后台同步补算最近 24 小时。
- 用户活跃来自用户日聚合。
- 内容供给来自 experiences、推荐可用池和 AI 可引用池。
- 聊聊质量来自议题数、消息数、清楚一点反馈。
- AI 状态来自 ai_call_logs 和 ai_jobs。
- 队列状态来自 job queue。

趋势展示：

- 核心指标卡同时展示今日值和近 7 天趋势。
- 7 天趋势不做复杂图表，使用小折线或涨跌标识。
- 推荐可用经验、AI 成功率、content_low 队列积压和公开原创转私密比例必须展示 7 天趋势，便于判断异常是否持续。

异常标识：

- AI 成功率低于阈值标红。
- content_low 队列积压超过阈值标黄。
- 推荐可用经验低于安全线标黄。
- API 错误率超过阈值标红。
- 推荐可用经验低于 1000 条标黄，低于 500 条标红。
- content_low 队列积压超过 500 个标黄，超过 2000 个标红。
- AI 用户实时功能成功率低于 97% 标黄，低于 92% 标红。
- 今日公开原创转私密比例超过 60% 标黄，提示可能存在发布质量或审核策略问题。

### 4.3 今日待处理

列表项：

- 待复查公开原创。
- 生产批次失败。
- AI 调用失败激增。
- 用户反馈未处理。
- 推荐池低于目标。
- 私密内容查看异常。

点击逻辑：

- 跳转到对应模块并自动带上筛选条件。
- 只展示需要运营或管理员采取行动的事项，不展示纯观察指标。

### 4.4 快捷操作

入口：

- 新建平台经验。
- 批量导入素材。
- 创建生产批次。
- 查看未处理反馈。
- 查看 AI 队列。

权限：

- Viewer 不显示写操作入口。

## 5. 数据看板

### 5.1 页面目标

数据看板用于判断产品增长、内容质量、用户行为和 AI 成本结构，不承担复杂 BI。

一级分页：

- 用户。
- 看看。
- 聊聊。
- 记下。
- 内容供给。
- AI 成本。

接口：

- `GET /api/v1/admin/analytics/users`
- `GET /api/v1/admin/analytics/feed`
- `GET /api/v1/admin/analytics/chat`
- `GET /api/v1/admin/analytics/notes`
- `GET /api/v1/admin/analytics/content-supply`
- `GET /api/v1/admin/analytics/ai-cost`

时间范围：

- 7 天。
- 30 天。
- 90 天。
- 自定义。

指标实现口径：

- 数据看板采用“事实事件 + 日聚合”的方案。
- 用户行为先记录事件，日常看板读取日聚合。
- 当天数据允许读取实时增量并与日聚合合并。
- 不在看板接口临时扫全量事件表。
- 看板不直接影响用户分发，但为推荐池、内容生产、AI 成本提供判断依据。

### 5.2 用户看板

指标：

- 新增用户。
- 活跃用户。
- 留存。
- Apple 登录成功率。
- 完善个人信息人数。
- 用户使用路径分布：先看、先聊、先记。
- 新用户第一天行为：看了几条、是否收藏、是否聊聊、是否记下。
- 有过记下的用户数。
- 有过聊聊的用户数。
- 有过收藏的用户数。

实现逻辑：

- 日维度写入 `user_daily_stats` 或统一 stats 表。
- 留存按首次登录日 cohort 计算。
- 用户行为用事件表和核心业务表交叉校验。
- DAU：当天有登录、看看曝光、收藏、有启发、记下或聊聊任一行为的去重用户数。
- 新增用户：当天首次完成 Apple 登录的用户数。
- 用户使用路径分布按注册后首个核心事件归类，核心事件包括 feed_expose、chat_message_send、note_save。
- 新用户第一天行为按注册后 24 小时窗口统计，不按自然日切分。

### 5.3 看看看板

指标：

- 推荐曝光。
- 上下滑切卡次数。
- 推荐翻面率。
- 推荐收藏率。
- 推荐有启发率。
- 推荐到收藏转化。
- 推荐到有启发转化。
- 收藏分页访问。
- 我的分页访问。
- 搜索次数。
- 搜索点击率。
- 搜索无结果率。
- 精选 / 原创在推荐中的曝光和反馈对比。
- 不可见占位展示次数。

实现逻辑：

- 原始行为写入 `experience_events`。
- 展示用聚合表，避免后台大范围扫事件表。
- 指标支持按经验类型、领域、创作者过滤。
- 看看曝光：`experience_events.expose` 去重后数量。
- 上下滑切卡次数：同一 recommendation_session 内从一个 experience_id 切到另一个 experience_id 的次数。
- 翻面率：qualified flip 数 / expose 数。
- 收藏率：collect 数 / expose 数。
- 有启发率：inspire 数 / expose 数。
- 推荐到收藏转化：推荐曝光后 24 小时内收藏的去重经验数 / 推荐曝光去重经验数。
- 推荐到有启发转化：推荐曝光后 24 小时内点有启发的去重经验数 / 推荐曝光去重经验数。
- 搜索点击率：search_click 数 / search_submit 数。
- 精选 / 原创对比按 recommendation_session 返回结果和后续行为关联计算。

### 5.4 聊聊看板

指标：

- 新增议题。
- 有效议题。
- 消息轮次。
- 继续历史议题次数。
- 清楚一点反馈数。
- 聊聊沉淀经验数。
- 引用经验展示数。
- 引用经验点击数。
- 引用后收藏数。
- 引用后有启发数。

实现逻辑：

- 议题来自 `chat_topics`。
- 消息来自 `chat_messages`。
- 引用来自 `chat_citations`。
- 清楚一点来自 `chat_session_feedback`。
- 聊聊会话数：有至少 1 条用户消息的会话数。
- 稳定议题数：进入 chat_topics 且未删除的议题数。
- 有效聊聊议题：至少 2 轮用户消息或产生清楚一点反馈的议题。
- 清楚一点率：“有一点”反馈数 / 有效反馈数，不含跳过。

### 5.5 记下看板

指标：

- 记下保存数。
- 公开保存数。
- 私密保存数。
- 默认公开保持率。
- 切换私密比例。
- 从聊聊沉淀数。
- 帮我整理点击数。
- 帮我整理采用率。
- 公开原创进入推荐候选数量。
- 公开原创获得收藏数。
- 公开原创获得有启发数。
- 公开转私密数。
- 记下后编辑率。
- 记下后删除率。
- 编辑后重新分层数量。

实现逻辑：

- 保存事实来自 `experiences.source_scene` 和 `created_at`。
- 默认公开保持率 = 记下保存时保持默认公开的数量 / 记下保存总数。
- 切换私密比例 = 用户在记下页主动切换为私密的保存数 / 记下保存总数。
- 帮我整理来自 `note_rewrite_click` 和对应 AI call。
- 公开原创进入推荐候选数量按 source_scene=note 且 quality_tier >= recommend_candidate 统计。
- 公开原创获得收藏 / 有启发按 note 经验与收藏、有启发事实表关联统计。
- 记下后编辑率 = 周期内编辑过的 note 经验数 / 周期内新增 note 经验数。
- 公开转私密统计需区分系统处理和用户主动转私密。
- 删除率 = 周期内删除的 note/chat 经验数 / 周期内新增 note/chat 经验数。

### 5.6 内容供给看板

指标：

- 经验总数。
- 公开经验数。
- 私密经验数。
- 平台精选数。
- 用户原创数。
- 推荐可用数。
- AI 可引用数。
- 各领域 / 子领域分布。
- 冷启动 3000 条进度。
- 来源类型分布。
- 创作者集中度。
- 低反馈经验数量。
- 下架数。

实现逻辑：

- 以 `experiences`、`source_materials`、`production_batches`、`candidate_experiences` 为事实源。
- 质量层级来自 `quality_tier`。
- 推荐可用由 `recommendation_status=eligible` 计算。
- 冷启动进度按 active 平台精选统计，不含 hidden、deleted、待确认候选。
- 可推荐数量和可 AI 引用数量单独统计，不能用总精选数替代。
- 内容供给只统计平台精选和公共原创供给，不混入用户私密经验表现。

### 5.7 AI 成本看板

指标：

- 按 function_type 的调用次数。
- 按 model 的调用次数。
- 按 call_source 的调用次数和成本。
- 成功率。
- 超时率。
- 平均延迟。
- token 用量。
- 成本估算。
- key alias 使用分布。
- 队列积压和失败。

实现逻辑：

- 事实数据来自 `ai_call_logs`。
- 聚合数据来自 `ai_usage_daily_stats`。
- 成本按配置单价估算，不作为财务凭证。
- call_source 至少区分 app_user、content_batch、admin_manual、scheduled_task。

## 6. 经验管理

### 6.1 页面目标

经验管理是统一内容资产管理中心，承载精选、公开原创、私密原创和不可见经验。

一级分页：

1. 全部经验
2. 精选
3. 原创公开
4. 原创私密
5. 待复查
6. 已下架

接口：

- `GET /api/v1/admin/experiences`
- `GET /api/v1/admin/experiences/:id`
- `PATCH /api/v1/admin/experiences/:id`
- `POST /api/v1/admin/experiences`
- `DELETE /api/v1/admin/experiences/:id`
- `POST /api/v1/admin/experiences/bulk-action`

### 6.2 列表字段

字段：

- id。
- 正文。
- 类型：精选 / 原创。
- 可见性：公开 / 私密。
- 生命周期：active / hidden / deleted / needs_review。
- unavailable 是前台占位展示态，不是 lifecycle_status。
- 领域。
- 子领域。
- 话题。
- 创建者名字。
- owner_user_id。
- source_reliability。
- quality_tier。
- recommendation_status。
- ai_citable。
- interpretation_status。
- 收藏数。
- 有启发数。
- 翻面率。
- 搜索点击数。
- AI 引用次数。
- 推荐分。
- 来源场景。
- 生产批次。
- 创建时间。
- 更新时间。

私密经验列表：

- 默认不显示正文，只显示「私密经验」占位。
- Owner/Admin 点击展开后写审计。
- Viewer 永远不显示明文。

### 6.3 筛选与排序

筛选：

- 经验类型。
- 可见性。
- 生命周期。
- 领域。
- 子领域。
- 话题。
- 创建者。
- owner_user_id。
- quality_tier。
- source_reliability。
- recommendation_status。
- ai_citable。
- interpretation_status。
- 来源场景。
- 生产批次。
- 创建时间。
- 收藏数区间。
- 有启发数区间。
- AI 引用次数区间。

搜索：

- 正文关键词。
- 创建者名字。
- 话题。
- id。

排序：

- 创建时间。
- 更新时间。
- 收藏数。
- 有启发数。
- 翻面率。
- 搜索点击数。
- AI 引用次数。
- 推荐分。
- quality_score。
- 低反馈优先。

### 6.4 经验详情

详情区块：

- 基础信息。
- 前台卡片预览。
- 经验解读。
- 质量与推荐。
- 数据表现。
- 来源追溯。
- 关联聊天或素材。
- 操作日志。

前台预览：

- 按用户 PRD 的卡片正反面规则展示。
- 不展示后台字段。
- 私密经验预览需要先展开私密明文。

来源追溯：

- 平台精选显示素材、原始链接、创作者、批次、候选记录。
- 用户原创显示来源场景：note / chat。
- 聊聊沉淀显示 source_chat_topic_id 和 source_chat_message_id。

分层与分发资格：

- 经验分层采用“公开资格、推荐资格、AI 引用资格、解读资格分离”的方案。
- quality_tier 是内容质量层级事实源。
- recommendation_status 是推荐分发开关，默认由 quality_tier 生成，但允许人工覆盖。
- recommendation_status 取值为 eligible / ineligible / suppressed。
- ai_citable 是公共 AI 引用开关，默认由 quality_tier 生成，但允许人工覆盖。
- visibility != public 或 lifecycle_status != active 时，recommendation_status 必须视为 ineligible，ai_citable 必须视为 false。
- quality_tier=public_visible 默认 recommendation_status=ineligible、ai_citable=false。
- quality_tier=recommend_candidate 默认 recommendation_status=eligible、ai_citable=false。
- quality_tier=ai_citable 或 high_trust 默认 recommendation_status=eligible、ai_citable=true。
- 人工覆盖 recommendation_status 或 ai_citable 必须记录 override_reason 和操作人。

评分口径：

- DeepSeek 输出 ai_quality_score 使用 1-10。
- 系统写入 quality_score 时统一归一为 0-100，默认换算为 ai_quality_score * 10。
- 人工复核可以直接调整 quality_score，但必须记录原因。
- quality_tier 由归一后的 quality_score、公开适配判断、来源可靠度和人工覆盖共同决定。
- 前台不展示 ai_quality_score、quality_score 或 quality_tier。

建议阈值：

- quality_score < 45：不适合公开。用户原创转私密；平台候选丢弃。
- 45-59：public_visible，只能公开展示和搜索，不进推荐。
- 60-74：recommend_candidate，可进入推荐候选。
- 75-84：ai_citable，可被公共 AI 引用。
- 85-100：high_trust，优先推荐和优先引用。

### 6.5 单条操作

通用操作：

- 编辑正文。
- 编辑领域、子领域、话题。
- 编辑创建者显示名。
- 修改 visibility。
- 修改 lifecycle_status。
- 修改 quality_tier。
- 修改 recommendation_status。
- 修改 ai_citable。
- 重新生成解读。
- 重新分层。
- 下架。
- 恢复。
- 删除。

操作逻辑：

- 编辑公开经验后设为 `needs_review`，并进入重新分层。
- 编辑私密经验不进入公共分层，除非改为公开。
- 下架后不再进入推荐、搜索、AI 公共引用。
- 恢复后按当前质量层级重新判断推荐资格。
- 删除为软删除，保留审计和必要聚合。
- 编辑平台精选正文后，interpretation_status 变为 stale，重新生成解读后才展示新反面。
- 编辑领域 / 子领域 / 话题立即影响搜索、推荐召回和后台筛选；如果领域变化较大，建议重新生成解读。
- 下架后收藏和聊天历史显示不可见。
- 恢复后重新计算 recommendation_status 和 ai_citable，不直接恢复原推荐排序。
- 降权只影响推荐，不改变用户收藏和历史聊天。
- 取消 AI 引用资格不影响推荐和搜索，只阻止后续公共 AI 引用。

用户原创限制：

- 只能后台下架、转私密、调整推荐资格。
- 后台不得编辑用户原创正文。
- 用户投诉处理或明显违规时，后台只能下架、转私密、删除异常内容、备注处理结果或触发复查，不能替用户改写正文。
- 用户仍有权在 App 编辑或删除自己创建的经验。
- 用户编辑公开原创后，lifecycle_status=needs_review，退出公共推荐、公共搜索和公共 AI 引用，重新处理；作者本人看到最新正文，其他用户收藏和历史引用显示 unavailable 占位。

平台精选：

- 后台可完整编辑。
- 编辑后必须记录操作人和原因。
- 来源创作者不能随意改为转述者。

### 6.6 批量操作

支持：

- 批量下架。
- 批量恢复。
- 批量调整推荐资格。
- 批量调整 AI 引用资格。
- 批量调整质量层级。
- 批量重新生成解读。
- 批量导出。

不支持：

- 批量硬删除。
- 批量编辑正文。
- 批量查看私密明文。

执行规则：

- 批量任务必须生成 operation_batch_id。
- 返回成功、失败、跳过数量和明细。
- 批量操作每次默认最多 200 条；超过需分批。
- 批量操作不可跳过权限和私密内容审计。

### 6.7 验收

- 精选、原创公开、原创私密在同一模型下可查询。
- 私密正文默认折叠，展开有审计。
- 下架经验不会在推荐、搜索和公共 AI 引用出现。
- 编辑公开经验会进入重新分层。
- 批量操作有成功、失败、跳过明细。

## 7. 内容生产

### 7.1 页面目标

内容生产模块管理平台精选经验的供给链路。

一级分页：

1. 素材库
2. 生产批次
3. 候选经验
4. 创作者池
5. 来源池
6. 采集名单
7. 排除名单
8. 内容周报
9. 任务日志

### 7.2 素材库

接口：

- `GET /api/v1/admin/source-materials`
- `POST /api/v1/admin/source-materials`
- `POST /api/v1/admin/source-materials/batch-import`
- `GET /api/v1/admin/source-materials/:id`
- `PATCH /api/v1/admin/source-materials/:id`
- `DELETE /api/v1/admin/source-materials/:id`

字段：

- title。
- source_url。
- source_platform。
- source_type：web / video / podcast / book / manual。
- access_type：public / login_required / paid / manual。
- source_derivation_type：direct_quote / expressed_principle / behavior_extraction。
- creator_name。
- transcript_or_excerpt。
- source_excerpt。
- source_location：页码、时间戳、章节或页面位置。
- language。
- transcription_method。
- confidence_score。
- display_name_policy。
- copyright_policy。
- collection_method。
- captured_by。
- raw_content_storage_policy。
- collected_by。
- collected_at。
- process_status。
- usable_for_extraction。

保存规则：

- 公开网页可保存全文或主要正文。
- 视频和播客可保存转写。
- 书籍只保存短片段、页码或位置，不保存整本内容。
- 登录态平台至少保存链接、创作者、短片段和抓取时间。
- 付费内容不自动抓取，不保存绕过付费墙获得的全文。
- 长转写和敏感素材全文默认折叠，后台查看需走私密 / 敏感内容审计。
- direct_quote 表示创作者原话。
- expressed_principle 表示创作者在原文中明确表达的原则。
- behavior_extraction 表示从传记或访谈叙事中的行为提炼出的原则，AI 高风险引用时需要降权。

导入逻辑：

- 单条导入立即保存素材。
- 批量导入生成 import job。
- 重复 URL 默认跳过，可手动强制更新。

### 7.3 生产批次

接口：

- `GET /api/v1/admin/production-batches`
- `POST /api/v1/admin/production-batches`
- `GET /api/v1/admin/production-batches/:id`
- `POST /api/v1/admin/production-batches/:id/start`
- `POST /api/v1/admin/production-batches/:id/pause`
- `POST /api/v1/admin/production-batches/:id/resume`
- `POST /api/v1/admin/production-batches/:id/cancel`
- `POST /api/v1/admin/production-batches/:id/retry-failed`
- `POST /api/v1/admin/production-batches/:id/export`

字段：

- name。
- source_scope。
- target_domains。
- target_sub_domains。
- target_count。
- target_phase：calibration / cold_start / daily / weekly_theme / hotspot。
- decision_status：pending / promoted / rejected。
- total_materials。
- extracted_count。
- approved_count。
- rejected_count。
- failed_count。
- ai_cost_estimate。
- created_by。
- started_at。
- completed_at。

状态流：

```text
draft -> queued -> running -> paused -> running -> completed
                       |          |
                       |          -> canceled
                       -> failed
```

执行流程：

1. 选择素材。
2. 创建 batch 和 processing_unit。
3. 生成 content_low 队列任务。
4. 执行 translation_normalization。
5. DeepSeek 执行 experience_extract。
6. 执行 experience_review 和质量审计标准。
7. 去重。
8. 分类领域、子领域、话题。
9. 执行安全、隐私和公开适配审核。
10. 高置信直接生成平台精选。
11. 中置信或边界低置信但仍可能有价值的内容进入待确认候选经验。
12. 高质量经验生成解读。
13. 写入批次报告。

AI 处理链固定为：

1. translation_normalization：必要时做翻译、繁简转换、格式清理和文本归一。
2. experience_extract：从归一后的素材中提取候选经验。
3. experience_review：按质量审计标准判断是否保留。
4. dedupe：去重和相似合并。
5. experience_classify：分类领域、子领域、话题。
6. moderation：判断安全、隐私和明显风险。
7. experience_interpretation：为高质量经验生成解读。

质量审计标准：

- 经验必须是普适、指导行为或认知变化的原则。
- 删除后用户没有损失的内容应删除。
- 删除个人经历、客观描述、纯鸡汤、段子、领域过窄、空泛内容。
- 去掉非必要外壳和格式杂质。
- 平台精选尽量摘取原文核心内容，不做无必要改写。
- 原文核心超过 100 字才适当压缩。
- 行为提炼必须能在原素材中找到明确行为依据；如果只是从故事中脑补道理，直接丢弃。
- 高风险领域相关的行为提炼不进入强引用池，默认只可作为背景视角。

冷启动供给目标：

- 上线前平台精选最低储备 3000 条。
- 3000 条不是数量填充，必须覆盖当前项目定义的 6 个一级领域和 35 个子领域。
- 不要求 3000 条全部生成解读；只有达到 recommend_candidate、ai_citable 或 high_trust 的经验才生成解读。
- 生产批次必须展示当前冷启动进度、领域缺口、子领域缺口、创作者集中度、来源类型分布和可 AI 引用数量。
- 冷启动内容气质以“有态度的活法”为主，粗比例目标为有态度约 60%、稳妥实用约 40%。
- 该比例只用于批次校准和周报复盘，不作为前台标签或数据库强字段。

冷启动一级领域比例：

- 意义：25%-30%。
- 认知：18%-22%。
- 工作：18%-22%。
- 关系：12%-16%。
- 生活：8%-12%。
- 生命：5%-8%。

冷启动子领域优先级：

- 意义：幸福 = 自我 > 情绪 > 使命 > 归属 > 信仰。
- 认知：思维 > 创造 = 表达 > 学习 > 信息 > 工具。
- 工作：创业 > 沟通 > 管理 > 效率 > 升职 > 求职。
- 关系：恋人 > 父母 > 朋友 > 夫妻 > 亲子 > 兄妹。
- 生活：娱乐 > 旅行 > 养护 > 购物 > 衣着 > 宠物。
- 生命：饮食 > 运动 > 健康 > 居住 > 出行。

生产节奏：

- 前 3 个校准批次每批 100 条素材 / 目标候选，用于校准审美、领域覆盖、创作者选择和经验提取尺度。
- 前 3 批必须形成校准报告，包含统计摘要、好样例、坏样例、争议样例和需要调整的提取标准。
- 校准稳定后，高置信内容可以自动入库。
- 后续主题批可以扩大到 300 或 500 条素材，但后台必须按 50-100 条的处理单元拆分任务、审核和回滚。
- 日常小批量保持 20-50 条，用于热点内容、单个新来源或小范围补缺。
- 校准稳定后的冷启动主题批可以扩大到 300 或 500 条素材，但必须拆成 50-100 条 processing_unit，单元级可暂停、重试、回滚和复查。

创作者集中度：

- 单个创作者在冷启动 3000 条中的入库软上限为 50 条。
- 特别高价值创作者可到 100 条，但超过 50 条必须在批次报告和周报中标记。
- 超过 100 条原则上不进入冷启动池，除非人工明确批准。
- 推荐曝光仍需另做打散，入库数量不等于推荐曝光。

来源可靠度：

- source_reliability 使用 high / medium / low 三档，只后台展示，不在前台展示。
- high：本人原始表达、本人博客 / 书籍、公开视频 / 公开访谈原话、可追溯时间戳。
- medium：可信二手记录、正式出版传记、权威媒体整理、可靠采访稿。
- low：平台转载、剪辑号、普通用户转述、出处不完整但内容可能有价值。
- source_reliability 影响 AI 引用资格、推荐权重、质量分层和复查优先级，但不直接决定经验质量。

周报必须包含：

- 本周新增精选数量。
- 高置信自动入库数量。
- 中置信候选数量。
- 丢弃数量和主要原因。
- 领域 / 子领域覆盖。
- 创作者集中度。
- 来源可靠度分布。
- 有态度经验 vs 稳妥实用经验的大致比例。
- 行为提炼经验数量和高风险引用限制数量。
- 达到推荐候选和 AI 可引用的数量。
- 反面解读覆盖数量。
- 冷启动 3000 条进度。
- 下周采集缺口。
- 代表性好样例、坏样例、争议样例。
- 有态度但可能有误用风险的样例。

### 7.4 候选经验

接口：

- `GET /api/v1/admin/candidate-experiences`
- `GET /api/v1/admin/candidate-experiences/:id`
- `POST /api/v1/admin/candidate-experiences/:id/approve`
- `POST /api/v1/admin/candidate-experiences/:id/reject`
- `POST /api/v1/admin/candidate-experiences/:id/promote`

字段：

- candidate_content。
- source_material_id。
- production_batch_id。
- creator_name。
- source_derivation_type。
- source_excerpt。
- domain。
- sub_domain。
- topic。
- ai_score。
- ai_reason。
- duplicate_group_id。
- status。

操作：

- 通过：生成平台精选经验。
- 拒绝：不入库。
- 不提供编辑正文后通过。
- 不提供人工合并重复。
- 如果缺领域、子领域、话题、质量分、解读资格等字段，通过后由对应 AI 功能或后台补齐任务处理。

候选去向：

- 高置信：ai_quality_score >= 8 或 quality_score >= 80，且 creator 明确、duplicate_confidence < 0.85、source_reliability 不低，可自动入库。
- 中置信：ai_quality_score 6-7 或 quality_score 60-79，或创作者 / 重复 / 经验边界低置信，进入候选池。
- 低置信：ai_quality_score < 6 或 quality_score < 60，或命中删除 6 类，直接丢弃但保留统计。
- behavior_extraction 类型候选即使质量分高，也必须检查误用风险；离职、投资、亲密关系、健康、法律等高风险主题默认不进入强 AI 引用。

候选数量：

- 单条短素材最多提取 3 条候选。
- 长文、访谈、书籍片段最多提取 8 条候选。
- 如果 AI 输出超过上限，只保留评分最高的候选。

去重和相似合并：

- 文本完全一致判重复。
- 核心谓词和对象一致判高度相似。
- 同一创作者、同一素材、表达接近时优先判重复。
- 不同创作者表达相似但态度不同，进入人工候选，不自动合并。
- 自动重复丢弃新候选并保留来源记录。
- 高度相似进入 duplicate_group 供系统去重和追溯。
- 第一阶段不提供人工合并重复操作；不确定相似项保留为待确认候选或直接丢弃。

### 7.5 创作者池

接口：

- `GET /api/v1/admin/creators`
- `POST /api/v1/admin/creators`
- `PATCH /api/v1/admin/creators/:id`

字段：

- display_name。
- aliases。
- preferred_display_name_zh。
- preferred_display_name_en。
- creator_type：public_figure / author / user / organization。
- bio_short。
- suitable_domains。
- suitable_sub_domains。
- priority。
- source_count。
- experience_count。
- quality_feedback。
- soft_cap_status。
- status。

规则：

- 创作者是经验真正提出者，不是转述者。
- 同一人不同译名用 aliases 合并。
- 前台显示名优先使用中文用户熟悉的名字；常见人物用中文常用名，不常见但专业圈熟悉的人可保留英文名。
- 后台保留中英文别名，用于去重、搜索、统计和推荐打散。
- 用户原创的创建者来自用户展示名，不进入平台创作者池作为可编辑对象。

### 7.6 来源池和采集名单

来源池字段：

- name。
- url。
- platform。
- content_type。
- source_quality：high / medium / low。
- sustainability：one_time / periodic / long_term。
- access_difficulty：low / medium / high。
- priority。
- login_required。
- scheduled_crawl_suitable。
- text_quality。
- suitable_domains。
- suitable_sub_domains。
- collected_count。
- imported_experience_count。
- quality_feedback。
- exclusion_status。
- collection_strategy。
- status。

采集名单字段：

- target_name。
- target_type：person / site / book / topic。
- priority。
- reason。
- target_count。
- collected_count。
- next_action。
- owner。
- exclusion_status。

实现逻辑：

- Mac Agent 负责需要登录、探索性强的采集。
- 服务器只做公开稳定来源的低频抓取。
- 后台记录采集计划、状态和结果，不承担复杂浏览器自动化控制台。
- Mac Agent 采集不能自动支付、改账号资料、点赞、评论、关注、私信或执行其他会改变账号状态的动作。
- 登录态平台优先保存链接、创作者、短片段和时间；需要用户判断的账号登录或授权由用户完成。
- 服务器定时任务只允许访问公开、稳定、低风险来源；遇到登录、验证码、付费墙或平台限制时停止并记录失败原因。

### 7.7 排除名单

接口：

- `GET /api/v1/admin/exclusion-list`
- `POST /api/v1/admin/exclusion-list`
- `PATCH /api/v1/admin/exclusion-list/:id`

字段：

- target_type：creator / source / material / keyword。
- target_name。
- target_id。
- reason。
- severity。
- created_by。
- created_at。
- status。

规则：

- 排除名单用于避免低质量源、营销号、搬运号、严重风险人物或不符合年糕气质的内容。
- 命中排除名单的来源和创作者不能进入新批次，除非 Owner 手动解除。
- 排除名单只影响后续生产，不自动下架已经入库的精选经验；是否下架由经验管理单独处理。

### 7.8 内容周报

接口：

- `GET /api/v1/admin/content-weekly-reports`
- `GET /api/v1/admin/content-weekly-reports/:id`
- `POST /api/v1/admin/content-weekly-reports/:id/export`

字段：

- report_week。
- selected_count。
- auto_imported_count。
- candidate_count。
- discarded_count。
- discard_reasons。
- domain_coverage。
- sub_domain_gaps。
- creator_concentration。
- source_reliability_distribution。
- recommend_candidate_count。
- ai_citable_count。
- interpretation_count。
- cold_start_progress。
- next_collection_gaps。
- good_examples。
- bad_examples。
- disputed_examples。

规则：

- 周报由生产批次和素材处理结果聚合生成，可人工补充备注。
- 周报保留在后台，可导出 Markdown。
- 周报只提示领域 / 子领域缺口和内容质量问题，不自动生成下一批采集计划。

### 7.9 任务日志

任务类型：

- source_import。
- translation_normalization。
- experience_extract。
- experience_review。
- experience_classify。
- moderation。
- experience_interpretation。
- batch_export。

字段：

- job_id。
- job_type。
- queue_name。
- status。
- input_summary。
- output_summary。
- error_message。
- retry_count。
- created_at。
- started_at。
- finished_at。

操作：

- 查看详情。
- 重试。
- 取消。
- 导出日志。

## 8. 用户与反馈

### 8.1 页面目标

用户与反馈模块用于观察真实用户使用情况、处理反馈、定位问题，不作为客服工单系统。

一级分页：

1. 用户列表
2. 用户详情
3. 反馈
4. 私密内容审计

### 8.2 用户列表

接口：

- `GET /api/v1/admin/users`

字段：

- user_id。
- display_name。
- Apple 登录标识脱敏。
- 注册时间。
- 最近活跃时间。
- 活跃天数。
- 看看浏览数。
- 有启发次数。
- 经验数。
- 原创公开经验数。
- 原创私密经验数。
- 收藏数。
- 聊聊议题数。
- 清楚一点次数。
- 反馈次数。
- 状态。

筛选：

- 注册时间。
- 最近活跃。
- 是否有原创经验。
- 是否有聊聊。
- 用户状态。

### 8.3 用户详情

接口：

- `GET /api/v1/admin/users/:id`
- `PATCH /api/v1/admin/users/:id/status`

区块：

- 基础信息。
- App 使用统计。
- 公开原创。
- 私密原创。
- 收藏。
- 聊聊议题。
- 聊天记录。
- 用户个人信息。
- 后台轻画像。
- 反馈记录。
- 操作日志。

私密内容：

- 私密原创、私密议题、聊天明文、后台轻画像默认折叠。
- 展开写审计。
- 不允许批量导出私密聊天明文。

用户状态：

- active。
- suspended。
- deleted。

操作：

- 暂停账号。
- 恢复账号。
- 查看审计。

注销逻辑：

- 用户从 App 发起注销。
- 后台只查看状态，不主动代用户注销，除非 Owner 处理合规请求。

### 8.4 反馈

接口：

- `GET /api/v1/admin/feedback`
- `GET /api/v1/admin/feedback/:id`
- `PATCH /api/v1/admin/feedback/:id`

字段：

- feedback_id。
- user_id。
- type。
- content。
- app_version。
- device。
- os_version。
- status。
- internal_note。
- created_at。

状态：

- pending：未处理。
- in_progress：处理中。
- resolved：已处理。
- no_action：无需处理。

操作：

- 标记处理中。
- 标记已处理。
- 标记无需处理。
- 添加内部备注。

不做：

- 回复用户。
- 分配客服。
- 工单 SLA。

### 8.5 私密内容审计

接口：

- `GET /api/v1/admin/privacy-audit-logs`

展示：

- 管理员。
- 查看对象。
- 对象类型。
- 用户 id。
- 查看时间。
- 操作原因。
- ip。

异常提示：

- 同一管理员短时间大量查看。
- Viewer 尝试访问私密内容。
- 非工作时间高频查看。

私密内容范围：

- 用户私密经验正文。
- 私密议题标题、摘要、消息。
- 聊天中引用的用户私密经验。
- 后台轻画像中可能暴露用户处境的字段。
- 登录态平台采集素材的长正文或长转写。
- 版权敏感素材的原文片段和处理备注。

查看流程：

1. 管理员点击展开。
2. 选择查看原因：用户反馈处理、安全复查、技术排查、用户请求、其他。
3. 写入 admin_private_access_logs。
4. 审计写入成功后才返回明文。
5. 本页面会话内可继续查看同一对象，刷新或离开后重新审计。

限制：

- Viewer 永远不能查看私密明文。
- 私密聊天不支持批量导出。
- 私密内容不进入公共内容运营列表。
- 私密内容不参与公共推荐、公共搜索、他人 AI 引用和公共质量表现统计。

异常阈值：

- 同一管理员 1 小时内查看私密对象超过 30 个，标记异常。
- 非工作时间连续查看超过 10 个私密对象，标记异常。
- Viewer 访问私密接口返回 403 并记录安全日志。

## 9. AI 与系统

### 9.1 页面目标

AI 与系统模块负责 DeepSeek 功能治理、成本治理、队列和后台系统配置。

AI 功能的调用时机、payload、输出 schema 和阶段变体，以 `docs/product/niangao-ai-functional-prompt-spec-v4.md` 为准；生产级 prompt 模板、评分锚点、正反例、golden cases 和上线门槛，以 `docs/product/niangao-ai-prompt-production-spec-v4.md` 为准；本章负责后台配置、日志、队列、预算和审计承接。

一级分页：

1. AI 功能配置
2. Key Alias
3. 调用日志
4. 任务队列
5. 限流与预算
6. 管理员与审计
7. 系统配置

### 9.2 AI 功能配置

接口：

- `GET /api/v1/admin/ai/function-configs`
- `PATCH /api/v1/admin/ai/function-configs/:id`

function_type：

- chat。
- chat_summary。
- chat_topic_classify。
- experience_rewrite。
- experience_extract。
- experience_review。
- experience_classify。
- experience_interpretation。
- recommendation_ai。
- moderation。
- translation_normalization。

字段：

- function_type。
- enabled。
- model。
- key_alias。
- prompt_version。
- schema_version。
- timeout_ms。
- max_tokens。
- temperature。
- queue_name。
- fallback_strategy。
- priority。
- max_retries。
- retry_backoff_seconds。
- per_minute_call_limit。
- per_minute_token_limit。
- daily_call_limit。
- daily_cost_limit。

规则：

- 用户实时链路优先使用高可用 key alias。
- 内容生产使用低优先级 alias。
- 修改配置写审计。
- 修改后新请求生效，不影响已入队任务。
- AI 与系统采用“功能级配置，而不是模型级配置”的方案。
- prompt_version 只选择已注册并通过对应 eval 的版本；eval 必须满足生产级 Prompt 规格中的 schema、product、quality 和 adversarial 四层评测；不得在后台直接编辑完整 prompt 文本。
- schema_version 用于校验 AI 输出解析规则，破坏性变更必须新建版本。
- 已入队任务使用创建时的配置快照，避免中途切换造成结果不可追溯。
- chat 最高优先级；experience_rewrite 次高；moderation 属于用户实时安全链路；chat_topic_classify 和 chat_summary 可后台延迟；内容生产最低。

第一阶段默认配置：

| function_type | key_alias | queue_name | 优先级 / 降级 |
| --- | --- | --- | --- |
| chat | deepseek_chat_primary | user_realtime | 最高优先级，失败保留用户消息并可重试 |
| moderation | deepseek_moderation_primary | user_realtime | 最高优先级，失败则延迟公开分发 |
| experience_rewrite | deepseek_user_primary | user_normal | 中高优先级，失败不影响保存 |
| chat_summary | deepseek_chat_primary | user_background | 可延迟补偿 |
| chat_topic_classify | deepseek_chat_primary 或 deepseek_user_primary | user_normal | 失败用临时标题和空分类 |
| recommendation_ai | deepseek_recommendation_primary | user_normal | 不在实时推荐主链路，失败退回规则 |
| experience_extract | deepseek_content_primary | content_low | 可暂停 |
| experience_review | deepseek_content_primary | content_low | 可暂停 |
| experience_classify | deepseek_content_primary | content_low | 可暂停 |
| experience_interpretation | deepseek_content_primary | content_low | 成本压力大时优先暂停 |
| translation_normalization | deepseek_content_primary | content_low | 成本压力大时优先暂停 |

### 9.3 Key Alias

接口：

- `GET /api/v1/admin/ai/key-aliases`
- `POST /api/v1/admin/ai/key-aliases`
- `PATCH /api/v1/admin/ai/key-aliases/:id`

字段：

- alias。
- env_var_name。
- provider。
- status。
- usage_scope。
- last_success_at。
- last_failure_at。

规则：

- 后台不展示真实 key。
- 创建 alias 时只填写 env var name。
- 服务端启动时从环境变量读取真实 key。
- key 不可从后台明文导出。

### 9.4 调用日志

接口：

- `GET /api/v1/admin/ai/call-logs`
- `GET /api/v1/admin/ai/call-logs/:id`

字段：

- request_id。
- function_type。
- key_alias。
- provider。
- model。
- prompt_version。
- schema_version。
- call_source。
- queue_name。
- priority。
- status。
- latency_ms。
- input_tokens。
- output_tokens。
- total_tokens。
- cost_estimate。
- error_code。
- attempt_no。
- retry_of_call_id。
- user_id。
- production_batch_id。
- source_material_id。
- candidate_experience_id。
- experience_id。
- chat_topic_id。
- chat_message_id。
- job_id。
- started_at。
- finished_at。
- created_at。

详情：

- 默认不展示完整 prompt 和完整输出。
- Owner 可查看脱敏后的输入输出摘要。
- 涉及用户私密内容时按私密内容审计处理。

### 9.5 任务队列

接口：

- `GET /api/v1/admin/ai/jobs`
- `POST /api/v1/admin/ai/jobs/:id/retry`
- `POST /api/v1/admin/ai/jobs/:id/cancel`
- `POST /api/v1/admin/ai/queues/:name/pause`
- `POST /api/v1/admin/ai/queues/:name/resume`

队列：

- user_realtime。
- user_normal。
- user_background。
- content_low。
- admin_manual。

规则：

- user_realtime 不允许后台暂停。
- content_low 可暂停，不影响 App 主流程。
- 失败任务可单条重试或批量重试。
- 同一任务最多自动重试 2 次，手动重试不超过 5 次。
- 输入格式错误、素材缺失不自动重试。
- 批次失败项超过 30% 时，批次标记 attention_required。

### 9.6 限流与预算

接口：

- `GET /api/v1/admin/ai/budget-policies`
- `PATCH /api/v1/admin/ai/budget-policies/:id`

维度：

- function_type。
- key_alias。
- 环境。
- 日预算。
- 月预算。
- 单用户频率。
- 全局并发。

降级规则：

- chat 失败返回温和失败文案。
- recommendation_ai 失败回退规则推荐。
- experience_interpretation 失败显示暂无解读。
- content_low 超预算自动暂停。
- experience_rewrite 失败不影响保存。
- chat_summary 和 chat_topic_classify 失败可延迟补偿。
- moderation 失败时延迟公开分发，不影响用户保存成功。

失败分类：

- provider_timeout：可自动重试。
- provider_rate_limited：按 key alias 限流退避。
- invalid_input：不自动重试，标记 skipped 或 failed。
- parse_error：可重试 1 次，仍失败进入人工或降级。
- budget_exceeded：按队列策略暂停或拒绝新任务。
- permission_denied：直接失败并记录安全日志。

### 9.7 管理员与审计

接口：

- `GET /api/v1/admin/admin-users`
- `POST /api/v1/admin/admin-users`
- `PATCH /api/v1/admin/admin-users/:id`
- `GET /api/v1/admin/audit-logs`

第一阶段：

- Owner 可创建 Admin 和 Viewer。
- 不做复杂组织和审批。
- 审计日志不可删除。

### 9.8 系统配置

配置项：

- 推荐池安全线。
- AI 告警阈值。
- 公开经验异步处理开关。
- 内容生产开关。
- App 功能开关。
- 领域和子领域只读展示。

规则：

- 领域和子领域使用当前项目定义，不在后台随意新增。
- 如需调整领域体系，必须走产品和数据库迁移流程。

## 10. 关键业务状态

### 10.1 经验生命周期

```text
active -> needs_review -> active
active -> hidden -> active
active -> deleted
hidden -> deleted
```

说明：

- `active` 表示经验有效。
- `needs_review` 表示公开内容被编辑后需要重新分层。
- `hidden` 表示前台不可见。
- `deleted` 表示软删除。

### 10.2 公开原创处理

用户从「记下」创建公开原创：

1. App 同步保存。
2. 用户看到「已记下」。
3. 后台进入 moderation。
4. 不适合公开则自动转 private，用户无感。
5. 可公开则进入分层。
6. 达到门槛后可推荐、可引用、可生成解读。

后台必须显示真实处理状态，但不得通过 App 通知用户审核失败。

### 10.3 平台精选入库

平台精选来源：

- 自动采集。
- 批量导入。
- 单条手动新增。

入库要求：

- creator_display_name 必填。
- content 必填且不超过 100 字。
- domain/sub_domain/topic 尽量填写。
- 通过质量审计。
- 高质量才生成解读。

## 11. API 汇总

运营总览：

- `GET /api/v1/admin/overview`

数据看板：

- `GET /api/v1/admin/analytics/users`
- `GET /api/v1/admin/analytics/feed`
- `GET /api/v1/admin/analytics/chat`
- `GET /api/v1/admin/analytics/notes`
- `GET /api/v1/admin/analytics/content-supply`
- `GET /api/v1/admin/analytics/ai-cost`

经验管理：

- `GET /api/v1/admin/experiences`
- `POST /api/v1/admin/experiences`
- `GET /api/v1/admin/experiences/:id`
- `PATCH /api/v1/admin/experiences/:id`
- `DELETE /api/v1/admin/experiences/:id`
- `POST /api/v1/admin/experiences/bulk-action`

内容生产：

- `GET /api/v1/admin/source-materials`
- `POST /api/v1/admin/source-materials`
- `POST /api/v1/admin/source-materials/batch-import`
- `GET /api/v1/admin/production-batches`
- `POST /api/v1/admin/production-batches`
- `POST /api/v1/admin/production-batches/:id/start`
- `POST /api/v1/admin/production-batches/:id/pause`
- `POST /api/v1/admin/production-batches/:id/resume`
- `POST /api/v1/admin/production-batches/:id/cancel`
- `GET /api/v1/admin/candidate-experiences`
- `POST /api/v1/admin/candidate-experiences/:id/approve`
- `POST /api/v1/admin/candidate-experiences/:id/reject`
- `GET /api/v1/admin/creators`
- `POST /api/v1/admin/creators`
- `GET /api/v1/admin/source-pools`
- `GET /api/v1/admin/collection-targets`
- `GET /api/v1/admin/exclusion-list`
- `GET /api/v1/admin/content-weekly-reports`
- `GET /api/v1/admin/jobs`

用户与反馈：

- `GET /api/v1/admin/users`
- `GET /api/v1/admin/users/:id`
- `PATCH /api/v1/admin/users/:id/status`
- `GET /api/v1/admin/feedback`
- `PATCH /api/v1/admin/feedback/:id`
- `GET /api/v1/admin/privacy-audit-logs`

AI 与系统：

- `GET /api/v1/admin/ai/function-configs`
- `PATCH /api/v1/admin/ai/function-configs/:id`
- `GET /api/v1/admin/ai/key-aliases`
- `POST /api/v1/admin/ai/key-aliases`
- `GET /api/v1/admin/ai/call-logs`
- `GET /api/v1/admin/ai/jobs`
- `POST /api/v1/admin/ai/jobs/:id/retry`
- `POST /api/v1/admin/ai/jobs/:id/cancel`
- `GET /api/v1/admin/ai/budget-policies`
- `GET /api/v1/admin/audit-logs`

## 12. 非功能需求

性能：

- 后台首屏接口 P95 不超过 1.5 秒。
- 普通列表查询 P95 不超过 1 秒。
- 大批量导出走异步任务。
- 统计看板使用聚合表，不直接扫大事件表。

安全：

- 后台必须 HTTPS。
- 管理员 JWT 独立于用户 App token。
- 私密内容访问必须审计。
- 真实 API key 不落库明文。
- 删除、下架、批量操作必须二次确认。

可靠性：

- 内容生产失败不影响用户 App。
- AI 内容生产队列可暂停和恢复。
- 批量任务可重试。
- 后台操作失败必须返回明确原因。

可维护性：

- 后台枚举与后端统一来源。
- 列表筛选参数进入 URL，方便复制和复现。
- 所有写操作通过统一 admin API client。

## 13. 第一阶段后台不做事项

第一阶段后台明确不做：

- 复杂 BI 自定义报表。
- 多级组织、审批流和细粒度岗位权限。
- 真实 AI key 明文查看、编辑或导出。
- 环境变量在线编辑。
- 客服工单系统、用户私信、客服聊天。
- 自动回复用户反馈。
- 私密内容批量导出。
- 自动内容采集规划器。
- 复杂审核工作流和多级复核。
- 用户标签运营。
- 社交关系管理。
- 创作者主页运营。
- 向 App 展示公开审核失败原因。
- 普通管理员绕过私密审计查看明文。

不做事项不应在后台放置不可用入口或“即将上线”占位。

## 14. 验收清单

- 六个一级导航完整可访问。
- 运营总览能显示用户、内容、AI、队列核心状态。
- 数据看板能按时间范围查看用户、看看、聊聊、记下、内容供给、AI 成本指标。
- 经验管理能统一管理精选、公开原创、私密原创。
- 私密内容默认折叠，展开写审计。
- 内容生产能完成素材导入、批次处理、候选入库、排除名单、内容周报、任务查看。
- 用户详情能查看统计、反馈、公开内容和受控私密内容。
- AI 与系统能管理 function config、prompt_version、schema_version、key alias、调用日志、队列和预算。
- Viewer 无法查看私密明文和执行写操作。
- 所有下架、删除、批量操作都有日志。

实现验收矩阵：

| 模块 | 必须验证 |
| --- | --- |
| 运营总览 | 指标来源正确，异常阈值能触发标黄 / 标红，待处理能跳到对应列表 |
| 数据看板 | 指标口径一致，7 天 / 30 天 / 90 天切换正确 |
| 经验管理 | 三类经验权限不同，状态变更能同步影响推荐、搜索、AI 引用 |
| 批量操作 | 成功 / 失败 / 跳过明细正确，批量操作有 operation_batch_id |
| 内容生产 | 批次、processing_unit、候选、高置信入库、周报完整闭环 |
| 冷启动 | 3000 条进度、领域缺口、创作者集中度、AI 可引用数量可见 |
| 私密审计 | 展开前必须写审计，Viewer 无法查看，异常查看能被标记 |
| AI 配置 | function_type、prompt_version、schema_version、key_alias、队列、预算、失败日志都能追溯，prompt 版本只允许选择已注册且通过 eval 的版本 |
