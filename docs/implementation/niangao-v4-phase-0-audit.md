# Niangao V4 Phase 0 Audit

Date: 2026-05-26

## 1. Purpose

This audit records the current state before V4 implementation starts. It is intended to prevent the old implementation from silently constraining the new product.

Phase 0 follows `docs/implementation/niangao-v4-app-backend-development-plan.md`.

## 2. Current Repository State

Current branch:

- `codex/niangao-v4-implementation`

Working tree state before implementation:

- Modified: `.learnings/ERRORS.md`
- Modified: `.learnings/LEARNINGS.md`
- Untracked: `docs/design-drafts/`
- Untracked: `docs/implementation/`
- Untracked: `docs/product-decisions-2026-05-24.md`
- Untracked: `docs/product/`
- Untracked: `scripts/`

Rule for later phases:

- Do not delete or overwrite these untracked/generated product documents.
- Do not stage `.learnings/*` unless explicitly requested.
- Implementation changes should be kept separate from prior documentation/design artifacts.

## 3. Verification Results

### 3.1 Backend

Original local command:

```bash
cd backend && go test ./...
```

Result:

- Failed before tests ran.
- Local shell cannot find `go`.
- `docker` is also not available in the current shell, so Docker-based Go verification was not available.

Updated direction:

- On 2026-05-26, the implementation workflow was aligned to the previous project deployment style: backend code is developed, tested, and built from this Mac, then Linux build artifacts and required migrations are uploaded to the cloud server.
- The production directory `/root/niangao` must not be used as a development workspace.
- Because this Mac is not Linux, deployment build commands must explicitly use `GOOS=linux GOARCH=amd64 CGO_ENABLED=0`.

Phase 0 fix:

- Installed a local Go toolchain outside the repo at `~/.local/toolchains/go1.26.3`.
- Added `scripts/backend-test.sh`.
- Added `scripts/backend-build-linux.sh`.

Latest commands:

```bash
./scripts/backend-test.sh
./scripts/backend-build-linux.sh /tmp/niangao-backend-verify
```

Latest result:

- Backend tests pass.
- Linux deployment binary builds successfully as `ELF 64-bit LSB executable, x86-64, statically linked`.

Risk:

- The existing cloud production directory `/root/niangao` is a dirty working tree and must not be used as the active development workspace.
- A plain local `go build` on macOS would produce a Darwin binary; deployment must use a Linux cross-compile command.

### 3.2 Mobile

Command:

```bash
cd mobile && npx tsc --noEmit
```

Result:

- Passed.

Command:

```bash
cd mobile && npx jest --runInBand
```

Original result:

- Failed before tests ran.
- Both current test suites fail with:

```text
ReferenceError: You are trying to `import` a file outside of the scope of the test code.
```

Likely cause:

- Expo/Jest version mismatch and current Jest setup incompatibility.

Phase 0 fix:

- Aligned Expo test dependencies:
  - `jest`: `~29.7.0`
  - `jest-expo`: `~54.0.17`
  - `react-native-worklets`: `0.5.1`
- Added test environment mocks for AsyncStorage, safe area, and React Navigation hooks.
- Updated stale HomeScreen and DetailScreen tests to assert current API signatures and current star/label UI.
- Added package scripts:
  - `npm run test`
  - `npm run typecheck`
  - `npm run expo:check`

Latest commands:

```bash
cd mobile && npm run test
cd mobile && npm run typecheck
cd mobile && npm run expo:check
```

Latest result:

- Jest passes: 2 suites, 8 tests.
- TypeScript check passes.
- Expo dependency check passes.

Command:

```bash
cd mobile && npx expo-doctor
```

Result:

- 17/18 checks passed after dependency alignment.
- Failed native tooling check: CocoaPods version check failed or CocoaPods unavailable.
- Package version check is now clean.

Risk:

- Mobile type checking is usable.
- Mobile Jest baseline is usable for current smoke coverage.
- iOS native build may fail until CocoaPods/tooling is repaired.

### 3.3 Admin

Command:

```bash
cd admin && npm run build
```

Result:

- Passed.
- Vite warns that the main JS chunk is larger than 500 kB.

Command:

```bash
cd admin && npm run lint
```

Result:

- Failed.
- 26 errors and 2 warnings.
- Main categories:
  - `react-hooks/set-state-in-effect`
  - `@typescript-eslint/no-explicit-any`
  - `react-hooks/static-components`

Risk:

- Admin is buildable but not lint-clean.
- Admin is not on the first implementation critical path.
- Later admin rewrite should not preserve old lint violations.

### 3.4 AI Service

Command:

```bash
cd ai-service && ./venv/bin/python -m pytest -q
```

Result:

- Passed: 12 tests.

Command:

```bash
cd ai-service && ./venv/bin/python -m ruff check .
```

Result:

- Passed.

Note:

- Local AI venv uses Python 3.9.6.
- CI workflow expects Python 3.11.

Risk:

