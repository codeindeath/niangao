# Errors

Command failures and integration errors.

---

## [ERR-20260528-039] search_card_action_overflow

**Logged**: 2026-05-28T02:10:00+08:00
**Priority**: high
**Status**: fixed
**Area**: frontend

### Summary
SearchCardScreen could show action buttons from the previous paged search card clipped at the top of the current card when opening a non-zero result index.

### Error
```
The simulator screenshot showed the 有启发 / 收藏 action bar at the top edge of SearchCardScreen, partially clipped by the status area.
```

### Context
- SearchCardScreen renders paged cards in a FlatList with `initialScrollIndex`.
- ExperienceCard positioned its action bar with `bottom: -32` so Home could float actions below the card.
- In paged search results, that negative offset could overflow into the next FlatList item and appear at the top after snapping to `initialIndex > 0`.

### Suggested Fix
Keep floating action offset contextual: preserve the Home default, but let SearchCardScreen pass a non-negative `actionBottom` so actions stay inside the current paged card. Add a regression test that asserts search-card action bars use non-negative bottom positioning.

### Metadata
- Reproducible: yes
- Related Files: mobile/src/components/ExperienceCard.tsx, mobile/src/screens/SearchCardScreen.tsx, mobile/src/__tests__/SearchCardScreen.test.tsx
- Evidence: /tmp/niangao-search-card-detail2.png, /tmp/niangao-searchcard-actions-fixed.png

---

## [ERR-20260527-038] rntl_find_query_false_negative

**Logged**: 2026-05-28T00:34:00+08:00
**Priority**: medium
**Status**: pending
**Area**: frontend

### Summary
React Native Testing Library public `findByText` / `findByLabelText` queries returned false negatives in `DetailScreen` tests even when the failure snapshot clearly contained the target text or accessibility label.

### Error
```
Unable to find an element with text: 测试经验
Unable to find an element with accessibility label: 私密经验
```

### Context
- The failure snapshot showed `测试经验`, `价值度`, stars, and `accessibilityLabel="私密经验"` already rendered.
- The older full-tree `UNSAFE_root.findAllByType(Text)` approach was slow and timed out under load.
- The stable fix was to use small test-local helpers over `toJSON()` for text assertions and targeted `UNSAFE_root.findAll` for accessibility labels / press targets.

### Suggested Fix
For this RN 0.81 / React 19 / jest-expo 54 setup, prefer focused render-tree helpers in flaky component tests when public RNTL queries produce false negatives despite rendered snapshots.

### Metadata
- Reproducible: yes
- Related Files: mobile/src/__tests__/DetailScreen.test.tsx
- See Also: ERR-20260527-021

---

## [ERR-20260527-037] git_add_ignored_tracked_paths

**Logged**: 2026-05-28T00:24:00+08:00
**Priority**: low
**Status**: pending
**Area**: workflow

### Summary
Staging an explicit mixed path list failed because ignored directory patterns include `.learnings` and `backend/cmd/server`, even though files inside them are already tracked.

### Error
```
The following paths are ignored by one of your .gitignore files:
.learnings
backend/cmd/server
```

### Context
- Attempted one `git add` command with tracked modified files plus new source/test files.
- `git add -u` successfully staged tracked modifications under ignored directories.
- A separate `git add` staged the new non-ignored backend source/test files.

### Suggested Fix
When tracked files live under ignored directory patterns, stage tracked changes with `git add -u`, then stage new files separately.

### Metadata
- Reproducible: yes
- Related Files: .gitignore
- See Also: ERR-20260527-035

---

## [ERR-20260527-036] remote_journal_scan_rg_missing

**Logged**: 2026-05-28T00:19:00+08:00
**Priority**: low
**Status**: pending
**Area**: infra

### Summary
The production server does not have `rg`, so remote `journalctl` scans using ripgrep fail even though local repo scans should still prefer `rg`.

### Error
```
bash: line 1: rg: command not found
```

