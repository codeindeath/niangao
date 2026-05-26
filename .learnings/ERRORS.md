# Errors

Command failures and integration errors.

---

## [ERR-20260526-001] production_smoke_psql_variable_mismatch

**Logged**: 2026-05-26T23:10:00+08:00
**Priority**: medium
**Status**: resolved
**Area**: infra

### Summary
Production authenticated smoke script initially failed because it used stale user schema assumptions and unsafe `psql -c` variable handling.

### Error
```
syntax error at or near ":"
column "wechat_openid" of relation "users" does not exist
invalid input syntax for type uuid
```

### Context
- Command attempted: create a temporary production smoke user, generate a JWT without printing secrets, call authenticated V4 endpoints, then clean up the user.
- Production `users` now uses `apple_user_id`; legacy `wechat_openid` has been removed by cleanup migrations.
- `psql -c` did not interpolate `:'var'` as expected in this context, and non-quiet insert output polluted the captured UUID.

### Suggested Fix
For production smoke scripts, read current table columns first when schema may have drifted, use `apple_user_id` for temporary users, and capture IDs with `psql -X -qAt` plus a `WITH inserted AS (...) SELECT id FROM inserted` query.

### Metadata
- Reproducible: yes
- Related Files: backend/migrations/006_cleanup_wechat.sql, backend/internal/handler/auth.go

---

## [ERR-20260525-001] local_playwright_missing

**Logged**: 2026-05-25T19:46:54+08:00
**Priority**: low
**Status**: pending
**Area**: frontend

### Summary
Local screenshot verification with Node Playwright failed because the project does not have `playwright` installed.

### Error
```
Cannot find module 'playwright'
```

### Context
- Command attempted: `node -e "console.log(require.resolve('playwright'))"`
- Fallback: use installed Google Chrome headless screenshot mode for static design verification.

### Suggested Fix
For design-draft workflows, either add a lightweight screenshot script dependency or standardize on Chrome headless commands.

### Metadata
- Reproducible: yes
- Related Files: docs/design-drafts/niangao-main-pages-three-directions.html

---

## [ERR-20260524-001] ui_ux_pro_max_missing_search_script

**Logged**: 2026-05-24T15:51:52+08:00
**Priority**: low
**Status**: pending
**Area**: frontend

### Summary
The ui-ux-pro-max skill documentation referenced a `scripts/search.py` helper, but the installed skill directory only contains `SKILL.md` and `_meta.json`.

### Error
```
can't open file '/Users/swt/.codex/skills/ui-ux-pro-max-2/scripts/search.py': [Errno 2] No such file or directory
```

### Context
- Command attempted: run the skill's documented design-system search helper for Niangao mobile design.
- Local skill directory inspected after failure and no `scripts/` directory exists.
- Fallback: apply the skill's design rules manually.

### Suggested Fix
Update the installed skill package to include the referenced script, or adjust `SKILL.md` to describe the no-script fallback workflow.

### Metadata
- Reproducible: yes
- Related Files: /Users/swt/.codex/skills/ui-ux-pro-max-2/SKILL.md

---

## [ERR-20260503-001] git_push_github

**Logged**: 2026-05-03T10:30:00+08:00
**Priority**: high
**Status**: resolved
**Area**: infra

### Summary
GitHub push failed multiple times: password auth deprecated, missing workflow scope, HTTPS timeout

### Error
```
1. remote: Invalid username or token. Password authentication is not supported
2. remote: refusing to allow PAT to create workflow without `workflow` scope
3. fatal: unable to access '...': Recv failure: Operation timed out
4. git@github.com: Permission denied (publickey)
```

### Context
- Push from China to GitHub using HTTPS with Personal Access Token
- First failure: token missing `workflow` scope for `.github/workflows/ci.yml`
- Second failure: HTTPS connection timed out (GFW interference)
- SSH attempt: public key not added to GitHub account

