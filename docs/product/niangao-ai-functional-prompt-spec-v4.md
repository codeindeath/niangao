# 年糕 AI 功能实现与 Prompt 设计规格 v4.5

本文档补齐年糕 v4.5 产品需求中的 AI 功能实现方案。它不是代码实现文档，而是开发、测试和运营可以共同执行的 AI 产品逻辑规格：什么时候调用 DeepSeek、传入什么结构、prompt 如何组织、输出如何校验、不同阶段如何变化、如何评测和回滚。

生产级 prompt 的完整 system/developer/user 模板、输入信任边界、质量锚点、正反例、golden cases 和上线门槛，以 `docs/product/niangao-ai-prompt-production-spec-v4.md` 为准。本文档中的“Prompt 模板要点”只作为功能级摘要，不作为最终可上线 prompt。

适用范围：

- 年糕 iOS App 内全部 AI 功能。
- 平台精选经验生产过程中的全部 AI 处理。
- 管理后台中的 AI 配置、日志、队列、成本和故障处理。
- 当前默认模型为 DeepSeek V4 Pro；模型和 key 可以按 `function_type` 切换，但产品逻辑以本文档为准。
- DeepSeek V4 Pro 必须显式设置 `thinking` 参数和 `response_format={"type":"json_object"}`；不能裸调用。若 `content` 为空，即使 `reasoning_content` 有内容，也按输出失败处理。

不适用范围：

- 自动采集素材本身的浏览器操作、账号登录、网页抓取脚本细节。
- 具体代码模块划分、数据库迁移文件、API handler 代码。
- 后续多模型实验和商业化额度体系。

## 1. 产品决策中的 AI 相关决策清单

### 1.1 全局 AI 治理

产品决策：

- 年糕 App 内全部 AI 功能由 DeepSeek 实现。
- 不同 AI 功能要拆 `function_type` 和 key alias，便于统计成本和用量。
- 用户实时请求优先于内容生产任务。
- 内容生产可以限流、暂停、恢复，不影响 App 主流程。
- 每次 AI 调用能追溯到功能、用户、批次、经验、议题或后台操作。
- prompt 内容不默认完整入库，避免保存大量私密文本。

实现承接：

- 所有 DeepSeek 调用必须经过 AI Gateway。
- AI Gateway 根据 `function_type` 读取 `ai_function_configs`，生成 prompt，调用模型，解析结构化输出，写 `ai_call_logs`。
- `ai_call_logs` 保存 `prompt_version`、输入输出 token、耗时、状态、错误码和业务对象 ID；默认只保存脱敏摘要，不保存完整 prompt。
- 入队任务保存配置快照，避免 prompt_version 或 key alias 中途变化导致结果不可追溯。

### 1.2 聊聊与长期上下文

产品决策：

- 聊聊默认参考人本主义的倾听、共情和澄清方式，但呈现上像陪伴的人，不像咨询师，不固定模板，不机械分点。
- 第一轮先接住用户表达，少下判断。
- AI 可以表达判断和偏好，但不替用户决定。
- 高风险决策更克制，做条件比较和边界分析。
- 新聊 / 续聊由 AI 和系统自动判断，用户不需要先选。
- 临时会话在主题明确后才生成稳定议题；信息不足可丢弃。
- AI 续聊使用议题摘要、相关议题、收藏经验、用户自己的经验和当前上下文。
- 摘要后台使用，用户不可见，不每轮更新，在离开聊聊或 App 进入后台时更新。
- 轻画像可用于理解、语气和经验选择，但不前台展示，不做用户管理对象。

实现承接：

- `chat` 负责生成用户可见回复。
- `chat_topic_classify` 负责临时会话是否变稳定议题、标题、领域、子领域和旧议题绑定。
- `chat_summary` 负责议题摘要、上下文压缩和轻画像候选补丁。
- `chat` prompt 必须支持场景变体：临时会话、稳定议题、历史续聊、强情绪、高风险决策、方法问题、无合适经验。

### 1.3 聊聊里沉淀经验

产品决策：

- AI 在关键时刻轻提示「这点要不要记下？」。
- 触发点包括用户说出新理解、做出决定、形成判断、被经验启发。
- 不频繁用 AI 文本打断。
- 用户确认后，AI 帮用户压缩成 100 字以内经验。
- 从聊聊生成经验默认私密，弹窗确认是否匿名公开。

实现承接：

- `chat` 输出可选 `note_suggestion`，只用于前端展示小动作，不直接保存经验。
- 用户点击「记下这点」后调用 `experience_rewrite`，输入包含用户确认的聊天片段和上下文。
- `experience_rewrite` 输出 100 字以内经验正文、领域、子领域、话题和整理置信度。
- 保存默认 `visibility=private`；用户选择匿名公开后进入公开异步处理链。

### 1.4 AI 引用经验

产品决策：

- AI 可引用经验必须公开、质量通过、内容清楚、边界明确、风险低。
- 用户自己的经验和收藏经验可作为更个性化的引用材料。
- 用户自己的私密经验只用于该用户自己的聊聊。
- 最终进入 DeepSeek 的候选经验通常 3-5 条。
- 情绪重时少引用、慢引用；方法问题可以更早引用。
- 如果没有找到合适经验，不告诉用户「没找到」，继续陪聊。
- 引用卡少出现，普通情况最多 1 张，多活法对照最多 3 张。
- AI 引用统计用于推荐分、AI 可引用池表现和内容生产反馈。

实现承接：

- 经验召回在调用 `chat` 前由后端完成，候选最多传 5 条。
- `chat` 不负责全库检索，只负责在候选中决定是否自然使用、是否展示小型引用卡。
- `chat` 输出 `citations[]`，每项包含 `experience_id`、`usage_type`、`show_card`、`reason_code`。
- 后端保存 `chat_citations`，并在用户点击、收藏、有启发、会话反馈后更新引用效果。

### 1.5 公开、审核与分层

产品决策：

- 前台不强调审核。
- 不适合公开的用户经验自动私密保存，用户无感，不解释原因，不引导修改。
- 低质量内容不一定拒绝公开，但不进推荐池，也不被 AI 引用。
- 公开经验默认可能被年糕 AI 引用，但必须经过质量分层。
- 后台分层包括公开可见、推荐候选、AI 可引用、高信任经验。