### Context
- Attempted remote backend/AI journal scans over SSH with `rg`.
- The scans were rerun with `grep -Ei` / `grep -c` and passed.
- This is a remote host tooling difference, not an application failure.

### Suggested Fix
Use `grep` for production-server one-off log scans unless `rg` is explicitly installed there.

### Metadata
- Reproducible: yes
- Related Files: docs/implementation/niangao-v4-phase-1-progress.md
- See Also: ERR-20260527-035

---

## [ERR-20260527-035] backend_test_script_workdir

**Logged**: 2026-05-27T23:15:00+08:00
**Priority**: low
**Status**: pending
**Area**: tests

### Summary
The backend full-test helper was first launched from `backend/`, but the script lives under the repository root `scripts/`.

### Error
```
zsh:1: no such file or directory: ./scripts/backend-test.sh
```

### Context
- Attempted command: `./scripts/backend-test.sh`
- Incorrect workdir: `/Users/swt/projects/niangao/backend`
- Correct workdir: `/Users/swt/projects/niangao`
- The command was rerun from the repository root and passed.

### Suggested Fix
Use repo-root workdir for project helper scripts under `scripts/`; use `backend/` only for raw Go package commands.

### Metadata
- Reproducible: yes
- Related Files: scripts/backend-test.sh
- See Also: ERR-20260527-020

---

## [ERR-20260527-034] intermittent_ssh_close_and_nested_awk_escape

**Logged**: 2026-05-27T21:16:00+08:00
**Priority**: low
**Status**: pending
**Area**: infra

### Summary
Production smoke verification hit intermittent SSH session closures and one nested shell quoting error around `awk`.

### Error
```
Connection closed by 115.190.177.146 port 22
awk: cmd. line:1: {print \}
```

### Context
- Operation attempted: deprecated-route request-id smoke and remote service/hash verification after deploying `/tmp/niangao-backend-v4-deprecated-request-id`.
- Public HTTP health and deprecated-route smoke worked while one SSH heredoc session closed before running.
- A separate SSH log scan succeeded and found the smoke request id with zero severe backend/AI log matches.
- A remote hash command failed because nested quoting produced an invalid `awk '{print \}'`.
- Retrying the SSH hash command after a short wait, without nested `awk`, succeeded and confirmed the expected backend binary hash.

### Suggested Fix
For remote hash checks, print full `sha256sum` output or use simpler quoting. If SSH closes before command execution but public health is OK, verify port/key health, wait briefly, then retry once before changing deployment state.

### Metadata
- Reproducible: no
- Related Files: docs/implementation/niangao-v4-phase-1-progress.md
- See Also: ERR-20260527-033

---

## [ERR-20260527-033] github_remote_internal_error_on_push

**Logged**: 2026-05-27T20:57:00+08:00
**Priority**: low
**Status**: pending
**Area**: infra

### Summary
A `git push` was rejected by the GitHub remote with an Internal Server Error after the local commit had already been created.

### Error
```
remote: Internal Server Error
remote: Request ID 73242514c48be1c0eca2407209e7c91e
error: failed to push some refs to 'github.com:codeindeath/niangao.git'
```

### Context
- Operation attempted: push commit `3d4273e docs: remove token-like examples`.
- Local branch was ahead by one commit after the failed push.
- `git ls-remote` confirmed the remote branch still pointed at the previous commit.
- A plain retry of `git push` succeeded and advanced the remote branch to the local commit.

### Suggested Fix
When a push fails with a remote 5xx after commit creation, verify local ahead state and the remote branch head, then retry the push before changing the commit.

### Metadata
- Reproducible: no
- Related Files: docs/admin-prd-v1.md
- See Also: ERR-20260527-028

---

## [ERR-20260527-032] backend_subdir_repo_relative_paths

**Logged**: 2026-05-27T18:52:00+08:00
**Priority**: low
**Status**: pending
**Area**: tests

### Summary
A backend formatting/test command was run from the `backend/` directory while still using repository-root-relative paths.

### Error
```
lstat backend/internal/handler/auth.go: no such file or directory
lstat backend/internal/handler/app_error_contract_test.go: no such file or directory
```