### Suggested Fix
- Ensure PAT has both ✅ repo AND ✅ workflow scopes
- Use SSH (git@github.com:) for China-based pushes (HTTPS unreliable)
- Generate ed25519 SSH key, add to github.com/settings/keys
- For quick workaround: use HTTPS with PAT but expect occasional timeouts

### Metadata
- Reproducible: yes (China network)
- Related Files: .github/workflows/ci.yml
- See Also: ERR-20260503-002

---

## [ERR-20260503-002] github_pages_404

**Logged**: 2026-05-03T10:30:00+08:00
**Priority**: high
**Status**: resolved
**Area**: infra

### Summary
Both project site and user GitHub Pages site returned 404 after Save in Pages settings

### Error
```
https://codeindeath.github.io/niangao → 404 "Site not found"
https://codeindeath.github.io/.well-known/apple-app-site-association → 404
```

### Context
- Pages settings configured correctly: main branch, /docs (project) or / (user)
- Save button confirmed active on both repos
- Waited >5 minutes between checks
- No error in GitHub UI — just silent deployment stall
- User site (codeindeath.github.io) is supposed to auto-deploy but first deployment never triggered

### Suggested Fix
Use Netlify Drop as primary static hosting. GitHub Pages only as backup when first deployment is already confirmed working. For AASA (apple-app-site-association), deploy via Netlify with netlify.toml redirects.

### Resolution
- **Resolved**: 2026-05-03T23:50:00+08:00
- **Notes**: Switched to Netlify Drop. Deployed AASA at cheerful-hotteok-4933a9.netlify.app with netlify.toml 200-rewrite for .well-known path. Verified JSON served correctly.

### Metadata
- Reproducible: yes (first-deploy queue issue)
- Related Files: netlify.toml, .well-known/apple-app-site-association

---

## [ERR-20260503-003] pingfang_font_oserror

**Logged**: 2026-05-03T10:30:00+08:00
**Priority**: low
**Status**: resolved
**Area**: frontend

### Summary
PIL ImageFont.truetype() failed with PingFang.ttc on macOS

### Error
```
OSError: cannot open resource
```

### Context
- System Python path: /System/Library/Fonts/PingFang.ttc
- Method: ImageFont.truetype(path, size)
- Root cause: PingFang.ttc is a TrueType Collection that requires explicit `index` parameter, but even with index=0 it fails on some macOS versions

### Suggested Fix
Use /System/Library/Fonts/Supplemental/Songti.ttc with `index=0` instead. For any .ttc TrueType Collection, always pass `index=0` to ImageFont.truetype().

### Metadata
- Reproducible: yes
- Related Files: None (generation script)

---

## [ERR-20260523-001] ai_service_invalid_dependency_pin

**Logged**: 2026-05-23T11:27:25+08:00
**Priority**: high
**Status**: resolved
**Area**: backend

### Summary
AI service dependency install failed because `pydantic-settings==2.14.0` is not available from the current package index.

### Error
```
ERROR: Could not find a version that satisfies the requirement pydantic-settings==2.14.0
```

### Context
- Command attempted: create local virtualenv, install `ai-service/requirements.txt`, then run pytest.
- `pip index versions pydantic-settings` showed `2.11.0` as the latest available version.

### Suggested Fix
Pin `pydantic-settings` to an available version and verify tests in a clean virtualenv.

### Metadata
- Reproducible: yes
- Related Files: ai-service/requirements.txt

---

## [ERR-20260526-002] repo_root_script_path_mismatch

**Logged**: 2026-05-26T23:59:00+08:00
**Priority**: low
**Status**: resolved
**Area**: tests

### Summary
Backend helper scripts are rooted at the repository root, not inside `backend/`.

### Error
```
zsh: no such file or directory: ./scripts/backend-test.sh
```

### Context
- The failed command was run with working directory `/Users/swt/projects/niangao/backend`.
- The script path is `/Users/swt/projects/niangao/scripts/backend-test.sh`.
- The same path assumption also caused an initial `gofmt` miss when using `backend/internal/...` from inside `backend/`.

