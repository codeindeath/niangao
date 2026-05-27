# 年糕管理后台 PRD v3.0

> 基于需求框架 v0.2 + 前后端代码审计 + 大厂PRD标准 完善 | 2026-05-07

---

## 〇、产品策略

### 〇.1 产品愿景

年糕管理后台是年糕 App 的**唯一运营控制中心**，让单人运营者能高效完成内容审核、用户管理、数据分析和系统配置，保障社区内容质量，用数据驱动产品迭代决策。

**核心价值**：用最小的人力成本，维持一个有质量保障的经验分享社区。

### 〇.2 成功指标

| 维度 | 指标 | 目标值 |
|------|------|--------|
| 审核效率 | 单条经验平均审核耗时 | < 30 秒 |
| 内容健康 | AI 审核通过率（不高于） | < 70%（过高说明标准太松） |
| 用户健康 | 违规用户发现到处理的平均时延 | < 1 小时 |
| 运营覆盖 | 待审核积压（超过 1h 未审） | < 5 条 |

### 〇.3 用户画像与使用场景

**唯一管理员**（产品运营者本人）:
- 使用时段：工作日碎片时间（手机远程查看） + 晚间集中处理（桌面端）
- 设备：Chrome 桌面端为主，偶尔手机浏览器应急
- 核心痛点：内容质量需要把关但人力有限、需要快速识别问题内容和用户、想了解社区增长趋势

**日常运营节奏**：
```
早晨（手机）  → 看一眼 Dashboard：昨日新增多少用户/经验？有没有积压审核？
午间（手机）  → 快速过几条待审核内容，通过/拒绝
晚间（桌面）  → 集中处理审核队列、查看统计趋势、调整配置
周末         → 批量生产平台经验、维护领域标签
```

### 〇.4 MVP 范围与优先级

| 优先级 | 模块 | 理由 |
|--------|------|------|
| **P0** 必须上线 | 登录认证、仪表盘、内容审核（队列+操作）、内容管理 | 不开工无法运营 |
| **P1** 快速迭代 | 用户管理、平台内容管理、系统配置 | 运营第一周内需要 |
| **P2** 体验完善 | 领域管理、数据统计、AI 服务管理、操作日志 | 运营稳定后优化用 |
| **P3** 锦上添花 | CSV 导出高级筛选、管理员多账号、角色权限 | 单人可以不做 |

---

### 〇.5 项目目标
为年糕 App 提供 Web 端管理后台，覆盖用户管理、内容审核、内容管理、数据分析、系统配置等运营场景。

### 〇.6 使用角色
- **唯一管理员**（单人运营，不涉及多角色权限）

### 〇.7 技术定位
- Web 单页应用（SPA），Chrome 桌面端使用
- 通过管理后台专用 API 与后端通信
- 管理后台 API 复用现有后端认证体系（JWT + 管理员标记）

---

## 一、总览仪表盘

### 1.1 核心指标卡片
- 展示 8 个实时指标：总用户数 / 总经验数 / **今日新增用户** / **今日新增经验** / 今日活跃用户 / 今日 AI 对话次数 / 待审核数 / 今日审核通过数
- 每个卡片含主数值 + 变化趋势副文本（↑/↓ vs 昨日）
- 待审核卡片红色高亮；新增指标卡片绿色边框高亮
- 点击"待审核"卡片 → 跳转审核队列

### 1.2 趋势折线图
- 近 7 天新增用户曲线
- 近 7 天新增经验曲线（按 UGC/平台分层）
- 近 7 天活跃用户曲线

### 1.3 审核管道概览
- 待审核数量
- 今日审核通过数
- 今日审核拒绝数
- **积压数量**（超过 1 小时未审核的经验数，红色高亮）

### 1.4 AI 服务状态面板
- 模型名称、API 连通性指示灯（绿/黄/红）
- 近 1 小时调用量、平均延迟、**失败率**、成功率
- 异常时标红并显示最近错误信息

### 1.5 审核队列预览
- 展示最新 5 条待审核经验
- 每条含：内容摘要(30字截断)、领域标签、AI判定标签(通过/拒绝+分数)、提交时间
- "查看全部"链接跳转审核页

### 1.6 快捷操作入口
- 一键跳转审核队列（按钮，始终可见）
- 一键跳转用户列表

---

## 二、用户管理

### 2.1 用户列表
| 功能 | 细节 |
|------|------|
| 分页 | 每页 20 条 |
| 展示字段 | 头像、昵称、称号、登录方式(Apple/Dev图标)、注册时间、状态(正常/禁用) |
| 搜索 | 按昵称模糊搜索、按 ID 精确搜索 |
| 筛选 | 登录方式 / 状态 / 注册时间范围 / 有无称号 |
| 排序 | 注册时间(默认) / 经验数 / 最近活跃 |
| 导出 | CSV 导出当前筛选结果 |

### 2.2 用户详情
- 基本信息区：ID、昵称、头像、称号、简介、注册时间、最近登录时间
- 登录信息区：登录方式、Apple 邮箱（如有）
- 统计区：发布数 / 获点赞 / 被收藏 / 浏览数 / 我点赞 / 我收藏 / 对话次数 / 消息条数
- 领域偏好图：按（发布×2 + 收藏×1）权重展示领域分布横向柱状图

### 2.3 用户经验列表
- 该用户发布的全部经验（含私密、未通过、已删除）
- 展示：内容摘要、领域、审核状态标签、发布时间
- 点击跳转经验详情页

### 2.4 用户状态管理
- 禁用：账号不可登录、不可发布，已有内容保留
- 启用：恢复
- 批量操作：勾选多条 → 批量禁用/启用
- 禁用时填写理由（内部备注）

---

## 三、内容审核

### 3.1 审核队列
| 功能 | 细节 |
|------|------|
| 分页 | 每页 20 条 |
| 排序 | 提交时间升序（最旧优先） |
| 展示字段 | 复选框、内容摘要(40字截断)、领域标签、AI判定标签、提交时间、操作按钮 |
| AI判定标签 | 通过(绿色+分数) / 拒绝(红色+分数) / 硬策略拒绝(红色"硬拒绝") |
| 高亮规则 | 硬策略拒绝行淡红背景、AI拒绝行淡橙背景 |

### 3.2 筛选
- 领域筛选（一级领域下拉）
- 来源类型（平台 / UGC / 全部）
- AI 判定结果（通过 / 拒绝 / 未判定）
- 提交时间范围

### 3.3 经验详情弹窗（审核视角）
左侧：经验内容 + 作者 + 领域 + 来源类型 + 审核状态  
右侧：AI 六维打分明细 + 解读内容  
底部：操作区（通过 / 拒绝 / 退回重审 / 备注）

### 3.4 单条审核操作
- **通过**：确认通过，覆盖 AI 判定
- **拒绝**：弹窗填写拒绝理由（必填）
- **退回重审**：强制重新调用 AI 审核接口
- **备注**：内部备注（用户不可见），附加到审核日志

### 3.5 批量操作
- 全选当前页
- 批量通过
- 批量拒绝（统一填写理由）
- 操作结果提示（成功 N 条 / 失败 M 条）

### 3.6 AI 打分复盘
- 查看单条经验的历史 AI 打分记录
- 标记"AI 误判"（误通过 / 误拒绝）
- 标记统计面板（用于调优 prompt）

### 3.7 审核日志
- 每条经验的操作时间线
- 展示：时间、操作类型（提交/AI审核/人工通过/人工拒绝/退回）、操作人、理由

---

## 四、内容管理

### 4.1 全量经验列表
| 功能 | 细节 |
|------|------|
| 分页 | 每页 20 条 |
| 展示字段 | 内容摘要、作者(昵称/创作者名)、领域、审核状态、发布状态、互动数(赞/藏/览)、发布时间 |
| 搜索 | 内容关键词、创作者名/昵称 |
| 筛选 | 领域 / 来源类型 / 审核状态 / 发布时间范围 |
| 排序 | 时间(默认) / 点赞数 / 收藏数 |

### 4.2 经验详情页
- 完整内容 + AI 解读
- 互动数据：点赞数 / 收藏数 / 浏览数
- 审核历史时间线
- 作者信息（可点击跳转用户详情）
- 操作按钮：编辑 / 下架 / 删除 / 修改审核状态

