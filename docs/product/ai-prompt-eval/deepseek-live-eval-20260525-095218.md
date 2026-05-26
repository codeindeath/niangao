# DeepSeek Live Prompt Eval

- generated_at: `2026-05-25T09:52:18`
- model: `deepseek-v4-pro`
- base_url: `https://api.deepseek.com`
- spec: `docs/product/niangao-ai-prompt-production-spec-v4.md`
- result: `15/16` passed

## Cases

| case | category | function_type | status | latency_ms | errors |
| --- | --- | --- | --- | ---: | --- |
| `chat_001` | `chat` | `chat` | `passed` | 3121 |  |
| `chat_002` | `chat` | `chat` | `passed` | 3726 |  |
| `privacy_001` | `privacy_summary` | `chat_summary` | `passed` | 5878 |  |
| `privacy_002` | `privacy_summary` | `chat_summary` | `passed` | 6922 |  |
| `classify_001` | `classification` | `experience_classify` | `passed` | 3060 |  |
| `classify_002` | `classification` | `experience_classify` | `passed` | 2988 |  |
| `content_016` | `content_production` | `experience_extract` | `passed` | 25391 |  |
| `content_017` | `content_production` | `experience_extract` | `passed` | 21794 |  |
| `content_086` | `content_production` | `experience_interpretation` | `passed` | 8123 |  |
| `content_087` | `content_production` | `experience_interpretation` | `passed` | 8067 |  |
| `content_056` | `content_production` | `experience_review` | `passed` | 25089 |  |
| `content_057` | `content_production` | `experience_review` | `failed` | 17304 | rule_in_failed:result.decision |
| `recommend_001` | `recommendation` | `recommendation_ai` | `passed` | 7371 |  |
| `recommend_002` | `recommendation` | `recommendation_ai` | `passed` | 9083 |  |
| `content_001` | `content_production` | `translation_normalization` | `passed` | 3768 |  |
| `content_002` | `content_production` | `translation_normalization` | `passed` | 4382 |  |

## Failed Output Previews

### content_057

Errors:

- `rule_in_failed:result.decision`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_review",
  "result": {
    "decision": "auto_import",
    "delete_category": null,
    "cleaned_content": "创业早期不要用想象替代用户反馈。",
    "cleaning_level": "none",
    "ai_quality_score": 7.0,
    "score_breakdown": {
      "clarity": 9,
      "transferability": 8,
      "judgment": 7,
      "boundary": 6,
      "non_obviousness": 5,
      "attitude": 7,
      "usefulness": 8
    },
    "quality_tier": "recommend_candidate",
    "ai_citable": false,
    "recommendation_eligible": true,
    "misuse_risk_level": "low",
    "misuse_risk_notes": [],
    "review_reason": "清晰的原则，可迁移性强，有明确判断和指导价值。",
    "needs_human_attention": false
  },
  "confidence": 0.95,
  "warnings": []
}
```