### Suggested Fix
Run repo helper scripts from `/Users/swt/projects/niangao`, or use paths relative to the selected working directory. From `backend/`, use `internal/...` paths for Go package files.

### Metadata
- Reproducible: yes
- Related Files: scripts/backend-test.sh, scripts/backend-build-linux.sh

---

## [ERR-20260527-001] production_smoke_cleanup_unreferenced_cte

**Logged**: 2026-05-27T00:18:00+08:00
**Priority**: medium
**Status**: resolved
## [ERR-20260527-010] production_chat_message_nil_reference_array

**Logged**: 2026-05-27T04:45:00+08:00
**Priority**: high
**Status**: resolved
**Area**: backend

### Summary
Production chat smoke exposed that user chat messages without citations passed a nil Go slice into `referenced_experience_ids`, violating the production NOT NULL `uuid[]` column.

### Error
```
ERROR: null value in column "referenced_experience_ids" of relation "chat_messages" violates not-null constraint (SQLSTATE 23502)
```

### Context
- `POST /api/v1/chat/temp-sessions/:id/messages` saved a user message with no referenced experiences.
- `AddChatMessage` passed `req.ReferencedExperienceIDs` directly to `$9::uuid[]`; when nil, PostgreSQL received NULL instead of the default empty array.
- The table definition requires `referenced_experience_ids UUID[] NOT NULL DEFAULT '{}'`.

### Suggested Fix
Normalize nil `ReferencedExperienceIDs` to `[]string{}` before executing the insert, and keep a regression test so citation-free chat messages persist an empty UUID array.

### Metadata
- Reproducible: yes
- Related Files: backend/internal/repository/chat_v4.go, backend/migrations/017_v4_core_foundation.sql

---

## [ERR-20260527-011] production_chat_candidate_future_column_drift

**Logged**: 2026-05-27T04:49:00+08:00
**Priority**: medium
**Status**: resolved
**Area**: backend

### Summary
Production chat smoke showed the chat candidate query referenced a future content-production field that is not present in the deployed V4 schema.

### Error
```
ERROR: column e.source_derivation_type does not exist (SQLSTATE 42703)
```

### Context
- The chat reply still degraded and returned 200 because candidate retrieval errors are non-fatal.
- Missing candidates would prevent 聊聊 from citing relevant reference experiences.
- Migration 017 includes `source_reliability` but does not add `experiences.source_derivation_type`.

### Suggested Fix
Do not directly reference future columns in active App-facing queries. Emit a compatible fallback value in the candidate payload until the content-production schema is introduced.

### Metadata
- Reproducible: yes
- Related Files: backend/internal/repository/chat_v4.go, backend/migrations/017_v4_core_foundation.sql

---

## [ERR-20260527-012] production_chat_candidate_collection_status_drift

**Logged**: 2026-05-27T04:52:00+08:00
**Priority**: medium
**Status**: resolved
**Area**: backend

### Summary
Production chat candidate retrieval used a legacy soft-delete predicate on `experience_collections`, but the V4 collection table uses `status`.

### Error
```
ERROR: column c.deleted_at does not exist (SQLSTATE 42703)
```

### Context
- The authenticated chat promotion smoke returned 200 because candidate retrieval is non-fatal.
- Backend logs showed the candidate query failed before AI citation candidates could be passed into the chat gateway.
- Migration 017 defines `experience_collections.status` with values `active` and `removed`, plus `removed_at`.

### Suggested Fix
Use `experience_collections.status='active'` in App-facing V4 queries. Avoid legacy `deleted_at` assumptions for V4 interaction tables.

### Metadata
- Reproducible: yes
- Related Files: backend/internal/repository/chat_v4.go, backend/migrations/017_v4_core_foundation.sql

---

**Area**: deployment