实现承接：

- `moderation` 负责安全、隐私和公开适配判断。
- `experience_review` 负责经验质量判断和初始质量分。
- `experience_classify` 负责领域、子领域、话题。
- 用户公开原创保存成功后异步进入上述流程；不通过公开适配时系统改为私密，不向用户展示失败原因。

### 1.6 推荐与匹配

产品决策：

- 推荐原则是先相关，再有用，再有差异。
- 第一阶段不要求用户维护兴趣标签。
- 推荐纠偏先靠收藏、有启发、翻面、搜索点击、AI 引用后反馈等行为信号。
- 少量反直觉经验可以混入，但用户情绪重时不急着给反向视角。
- `recommendation_ai` 第一阶段可以少用，但需要独立统计。

实现承接：

- 第一阶段实时推荐主链路使用规则召回、评分和打散，不依赖实时 AI。
- `recommendation_ai` 只用于离线评估、批量校验、候选重排实验和后续增强；失败回退规则推荐。
- `chat` 可以根据 `risk_level` 和情绪场景决定是否使用反直觉候选，但不能突破经验的 `ai_citable` 和风险限制。

### 1.7 平台经验生产

产品决策：

- 平台经验生产由 agent + DeepSeek + 年糕后台共同完成。
- agent 负责找矿和获取原料，DeepSeek 负责炼矿，后台负责把经验变成内容资产。
- 冷启动上线前至少 3000 条精选经验。
- 前 3 批每批 100 条用于校准审美、领域覆盖、创作者选择和经验提取尺度。
- 稳定后高置信内容可以自动入库。
- 平台经验尽量摘取原文核心内容，不统一改成年糕语气；超过 100 字才压缩。
- 经验质量审计标准 v3.0：删个人经历、客观描述、纯鸡汤、段子、领域过窄、空泛；清理非必要外壳和格式杂质。
- 谁真正提出经验，创作者就是谁，转述者不是创作者。
- 行为提炼类经验可以展示，但高风险场景里不作为强建议。
- 多语言内容由 DeepSeek 翻译，前台展示中文，后台保留原文短片段。

实现承接：

- 平台生产固定 AI 链为：`translation_normalization -> experience_extract -> experience_review -> dedupe -> experience_classify -> moderation -> experience_interpretation`。
- `experience_extract` 必须尽量保留原文核心表达，只做必要去壳和超过 100 字时的压缩。
- `experience_review` 必须执行 v3.0 审计标准，输出丢弃原因、质量分、分层建议和 AI 引用资格建议。
- `experience_classify` 必须使用固定 6 个一级领域和 35 个子领域，不创造新领域。
- 前 3 个校准批不直接全自动入库；AI 输出自动入库建议，批次策略决定是否需要人工确认。

## 2. AI 功能总表

| function_type | 触发时机 | 同步性 | 输入来源 | 输出结果 | 默认队列 | 失败策略 |
| --- | --- | --- | --- | --- | --- | --- |
| chat | 用户发送聊聊消息并完成上下文构建后 | 同步 | 当前消息、摘要、最近消息、相关经验、个人信息、轻画像 | 回复文本、引用决策、记下提示、风险/情绪信号 | user_realtime | 保留用户消息，AI 消息 failed 或不落库，前端可重试 |
| chat_topic_classify | 临时会话首轮后、连续两轮同主题后、用户离开时 | 可异步 | 临时消息、候选旧议题、last_chat_context | 是否生成议题、标题、领域、子领域、绑定旧议题 | user_normal | 临时标题或空分类，不阻塞聊天 |
| chat_summary | 用户离开聊聊、App 后台、token 接近上限 | 异步 | 议题消息、旧摘要、用户反馈 | 新摘要、开放问题、轻画像候选补丁 | user_background | 继续使用最近消息，稍后重试 |
| experience_rewrite | 用户主动点「帮我整理」或聊聊中点「记下这点」 | 同步或短异步 | 原始文本、聊天片段、来源场景、可选领域 | 100 字内经验、分类建议、整理理由 | user_normal | 保存原文或提示整理失败，不影响用户保存 |
| moderation | 用户公开原创入库前、平台候选入库前、必要的风险判断 | 同步/异步按场景 | 经验正文、上下文、可见性、来源 | allow_public/private_only/block_public/risk flags | user_realtime | 用户保存成功，公开分发延迟 |
| translation_normalization | 平台素材是外语、文言、繁体、字幕杂质时 | 异步 | 素材文本、语言、来源类型 | 归一文本、语言、保留表达说明 | content_low | 暂停该素材处理 |
| experience_extract | 素材进入生产批次后 | 异步 | 归一素材、来源、创作者、阶段、领域缺口 | 候选经验数组、原文依据、创作者归属 | content_low | 暂停该处理单元或进入 failed |
| experience_review | 候选经验提取后、用户公开原创异步分层时 | 异步 | 候选正文、原文依据、来源可靠度、领域 | 是否丢弃、质量分、质量层、引用资格 | content_low/user_normal | parse_error 重试 1 次，仍失败入人工或降级 |
| experience_classify | 候选通过审核后、缺分类字段时 | 异步 | 经验正文、原文片段、已有分类 | 领域、子领域、话题、置信度 | content_low | 保留待分类状态 |
| experience_interpretation | 高质量经验入库或补齐解读时 | 异步 | 经验正文、创作者、领域、来源片段 | 结构化解读 | content_low | 经验仍展示，反面暂无解读 |
| recommendation_ai | 离线推荐质量检查、批量 rerank 实验、后续增强 | 异步 | 规则候选集、用户弱画像、行为摘要 | rerank 建议、异常原因、覆盖检查 | user_normal | 回退规则推荐 |

## 3. Prompt 注册、版本和结构化输出

### 3.1 Prompt 注册

AI Gateway 维护 prompt registry。每个 prompt 条目包含：

