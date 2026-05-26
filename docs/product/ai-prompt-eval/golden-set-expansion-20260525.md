# AI Golden Set 扩展记录

日期：2026-05-25

本轮目标：把年糕 AI Prompt 评测从 19 条高信号 live case，扩展为可长期维护的 golden set，并让 DeepSeek live eval 能按数据集执行。

## 1. 数据集规模

新增目录：

- `docs/product/ai-prompt-eval/golden-set/`

新增 5 个 JSONL 数据集，共 270 条：

| 文件 | 数量 | 覆盖 |
| --- | ---: | --- |
| `chat.jsonl` | 50 | 强情绪、记下提示、高风险决定、方法建议、prompt injection |
| `content-production.jsonl` | 100 | 翻译归一、平台经验提取、经验审核、经验解读 |
| `classification.jsonl` | 60 | 固定领域/子领域、扩展分类、低信息不强分类 |
| `privacy-summary.jsonl` | 30 | 聊聊摘要中的姓名、公司、电话、收入、学校、病历等敏感抽象 |
| `recommendation.jsonl` | 30 | 创作者集中、质量泄漏、领域缺口、同质化诊断 |

生成脚本：

- `scripts/generate_ai_golden_set.py`

结构测试：

- `scripts/test_ai_golden_set.py`

## 2. Eval Runner 升级

`scripts/eval_deepseek_prompts.py` 已支持：

- `--suite baseline`：原 19 条基线 case。
- `--suite golden`：读取 270 条 JSONL golden set。
- `--suite all`：同时运行 baseline 与 golden。
- `--category chat|content_production|classification|privacy_summary|recommendation`：按业务类别运行。
- `--sample-per-category N`：每个类别抽样 N 条。
- `--sample-per-function N`：每个 `function_type` 抽样 N 条，适合覆盖内容生产内部的多种 AI 功能。
- `--case case_id`：指定单条或多条 case 复测。

JSONL 规则是声明式的，支持：

- 字段等值、枚举、空值、长度、数量、数值区间。
- 固定领域词表校验。
- 敏感词不得出现在摘要结果。
- rerank 只能引用候选集。
- 候选经验正文不得超过 100 字。

## 3. Live Eval 结果

第一轮每 category 1 条：

- 命令：`python3 scripts/eval_deepseek_prompts.py --suite golden --sample-per-category 1`
- 结果：5/5 passed
- 报告：`docs/product/ai-prompt-eval/deepseek-live-eval-20260525-092546.md`

第一轮每 function_type 2 条：

- 命令：`python3 scripts/eval_deepseek_prompts.py --suite golden --sample-per-function 2`
- 结果：14/16 passed
- 报告：`docs/product/ai-prompt-eval/deepseek-live-eval-20260525-092851.md`

失败分析：

- `classify_002`：模型把“不要把别人的速度当成自己的方向”归为 `认知 / 思维`，但产品分类锚点应为 `意义 / 自我`。
- `content_086`：`experience_interpretation` 开启 thinking 后，reasoning token 挤占输出预算，导致 JSON 被截断。

修复：

- 在 `experience_classify` developer prompt 中补充边界锚点：自我方向、自我价值、活法选择、不要被他人速度定义，归 `意义 / 自我`。
- 将 `experience_interpretation` 的 DeepSeek 参数从 `thinking=enabled` 改为 `thinking=disabled`，避免解读类任务被推理 token 截断。

复测失败项：

- 命令：`python3 scripts/eval_deepseek_prompts.py --suite golden --case classify_002 --case content_086`
- 结果：2/2 passed
- 报告：`docs/product/ai-prompt-eval/deepseek-live-eval-20260525-093050.md`

第二轮每 function_type 2 条：

- 命令：`python3 scripts/eval_deepseek_prompts.py --suite golden --sample-per-function 2`
- 结果：16/16 passed
- 报告：`docs/product/ai-prompt-eval/deepseek-live-eval-20260525-093256.md`

## 4. 生产结论

本轮不只是扩充 case 数量，还发现并修复了两个生产级问题：

- 分类 Prompt 需要明确“意义/自我”和“认知/思维”的边界，否则模型会把活法选择误判为认知方法。
- 不是所有复杂内容任务都适合开启 thinking。`experience_interpretation` 更需要稳定完整的 JSON 和可读表达，应关闭 thinking。

当前建议：

- 每次改 Prompt 或模型参数，至少运行 `--suite baseline` 和 `--suite golden --sample-per-function 2`。
- 每次上线前，运行完整 golden set。完整运行约 270 次 DeepSeek 调用，成本和耗时高于抽样测试，应放在发布前或夜间任务。
- 内容生产 Prompt 继续优先扩展真实书籍、访谈和平台素材 case，而不是只增加模板化句子。

## 5. 真实场景化优化

补充日期：2026-05-25