### Context
- Operation attempted: `gofmt` and targeted Go tests after changing `backend/internal/handler/auth.go`.
- The command's working directory was `/Users/swt/projects/niangao/backend`, so paths starting with `backend/` pointed to a non-existent nested directory.
- The corrected command used `internal/handler/auth.go` and `internal/handler/app_error_contract_test.go` from the backend workdir.

### Suggested Fix
When `workdir` is `backend/`, use backend-relative paths for Go files and packages. Use repository-root paths only when the command workdir is the repository root.

### Metadata
- Reproducible: yes
- Related Files: backend/internal/handler/auth.go
- See Also: ERR-20260527-031

---

## [ERR-20260527-031] command_substitution_grep_pipefail_count

**Logged**: 2026-05-27T18:31:00+08:00
**Priority**: low
**Status**: pending
**Area**: infra

### Summary
A production journal scan script exited before printing results because `grep` found no severe-log matches inside a command substitution while `set -euo pipefail` was active.

### Error
```
ssh log scan exited with code 1 and no diagnostic output.
```

### Context
- Operation attempted: count backend/AI severe log patterns after the account-cancellation copy deploy.
- The script used `backend_severe=$(grep -Eai '...' "$backend_log" | wc -l | tr -d ' ')`.
- With `set -euo pipefail`, no `grep` matches made the command substitution fail before the script printed counts.
- The corrected script used `(grep -Eai '...' "$backend_log" || true) | wc -l`, then verified severe counts were zero and all smoke request IDs appeared in backend logs.

### Suggested Fix
When counting optional matches under `set -euo pipefail`, wrap `grep` with `|| true` inside the command substitution, not only in top-level pipelines.

### Metadata
- Reproducible: yes
- Related Files: docs/implementation/niangao-v4-phase-1-progress.md
- See Also: ERR-20260527-025

---

## [ERR-20260527-030] macos_tar_appledouble_python_compile_failure

**Logged**: 2026-05-27T17:45:11+08:00
**Priority**: medium
**Status**: pending
**Area**: infra

### Summary
A production AI service source archive created on macOS included AppleDouble `._*.py` metadata files, causing remote Python compile verification to fail.

### Error
```
*** Error compiling 'app/._main.py'...
Sorry: ValueError: source code string cannot contain null bytes
```

### Context
- The AI service source tarball was created locally and extracted on the Linux production host.
- Remote pytest passed, but `compileall` failed on AppleDouble metadata files such as `app/._main.py` and `tests/._test_gateway.py`.
- The service had not been restarted yet, so the bad metadata did not become a running service issue.
- Removing `._*` files from `app` and `tests` allowed `compileall` to pass and the service restart to complete.

### Suggested Fix
For macOS-created deployment tarballs, set `COPYFILE_DISABLE=1` or delete `._*` files before remote compile/restart. Keep `find app tests -name '._*' -delete` as a defensive remote cleanup step.

### Metadata
- Reproducible: yes
- Related Files: ai-service/app/main.py
- See Also: ERR-20260527-029

---

## [ERR-20260527-029] zsh_path_variable_breaks_smoke_script

**Logged**: 2026-05-27T16:34:48+08:00
**Priority**: low
**Status**: pending
**Area**: infra

### Summary
A production smoke script used a local variable named `path` in zsh, which overwrote the shell's command search path and made basic commands unavailable.

### Error
```
request:10: command not found: curl
request:12: command not found: jq
zsh:23: command not found: rm
```

### Context
- The smoke helper accepted a route path argument but assigned it to `path`.
- In zsh, `path` is tied to `PATH`, so setting it inside the function broke command lookup for later commands.
- The backend service was not implicated; the failure happened before valid HTTP assertions.

### Suggested Fix
Avoid variable names `path` and `PATH` in zsh smoke scripts. Use names such as `route_path` for URL paths.

### Metadata
- Reproducible: yes
- Related Files: docs/implementation/niangao-v4-phase-1-progress.md
- See Also: ERR-20260527-018

