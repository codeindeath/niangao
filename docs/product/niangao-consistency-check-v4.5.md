# 年糕 PRD v4.5 跨功能一致性检测报告

日期：2026-05-24

检查范围：

- `docs/product/niangao-user-prd-v4.md`
- `docs/product/niangao-admin-prd-v4.md`
- `docs/product/niangao-technical-architecture-v4.md`

## 1. 检查结论

v4.5 已完成跨功能一致性修订。

本轮重点修复的是修订前版本中“单个功能成立，但跨功能连接不够严密”的问题。当前推荐、搜索、记下、聊聊、后台分层、内容生产、私密审计、AI 治理和技术架构之间没有发现剩余结构性冲突。

## 2. 已修复的一致性问题

### 2.1 推荐门槛与质量分层

问题：

- 用户 PRD 曾写推荐数据源最低 `public_visible`。
- 后台 PRD 写 `public_visible` 只能公开展示和搜索，不进推荐。

修订：

- 推荐流统一要求 `quality_tier >= recommend_candidate`。
- 新鲜探索池也统一到 `recommend_candidate`。
- `public_visible` 只进入公开展示和搜索，不进入推荐和公共 AI 引用。

状态：已一致。

### 2.2 unavailable 状态

问题：

- `unavailable` 曾被写进用户侧生命周期。
- 后台和技术架构实际生命周期不包含 unavailable。

修订：

- `unavailable` 统一为前台占位展示态。
- experiences.lifecycle_status 只使用 active / needs_review / hidden / deleted。

状态：已一致。

### 2.3 临时聊聊会话

问题：

- 产品要求临时会话可丢弃。
- 技术接口默认围绕 `chat_topics/:id/messages`，缺少临时态承接。

修订：

- 新增 `chat_temp_sessions`。
- 临时消息写入 chat_messages，topic_id 为空，temp_session_id 不为空。
- 议题清晰后创建 chat_topic 并绑定临时消息。
- 离开仍不清晰则标记 discarded，24 小时内清理明文，不进入最近聊过或 AI 上下文。

状态：已一致。

### 2.4 chat_topic_classify 触发时机

问题：

- 原文写“议题形成时触发”，但又用它判断是否形成议题。

修订：

- 统一为临时会话首轮后、连续两轮后、用户离开时触发。
- 它用于判断是否形成稳定议题，而不是议题形成后才运行。

状态：已一致。

### 2.5 内容生产 AI 链路顺序

问题：

- translation_normalization 曾排在 extract / review / classify 后面。

修订：

- 统一为 translation_normalization -> experience_extract -> experience_review -> dedupe -> experience_classify -> moderation -> experience_interpretation。

状态：已一致。

### 2.6 命名统一

问题：

- 清楚一点反馈同时出现 chat_feedback 和 chat_session_feedback。
- 私密审计同时出现 admin_audit_logs 和 admin_private_access_logs。
- 批量操作同时出现 batch_operation_id 和 operation_batch_id。

修订：

- 会话反馈统一为 `chat_session_feedback`。
- 通用后台操作审计使用 `admin_audit_logs`。
- 私密明文查看审计使用 `admin_private_access_logs`。
- 批量操作统一使用 `operation_batch_id`。

状态：已一致。

### 2.7 质量层级、推荐资格和 AI 引用资格

问题：

- `quality_tier`、`recommendation_status`、`ai_citable` 都能表达分发资格，但缺不变量。

修订：

- `quality_tier` 是质量层级事实源。
- `recommendation_status` 是推荐分发开关。
- `ai_citable` 是公共 AI 引用开关。
- 后两者默认由 quality_tier 生成，但允许人工覆盖并记录原因。
- 非 public 或非 active 时，推荐和 AI 引用必须视为不可用。

状态：已一致。

### 2.8 公开原创编辑后的外部可见性

原问题：

- 作者编辑公开原创后，作者侧显示新正文；其他用户收藏、搜索、历史引用的展示规则曾缺少统一口径。

修订：

- needs_review 期间，作者本人看到最新正文。
- 其他用户收藏、搜索结果卡片集和聊天历史引用显示 unavailable 占位。
- 重新分层通过后恢复对应公共可见性和分发资格；不通过则转私密。

状态：已一致。

### 2.9 个人信息和推荐

问题：

- 个人信息写了“推荐弱辅助”，但推荐公式没有接入个人信息字段。