### Summary
PostgreSQL did not execute a DELETE hidden inside an unreferenced CTE during production smoke cleanup.

### Error
```
cleanup_remaining=1
```

### Context
- Authenticated production smoke inserted a temporary user with `apple_user_id='codex-smoke-handler-cleanup-20260527001542'`.
- Cleanup query used `WITH deleted AS (DELETE FROM users ...) SELECT COUNT(*) ...`, but the `deleted` CTE was not referenced by the final SELECT.
- PostgreSQL can skip unreferenced data-modifying CTEs, so the temporary row remained until an explicit `DELETE FROM users ...` command was run.

### Suggested Fix
For production smoke cleanup, run `DELETE ...` as its own statement, then run a separate `SELECT COUNT(*) ...` verification. Do not rely on an unreferenced CTE for side effects.

### Metadata
- Reproducible: yes
- Related Files: None

---

## [ERR-20260527-008] zsh_status_readonly_in_smoke_script

**Logged**: 2026-05-27T03:22:00+08:00
**Priority**: low
**Status**: resolved
**Area**: deployment

### Summary
A production smoke script attempted to assign curl's exit code to `status`, but `status` is read-only in zsh.

### Error
```
zsh: read-only variable: status
```

### Context
- Public smoke expected deprecated endpoints to return HTTP 410, so the script needed to tolerate curl's nonzero exit under `-f`.
- The command used `|| status=$?`, which fails under zsh because `status` is a special read-only parameter.
- The endpoint status lines had already shown `health=200`, `recommend=200`, `search=200`, and `deprecated_list=410`; only the wrapper variable caused the command failure.

### Suggested Fix
Avoid assigning to `status` in zsh scripts. Use a different variable name, or avoid `curl -f` for expected non-2xx smoke checks and compare `%{http_code}` directly.

### Metadata
- Reproducible: yes
- Related Files: None

---

## [ERR-20260527-006] backend_subdir_path_prefix

**Logged**: 2026-05-27T02:51:10+08:00
**Priority**: low
**Status**: resolved
**Area**: tests

### Summary
Ran `gofmt` from the `backend/` directory while still prefixing paths with `backend/`, causing a no-op path error.

### Error
```
lstat backend/internal/handler/experience.go: no such file or directory
```

### Context
- The command was executed with `workdir=/Users/swt/projects/niangao/backend`.
- Repository-root paths such as `backend/internal/...` only work from `/Users/swt/projects/niangao`.
- The corrected command used `internal/handler/...` paths from the `backend/` directory and passed.

### Suggested Fix
When `workdir` is `backend/`, use `internal/...` paths for Go package files. Use `backend/internal/...` only from the repository root.

### Metadata
- Reproducible: yes
- Related Files: backend/internal/handler/experience.go

---

## [ERR-20260527-007] psql_returning_command_tag_in_smoke_user_id

**Logged**: 2026-05-27T02:55:48+08:00
**Priority**: medium
**Status**: resolved
**Area**: deployment

### Summary
Production smoke JWT generation captured both the `INSERT ... RETURNING id` row and the trailing `INSERT 0 1` command tag, producing a malformed JWT `user_id`.

### Error
```
invalid input syntax for type uuid: "... INSERT 0 1"
```

### Context
- The smoke script assigned `USER_ID=$(psql ... -c "INSERT ... RETURNING id")`.
- The backend correctly returned 500 because the signed JWT contained a non-UUID `user_id`.
- Temporary data from the failed smoke was cleaned up, then the script was rerun with `psql -XqAt ... | sed -n '1p'` and passed.

### Suggested Fix
When capturing PostgreSQL `RETURNING` values for smoke scripts, use quiet tuple-only output and explicitly keep the first row. Never sign a JWT from unsanitized multiline command output.

### Metadata
- Reproducible: yes
- Related Files: None

---

## [ERR-20260527-009] production_expose_dedupe_context_id_cast