---

## [ERR-20260527-028] scp_connection_closed_during_backend_deploy

**Logged**: 2026-05-27T14:50:00+08:00
**Priority**: low
**Status**: pending
**Area**: infra

### Summary
A production backend deploy upload failed because the SCP connection was closed before the artifact reached the timestamped deployment directory.

### Error
```
Connection closed by 115.190.177.146 port 22
scp: Connection closed
```

### Context
- Operation attempted: upload `/tmp/niangao-backend-v4-profile-error-copy` to `/root/niangao/deployments/20260527144814/server`.
- The failure happened before replacing `/root/niangao/backend/server`; a follow-up SSH check confirmed `niangao-backend` was still active and still running the previous artifact hash.
- The upload was retried, the uploaded hash matched the local artifact, and only then was the backend binary replaced and restarted.

### Suggested Fix
After any interrupted deploy upload, verify the running binary hash and service health before retrying. Do not proceed to `install` or restart until the uploaded artifact hash matches the local artifact.

### Metadata
- Reproducible: no
- Related Files: docs/implementation/niangao-v4-phase-1-progress.md
- See Also: ERR-20260527-004

---

## [ERR-20260527-027] python39_union_type_annotation

**Logged**: 2026-05-27T04:40:51Z
**Priority**: low
**Status**: pending
**Area**: ai-service

### Summary
AI service tests failed because new middleware used Python 3.10 union type syntax while the service venv runs Python 3.9.

### Error
```
TypeError: unsupported operand type(s) for |: 'type' and 'NoneType'
```

### Context
- `ai-service/app/middleware/request_id.py` initially annotated `value: str | None`.
- The local and production AI service runtime use Python 3.9, where PEP 604 union syntax is not supported without newer Python.
- The fix changed the annotation to `Optional[str]` and reran the AI request-id test, full AI pytest suite, and ruff.

### Suggested Fix
Use Python 3.9-compatible type annotations in `ai-service` unless the runtime is explicitly upgraded; prefer `typing.Optional` over `T | None`.

### Metadata
- Reproducible: yes
- Related Files: ai-service/app/middleware/request_id.py
- See Also: ERR-20260527-023

---

## [ERR-20260527-026] shell_printf_dash_format

**Logged**: 2026-05-27T04:40:51Z
**Priority**: low
**Status**: pending
**Area**: infra

### Summary
A remote diagnostic script failed because `printf` received a format string starting with dashes.

### Error
```
bash: line 6: printf: --: invalid option
printf: usage: printf [-v var] format [arguments]
```

### Context
- The script used `printf '--- backend request id matches ---\n'`.
- In this shell, the leading dashes were treated as an option-like format argument.
- The corrected command used `printf '%s\n' '--- backend request id matches ---'`.

### Suggested Fix
For diagnostic separators in remote shell scripts, use `printf '%s\n' '--- text ---'` instead of putting dash-prefixed text in the format string.

### Metadata
- Reproducible: yes
- Related Files: docs/implementation/niangao-v4-phase-1-progress.md
- See Also: ERR-20260527-024

---

## [ERR-20260527-025] shell_pipefail_grep_count

**Logged**: 2026-05-27T04:40:51Z
**Priority**: low
**Status**: pending
**Area**: infra

### Summary
A production smoke script exited before cleanup verification because `grep` found no request-id log matches under `set -euo pipefail`.

### Error
```
Command exited after the successful HTTP smoke output, before printing log-match counts.
```

### Context
- The smoke wrote successful backend health and rewrite results, but then counted journal matches with `journalctl ... | grep -F "$request_id" | wc -l`.
- With `pipefail`, a no-match `grep` makes the pipeline fail even when `wc` would produce `0`.
- The corrected smoke wrapped grep with `|| true` before `wc -l` and added a cleanup trap for the temporary smoke user.

### Suggested Fix
When counting optional log matches under `pipefail`, use `(journalctl ... | grep -F "$needle" || true) | wc -l`, and add cleanup traps around production smoke data.