### 4.3 经验编辑
- 修改内容
- 修改领域（一级 + 二级联动）
- 修改来源类型（平台 / UGC）
- 修改创作者名称
- 修改打分理由（≤15 字，后端 rune 计数校验）
- 保存后记录操作日志

### 4.4 经验状态管理
- **软删除**：标记 deleted_at（已入池经验）
- **恢复**：清除 deleted_at
- **硬删除**：物理删除（未入池经验）
- **下架**：从推荐池移除（保留数据）
- **修改审核状态**：手动改为通过/拒绝/私密

### 4.5 批量操作
- 批量修改领域
- 批量下架
- 批量软删除
- 导出 CSV

### 4.6 互动数据
- 查看某条经验的点赞用户列表
- 查看收藏用户列表
- 查看浏览用户列表

---

## 五、平台内容管理

### 5.1 平台经验列表
- 展示 source_type=platform 的所有经验
- 字段：内容摘要、创作者名、领域、打分/理由、发布时间、有无 AI 解读
- 筛选：领域 / 有无解读 / 有无打分
- 搜索：创作者名、内容关键词

### 5.2 新建平台经验
- 表单：内容(10-100字)、一级领域(下拉)、二级领域(联动)、创作者名称、来源标签、打分理由(≤15字)
- 保存草稿：review_status=pending，不公开
- 直接发布：review_status=approved

### 5.3 编辑平台经验
- 同 4.3，额外可编辑创作者名称和来源标签

### 5.4 批量导入
- 上传 CSV/JSON 文件
- 预览界面：展示解析结果，标注校验失败项
- 确认导入
- 结果报告：成功 N 条 / 失败 M 条（含失败原因）

### 5.5 批量 AI 处理
- 选中未评分的平台经验
- 一键触发批量 AI 打分+解读
- 进度条展示
- 结果报告
- 失败项可重试

### 5.6 单条重打分
- 对已有评分的平台经验 → 重新调用 AI 覆盖

---

## 六、领域与标签管理

### 6.1 一级领域列表
- 表格：显示名(中文)、标识名(英文)、图标、子领域数、经验总数
- 行内编辑
- 拖拽排序

### 6.2 一级领域编辑
- 新增：填写显示名/标识名/图标
- 编辑：修改显示名/图标
- 禁用：禁用后发布页不可选，已有经验不变

### 6.3 二级领域列表
- 按父领域分组展示
- 每项：显示名、标识名、所属父领域、经验数
- 组内拖拽排序

### 6.4 二级领域编辑
- 新增：选择父领域 + 填写显示名/标识名
- 编辑：修改显示名、修改所属父领域（迁移经验）
- 禁用/启用

---

## 七、AI 服务管理

### 7.1 服务状态
- API 连通性检测按钮 + 自动定时检测
- 显示当前模型名称、API 地址

### 7.2 调用统计
- 按接口分组（review / chat/send / generate-interpretation / **图像生成**）
- 时间维度：今日 / 昨日 / 近7天 / 近30天
- 指标：调用次数、平均延迟、成功率、失败率
- 折线图 + 表格

### 7.3 批量任务管理
- 批量打分任务列表
- 进度条（已完成/总数/失败数）
- 取消进行中任务
- 重试失败项

### 7.4 费用估算（可选，需配置 API 单价）
- Token 消耗统计
- 费用估算
- 日/月趋势

### 7.5 Prompt 查看
- 审核 prompt 模板（只读）
- 解读生成 prompt 模板（只读）
- 对话系统 prompt 模板（只读）

---

## 八、数据统计

### 8.1 用户增长
- 日/周/月新增用户折线图
- 累计用户面积图
- 按登录方式分层

### 8.2 内容增长
- 日/周/月新增经验折线图
- 按来源（平台/UGC）分层
- 按审核结果过滤

### 8.3 互动统计
- 日点赞/收藏/浏览趋势折线图
- 人均互动次数

### 8.4 领域分布
- 一级领域经验数饼图
- 二级领域经验数横向柱状图
- 时间范围筛选

### 8.5 审核统计
- AI 通过率趋势
- 硬策略拦截率
- 人工覆盖 AI 比例（分两个方向统计：**"AI 通过 → 人工拒绝"** 和 **"AI 拒绝 → 人工通过"**，分别展示趋势）

### 8.6 AI 使用统计
- 日 AI 对话用户数 + 人均轮次
- 解读生成次数趋势

### 8.7 留存概览（需埋点）
- 次日 / 7 日留存率

### 8.8 导出
- 所有图表可导出 CSV

---

## 九、系统配置

### 9.1 审核设置
- **模式切换**：全自动（AI 判定即最终）/ 人工复核（AI 判定后需人工确认）
- **硬策略开关**：逐条开关，可独立启用/禁用各项硬策略：
  - 字数下限检测
  - 有意义内容检测（汉字/字母存在性）
  - 敏感词过滤
  - 重复字符检测

### 9.2 内容限制
- 经验字数上下限、解读字数上限、称号字数上限
- 输入框直接修改，保存生效

### 9.3 频率限制
- 单人每日发布上限
- 单人每日 AI 对话轮次上限
- **全局 AI API 调用频率限制**（防止 API 费用失控）

### 9.4 敏感词管理
- 敏感词列表（表格，支持搜索）
- 新增 / 删除
- 批量导入（每行一词）
- 启用/禁用敏感词检查

### 9.5 功能开关
- 注册开关（关闭后新用户无法登录）
- AI 解读生成开关
- **搜索功能开关**（关闭后移动端搜索入口隐藏）

### 9.6 登录方式
- 查看已启用登录方式
- Dev Login 生产环境强制关闭（不可开启）

---

## 十、操作日志

### 10.1 日志列表
- 表格 + 分页
- 字段：时间、操作人、操作类型、操作对象（类型+ID）、详情（JSON）、结果
- 筛选：操作人、操作类型、时间范围

### 10.2 操作类型覆盖
- 审核（通过/拒绝/退回）
- 内容编辑
- 内容删除/下架/恢复
- 用户禁用/启用
- 配置变更
- 批量操作

### 10.3 日志保留
- **保留周期**：90 天（可配置），超期日志自动清理
- **日志导出**：支持按筛选条件导出 CSV（用于审计或离线分析）

### 10.4 敏感操作高亮
- 删除、禁用、配置变更 → 行标红

### 10.5 管理员管理（预留）
- 管理员列表
- 新增/删除管理员（单人阶段一个即可）
- 修改密码

---

## 十一、交互与业务规则

> 本章为 10 个模块的详细 UX 规格，每模块包含 User Story、页面状态矩阵、交互规则、业务规则与边界条件、验收标准。

### 11.1 仪表盘

**User Story**：作为运营人员，我进入管理后台后能一眼看清平台核心指标与待处理事项，无需任何手动操作。

**页面状态矩阵**：

| 状态 | 触发条件 | 展示内容 |
|------|---------|---------|
| Loading 态 | 首次加载，API 未返回 | 8 个指标卡片骨架屏（skeleton），趋势图浅灰占位块 |
| 正常态 | API 成功，数据非空 | 指标卡片渲染实际数值；趋势图折线；待审核预览 5 条 |
| 空态 | 平台无任何数据（新部署） | 卡片显示 `0` 或 `--`；趋势图「暂无数据」插画 |
| 部分数据缺失 | 某些指标后端返回 `null` | 对应卡片显示 `--`（区别于 0） |
| 错误态 | API 请求失败 | 顶部红色 Banner：「数据加载失败，请刷新重试」+ 缓存降级数据 |
| 趋势图无数据 | 7 天内无任何发布 | 空坐标轴 + 居中提示「近 7 天无数据」 |

**交互规则**：
- 页面进入即自动加载，不设手动刷新按钮（数据 T+0 日级即可）
- 待审核卡片点击 → 跳转审核队列并锚定（`?focus=exp_id`）
- 趋势图默认 7 天，提供「7/14/30 天」切换
- 页面聚焦（`visibilitychange`）时静默刷新；切回时重新请求；无轮询
- 数字 ≥10,000 显示为 `1.2万`（保留一位小数）

