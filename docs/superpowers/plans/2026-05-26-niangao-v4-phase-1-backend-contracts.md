# Niangao V4 Phase 1 Backend Contracts Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use task-quality-gate before each task. Use test-driven-development for Go behavior changes. Use verification-before-completion before claiming completion. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Establish the V4 backend contract and schema foundation for the App-first implementation line.

**Architecture:** Keep the existing Go/Gin backend and PostgreSQL deployment path, but introduce V4 canonical fields and tables before rewriting App flows. Existing legacy fields remain only as compatibility inputs while new APIs and downstream features read V4 facts: `experience_type`, `visibility`, `lifecycle_status`, `quality_tier`, `recommendation_status`, `ai_citable`, event tables, collection/inspiration tables, chat topic tables, and AI Gateway metadata.

**Tech Stack:** Go/Gin, pgx/PostgreSQL migrations, Expo React Native contracts, DeepSeek through server-side AI Gateway later.

---

## File Map

- Modify: `backend/internal/model/models.go`
  - Add V4 enum types, validation helpers, request/response structs, and compatibility conversion from old source/private/status fields.
- Modify: `backend/internal/model/models_test.go`
  - Add table-driven tests for V4 enum defaults, distribution eligibility, and optional domain/subdomain validation.
- Create: `backend/migrations/017_v4_core_foundation.sql`
  - Add V4 columns to `users` and `experiences`.
  - Backfill V4 fields from existing data.
  - Create `experience_collections`, `experience_inspirations`, `experience_events`, `experience_metrics`, `recommendation_sessions`, `chat_topics`, `chat_temp_sessions`, `chat_citations`, `ai_function_configs`, `ai_prompt_registry`, `ai_call_logs`, and `ai_jobs`.
- Create: `docs/implementation/niangao-v4-phase-1-contracts.md`
  - Record the App/backend API contracts for the first implementation line.
- Create or modify later: `backend/internal/repository/experience_v4.go`
  - Implement V4 feed/search repository methods once the migration and model tests are green.
- Create or modify later: `backend/internal/handler/feed.go`
  - Implement `/api/v1/feed/recommend`, `/api/v1/feed/collections`, `/api/v1/feed/mine`.
- Create or modify later: `backend/internal/handler/inspiration.go`
  - Implement `/api/v1/experiences/:id/inspire`.

---

## Task 1: V4 Model Foundation

**Files:**
- Modify: `backend/internal/model/models.go`
- Modify: `backend/internal/model/models_test.go`

- [ ] **Step 1: Write failing tests for V4 enum behavior**

Add tests that require:

```go
func TestV4ExperienceEnums(t *testing.T) {
    if !IsValidExperienceType(ExperienceTypePlatformSelected) {
        t.Fatal("platform_selected should be valid")
    }
    if !IsValidVisibility(VisibilityPublic) || !IsValidVisibility(VisibilityPrivate) {
        t.Fatal("public/private should be valid visibility values")
    }
    if !IsValidLifecycleStatus(LifecycleNeedsReview) {
        t.Fatal("needs_review should be a valid lifecycle status")
    }
    if !IsValidQualityTier(QualityTierAICitable) {
        t.Fatal("ai_citable should be a valid quality tier")
    }
    if !IsValidRecommendationStatus(RecommendationEligible) {
        t.Fatal("eligible should be a valid recommendation status")
    }
}
```

- [ ] **Step 2: Run RED**

Run:

```bash
./scripts/backend-test.sh
```

Expected: FAIL because V4 enum helpers do not exist.

- [ ] **Step 3: Implement V4 enum types and helpers**

Add:

```go
type ExperienceType string
const (
    ExperienceTypePlatformSelected ExperienceType = "platform_selected"
    ExperienceTypeUserOriginal ExperienceType = "user_original"
)

type Visibility string
const (
    VisibilityPublic Visibility = "public"
    VisibilityPrivate Visibility = "private"
)

type LifecycleStatus string
const (
    LifecycleActive LifecycleStatus = "active"
    LifecycleHidden LifecycleStatus = "hidden"
    LifecycleDeleted LifecycleStatus = "deleted"
    LifecycleNeedsReview LifecycleStatus = "needs_review"
)

type QualityTier string
const (
    QualityTierUnreviewed QualityTier = "unreviewed"
    QualityTierPrivateOnly QualityTier = "private_only"
    QualityTierPublicVisible QualityTier = "public_visible"
    QualityTierRecommendCandidate QualityTier = "recommend_candidate"
    QualityTierAICitable QualityTier = "ai_citable"
    QualityTierHighTrust QualityTier = "high_trust"
)

type RecommendationStatus string
const (
    RecommendationEligible RecommendationStatus = "eligible"
    RecommendationIneligible RecommendationStatus = "ineligible"
    RecommendationSuppressed RecommendationStatus = "suppressed"
)
```