- Current AI service is locally healthy.
- Runtime version should be aligned with CI and production before implementation.

### 3.5 Remote Service Health

Backend:

```bash
curl http://115.190.177.146:8080/health
```

Result:

```json
{"status":"ok"}
```

Nginx root health:

```bash
curl http://115.190.177.146/health
```

Result:

- 404.

Public API through Nginx:

```bash
curl 'http://115.190.177.146/api/v1/experiences?page=1&page_size=1'
```

Result:

- Reachable.
- Returned `total: 385`.

AI service:

```bash
curl http://115.190.177.146:8000/health
```

Result:

```json
{"status":"ok"}
```

Risk:

- Current backend and AI service are running remotely.
- `/health` is not exposed through the Nginx root path, while `/api/v1/*` is exposed.
- Later deployment should standardize health endpoints.

## 4. Current Implementation Summary

### 4.1 Mobile

Current stack:

- Expo `~54.0.33`
- React Native `0.81.5`
- React `19.1.0`
- React Navigation
- Apple authentication
- AsyncStorage token storage

Current useful assets:

- Full-screen experience card feed.
- Vertical card switching.
- Horizontal tab switching.
- Card flip interaction.
- Existing domain/subdomain selection interaction in record page.
- Apple login integration.
- Profile/statistics page patterns.

Current conflicts with V4:

- Bottom tabs still use old names: 首页 / 对话 / 记录 / 我的.
- Login slogan is old: 记录经验，年年成长.
- App has no guest-first 看看 flow matching the final design.
- API base URL is hardcoded to `http://115.190.177.146`.
- `AI_BASE` exists in mobile config and `generateInterpretation` calls AI service directly.
- Experience action semantics still use like/liked/like_count instead of 有启发/inspiration.
- Tests use old domains such as `life` and old subdomains such as `time-mgmt`.
- My page still has old entries such as 经验包 and 对话人格.

### 4.2 Backend

Current stack:

- Go/Gin backend.
- PostgreSQL via pgx.
- JWT and Apple login.
- Existing health endpoint.
- Existing admin routes.

Current useful assets:

- Server bootstrapping and route grouping.
- Auth/JWT infrastructure.
- Apple login handler.
- Domain v3 constants include `meaning/emotion`.
- Existing repositories and handler tests provide starting points.
- Current remote API is reachable and returns real data.

Current conflicts with V4:

- Database schema is an accumulated legacy schema, not V4.
- Old fields still drive behavior: `review_status`, `status`, `like_count`, `source_type`, `is_official`.
- `likes` table still represents 点赞, while V4 needs 有启发.
- User publish flow synchronously returns review status/reason to the user; V4 says users should not perceive failed public review.
- Recommendation is bookmark-count threshold + random/domain preference, not V4 multi-pool recommendation.
- Chat is a single conversation per user, not V4 temporary sessions plus topics.
- Chat context uses inferred domain and recent bookmarks, but lacks topic records, citations, and structured save suggestions.
- AI review/translation/interpretation calls are ad hoc HTTP calls, not an AI Gateway with function_type, key alias, prompt version, schema version, logging, and budget.

### 4.3 AI Service

Current useful assets:

- FastAPI service.
- DeepSeek client wrapper.
- Chat prompt has useful humanistic style direction.
- Normalize, review, translate, interpretation endpoints exist.
- Tests pass.

Current conflicts with V4:

- No AI Gateway abstraction.
- No function_type config table.
- No key alias separation.
- No prompt_version/schema_version logging.
- Most endpoints return ad hoc text or simple JSON, not the production schemas in the AI spec.
- Chat returns no real citation records.
- Prompt pack in product docs is more advanced than the current implementation.

### 4.4 Admin

Current useful assets:

- React/Vite app builds.
- Admin auth skeleton.
- Layout and table pages can be reused structurally.
- Existing API client pattern is usable.

Current conflicts with V4:

- Admin is still review-queue/content-management centered.
- UI and modules do not match V4 admin positioning.
- Lint is currently failing.
- Admin is no longer first implementation priority.

## 5. Keep / Refactor / Delete Matrix

### 5.1 Mobile

Keep:

- Expo project structure.
- React Navigation stack.
- Apple login integration.
- AsyncStorage token persistence.
- Experience card flip and vertical feed mechanics as implementation reference.
- Domain/subdomain selection logic from CreateScreen.
- Existing screen/test folder organization.

Refactor:

- `HomeScreen` into 看看 with final B2 design, star quality, creator icon, recommended/bookmarked/mine tabs, proper guest mode.
- `CreateScreen` into 记下 with final copy, default public behavior, weak public/private control, 帮我改改, optional domain/subdomain/topic.
- `ChatScreen` into 聊聊 with topic button, reference experience cards, save-suggestion message, and V4 message states.
- `ProfileScreen` into 我的 with recent harvest, accumulation, contribution feedback, and final visual hierarchy.
- `LoginScreen` into selected A visual design with background image, 年糕, 生活有态度, Apple登录, 先看看.
- `services/api.ts` into V4 API contracts.
- `services/config.ts` into environment-based config.
- Existing tests after API and UI contracts are updated.

