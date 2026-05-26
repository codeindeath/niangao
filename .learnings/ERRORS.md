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
