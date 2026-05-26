# Niangao V4 App And Backend Development Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use task-quality-gate before starting each phase. Use Code for implementation and verification. Use frontend-design when touching App UI. Use systematic-debugging for failures. Use verification-before-completion before claiming a phase is complete.

**Goal:** Build the production-grade Niangao iOS App and backend first, with enough AI and data infrastructure to support real users and later content production, while keeping management admin and actual content production as later tracks.

**Architecture:** The first implementation line prioritizes App plus backend usability. Go backend owns business state, PostgreSQL owns durable data, Python AI service is accessed through an AI Gateway, and the React Native App consumes stable API contracts. Backend development and verification follow the previous project workflow: develop, test, and build from this Mac, then upload Linux build artifacts and required migrations to the cloud server. Management admin and content production execution are deliberately delayed so they do not block the App/backend vertical slice.

**Tech Stack:** Expo React Native iOS App, Go/Gin backend, PostgreSQL, Python/FastAPI AI service, DeepSeek V4 Pro through server-side AI Gateway, React/Vite admin later.

---

## 1. Scope

This plan includes:

- iOS App implementation.
- Backend APIs, database schema, authentication, recommendation, chat, experience, user statistics, and deployment readiness.
- User-facing AI capabilities: chat response, topic handling, experience rewrite, classification, interpretation, recommendation support.
- Content production capability infrastructure: source records, batch records, candidate experiences, processing states, queue jobs, AI processing contracts, and later admin entry points.
- A later management admin MVP after App and backend are usable.

This plan does not include:

- Actual internet content crawling.
- Actual large-scale content cleaning and production.
- Cold-start production of 3000 platform experiences.
- Creator/book/site sourcing plans.
- Operational scheduling for content production.

Actual content production must be handled by a separate content production plan after the App/backend foundation is stable.

## 2. Product And Design Baseline

Implementation must follow these local artifacts:

- `docs/product/niangao-user-prd-v4.md`
- `docs/product/niangao-admin-prd-v4.md`
- `docs/product/niangao-technical-architecture-v4.md`
- `docs/product/niangao-ai-functional-prompt-spec-v4.md`
- `docs/product/niangao-ai-prompt-production-spec-v4.md`
- `docs/product/ai-prompt-eval/golden-set-expansion-20260525.md`
- `docs/design-drafts/niangao-app-complete-screens.html`
- `docs/design-drafts/niangao-app-state-screens.html`
- `docs/design-drafts/assets/niangao-login-bg.png`

Implementation rule:

- Use the latest product and design baseline as source of truth.
- Reuse old implementation only when it matches the new product semantics.
- Rewrite or delete old implementation when preserving it would create mixed naming, mixed domain logic, or mixed user mental models.
- Do not preserve compatibility layers for deprecated concepts such as 求经验, 速记, 宝库, old domain enums, WeChat login residue, or review-queue-centered admin structure.

## 3. Phase 0: Baseline Audit And Implementation Split

Goal: decide what can be reused, what must be rewritten, and what must be deleted before coding starts.

Actions:

- Run existing build and test commands for `mobile`, `backend`, `ai-service`, and `admin`.
- Run backend build and test commands from this Mac before deployment. Production deployment uploads Linux build artifacts and required migrations to the cloud server; do not use the production server directory as the development workspace.
- Inspect current routes, schemas, screens, services, and deployment files.
- Compare existing code against the latest PRD, technical architecture, AI specs, and design drafts.
- Audit database migrations for legacy semantics that conflict with the new model.
- Audit environment and deployment references from local project files and Hermes records without printing secrets.
- Produce a keep/refactor/delete matrix for old code.
- Produce App/backend MVP API and data model gap list.

Acceptance:

- Current build/test state is known.
- Dirty and untracked files are identified.
- Old code treatment is explicit.
- Phase 1 implementation tasks can start without guessing.

## 4. Phase 1: Backend Core Skeleton

Goal: provide stable APIs and data models needed by the App.

Implement:

- Apple login, guest mode, JWT, current user profile.
- Domain, subdomain, and topic dictionary.
- Experience model with selected/original, public/private, creator name, star quality, domain, subdomain, topic, interpretation, ownership, and lifecycle state.
- Experience actions: bookmark, inspiration feedback, exposure, flip, search click.
- Feed APIs: recommended, bookmarked, mine, cursor pagination, deduplication, and fallback.
- Note APIs: create, edit, delete, public/private toggle, rewrite request, classification result.
- Chat APIs: session, topic, messages, cited experiences, save-suggestion card, feedback.
- My page APIs: recent harvest, accumulation stats, contribution feedback, recent responded experiences, display name.
- Unified error format, request id, structured logs, health checks.

Defer:

- Full admin workflows.
- Actual content production execution.
- Large-scale production scheduling.

Acceptance:

- Backend tests pass.
- Backend tests pass on this Mac before build/upload.
- Core App API contracts are stable.
- Legacy schema fields do not leak into new semantics.

## 5. Phase 2: AI Gateway And User-Facing AI

Goal: make AI behavior usable in the real App before building lower-priority systems.

Implement:

