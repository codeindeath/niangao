# Niangao V4 Phase 1 App/Backend Contracts

Date: 2026-05-26

This document freezes the first App/backend contracts for V4 implementation. It intentionally focuses on the App-first backend line. Admin and actual content production execution are later tracks.

Source of truth:

- `docs/product/niangao-user-prd-v4.md`
- `docs/product/niangao-technical-architecture-v4.md`
- `docs/product/niangao-ai-functional-prompt-spec-v4.md`
- `docs/implementation/niangao-v4-app-backend-development-plan.md`

## 1. Contract Principles

- New user-facing naming is fixed: 看看 / 聊聊 / 记下 / 我的.
- New user-facing feedback is fixed: 有启发, not 点赞.
- App does not call the AI service directly. App calls Go backend; backend calls AI Gateway or existing AI worker during transition.
- Logged-out users can use `先看看` for public recommend/search/detail browsing. Actions that create personal state require login: 聊聊, 记下, 收藏, 有启发, 看看-收藏, 看看-我的, 我的页 stats/profile.
- User-created experience fields `domain`, `sub_domain`, and `topic` are optional.
- User publish failures from public moderation are not surfaced as review failure. User sees only saved or request-level validation errors.
- Public distribution uses V4 facts: `visibility`, `lifecycle_status`, `quality_tier`, `recommendation_status`, `ai_citable`.
- Legacy fields may be backfilled and read as compatibility fallbacks during migration, but new handlers should not expose legacy semantics.

## 2. Shared Experience Card Shape

Backend response object:

```json
{
  "id": "uuid",
  "owner_user_id": "uuid",
  "content": "100字以内经验正文",
  "experience_type": "platform_selected",
  "visibility": "public",
  "lifecycle_status": "active",
  "domain": "meaning",
  "sub_domain": "self",
  "topic": "#自我",
  "creator_display_name": "李小龙",
  "interpretation_status": "ready",
  "interpretation_summary_available": true,
  "quality_tier": "ai_citable",
  "star_rating": 4,
  "inspiration_count": 12,
  "collection_count": 8,
  "is_collected": false,
  "is_inspired": false,
  "unavailable_reason": ""
}
```

Allowed values:

- `experience_type`: `platform_selected`, `user_original`.
- `visibility`: `public`, `private`.
- `lifecycle_status`: `active`, `hidden`, `deleted`, `needs_review`.
- `quality_tier`: `unreviewed`, `private_only`, `public_visible`, `recommend_candidate`, `ai_citable`, `high_trust`.
- `recommendation_status`: `eligible`, `ineligible`, `suppressed`.
- `interpretation_status`: `none`, `pending`, `ready`, `stale`, `failed`.

Visibility rules:

- Public card returns content only when `visibility=public`, `lifecycle_status=active`, and user permission allows it.
- Private card returns content only for owner.
- Unavailable card returns `id`, `unavailable_reason`, and minimal collection state; it does not return original content.
- Feed/search cards include `owner_user_id` when the current App may need owner actions; App uses it before legacy `author_id`.

## 3. 记下

### 3.1 Create Experience

Endpoint:

```http
POST /api/v1/experiences
```

Auth:

- Required.

Request:

```json
{
  "content": "我刚才想明白的一点",
  "visibility": "public",
  "domain": "meaning",
  "sub_domain": "self",
  "topic": "#自我",
  "source_scene": "note",
  "source_chat_topic_id": null,
  "source_chat_message_id": null,
  "source_chat_message_snapshot": null,
  "source_message_ids": []
}
```

Validation:

- `content`: required, 1-100 runes.
- `visibility`: optional, defaults to `public`; must be `public` or `private`.
- `domain`: optional; if present must be one of the fixed V3 domain keys.
- `sub_domain`: optional; if present must belong to `domain`.
- `topic`: optional, max 200 runes.
- `source_scene`: optional; for this endpoint defaults to `note`.
- `source_chat_topic_id`: optional UUID; only used when a note is created from an existing stable chat topic.
- `source_chat_message_id`: optional UUID; when empty and `source_scene=chat`, backend derives it from the latest valid ID in `source_message_ids`.
- `source_chat_message_snapshot`: optional safe source snapshot or source-message-id list for temp-session traceability.
- `source_message_ids`: optional source chat message IDs; backend compacts empty values and preserves them in `source_chat_message_snapshot` when no explicit snapshot is provided.

