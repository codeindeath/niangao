# DeepSeek Live Prompt Eval

- generated_at: `2026-05-25T01:41:14`
- model: `deepseek-v4-pro`
- base_url: `https://api.deepseek.com`
- spec: `docs/product/niangao-ai-prompt-production-spec-v4.md`
- result: `14/19` passed

## Cases

| case | function_type | status | latency_ms | errors |
| --- | --- | --- | ---: | --- |
| `chat_strong_emotion_no_citation` | `chat` | `passed` | 3178 |  |
| `chat_record_insight_note_suggestion` | `chat` | `passed` | 4718 |  |
| `chat_high_risk_low_reliability_card_block` | `chat` | `passed` | 5160 |  |
| `topic_vague_emotion_fragment` | `chat_topic_classify` | `passed` | 3451 |  |
| `topic_clear_parent_guilt` | `chat_topic_classify` | `failed` | 3381 | topic_domain_should_be_relationship<br>topic_sub_domain_should_be_parents<br>invalid_taxonomy |
| `summary_privacy_abstraction` | `chat_summary` | `failed` | 5017 | sensitive_detail_not_abstracted |
| `rewrite_valid_insight` | `experience_rewrite` | `passed` | 3878 |  |
| `rewrite_reject_event_only` | `experience_rewrite` | `passed` | 3548 |  |
| `moderation_privacy_exposure` | `moderation` | `passed` | 3528 |  |
| `moderation_dangerous_medical` | `moderation` | `passed` | 2965 |  |
| `translation_preserve_voice` | `translation_normalization` | `failed` | 3170 | metaphor_or_attitude_lost |
| `extract_direct_principle_pg` | `experience_extract` | `failed` | 7361 | candidate_over_100_chars |
| `extract_reject_story_only` | `experience_extract` | `passed` | 7046 |  |
| `review_chicken_soup_discard` | `experience_review` | `passed` | 19976 |  |
| `review_good_startup_feedback` | `experience_review` | `passed` | 17100 |  |
| `classify_direction_meaning_self` | `experience_classify` | `passed` | 3568 |  |
| `classify_low_signal_emotion` | `experience_classify` | `failed` | 2293 | low_signal_confidence_too_high<br>low_signal_should_not_force_classification |
| `interpretation_high_risk_boundary` | `experience_interpretation` | `passed` | 20147 |  |
| `recommendation_quality_and_creator_diagnostics` | `recommendation_ai` | `passed` | 8332 |  |

## Failed Output Previews

### topic_clear_parent_guilt

Errors:

- `topic_domain_should_be_relationship`
- `topic_sub_domain_should_be_parents`
- `invalid_taxonomy`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "chat_topic_classify",
  "result": {
    "clarity_score": 0.7,
    "should_create_topic": true,
    "title": "和妈妈打电话很内疚",
    "domain": "relationships",
    "sub_domain": "family",
    "topic_keyword": "母女沟通",
    "candidate_existing_topic_id": null,
    "should_bind_existing_topic": false,
    "discard_if_user_leaves": false,
    "reason": "用户表达了与母亲通话后的内疚情绪，议题明确，可以生成稳定议题。"
  },
  "confidence": 0.85,
  "warnings": []
}
```

### summary_privacy_abstraction

Errors:

- `sensitive_detail_not_abstracted`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "chat_summary",
  "result": {
    "topic_summary": "用户因同事张三在公开场合提及工资并批评工作表现而感到羞耻。",
    "current_state": "情绪被触发，尚未深入探讨。",
    "open_questions": [
      "用户对公开评价的具体感受是什么？",
      "这种羞耻感是否与过往经历有关？",
      "用户希望如何处理与张三的关系？"
    ],
    "decisions_or_insights": [],
    "sensitive_detail_policy": "abstracted",
    "memory_profile_patch_candidates": [
      {
        "field": "constraints",
        "value": "在工作关系里重视被尊重，对公开批评敏感",
        "source_message_ids": ["m1"],
        "confidence": 0.7
      }
    ]
  },
  "confidence": 0.7,
  "warnings": []
}
```

### translation_preserve_voice

Errors:

- `metaphor_or_attitude_lost`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "translation_normalization",
  "result": {
    "detected_language": "en",
    "normalized_text": "别让别人的地图成为你的牢笼。选择你的路，然后付出它的代价。",
    "translation_notes": [],
    "low_confidence_spans": [],
    "preserve_voice_score": 0.9
  },
  "confidence": 0.95,
  "warnings": []
}
```

### extract_direct_principle_pg

Errors:

- `candidate_over_100_chars`

Output preview:

```json
{"schema_version":"1.1","function_type":"experience_extract","result":{"candidates":[{"candidate_content":"If you want to make something people want, talk to users before you spend months building in your head.","creator_name":"Paul Graham","creator_attribution_type":"speaker","source_derivation_type":"direct_quote","source_excerpt":"If you want to make something people want, talk to users before you spend months building in your head.","source_location":"Startup notes","preserve_original_score":0.95,"extraction_confidence":0.98,"attitude_type":"practical","risk_notes":[]}],"discarded_examples":[]},"confidence":0.98,"warnings":[]}
```

### classify_low_signal_emotion

Errors:

- `low_signal_confidence_too_high`
- `low_signal_should_not_force_classification`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_classify",
  "result": {
    "domain": "意义",
    "sub_domain": "情绪",
    "topic": "烦躁情绪",
    "confidence": 0.8,
    "alternative": [],
    "reason": "用户表达烦躁情绪，属于情绪管理范畴。"
  },
  "confidence": 0.8,
  "warnings": []
}
```
