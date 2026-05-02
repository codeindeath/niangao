# 年糕测试体系

> 状态：已覆盖可离线测试的模块，集成/端到端测试需 Supabase 环境就绪后执行

---

## 测试金字塔

```
           ┌─────────┐
           │  E2E    │ 0 个 — 需真机 + 生产环境
           ├─────────┤
           │ 集成测试 │ 0 个 — 需 Supabase + DeepSeek API
           ├─────────┤
           │ 单元测试 │ ✅ 47 个用例已编写
           └─────────┘
```

---

## 已覆盖的测试（47 个用例）

### Go 后端 — 模型层 (models_test.go) — 15 个用例
| 测试 | 覆盖内容 |
|------|---------|
| `TestIsValidDomain` | 5 个合法领域 + 空值 + 不存在 + 大小写敏感 |
| `TestValidDomainsMapping` | 领域枚举和中文名映射完整性 |
| `TestCreateExperienceRequestValidation` | 6 个场景：正常/超100字/恰好100字/空内容/无效领域/中文 |
| `TestInterpretationLength` | 6 个场景：空/短/恰好500/超500/中文500/中文501 |
| `TestExperienceStatusValues` | 状态枚举：published/hidden/flagged/deleted/空 |
| `TestMessageRoleValues` | 角色枚举：user/assistant/system/空/大小写 |
| `TestExperienceListQueryDefaults` | 分页参数默认值 |
| `TestChatRequestValidation` | 正常/空/超长消息 |

### Go 后端 — 中间件层 (auth_test.go) — 10 个用例
| 测试 | 覆盖内容 |
|------|---------|
| `TestCORSMiddleware` | GET/OPTIONS/POST 请求的 CORS 头 |
| `TestAuthMiddlewareNoToken` | 无 token 请求正常通过（不阻塞公开 API） |
| `TestAuthMiddlewareInvalidToken` | 无效 token 不崩溃 |
| `TestAuthMiddlewareMalformedHeader` | 3 种畸形 Authorization 头 |
| `TestRequireAuthMiddleware` | 公开/私有端点、空 user_id 拒绝 |

### Go 后端 — Handler 层 (handler_test.go) — 12 个用例
| 测试 | 覆盖内容 |
|------|---------|
| `TestHealthEndpoint` | /health 返回 ok |
| `TestExperienceCreateValidation` | 10 个场景：正常/缺字段/空body/超100字/恰好100字/中文/带解读/解读超500/恰好500 |
| `TestExperienceCreateRequiresAuth` | 未登录创建经验返回 401 |
| `TestExperienceListQueryParams` | 5 个场景：无参数/domain/sort/page/全部参数 |

### Go 后端 — 配置层 (config_test.go) — 4 个用例
| 测试 | 覆盖内容 |
|------|---------|
| `TestLoadDefaults` | 默认端口 8080、AI URL 默认值 |
| `TestLoadFromEnv` | 从环境变量正确加载全部字段 |
| `TestGetEnvFallback` | 环境变量不存在时的回退逻辑 |
| `TestAllConfigFieldsExist` | 所有字段非空 |

### Python AI 服务 — Prompt 构建 (test_prompts.py) — 12 个用例
| 测试 | 覆盖内容 |
|------|---------|
| `test_prompt_contains_humanistic_principles` | 4 个人本主义原则必须存在 |
| `test_prompt_handles_empty_experiences` | 空经验不报错，无占位符残留 |
| `test_prompt_formats_experiences` | 经验编号列表格式 |
| `test_prompt_truncates_to_max_5` | 最多 5 条经验 |
| `test_prompt_template_is_unchanged` | 核心 Prompt 模板完整性 |
| `test_builds_correct_message_structure` | system + user 消息结构 |
| `test_includes_history` | 历史消息正确插入 |
| `test_truncates_long_history` | 超过 20 条截断 |
| `test_empty_history` | 空历史不报错 |
| `test_default_values` | 配置默认值 |
| `test_experiences_with_missing_fields` | 缺字段的经验不崩溃 |
| `test_special_characters_in_experiences` | XSS/换行符不破坏 Prompt |

---

## 待补的测试（需 Supabase 环境）

| 层级 | 内容 | 优先级 |
|------|------|--------|
| Repository 集成测试 | 真实数据库 CRUD 操作 | P0 |
| API 端到端测试 | HTTP 请求完整流程 | P0 |
| AI 对话集成测试 | 真实 DeepSeek API 调用验证 | P0 |
| 向量检索测试 | pgvector 相似度搜索准确性 | P1 |
| 并发测试 | 点赞/收藏的原子性 | P1 |
| 压力测试 | 1000 并发下的表现 | P2 |
| RN 组件测试 | Jest + React Native Testing Library | P1 |

---

## 下一步

1. 配置 Supabase → 跑集成测试
2. 配置 DeepSeek API → 跑 AI 集成测试
3. 补 Repository 层测试
4. 全部通过后 → 进入阶段 1 剩余开发