**业务规则**：
- 仪表盘指标与对应模块列表同一数据源，确保数字对齐
- 待审核口径：`review_status='pending' AND status='published' AND is_private=false`
- 今日统计以服务器时区 `Asia/Shanghai` 00:00 为界
- API 失败时展示 localStorage 缓存（有效期 1h），黄色提示「显示的是缓存数据」

**验收标准**：
1. 8 个指标卡片在 2 秒内加载完成并显示正确数值
2. 趋势图默认 7 天折线，切换「14/30 天」数据正确更新
3. 待审核卡片点击能正确跳转审核页并锚定对应经验
4. API 失败时显示红色 Banner，缓存卡片仍正常展示（不白屏）

### 11.2 用户管理

**User Story**：作为运营人员，我能搜索、查看、禁用/启用用户，对违规用户执行处置并记录理由。

**页面状态矩阵**：

| 状态 | 触发条件 | 展示内容 |
|------|---------|---------|
| Loading 态 | 列表数据请求中 | 5 行骨架屏；搜索框可用但不触发请求 |
| 正常态 | 列表数据返回 | 表格：头像+昵称、Apple ID（脱敏前8后4）、经验数、收藏数、注册时间、状态、操作 |
| 空态 | 搜索无匹配 | 「未找到匹配用户」+「请尝试其他关键词」 |
| 全空态 | 平台无用户 | 「暂无用户」空状态插画 |
| 错误态 | API 失败 | 「加载失败」+ 重试按钮 |

**交互规则**：
- 搜索防抖 300ms；右侧「×」清空搜索恢复全量列表
- 分页每页 20 条
- 禁用操作：二次确认弹窗 + 必填理由（≤200 字）+ Toast「用户已禁用」
- 启用操作：简单二次确认 + Toast「用户已启用」
- 已禁用用户显示「查看理由」按钮，弹窗只读展示
- 批量操作上限 **50 条**，超限按钮置灰 + tooltip「单次最多操作 50 条」
- 排序：注册时间 / 经验数 / 收藏数（点击列头切换升降序）

**业务规则**：
- 管理员自身行「禁用」按钮置灰，tooltip「不可禁用当前账号」
- Apple User ID 前端脱敏：前 8 位 + `****` + 后 4 位
- 禁用理由存储到独立 `user_moderation_log` 表
- 幂等性：重复禁用已禁用用户 → Toast「该用户已被禁用」

**验收标准**：
1. 搜索框输入停止 300ms 后自动搜索，结果正确过滤
2. 禁用用户 → 二次确认 + 理由 → 状态变为禁用 + Toast
3. 已禁用用户「查看理由」可点击，弹窗展示正确理由和时间
4. 选中 51 条 → 批量禁用按钮置灰 + tooltip
5. 管理员自身行禁用按钮置灰

### 11.3 内容审核

**User Story**：作为运营人员，我能高效审核用户发布的公开经验，逐条或批量做出通过/拒绝/退回决策。

**页面状态矩阵**：

| 状态 | 触发条件 | 展示内容 |
|------|---------|---------|
| Loading 态 | 审核队列加载中 | 骨架屏 + 计数器 `--` |
| 正常态 | 队列有数据 | 顶部「待审核：N 条」；列表：作者+时间+领域标签+内容+AI建议折叠+操作按钮 |
| 空态 | 无待审核 | 「🎉 暂无待审核内容」 |
| 错误态 | API 失败 | 「加载审核队列失败」+ 重试按钮 |
| 单条操作后 | 通过/拒绝/退回 | 该项淡出消失（200ms），计数 -1，不重新请求全量（本地移除） |
| 批量操作后 | 批量处理 | 选中项全部消失，全选自动取消 |
| 审核完成 | 最后一条被处理 | 自动切换至空态 |

**交互规则**：
- 排序：始终 `created_at ASC`（最旧优先），固定不提供切换
- 单条通过：即时乐观移除 → PATCH → Toast「已通过」；失败则该项重新出现 + 错误提示
- 单条拒绝：弹出对话框 + 必填理由（≤500 字，预置常见理由下拉）+ 可选建议 → 确认后移除
- 单条退回：必填退回理由 + 提示「退回后用户可修改重新提交」
- 批量通过/拒绝：勾选 N 条 → 确认弹窗 → 全部移除 → Toast
- 全选：表头复选框全选当前页；勾选后底部浮动操作栏「已选 M 条」「批量通过」「批量拒绝」
- AI 审核建议默认折叠，点击展开：AI 判断 + 置信度 + 6 维评分 + 评语摘要
- 无限滚动加载（非传统分页器）

**业务规则**：
- 审核队列仅含 `review_status='pending'` 的公开经验
- 幂等性：重复操作 → 后端 409 → 前端自动移除该项
- 拒绝操作不可逆（用户无法重新提交），弹窗附加黄色警告
- 退回操作可逆（用户可修改重提）
- 拒绝理由后端双重校验非空（前端+后端）
- 所有审核操作写入 `review_log` 表

**验收标准**：
1. 审核队列按发布时间从旧到新排序
2. 单条通过后该项消失 + 计数 -1 + Toast；刷新后不再出现
3. 拒绝不填理由 →「确认拒绝」按钮置灰
4. 批量勾选 3 条通过 → 全部消失 + 全选取消
5. 对已审核经验重复操作 → 409 → 该项自动移除

### 11.4 内容管理

**User Story**：作为运营人员，我能查看、编辑、软删除、硬删除平台所有经验（含已删除），进行全生命周期管理。

**页面状态矩阵**：

| 状态 | 触发条件 | 展示内容 |
|------|---------|---------|
| Loading 态 | 列表加载中 | 表格骨架屏（8 列），筛选器可见但不可交互 |
| 正常态 | 列表返回 | 表格：内容摘要、作者、领域、review_status 彩色 Tag、status、质量分、时间、操作 |
| 空态 | 筛选无匹配 | 「未找到匹配的经验」+「请调整筛选条件」 |
| 全空态 | 平台无经验 | 「暂无经验」空状态 |
| 错误态 | API 失败 | 「加载失败」+ 重试 |
| 已删除行 | status='deleted' | 整行浅红背景 `#fff1f0`，操作列仅「恢复」「硬删除」 |

**交互规则**：
- 默认展示所有 status（含 deleted/hidden/flagged）；筛选器任一条件变更自动刷新
- 编辑：模态弹窗 → 可编辑字段 → 保存（无修改时按钮置灰）→ Toast「已保存」
- 软删除：确认弹窗 → PATCH status='deleted' → 行变浅红 + Toast「已软删除」
- 恢复：简单确认 → PATCH status='published' → 恢复正常 + Toast
- **硬删除**：红色警告 + 「⚠️ 不可恢复」+ **输入「永久删除」四字**才能启用确认按钮
- 排序：发布时间 / 质量分 / 点赞数

**业务规则**：
- 编辑使用乐观锁（`updated_at` 版本检查），被他人修改后返回 409
- 硬删除前置条件：必须先软删除（两步），非 deleted 状态拒绝返回 400
- 运营可手动覆盖 review_status，记录到 review_log 表标记 `source='manual_override'`
- 修改一级领域 → 二级领域自动清空（防数据不一致）

**验收标准**：
1. 筛选 review_status=approved → 仅展示通过经验
2. 编辑内容后保存 → 弹窗关闭 + Toast + 行内容更新
3. 软删除后行变浅红 + 操作变为「恢复」「硬删除」
4. 硬删除必须输入「永久删除」才启用确认按钮
5. 对非 deleted 经验硬删除 → 后端 400 + Toast「请先软删除」

### 11.5 平台内容管理

**User Story**：作为运营人员，我能管理平台官方经验库，支持手动新建、CSV 批量导入、批量 AI 评分。

**页面状态矩阵**：

| 状态 | 触发条件 | 展示内容 |
|------|---------|---------|
| Loading 态 | 列表加载中 | 表格骨架屏（7 列），新建/导入按钮不可点击 |
| 正常态 | 列表返回 | 表格：内容摘要、领域、来源标注、质量分、review_status、时间、操作 |
| 空态 | 无平台经验 | 「暂无平台经验」+ CTA |
| 错误态 | API 失败 | 「加载失败」+ 重试 |
| CSV 预览中 | 文件上传后 | 模态弹窗：前 10 行预览 + 列映射 + 错误行标记 + 总行数 |
| 批量 AI 进行中 | 批量评分 | 进度弹窗「正在评分 M/N」+ 成功失败列表 + 可中途取消 |