Display-name gate:

- If `visibility=public` and user `display_name` is empty, return:

```json
{
  "error": {
    "code": "display_name_required",
    "message": "需要先设置展示名"
  }
}
```

- The experience is not created.
- If the user switches to `private`, no display name is required.

Success response:

```json
{
  "experience": {
    "id": "uuid",
    "content": "我刚才想明白的一点",
    "experience_type": "user_original",
    "visibility": "public",
    "lifecycle_status": "active",
    "domain": "meaning",
    "sub_domain": "self",
    "topic": "#自我",
    "creator_display_name": "用户展示名",
    "quality_tier": "unreviewed",
    "recommendation_status": "ineligible",
    "ai_citable": false
  }
}
```

Synchronous behavior:

- Save the experience immediately.
- Private saves complete synchronously.
- Public saves enqueue moderation/classification/quality processing and still return saved success.
- Frontend shows only `已记下`.

Asynchronous behavior:

- Public content moves through moderation, classify, quality tier, and optional interpretation.
- If public moderation says private-only, backend changes `visibility=private` and records internal reason; the user is not notified and sees the item in 看看-我的.
- If classification fails, the experience remains saved with empty domain/sub_domain until retry or manual repair.

### 3.2 Experience Detail

Endpoint:

```http
GET /api/v1/experiences/:id
```

Auth:

- Optional.
- Guests can read V4 `visibility=public` and `lifecycle_status=active` experiences.
- Logged-in users can read their own non-deleted experiences by `COALESCE(owner_user_id, author_id)=current_user`, including private and needs-review items.

Response behavior:

- Detail response exposes V4 ownership and display fields needed by the App:
  - `owner_user_id`
  - `creator_display_name`
  - `experience_type`
  - `visibility`
  - `lifecycle_status`
  - `quality_tier`
  - `inspiration_count`
  - `collection_count`
- App owner actions use `owner_user_id` first and fall back to `author_id` only for legacy records.
- App treats `404` detail responses as unavailable/deleted state, not as weak-network failure.
- For public own experiences, destructive delete must offer `转为私密` before `删除`.

### 3.3 Rewrite

Endpoint:

```http
POST /api/v1/experiences/rewrite
```

Request:

```json
{
  "content": "用户原始输入"
}
```

Behavior:

- Backend uses AI Gateway `function_type=experience_rewrite`.
- Output must be 1-100 runes.
- Return rewritten text only; do not save automatically.
- Failure does not block original save.

Response:

```json
{
  "rewritten_content": "更清楚的一条经验",
  "domain": "meaning",
  "sub_domain": "self",
  "topic": "#自我"
}
```

## 4. 看看

### 4.1 Recommend Feed

Endpoint:

```http
GET /api/v1/feed/recommend?cursor=&limit=20
```

Auth:

- Optional. Logged-out users receive cold-start public feed.
- App behavior: guest users can stay in 推荐 and search public experiences. Switching to 收藏 or 我的 opens the login gate instead of calling authenticated feed APIs.

Backend behavior:

1. Load or create `recommendation_session`.
2. Recall candidate pools.
3. Hard-filter public, active, `recommendation_status=eligible`, `quality_tier >= recommend_candidate`.
4. Score by V4 formula.
5. Apply creator/source/domain/sub_domain diversity.
6. Persist session order and next cursor.

Response:

```json
{
  "data": [],
  "next_cursor": "signed-or-session-cursor",
  "session_id": "uuid",
  "has_more": true
}
```

Fallbacks:

- If session persistence fails, return first page from immediate score but `next_cursor=""`.
- If metrics are missing, use neutral engagement score.
- If profile is missing, use cold-start mix.

### 4.2 Collections Feed

Endpoint:

```http
GET /api/v1/feed/collections?cursor=&limit=20&view=card
```

