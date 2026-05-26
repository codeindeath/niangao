# DeepSeek Live Eval 结果分析与优化记录

日期：2026-05-25

测试目标：用本机 Hermes 配置中的 `DEEPSEEK_API_KEY`，真实调用 DeepSeek V4 Pro，验证年糕生产级 AI Prompt 是否能稳定支撑产品决策中的 AI 功能。

安全说明：测试过程只读取密钥用于请求，不在终端、报告或文档中输出密钥明文。

## 1. 覆盖范围

本轮 live eval 覆盖 11 个 `function_type`：

- `chat`
- `chat_topic_classify`
- `chat_summary`
- `experience_rewrite`
- `moderation`
- `translation_normalization`
- `experience_extract`
- `experience_review`
- `experience_classify`
- `experience_interpretation`
- `recommendation_ai`

共 19 个高信号 case，覆盖：

- 强情绪陪伴不引用经验。
- 聊聊中形成新理解后轻提示记下。
- 高风险离职场景不使用低可靠经验卡。
- 临时议题识别与低信息不建议建题。
- 摘要隐私抽象。
- 用户经验整理与拒绝无经验内容。
- 公开适配、隐私和危险医疗判断。
- 多语言素材归一。
- 平台经验提取、故事拒绝、质量审核。
- 固定领域分类。
- 高风险经验解读边界。
- 推荐诊断中的质量泄漏和创作者集中。

## 2. 关键发现

### 2.1 DeepSeek V4 Pro 不能裸调用

实测发现：`deepseek-v4-pro` 如果不显式传 `thinking`，部分请求会 `content` 为空，结果出现在 `reasoning_content` 中。

结论：

- 生产环境必须显式设置 `thinking`。
- 业务只能解析 `content`。
- `reasoning_content` 不得作为用户可见内容或业务兜底输出。
- `content` 为空必须视为 `empty_content` 错误。

已补入：

- `docs/product/niangao-ai-prompt-production-spec-v4.md`
- `docs/product/niangao-ai-functional-prompt-spec-v4.md`
- `docs/product/niangao-technical-architecture-v4.md`

### 2.2 只写 Schema 文档不够，必须注入模型请求

第一轮测试 0/19 通过。模型能理解任务，但只输出了简化 JSON，例如只给 `reply_text`、`note_suggestion`，没有外层 `schema_version/function_type/result`。

根因：实际请求里没有把“输出 Schema 契约”强注入 Prompt。

已修复：

- 评测脚本从 Prompt 规格中抽取 `输出 Schema` 并注入请求。
- 生产 Prompt 规格明确 Gateway 必须注入 `output_schema_contract + output_schema`。
- 技术架构补充 `output_schema_contract_ref`。

### 2.3 Prompt 策略问题集中在 5 类

第二轮测试 14/19 通过，失败项暴露出真实产品风险：

| 失败点 | 风险 | 修复 |
| --- | --- | --- |
| 议题识别输出英文分类 | 会破坏固定领域体系 | 给 `chat_topic_classify` 补固定词表和禁止创造分类 |
| 议题识别过保守 | 用户明确重复困扰也不建题 | 增加“稳定对象 + 反复场景 + 明确后果”可建题规则 |
| 摘要保留姓名 | 私密摘要可能沉淀敏感识别信息 | 增加所有摘要字段的敏感抽象硬规则 |
| 平台提取保留英文长句 | 前台经验不可读且可能超过 100 字 | 要求 `candidate_content` 默认中文，原文放 `source_excerpt` |
| 情绪碎片被强分类 | 把非经验当经验进入分类链 | 增加“先判断是否为经验，低信息置空分类”规则 |

第三轮复测 19/19 通过。

最终报告：

- `docs/product/ai-prompt-eval/deepseek-live-eval-20260525-015224.json`
- `docs/product/ai-prompt-eval/deepseek-live-eval-20260525-015224.md`

## 3. 本轮优化结果

已新增：

- `scripts/eval_deepseek_prompts.py`

该脚本支持：

- 自动读取 `docs/product/niangao-ai-prompt-production-spec-v4.md`。
- 抽取每个 Prompt Pack 的 system、developer、user template、输出 schema。
- 从环境变量或 `~/.hermes/.env` 读取 DeepSeek 配置。
- 按 function_type 设置模型参数。
- 真实调用 DeepSeek V4 Pro。
- 校验 JSON 外层包、字段、业务行为和产品边界。
- 生成 JSON 与 Markdown 报告。

已补强的生产规格：

- DeepSeek V4 Pro adapter 参数。
- 输出 Schema 注入方式。
- 空 `content` 判错。
- `reasoning_content` 使用边界。
- 议题识别固定词表。
- 摘要隐私抽象。
- 平台经验提取的中文展示、原文证据和 100 字约束。
- 低信息内容不强分类。

## 4. 人工质量抽查结论

抽查样例显示：

- 强情绪聊天回复短、稳，不急于建议，不引用经验。
- 新理解场景能生成可记下的 100 字内经验文本。
- 高风险离职场景不展示低可靠经验卡。
- 摘要能把具体姓名和收入抽象为关系角色与问题类型。
- 平台经验提取能输出中文正文，同时保留英文原文证据。
- 高风险经验解读能解释使用边界，不变成行动命令。

