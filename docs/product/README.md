# 年糕新版产品文档索引

本目录是 2026-05-24 之后的新版产品与架构基准。

## 文档

- `niangao-user-prd-v4.md`：用户产品 PRD v4.5，覆盖 iOS App 的看看、聊聊、记下、我的；每个功能章节内合并目标、交互和实现方案。
- `niangao-admin-prd-v4.md`：管理后台 PRD v4.5，覆盖运营总览、数据看板、经验管理、内容生产、用户与反馈、AI 与系统；每个后台模块内合并运营目标和实现方案。
- `niangao-technical-architecture-v4.md`：新版技术架构设计 v4.5，已按 PRD v4.5 的功能实现方案做一致性校验。
- `niangao-ai-functional-prompt-spec-v4.md`：AI 功能实现与 Prompt 设计规格 v4.5，覆盖 DeepSeek 调用时机、payload、prompt 模板、输出 schema、阶段变体和 eval。
- `niangao-ai-prompt-production-spec-v4.md`：生产级 AI Prompt 规格 v4.6，覆盖完整 Prompt Pack、输入信任边界、评分锚点、正反例、golden cases 和上线门槛。
- `ai-prompt-eval/golden-set-expansion-20260525.md`：AI golden set 扩展记录，覆盖 270 条 case 的结构、运行方式、live eval 结果和本轮 Prompt/参数修正。
- `niangao-consistency-check-v4.5.md`：PRD v4.5 跨功能一致性检测报告。
- `niangao-prd-decision-alignment-v4.4.md`：PRD v4.4 与产品决策记录的详细对齐检查。
- `niangao-prd-decision-alignment-v4.3.md`：PRD v4.3 与产品决策记录的对齐检查。
- `niangao-gap-analysis-v4.md`：新版设计与现有实现差异清单。
- `../product-decisions-2026-05-24.md`：产品讨论和决策原始汇总。

## 设计稿

- `../design-drafts/niangao-main-pages-three-directions.html`：用户产品四个主页面的三版视觉方向。

## 基准原则

- 以最新版产品设计为准。
- 旧实现能用则迁移。
- 旧实现如果会造成新版产品心智混乱，应删除或重构。
- 不为了兼容历史实现保留不伦不类的中间态。