- `function_type`。
- `prompt_version`，格式为 `function.major.minor.patch`，例如 `chat.1.0.0`。
- `schema_version`，输入输出 schema 变化时递增。
- `stage_policy`，定义哪些阶段使用同一 prompt 或变体。
- `system_template`，具体内容来自 `docs/product/niangao-ai-prompt-production-spec-v4.md`。
- `developer_template`，具体内容来自 `docs/product/niangao-ai-prompt-production-spec-v4.md`。
- `user_template`，具体内容来自 `docs/product/niangao-ai-prompt-production-spec-v4.md`。
- `output_schema_contract`，用于强制外层包、必填字段和枚举范围。
- `output_schema`。
- `parser_policy`。
- `eval_suite_id`，必须覆盖生产级 Prompt 规格中的 seed eval 和上线门槛。
- `status`：draft / active / deprecated / disabled。

规则：

- 业务 handler 不能直接拼 DeepSeek prompt。
- Prompt 只在 AI Gateway 或 AI Worker 内生成。
- prompt 变更必须生成新 `prompt_version`，不能静默覆盖 active 版本。
- 任何 active prompt 都必须通过 `docs/product/niangao-ai-prompt-production-spec-v4.md` 定义的 schema eval、product eval、quality eval 和 adversarial eval。
- `ai_call_logs` 记录 `prompt_version` 和 `schema_version`。
- 默认不保存完整 prompt；只保存脱敏 payload summary 和输出摘要。
- DeepSeek 如不稳定支持 `developer` role，则 Gateway 将 `developer_template + output_schema_contract + output_schema` 合并进 system message，但日志仍按三层模板分别记录版本。

### 3.2 通用 Prompt 结构

所有 prompt 使用五段结构：

1. 身份和任务边界：说明年糕的产品定位、该功能要做什么、不做什么。
2. 产品规则：引用与该功能相关的硬规则，例如 100 字以内、固定领域、公开无感、不得过度改写。
3. 输入上下文：由 AI Gateway 从 payload 格式化，字段名稳定。
4. 输出契约：要求返回 JSON 或指定结构。
5. 自检要求：输出前检查是否违反长度、领域词表、隐私、引用资格、原文保留等规则。

实际发送给 DeepSeek 的消息必须包含 output schema contract 和 output schema。否则模型容易只输出 `result` 内字段，导致 schema 校验失效。

### 3.3 通用输出包

除 `chat` 可返回自然语言回复字段外，所有 AI 功能都必须返回 JSON。统一外层字段：

```json
{
  "schema_version": "1.0",
  "function_type": "chat",
  "result": {},
  "confidence": 0.0,
  "warnings": [],
  "reject_reason": null
}
```

解析规则：

- 缺少 `schema_version` 或 `result`：按 parse_error。
- JSON 外包裹 markdown 代码块：解析器可清理一次。
- 字段类型错误：按 parse_error 自动重试 1 次，重试 prompt 要求“只输出合法 JSON”。
- 输出字段越权，例如创造不存在的领域：按 invalid_output，不直接入库。
- AI 不决定最终数据库状态；业务层根据规则和输出结果落状态。

### 3.4 阶段变体

Prompt 不按“重新写一套”扩散，而是使用 `stage` 和 `context_flags` 调整。

通用阶段：

- `calibration`：前 3 批内容生产，要求输出更多理由、好坏样例标记、争议点；高置信也先进入待确认或批次确认。
- `cold_start`：上线前集中生产，强调领域覆盖、创作者多样性、原文保留和质量门槛。
- `daily`：上线后日常补充，强调与用户反馈和内容缺口匹配。
- `hotspot`：热点人物或热点内容，强调不追事件、不做新闻流，只提取可沉淀经验。
- `user_realtime`：用户实时链路，强调低延迟、少字段、可降级。
- `admin_manual`：后台人工触发重跑，要求输出更完整的原因，便于运营判断。

聊聊场景 flags：

- `temp_session`：临时会话，少使用历史假设。
- `stable_topic`：稳定议题，可使用摘要和相关议题。
- `history_continue`：历史议题继续聊，不主动长摘要，不复述敏感细节。
- `strong_emotion`：回复更短，先接住，少引用或不引用经验。
- `method_question`：可以更直接给方法和经验对照。
- `high_risk_decision`：做条件比较和边界分析，不替用户做决定。
- `no_experience_match`：不提“没找到经验”，正常陪聊。

## 4. 固定领域词表

AI 分类、后台编辑、平台精选入库必须使用以下固定词表。

| 一级领域 | 子领域 |
| --- | --- |
| 意义 | 幸福、自我、情绪、使命、归属、信仰 |
| 认知 | 学习、思维、信息、工具、创造、表达 |
| 工作 | 求职、升职、创业、沟通、管理、效率 |
| 关系 | 夫妻、恋人、朋友、亲子、父母、兄妹 |
| 生活 | 宠物、旅行、衣着、养护、购物、娱乐 |
| 生命 | 健康、居住、出行、饮食、运动 |

输出校验：

- `sub_domain` 必须属于对应 `domain`。
- 用户发布可空；AI 分类和平台精选入库不可输出非法组合。
- 不确定时输出 `domain=null, sub_domain=null, confidence<0.5`，业务层进入待分类。
- AI 不得创造“心理”“成长”“家庭”等新领域名。

## 5. 各 AI 功能实现方案

### 5.1 chat

调用时机：

1. 后端校验登录和 topic/temp_session 权限。
2. 保存用户消息。
3. 必要时触发 `chat_topic_classify`。
4. 构建上下文。
5. 召回相关经验。
6. 调用 `chat`。
7. 保存 AI 消息和引用记录。

输入 payload：