### Metadata
- Reproducible: yes
- Related Files: docs/implementation/niangao-v4-phase-1-progress.md
- See Also: ERR-20260527-019

---

## [ERR-20260527-024] nested_remote_python_fstring_quotes

**Logged**: 2026-05-27T03:11:57Z
**Priority**: low
**Status**: pending
**Area**: infra

### Summary
An AI Gateway smoke parser failed after a successful 200 response because a nested remote Python f-string lost quotes around a dictionary key.

### Error
```
NameError: name 'function_type' is not defined
```

### Context
- The remote curl wrote a successful AI Gateway `chat_topic_classify` response and status file.
- The subsequent parser used nested shell quoting around `body.get("function_type")`, which was transformed into an unquoted name inside the remote Python script.
- Re-running the parser with single-quoted Python dictionary keys printed the expected function type, topic decision, and clarity score.

### Suggested Fix
For Python snippets embedded inside remote shell commands, prefer single-quoted dictionary keys inside the heredoc or avoid f-strings that require escaped quotes.

### Metadata
- Reproducible: yes
- Related Files: docs/implementation/niangao-v4-phase-1-progress.md
- See Also: ERR-20260527-023

---

## [ERR-20260527-023] production_ai_venv_missing_ruff

**Logged**: 2026-05-27T03:07:36Z
**Priority**: low
**Status**: pending
**Area**: infra

### Summary
The AI service deploy script stopped after syncing code because the production virtualenv does not include the development-only `ruff` executable.

### Error
```
bash: line 5: ./venv/bin/ruff: No such file or directory
```

### Context
- Local AI verification had already run `ai-service/venv/bin/ruff check ai-service/app ai-service/tests` successfully.
- The production sync completed, and remote `pytest tests/test_llm.py` passed before the script reached the missing ruff executable.
- The corrected remote verification used the production virtualenv for `pytest tests -q` plus `compileall`, then restarted `niangao-ai`.

### Suggested Fix
Keep ruff as a local development gate unless production explicitly installs dev tooling; for production post-sync checks, use service-runtime checks such as pytest, compileall, service restart, health checks, and journal scans.

### Metadata
- Reproducible: yes
- Related Files: ai-service/app/services/llm.py, ai-service/app/core/config.py
- See Also: ERR-20260527-020

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

---

## [ERR-20260527-018] remote_python_here_doc_quoting

**Logged**: 2026-05-27T07:25:59+08:00
**Priority**: low
**Status**: pending
**Area**: infra

### Summary
A remote health smoke command failed because a Python heredoc lost quotes around a file path inside nested shell quoting.

### Error
```
File "<stdin>", line 2
    p=json.load(open(/tmp/feed.out))
                     ^
SyntaxError: invalid syntax
```

### Context
- Command attempted a remote `curl` health/feed smoke plus inline Python JSON parsing over `ssh`.
- The nested local shell, remote shell, and Python heredoc quoting stripped the quotes around `/tmp/feed.out`.
- The backend health check in the same command still returned `{"status":"ok"}`; the JSON parse portion must be retried with safer quoting.

### Suggested Fix
For remote smoke parsing, prefer `python3 -c '...'` with a simple quoted string, or use a single-quoted remote heredoc (`ssh host 'bash -s' <<'REMOTE'`) so Python code is not altered by nested shell interpolation.

### Metadata
- Reproducible: yes
- Related Files: docs/implementation/niangao-v4-phase-1-progress.md
- See Also: ERR-20260527-015

---

## [ERR-20260527-021] mobile_api_config_path_assumption

**Logged**: 2026-05-27T09:08:00+08:00
**Priority**: low
**Status**: pending
**Area**: frontend

### Summary
A follow-up inspection command used the stale path `mobile/src/config.ts`; the active App API config file is `mobile/src/services/config.ts`.

### Error
```
sed: mobile/src/config.ts: No such file or directory
```