Add validation maps and `CanDistributePublicly()`, `CanBeAICitedPublicly()` helpers.

- [ ] **Step 4: Run GREEN**

Run:

```bash
./scripts/backend-test.sh
```

Expected: PASS.

## Task 2: V4 Core Migration

**Files:**
- Create: `backend/migrations/017_v4_core_foundation.sql`

- [ ] **Step 1: Write migration SQL**

Migration must be idempotent and safe on the existing accumulated schema.

Key rules:

- `users.display_name` defaults from `nickname`.
- `experiences.owner_user_id` defaults from `author_id`.
- `experiences.experience_type` defaults from `source_type='platform' OR is_official=true`.
- `experiences.visibility` defaults from `is_private`.
- `experiences.lifecycle_status` defaults from `deleted_at`, `status`, and `review_status`.
- `experiences.quality_tier` defaults from `review_status`, `is_private`, and `quality_score`.
- `experiences.recommendation_status` and `ai_citable` default from `visibility`, `lifecycle_status`, and `quality_tier`.
- New interaction tables must not drop or rewrite legacy `likes/bookmarks`; they run in parallel during migration.

- [ ] **Step 2: Static migration inspection**

Run:

```bash
rg -n "CREATE TABLE IF NOT EXISTS|ALTER TABLE experiences|ALTER TABLE users|CREATE INDEX IF NOT EXISTS" backend/migrations/017_v4_core_foundation.sql
```

Expected: shows every V4 table and index.

- [ ] **Step 3: Run backend tests**

Run:

```bash
./scripts/backend-test.sh
```

Expected: PASS; migration file should not break compilation.

## Task 3: V4 API Contract Document

**Files:**
- Create: `docs/implementation/niangao-v4-phase-1-contracts.md`

- [ ] **Step 1: Document the first App/backend API contracts**

Cover:

- `POST /api/v1/experiences`
- `POST /api/v1/experiences/rewrite`
- `GET /api/v1/feed/recommend`
- `GET /api/v1/feed/collections`
- `GET /api/v1/feed/mine`
- `POST /api/v1/experiences/:id/inspire`
- `POST /api/v1/experiences/:id/collect`
- `DELETE /api/v1/experiences/:id/collect`
- `GET /api/v1/search/experiences`
- initial chat topic/temp-session endpoints.

- [ ] **Step 2: Self-check contracts against PRD**

Check:

- Uses 看看 / 聊聊 / 记下 / 我的 names.
- No old 点赞 wording in new contracts.
- Domain/subdomain/topic are optional for user publish.
- Public review failure is not user-visible.
- User-facing AI calls go through backend only.

## Task 4: First Repository/API Slice

**Files:**
- Create or modify: `backend/internal/repository/experience_v4.go`
- Create or modify: `backend/internal/handler/feed.go`
- Modify: `backend/cmd/server/main.go`
- Add tests under `backend/internal/handler` and/or `backend/internal/model`.

- [ ] **Step 1: Write failing tests for feed route registration and response shape**

Use Gin tests to assert:

- `GET /api/v1/feed/recommend` route exists.
- Response object has `data` and `next_cursor`.
- Unauthenticated users can read recommend.
- Collections and mine require auth.

- [ ] **Step 2: Implement minimal feed handlers**

Implement handlers that read V4 columns and preserve old list behavior only as fallback.

- [ ] **Step 3: Run backend tests**

Run:

```bash
./scripts/backend-test.sh
```

Expected: PASS.

## Task 5: Verification Gate

- [ ] Run:

```bash
./scripts/backend-test.sh
./scripts/backend-build-linux.sh /tmp/niangao-backend-phase1
cd mobile && npm run test
cd mobile && npm run typecheck
cd mobile && npm run expo:check
```

- [ ] Inspect:

```bash
git diff --check
git diff --stat
```

- [ ] Update:

```text
docs/implementation/niangao-v4-phase-0-audit.md
docs/implementation/niangao-v4-phase-1-contracts.md
```

with the actual completed state and remaining risks.

---

## Self-Review

- Spec coverage: This plan covers the first App/backend implementation foundation: V4 experience facts, behavior events, feed/search/chat/AI Gateway storage, and first feed/API contracts. It intentionally excludes admin UI and actual content production execution.
- Placeholder scan: No task relies on unspecified behavior; later tasks name exact files and acceptance checks.
- Type consistency: V4 names match the PRD and architecture: `experience_type`, `visibility`, `lifecycle_status`, `quality_tier`, `recommendation_status`, `ai_citable`, `experience_collections`, `experience_inspirations`, and `experience_events`.