**交互规则**：
- 新建：表单（内容必填 ≤100 字、领域必填）→ POST（自动设置 `review_status='approved'`, `is_official=true`）→ 列表顶部插入新高亮 2 秒 + Toast
- CSV 导入流程：上传 → 后端解析 + 校验 → 预览弹窗 → 确认列映射 → 确认导入 → Toast
- CSV 格式：必需列 content/domain/sub_domain；可选列 interpretation/source_label
- 批量 AI 评分：勾选 N 条 → 确认 → 进度弹窗（逐条处理，失败不影响其他）→ 完成后列表刷新
- 单条重新评分：确认 → 调用 AI → 质量分更新 + Toast
- 下架/上架：PATCH status → Toast

**业务规则**：
- 新建平台经验直接 `review_status='approved'`，不经过审核（运营特权）
- CSV 去重：`content + author_id` 联合检查，重复跳过
- CSV 上限：单次 500 条（≈50KB），超限前端拒绝
- CSV 编码：UTF-8 / UTF-8 BOM 自动检测
- 批量 AI 上限：单次 50 条
- 重新评分覆盖原分数，记录到 review_log（`source='manual_rescore'`）

**验收标准**：
1. 新建平台经验 → 列表顶部新高亮 2 秒 + review_status=approved + Toast
2. 上传合法 CSV（20 条）→ 预览前 10 行 → 确认导入 → Toast「成功 20 条」
3. CSV 含 3 条空 content → 预览红色标记 + 错误提示
4. 批量 AI 评分 10 条 → 进度弹窗实时更新 → 完成后质量分全更新
5. 上传 600 行 CSV → 前端拒绝 + 提示

### 11.6 领域与标签管理

**User Story**：作为运营人员，我能灵活管理 App 的经验领域与子领域（CRUD/排序/启用禁用），确保内容分类与业务一致。

**页面状态矩阵**：

| 状态 | 触发条件 | 展示内容 |
|------|---------|---------|
| Loading 态 | 首次进入/手动刷新 | 骨架屏（5 行占位，含缩进），操作按钮置灰 |
| 空态 | 领域列表为空 | 「暂无领域分类」+ 新建 CTA |
| 错误态 | 接口超时/5xx | Toast 错误 + 保留上次数据 + 顶部「加载失败，重试」横幅 |
| 禁用冲突 | 禁用有活跃经验的领域 | 弹窗列出受影响经验（前 5 条 +「等 N 条」），提示影响范围 |
| 拖拽保存失败 | 排序接口异常 | Toast「排序保存失败，已恢复原顺序」，列表回滚 |

**交互规则**：
- 删除领域：下属有子领域或经验时按钮置灰 + tooltip「请先移除下属子领域及经验」
- 拖拽排序：行首 6 点手柄，拖拽中半透明 + 蓝色虚线，松手后保存
- CRUD 成功 Toast（绿色 3 秒），失败 Toast（红色 5 秒）
- 启用/禁用：Switch 即时 UI 更新 + Toast
- 页面聚焦时静默刷新

**业务规则**：
- 排序序号必须连续（1,2,3...），后端整体重排
- 子领域名称同父领域下不可重复（409）
- 禁用父领域 → 子领域联动禁用
- 删除领域不可逆，需输入领域名称确认

**验收标准**：
1. 新建领域 → 列表即时出现 + sort_order 自动追加末尾
2. 拖拽从 3→1 → 所有领域 sort_order 连续
3. 禁用有经验归属的领域 → 弹窗影响范围 → 确认后生效
4. 删除领域需输入名称确认，不匹配按钮置灰
5. 网络断开编辑 → Toast 错误 + 回退原值

### 11.7 AI 服务管理

**User Story**：作为运营人员，我能实时掌握 AI 服务状态、调用量及预估费用，及时发现异常并控制成本。

**页面状态矩阵**：

| 状态 | 触发条件 | 展示内容 |
|------|---------|---------|
| Loading 态 | 首次进入 | 状态指示灯脉冲、统计卡片 `--`、费用骨架屏 |
| 空态 | 所选时间无调用 | 统计显示 0，折线图零值坐标轴 +「无调用记录」 |
| 错误态（健康超时） | AI 服务不响应 | 🔴 指示灯 +「服务不可达」横幅 + 最近成功探针时间 |
| 错误态（统计异常） | 统计接口 5xx | 卡片半透明遮罩 +「加载失败，重试」 |
| 费用异常 | 当日费用 > 昨日 200% | 费用标红 + 橙色 warning 标签 + hover 对比值 |

**交互规则**：
- 每 30 秒自动轮询健康检查（页面不可见时暂停，切回立即检查）
- 时间范围切换：折线图/柱状图 300ms 平滑过渡
- 费用 hover Tooltip：「费用为近似估算值，非精确账单」
- Prompt 模板只读模式（灰度背景，无编辑入口）
- 默认时间范围「最近 7 天」

**业务规则**：
- 健康状态：🟢 响应 <2s → 🟡 连续 2 次 >2s → 🔴 连续 3 次超时/5xx → ⚪ 首次加载中
- 费用公式：Token 数 × 单价（单价从系统配置读取）
- 费用卡片底部固定注释「费用为近似估算值」
- AI 统计数据保留 180 天

**验收标准**：
1. 加载后 3 秒内完成健康检查并显示状态指示灯
2. 健康检查每 30 秒轮询；切标签页暂停，切回立即恢复
3. 切换时间范围 → 图表 300ms 内完成过渡
4. 费用卡片 hover 显示澄清 Tooltip
5. 健康状态绿→红 → 指示灯平滑过渡 + 顶部告警横幅

### 11.8 数据统计

**User Story**：作为运营人员，我能查看多维度数据统计图表和用户留存率，用数据驱动运营决策。

**页面状态矩阵**：

| 状态 | 触发条件 | 展示内容 |
|------|---------|---------|
| Loading 态 | 首次进入/切换时间 | 图表骨架屏、数值卡片 `--` |
| 空态 | 所选时间全零数据 | 坐标轴 + 零值折线 +「所选时间段暂无数据」 |
| 留存率空态 | 未接入埋点 | 「需埋点支持」占位 + hover 说明 |
| 错误态 | 接口 5xx/超时 | 对应图表错误插画 + 重试；其余图表独立不阻塞 |
| 日期超限 | 跨度 > 90 天 | 前端截断 + 橙色提示「已自动截断」 |

**交互规则**：
- 默认「最近 7 天」+ 第一个统计维度 Tab
- 切换时间范围：图表 300ms 平滑过渡，各图表并行加载
- 切换维度 Tab：仅加载激活 Tab，已加载 Tab 缓存不重复请求
- 图表支持 hover Tooltip + 滚轮缩放 + 拖拽平移 + 图例点击显隐
- 数据仅用户主动操作时加载，无自动轮询

**业务规则**：
- 7 维统计：用户增长/内容增长/互动/领域分布/审核统计/AI 使用/留存
- 留存率依赖埋点，返回 null 时展示「需埋点支持」不显示假零值
- 全部统计服务端统一口径，前端不做二次计算
- 时区基于 UTC+8（Asia/Shanghai）
- 同时间范围+维度组合页面生命周期内缓存

**验收标准**：
1. 默认 7 天数据，7 个维度 Tab 正常切换且图表独立渲染
2. 切换「30 天」→ 已激活 Tab 图表 1 秒内完成更新
3. 留存率无埋点数据 → 显示「需埋点支持」，不影响其他维度
4. 自定义日期 >90 天 → 自动截断 + 提示
5. 任一维度接口异常 → 仅该图表错误态 + 重试，其余正常

### 11.9 系统配置

**User Story**：作为运营人员，我能集中管理系统配置和敏感词库，灵活调整 App 运行参数而无需开发介入。

**页面状态矩阵**：