```json
{
  "message_id": "msg_123",
  "user_message": "我最近很想辞职，但又怕后悔",
  "session_state": "temp_session | stable_topic | history_continue",
  "topic": {
    "topic_id": "topic_123",
    "title": "工作里的不甘心",
    "domain": "工作",
    "sub_domain": "沟通",
    "summary": "后台摘要，不直接给用户看",
    "clarity_score": 0.72
  },
  "recent_messages": [
    {"role": "user", "content": "...", "created_at": "..."},
    {"role": "assistant", "content": "...", "created_at": "..."}
  ],
  "related_topic_summaries": [
    {"topic_id": "topic_99", "title": "换工作的犹豫", "summary": "...", "similarity_reason": "..."}
  ],
  "user_profile_relevant": {
    "occupation_stage": "职场前期",
    "relationship_status": null,
    "common_troubles": ["工作压力"]
  },
  "memory_profile_relevant": {
    "common_issue_domains": [{"domain": "工作", "weight": 0.6}],
    "constraints": ["时间精力有限"],
    "preferred_experience_types": ["有态度", "具体做法"]
  },
  "candidate_experiences": [
    {
      "experience_id": "exp_1",
      "content": "经验正文",
      "creator_name": "创作者",
      "experience_type": "精选 | 原创",
      "visibility": "public | private",
      "source_relation": "own | collected | public",
      "quality_tier": "ai_citable",
      "source_derivation_type": "original_quote | behavior_extraction",
      "risk_notes": []
    }
  ],
  "context_flags": ["strong_emotion", "high_risk_decision"],
  "limits": {
    "max_reply_chars_soft": 500,
    "max_citation_cards": 1
  }
}
```

Prompt 模板要点：

```text
你是「年糕」。你参考人本主义的倾听、共情和澄清方式，但你不是治疗师，也不把自己包装成真人；你是一个会认真听、会帮用户把事情想清楚一点的陪伴者。

对话原则：
- 先回应用户已经说出的东西，再决定是否追问、澄清、给判断或借经验。
- 不固定分点，不写成报告，不使用僵硬结构。
- 可以有自己的判断，例如「我会更倾向于...」，但必须保留用户自己的决定权。
- 一次最多问一个关键问题。
- 用户情绪重时回复更短，先接住，少引用经验。
- 高风险决策只做条件比较、边界和后果提醒，不替用户决定。
- 如果没有合适经验，不要说明没有找到，继续自然回应。

经验使用规则：
- 你只能使用输入里给出的 candidate_experiences。
- 用户自己的经验可以说「你之前记下过...」。
- 收藏或公共经验可以说「有人在类似处境里...」。
- 不要为了引用而引用；情绪重时可以不用经验。
- 普通情况最多展示 1 张卡，多活法对照最多 3 张。

输出必须是 JSON，reply_text 是用户可见回复。
```

输出 schema：

```json
{
  "schema_version": "1.0",
  "function_type": "chat",
  "result": {
    "reply_text": "用户可见的自然回复",
    "citations": [
      {
        "experience_id": "exp_1",
        "usage_type": "natural_mention | card",
        "show_card": true,
        "citation_sentence": "有人在类似处境里试过...",
        "reason_code": "high_relevance | own_experience | comparison"
      }
    ],
    "note_suggestion": {
      "should_show": false,
      "suggested_text": null,
      "source_message_ids": []
    },
    "emotion_level": "low | medium | high",
    "risk_level": "normal | high_decision | safety_sensitive",
    "followup_question_count": 0
  },
  "confidence": 0.8,
  "warnings": []
}
```

业务层校验：

- `reply_text` 为空则失败。
- `citations[].experience_id` 不在候选列表则丢弃该引用。
- `show_card=true` 数量超过限制时，只保留排序最高的前 N 张。
- `note_suggestion.should_show=true` 时，必须满足：用户有明确新理解 / 决定 / 判断；当前不是强情绪；距离上次提示超过限频阈值。
- `risk_level=safety_sensitive` 第一阶段不触发复杂危机转介，只要求回复克制、不提供危险具体步骤，并可打后台风险标记。

阶段变体：

- `temp_session`：不要假设用户之前长期状态；少用轻画像。
- `history_continue`：可轻承接旧摘要，但不主动复述敏感细节。
- `strong_emotion`：`reply_text` 通常 1-3 句，`citations` 默认空。
- `method_question`：可以给 1-2 个具体选择和经验对照。
- `high_risk_decision`：不得输出“你应该立刻...”；输出条件、边界、可能后果和下一步小确认。

### 5.2 chat_topic_classify

调用时机：

- 临时会话首轮后。
- 连续两轮围绕同一问题后。
- 用户离开聊聊时。
- 稳定议题中用户明显转向新问题时，可后台补偿判断。

输入 payload：

```json
{
  "temp_session_id": "temp_1",
  "messages": [
    {"role": "user", "content": "第一条消息"},
    {"role": "assistant", "content": "回复"},
    {"role": "user", "content": "第二条消息"}
  ],
  "recent_topics": [
    {"topic_id": "topic_1", "title": "工作里的不甘心", "summary": "...", "updated_at": "..."}
  ],
  "user_clicked_new_topic": true,
  "domain_taxonomy": {}
}
```

Prompt 模板要点：

```text
判断这段临时聊天是否已经形成一个用户之后值得找回的议题。

议题标题要像真实心事，不像分类标签。
示例：写「工作里的不甘心」，不要写「工作压力问题分析」。

如果信息不足，不要硬生成标题。
如果用户点过「换个事聊」，即使命中旧议题，也默认不绑定旧议题。
领域和子领域只能从固定词表选择，不确定可以为空。
```

输出 schema：

```json
{
  "schema_version": "1.0",
  "function_type": "chat_topic_classify",
  "result": {
    "clarity_score": 0.72,
    "should_create_topic": true,
    "title": "工作里的不甘心",
    "domain": "工作",
    "sub_domain": "沟通",
    "topic_keyword": "和上级沟通",
    "candidate_existing_topic_id": null,
    "should_bind_existing_topic": false,
    "discard_if_user_leaves": false,
    "reason": "用户连续两轮都在谈同一个工作冲突"
  },
  "confidence": 0.78,
  "warnings": []
}
```

业务层规则：

- `clarity_score >= 0.65` 创建稳定议题。
- `0.45 <= clarity_score < 0.65` 保持临时会话，连续两轮后再试。
- `<0.45` 用户离开后标记 discarded。
- `user_clicked_new_topic=true` 时，不自动绑定旧议题。
- 标题为空但需要创建议题时，业务层使用临时标题，稍后重试。

### 5.3 chat_summary

调用时机：

- 用户离开聊聊。
- App 进入后台。
- 同一前台会话过长并接近 token 上限。
- 后台补偿失败摘要。

