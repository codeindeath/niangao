# DeepSeek Live Prompt Eval

- generated_at: `2026-05-25T01:37:42`
- model: `deepseek-v4-pro`
- base_url: `https://api.deepseek.com`
- spec: `docs/product/niangao-ai-prompt-production-spec-v4.md`
- result: `0/19` passed

## Cases

| case | function_type | status | latency_ms | errors |
| --- | --- | --- | ---: | --- |
| `chat_strong_emotion_no_citation` | `chat` | `failed` | 2240 | schema_version_not_1.1<br>function_type_mismatch<br>result_not_object |
| `chat_record_insight_note_suggestion` | `chat` | `failed` | 3371 | schema_version_not_1.1<br>function_type_mismatch<br>result_not_object |
| `chat_high_risk_low_reliability_card_block` | `chat` | `failed` | 3993 | schema_version_not_1.1<br>function_type_mismatch<br>result_not_object |
| `topic_vague_emotion_fragment` | `chat_topic_classify` | `failed` | 1842 | schema_version_not_1.1<br>function_type_mismatch<br>result_not_object |
| `topic_clear_parent_guilt` | `chat_topic_classify` | `failed` | 1954 | schema_version_not_1.1<br>function_type_mismatch<br>result_not_object |
| `summary_privacy_abstraction` | `chat_summary` | `failed` | 2961 | schema_version_not_1.1<br>function_type_mismatch<br>result_not_object |
| `rewrite_valid_insight` | `experience_rewrite` | `failed` | 1632 | schema_version_not_1.1<br>function_type_mismatch<br>result_not_object |
| `rewrite_reject_event_only` | `experience_rewrite` | `failed` | 1433 | schema_version_not_1.1<br>function_type_mismatch<br>result_not_object |
| `moderation_privacy_exposure` | `moderation` | `failed` | 1322 | schema_version_not_1.1<br>function_type_mismatch<br>result_not_object |
| `moderation_dangerous_medical` | `moderation` | `failed` | 1338 | schema_version_not_1.1<br>function_type_mismatch<br>result_not_object |
| `translation_preserve_voice` | `translation_normalization` | `failed` | 2153 | schema_version_not_1.1<br>function_type_mismatch<br>result_not_object |
| `extract_direct_principle_pg` | `experience_extract` | `failed` | 7884 | schema_version_not_1.1<br>function_type_mismatch<br>result_not_object |
| `extract_reject_story_only` | `experience_extract` | `failed` | 3678 | schema_version_not_1.1<br>function_type_mismatch<br>result_not_object |
| `review_chicken_soup_discard` | `experience_review` | `failed` | 17166 | schema_version_not_1.1<br>function_type_mismatch<br>result_not_object |
| `review_good_startup_feedback` | `experience_review` | `failed` | 23803 | schema_version_not_1.1<br>function_type_mismatch<br>result_not_object |
| `classify_direction_meaning_self` | `experience_classify` | `failed` | 1326 | schema_version_not_1.1<br>function_type_mismatch<br>result_not_object |
| `classify_low_signal_emotion` | `experience_classify` | `failed` | 1702 | schema_version_not_1.1<br>function_type_mismatch<br>result_not_object |
| `interpretation_high_risk_boundary` | `experience_interpretation` | `failed` | 18468 | schema_version_not_1.1<br>function_type_mismatch<br>result_not_object |
| `recommendation_quality_and_creator_diagnostics` | `recommendation_ai` | `failed` | 10141 | schema_version_not_1.1<br>function_type_mismatch<br>result_not_object |

## Failed Output Previews

### chat_strong_emotion_no_citation

Errors:

- `schema_version_not_1.1`
- `function_type_mismatch`
- `result_not_object`

Output preview:

```json
{
  "reply_text": "嗯，听起来你现在特别烦，不想被任何人指手画脚。那就先不解决问题，我在这儿陪你待一会儿。",
  "note_suggestion": null,
  "citation_cards": []
}
```

### chat_record_insight_note_suggestion

Errors:

