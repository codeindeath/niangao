# Errors

Command failures and integration errors.

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