Auth:

- Required.
- App behavior: guest users are prompted to log in before opening this tab.

Behavior:

- Fact source is `experience_collections` with `status=active`.
- Visibility uses canonical V4 `visibility` and `lifecycle_status`: public active experiences are visible, owners can still see their own non-deleted collected experiences, and non-owner private/deleted rows return placeholders.
- Sort by `collected_at desc, id desc`.
- Unavailable experiences return placeholder cards and keep collection relation.

### 4.3 Mine Feed

Endpoint:

```http
GET /api/v1/feed/mine?cursor=&limit=20&view=card
```

Auth:

- Required.
- App behavior: guest users are prompted to log in before opening this tab.

Behavior:

- Query `experiences.owner_user_id=current_user`.
- Include public, private, and needs_review rows by canonical V4 lifecycle, excluding `lifecycle_status=deleted` and physically deleted rows.
- Sort by `created_at desc, id desc`.
- Private full-screen cards include only a light `仅自己可见` tag.

## 5. Interactions

Shared App behavior:

- Guest users see the login gate before any optimistic personal-state update is applied.
- If an optimistic 有启发 or 收藏 request fails, the App rolls the local card state back first.
- `401` failures use the unified expired-auth flow; non-auth request failures show `操作失败` with the backend/network message or `请稍后再试`.

### 5.1 有启发

Endpoint:

```http
POST /api/v1/experiences/:id/inspire
```

Auth:

- Required.
- App behavior: guest users are prompted to log in before the optimistic 有启发 update is applied.

Behavior:

- Insert into `experience_inspirations`.
- Write `experience_events` with `event_type=inspire`.
- Same user can inspire the same experience only once.
- Eligibility uses canonical V4 `visibility` and `lifecycle_status`: public active experiences are eligible, and owners can act on their own non-deleted experiences.
- Public experiences update public metric counters.
- Private experiences only affect personal history.

Responses:

- `200 {"inspired": true}`
- `409 {"inspired": true, "code": "already_inspired"}`
- `404` or `410` when not visible.

### 5.2 收藏

Endpoints:

```http
POST /api/v1/experiences/:id/collect
DELETE /api/v1/experiences/:id/collect
```

Behavior:

- Auth is required. App prompts guest users to log in before the optimistic 收藏/取消收藏 update is applied.
- Eligibility uses canonical V4 `visibility` and `lifecycle_status`: public active experiences are eligible, and owners can collect their own non-deleted experiences.
- Active collection lives in `experience_collections`.
- Removing collection sets `status=removed`; do not hard delete.
- Write `experience_events.collect` or `experience_events.uncollect`.

### 5.3 Passive Experience Events

Endpoint:

```http
POST /api/v1/experiences/:id/events
```

Auth:

- Optional.
- Guest events are allowed for public active experiences and write `user_id=null`.
- Logged-in users may record events for visible public experiences and their own private experiences.

Request:

```json
{
  "event_type": "search_click",
  "source_context": "search",
  "context_id": "uuid-or-empty",
  "metadata": {"query": "姜文", "rank": 0}
}
```

Behavior:

- Allowed passive event types are `expose`, `flip`, `search_click`, `chat_citation_show`, and `chat_citation_click`.
- Action events `collect`, `uncollect`, and `inspire` must use their dedicated authenticated endpoints and are rejected here.
- Invisible or deleted experiences return `404`.
- Eligibility uses canonical V4 `visibility` and `lifecycle_status`; fallback public/active predicates must not drive passive event eligibility.
- Response is `204 No Content`.

## 6. Search

Endpoint:

```http
GET /api/v1/search/experiences?q=&cursor=&limit=20
```

Auth:

- Optional.

Behavior:

- First stage uses rules and PostgreSQL search, not realtime AI.
- Search dimensions: content, creator_display_name, domain/sub_domain, topic, situation words.
- Public search returns public_visible and above.
- User-owned private search is allowed only in own scope when logged in.
- Public search eligibility uses V4 `visibility`, `lifecycle_status`, and `quality_tier` facts directly; legacy fallback predicates must not drive the public gate.
- Empty result does not show hot search or suggestions.
- App behavior: when a stored optional token makes search return `401`, treat it as expired auth, clear local auth, and route through the unified login gate instead of showing a weak-network search error.