| 状态 | 触发条件 | 展示内容 |
|------|---------|---------|
| Loading 态 | 首次进入 | 配置列表骨架屏（5 行 KV 占位），保存按钮置灰 |
| 空态（配置） | 配置列表为空 | 「暂无配置项」+「联系开发人员」 |
| 空态（敏感词） | 未添加敏感词 | 「暂无敏感词」+ CTA |
| 错误态 | 保存失败 | Toast 错误 + 编辑内容保留不丢失 |
| 核心配置修改 | 修改审核模式/注册开关等 | 二次确认弹窗：「此配置影响全局业务，确认修改？」+ 变更前后对比 |

**交互规则**：
- 单项保存：值未变更时按钮置灰
- 普通配置保存：Toast「保存成功」+ 即时退出编辑
- 核心配置保存：二次确认 → Toast + 1 秒高亮闪烁
- 敏感词：单个添加（输入框+按钮）、批量导入（CSV 每行一词）、表格搜索过滤、多选批量删除
- 配置修改即时生效，无需重启

**业务规则**：
- 核心配置标记：审核模式 / 注册开关 / AI 频次上限，修改时强制二次确认
- Key 系统预设不可新增删除，仅可改 Value
- 值类型校验：String/Boolean/Number/JSON，非法值输入框红框 + 错误提示
- 敏感词去重：重复词提示「N 条已存在，已自动跳过」
- 配置变更自动写入操作日志（记录前后值）

**验收标准**：
1. 修改普通配置 → Toast 成功 + 后端即时生效 + 刷新持久化
2. 修改「审核模式」→ 二次确认弹窗 + 变更对比 + 确认后 Toast
3. 敏感词输入空白 →「添加」按钮置灰
4. CSV 导入 100 条含 10 条重复 → 提示跳过 + 新增 90 条
5. Number 类型输入非数字 → 红框 + 行内错误提示

### 11.10 操作日志

**User Story**：作为运营人员，我能多维筛选查看操作日志，追溯问题操作和审计敏感行为。

**页面状态矩阵**：

| 状态 | 触发条件 | 展示内容 |
|------|---------|---------|
| Loading 态 | 首次进入/切换筛选/翻页 | 表格骨架屏（10 行），筛选器可用但表格有 loading 遮罩 |
| 空态 | 筛选无匹配 | 空插画 +「当前筛选条件下无操作日志」+ 清空筛选链接 |
| 全空态 | 无任何日志 | 「暂无操作记录」插画 |
| 错误态 | 接口超时/5xx | 错误提示 + 重试；筛选器保持可用 |
| 翻页越界 | 页码 > 总页数 | 自动回退最后一页 + Toast |

**交互规则**：
- 默认按操作时间倒序（最新在上）
- 筛选变更即时触发查询（操作人/类型/时间范围/关键词）
- 敏感操作行：红色浅底高亮（`#FFF1F0`）+ 行首「敏感」Badge
- 点击日志行 → 弹窗展示详情（JSON Diff 只读）
- 导出日志：筛选结果 >10,000 条时拦截提示
- 不自动轮询；顶部「数据截止至 HH:mm:ss」可点击刷新
- 分页默认 20 条/页，支持切换 10/20/50

**业务规则**：
- 敏感操作：删除领域/经验、修改核心配置、批量删除敏感词、用户封禁
- 日志 90 天自动清理（定时任务每日凌晨 2:00）
- 操作日志不可手动删除（前端无删除入口）
- 导出上限 10,000 条

**验收标准**：
1. 日志列表时间倒序 + 每页 20 条 + 翻页正常
2. 筛选「删除领域」→ 仅匹配日志 + 敏感行红底 + Badge
3. 点击日志行 → 详情弹窗 JSON Diff 展示变更前后对比
4. 关键词搜索 300ms 防抖；清空恢复原结果
5. 导出 12,000 条 → 拦截提示「超过上限，请缩小范围」

---

## 十二、非功能性需求

### 12.1 性能指标

| 指标 | 目标值 | 测量方式 |
|------|--------|---------|
| 首屏加载（仪表盘） | < 2 秒 | Lighthouse / Chrome DevTools |
| 列表分页查询 | < 500ms | API 响应时间 P95 |
| 单条审核操作 | < 300ms | 点击到 Toast 反馈 |
| 批量操作（50 条） | < 3 秒 | 全流程耗时 |
| CSV 导入（500 条） | < 10 秒 | 上传→预览→确认→导入完成 |

### 12.2 安全规范

- **密码强度**：管理员密码 ≥8 位，含字母+数字
- **会话超时**：JWT 7 天有效，Refresh Token 30 天
- **操作日志不可篡改**：`admin_logs` 表只 INSERT 不 UPDATE/DELETE
- **敏感信息脱敏**：Apple User ID 前端仅显示前 8 后 4 位
- **CSRF 防护**：JWT Bearer Token 天然防 CSRF
- **SQL 注入防护**：全部使用 pgx 参数化查询（`$1, $2, ...`）

### 12.3 浏览器兼容

- **唯一支持**：Chrome 桌面端最新两个大版本（当前 v130+）
- 不要求 Safari / Edge / Firefox 兼容
- 不要求移动端浏览器适配
- 最低分辨率：1366×768

### 12.4 并发与容错

- **多 Tab 支持**：同一管理员可开多个 Tab，操作互不冲突（乐观锁防编辑覆盖）
- **API 超时**：前端 fetch 默认超时 15 秒，超时显示错误态 + 重试按钮
- **网络断开**：操作失败保留用户输入不丢失，恢复后提示重试
- **后端重启**：前端自动重新认证（Token 检测 401 → 跳转登录页）
- **数据库故障**：后端返回 500，前端显示降级提示

---

## 十三、数据埋点方案

管理后台自身需要采集的数据（用于优化运营效率）：

| 埋点事件 | 触发时机 | 采集字段 |
|---------|---------|---------|
| `admin_login` | 管理员登录成功 | admin_id, timestamp, ip |
| `admin_page_view` | 进入任一页面 | page_name, timestamp, admin_id |
| `admin_review_action` | 审核操作（通过/拒绝/退回） | action_type, exp_id, review_duration_ms, ai_verdict |
| `admin_batch_action` | 批量操作 | action_type, affected_count, admin_id |
| `admin_user_disable` | 禁用用户 | user_id, reason, admin_id |
| `admin_config_change` | 修改系统配置 | config_key, old_value, new_value, admin_id |
| `admin_csv_import` | CSV 导入 | total_rows, success_count, fail_count, duration_ms |
| `admin_ai_batch_score` | 批量 AI 评分 | total, success, fail, duration_ms |
| `admin_export` | 导出 CSV | export_type, row_count, admin_id |

**目的**：追踪运营效率（审核耗时、操作频次）、发现瓶颈（高频操作/页面）、审计关键操作。

---

## 十四、开发优先级矩阵

| 优先级 | 模块 | 功能点 | 依赖 | 预估工时 |
|--------|------|--------|------|---------|
| **P0** | 登录认证 | 用户名+密码登录、JWT 签发 | 数据库 | ✅ 已完成 |
| **P0** | 仪表盘 | 8 指标卡片、趋势图、审核预览 | admin_dashboard API | 基本完成 |
| **P0** | 内容审核 | 审核队列、单条/批量通过/拒绝/退回 | admin_review API | 🔴 待重做 |
| **P0** | 内容管理 | 经验列表、编辑、软删/硬删/恢复 | admin_content API | 基本完成 |
| **P1** | 用户管理 | 用户列表、搜索、禁用/启用、详情 | admin_users API | 基本完成 |
| **P1** | 平台内容 | 新建、CSV 导入、批量 AI 评分 | admin_platform API | 较好 |
| **P1** | 系统配置 | 配置 KV 编辑、敏感词管理 | admin_config API | 待完善 |
| **P2** | 领域管理 | 领域 CRUD、排序、启用禁用 | admin_domains API | 待完善 |
| **P2** | 数据统计 | 7 维图表、留存率 | admin_stats API | 待完善 |
| **P2** | AI 服务 | 健康检查、调用统计、费用估算 | admin_ai API | 待完善 |
| **P2** | 操作日志 | 多维度筛选、敏感高亮、导出 | admin_logs API | 待完善 |
| **P3** | CSV 导出 | 用户/经验列表导出 | admin_export API | 基本完成 |

