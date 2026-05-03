# Learnings

Corrections, insights, and knowledge gaps captured during development.

**Categories**: correction | insight | knowledge_gap | best_practice

---

## [LRN-20260503-001] correction

**Logged**: 2026-05-03T10:30:00+08:00
**Priority**: critical
**Status**: resolved
**Area**: frontend

### Summary
auth.ts WeChat login logic: !errCode reversed — success (errCode=0) treated as failure

### Details
In `mobile/src/services/auth.ts`, the condition `if (!authResp.errCode)` evaluates to `true` when errCode is 0 (WeChat SDK success). This caused successful WeChat authorizations to be reported as "用户取消登录". The correct check is `if (authResp.errCode !== 0)`.

### Suggested Action
Fixed in commit d2045a0. Added TypeScript lint awareness: `!` on numeric returns from SDKs is a recurring footgun.

### Metadata
- Source: user_feedback
- Related Files: mobile/src/services/auth.ts
- Tags: wechat, login, reactivity-pattern, numeric-truthiness

---

## [LRN-20260503-002] insight

**Logged**: 2026-05-03T10:30:00+08:00
**Priority**: medium
**Status**: pending
**Area**: infra

### Summary
GitHub Pages first-deploy is unreliable; use Netlify Drop as primary

### Details
Both project site (codeindeath.github.io/niangao) and user site (codeindeath.github.io) failed to deploy after Save in Pages settings. Pages deployment queue stalls silently with no error. Netlify Drop works instantly but subdirectories (.well-known/) may not be deployed — requires root copy + netlify.toml redirect.

### Suggested Action
For any future static hosting needs (landing pages, AASA, etc.), start with Netlify Drop. Only try GitHub Pages as fallback. Document the Netlify Drop workflow as a reusable skill.

### Metadata
- Source: error
- Related Files: netlify.toml, .well-known/apple-app-site-association
- Tags: deployment, github-pages, netlify, china-network

---

## [LRN-20260503-003] insight

**Logged**: 2026-05-03T10:30:00+08:00
**Priority**: medium
**Status**: resolved
**Area**: config

### Summary
config_test.go compiled against stale Supabase fields after migration to WeChat auth

### Details
After migrating the Config struct from Supabase-dependent fields (SupabaseJWTSecret, SupabaseURL) to WeChat-dependent fields (JWTSecret, WechatAppID, WechatAppSecret), config_test.go still referenced the old field names. Tests failed to compile.

### Suggested Action
When renaming config fields, run `grep` for all references before committing. Added cross-repo search step to migration workflow.

### Metadata
- Source: error
- Related Files: backend/internal/config/config.go, backend/internal/config/config_test.go
- Tags: refactoring, tests, compile-error, supabase-cleanup

---

## [LRN-20260503-004] insight

**Logged**: 2026-05-03T10:30:00+08:00
**Priority**: high
**Status**: resolved
**Area**: backend

### Summary
User stats cache fields (experience_count, bookmark_count, practiced_count) had no trigger sync

### Details
The `users` table had three counter cache fields but no database triggers to update them when experiences were created/deleted, bookmarks were added/removed, or `practiced` flag changed on bookmarks. Values would always be 0.

### Suggested Action
Added 3 PostgreSQL triggers in migration schema. For any future counter cache fields, always add corresponding triggers in the same migration.

### Metadata
- Source: code_review
- Related Files: backend/migrations/001_initial_schema.sql
- Tags: postgresql, triggers, cache-coherence, data-integrity

---

## [LRN-20260503-005] insight

**Logged**: 2026-05-03T10:30:00+08:00
**Priority**: low
**Status**: resolved
**Area**: tests

### Summary
models_test.go used len() (byte count) instead of utf8.RuneCountInString() (character count)

### Details
Chinese characters are 3 bytes in UTF-8. Using len() to check "max 100 characters" would allow content with >100 Chinese characters but <300 bytes. The Gin binding uses characters, so the test was testing the wrong constraint.

### Suggested Action
Use `utf8.RuneCountInString()` for all character-length assertions in Go tests involving CJK text.

### Metadata
- Source: code_review
- Related Files: backend/internal/model/models_test.go
- Tags: unicode, testing, character-encoding, cjk

---

## [LRN-20260503-006] knowledge_gap

**Logged**: 2026-05-03T10:30:00+08:00
**Priority**: medium
**Status**: pending
**Area**: infra

### Summary
DeepSeek V4 Pro does not support image/vision — browser_vision and vision_analyze fail with "unknown variant image_url"

### Details
All calls to browser_vision and vision_analyze return 400 with the error message indicating the model doesn't accept image content. This means screenshots for QA cannot be analyzed by the primary model. Workaround: use browser_snapshot for text-based DOM verification, or switch to a vision-capable model (Claude/GPT-4o) when visual analysis is needed.

### Suggested Action
- Primary model (DeepSeek): use browser_snapshot + DOM-based verification
- When visual verification is required: temporarily switch to vision-capable model
- Document which models support vision for this deployment

### Metadata
- Source: error
- Tags: model-capability, vision, deepseek, workaround

---

## [LRN-20260503-007] insight

**Logged**: 2026-05-03T10:30:00+08:00
**Priority**: low
**Status**: resolved
**Area**: infra

### Summary
PIL/Pillow icon generation: complex shapes (rice cakes, plants) look amateur; simple geometry + gradients wins

### Details
After designing 4 icon iterations, the user rejected flat typography ("太丑了"), nature sprout ("还是太丑了"), and ultimately selected the glow/premium style (dark green gradient + glowing circle + star spark). PIL drawing primitives cannot produce professional-quality icons with complex shapes. The winning design used only: rounded rect, ellipse, gradient, polygon — all mathematically simple operations that PIL handles cleanly.

### Suggested Action
For future PIL icon generation: skip complex shapes entirely. Offer: 1) flat typography, 2) nature/symbol, 3) glow/premium. Start with glow/premium as default.

### Metadata
- Source: user_feedback
- Tags: design, pil, icon-generation, aesthetic-judgment