Delete / replace:

- Old tab labels 首页 / 对话 / 记录.
- Old login slogan.
- Direct mobile `AI_BASE` calls.
- Hardcoded API URL in multiple files.
- Like/liked/like_count user-facing semantics.
- Placeholder entries 经验包 and 对话人格.
- Old review status UI that exposes rejected/private audit semantics to users.

### 5.2 Backend

Keep:

- Gin server entrypoint.
- CORS/auth middleware structure where still valid.
- Apple/JWT auth infrastructure.
- Domain/subdomain constants, with validation.
- Existing repository structure as a starting pattern.
- Existing tests as regression references.

Refactor:

- Experience schema to V4 lifecycle and quality model.
- Like flow into inspiration feedback.
- Review pipeline into asynchronous classification/moderation without telling users why public review did not pass.
- Recommendation into V4 multi-pool feed sessions and event tracking.
- Chat into temporary sessions, topics, message citations, summaries, and save suggestions.
- Stats into V4 My page aggregation.
- AI calls into AI Gateway.
- Environment and deployment config.

Delete / replace:

- `review_status` as primary user-facing state.
- `like_count` as user-facing 点赞.
- Synchronous public review response messages.
- Old platform/manual production paths from the first implementation line.
- Old WeChat-related config comments and residue.

### 5.3 AI Service

Keep:

- FastAPI application.
- DeepSeek client wrapper.
- Existing humanistic chat prompt as a source reference.
- Normalize/review/translate tests as regression references.

Refactor:

- Add production AI Gateway contract.
- Implement function_type-based routing.
- Add prompt version and schema version handling.
- Add structured outputs and parser validation.
- Add call logs, retry strategy, timeout, and degradation.
- Connect golden-set regression to implementation verification.

Delete / replace:

- Direct ad hoc model calls from handlers.
- Mobile-direct interpretation endpoint usage.
- Non-versioned prompt behavior.

### 5.4 Admin

Keep:

- Vite/React project foundation.
- Admin login shell.
- API client pattern.
- Table/list page scaffolding where useful.

Refactor later:

- Navigation and modules into V4 admin structure.
- Private content permission and audit.
- AI config and call logs.
- Experience management around V4 lifecycle.

Delete / replace later:

- ReviewQueue-centered information architecture.
- Old status/like/source labels that conflict with V4.

## 6. App / Backend MVP Gaps

P0 gaps before coding:

- iOS CocoaPods/tooling needs repair before native build verification.
- API contracts still reflect old semantics.
- Database migrations need a V4 migration strategy instead of continuing legacy field mixing.
- Backend lacks event tables for exposure, flip, inspiration, collect, search click, and AI call logs.
- Backend lacks V4 feed session/cursor model.
- Backend lacks chat topics, citations, temp sessions, save suggestions, summaries, and topic lifecycle.
- Backend lacks AI Gateway tables and service layer.
- App lacks final visual design implementation.
- App lacks guest-first 看看 flow.
- App still has direct AI service config.
- Remote Nginx lacks `/health` forwarding.

## 7. Phase 1 Entry Tasks

Before feature implementation:

1. Define V4 App/backend API contracts before UI rewrite.
2. Define V4 migration strategy:
   - Either add forward migrations from current DB.
   - Or create clean V4 schema plus controlled data migration.
3. Implement backend P0 data model and API contracts.
4. Implement AI Gateway foundation before App depends on AI features.
5. Implement App screens against V4 contracts and final design.
6. Repair CocoaPods availability before iOS native build verification.

## 8. Risk Assessment

High:

- Current DB schema mixes old and new concepts. Continuing directly will make V4 implementation inconsistent.
- App direct AI service access violates the server-owned AI architecture.
- Deployment builds require explicit Linux cross-compilation.

Medium:

- iOS CocoaPods/tooling still needs repair before native build verification.
- Admin lint failures are broad but admin is no longer first-path.
- Remote services are running, but health routing is inconsistent.
- AI prompt implementation is materially behind product prompt specs.

Low:

- Backend local test/build baseline is now available through scripts.
- Mobile Jest and TypeScript baselines are now passing.
- Existing App interaction code provides useful implementation references.
- AI service tests and ruff are currently clean.
- Admin build currently passes.

## 9. Phase 0 Conclusion

The project is ready to enter Phase 1 backend/API contract work. The local verification baseline is now usable, but Phase 1 must still start with schema/API contract cleanup before large App UI work.

The recommended implementation order is:

1. Lock V4 backend API and schema contracts.
2. Implement backend core.
3. Implement AI Gateway.
4. Implement App vertical slices.
5. Deploy App/backend usable version.
6. Add admin MVP.
7. Add content production capability.
8. Plan and execute actual content production separately.