**Logged**: 2026-05-27T03:49:44+08:00
**Priority**: high
**Status**: resolved
**Area**: backend

### Summary
Authenticated production expose-dedupe smoke returned 500 because optional UUID `NULLIF` expressions were parsed as uuid inside the new dedupe SQL.

### Error
```
ERROR: invalid input syntax for type uuid: "" (SQLSTATE 22P02)
```

### Context
- Operation: production authenticated smoke for posting two `expose` events to the same experience.
- The new dedupe CTE first used `NULLIF($5, '')::uuid` for `context_id`; the first fix changed it to `NULLIF($5::text, '')::uuid`.
- A second smoke still failed because `$1` was inferred as uuid elsewhere in the same CTE, so `ev.user_id = NULLIF($1, '')::uuid` coerced the empty-string branch to uuid.
- The authenticated dedupe helper is only called when `userID != ""`, so it should compare `ev.user_id = $1::uuid` instead of using optional-user `NULLIF`.
- Runtime parameter inference allowed empty-string branches to hit uuid parsing before becoming NULL.
- Temporary smoke rows were cleaned up and verified separately.

### Suggested Fix
For optional UUID parameters in SQL queries, force text before `NULLIF`: `NULLIF($param::text, '')::uuid`. For authenticated-only parameters, avoid optional-user `NULLIF` and cast the required parameter directly. Include empty optional UUID values in authenticated production smoke checks.

### Metadata
- Reproducible: yes
- Related Files: backend/internal/repository/experience_actions_v4.go

---

## [ERR-20260527-004] production_backend_wrong_binary_path

**Logged**: 2026-05-27T02:20:00+08:00
**Priority**: high
**Status**: resolved
**Area**: deployment

### Summary
Production deploy initially copied the verified backend artifact to `/root/niangao/server`, but systemd runs `/root/niangao/backend/server`.

### Error
```
deprecated-list 200
```

### Context
- The service `niangao-backend` has `WorkingDirectory=/root/niangao/backend` and `ExecStart=/root/niangao/backend/server`.
- Copying the artifact to `/root/niangao/server` and restarting the service left production behavior unchanged.
- Directly copying over `/root/niangao/backend/server` while running failed with `Text file busy`.
- The successful pattern was to copy the artifact to `/root/niangao/backend/server.new`, stop the service, `mv -f` it into place, then start the service and verify the running binary SHA.

### Suggested Fix
For backend production deploys, deploy to the actual systemd `ExecStart` path: `/root/niangao/backend/server`. Use `server.new` plus stop/move/start instead of overwriting the running binary in place.

### Metadata
- Reproducible: yes
- Related Files: /etc/systemd/system/niangao-backend.service

---

## [ERR-20260527-005] production_smoke_cleanup_same_statement_snapshot

**Logged**: 2026-05-27T02:22:00+08:00
**Priority**: medium
**Status**: resolved
**Area**: deployment

### Summary
Production smoke cleanup verified remaining rows in the same PostgreSQL statement as data-modifying CTEs, so the SELECT observed the pre-delete snapshot.

### Error
```
1|1
```

### Context
- Authenticated smoke cleanup used a single `WITH ... DELETE ... SELECT count(...)` statement.
- The temporary user and chat temp session were actually verified clean only after rerunning explicit DELETE statements followed by a separate SELECT statement.
- The correct cleanup verification returned `0|0`.

### Suggested Fix
Keep smoke cleanup as explicit DELETE statements, then run a separate SQL statement for remaining-row verification. Avoid relying on same-statement visibility for post-delete counts.

### Metadata
- Reproducible: yes
- Related Files: None

---

## [ERR-20260527-002] production_smoke_cleanup_wrong_chat_citation_column

**Logged**: 2026-05-27T01:03:00+08:00
**Priority**: medium
**Status**: resolved
**Area**: deployment

### Summary
Production auth smoke cleanup used the wrong `chat_citations` foreign-key column.

### Error
```
ERROR: column "assistant_message_id" does not exist
```

