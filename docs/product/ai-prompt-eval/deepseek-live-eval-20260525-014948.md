# DeepSeek Live Prompt Eval

- generated_at: `2026-05-25T01:49:48`
- model: `deepseek-v4-pro`
- base_url: `https://api.deepseek.com`
- spec: `docs/product/niangao-ai-prompt-production-spec-v4.md`
- result: `4/5` passed

## Cases

| case | function_type | status | latency_ms | errors |
| --- | --- | --- | ---: | --- |
| `topic_clear_parent_guilt` | `chat_topic_classify` | `passed` | 3856 |  |
| `summary_privacy_abstraction` | `chat_summary` | `passed` | 4193 |  |
| `translation_preserve_voice` | `translation_normalization` | `passed` | 2970 |  |
| `extract_direct_principle_pg` | `experience_extract` | `passed` | 10036 |  |
| `classify_low_signal_emotion` | `experience_classify` | `failed` | 2652 | low_signal_confidence_too_high |

## Failed Output Previews

### classify_low_signal_emotion

Errors:

- `low_signal_confidence_too_high`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_classify",
  "result": {
    "domain": null,
    "sub_domain": null,
    "topic": null,
    "confidence": 0.0,
    "alternative": [],
    "reason": "输入为情绪表达，无复用性判断，不构成经验。"
  },
  "confidence": 0.0,
  "warnings": []
}
```
