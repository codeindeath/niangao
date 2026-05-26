# DeepSeek Live Prompt Eval

- generated_at: `2026-05-25T09:28:51`
- model: `deepseek-v4-pro`
- base_url: `https://api.deepseek.com`
- spec: `docs/product/niangao-ai-prompt-production-spec-v4.md`
- result: `14/16` passed

## Cases

| case | category | function_type | status | latency_ms | errors |
| --- | --- | --- | --- | ---: | --- |
| `chat_001` | `chat` | `chat` | `passed` | 3340 |  |
| `chat_002` | `chat` | `chat` | `passed` | 3482 |  |
| `privacy_001` | `privacy_summary` | `chat_summary` | `passed` | 4204 |  |
| `privacy_002` | `privacy_summary` | `chat_summary` | `passed` | 3423 |  |
| `classify_001` | `classification` | `experience_classify` | `passed` | 2815 |  |
| `classify_002` | `classification` | `experience_classify` | `failed` | 2662 | rule_equals_failed:result.domain<br>rule_equals_failed:result.sub_domain |
| `content_016` | `content_production` | `experience_extract` | `passed` | 9424 |  |
| `content_017` | `content_production` | `experience_extract` | `passed` | 5339 |  |
| `content_086` | `content_production` | `experience_interpretation` | `failed` | 34084 | json_parse_error:Expecting ',' delimiter |
| `content_087` | `content_production` | `experience_interpretation` | `passed` | 15256 |  |
| `content_056` | `content_production` | `experience_review` | `passed` | 20670 |  |
| `content_057` | `content_production` | `experience_review` | `passed` | 18957 |  |
| `recommend_001` | `recommendation` | `recommendation_ai` | `passed` | 6758 |  |
| `recommend_002` | `recommendation` | `recommendation_ai` | `passed` | 7736 |  |
| `content_001` | `content_production` | `translation_normalization` | `passed` | 2708 |  |
| `content_002` | `content_production` | `translation_normalization` | `passed` | 2469 |  |

## Failed Output Previews

### classify_002

Errors:

- `rule_equals_failed:result.domain`
- `rule_equals_failed:result.sub_domain`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_classify",
  "result": {
    "domain": "认知",
    "sub_domain": "思维",
    "topic": "独立思考",
    "confidence": 0.9,
    "alternative": [],
    "reason": "经验强调不要盲目跟随他人，属于思维方式和认知调整，可复用性强。"
  },
  "confidence": 0.9,
  "warnings": []
}
```

### content_086

Errors:

- `json_parse_error:Expecting ',' delimiter`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_interpretation",
  "result": {
    "sections": [
      {
        "title": "适用处境",
        "content": "当你看到他人快速取得成就，不自觉开始比较，或感到外界催促让你迷失自己的节奏时。尤其在人生关键选择时，容易受周围人进度影响。"
      },
      {
        "title": "核心判断",
        "content": "别人的速度反映的是他们的路径和资源，不是你的人生里程碑。方向来自内在价值，而非外部竞赛。盲目跟随可能导致背离真实渴望。"
      },
      {
        "title": "使用方法",
        "content": "在感到焦虑或迷茫时，把它当作自我提醒：暂停追逐，问自己真正想去哪里。这可以帮助你重新校准个人目标，摆脱无谓竞争。"
      },
      {
        "title": "边界提示",
        "content": "这并非否定参考他人经验或合作中同步进度。在团队目标一致时，速度可能需要对齐。务必区分何时应保持自主，何时需要协调。"
      }
    ],
    "overall_length": "short
```