修订：

- 第一阶段看看推荐不直接读取个人信息字段，避免标签化用户。
- 个人信息只用于聊聊理解。
- 由聊聊形成的近期议题处境信号可弱参与推荐。

状态：已一致。

### 2.10 画像失效

原问题：

- 用户删除议题、删除经验、取消收藏后，轻画像和推荐画像的失效规则曾缺少统一口径。

修订：

- 删除议题、删除经验、取消收藏、清空个人信息时写入 profile_invalidation 任务。
- user_memory_profiles 和 user_preference_profiles 异步重算。
- profile_version 增加后，旧推荐 cursor 只允许继续一页，刷新或新 session 使用新画像。

状态：已一致。

### 2.11 目标表清单

问题：

- 技术架构前面的第一阶段目标表漏掉工程化新增表。

修订：

- 补入 recommendation_sessions、search_sessions、processing_units、production_batch_reports、admin_private_access_logs、admin_security_events、user_preference_profiles、chat_temp_sessions、experience_public_snapshots。

状态：已一致。

### 2.12 产品决策差距修订

问题：

- 用户 PRD 对聊聊引用卡、首次公开展示名、公开原创删除转私密、专业问题轻提醒、历史议题打开、单条消息删除、平台精选前台边界等规则描述不完整。
- 管理后台 PRD 的数据看板仍偏“内容 / 互动 / AI”泛口径，未完全回到用户产品一级结构。
- 管理后台内容生产缺少排除名单、内容周报、完整来源池 / 素材字段和采集边界。
- 技术架构缺少部分承接表、引用统计字段、敏感素材审计和测试项。

修订：

- 用户 PRD 补齐引用小卡 / 完整卡差异、引用统计、匿名贡献、首次公开展示名、删除转私密、专业问题轻提醒、历史议题不自动回复和第一阶段不做事项。
- 管理后台数据看板改为用户、看看、聊聊、记下、内容供给、AI 成本六个分页。
- 管理后台补齐用户列表运营字段、反馈状态、排除名单、内容周报、来源池字段、素材字段和采集边界。
- 技术架构补齐 feedback、content_weekly_reports、source_exclusion_items、note_daily_stats、引用统计字段、敏感素材审计和测试承接。

状态：已一致。

### 2.13 P0/P1/P2 完备度修订

原问题：

- 后台 PRD 对用户原创正文留有“可改写”口子，与用户拥有原创编辑权的决策冲突。
- 候选经验曾保留人工改写和人工去重合并动作，与“运营只做通过 / 拒绝”的简化原则冲突。
- 聊聊新聊 / 续聊 / 换个事聊、临时会话接口、高风险决策策略、轻画像字段、后台看板指标、经验管理排序、AI 日志字段仍不够可实现。
- 平台内容气质、行为提炼、多语言创作者显示名和决策文档旧数据口径缺少明确承接。

修订：

- 后台 PRD 明确后台不得编辑用户原创正文；投诉或违规只能下架、转私密、删除异常内容、备注或触发复查。
- 候选经验移除编辑和人工合并操作，只保留通过 / 拒绝及必要 AI 补齐任务；通过由 promote 接口完成。
- 用户 PRD 补齐换个事聊、默认续聊、新聊判断、临时会话消息接口和临时会话记下接口。
- 用户 PRD 和技术架构补齐高风险决策策略、相关历史议题、轻画像字段和失效重算。
- 后台 PRD 补齐用户路径、新用户首日行为、上下滑切卡、推荐转化、默认公开保持率、推荐候选数量、AI 引用次数、推荐分、低反馈优先、AI call_source / model / retry / token / cost 字段。
- 内容生产补齐有态度 vs 稳妥实用周报校准、source_derivation_type、behavior_extraction 风险规则和中英文创作者名规则。
- 产品决策记录顶部标注 v4.5 修正：unavailable 不是 lifecycle_status，quality_score 采用 0-100 归一分。

状态：已一致。

## 3. 当前一致性判断

### 3.1 状态模型

一致：

- visibility：public / private。
- lifecycle_status：active / needs_review / hidden / deleted。
- quality_tier：public_visible / recommend_candidate / ai_citable / high_trust。
- recommendation_status：eligible / ineligible / suppressed。
- ai_citable：true / false。
- unavailable：前台占位展示态，不是数据库生命周期。
- domain / sub_domain：使用固定 6 个一级领域和 35 个子领域词表，用户发布可空，AI 分类和后台编辑必须校验合法关系。
- candidate_experiences.decision_status：pending / promoted / rejected。
- 待确认候选：只保存中置信或边界低置信但仍可能有价值的候选；明确低质直接丢弃并保留统计或任务日志。