### Context
- Authenticated production smoke passed for `/me/profile`, `/me/stats/assets`, `/feed/mine`, and `/chat/temp-sessions`.
- Cleanup initially attempted `DELETE FROM chat_citations WHERE assistant_message_id IN (...)`.
- The actual V4 table column is `message_id`, so cleanup aborted before deleting the temporary smoke user and temp session.
- A follow-up explicit cleanup using `message_id` removed the temp session and user, then verified `cleanup_remaining=0`.

### Suggested Fix
For future production smoke cleanup, inspect or remember the actual V4 chat citation schema: `chat_citations.message_id -> chat_messages.id`. Keep cleanup as explicit DELETE statements followed by a separate remaining-row SELECT.

### Metadata
- Reproducible: yes
- Related Files: backend/migrations/017_v4_core_foundation.sql

---

## [ERR-20260527-003] ssh_nested_heredoc_sql_quoting

**Logged**: 2026-05-27T01:03:00+08:00
**Priority**: low
**Status**: resolved
**Area**: deployment

### Summary
Nested `ssh '...'` SQL cleanup stripped UUID quotes and produced invalid PostgreSQL syntax.

### Error
```
ERROR: syntax error at or near "aaab41"
```

### Context
- A cleanup retry embedded SQL inside a single-quoted remote ssh command.
- The UUID literal quotes were stripped before reaching PostgreSQL.
- Sending the script with `ssh root@host 'bash -s' <<'REMOTE' ... REMOTE` preserved SQL quoting and cleanup succeeded.

### Suggested Fix
For multiline production SQL over SSH, prefer a local quoted heredoc into `bash -s` over deeply nested shell quoting.

### Metadata
- Reproducible: yes
- Related Files: None

---

## [ERR-20260527-013] backend_go_test_wrong_workdir

**Logged**: 2026-05-27T05:16:11+08:00
**Priority**: low
**Status**: resolved
**Area**: tests

### Summary
Backend Go tests failed when launched from the repository root instead of the `backend/` module directory.

### Error
```
go: cannot find main module, but found .git/config in /Users/swt/projects/niangao
```

### Context
- The backend Go module lives under `/Users/swt/projects/niangao/backend`.
- `go test ./internal/repository ...` must be run from `backend/`, while repo-root paths are only valid for shell scripts that handle their own working directory.
- Rerunning the same command from `backend/` passed.
- The same path rule applies to `gofmt`: use repo-root paths from the repository root, or backend-relative paths from `backend/`.

### Suggested Fix
Run raw backend Go package tests with `workdir=/Users/swt/projects/niangao/backend`. For `gofmt`, keep paths aligned with the selected working directory. Use repo root only for project scripts such as `./scripts/backend-test.sh`.

### Metadata
- Reproducible: yes
- Related Files: backend/go.mod

---

## [ERR-20260527-014] production_smoke_profile_wrong_key

**Logged**: 2026-05-27T05:21:00+08:00
**Priority**: low
**Status**: resolved
**Area**: deployment

### Summary
Authenticated production smoke for `/api/v1/me/profile` initially checked for a nonexistent top-level `id` key.

### Error
```
profile=200
missing_key=id
```

### Context
- The V4 profile response is `MeProfile` and contains fields such as `display_name`, `career_stage`, `common_issues`, and `profile_version`; it does not include `id`.
- The smoke failure was in the validation script after HTTP 200, not in backend behavior.
- The temporary smoke user from the failed run was cleaned up, and the corrected smoke checked `display_name` and passed.

### Suggested Fix
For `/api/v1/me/profile` smoke checks, validate `display_name` or another actual V4 profile field instead of `id`.

### Metadata
- Reproducible: yes
- Related Files: backend/internal/model/models.go, backend/internal/handler/me_profile_v4.go

---

## [ERR-20260527-015] production_smoke_cleanup_cte_scope_and_schema_assumption