## 7. 聊聊

### 7.1 Topic And Temp Session Endpoints

Endpoints:

```http
GET /api/v1/chat/recent-topics
GET /api/v1/chat/topics?cursor=&limit=20
POST /api/v1/chat/temp-sessions
POST /api/v1/chat/topics
PATCH /api/v1/chat/topics/:id
DELETE /api/v1/chat/topics/:id
```

Behavior:

- New chat starts as `chat_temp_sessions` unless recent active topic rules choose a stable topic.
- App entry resumes the most recent active stable topic only when `last_opened_at` / `updated_at` / `created_at` is within 2 hours; stale or timestampless topics fall back to a normal temp session.
- User-clicked `换个事聊` creates a temp session with `forced_new_topic=true` and must not auto-bind to an existing topic.
- First message in a temp session does not create a stable topic.
- `chat_topic_classify` can promote temp session to topic when `clarity_score >= 0.65`.
- Deleted topics stop contributing to AI context.

### 7.2 Send Message

Endpoints:

```http
POST /api/v1/chat/topics/:id/messages
POST /api/v1/chat/temp-sessions/:id/messages
```

Request:

```json
{
  "content": "最近有什么想不清的",
  "client_message_id": "client-generated-id"
}
```

Behavior:

- Save user message first.
- Build context from summary, recent messages, relevant user info, lightweight memory, and 3-5 candidate experiences.
- Call AI Gateway `function_type=chat`.
- Save assistant message and `chat_citations`.
- For temp-session messages, backend then calls AI Gateway `function_type=chat_topic_classify`.
- If `chat_topic_classify` returns `should_create_topic=true` and `clarity_score >= 0.65`, backend creates a stable `chat_topics` row, moves all messages from `temp_session_id` to `topic_id`, marks the temp session `promoted`, and returns `promoted_topic`.
- If topic classification fails or score is below threshold, chat reply still succeeds and the session remains `temp_session`.
- If AI fails, keep user message and return retryable error.
- Candidate experience retrieval is non-fatal; it must use currently deployed V4 fields only. Until content-production fields land, `source_derivation_type` is emitted as a compatible fallback rather than read from `experiences`.
- Candidate and historical reference-card visibility use canonical V4 `visibility` and `lifecycle_status`; public active experiences are visible, owner active experiences are usable as own context, and non-visible citations return placeholder cards instead of leaking content.

Response:

```json
{
  "user_message": {
    "id": "uuid",
    "role": "user",
    "content": "最近有什么想不清的",
    "topic_id": "uuid-if-promoted"
  },
  "message": {
    "id": "uuid",
    "role": "assistant",
    "content": "自然回复",
    "topic_id": "uuid-if-promoted"
  },
  "reference_cards": [
    {
      "experience_id": "uuid",
      "content": "参考经验正文",
      "is_collected": false
    }
  ],
  "note_suggestion": {
    "should_show": true,
    "suggested_text": "先做一小步，再用结果修正判断。",
    "source_message_ids": ["user-message-id", "assistant-message-id"]
  },
  "session_state": "temp_session | stable_topic",
  "promoted_topic": {
    "id": "uuid",
    "status": "active",
    "title": "工作里的不甘心",
    "domain": "work",
    "sub_domain": "work-comm",
    "topic": "和上级沟通",
    "clarity_score": 0.78
  }
}
```

App behavior:

- If `promoted_topic` is present, Chat switches from `tempSessionId` to `activeTopic`.
- Follow-up messages use `POST /api/v1/chat/topics/:id/messages`.
- Citation show/click events after promotion use `topic_id`, not the old `temp_session_id`.
- `429` quota errors display the backend-provided `message` when present; the App must not hard-code a numeric daily chat limit.
- Backend V4 chat quota is enforced before saving a new user message or calling AI. The limit is read from `system_config.chat_limit_per_day` with a defensive default of 50, and usage is counted from today's non-deleted V4 `chat_messages` rows where `role='user'`.