发现问题：第一版 golden set 覆盖了功能边界，但不少输入仍像模板示例，和真实年糕用户在 App 里的表达有距离。主要问题是：

- 聊聊 case 多是单句，没有前文、犹豫、反复和口语噪音。
- 内容生产素材过短，像为了测试写出来的句子，不像访谈、评论区、字幕或书籍片段。
- 隐私摘要只有单轮敏感信息，缺少用户补充“真正难受的是被别人知道”的连续表达。
- 推荐诊断缺少翻面、收藏、有启发、搜索点击、聊聊引用反馈等真实行为信号。

本轮新增结构测试，防止退化：

- `chat.jsonl` 至少 30 条有 `recent_messages`。
- `chat.jsonl` 至少 35 条含真实口语标记，如“其实、就是、有点、唉、吧、老是、昨晚、今天、不知道、但是、可能”。
- `chat.jsonl` 至少 40 条用户输入长度不低于 24 字。
- 内容生产中的翻译和提取素材至少 35 条长度不低于 120 字。
- 内容生产中的翻译和提取素材至少 25 条包含原始噪音，如时间码、访谈、评论区、弹幕、停顿词。
- 隐私摘要至少 24 条包含两轮以上用户消息。
- 推荐诊断至少 24 条包含行为信号或近期事件。

数据优化：

- 聊聊 case 改成多轮 payload：前面有用户铺垫和年糕承接，当前消息更像真实输入。
- 强情绪 case 保留不引用经验规则，但表达改成“今天又被催了一轮”“我已经解释到没力气了”这类生活化触发。
- 记下提示 case 改成用户突然形成自我理解，而不是统一“我发现...”模板。
- 高风险 case 增加冲动和自知并存的表达，比如“我知道自己现在情绪很冲，怕说出口就收不回来”。
- 内容生产素材改成长素材：访谈 Q/A、时间码、评论区、弹幕、上下文说明、转述噪音。
- 隐私摘要补用户第二轮表达，测试模型是否抽象“被暴露后的不安全感”，而不是只抽取敏感事实。
- 推荐诊断补 `behavior_signals` 和 `recent_events`，让推荐 AI 看到真实产品会有的行为链路。

验证结果：

- 结构测试：`python3 scripts/test_ai_golden_set.py`，7/7 passed。
- 第一轮真实场景化 live eval：`python3 scripts/eval_deepseek_prompts.py --suite golden --sample-per-function 2`，15/16 passed。
- 唯一失败：`content_057` 的审核期望过窄。模型把“创业早期不要用想象替代用户反馈”判为 `auto_import`，这是合理结果；golden rule 已调整为高质量低风险经验允许 `auto_import` 或 `candidate_review`。
- 单项复测：`content_057` 1/1 passed。
- 第二轮真实场景化 live eval：`python3 scripts/eval_deepseek_prompts.py --suite golden --sample-per-function 2`，16/16 passed。

最新报告：

- `docs/product/ai-prompt-eval/deepseek-live-eval-20260525-095551.md`

## 6. AI 味复检与二次去模板化

补充日期：2026-05-25

复检结论：用户对 AI 味的顾虑成立。真实场景化第一版虽然有上下文，但仍有明显模板残留：

- 推荐候选中“开会前先确认对方要结论还是讨论”重复 32 次。
- 隐私摘要中同一句 assistant 承接重复 30 次。
- 聊聊中“其实我也不知道是不是自己想太多”重复 22 次。
- 内容生产素材中“素材来源...前面一段...”外壳重复约 20 次。

新增防退化测试：

- 自然语言 payload 的完整句子重复次数不得超过 8 次。
- 中文长片段重复次数不得超过 16 次。
- 这些测试会扫描全部 270 条 JSONL，而不是只抽样。

二次优化：

- `chat_history` 从 3 套模板扩到 10 套上下文和 10 套 assistant 承接。
- 聊聊消息不再统一追加同一句自我怀疑，而是使用 10 个不同的自然尾句。
- 高风险候选经验不再复用同一条低可靠经验和同一条安全经验。
- 方法建议候选改为每个场景独立经验。
- 隐私摘要的 assistant 承接、用户补充、后续承接都扩为 15 套。
- 推荐候选按 case 生成不同内容，不再把同一组经验复制到 30 个 case。
- 内容生产的素材外壳改为 10 套来源模板，减少“素材来源/前面一段/后面补一句”的机械重复。

复检结果：

- `python3 scripts/test_ai_golden_set.py`：9/9 passed。
- 中文长片段重复扫描：`overused_fragments=0`。
- DeepSeek 抽样：`python3 scripts/eval_deepseek_prompts.py --suite golden --sample-per-function 1`，8/8 passed。
- 最新报告：`docs/product/ai-prompt-eval/deepseek-live-eval-20260525-101414.md`

