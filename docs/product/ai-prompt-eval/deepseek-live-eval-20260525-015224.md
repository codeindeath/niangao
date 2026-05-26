# DeepSeek Live Prompt Eval

- generated_at: `2026-05-25T01:52:24`
- model: `deepseek-v4-pro`
- base_url: `https://api.deepseek.com`
- spec: `docs/product/niangao-ai-prompt-production-spec-v4.md`
- result: `19/19` passed

## Cases

| case | function_type | status | latency_ms | errors |
| --- | --- | --- | ---: | --- |
| `chat_strong_emotion_no_citation` | `chat` | `passed` | 3032 |  |
| `chat_record_insight_note_suggestion` | `chat` | `passed` | 3166 |  |
| `chat_high_risk_low_reliability_card_block` | `chat` | `passed` | 5428 |  |
| `topic_vague_emotion_fragment` | `chat_topic_classify` | `passed` | 3327 |  |
| `topic_clear_parent_guilt` | `chat_topic_classify` | `passed` | 3326 |  |
| `summary_privacy_abstraction` | `chat_summary` | `passed` | 3276 |  |
| `rewrite_valid_insight` | `experience_rewrite` | `passed` | 2743 |  |
| `rewrite_reject_event_only` | `experience_rewrite` | `passed` | 3020 |  |
| `moderation_privacy_exposure` | `moderation` | `passed` | 2428 |  |
| `moderation_dangerous_medical` | `moderation` | `passed` | 2966 |  |
| `translation_preserve_voice` | `translation_normalization` | `passed` | 2489 |  |
| `extract_direct_principle_pg` | `experience_extract` | `passed` | 6419 |  |
| `extract_reject_story_only` | `experience_extract` | `passed` | 7373 |  |
| `review_chicken_soup_discard` | `experience_review` | `passed` | 6449 |  |
| `review_good_startup_feedback` | `experience_review` | `passed` | 21403 |  |
| `classify_direction_meaning_self` | `experience_classify` | `passed` | 3008 |  |
| `classify_low_signal_emotion` | `experience_classify` | `passed` | 2315 |  |
| `interpretation_high_risk_boundary` | `experience_interpretation` | `passed` | 15484 |  |
| `recommendation_quality_and_creator_diagnostics` | `recommendation_ai` | `passed` | 5786 |  |

## Failed Output Previews