> **当前最大缺口**：审核队列页面（ReviewQueue.tsx 空文件）。这是 P0 运营核心功能，需要立即补齐。

---

## 附 A：API 接口概览

| 模块 | 端点前缀 | 新增端点 |
|------|---------|---------|
| 仪表盘 | `GET /api/v1/admin/dashboard` | 1 个聚合端点返回所有卡片数据 |
| 用户管理 | `GET /api/v1/admin/users` ... | CRUD + 禁用/启用 |
| 内容审核 | `GET /api/v1/admin/reviews` ... | 队列列表 + 审核操作 |
| 内容管理 | `GET /api/v1/admin/experiences` ... | CRUD + 批量操作 |
| 平台内容 | `GET /api/v1/admin/platform-experiences` ... | CRUD + 批量导入/AI |
| 领域管理 | `GET /api/v1/admin/domains` ... | CRUD + 排序 |
| AI 服务 | `GET /api/v1/admin/ai-status` ... | 状态+统计 |
| 数据统计 | `GET /api/v1/admin/stats/*` ... | 各维度统计端点 |
| 系统配置 | `GET /api/v1/admin/config` ... | 读/写配置 |
| 操作日志 | `GET /api/v1/admin/logs` ... | 列表查询 |

---

## 附 B：设计稿

设计稿 HTML 文件：`designs/admin-v1.html`
展示页面：总览仪表盘 + 审核队列 + 详情弹窗

---

# 详细设计附录（基于真实代码实现）

> 以下内容基于 2026-05-07 前后端代码审计，与 `backend/internal/handler/admin_*.go` 完全对齐。

---

## 附 C：认证体系

### C.1 中间件链路

```
HTTP Request → CORS() → AuthMiddleware(jwtSecret, db) → [业务路由]
```

**AuthMiddleware**（`middleware/auth.go`）:
- 解析 `Authorization: Bearer <token>`，验证 HS256 JWT 签名
- 检查 token 是否被吊销（`token_revocations` 表）
- 检查用户是否被禁用（`users.deleted_at IS NOT NULL`）
- 注入 `user_id`, `open_id`, `nickname`, `jti` 到 Gin Context
- **无 token 时不拦截**（公开接口可访问）

**RequireAdmin**（`middleware/admin.go`）:
- 查询 `users.is_admin` 字段确认管理员身份
- `user_id` 不存在 → `401 {"error": "请先登录"}`
- `is_admin = false` → `403 {"error": "需要管理员权限"}`
- DB 为 nil（测试模式）→ 允许通过
- **加在 admin router group 上**，所有 `/api/v1/admin/*` 路由都需要

### C.2 JWT 体系

| 项目 | 值 |
|------|-----|
| 签名算法 | HS256 |
| 有效期 | 7 天 |
| Claims | `user_id`, `open_id`, `nickname`, `is_admin`, `jti` |
| Refresh Token | 随机 32 字节 hex，SHA256 哈希存储，30 天有效 |
| 吊销 | `token_revocations` 表（jti 黑名单），7 天自动清理 |

### C.3 管理员登录

**`POST /api/v1/auth/admin/login`**（`handler/admin_auth.go`）

```json
// Request
{ "username": "<admin-user>", "password": "<从生产环境安全配置获取，不写入仓库文档>" }

// Response 200
{
  "token": "<jwt>",
  "refresh_token": "<refresh-token>",
  "user": {
    "id": "uuid",
    "nickname": "admin",
    "is_admin": true
  }
}
```

后端流程：`users` 表按 nickname 查找 → bcrypt 验证 → 检查 `is_admin=true` → 签发 JWT + Refresh Token。

**开发模式**（`POST /api/v1/auth/admin/dev/login`，仅 `ENV != production` 时注册路由）：
```json
// Request
{ "nickname": "dev-admin" }
```

使用 `INSERT ... ON CONFLICT (apple_user_id) DO UPDATE` 自动创建 `dev-admin-{nickname}` 用户并设置 `is_admin=TRUE`。

---

## 附 D：数据模型

### D.1 AdminDashboard（仪表盘聚合）

```go
type AdminDashboard struct {
    TotalUsers        int  `json:"total_users"`
    TotalExperiences  int  `json:"total_experiences"`
    TodayNewUsers     int  `json:"today_new_users"`
    TodayNewExps      int  `json:"today_new_exps"`
    TodayActiveUsers  int  `json:"today_active_users"`
    TodayAIChats      int  `json:"today_ai_chats"`
    PendingReviews    int  `json:"pending_reviews"`
    TodayApproved     int  `json:"today_approved"`
    TodayRejected     int  `json:"today_rejected"`
    YesterdayNewUsers int  `json:"yesterday_new_users"`     // 用于 ↑/↓ 对比
    YesterdayNewExps  int  `json:"yesterday_new_exps"`
    ReviewPreview     []ReviewItem `json:"review_preview"`   // 最近5条待审
}
```

### D.2 ReviewItem（审核列表项）

```go
type ReviewItem struct {
    ID              string          `json:"id"`
    Content         string          `json:"content"`
    Domain          string          `json:"domain"`
    SubDomain       *string         `json:"sub_domain"`
    SourceType      string          `json:"source_type"`
    ReviewStatus    string          `json:"review_status"`
    AIVerdict       *string         `json:"ai_verdict"`       // "approved" / "rejected"
    AIScore         *float64        `json:"ai_score"`          // 0-10
    AIScoreDetail   json.RawMessage `json:"ai_score_detail"`   // 6维度明细JSONB
    AIInterpretation *string        `json:"ai_interpretation"`
    HardPolicyResult json.RawMessage `json:"hard_policy_result"`
    AuthorName      string          `json:"author_name"`
    SubmittedAt     string          `json:"submitted_at"`
}
```

**AI判定标签规则**：
- 硬策略拒绝：`hard_policy_result.passed=false` → 红色 **"硬拒绝"**
- AI 拒绝：`ai_verdict="rejected"` → 橙色
- AI 通过：`ai_verdict="approved"` → 绿色 + 显示分数

### D.3 AdminUserItem / AdminUserDetail

```go
type AdminUserItem struct {
    ID            string  `json:"id"`
    Nickname      string  `json:"nickname"`
    AvatarURL     *string `json:"avatar_url"`
    Title         *string `json:"title"`
    AuthProvider  string  `json:"auth_provider"`  // "apple" / "dev"
    IsActive      bool    `json:"is_active"`      // deleted_at IS NULL
    CreatedAt     string  `json:"created_at"`
    ExpCount      int     `json:"exp_count"`
    BookmarkCount int     `json:"bookmark_count"`
}

type AdminUserDetail struct {
    AdminUserItem
    LikeReceived     int     `json:"like_received"`
    BookmarkReceived int     `json:"bookmark_received"`
    ViewedCount      int     `json:"viewed_count"`
    LikedCount       int     `json:"liked_count"`
    BookmarkedCount  int     `json:"bookmarked_count"`
    ChatCount        int     `json:"chat_count"`
    MsgCount         int     `json:"msg_count"`
    DomainPrefs      []DomainPref `json:"domain_prefs"`  // 发布×2+收藏×1
}
```

### D.4 通用分页

```go
type PaginatedResponse struct {
    Data       interface{} `json:"data"`
    Page       int         `json:"page"`
    PageSize   int         `json:"page_size"`
    TotalCount int         `json:"total_count"`
    TotalPages int         `json:"total_pages"`
}
```

### D.5 趋势数据（用户/经验/互动/AI/审核）

```go
type TrendPoint struct {
    Date  string `json:"date"`
    Count int    `json:"count"`
}
```

---

## 附 E：完整 API 契约

### E.1 仪表盘

| 方法 | 路径 | 查询参数 | 响应 |
|------|------|---------|------|
| GET | `/api/v1/admin/dashboard` | — | `AdminDashboard` |
| GET | `/api/v1/admin/trends` | `days`（默认7, 最大90） | `{days, users: TrendPoint[], experiences: TrendPoint[]}` |

