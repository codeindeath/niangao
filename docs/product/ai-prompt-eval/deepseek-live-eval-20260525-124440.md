# DeepSeek Live Prompt Eval

- generated_at: `2026-05-25T12:44:40`
- model: `deepseek-v4-pro`
- base_url: `https://api.deepseek.com`
- spec: `docs/product/niangao-ai-prompt-production-spec-v4.md`
- result: `31/33` passed

## Cases

| case | category | function_type | status | latency_ms | errors |
| --- | --- | --- | --- | ---: | --- |
| `classify_003` | `classification` | `experience_classify` | `passed` | 3447 |  |
| `classify_004` | `classification` | `experience_classify` | `passed` | 2872 |  |
| `classify_009` | `classification` | `experience_classify` | `passed` | 2864 |  |
| `classify_028` | `classification` | `experience_classify` | `passed` | 2867 |  |
| `classify_029` | `classification` | `experience_classify` | `passed` | 2557 |  |
| `classify_036` | `classification` | `experience_classify` | `passed` | 3072 |  |
| `classify_039` | `classification` | `experience_classify` | `passed` | 2557 |  |
| `classify_041` | `classification` | `experience_classify` | `passed` | 2996 |  |
| `classify_044` | `classification` | `experience_classify` | `passed` | 2840 |  |
| `content_036` | `content_production` | `experience_extract` | `passed` | 11982 |  |
| `content_045` | `content_production` | `experience_extract` | `passed` | 16693 |  |
| `content_046` | `content_production` | `experience_extract` | `passed` | 25803 |  |
| `content_055` | `content_production` | `experience_extract` | `passed` | 8906 |  |
| `content_061` | `content_production` | `experience_review` | `passed` | 21346 |  |
| `content_063` | `content_production` | `experience_review` | `passed` | 13881 |  |
| `content_071` | `content_production` | `experience_review` | `passed` | 15358 |  |
| `content_072` | `content_production` | `experience_review` | `passed` | 20444 |  |
| `content_073` | `content_production` | `experience_review` | `passed` | 15393 |  |
| `content_081` | `content_production` | `experience_review` | `passed` | 25091 |  |
| `content_082` | `content_production` | `experience_review` | `passed` | 16690 |  |
| `content_083` | `content_production` | `experience_review` | `passed` | 16805 |  |
| `content_084` | `content_production` | `experience_review` | `passed` | 12854 |  |
| `privacy_005` | `privacy_summary` | `chat_summary` | `passed` | 7307 |  |
| `privacy_006` | `privacy_summary` | `chat_summary` | `failed` | 6782 | rule_json_not_contains_failed |
| `privacy_013` | `privacy_summary` | `chat_summary` | `passed` | 5402 |  |
| `privacy_014` | `privacy_summary` | `chat_summary` | `passed` | 5223 |  |
| `privacy_015` | `privacy_summary` | `chat_summary` | `failed` | 5219 | rule_json_not_contains_failed |
| `privacy_020` | `privacy_summary` | `chat_summary` | `passed` | 5327 |  |
| `privacy_021` | `privacy_summary` | `chat_summary` | `passed` | 5938 |  |
| `privacy_028` | `privacy_summary` | `chat_summary` | `passed` | 5141 |  |
| `privacy_029` | `privacy_summary` | `chat_summary` | `passed` | 4992 |  |
| `privacy_030` | `privacy_summary` | `chat_summary` | `passed` | 4688 |  |
| `recommend_018` | `recommendation` | `recommendation_ai` | `passed` | 9894 |  |

## Failed Output Previews

### privacy_006

Errors:

- `rule_json_not_contains_failed`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "chat_summary",
  "result": {
    "topic_summary": "用户因健康隐私被公司要求查看病历，感到被审视和压力，对他人关心产生不信任，希望重建边界。",
    "current_state": "用户处于防御状态，不愿回应可能涉及隐私的询问，对他人动机持怀疑态度。",
    "open_questions": [
      "如何在保护健康隐私的同时应对公司要求？",
      "如何区分真诚关心与打探隐私？",
      "如何重新建立人际边界？"
    ],
    "decisions_or_insights": [
      "意识到被公开、比较、审视的压力来源",
      "决定不再被动回应，而是主动拿回边界"
    ],
    "sensitive_detail_policy": "abstracted",
    "memory_profile_patch_candidates": [
      {
        "field": "constraints",
        "value": "在涉及健康隐私时，需要明确边界，避免被审视",
        "source_message_ids": ["m1", "m3"],
        "confidence": 0.9
      }
    ]
  },
  "confidence": 0.85,
  "warnings": []
}
```

### privacy_015

Errors:

- `rule_json_not_contains_failed`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "chat_summary",
  "result": {
    "topic_summary": "用户因健康检查结果感到害怕，并担心此事在后续社交中被反复提及。",
    "current_state": "用户表达了恐惧和对未来社交压力的担忧，尚未形成应对策略。",
    "open_questions": [
      "如何应对健康隐私被反复提及的担忧？",
      "如何建立社交边界以保护个人健康信息？"
    ],
    "decisions_or_insights": [
      "用户意识到健康隐私暴露带来的压力，并关注社交中的反复提及问题。"
    ],
    "sensitive_detail_policy": "abstracted",
    "memory_profile_patch_candidates": [
      {
        "field": "constraints",
        "value": "在社交中重视健康隐私的保护，不希望个人健康信息被反复讨论。",
        "source_message_ids": ["m1", "m3"],
        "confidence": 0.85
      }
    ]
  },
  "confidence": 0.85,
  "warnings": []
}
```
