# DeepSeek Live Prompt Eval

- generated_at: `2026-05-25T01:38:45`
- model: `deepseek-v4-pro`
- base_url: `https://api.deepseek.com`
- spec: `docs/product/niangao-ai-prompt-production-spec-v4.md`
- result: `2/3` passed

## Cases

| case | function_type | status | latency_ms | errors |
| --- | --- | --- | ---: | --- |
| `chat_strong_emotion_no_citation` | `chat` | `passed` | 2850 |  |
| `topic_clear_parent_guilt` | `chat_topic_classify` | `failed` | 3785 | clear_topic_should_create<br>topic_title_too_short<br>topic_domain_should_be_relationship<br>topic_sub_domain_should_be_parents |
| `review_chicken_soup_discard` | `experience_review` | `passed` | 11061 |  |

## Failed Output Previews

### topic_clear_parent_guilt

Errors:

- `clear_topic_should_create`
- `topic_title_too_short`
- `topic_domain_should_be_relationship`
- `topic_sub_domain_should_be_parents`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "chat_topic_classify",
  "result": {
    "clarity_score": 0.55,
    "should_create_topic": false,
    "title": null,
    "domain": null,
    "sub_domain": null,
    "topic_keyword": null,
    "candidate_existing_topic_id": null,
    "should_bind_existing_topic": false,
    "discard_if_user_leaves": false,
    "reason": "用户表达了与母亲通话后的内疚情绪，但信息尚不足以形成稳定议题，缺乏具体情境和问题线索。"
  },
  "confidence": 0.7,
  "warnings": []
}
```