### 3.2 用户侧主链路

一致：

- 推荐只使用 public + active + eligible + recommend_candidate 及以上经验。
- 搜索可展示 public_visible 及以上公开经验。
- 记下公开失败自动转私密，用户无感。
- 首次公开原创前必须补齐展示名，私密保存不要求展示名。
- 公开原创删除前优先提供转私密。
- 首次进入看看有一次性轻教学，不影响主体验。
- 聊聊支持默认续聊、换个事聊、临时会话、绑定旧议题和丢弃临时会话。
- 聊聊临时会话可丢弃，稳定议题长期保存。
- 历史议题打开不触发 AI 自动回复。
- 第一阶段不支持单条聊天消息删除。
- 高风险决策场景有独立回复和引用策略。
- 轻画像有字段、来源、使用边界和失效重算规则。
- 聊聊经验引用只从用户相关经验、收藏经验、公共高质量经验中取。
- 聊聊引用小卡和完整经验卡的交互边界已区分。
- 聊聊引用统计有 show / click / collect / inspire 和会话反馈承接。
- 我的统计不承载内容列表。

### 3.3 后台和技术承接

一致：

- 后台可管理公开、私密、精选、原创。
- 后台不得编辑用户原创正文。
- 私密明文查看必须写 private access audit。
- 数据看板按用户、看看、聊聊、记下、内容供给、AI 成本六个分页组织。
- 数据看板指标有事实源、聚合口径和关键转化定义。
- 经验管理包含来源可靠度、搜索点击、AI 引用、推荐分和低反馈优先。
- 候选经验运营动作保持通过 / 拒绝为主，不做复杂人工编辑合并。
- 内容生产有批次、处理单元、候选、入库、排除名单、周报、内容气质校准和行为提炼风险规则。
- 登录态素材长正文和长转写查看进入敏感内容审计。
- AI Gateway 按 function_type、key_alias、model、call_source、queue、budget、retry 和成本日志治理。
- 技术架构有对应表、队列、缓存和测试承接。

### 3.4 PRD 功能方案结构

一致：

- 用户 PRD 已取消后置“功能实现方案详规”和“工程化判定规则”重复章节。
- 用户侧推荐、收藏、我的、搜索、记下、聊聊、经验卡、个人信息、统计和事件口径均合并到各自功能章节。
- 管理后台 PRD 已取消后置“后台功能实现方案详规”和“后台工程化判定规则”重复章节。
- 后台侧运营总览、数据看板、经验管理、内容生产、用户与反馈、私密审计、AI 与系统的实现口径均合并到各自模块章节。
- 后续验收清单只保留验收矩阵，不再承载另一套功能方案。

## 4. 剩余风险

当前没有结构性冲突。

剩余风险属于实现阶段需要用测试覆盖的问题：

- 推荐评分公式参数是否需要调优。
- PostgreSQL queue 在大批量内容生产时是否足够。
- profile_invalidation 重算延迟是否影响用户感知。
- needs_review 期间 unavailable 占位是否会让少量收藏用户困惑。
- 首次公开展示名弹窗需要足够轻，避免增加记下负担。
- 内容采集边界需要在实际 agent 工作流中继续校验，避免账号状态变化和素材保存越界。
- 聊聊新聊 / 续聊自动判断需要用真实对话样本调阈值。
- 高风险场景识别需要避免过度触发，防止普通陪伴对话变得僵硬。

这些不阻塞进入实现拆解。

## 5. 本轮最终校验

校验范围：

- 用户 PRD、管理后台 PRD、技术架构文档、产品决策记录和本一致性报告。

校验结论：

- P0：后台不得编辑用户原创正文；候选经验不提供人工编辑后通过和人工合并重复，已在后台 PRD 和技术架构口径中统一。
- P1：聊聊新聊 / 续聊 / 换个事聊、临时会话接口、高风险决策策略、轻画像字段和失效、后台运营指标、经验管理字段、AI 日志和成本字段已补齐到功能方案层面。
- P2：内容气质比例、行为提炼来源类型、创作者中英文展示名、状态枚举、领域词表和决策文档数据模型口径已补齐。
- 功能实现完备度：未发现未落地标记。
- 一致性：未发现仍在 PRD 或技术架构中使用旧冲突口径的规则。
- 格式：目标 Markdown 文件未发现尾随空白；git diff check 通过。