- `schema_version_not_1.1`
- `function_type_mismatch`
- `result_not_object`

Output preview:

```json
{
  "reply_text": "这句话很关键——你把“怕换工作”重新看清楚了。怕的不是变化本身，是怕再经历一次“选错”的感觉。",
  "note_suggestion": "这点要不要记下？",
  "citation_cards": []
}
```

### chat_high_risk_low_reliability_card_block

Errors:

- `schema_version_not_1.1`
- `function_type_mismatch`
- `result_not_object`

Output preview:

```json
{
  "reply_text": "我听到你话里那股“不想再忍了”的劲儿，好像今晚不切断，明天就又会心软。这种冲动本身不是错，它是在替你喊停。",
  "note_suggestion": null,
  "citation_cards": [
    {
      "experience_id": "exp_high_boundary",
      "display_text": "重大决定先分清触发情绪和长期底线，再做可逆的小动作。",
      "source_label": "精选经验"
    }
  ]
}
```

### topic_vague_emotion_fragment

Errors:

- `schema_version_not_1.1`
- `function_type_mismatch`
- `result_not_object`

Output preview:

```json
{"topic_title":null,"domain":null,"sub_domain":null,"clarity_score":0.1,"should_bind_existing_topic":false,"existing_topic_id":null}
```

### topic_clear_parent_guilt

Errors:

- `schema_version_not_1.1`
- `function_type_mismatch`
- `result_not_object`

Output preview:

```json
{
  "should_create_topic": true,
  "clarity_score": 0.75,
  "title": "和妈妈打电话很内疚",
  "domain": "relationships",
  "sub_domain": "family",
  "bind_existing_topic_id": null
}
```

### summary_privacy_abstraction

Errors:

- `schema_version_not_1.1`
- `function_type_mismatch`
- `result_not_object`

Output preview:

```json
{
  "summary": "用户提到张三在公司会上公开评价其工资五万且工作表现不佳，感到羞耻。年糕回应指出这是被公开评价刺到。",
  "light_profile_candidates": [
    {
      "field": "common_issue_domains",
      "value": "在工作关系里重视被尊重",
      "source_message_ids": ["m1"]
    }
  ]
}
```

### rewrite_valid_insight

Errors:

- `schema_version_not_1.1`
- `function_type_mismatch`
- `result_not_object`

Output preview:

```json
{
  "experience": "找朋友帮忙要提前说清楚边界，不然最后两个人都不舒服。",
  "rewrite_level": "light"
}
```

### rewrite_reject_event_only

Errors:

- `schema_version_not_1.1`
- `function_type_mismatch`
- `result_not_object`

Output preview:

```json
{
  "can_rewrite": false,
  "rewrite_level": "reject",
  "experience": ""
}
```

### moderation_privacy_exposure

Errors:

- `schema_version_not_1.1`
- `function_type_mismatch`
- `result_not_object`

Output preview:

```json
{
  "public_action": "block_public",
  "ai_citation_allowed": false
}
```

### moderation_dangerous_medical

Errors:

- `schema_version_not_1.1`
- `function_type_mismatch`
- `result_not_object`

Output preview:

```json
{
  "public_action": "block_public",
  "ai_citation_allowed": false
}
```

### translation_preserve_voice

Errors:

- `schema_version_not_1.1`
- `function_type_mismatch`
- `result_not_object`

Output preview:

```json
{
  "source_material": "00:01 Don't let someone else's map become your prison. 00:04 Pick your road, then pay its price.",
  "source_language": "en",
  "normalized_text": "别让他人的地图成为你的牢笼。选择你的路，然后付出它的代价。",
  "low_confidence_spans": []
}
```

### extract_direct_principle_pg

Errors:

- `schema_version_not_1.1`
- `function_type_mismatch`
- `result_not_object`

Output preview:

```json
{
  "candidates": [
    {
      "content": "If you want to make something people want, talk to users before you spend months building in your head.",
      "creator_name": "Paul Graham",
      "source_derivation_type": "direct_statement",
      "source_excerpt": "If you want to make something people want, talk to users before you spend months building in your head.",
      "risk_notes": null
    }
  ]
}
```