这些结果符合当前年糕的核心产品气质：陪伴、澄清、借鉴经验，但不替用户做决定。

## 5. 仍需进入生产前扩展的评测

19 个 live case 只能证明当前 Prompt 和接入策略通过了高信号基线，不等于已经覆盖所有生产风险。

生产上线前应继续扩展：

- `chat` 至少 50 个 case：强情绪、短消息、长消息、反复倾诉、高风险决定、经验引用、多活法对照、prompt injection。
- 内容生产至少 100 个素材 case：名人访谈、书籍片段、传记转述、英文素材、字幕噪音、鸡汤、段子、客观描述、行为提炼、高风险观点。
- 分类至少 60 个 case：6 个一级领域和全部子领域都要覆盖。
- 隐私摘要至少 30 个 case：姓名、公司、学校、收入、电话、地址、病历、关系身份。
- 推荐诊断至少 30 个 case：质量泄漏、创作者集中、领域缺口、同质化、弱画像缺失。

当前脚本已经可以作为这套 golden set 的执行器继续扩展。

## 6. Golden Set 全量实测结论

补充日期：2026-05-25

基于扩展后的 270 条 golden set，本轮继续真实调用 DeepSeek V4 Pro 做全量验证。测试覆盖聊天、议题识别、摘要、经验整理、审核、翻译归一、平台经验提取、领域分类、经验解读和推荐诊断。

最终结果：

- 命令：`python3 scripts/eval_deepseek_prompts.py --suite golden --request-timeout 90 --retries 1`
- 结果：270/270 passed。
- 覆盖：chat 50、classification 60、content_production 100、privacy_summary 30、recommendation 30。
- 函数分布：`chat` 50、`experience_classify` 60、`translation_normalization` 15、`experience_extract` 40、`experience_review` 30、`experience_interpretation` 15、`chat_summary` 30、`recommendation_ai` 30。
- 延迟：最小 2609ms，中位 5836ms，P95 22734ms，最大 39892ms。慢请求主要集中在长素材抽取和审核。
- Token 用量：prompt tokens 298087，completion tokens 98381。
- 最终报告：`docs/product/ai-prompt-eval/deepseek-live-eval-20260525-164913.md`
- JSON 报告：`docs/product/ai-prompt-eval/deepseek-live-eval-20260525-164913.json`

本轮不是一次性通过，而是经历了多轮失败定位和修复：

- 从 237/270 到 255/270：主要修复分类边界、隐私抽象、平台经验提取和推荐诊断规则。
- 从 255/270 到 268/270：主要修复高风险观点审核、领域缺口诊断、内容提取输出长度和噪音拒绝。
- 从 268/270 到 270/270：主要修复 prompt injection 表达泄露、非经验素材被强提取、风险说明缺失，以及 DeepSeek 偶发空内容 / JSON 解析失败的评测器重试策略。

Prompt 侧关键升级：

- 聊聊：补强内部规则防泄露和非机械化表达要求，避免出现“系统提示词”“内部规则”“作为 AI”等破坏陪伴感的表达。
- 摘要：健康、孩子、学校、收入等敏感细节必须抽象为问题类型、边界诉求和情绪状态。
- 平台提取：明确采集说明、评论评价、测试描述、质量审计说明和反例说明不是经验；候选经验正文必须中文、短、可直接展示。
- 审核：短而有态度的活法不能因为不完整、不温和就被误删；高风险态度观点必须进入候选审核并标风险。
- 分类：固定领域和子领域边界明确，禁止创造非法子领域。
- 推荐诊断：当候选池与当前议题严重不匹配时，AI 允许拒绝重排，而不是强行从坏候选里选。

评测器侧关键升级：

- 默认禁用系统代理，避免本机代理影响 DeepSeek 连接；需要代理时用 `--use-system-proxy` 显式开启。
- 支持 `--request-timeout` 和 `--retries`，并对 timeout、URL 错误、空 `content`、可重试 JSON 解析失败做一次重试。
- 内容生产类输出预算提高，降低长素材场景 JSON 截断风险。
- 新增 `scripts/test_eval_deepseek_prompts.py`，覆盖代理和重试逻辑。

人工质量抽查：

- prompt injection case 能拒绝内部规则泄露，并自然回到用户真实议题。
- 非经验素材能被拒绝，不会把测试说明或评论噪音炼成经验。
- 高风险观点能保留观点锋芒，但显式标注风险，不包装成强推荐。
- 隐私摘要保留“用户为什么难受”和“需要什么边界”，不保留敏感识别细节。
- 推荐诊断能识别领域缺口和候选池不可用，必要时返回空重排。

结论：

当前生产级 Prompt 规格和 eval runner 已经具备可复测的上线前基线能力。后续进入工程实现时，AI Gateway 必须按本文档和 `niangao-ai-prompt-production-spec-v4.md` 注入同一套 Prompt、Schema、模型参数和重试边界；任何 Prompt、模型参数、输出 Schema 或后处理逻辑变更，都必须重新跑 golden set。
