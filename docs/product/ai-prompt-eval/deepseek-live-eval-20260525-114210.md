# DeepSeek Live Prompt Eval

- generated_at: `2026-05-25T11:42:10`
- model: `deepseek-v4-pro`
- base_url: `https://api.deepseek.com`
- spec: `docs/product/niangao-ai-prompt-production-spec-v4.md`
- result: `7/8` passed

## Cases

| case | category | function_type | status | latency_ms | errors |
| --- | --- | --- | --- | ---: | --- |
| `chat_001` | `chat` | `chat` | `passed` | 3501 |  |
| `privacy_001` | `privacy_summary` | `chat_summary` | `passed` | 6264 |  |
| `classify_001` | `classification` | `experience_classify` | `passed` | 3033 |  |
| `content_016` | `content_production` | `experience_extract` | `failed` | 14374 | rule_all_items_max_len_failed:result.candidates.candidate_content |
| `content_086` | `content_production` | `experience_interpretation` | `passed` | 9062 |  |
| `content_056` | `content_production` | `experience_review` | `passed` | 15111 |  |
| `recommend_001` | `recommendation` | `recommendation_ai` | `passed` | 9416 |  |
| `content_001` | `content_production` | `translation_normalization` | `passed` | 6374 |  |

## Failed Output Previews

### content_016

Errors:

- `rule_all_items_max_len_failed:result.candidates.candidate_content`

Output preview:

```json
{"schema_version":"1.1","function_type":"experience_extract","result":{"candidates":[{"candidate_content":"If you want to make something people want, talk to users before you spend months building in your head.","creator_name":"Paul Graham","creator_attribution_type":"speaker","source_derivation_type":"expressed_principle","source_excerpt":"If you want to make something people want, talk to users before you spend months building in your head.","source_location":"Paul Graham素材","preserve_original_score":1.0,"extraction_confidence":0.95,"attitude_type":"practical","risk_notes":[]}],"discarded_examples":[]},"confidence":0.95,"warnings":[]}
```