输入 payload：

```json
{
  "topic_id": "topic_1",
  "old_summary": "旧摘要",
  "messages_since_last_summary": [],
  "session_feedback": "clearer | not_much | unsure | null",
  "existing_memory_profile_refs": []
}
```

Prompt 模板要点：

```text
为后续 AI 续聊生成后台摘要，不给用户展示。

摘要目标：
- 保留议题主线、用户当前卡住的点、已经形成的判断、仍未解决的问题。
- 不写成心理诊断。
- 敏感细节只保留后续理解必要的抽象描述。
- 可以提出轻画像候选，但不要把用户固定成某种人。
- 轻画像候选必须带来源消息 ID，便于删除议题后失效。
```

输出 schema：

```json
{
  "schema_version": "1.0",
  "function_type": "chat_summary",
  "result": {
    "topic_summary": "压缩后的议题摘要",
    "current_state": "用户现在主要卡在...",
    "open_questions": ["还没确定是否要和对方谈一次"],
    "decisions_or_insights": ["用户意识到自己更在意被尊重"],
    "sensitive_detail_policy": "abstracted",
    "memory_profile_patch_candidates": [
      {
        "field": "constraints",
        "value": "时间精力有限",
        "source_message_ids": ["msg_1", "msg_2"],
        "confidence": 0.7
      }
    ]
  },
  "confidence": 0.82,
  "warnings": []
}
```

业务层规则：

- 摘要只后台使用，不显示在议题列表。
- 删除议题时，摘要和相关轻画像来源一起失效。
- `memory_profile_patch_candidates` 只进入后台轻画像候选，业务层按置信度和来源规则合并，不让 AI 直接覆盖画像。
- 摘要失败不影响继续聊天。

### 5.4 experience_rewrite

调用时机：

- 用户在「记下」页点击「帮我整理」。
- 用户在聊聊中点击「记下这点」。

输入 payload：

```json
{
  "source": "manual_note | chat_note",
  "raw_text": "用户原文或聊天片段",
  "source_message_ids": ["msg_1"],
  "default_visibility": "public | private",
  "user_selected_domain": "意义",
  "user_selected_sub_domain": "情绪",
  "topic_context": "可选上下文"
}
```

Prompt 模板要点：

```text
把用户想记下的内容整理成一条 100 字以内的经验。

规则：
- 不替用户发明没有表达过的结论。
- 不写成鸡汤。
- 保留用户的真实判断和语气。
- 只做压缩、去重复、理顺表达。
- 如果用户原文已经足够清楚，可以少改。
- 领域和子领域可参考用户选择；用户未选时再判断。
- 输出不要要求用户预览确认，保存后用户可以编辑。
```

输出 schema：

```json
{
  "schema_version": "1.0",
  "function_type": "experience_rewrite",
  "result": {
    "content": "100 字以内经验正文",
    "domain": "意义",
    "sub_domain": "情绪",
    "topic": "情绪边界",
    "rewrite_level": "light | medium",
    "source_preservation": "high",
    "needs_user_edit": false,
    "reason": "去掉重复表达，保留核心判断"
  },
  "confidence": 0.86,
  "warnings": []
}
```

业务层规则：

- `content` 超过 100 字，自动截断前必须返回给 AI 重试一次。
- 非法领域组合直接丢弃分类字段，不阻塞保存。
- 失败时保留用户原文，允许用户直接保存。
- 从「记下」页进入默认公开；从「聊聊」生成默认私密。

### 5.5 moderation

调用时机：

- 用户公开原创经验保存后，进入公开分发前。
- 平台候选经验入库为精选前。
- 公开经验编辑后重新分层。
- 聊聊中只在需要风险标记时辅助判断，不作为复杂危机流程。

输入 payload：

```json
{
  "content_type": "experience | chat_message | source_material",
  "content": "待判断文本",
  "visibility_target": "public | private",
  "source_context": "manual_note | chat_note | platform_candidate",
  "metadata": {
    "creator_type": "user | platform",
    "contains_personal_info": false
  }
}
```

Prompt 模板要点：

```text
判断内容是否适合公开展示或进入公共 AI 引用池。

必须拦截：
- 违法违规内容。
- 明显伤害性内容。
- 隐私暴露。
- 具体人身攻击。
- 仇恨歧视。
- 极端危险建议，例如自伤、违法行为、医疗偏方等。

不要帮用户改写，不要输出给用户看的解释。
低质量但不危险的内容可以公开可见，但不得推荐或 AI 引用。
```

输出 schema：

```json
{
  "schema_version": "1.0",
  "function_type": "moderation",
  "result": {
    "public_action": "allow_public | private_only | block_public",
    "risk_flags": ["privacy_exposure"],
    "risk_level": "low | medium | high",
    "recommendation_allowed": false,
    "ai_citation_allowed": false,
    "internal_reason": "含具体可识别隐私"
  },
  "confidence": 0.91,
  "warnings": []
}
```

业务层规则：

- `private_only`：用户内容自动转私密，用户无感。
- `block_public`：不进入公开分发；必要时只保留私密或标记后台复查。
- 不向用户展示 `internal_reason`。
- moderation 失败时保持待处理，不公开分发。

### 5.6 translation_normalization

调用时机：

- 素材是外语、繁体、文言、字幕转写、格式杂质明显。
- 平台经验生产链第一步。

输入 payload：

```json
{
  "source_material_id": "mat_1",
  "source_type": "book | interview | podcast | video | blog | content_platform",
  "language_hint": "en",
  "raw_text": "source text",
  "creator_name_hint": "Steve Jobs",
  "preserve_voice": true
}
```

Prompt 模板要点：

```text
把素材归一成便于提取经验的中文文本。

规则：
- 外语翻译成中文，但尽量保留原作者表达味道。
- 不翻译成年糕统一口吻。
- 文言或古文转成现代中文。
- 字幕去掉时间码、重复口癖和明显识别错误。
- 不做经验提取，不做质量判断。
```

输出 schema：