**关键 SQL 模式**：
- 总量过滤：`WHERE deleted_at IS NULL`
- 今日活跃：`COUNT(DISTINCT user_id) FROM user_views WHERE viewed_at >= CURRENT_DATE`
- 今日审核通过：`COUNT(*) FROM admin_logs WHERE action_type='review_approve' AND created_at >= CURRENT_DATE`（**不用** `experiences.updated_at`——编辑也会更新）
- 趋势：`generate_series(CURRENT_DATE - (N-1), CURRENT_DATE, '1 day') LEFT JOIN ...`

### E.2 审核管理

**`GET /api/v1/admin/reviews`**

查询参数：
| 参数 | 类型 | 默认 | 说明 |
|------|------|------|------|
| `review_status` | string | `"pending"` | approved/rejected/pending/private |
| `domain` | string | — | 一级领域过滤 |
| `source_type` | string | — | platform/user |
| `ai_verdict` | string | — | approved/rejected（AI 判定结果过滤） |
| `date_from` | string | — | YYYY-MM-DD |
| `date_to` | string | — | YYYY-MM-DD |
| `page` | int | 1 | |
| `page_size` | int | 20 | |

**`POST /api/v1/admin/reviews/:id/approve`** — 通过审核
```json
// Response 200
{ "status": "ok" }
```
后端：`UPDATE experiences SET review_status='approved', updated_at=NOW() WHERE id=$1` + 写入 `admin_logs`

**`POST /api/v1/admin/reviews/:id/reject`** — 驳回审核
```json
// Request
{ "reason": "内容不符合社区规范" }
// Response 200
{ "status": "ok" }
```

**`POST /api/v1/admin/reviews/:id/retry`** — 退回重审
重置 `review_status='pending'`，清除旧 AI 结果。记录 `action_type='review_retry'`。

**`POST /api/v1/admin/reviews/:id/misjudge`** — 标记误判
```json
// Request
{ "reason": "误拒绝：内容实际是可用经验" }
```
重置为 pending，记录 `action_type='review_misjudge'`。

**`POST /api/v1/admin/reviews/batch`** — 批量审核
```json
// Request
{ "ids": ["uuid1", "uuid2", ...], "action": "approve", "reason": "批量通过" }
// Response 200
{ "status": "ok", "action": "approve", "affected_count": 5 }
```

### E.3 内容管理

**`GET /api/v1/admin/experiences`**

查询参数：`domain`, `source_type`, `review_status`, `search`（ILIKE 匹配 content/creator_name/nickname）, `page`, `page_size`

不筛 `status`——管理员可看所有状态（published/hidden/flagged 及已删除）。

**`PUT /api/v1/admin/experiences/:id`** — 编辑
```json
// Request（所有字段可选）
{
  "content": "修改后的经验内容",
  "domain": "career",
  "sub_domain": "skill-building",
  "source_type": "platform",
  "creator_name": "张小龙",
  "score_reason": "深度思考的体现"
}
```

**`POST /api/v1/admin/experiences/:id/restore`** — 恢复软删除
清除 `deleted_at`，不恢复 `status`（如果之前 status=hidden 仍保持 hidden）。

**`POST /api/v1/admin/experiences/:id/hard-delete`** — 物理删除
**仅限** `review_status != 'approved'` 的经验。已入池经验拒绝硬删除。

**`PUT /api/v1/admin/experiences/:id/review-status`** — 直接改审核状态
```json
// Request
{ "review_status": "approved", "reason": "手动修正" }
```

### E.4 平台经验管理

**`POST /api/v1/admin/platform-experiences`** — 创建
```json
// Request
{
  "content": "好的产品用完即走",
  "domain": "career",
  "sub_domain": "career-planning",
  "creator_name": "张小龙",
  "source_label": "张小龙·微信产品观",
  "score_reason": "经典产品哲学"
}
// Response 201
{ "id": "uuid", "status": "created" }
```
自动设置：`author_id=00000000-0000-0000-0000-000000000000`, `source_type='platform'`, `is_official=true`, `review_status='approved'`

**`POST /api/v1/admin/platform-experiences/batch-ai`** — 批量 AI 评分
```json
// Request
{ "ids": ["uuid1", "uuid2", ...] }
// Response
{ "total": 10, "success": 8, "failed": 2 }
```
异步调用 AI Service 的 `/api/v1/review` + `/api/v1/chat/generate-interpretation`，并发量 5。

**`POST /api/v1/admin/platform-experiences/:id/rescore`** — 单条重打分
重新调用 AI 打分覆盖原分数。

**`POST /api/v1/admin/platform-experiences/import-csv`** — CSV 导入
```json
// Request
{
  "data": "content,domain,sub_domain,creator_name,source_label,score_reason\n内容1,career,career-planning,作者1,来源1,理由1\n..."
}
// Response
{ "total": 50, "inserted": 45 }
```
CSV 表头固定：`content, domain, sub_domain, creator_name, source_label, score_reason`。跳过领域无效的行。

### E.5 用户管理

**`GET /api/v1/admin/users`**

查询参数：`search`（昵称模糊）, `auth_provider`（apple/dev）, `is_active`（true/false）, `date_from`, `date_to`, `sort`（activity=按经验数降序，默认按注册时间）, `has_title`（有无称号）, `page`, `page_size`

**`PUT /api/v1/admin/users/:id/status`**
```json
// Request
{ "active": false, "reason": "违规发布广告" }
// Response
{ "status": "ok", "active": false }
```
启用 → `SET deleted_at = NULL`；禁用 → `SET deleted_at = NOW()`。操作记录到 `admin_logs`（action_type=`"启用"`/`"禁用"`，中文）。

**`POST /api/v1/admin/users/batch-status`**
```json
// Request
{ "ids": ["uuid1", ...], "active": false, "reason": "批量清理" }
```

### E.6 领域管理

领域元数据存储在 `system_config` 表（非独立表），key 格式：
- `domain_display_{name}` — 显示名
- `domain_active_{name}` — 启用/禁用
- `domain_order_{parent}` — 排序（JSON 数组）
- `sub_domains_{parent}` — 子领域定义（JSON）

**`GET /api/v1/admin/domains`** 返回层级树（硬编码基础结构 + system_config 覆盖）。

**`PUT /api/v1/admin/domains/reorder`**
```json
// Request（一级领域排序）
{ "names": ["career", "cognition", "life", "relationship", "emotion"] }
// Request（二级领域排序）
{ "parent_name": "career", "names": ["career-planning", "skill-building", "side-hustle", "workplace-comm"] }
```

### E.7 数据统计

| 端点 | 参数 | 响应 |
|------|------|------|
| `GET /admin/stats/users` | `days` | `{days, data: TrendPoint[]}` |
| `GET /admin/stats/experiences` | `days`, `source_type` | `{days, source_type, data: TrendPoint[]}` |
| `GET /admin/stats/interactions` | `days` | `{days, likes[], bookmarks[], views[]}` |
| `GET /admin/stats/reviews` | `days` | `{days, approved[], rejected[]}` |
| `GET /admin/stats/domains` | — | `{data: [{domain, count}]}` |
| `GET /admin/stats/ai` | `days` | `{days, chats[], messages[], interpretations[]}` |
| `GET /admin/stats/retention` | `days` | `{days, day1[], day7[], day30[]}` |

**统一 SQL 模式**：`generate_series(CURRENT_DATE - (N-1), CURRENT_DATE, '1 day') LEFT JOIN` 目标表，无数据日期返回 0。

**留存率算法**：每日注册用户中，N 天内有过任何活动（浏览/点赞/收藏/对话）的用户占比。

### E.8 AI 服务状态

**`GET /api/v1/admin/ai-status`**

HTTP GET `AI_SERVICE_URL/health`（默认 `http://localhost:8000`），5 秒超时。返回：
- `healthy: bool` — 连通性
- `tier_stats`: Review/Chat/Interpretation 的今日/总量
- `daily_cost` — 今日/本月估算费用（review=$0.002/条, chat=$0.001/条）
- `prompt_config`: Review Prompt 和 Chat System Prompt 的字符长度（只读）
- `batch_tasks`: 最近 5 条批量任务记录

### E.9 配置管理

**配置存储**：`system_config` 表，key-value (JSONB) 结构。

**`PUT /api/v1/admin/config`**
```json
// Request
{ "key": "review_mode", "value": "auto" }
// Response
{ "status": "ok" }
```