- Single server-side AI Gateway for all DeepSeek calls.
- Function-level configuration: function_type, prompt version, key alias, timeout, retry, budget, logging, and output schema.
- Chat response style matching Niangao: warm companionship, humanistic tone, not therapy, not rigid structured advice.
- Topic handling that updates gradually from conversation context rather than trusting only the first user message.
- Experience citation retrieval for chat based on current topic, closest domain, and recent/high-quality relevant experiences.
- Note rewrite under 帮我改改, preserving user intent and reducing expression burden.
- Experience classification: domain, subdomain, topic, quality star, privacy/public risk, interpretation eligibility.
- Experience interpretation only for high-quality public experiences.
- Golden-set regression for chat, content extraction/classification, privacy summary, and recommendation diagnosis.

Acceptance:

- App-facing AI calls work through backend only.
- Golden-set failures are traceable to prompt, parser, model output, or business rule.
- No model API key is exposed to App or Admin.

## 6. Phase 3: iOS App Main User Flows

Goal: ship the core user product, not a demo shell.

Implementation order:

1. Login: selected A background, 年糕, 生活有态度, Apple登录, 先看看.
2. Navigation: 看看 / 聊聊 / 记下 / 我的.
3. 看看: 推荐 / 收藏 / 我的, vertical card switching, horizontal page switching, card flip, bookmark, inspiration feedback, search entry.
4. 记下: quick record, 帮我改改, domain/subdomain, topic optional, weak public/private control, edit and delete after publish.
5. 聊聊: fixed title, topic button, message stream, compact reference experiences, centered save-suggestion card.
6. 我的: recent harvest, my accumulation, contribution feedback, recent responded experiences, display name setup.
7. States: logged out, empty, loading, error, weak network, forbidden, end of feed, AI timeout.

Acceptance:

- iOS simulator can run the App.
- Four main pages and key states match design drafts.
- A user can complete: guest browse, Apple login, browse feed, flip card, bookmark, mark inspiration, chat, save an experience, create a note, edit own experience, view personal accumulation.

## 7. Phase 4: App And Backend Usable Deployment

Goal: produce the first real usable system.

Implement:

- Deploy backend and AI service to the existing server or staging environment.
- Connect App to the real backend environment.
- Apply database migrations repeatably.
- Seed enough safe test data for App validation.
- Run smoke tests for login, feed, note, chat, my page, and AI calls.
- Verify logs, health checks, and AI call records.

Acceptance:

- The App does not depend on local mock data for core flows.
- Health endpoints and logs are reachable.
- AI call cost and failure records are inspectable.

## 8. Phase 5: Management Admin MVP Later

Goal: support operation after App/backend usability is proven.

Implement later:

- Admin login and roles.
- Experience list, search, filter, edit, delete, re-review.
- User list, basic statistics, private content access audit.
- Operations overview for users, experiences, feedback, AI calls, and recommendation behavior.
- AI configuration: function_type, prompt version, key alias, failure rate, cost.
- System audit logs.

Acceptance:

- Admin can inspect and manage real App data.
- Private content access is permissioned and audited.

## 9. Phase 6: Content Production Capability Later

Goal: build the system capabilities required for content production, but not execute production in this plan.

Implement later:

- Content source records: platform, URL, creator, raw text, collection time.
- Production batch records: status, count, processing version, owner, result summary.
- Candidate experience records: extracted core, creator, classification, star quality, interpretation, review state.
- Queue states: pending, processing, succeeded, failed, retrying, dead-letter.
- AI processing contracts: extraction, quality audit, classification, interpretation, deduplication.
- Admin entry points: batch view, candidate review, publish, unpublish.

Explicitly excluded:

- Actual crawling.
- Actual large-scale production.
- Cold-start 3000 selected experiences.
- Production source planning.

## 10. Phase 7: Production Hardening

Goal: make the usable App/backend suitable for early production.

Checks:

- Backend API latency, slow query logs, indexes, and connection pool.
- AI timeout, retry, degradation, and cost ceiling.
- Feed exposure deduplication and fallback.
- Private content permissions and admin audit.
- Apple login security.
- Secret scan.
- Database backup and deployment rollback.
- App crash, weak network, repeated taps, stale token, and AI timeout handling.

Acceptance:

- Core paths meet early 10k DAU assumptions.
- Deployment is repeatable.
- Critical failures have diagnosis and fallback paths.

## 11. Priority

P0:

- Backend core model and APIs.
- AI Gateway.
- App four main pages.
- Login, 看看, 聊聊, 记下, 我的.
- Feed, search, bookmark, inspiration feedback.
- Real environment deployment.

P1:

- Management admin MVP.
- Content production capability infrastructure.
- More complete operations dashboard.
- AI cost and prompt management.

P2:

- Actual content production plan.
- Automatic crawling.
- Cold-start 3000 selected experiences.
- Advanced operation strategy and batch workflows.

## 12. Verification Gate

Before claiming any phase complete:

- Backend: run `go test ./...` in `backend` on this Mac. Deployment builds must target the ECS runtime explicitly with `GOOS=linux GOARCH=amd64 CGO_ENABLED=0`.
- Mobile: run available tests in `mobile`, then run the App in iOS simulator when UI is touched.
- AI service: run available tests in `ai-service`, plus golden-set regression when AI behavior is touched.
- Admin: run build when admin code is touched.
- Database: verify migrations are repeatable in a clean local database or documented otherwise.
- UI: compare key screenshots against design drafts when UI is touched.
- Security: verify no API keys or secrets are added to frontend or Git.

## 13. Phase 0 Output Files

Phase 0 should produce:

- `docs/implementation/niangao-v4-phase-0-audit.md`
- A keep/refactor/delete matrix for old code.
- Build/test result summary.
- App/backend MVP gap list.
- Risk list for legacy schema, deployment, AI, and App runtime.