## 6. 追加全面一致性检查

本轮追加检查范围：

- 产品决策记录和用户 PRD 的一致性。
- 产品决策记录和管理后台 PRD 的一致性。
- 用户 PRD、后台 PRD 和技术架构之间的字段、状态、接口、事实源和降级策略一致性。
- 功能实现方案是否仍存在目标型描述、未落地标记或互相冲突的状态口径。

追加修订：

- 产品决策记录第 15 章不再保留旧数据模型冲突口径，`lifecycle_status`、`quality_score`、`recommendation_status` 和 `candidate_experiences.decision_status` 已和 PRD / 技术架构统一。
- 用户 PRD 和技术架构补齐固定领域词表：6 个一级领域和 35 个子领域；用户发布可空，AI 分类、后台编辑和平台精选入库必须校验合法关系。
- 候选经验统一为待确认候选口径：中置信或边界低置信但仍可能有价值的内容进入候选池，明确低质直接丢弃并保留统计或任务日志。
- 候选经验后台动作统一为通过 / 拒绝；拒绝对应 decision_status=rejected，不进入前台、经验管理和推荐。
- 技术架构推荐召回硬过滤统一为 public + active + recommendation_status=eligible + quality_tier >= recommend_candidate。

追加结论：

- 产品需求与产品决策已对齐，没有发现仍需改写的产品口径差异。
- 功能实现方案已覆盖看看、搜索、记下、聊聊、我的、登录、埋点、后台运营、经验管理、内容生产、用户与反馈、私密审计、AI 与系统、技术架构承接。
- 跨功能关联已统一：经验状态、公开 / 私密、needs_review、unavailable 占位、收藏关系、推荐资格、AI 引用资格、解读资格、候选经验、轻画像失效和私密审计均有一致口径。
- 剩余风险均属于实现参数调优、性能验证和真实数据校准，不属于产品需求或功能方案缺口。

## 7. AI 功能专项一致性修订

本轮专项检查范围：

- 产品决策记录中所有 AI 相关决策。
- 用户 PRD、管理后台 PRD、技术架构中的 AI 调用、AI 治理和内容生产链路。
- 当前文档是否已经说明什么时候调用 DeepSeek、输入 payload、prompt 模板、输出 schema、阶段变体、失败降级和 eval。

检查发现：

- 原文档已经覆盖 `function_type`、key alias、队列、预算、调用日志、降级和部分业务触发时机。
- 原文档没有完整覆盖 prompt 级实现方案：每个 AI 功能的 payload、prompt 模板、输出 schema、阶段变体和 eval 只零散出现或缺失。
- 因此上一轮“功能实现方案已覆盖全部 AI 与系统”的结论，只对 AI 治理链路成立；对 AI prompt / schema / eval 层不完整。

本轮修订：

- 新增 `docs/product/niangao-ai-functional-prompt-spec-v4.md`。
- 从产品决策中抽取 AI 相关决策，覆盖聊聊、议题识别、摘要、轻画像、经验引用、记下整理、公开分层、推荐辅助、平台经验生产、翻译归一、质量审计、解读、后台治理。
- 为 11 个 `function_type` 补齐调用时机、同步性、输入 payload、prompt 模板要点、输出 schema、业务层校验、阶段变体和失败策略。
- 补齐 Prompt Registry、`schema_version`、prompt_version 灰度 / 回滚、最小 eval 集和校准批评测。
- 用户 PRD 引用 `experience_rewrite` 和 `chat` 的 AI 规格。
- 管理后台 PRD 补充 `schema_version`、prompt_version 只能选择已注册并通过 eval 的版本、后台不得直接编辑完整 prompt。
- 技术架构补充 `ai_prompt_registry`、prompt registry 加载、output_schema 校验、schema_version 解析、prompt eval 测试承接。

产品决策一致性结论：