```json
{
  "schema_version": "1.0",
  "function_type": "translation_normalization",
  "result": {
    "detected_language": "en",
    "normalized_text": "归一后的中文文本",
    "translation_notes": ["保留了作者原有比喻"],
    "low_confidence_spans": [],
    "preserve_voice_score": 0.82
  },
  "confidence": 0.84,
  "warnings": []
}
```

业务层规则：

- `confidence < 0.5` 或低置信片段过多，素材进入低置信，不自动入库。
- 书籍和版权敏感来源只保存与候选经验相关短片段，不保存完整归一文本到最终经验。

### 5.7 experience_extract

调用时机：

- 素材完成归一后。
- 后台手动新增素材、CSV 导入、批次处理。

输入 payload：

```json
{
  "source_material_id": "mat_1",
  "production_batch_id": "batch_1",
  "stage": "calibration | cold_start | daily | hotspot",
  "source_type": "interview",
  "source_reliability": "high | medium | low",
  "creator_candidates": [
    {"name": "乔布斯", "role": "speaker"}
  ],
  "material_text": "归一后的素材文本",
  "target_domains": ["意义", "认知"],
  "domain_gap_context": ["意义/幸福缺口较大"],
  "max_candidates": 3
}
```

Prompt 模板要点：

```text
从素材中提取候选经验。经验是普适、能指导行为或认知变化的原则。

提取规则：
- 优先摘取原文核心内容，不做修改。
- 如果核心内容 100 字以内，尽量不改。
- 如果超过 100 字，才适当压缩，并尽量保留原文措辞。
- 不统一改成年糕语气。
- 不从故事里强行脑补道理。
- 单条候选只保留一个判断。
- 谁真正提出经验，creator_name 就是谁；转述者不是创作者。
- 行为提炼必须有明确行为依据。

删除：
- 个人经历。
- 客观描述。
- 纯鸡汤。
- 段子。
- 领域过窄。
- 8 字以内且无实质内容。
```

输出 schema：

```json
{
  "schema_version": "1.0",
  "function_type": "experience_extract",
  "result": {
    "candidates": [
      {
        "candidate_content": "100 字以内候选经验",
        "creator_name": "乔布斯",
        "creator_attribution_type": "speaker | book_author | subject | quoted_person",
        "source_derivation_type": "original_quote | cleaned_quote | compressed_quote | behavior_extraction",
        "source_excerpt": "与经验直接相关的原文短片段",
        "source_location": "章节/时间戳/段落",
        "preserve_original_score": 0.9,
        "extraction_confidence": 0.86,
        "risk_notes": []
      }
    ],
    "discarded_examples": [
      {"text": "被丢弃片段", "reason": "个人经历"}
    ]
  },
  "confidence": 0.8,
  "warnings": []
}
```

业务层规则：

- 单条短素材最多 3 条候选。
- 长文、访谈、书籍片段最多 8 条候选。
- 输出超过上限时，业务层只保留 `extraction_confidence` 和质量预估最高的前 N 条。
- `source_derivation_type=behavior_extraction` 的候选后续必须做误用风险检查。
- `stage=calibration` 时候选不直接自动入库，批次报告必须保留好/坏/争议样例。

### 5.8 experience_review

调用时机：

- `experience_extract` 后。
- 用户公开原创经验进入后台分层时。
- 后台重跑质量分层时。

输入 payload：

```json
{
  "experience_id": "可选",
  "candidate_experience_id": "cand_1",
  "content": "候选经验正文",
  "source_excerpt": "原文依据",
  "source_reliability": "high",
  "source_derivation_type": "cleaned_quote",
  "creator_name": "乔布斯",
  "domain": "认知",
  "sub_domain": "思维",
  "stage": "calibration"
}
```

Prompt 模板要点：

```text
按年糕经验质量审计标准 v3.0 判断候选是否值得保留。

定义：经验 = 普适、指导行为或认知变化的原则。
核心测试：删掉后用户损失什么？如果什么都没损失，就删。

删除 6 类：
1. 个人经历。
2. 客观描述。
3. 纯鸡汤。
4. 段子。
5. 领域过窄。
6. 空泛，尤其 8 字以内无实质内容。

清理 2 类：
1. 非必要外壳。
2. 格式杂质。

判断时区分内容质量和来源可靠度。来源可靠不等于经验质量高。
质量分使用 1-10 分。前台不展示分数。
```

输出 schema：

```json
{
  "schema_version": "1.0",
  "function_type": "experience_review",
  "result": {
    "decision": "auto_import | candidate_review | discard",
    "delete_category": null,
    "cleaned_content": "必要清理后的正文",
    "cleaning_level": "none | shell_only | compressed",
    "ai_quality_score": 8.2,
    "quality_tier": "public_visible | recommend_candidate | ai_citable | high_trust",
    "ai_citable": true,
    "recommendation_eligible": true,
    "misuse_risk_level": "low | medium | high",
    "misuse_risk_notes": [],
    "review_reason": "内容有清晰判断和可迁移性",
    "needs_human_attention": false
  },
  "confidence": 0.84,
  "warnings": []
}
```

业务层规则：

- `ai_quality_score` 是 DeepSeek 原始 1-10 分；系统另行归一为 `quality_score` 0-100。
- `decision=discard` 不进入候选池，只保留统计或任务日志。
- `candidate_review` 进入待确认候选，后台只做通过 / 拒绝。
- `auto_import` 只有在批次阶段允许、创作者明确、重复置信低、来源可靠度不低、误用风险不高时才自动入库。
- `cleaning_level` 超过 `shell_only` 时，必须保留原文短片段供追溯。

### 5.9 experience_classify

调用时机：

- 候选通过审核后。
- 用户公开原创分层时。
- 经验编辑后 needs_review 重新处理。
- 候选通过但缺领域、子领域、话题时补跑。

输入 payload：

```json
{
  "content": "经验正文",
  "creator_name": "创作者",
  "source_excerpt": "可选原文片段",
  "existing_domain": null,
  "existing_sub_domain": null,
  "domain_taxonomy": {}
}
```

Prompt 模板要点：

```text
把经验分类到固定领域词表。

只能选择固定 6 个一级领域和对应 35 个子领域。
不要创造新领域。
话题是自由短词，用于搜索和理解，不放入轻标签。
如果不确定，可以给低置信度。
```