### Context
- The command tried to inspect API error parsing after the structured backend error deployment.
- The App imports API transport from `mobile/src/services/config.ts`, and tests live under `mobile/src/__tests__/config.test.ts`.
- The corrected command read `mobile/src/services/config.ts` successfully.

### Suggested Fix
For App API transport and `ApiError` behavior, inspect `mobile/src/services/config.ts`; do not use the older top-level `mobile/src/config.ts` path.

### Metadata
- Reproducible: yes
- Related Files: mobile/src/services/config.ts
- See Also: ERR-20260527-020

---

## [ERR-20260527-022] react_native_actual_spread_triggers_getters

**Logged**: 2026-05-27T09:36:00+08:00
**Priority**: low
**Status**: pending
**Area**: tests

### Summary
A ChatScreen FlatList test mock initially spread `jest.requireActual('react-native')`, which triggered React Native index getters and failed on unavailable native modules.

### Error
```
Invariant Violation: TurboModuleRegistry.getEnforcing(...): 'DevMenu' could not be found.
```

### Context
- Full mobile Jest verification surfaced a React Native `VirtualizedList` delayed `act(...)` warning from ChatScreen tests.
- The first mock attempted `return {...RN, FlatList}`, which enumerated React Native getters such as DevMenu and deprecated core modules.
- The corrected mock uses `Object.defineProperty(RN, 'FlatList', {value: FlatList})` and returns `RN` without spreading, avoiding getter evaluation.

### Suggested Fix
When overriding one React Native export in Jest, patch the actual module object with `Object.defineProperty` instead of spreading `jest.requireActual('react-native')`.

### Metadata
- Reproducible: yes
- Related Files: mobile/src/__tests__/ChatScreen.test.tsx
- See Also: ERR-20260527-021

---

---

## [ERR-20260527-020] go_test_from_repo_root

**Logged**: 2026-05-27T08:54:00+08:00
**Priority**: low
**Status**: pending
**Area**: tests

### Summary
A targeted Go test failed because it was launched from the repo root instead of the `backend/` Go module directory.

### Error
```
go: cannot find main module, but found .git/config in /Users/swt/projects/niangao
```

### Context
- The command combined root-relative `gofmt` paths with a raw `go test ./internal/middleware ...`.
- `gofmt` succeeded from the repo root, but raw Go test commands need `workdir=/Users/swt/projects/niangao/backend`.
- The test was rerun from `backend/` and passed.

### Suggested Fix
Run raw Go test commands from `backend/`, or use repo-root scripts such as `./scripts/backend-test.sh` for full-suite verification.

### Metadata
- Reproducible: yes
- Related Files: backend/internal/middleware/logging_test.go
- See Also: ERR-20260527-015

---

## [ERR-20260527-019] production_smoke_users_schema_assumption

**Logged**: 2026-05-27T07:31:00+08:00
**Priority**: medium
**Status**: pending
**Area**: infra

### Summary
The first production authenticated quota smoke assumed the legacy `users.wechat_openid` column still exists, but the production `users` table now uses Apple-auth fields without that column.

### Error
```
ERROR: column "wechat_openid" of relation "users" does not exist
quota-smoke-http 401
{"error":"请先登录"}
```

### Context
- The smoke inserted a temporary user with `(wechat_openid, apple_user_id, nickname, display_name)`.
- Production schema inspection showed current user columns include `apple_user_id`, `nickname`, `display_name`, and related V4 fields, but not `wechat_openid`.
- The script also lacked `ON_ERROR_STOP=1`, so later steps continued with an empty user id and produced secondary UUID errors.
- The corrected smoke uses `apple_user_id` only and adds `ON_ERROR_STOP=1` for all production SQL.

### Suggested Fix
Before writing production smoke setup SQL, inspect current production columns or reuse existing smoke helpers. Always pass `-v ON_ERROR_STOP=1` to `psql` in smoke scripts so setup failures stop before API assertions.

### Metadata
- Reproducible: yes
- Related Files: docs/implementation/niangao-v4-phase-1-progress.md
- See Also: ERR-20260527-015