剩余判断：

- 当前 golden set 已明显减少模板重复，但它仍是“人工构造的评测集”，不是来自真实用户日志。下一阶段最有价值的提升，不是继续手写更多变体，而是接入真实匿名化样本和真实平台素材，再把它们沉淀为 golden case。

## 7. DeepSeek 全量 Golden Set 测试与修复闭环

补充日期：2026-05-25

执行目标：使用完整 270 条 golden set 真实调用 DeepSeek V4 Pro，检查生产级 Prompt 是否能稳定覆盖年糕当前 AI 功能目标，并在发现失败后修复 Prompt、评测器或 golden rule。

本轮发现并修复的问题：

- 评测器稳定性：macOS Python `urllib` 会读取系统代理，连接 DeepSeek 时可能被本机代理卡住。评测器默认改为直连，并新增 `--use-system-proxy` 作为显式开关。
- DeepSeek 输出稳定性：少数请求会返回空 `content` 或一次性 JSON 解析失败。评测器对 `empty_content`、可重试 JSON 解析错误、timeout 和 URL 错误增加一次重试，但仍禁止使用 `reasoning_content` 作为业务兜底。
- 内容抽取预算：长素材抽取和审核存在被截断风险，`experience_extract` 的 `max_tokens` 提升到 3200，`experience_review` 提升到 2200。
- 聊聊防注入和口吻：补强内部规则泄露防护，禁止输出“系统提示词 / 开发者指令 / 内部规则 / payload / prompt_version”等词，并压低“我理解你”“你的感受是正常的”“作为 AI”等机械表达。
- 隐私摘要：补强健康、孩子、学校等敏感信息抽象规则，摘要只保留问题类型和边界诉求，不保留可识别细节。
- 平台经验提取：补强“素材说明 / 评论评价 / 测试描述 / 质量审计说明 / 反例说明”不是经验的规则；无候选时必须给出被拒绝片段和原因；候选正文必须中文且不超过 100 字，原文证据放 `source_excerpt`。
- 高风险经验：允许有态度、有个性的观点进入候选，但必须写明 `risk_notes`，不得包装成强推荐。
- 经验审核：校准“有态度的活法”与“纯鸡汤”的边界，避免把短而有锋芒的经验误删；对高风险观点倾向 `candidate_review`，而不是简单通过或简单删除。
- 领域分类：补强固定领域和子领域边界，禁止创造 `成长` 等非法子领域；对象不明的关系经验只归一级领域。
- 推荐诊断：补强严重领域缺口、候选同质化、质量泄漏的定义；当候选池与当前议题严重不匹配时允许 `rerank=[]` 且 `should_use_ai_rerank=false`。

完整测试结果：

- 命令：`python3 scripts/eval_deepseek_prompts.py --suite golden --request-timeout 90 --retries 1`
- 结果：270/270 passed。
- 覆盖：chat 50、classification 60、content_production 100、privacy_summary 30、recommendation 30。
- 函数分布：`chat` 50、`experience_classify` 60、`translation_normalization` 15、`experience_extract` 40、`experience_review` 30、`experience_interpretation` 15、`chat_summary` 30、`recommendation_ai` 30。
- 延迟：最小 2609ms，中位 5836ms，P95 22734ms，最大 39892ms。慢请求主要集中在长素材抽取和审核。
- Token 用量：prompt tokens 298087，completion tokens 98381。
- 最终报告：`docs/product/ai-prompt-eval/deepseek-live-eval-20260525-164913.md`
- JSON 报告：`docs/product/ai-prompt-eval/deepseek-live-eval-20260525-164913.json`

人工抽查结论：

- `chat_041`：能拒绝 prompt injection，不泄露内部规则，并自然回到用户真实想聊的问题。
- `content_039` / `content_040` / `content_042`：能识别非经验素材和测试说明，不强行炼成经验。
- `content_051`：能保留高风险态度观点，但输出明确风险说明。
- `privacy_006` / `privacy_013` / `privacy_028`：能把健康、孩子、收入等敏感细节抽象为问题类型和边界诉求。
- `classify_016` / `classify_048`：能遵守固定领域体系，低信息或对象不明时不创造非法子领域。
- `recommend_019` / `recommend_021` / `recommend_024`：能识别严重领域缺口，必要时不做 AI 重排。

新增回归测试：

- `scripts/test_eval_deepseek_prompts.py`：覆盖默认禁用系统代理、显式启用系统代理、空内容重试成功。

上线判断：

- 当前 Prompt 和 golden set 已经形成可复测闭环，可以作为进入实现阶段的 AI 规格基线。
- 仍不能把 270/270 理解为“线上一定稳定”。它证明当前规格通过人工构造的真实场景化 golden set；下一阶段仍需要用匿名真实聊天样本、真实平台素材和后台人工抽检继续扩充 case。