输出 schema：

```json
{
  "schema_version": "1.0",
  "function_type": "experience_classify",
  "result": {
    "domain": "意义",
    "sub_domain": "情绪",
    "topic": "情绪边界",
    "confidence": 0.82,
    "alternative": [
      {"domain": "关系", "sub_domain": "朋友", "confidence": 0.31}
    ],
    "reason": "核心是识别并处理自己的情绪边界"
  },
  "confidence": 0.82,
  "warnings": []
}
```

业务层规则：

- 非法领域组合直接 rejected output，不写入经验。
- 低置信输出进入待分类，不阻塞私密保存。
- 搜索使用 `topic`，但前台经验轻标签不混入 topic。

### 5.10 experience_interpretation

调用时机：

- 达到 `recommend_candidate`、`ai_citable` 或 `high_trust` 的公开经验入库后。
- 平台精选高质量经验优先。
- 私密和低质经验不生成解读。

输入 payload：

```json
{
  "experience_id": "exp_1",
  "content": "经验正文",
  "creator_name": "创作者",
  "domain": "意义",
  "sub_domain": "自我",
  "source_excerpt": "可选原文依据",
  "quality_tier": "ai_citable",
  "source_derivation_type": "original_quote"
}
```

Prompt 模板要点：

```text
为一条高质量经验生成反面/详情解读，帮助用户理解如何运用。

解读要与经验正文有 UI 区分，不要把经验正文扩成一大段。
内容可以比经验正文长，但要克制，不长篇大论。
可以结构化呈现背景、适合谁、怎么理解、怎么运用、边界。
不要做心理治疗式分析。
如果是行为提炼或高风险主题，必须强调适用边界。
```

输出 schema：

```json
{
  "schema_version": "1.0",
  "function_type": "experience_interpretation",
  "result": {
    "sections": [
      {"title": "这条经验在说什么", "content": "简短解释"},
      {"title": "什么时候有用", "content": "适用背景"},
      {"title": "怎么用", "content": "具体理解和使用方式"},
      {"title": "边界", "content": "不适用或容易误用的地方"}
    ],
    "overall_length": "medium",
    "risk_boundary_required": false
  },
  "confidence": 0.82,
  "warnings": []
}
```

业务层规则：

- 只对公开且高质量经验生成。
- 失败不影响经验展示。
- 解读更新不改变经验正文、质量分和推荐资格。

### 5.11 recommendation_ai

调用时机：

- 第一阶段不进入实时推荐主链路。
- 每日或每周离线抽样检查推荐结果。
- 后台手动选择某个用户画像或推荐 session 做质量诊断。
- 后续若接入 AI rerank，只能对规则候选集做重排，不能绕过硬过滤。

输入 payload：

```json
{
  "recommendation_session_id": "rec_1",
  "user_signal_summary": {
    "recent_topic_domains": ["工作"],
    "positive_domains": ["意义", "认知"],
    "negative_or_absent": []
  },
  "candidate_experiences": [
    {"experience_id": "exp_1", "content": "经验", "score": 0.72, "creator_name": "创作者"}
  ],
  "business_rules": {
    "hard_filter": "public + active + eligible + recommend_candidate_or_above",
    "diversity_required": true
  }
}
```

Prompt 模板要点：

```text
检查这组推荐候选是否符合年糕原则：先相关，再有用，再有差异。

你不能新增候选，不能推荐不在列表里的经验。
你不能突破硬过滤。
你只输出排序建议、问题诊断和覆盖缺口。
```

输出 schema：

```json
{
  "schema_version": "1.0",
  "function_type": "recommendation_ai",
  "result": {
    "rerank": [
      {"experience_id": "exp_1", "rank": 1, "reason": "与近期议题相关且质量高"}
    ],
    "diagnostics": {
      "too_similar": false,
      "creator_concentration": false,
      "domain_gap": ["关系"]
    },
    "should_use_ai_rerank": false
  },
  "confidence": 0.78,
  "warnings": []
}
```

业务层规则：

- 实时推荐接口默认不等待 `recommendation_ai`。
- AI rerank 结果只能作为实验或后台诊断，第一阶段线上主排序仍以规则分为准。
- AI 输出不能改变经验初始质量分。

## 6. 跨功能调用链

### 6.1 聊聊消息链路

```text
User sends message
  -> Save user message
  -> chat_topic_classify if needed
  -> Build context
       topic summary
       recent messages <= 12
       related topic summaries <= 2
       relevant user profile
       relevant memory profile
  -> Recall experiences
       own experiences
       collected experiences, newest 50 within related domain/topic
       public high-quality ai_citable pool
       final candidates <= 5
  -> AI Gateway chat
  -> Validate reply JSON
  -> Save assistant message
  -> Save chat_citations
  -> Return reply and optional citation cards
```

一致性规则：

- 经验召回失败不阻塞聊天。
- AI 没有引用经验时，不向用户解释。
- 用户私密经验只进入该用户自己的候选。
- 历史议题打开不自动触发 AI 回复；用户发新消息后才走链路。

### 6.2 记下 / 公开原创链路

```text
User writes note
  -> optional experience_rewrite
  -> Save original/rewritten content
  -> if default public
       moderation
       experience_review
       experience_classify
       set quality_tier / recommendation_status / ai_citable
       if not suitable public -> visibility private, user unaware
  -> if private
       no public distribution
       no public AI citation
```

一致性规则：

- 记下页默认公开，有弱公开/私密按钮。
- 聊聊生成经验默认私密，并弹窗确认是否匿名公开。
- 用户发布经验时不做预览确认，允许事后编辑。
- 用户编辑公开经验后进入 needs_review，重新分层。

### 6.3 平台精选生产链路

```text
Agent/server collects source material
  -> source_materials
  -> production_batch / processing_unit
  -> translation_normalization
  -> experience_extract
  -> experience_review
  -> dedupe
  -> experience_classify
  -> moderation
  -> auto import or candidate pool
  -> experience_interpretation for high-quality items
  -> batch report / weekly report
```

一致性规则：