**Logged**: 2026-05-27T05:53:44+08:00
**Priority**: low
**Status**: resolved
**Area**: infra

### Summary
Chat production smoke cleanup failed because the script reused a CTE name across separate SQL statements and then checked a non-existent `ai_call_logs.request_payload` column.

### Error
```
ERROR: relation "smoke_messages" does not exist
ERROR: column "request_payload" does not exist
```

### Context
- The chat-gate production smoke itself passed: messages endpoint returned 200, send endpoint returned 200, and reference cards preserved unavailable placeholders.
- Cleanup initially defined `smoke_messages` in one statement, then referenced it from later DELETE statements where that CTE was out of scope.
- The retry then verified cleanup by assuming `ai_call_logs.request_payload` exists; the production table does not have that column.
- A corrected cleanup used explicit subqueries per DELETE statement and verified temporary users, experiences, topics, messages, and citations as `0|0|0|0|0`.

### Suggested Fix
For production smoke cleanup, do not reuse CTE names across SQL statements. Use explicit subqueries or temp tables, and inspect table schema before adding verification checks for optional log columns.

### Metadata
- Reproducible: yes
- Related Files: backend/internal/repository/chat_v4.go, docs/implementation/niangao-v4-phase-1-progress.md
- See Also: ERR-20260527-001, ERR-20260527-005

---

## [ERR-20260527-016] expo_run_ios_missing_cocoapods_cli

**Logged**: 2026-05-27T06:31:00+08:00
**Priority**: medium
**Status**: pending
**Area**: frontend

### Summary
`npx expo run:ios` cannot be used directly in the current shell because CocoaPods CLI is not on PATH and the automatic installers are unavailable.

### Error
```
Failed to install CocoaPods CLI with gem (recommended)
Cause: gem install cocoapods --no-document exited with non-zero code: 1
Failed to install CocoaPods with Homebrew
Cause: spawn brew ENOENT
```

### Context
- Command attempted from `mobile/`: `npx expo run:ios --device A41B8DA3-B22F-4FF2-9F2B-DE340A07DB14 --port 8081`.
- `expo run:ios` attempted to install CocoaPods before building even though `mobile/ios/Pods` already exists.
- Homebrew is not installed on this machine, so Expo's fallback installer cannot run.
- Use the existing Xcode workspace/build path or install/expose CocoaPods before relying on `expo run:ios`.

### Suggested Fix
For simulator runtime checks in this repo, prefer the already documented `xcodebuild -workspace ios/mobile.xcworkspace ... build` path plus an existing app install/Metro workflow, or add CocoaPods CLI to PATH before invoking `expo run:ios`.

### Metadata
- Reproducible: yes
- Related Files: mobile/package.json, docs/implementation/niangao-v4-phase-1-progress.md
- See Also: ERR-20260527-013

---

## [ERR-20260527-017] osascript_simulator_click_accessibility_denied

**Logged**: 2026-05-27T07:03:08+08:00
**Priority**: medium
**Status**: pending
**Area**: frontend

### Summary
AppleScript/System Events cannot currently be used for simulator tap automation because `osascript` lacks macOS Accessibility permission.

### Error
```
execution error: "System Events" encountered an error: "osascript" is not allowed assistive access. (-25211)
```

### Context
- `xcrun simctl io` in the current Xcode does not support tap input.
- `open -a Simulator` and System Events window inspection work; the iPhone 17 simulator window was visible and measurable.
- Attempted coordinate click: `osascript -e 'tell application "System Events" to click at {720, 715}'`.
- The click path failed before interaction because macOS Accessibility permissions block `osascript`.

### Suggested Fix
Grant Accessibility permission for the terminal/Codex host that runs `osascript`, or use another tap automation path that can generate trusted CGEvents without that permission.

### Metadata
- Reproducible: yes
- Related Files: docs/implementation/niangao-v4-phase-1-progress.md
- See Also: ERR-20260527-016