### 7.3 Save Chat Note Suggestion

First-phase App flow:

- Chat renders the `note_suggestion` card returned by the message endpoint.
- Tapping `记下` opens the App `记下` editor rather than saving automatically.
- The final save reuses `POST /api/v1/experiences` so the user can still edit text, keep private, or choose anonymous public contribution.

Required create payload fields for chat-sourced saves:

```json
{
  "content": "先做一小步，再用结果修正判断。",
  "visibility": "private",
  "source_scene": "chat",
  "source_message_ids": ["user-message-id", "assistant-message-id"],
  "source_chat_topic_id": "stable-topic-id-if-any",
  "source_chat_message_id": "source-message-id",
  "source_chat_message_snapshot": "safe-source-message-id-list-or-summary"
}
```

Behavior:

- `source_chat_topic_id` is sent only when the App is saving from a stable topic.
- `source_chat_message_id` points at the assistant message that produced the note suggestion.
- `source_chat_message_snapshot` stores source identifiers or a safe short snapshot for temp-session traceability.
- Public anonymous contribution must not expose chat title, raw chat context, or private source messages in public experience fields.

## 8. 我的

### 8.1 Profile

Endpoints:

```http
GET /api/v1/me/profile
PATCH /api/v1/me/profile
```

Auth:

- Required.

Behavior:

- `display_name` is the user-facing creator name for user-original experiences.
- `display_name` accepts up to 30 characters; App entry points that set it, including the first-public-note gate, must mirror this limit.
- Patch can update `display_name`, `free_description`, and lightweight profile fields.
- When `display_name` changes, backend syncs `creator_display_name` on the user's non-deleted original experiences.

### 8.2 Stats

Endpoints:

```http
GET /api/v1/me/stats/assets
GET /api/v1/me/stats/contribution
GET /api/v1/me/stats/change
GET /api/v1/me/stats/recent-harvest?range=7d|30d|all
GET /api/v1/me/recent-responded-experiences?limit=3
```

Behavior:

- `assets` returns long-term accumulation: own experiences, active collections, public/private split, and note/chat source split.
- `contribution` returns only public user-original contribution feedback; platform-selected experiences are excluded.
- `change` returns chat topics, clearer count, and chat-derived experiences.
- `recent-harvest` supports `7d`, `30d`, and `all`; it returns note-created experiences, chat-derived experiences, new inspired users, and new collections in the selected range.
- `recent-responded-experiences` returns public user-original experiences with real response data. Cards include content, star rating, domain/sub_domain, inspiration count, collection count, and latest response time.
- App-facing stats use V4 `visibility` and `lifecycle_status` facts for public/private, contribution, recent-harvest, and recent-responded eligibility; legacy `is_private` and fallback lifecycle predicates must not be used as runtime gates.
- Stats failures must not be shown as zero in the App.

### 8.3 Account Actions

Endpoints:

```http
DELETE /api/v1/me/account
```

Auth:

- Required.

Behavior:

- App uses the V4 `/me` route for account cancellation from 我的.
- The legacy `/user/account` route may remain temporarily for compatibility, but new App code must not call it.
- Account deletion/anonymization policy still needs the production data-retention decision before final deployment.

## 9. AI Gateway Foundation

First stage tables:

- `ai_function_configs`
- `ai_prompt_registry`
- `ai_call_logs`
- `ai_jobs`

Rules:

- Handlers pass `function_type` and payload.
- Gateway resolves `key_alias`, prompt version, schema version, timeout, retry, and budget.
- DeepSeek calls are never made directly by mobile.
- Existing AI service calls may be retained only as transitional worker implementation behind Gateway-compatible functions.

## 10. Phase 1 Acceptance

- Backend has V4 canonical schema fields and event tables.
- Backend model layer has V4 enums, validation helpers, and eligibility helpers.
- API contracts above are reflected in implementation tasks.
- Old user-facing semantics are not used for new endpoints.
- Backend tests and Linux cross-build pass.