- 前 3 个校准批必须输出校准报告，不直接以数量为目标。
- 高置信自动入库只在校准稳定后启用。
- 明确低质直接丢弃，不进入后台候选池。
- 待确认候选只需要运营通过 / 拒绝。
- 后台可批量下架、降权、复查批次。

### 6.4 轻画像更新链路

```text
chat_summary
  -> memory_profile_patch_candidates
  -> business rule merge
  -> user_memory_profiles
  -> profile_version++
  -> recommendation cursor invalidation
```

一致性规则：

- 用户前台不查看、不编辑、不删除轻画像。
- 删除议题、删除经验、取消收藏、清空个人信息时写入 profile_invalidation。
- 轻画像只弱参与推荐，主要用于聊聊理解和经验选择。
- AI 不向用户显性说明“你是某类人”。

## 7. Prompt 版本、灰度和回滚

版本规则：

- 小文案、少量约束调整：patch，例如 `chat.1.0.1`。
- 输出字段增加但兼容旧解析：minor，例如 `chat.1.1.0`。
- 输出 schema 破坏性变化：major，例如 `chat.2.0.0`。

发布规则：

- 新 prompt 先在 staging 或后台手动样例跑 eval。
- 通过 eval 后设为 active。
- 内容生产类 prompt 可按 batch 灰度。
- 用户实时 `chat` prompt 首次上线不做大规模多版本并行，先保持单 active 版本，降低排障复杂度。
- 出现 parse_error、投诉、回复风格异常、引用越权时，回滚到上一 active 版本。

日志规则：

- `ai_call_logs` 必须记录 `prompt_version`。
- prompt 完整内容默认不入库。
- eval 样例可以保存脱敏输入输出，用于回归。

## 8. AI 评测方案

### 8.1 上线前必须有的 eval

| function_type | 最小 eval 集 | 通过标准 |
| --- | --- | --- |
| chat | 60 条真实/模拟对话，覆盖强情绪、方法问题、高风险决策、无经验、引用自己经验、引用公共经验 | 无越权引用；强情绪不过度建议；高风险不替用户决定；JSON 解析成功率 >= 98% |
| chat_topic_classify | 40 条临时会话，覆盖信息不足、明确议题、换个事聊、命中旧议题 | clarity_score 分段符合人工判断；标题不像分类标签 |
| chat_summary | 30 条多轮议题 | 摘要能续聊；不写心理诊断；轻画像候选有来源 |
| experience_rewrite | 50 条用户原文和聊天片段 | 100 字以内；不发明用户未表达结论；保留原意 |
| moderation | 80 条公开/私密/风险样例 | 高风险拦截；低质量不误判为违规；不输出用户可见原因 |
| translation_normalization | 30 条英文、繁体、文言、字幕 | 语义准确；不统一年糕腔；保留表达味道 |
| experience_extract | 80 条素材，含访谈、书籍、播客、普通平台内容 | 不脑补；候选数量合规；创作者归属正确 |
| experience_review | 120 条候选，覆盖删除 6 类和高质量经验 | 删除类别准确；质量分层与人工基准一致率 >= 85% |
| experience_classify | 100 条经验 | domain/sub_domain 合法率 100%；人工一致率 >= 85% |
| experience_interpretation | 40 条高质量经验 | 解读有边界；不与正文混成大段；不治疗化 |
| recommendation_ai | 20 组候选集 | 不新增候选；不突破硬过滤；诊断有用 |

### 8.2 校准批评测

前 3 个内容生产校准批每批结束后必须检查：

- 自动提取候选数量。
- 被丢弃数量和原因分布。
- 好 / 坏 / 争议样例。
- 创作者归属错误数。
- 过度改写比例。
- 纯鸡汤漏检数。
- 行为提炼误用风险漏检数。
- 领域 / 子领域覆盖。
- 可 AI 引用数量。

如果出现以下情况，不进入高置信自动入库：

- 过度改写明显。
- 纯鸡汤或客观描述漏检高。
- 创作者归属错误多。
- 行为提炼经常脑补。
- 字段非法或 JSON 解析不稳定。

### 8.3 聊聊评测

聊聊评测不只看“回答是否正确”，还要看产品气质：

- 是否像陪伴的人，而不是咨询师或客服。
- 是否没有僵硬结构。
- 是否先接住用户，再追问或判断。
- 是否不过度引用经验。
- 是否没有为了展示记忆而展示记忆。
- 是否在高风险决策里克制。
- 是否能自然从用户自己的经验、收藏经验、公共经验中选择。

## 9. 与 PRD / 技术架构的关系

权威边界：

- 用户 PRD 定义用户可见行为、流程和状态。
- 管理后台 PRD 定义运营、配置、日志、权限和审计。
- 技术架构定义服务、数据、队列、缓存、部署和测试承接。
- 本文档定义 AI 功能的 prompt、payload、输出 schema、阶段变体和 eval。

如果出现冲突：

1. 用户可见行为以用户 PRD 为准。
2. 后台操作和权限以后台 PRD 为准。
3. AI prompt 结构、输出 schema 和评测以本文档为准。
4. 数据落库和服务边界以技术架构为准。

## 10. 本轮完备性校验

对照产品决策，本轮已补齐：

- DeepSeek 何时调用：每个 `function_type` 均有触发时机。
- 调用输入是什么：每个功能均有 payload 示例。
- 输出如何被系统使用：每个功能均有 output schema 和业务层校验。
- prompt 如何写：每个功能均有 prompt 模板要点和硬规则。
- 不同阶段如何变化：有通用阶段和聊聊场景 flags。
- AI 引用、轻画像、公开分层、平台生产之间的关联：已在跨功能调用链中统一。
- 失败如何降级：已按功能定义。
- 如何验证 prompt 可靠：已定义最小 eval 集和校准批评测。

仍属于实现阶段调参而非产品方案缺口：

- 具体 timeout、token 上限和成本阈值。
- eval 样例集后续可以继续扩充，但第一批 golden cases、评分锚点和上线门槛已经在 `docs/product/niangao-ai-prompt-production-spec-v4.md` 中定义，不再是产品方案缺口。
- 各领域质量分阈值是否随真实数据调整。
- `recommendation_ai` 何时从离线诊断升级为线上 rerank。