### extract_reject_story_only

Errors:

- `schema_version_not_1.1`
- `function_type_mismatch`
- `result_not_object`

Output preview:

```json
{"candidates": []}
```

### review_chicken_soup_discard

Errors:

- `schema_version_not_1.1`
- `function_type_mismatch`
- `result_not_object`

Output preview:

```json
{"decision":"delete","reason":"chicken_soup","quality_score":2}
```

### review_good_startup_feedback

Errors:

- `schema_version_not_1.1`
- `function_type_mismatch`
- `result_not_object`

Output preview:

```json
{"deletion_decision": false, "deletion_reason": null, "quality_score": 6, "quality_dimensions": {"clarity": 8, "transferability": 7, "judgment": 7, "boundary": 4, "non_obviousness": 5, "attitude": 6, "usefulness": 7}}
```

### classify_direction_meaning_self

Errors:

- `schema_version_not_1.1`
- `function_type_mismatch`
- `result_not_object`

Output preview:

```json
{
  "domain": "意义",
  "sub_domain": "自我",
  "topic": "自我节奏",
  "confidence": 0.9
}
```

### classify_low_signal_emotion

Errors:

- `schema_version_not_1.1`
- `function_type_mismatch`
- `result_not_object`

Output preview:

```json
{
  "domain": "意义",
  "sub_domain": "情绪",
  "topic": "烦躁情绪",
  "confidence": 0.9
}
```

### interpretation_high_risk_boundary

Errors:

- `schema_version_not_1.1`
- `function_type_mismatch`
- `result_not_object`

Output preview:

```json
{
  "interpretation": {
    "sections": [
      {
        "heading": "适用于哪种创业处境",
        "body": "当项目需要全力冲刺、快速验证商业模式，且个人能承受短期无收入风险时，all in 心态有助于集中资源突破瓶颈。它更适用于早期单点突破阶段，而非成熟扩张期。"
      },
      {
        "heading": "“别留退路”的核心判断",
        "body": "这条经验假设“留有退路会分散注意力和资源，降低成功概率”。它认为人在绝境中能爆发更强的创造力与执行力，但忽略了并非所有人都能承受极端压力。"
      },
      {
        "heading": "如何使用这条经验",
        "body": "把“all in”理解为阶段性策略：设定明确的时间窗口和可量化的止损条件，而非盲目永久切断后路。可以投入时间、专注度，但需保留基本生活保障。"
      },
      {
        "heading": "高风险下的边界提醒",
        "body": "当创业涉及家庭生计、借高息贷款或押上全部身家时，不留退路可能带来灾难性后果。建议区分财务风险与精力投入，永远为健康、基本生存留一道安全门。"
      }
    ]
  }
}
```

### recommendation_quality_and_creator_diagnostics

Errors:

- `schema_version_not_1.1`
- `function_type_mismatch`
- `result_not_object`

Output preview:

```json
{
  "diagnosis": {
    "relevance_check": {
      "domain_match": "工作",
      "sub_domain_match": "沟通",
      "relevant_candidates": ["e1", "e2", "e3"],
      "irrelevant_candidates": ["e4", "e_low"]
    },
    "quality_check": {
      "tier_distribution": {
        "ai_citable": ["e1", "e2"],
        "recommend_candidate": ["e3"],
        "high_trust": ["e4"],
        "public_visible": ["e_low"]
      },
      "low_quality_mixed": true,
      "low_quality_ids": ["e_low"]
    },
    "diversity_check": {
      "creator_concentration": {
        "A": 3,
        "B": 1,
        "C": 1
      },
      "domain_narrowness": "工作领域集中，但子领域均为沟通，缺乏差异视角",
      "content_similarity": {
        "e1_vs_e2": "高度相似，均强调沟通前确认目标",
        "e1_vs_e3": "相关但侧重不同（确认需求 vs 明确下一步）",
        "e2_vs_e3": "相关但侧重不同（确认目标
```