- DeepSeek 负责 App 内全部 AI 功能和平台经验生产 AI 处理：已由 AI Gateway 和 11 个 function_type 承接。
- 聊聊要像陪伴的人、不像咨询师、不固定模板、强情绪少引用、高风险不替用户决定：已由 `chat` prompt、context flags 和 eval 承接。
- 新聊 / 续聊 / 临时会话 / 议题标题：已由 `chat_topic_classify` payload、clarity_score 和标题规则承接。
- 议题摘要、后台轻画像、删除失效：已由 `chat_summary` 输出和业务合并规则承接。
- 聊聊里轻提示记下、默认私密、确认匿名公开：已由 `chat.note_suggestion` 和 `experience_rewrite` 承接。
- AI 引用经验优先级、候选 3-5 条、私密经验只给本人、无匹配不提示：已由 chat 前置召回和 `chat` 输出校验承接。
- 用户公开经验无感审核、不适合公开自动私密、低质不推荐不引用：已由 `moderation`、`experience_review` 和业务状态规则承接。
- 推荐第一阶段不依赖实时 AI，`recommendation_ai` 只做离线检查和后续增强：已在 AI 规格和用户 PRD 中一致。
- 平台经验尽量保留原文、v3.0 质量审计、创作者归属、行为提炼风险、前 3 批校准、冷启动 3000 条：已由 `translation_normalization`、`experience_extract`、`experience_review`、`experience_classify`、`experience_interpretation` 和校准 eval 承接。
- 管理后台按功能拆 key alias、成本、队列、日志、预算，不展示真实 key：原后台 PRD 已覆盖，并补充 prompt_version / schema_version 约束。

功能完备性结论：

- AI 相关功能不再只停留在“目标型描述”或“治理字段”，已经具备可实现的产品功能方案。
- 每个 AI 功能都有调用入口、输入、输出、失败策略和评测要求。
- PRD 负责用户和后台行为，AI 规格负责 prompt / schema / eval，技术架构负责服务和数据承接，边界清晰。

剩余事项：

- 具体 prompt 文案、eval 样例文本、timeout、token 和预算阈值仍需在实现阶段落到配置、测试和样例集中。
- 这些属于实现落地和参数调优，不再是产品需求缺口。

## 8. AI Prompt 生产级升级

本轮升级背景：

- 上一版 AI 功能规格已经覆盖功能清单、调用时机、payload、输出 schema 和治理链路。
- 但 prompt 本身仍偏“模板要点”，不足以支撑真实生产环境中的年糕体验和平台经验冷启动质量。
- 主要风险包括：聊聊回复风格不可控、高风险引用策略不完整、prompt injection 边界不足、内容生产评分无锚点、经验解读可能固定模板化、eval 偏数量而非质量。

本轮修订：

- 新增 `docs/product/niangao-ai-prompt-production-spec-v4.md`，作为生产级 prompt 和 eval 的事实源。
- 为核心 AI 功能补齐完整 Prompt Pack：`chat`、`chat_topic_classify`、`chat_summary`、`experience_rewrite`、`moderation`、`translation_normalization`、`experience_extract`、`experience_review`、`experience_classify`、`experience_interpretation`、`recommendation_ai`。
- 每个 Prompt Pack 补齐 system / developer / user 分层、输入信任边界、输出 schema、质量锚点或 seed eval。
- `chat` 补齐人本主义陪伴语气、意图识别、强情绪、高风险决策、引用经验、记下提示、prompt injection 和回复质量 rubric。
- 高风险场景补齐 `source_reliability`、`source_derivation_type`、`citation_policy`，避免低可靠来源和行为提炼经验被强引用。
- `experience_extract` 补齐显性经验优先、创作者归属、行为提炼证据、过度改写控制和正反例。
- `experience_review` 补齐 1-10 分质量锚点、维度拆分、质量层级映射、误用风险和 seed eval。
- `experience_interpretation` 改为动态 3-5 个小节，不固定四段模板，和产品决策保持一致。
- eval 从“数量要求”升级为 schema / product / quality / adversarial 四层评测，并定义 chat rubric、内容生产抽样阈值和必备 golden cases。

升级后结论：

- AI 功能实现不再只是覆盖功能和治理字段，而是具备生产级 prompt 设计的主体框架。
- 仍需要在实现阶段把这些 Prompt Pack 落为实际模板文件、eval 数据集和 CI / 后台评测任务，但样例类型、评分锚点和上线门槛已经明确，不再是空白。
- 真实效果仍必须通过首批人工评测和校准批验证，不能仅凭文档宣布 prompt 已经可直接规模化上线。