**默认配置项**（迁移 `011_admin_system.sql` 预设）：
```
review_mode: "human_review"
content_min_length: 10          content_max_length: 100
interpretation_max_length: 300  title_max_length: 15
publish_limit_per_day: 20       chat_limit_per_day: 50
sensitive_words_enabled: true
registration_enabled: true
ai_interpretation_enabled: true
search_enabled: true
```

**敏感词**：独立表 `sensitive_words(id SERIAL PRIMARY KEY, word VARCHAR UNIQUE NOT NULL)`。支持 GET（列表）, POST（新增）, DELETE（按 ID 删除）。批量导入通过逐条调用 POST。

### E.10 操作日志

**`GET /api/v1/admin/logs`**

查询参数：`admin_id`, `action_type`, `date_from`, `date_to`, `is_sensitive`（bool，过滤敏感操作）, `page`, `page_size`

```json
// Response item
{
  "id": "uuid",
  "admin_id": "uuid",
  "admin_name": "admin",
  "action_type": "review_approve",
  "target_type": "experience",
  "target_id": "uuid",
  "detail": { "reason": "内容优质" },
  "result": "success",
  "created_at": "2026-05-07T12:00:00Z"
}
```

**已记录的操作类型**：
| action_type | 触发场景 |
|-------------|---------|
| `review_approve` / `review_reject` | 单条审核 |
| `review_retry` / `review_misjudge` | 退回重审/标记误判 |
| `review_batch_approve` / `review_batch_reject` | 批量审核 |
| `"启用"` / `"禁用"` | 用户状态变更（中文） |
| `batch_enable` / `batch_disable` | 批量用户操作 |
| `config_update` | 系统配置修改 |
| `sensitive_word_add` / `sensitive_word_delete` | 敏感词管理 |

**敏感操作高亮条件**：`user_disable`, `user_enable`, `batch_disable`, `batch_enable`, `config_update`, `review_approve`, `review_reject`, `hard_delete`

### E.11 CSV 导出

| 端点 | 响应类型 | 导出字段 |
|------|---------|---------|
| `GET /admin/export/users` | `text/csv` (UTF-8 BOM) | ID, 昵称, 注册时间, 是否活跃, 是否管理员, 经验数 |
| `GET /admin/export/experiences` | `text/csv` (UTF-8 BOM) | ID, 内容, 作者, 领域, 子领域, 来源, 审核状态, 质量分, 创建时间 |

UTF-8 BOM（`\xEF\xBB\xBF`）确保 Excel 正确识别中文编码。

---

## 附 F：移动端功能对照表（管理后台管理的对象）

| 移动端功能 | 管理后台对应模块 | 控制点 |
|-----------|---------------|-------|
| Apple Sign In 登录 | 用户管理 | 查看用户→禁用/启用→批量操作 |
| 首页推荐流 | 内容管理 + 审核 | 审核通过才入池；下架/删除会影响推荐 |
| 发布经验 | 内容审核 | 硬策略+AI 双层审核；配置可调硬策略开关 |
| 3D 翻转卡片（AI 解读） | 内容管理 + AI 服务 | AI 解读开关控制生成；内容编辑可修改解读 |
| 价值度星级 | 内容管理 + 平台内容 | quality_score + score_reason 可编辑/重打分 |
| 我的经验 / 收藏 | 用户管理 | 用户详情页查看全部经验列表 |
| AI 对话 | AI 服务 + 系统配置 | chat_limit_per_day 控制上限；AI 使用统计观察趋势 |
| 搜索 | 系统配置 | search_enabled 功能开关 |
| 领域标签体系 | 领域管理 | 5 大领域 + 21 子领域可编辑/禁用/排序 |
| 个人资料编辑 | 用户管理 | 用户详情可查看；title/nickname 字段 |
| 删除经验 | 内容管理 | 软删除=设 deleted_at；硬删除=物理删除 |
| 敏感词过滤 | 系统配置 | 敏感词列表增删 + 开关 |
| 平台「官」标签 | 平台内容管理 | source_type='platform' 的内容通过平台内容 CRUD 管理 |
| 名人经验种子 | 平台内容管理 | CSV 导入；batch-ai 批量评分 |

---

## 附 G：前端页面实现规格

### G.1 页面路由与组件树

```
/admin/login          → Login.tsx（用户名+密码）
/admin/                → Layout.tsx（侧边栏+顶栏+Outlet）
  /admin/dashboard      → Dashboard.tsx
  /admin/users          → UserManagement.tsx
  /admin/reviews        → ReviewQueue.tsx
  /admin/content        → ContentManagement.tsx
  /admin/platform       → PlatformContent.tsx
  /admin/domains        → DomainManagement.tsx
  /admin/stats          → Statistics.tsx
  /admin/config         → SystemConfig.tsx
  /admin/logs           → AdminLogs.tsx
```

**技术栈**：React + TypeScript + Vite + React Router v6 + CSS Modules

**部署**：SPA 子路径 `/admin/`，Nginx `root /var/www; try_files $uri /admin/index.html;`

### G.2 页面通用模式

所有列表页遵循统一模式（`admin-page-pattern`）：
```
┌─ Toolbar ──────────────────────────────────┐
│ [搜索框] [筛选下拉] [筛选下拉] ... [操作按钮] │
├─ Table ────────────────────────────────────┤
│ ☐ | 字段1 | 字段2 | 字段3 | ... | 操作      │
│ ☐ | ...                                    │
├─ Pagination ───────────────────────────────┤
│ < 1 2 3 ... >  共 N 条                      │
└────────────────────────────────────────────┘
```

**Toolbar 模式**：搜索框（防抖 300ms）+ 筛选下拉（级联：一级领域→二级领域）+ 操作按钮（新建/导出/批量操作）。

**Table 模式**：固定表头、斑马纹行、高亮标记（红色=拒绝、橙色=AI拒绝）、复选框列。

**Modal 模式**：详情/编辑/确认操作统一使用 Modal，半屏宽，遮罩点击关闭。

### G.3 关键交互规格

**审核操作流程**：
1. 审核列表 → 点击行 → Modal 弹出详情（左侧经验内容+作者，右侧 AI 六维打分+解读）
2. Modal 底部操作栏：[通过] [拒绝（弹出理由输入框）] [退回重审] [添加备注]
3. 操作后刷新列表，该条目消失

**批量审核流程**：
1. 勾选多条 → 顶部出现批量操作栏
2. [全选当前页] [批量通过] [批量拒绝]
3. 确认后显示结果提示「成功 N 条 / 失败 M 条」

**用户禁用流程**：
1. 用户列表 → 点击行 → Modal 用户详情
2. [禁用用户] 按钮 → 二次确认弹窗 → 填写理由 → 确认
3. 用户状态标签变为红色「已禁用」

---

## 附 H：数据库表关系（管理后台视角）

```
users ──1:N── experiences ──1:N── likes (user_id, experience_id)
  │              │                └─ bookmarks (user_id, experience_id)
  │              ├── quality_score, score_reason, score_details (JSONB)
  │              ├── review_status (ENUM: pending/approved/rejected/private)
  │              ├── source_type (platform/user)
  │              └── deleted_at (软删除时间戳)
  │
  ├── is_admin (BOOLEAN)  ← admin 中间件检查
  ├── deleted_at ← 用户禁用机制
  └── apple_user_id (UNIQUE)

system_config (key TEXT PK, value JSONB)
  ├── review_mode, content_min_length, ...
  └── domain_display_{name}, domain_active_{name}, ...

admin_logs (id UUID PK, admin_id FK→users, action_type, target_type, target_id, detail JSONB, result, created_at)

sensitive_words (id SERIAL PK, word VARCHAR UNIQUE)

conversations (id, user_id FK→users, title, created_at)
  └── messages (id, conversation_id FK→conversations, role, content, referenced_experience_ids, created_at)

user_views (user_id FK→users, experience_id FK→experiences, viewed_at)
```

标记 `→` 的行表示外键关系（级联）。`admin_logs` 不设 FK 级联（管理员删除后日志保留）。

截图：MEDIA:/Users/swt/.hermes/cache/screenshots/browser_screenshot_c378e2b9cb4c478cbb467ebd87b0b50e.png
