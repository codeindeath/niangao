# DeepSeek Live Prompt Eval

- generated_at: `2026-05-25T09:36:39`
- model: `deepseek-v4-pro`
- base_url: `https://api.deepseek.com`
- spec: `docs/product/niangao-ai-prompt-production-spec-v4.md`
- result: `19/19` passed

## Cases

| case | category | function_type | status | latency_ms | errors |
| --- | --- | --- | --- | ---: | --- |
| `chat_strong_emotion_no_citation` | `baseline` | `chat` | `passed` | 3066 |  |
| `chat_record_insight_note_suggestion` | `baseline` | `chat` | `passed` | 4120 |  |
| `chat_high_risk_low_reliability_card_block` | `baseline` | `chat` | `passed` | 5402 |  |
| `topic_vague_emotion_fragment` | `baseline` | `chat_topic_classify` | `passed` | 2973 |  |
| `topic_clear_parent_guilt` | `baseline` | `chat_topic_classify` | `passed` | 3376 |  |
| `summary_privacy_abstraction` | `baseline` | `chat_summary` | `passed` | 4415 |  |
| `rewrite_valid_insight` | `baseline` | `experience_rewrite` | `passed` | 3689 |  |
| `rewrite_reject_event_only` | `baseline` | `experience_rewrite` | `passed` | 2951 |  |
| `moderation_privacy_exposure` | `baseline` | `moderation` | `passed` | 3278 |  |
| `moderation_dangerous_medical` | `baseline` | `moderation` | `passed` | 3071 |  |
| `translation_preserve_voice` | `baseline` | `translation_normalization` | `passed` | 2149 |  |
| `extract_direct_principle_pg` | `baseline` | `experience_extract` | `passed` | 20277 |  |
| `extract_reject_story_only` | `baseline` | `experience_extract` | `passed` | 5116 |  |
| `review_chicken_soup_discard` | `baseline` | `experience_review` | `passed` | 17223 |  |
| `review_good_startup_feedback` | `baseline` | `experience_review` | `passed` | 23329 |  |
| `classify_direction_meaning_self` | `baseline` | `experience_classify` | `passed` | 2888 |  |
| `classify_low_signal_emotion` | `baseline` | `experience_classify` | `passed` | 2448 |  |
| `interpretation_high_risk_boundary` | `baseline` | `experience_interpretation` | `passed` | 6638 |  |
| `recommendation_quality_and_creator_diagnostics` | `baseline` | `recommendation_ai` | `passed` | 7068 |  |

## Failed Output Previews
