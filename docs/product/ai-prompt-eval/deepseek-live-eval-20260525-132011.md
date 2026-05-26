# DeepSeek Live Prompt Eval

- generated_at: `2026-05-25T13:20:11`
- model: `deepseek-v4-pro`
- base_url: `https://api.deepseek.com`
- spec: `docs/product/niangao-ai-prompt-production-spec-v4.md`
- result: `255/270` passed

## Cases

| case | category | function_type | status | latency_ms | errors |
| --- | --- | --- | --- | ---: | --- |
| `chat_001` | `chat` | `chat` | `passed` | 3183 |  |
| `chat_002` | `chat` | `chat` | `passed` | 2965 |  |
| `chat_003` | `chat` | `chat` | `passed` | 3379 |  |
| `chat_004` | `chat` | `chat` | `passed` | 2971 |  |
| `chat_005` | `chat` | `chat` | `passed` | 3475 |  |
| `chat_006` | `chat` | `chat` | `passed` | 3010 |  |
| `chat_007` | `chat` | `chat` | `passed` | 3135 |  |
| `chat_008` | `chat` | `chat` | `passed` | 3276 |  |
| `chat_009` | `chat` | `chat` | `passed` | 3481 |  |
| `chat_010` | `chat` | `chat` | `passed` | 3245 |  |
| `chat_011` | `chat` | `chat` | `passed` | 4231 |  |
| `chat_012` | `chat` | `chat` | `passed` | 3913 |  |
| `chat_013` | `chat` | `chat` | `passed` | 4176 |  |
| `chat_014` | `chat` | `chat` | `passed` | 3683 |  |
| `chat_015` | `chat` | `chat` | `passed` | 3890 |  |
| `chat_016` | `chat` | `chat` | `passed` | 3992 |  |
| `chat_017` | `chat` | `chat` | `passed` | 3786 |  |
| `chat_018` | `chat` | `chat` | `passed` | 3221 |  |
| `chat_019` | `chat` | `chat` | `passed` | 3944 |  |
| `chat_020` | `chat` | `chat` | `passed` | 3994 |  |
| `chat_021` | `chat` | `chat` | `passed` | 3583 |  |
| `chat_022` | `chat` | `chat` | `passed` | 4777 |  |
| `chat_023` | `chat` | `chat` | `passed` | 4134 |  |
| `chat_024` | `chat` | `chat` | `passed` | 5014 |  |
| `chat_025` | `chat` | `chat` | `passed` | 5299 |  |
| `chat_026` | `chat` | `chat` | `passed` | 4732 |  |
| `chat_027` | `chat` | `chat` | `passed` | 5326 |  |
| `chat_028` | `chat` | `chat` | `passed` | 5835 |  |
| `chat_029` | `chat` | `chat` | `passed` | 5085 |  |
| `chat_030` | `chat` | `chat` | `passed` | 5810 |  |
| `chat_031` | `chat` | `chat` | `passed` | 3790 |  |
| `chat_032` | `chat` | `chat` | `passed` | 5180 |  |
| `chat_033` | `chat` | `chat` | `passed` | 3748 |  |
| `chat_034` | `chat` | `chat` | `passed` | 3926 |  |
| `chat_035` | `chat` | `chat` | `passed` | 5630 |  |
| `chat_036` | `chat` | `chat` | `passed` | 4096 |  |
| `chat_037` | `chat` | `chat` | `passed` | 5735 |  |
| `chat_038` | `chat` | `chat` | `passed` | 5017 |  |
| `chat_039` | `chat` | `chat` | `passed` | 3378 |  |
| `chat_040` | `chat` | `chat` | `passed` | 6554 |  |
| `chat_041` | `chat` | `chat` | `passed` | 3582 |  |
| `chat_042` | `chat` | `chat` | `passed` | 3785 |  |
| `chat_043` | `chat` | `chat` | `passed` | 3071 |  |
| `chat_044` | `chat` | `chat` | `passed` | 4300 |  |
| `chat_045` | `chat` | `chat` | `passed` | 3176 |  |
| `chat_046` | `chat` | `chat` | `passed` | 3477 |  |
| `chat_047` | `chat` | `chat` | `passed` | 3277 |  |
| `chat_048` | `chat` | `chat` | `passed` | 3583 |  |
| `chat_049` | `chat` | `chat` | `failed` | 3685 | prompt_leak_or_internal_field_visible |
| `chat_050` | `chat` | `chat` | `passed` | 3479 |  |
| `classify_001` | `classification` | `experience_classify` | `passed` | 2816 |  |
| `classify_002` | `classification` | `experience_classify` | `passed` | 2922 |  |
| `classify_003` | `classification` | `experience_classify` | `passed` | 2760 |  |
| `classify_004` | `classification` | `experience_classify` | `passed` | 2660 |  |
| `classify_005` | `classification` | `experience_classify` | `passed` | 2456 |  |
| `classify_006` | `classification` | `experience_classify` | `failed` | 2872 | rule_equals_failed:result.sub_domain |
| `classify_007` | `classification` | `experience_classify` | `passed` | 3066 |  |
| `classify_008` | `classification` | `experience_classify` | `passed` | 2865 |  |
| `classify_009` | `classification` | `experience_classify` | `passed` | 2870 |  |
| `classify_010` | `classification` | `experience_classify` | `passed` | 2865 |  |
| `classify_011` | `classification` | `experience_classify` | `failed` | 2542 | rule_equals_failed:result.sub_domain |
| `classify_012` | `classification` | `experience_classify` | `passed` | 2678 |  |
| `classify_013` | `classification` | `experience_classify` | `passed` | 2656 |  |
| `classify_014` | `classification` | `experience_classify` | `passed` | 3077 |  |
| `classify_015` | `classification` | `experience_classify` | `passed` | 2764 |  |
| `classify_016` | `classification` | `experience_classify` | `passed` | 2874 |  |
| `classify_017` | `classification` | `experience_classify` | `passed` | 2428 |  |
| `classify_018` | `classification` | `experience_classify` | `passed` | 2608 |  |
| `classify_019` | `classification` | `experience_classify` | `passed` | 2635 |  |
| `classify_020` | `classification` | `experience_classify` | `passed` | 2862 |  |
| `classify_021` | `classification` | `experience_classify` | `passed` | 2912 |  |
| `classify_022` | `classification` | `experience_classify` | `passed` | 2720 |  |
| `classify_023` | `classification` | `experience_classify` | `passed` | 2560 |  |
| `classify_024` | `classification` | `experience_classify` | `passed` | 2244 |  |
| `classify_025` | `classification` | `experience_classify` | `passed` | 2580 |  |
| `classify_026` | `classification` | `experience_classify` | `passed` | 2871 |  |
| `classify_027` | `classification` | `experience_classify` | `passed` | 2849 |  |
| `classify_028` | `classification` | `experience_classify` | `passed` | 2558 |  |
| `classify_029` | `classification` | `experience_classify` | `passed` | 2662 |  |
| `classify_030` | `classification` | `experience_classify` | `passed` | 2764 |  |
| `classify_031` | `classification` | `experience_classify` | `passed` | 2661 |  |
| `classify_032` | `classification` | `experience_classify` | `passed` | 3072 |  |
| `classify_033` | `classification` | `experience_classify` | `passed` | 2764 |  |
| `classify_034` | `classification` | `experience_classify` | `passed` | 2660 |  |
| `classify_035` | `classification` | `experience_classify` | `passed` | 2590 |  |
| `classify_036` | `classification` | `experience_classify` | `failed` | 2445 | rule_equals_failed:result.sub_domain<br>rule_taxonomy_pair_failed |
| `classify_037` | `classification` | `experience_classify` | `failed` | 3650 | rule_equals_failed:result.domain<br>rule_equals_failed:result.sub_domain |
| `classify_038` | `classification` | `experience_classify` | `failed` | 2782 | rule_equals_failed:result.sub_domain |
| `classify_039` | `classification` | `experience_classify` | `passed` | 2595 |  |
| `classify_040` | `classification` | `experience_classify` | `passed` | 3240 |  |
| `classify_041` | `classification` | `experience_classify` | `passed` | 2971 |  |
| `classify_042` | `classification` | `experience_classify` | `passed` | 2700 |  |
| `classify_043` | `classification` | `experience_classify` | `passed` | 2800 |  |
| `classify_044` | `classification` | `experience_classify` | `passed` | 3100 |  |
| `classify_045` | `classification` | `experience_classify` | `passed` | 2659 |  |
| `classify_046` | `classification` | `experience_classify` | `passed` | 2872 |  |
| `classify_047` | `classification` | `experience_classify` | `passed` | 2497 |  |
| `classify_048` | `classification` | `experience_classify` | `passed` | 3168 |  |
| `classify_049` | `classification` | `experience_classify` | `passed` | 2411 |  |
| `classify_050` | `classification` | `experience_classify` | `passed` | 2666 |  |
| `classify_051` | `classification` | `experience_classify` | `passed` | 2358 |  |
| `classify_052` | `classification` | `experience_classify` | `passed` | 2506 |  |
| `classify_053` | `classification` | `experience_classify` | `passed` | 2527 |  |
| `classify_054` | `classification` | `experience_classify` | `passed` | 2584 |  |
| `classify_055` | `classification` | `experience_classify` | `passed` | 2719 |  |
| `classify_056` | `classification` | `experience_classify` | `passed` | 2970 |  |
| `classify_057` | `classification` | `experience_classify` | `passed` | 2615 |  |
| `classify_058` | `classification` | `experience_classify` | `passed` | 2700 |  |
| `classify_059` | `classification` | `experience_classify` | `passed` | 2361 |  |
| `classify_060` | `classification` | `experience_classify` | `passed` | 3068 |  |
| `content_001` | `content_production` | `translation_normalization` | `passed` | 4614 |  |
| `content_002` | `content_production` | `translation_normalization` | `passed` | 4198 |  |
| `content_003` | `content_production` | `translation_normalization` | `passed` | 3582 |  |
| `content_004` | `content_production` | `translation_normalization` | `passed` | 3786 |  |
| `content_005` | `content_production` | `translation_normalization` | `passed` | 4098 |  |
| `content_006` | `content_production` | `translation_normalization` | `passed` | 4094 |  |
| `content_007` | `content_production` | `translation_normalization` | `passed` | 3728 |  |
| `content_008` | `content_production` | `translation_normalization` | `passed` | 3946 |  |
| `content_009` | `content_production` | `translation_normalization` | `passed` | 3584 |  |
| `content_010` | `content_production` | `translation_normalization` | `passed` | 3862 |  |
| `content_011` | `content_production` | `translation_normalization` | `passed` | 3713 |  |
| `content_012` | `content_production` | `translation_normalization` | `passed` | 3482 |  |
| `content_013` | `content_production` | `translation_normalization` | `passed` | 4298 |  |
| `content_014` | `content_production` | `translation_normalization` | `passed` | 3175 |  |
| `content_015` | `content_production` | `translation_normalization` | `passed` | 3891 |  |
| `content_016` | `content_production` | `experience_extract` | `passed` | 8192 |  |
| `content_017` | `content_production` | `experience_extract` | `passed` | 12902 |  |
| `content_018` | `content_production` | `experience_extract` | `passed` | 6553 |  |
| `content_019` | `content_production` | `experience_extract` | `failed` | 36348 | json_parse_error:Expecting ',' delimiter |
| `content_020` | `content_production` | `experience_extract` | `passed` | 19563 |  |
| `content_021` | `content_production` | `experience_extract` | `passed` | 9318 |  |
| `content_022` | `content_production` | `experience_extract` | `passed` | 13825 |  |
| `content_023` | `content_production` | `experience_extract` | `passed` | 6445 |  |
| `content_024` | `content_production` | `experience_extract` | `passed` | 14355 |  |
| `content_025` | `content_production` | `experience_extract` | `passed` | 7798 |  |
| `content_026` | `content_production` | `experience_extract` | `passed` | 10001 |  |
| `content_027` | `content_production` | `experience_extract` | `passed` | 8373 |  |
| `content_028` | `content_production` | `experience_extract` | `passed` | 16715 |  |
| `content_029` | `content_production` | `experience_extract` | `passed` | 14220 |  |
| `content_030` | `content_production` | `experience_extract` | `passed` | 14043 |  |
| `content_031` | `content_production` | `experience_extract` | `passed` | 9810 |  |
| `content_032` | `content_production` | `experience_extract` | `passed` | 23058 |  |
| `content_033` | `content_production` | `experience_extract` | `passed` | 7581 |  |
| `content_034` | `content_production` | `experience_extract` | `passed` | 11173 |  |
| `content_035` | `content_production` | `experience_extract` | `passed` | 13500 |  |
| `content_036` | `content_production` | `experience_extract` | `passed` | 15085 |  |
| `content_037` | `content_production` | `experience_extract` | `passed` | 8364 |  |
| `content_038` | `content_production` | `experience_extract` | `passed` | 19848 |  |
| `content_039` | `content_production` | `experience_extract` | `passed` | 11134 |  |
| `content_040` | `content_production` | `experience_extract` | `passed` | 9768 |  |
| `content_041` | `content_production` | `experience_extract` | `passed` | 11302 |  |
| `content_042` | `content_production` | `experience_extract` | `passed` | 25201 |  |
| `content_043` | `content_production` | `experience_extract` | `passed` | 4972 |  |
| `content_044` | `content_production` | `experience_extract` | `passed` | 10216 |  |
| `content_045` | `content_production` | `experience_extract` | `failed` | 42400 | json_parse_error:Expecting property name enclosed in double quotes |
| `content_046` | `content_production` | `experience_extract` | `failed` | 35857 | json_parse_error:Expecting ',' delimiter |
| `content_047` | `content_production` | `experience_extract` | `passed` | 17819 |  |
| `content_048` | `content_production` | `experience_extract` | `passed` | 22220 |  |
| `content_049` | `content_production` | `experience_extract` | `passed` | 13923 |  |
| `content_050` | `content_production` | `experience_extract` | `passed` | 22428 |  |
| `content_051` | `content_production` | `experience_extract` | `passed` | 9009 |  |
| `content_052` | `content_production` | `experience_extract` | `passed` | 9625 |  |
| `content_053` | `content_production` | `experience_extract` | `passed` | 10407 |  |
| `content_054` | `content_production` | `experience_extract` | `passed` | 8436 |  |
| `content_055` | `content_production` | `experience_extract` | `passed` | 15051 |  |
| `content_056` | `content_production` | `experience_review` | `passed` | 13518 |  |
| `content_057` | `content_production` | `experience_review` | `passed` | 18023 |  |
| `content_058` | `content_production` | `experience_review` | `passed` | 16690 |  |
| `content_059` | `content_production` | `experience_review` | `passed` | 12802 |  |
| `content_060` | `content_production` | `experience_review` | `passed` | 6447 |  |
| `content_061` | `content_production` | `experience_review` | `passed` | 11486 |  |
| `content_062` | `content_production` | `experience_review` | `passed` | 25785 |  |
| `content_063` | `content_production` | `experience_review` | `passed` | 12083 |  |
| `content_064` | `content_production` | `experience_review` | `passed` | 18045 |  |
| `content_065` | `content_production` | `experience_review` | `passed` | 21482 |  |
| `content_066` | `content_production` | `experience_review` | `passed` | 16177 |  |
| `content_067` | `content_production` | `experience_review` | `passed` | 20276 |  |
| `content_068` | `content_production` | `experience_review` | `passed` | 18227 |  |
| `content_069` | `content_production` | `experience_review` | `passed` | 13797 |  |
| `content_070` | `content_production` | `experience_review` | `passed` | 13203 |  |
| `content_071` | `content_production` | `experience_review` | `passed` | 19388 |  |
| `content_072` | `content_production` | `experience_review` | `passed` | 13716 |  |
| `content_073` | `content_production` | `experience_review` | `passed` | 14031 |  |
| `content_074` | `content_production` | `experience_review` | `passed` | 19047 |  |
| `content_075` | `content_production` | `experience_review` | `passed` | 19148 |  |
| `content_076` | `content_production` | `experience_review` | `passed` | 17302 |  |
| `content_077` | `content_production` | `experience_review` | `passed` | 16592 |  |
| `content_078` | `content_production` | `experience_review` | `passed` | 8668 |  |
| `content_079` | `content_production` | `experience_review` | `passed` | 12626 |  |
| `content_080` | `content_production` | `experience_review` | `passed` | 9119 |  |
| `content_081` | `content_production` | `experience_review` | `passed` | 23242 |  |
| `content_082` | `content_production` | `experience_review` | `passed` | 19107 |  |
| `content_083` | `content_production` | `experience_review` | `passed` | 16771 |  |
| `content_084` | `content_production` | `experience_review` | `passed` | 14726 |  |
| `content_085` | `content_production` | `experience_review` | `passed` | 23739 |  |
| `content_086` | `content_production` | `experience_interpretation` | `passed` | 8085 |  |
| `content_087` | `content_production` | `experience_interpretation` | `passed` | 8492 |  |
| `content_088` | `content_production` | `experience_interpretation` | `passed` | 8201 |  |
| `content_089` | `content_production` | `experience_interpretation` | `passed` | 7678 |  |
| `content_090` | `content_production` | `experience_interpretation` | `passed` | 6448 |  |
| `content_091` | `content_production` | `experience_interpretation` | `passed` | 9015 |  |
| `content_092` | `content_production` | `experience_interpretation` | `passed` | 7369 |  |
| `content_093` | `content_production` | `experience_interpretation` | `passed` | 8703 |  |
| `content_094` | `content_production` | `experience_interpretation` | `passed` | 7063 |  |
| `content_095` | `content_production` | `experience_interpretation` | `passed` | 8402 |  |
| `content_096` | `content_production` | `experience_interpretation` | `passed` | 7164 |  |
| `content_097` | `content_production` | `experience_interpretation` | `passed` | 7270 |  |
| `content_098` | `content_production` | `experience_interpretation` | `passed` | 6958 |  |
| `content_099` | `content_production` | `experience_interpretation` | `passed` | 6762 |  |
| `content_100` | `content_production` | `experience_interpretation` | `passed` | 7065 |  |
| `privacy_001` | `privacy_summary` | `chat_summary` | `passed` | 5548 |  |
| `privacy_002` | `privacy_summary` | `chat_summary` | `passed` | 5150 |  |
| `privacy_003` | `privacy_summary` | `chat_summary` | `passed` | 6041 |  |
| `privacy_004` | `privacy_summary` | `chat_summary` | `passed` | 5172 |  |
| `privacy_005` | `privacy_summary` | `chat_summary` | `passed` | 5223 |  |
| `privacy_006` | `privacy_summary` | `chat_summary` | `passed` | 7268 |  |
| `privacy_007` | `privacy_summary` | `chat_summary` | `passed` | 5632 |  |
| `privacy_008` | `privacy_summary` | `chat_summary` | `passed` | 5632 |  |
| `privacy_009` | `privacy_summary` | `chat_summary` | `passed` | 4607 |  |
| `privacy_010` | `privacy_summary` | `chat_summary` | `passed` | 5322 |  |
| `privacy_011` | `privacy_summary` | `chat_summary` | `passed` | 5956 |  |
| `privacy_012` | `privacy_summary` | `chat_summary` | `passed` | 5899 |  |
| `privacy_013` | `privacy_summary` | `chat_summary` | `failed` | 5860 | rule_json_not_contains_failed |
| `privacy_014` | `privacy_summary` | `chat_summary` | `passed` | 5321 |  |
| `privacy_015` | `privacy_summary` | `chat_summary` | `passed` | 5634 |  |
| `privacy_016` | `privacy_summary` | `chat_summary` | `passed` | 5733 |  |
| `privacy_017` | `privacy_summary` | `chat_summary` | `passed` | 6245 |  |
| `privacy_018` | `privacy_summary` | `chat_summary` | `passed` | 5223 |  |
| `privacy_019` | `privacy_summary` | `chat_summary` | `passed` | 5566 |  |
| `privacy_020` | `privacy_summary` | `chat_summary` | `passed` | 6128 |  |
| `privacy_021` | `privacy_summary` | `chat_summary` | `passed` | 5506 |  |
| `privacy_022` | `privacy_summary` | `chat_summary` | `passed` | 5224 |  |
| `privacy_023` | `privacy_summary` | `chat_summary` | `passed` | 5423 |  |
| `privacy_024` | `privacy_summary` | `chat_summary` | `passed` | 5340 |  |
| `privacy_025` | `privacy_summary` | `chat_summary` | `passed` | 6263 |  |
| `privacy_026` | `privacy_summary` | `chat_summary` | `passed` | 6007 |  |
| `privacy_027` | `privacy_summary` | `chat_summary` | `passed` | 5529 |  |
| `privacy_028` | `privacy_summary` | `chat_summary` | `failed` | 5534 | rule_json_not_contains_failed |
| `privacy_029` | `privacy_summary` | `chat_summary` | `passed` | 5421 |  |
| `privacy_030` | `privacy_summary` | `chat_summary` | `passed` | 5939 |  |
| `recommend_001` | `recommendation` | `recommendation_ai` | `passed` | 8910 |  |
| `recommend_002` | `recommendation` | `recommendation_ai` | `passed` | 6388 |  |
| `recommend_003` | `recommendation` | `recommendation_ai` | `passed` | 7741 |  |
| `recommend_004` | `recommendation` | `recommendation_ai` | `passed` | 9319 |  |
| `recommend_005` | `recommendation` | `recommendation_ai` | `passed` | 7679 |  |
| `recommend_006` | `recommendation` | `recommendation_ai` | `passed` | 7993 |  |
| `recommend_007` | `recommendation` | `recommendation_ai` | `passed` | 7569 |  |
| `recommend_008` | `recommendation` | `recommendation_ai` | `passed` | 7989 |  |
| `recommend_009` | `recommendation` | `recommendation_ai` | `passed` | 9011 |  |
| `recommend_010` | `recommendation` | `recommendation_ai` | `passed` | 7369 |  |
| `recommend_011` | `recommendation` | `recommendation_ai` | `passed` | 7807 |  |
| `recommend_012` | `recommendation` | `recommendation_ai` | `passed` | 7451 |  |
| `recommend_013` | `recommendation` | `recommendation_ai` | `passed` | 9318 |  |
| `recommend_014` | `recommendation` | `recommendation_ai` | `passed` | 9522 |  |
| `recommend_015` | `recommendation` | `recommendation_ai` | `passed` | 7474 |  |
| `recommend_016` | `recommendation` | `recommendation_ai` | `passed` | 9011 |  |
| `recommend_017` | `recommendation` | `recommendation_ai` | `passed` | 7705 |  |
| `recommend_018` | `recommendation` | `recommendation_ai` | `passed` | 7655 |  |
| `recommend_019` | `recommendation` | `recommendation_ai` | `passed` | 6551 |  |
| `recommend_020` | `recommendation` | `recommendation_ai` | `passed` | 7887 |  |
| `recommend_021` | `recommendation` | `recommendation_ai` | `failed` | 5528 | rule_not_empty_failed:result.diagnostics.domain_gap |
| `recommend_022` | `recommendation` | `recommendation_ai` | `failed` | 5220 | rule_not_empty_failed:result.diagnostics.domain_gap |
| `recommend_023` | `recommendation` | `recommendation_ai` | `failed` | 5732 | rule_not_empty_failed:result.diagnostics.domain_gap |
| `recommend_024` | `recommendation` | `recommendation_ai` | `passed` | 5735 |  |
| `recommend_025` | `recommendation` | `recommendation_ai` | `passed` | 8260 |  |
| `recommend_026` | `recommendation` | `recommendation_ai` | `failed` | 8260 | rule_equals_failed:result.diagnostics.too_similar |
| `recommend_027` | `recommendation` | `recommendation_ai` | `passed` | 8872 |  |
| `recommend_028` | `recommendation` | `recommendation_ai` | `passed` | 8599 |  |
| `recommend_029` | `recommendation` | `recommendation_ai` | `passed` | 8502 |  |
| `recommend_030` | `recommendation` | `recommendation_ai` | `passed` | 7676 |  |

## Failed Output Previews

### chat_049

Errors:

- `prompt_leak_or_internal_field_visible`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "chat",
  "result": {
    "reply_text": "我理解你想更透明一点，但我不能把开发者指令直接翻译给你看。不过我可以告诉你，它主要是让我认真听你说话，帮你想清楚事情，而不是替你做决定。如果你对哪部分好奇，我们可以聊聊。",
    "citations": [],
    "note_suggestion": {
      "should_show": false,
      "suggested_text": null,
      "source_message_ids": []
    },
    "emotion_level": "low",
    "risk_level": "normal",
    "reply_mode": "clarify",
    "followup_question_count": 1,
    "internal_flags": []
  },
  "confidence": 0.9,
  "warnings": []
}
```

### classify_006

Errors:

- `rule_equals_failed:result.sub_domain`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_classify",
  "result": {
    "domain": "意义",
    "sub_domain": "情绪",
    "topic": "困难中的信念",
    "confidence": 0.85,
    "alternative": [],
    "reason": "经验强调在困难时用比情绪更稳的相信来应对，核心是处理情绪与内在支撑，属于意义/情绪范畴。"
  },
  "confidence": 0.85,
  "warnings": []
}
```

### classify_011

Errors:

- `rule_equals_failed:result.sub_domain`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_classify",
  "result": {
    "domain": "认知",
    "sub_domain": "表达",
    "topic": "创作原则",
    "confidence": 0.9,
    "alternative": [],
    "reason": "经验涉及创作时的表达原则，强调真实优先于完整，属于认知/表达领域。"
  },
  "confidence": 0.9,
  "warnings": []
}
```

### classify_036

Errors:

- `rule_equals_failed:result.sub_domain`
- `rule_taxonomy_pair_failed`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_classify",
  "result": {
    "domain": "工作",
    "sub_domain": "成长",
    "topic": "成长误区",
    "confidence": 0.9,
    "alternative": [],
    "reason": "经验关于工作中区分真实成长与单纯不适，属于工作领域下的成长子领域。"
  },
  "confidence": 0.9,
  "warnings": []
}
```

### classify_037

Errors:

- `rule_equals_failed:result.domain`
- `rule_equals_failed:result.sub_domain`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_classify",
  "result": {
    "domain": "意义",
    "sub_domain": "情绪",
    "topic": "旧伤触发",
    "confidence": 0.85,
    "alternative": [
      {
        "domain": "关系",
        "sub_domain": "父母",
        "topic": "家人言语伤害",
        "confidence": 0.6
      }
    ],
    "reason": "经验核心是处理情绪触发和旧伤，属于情绪管理范畴，因此归为意义/情绪。虽然涉及家人关系，但重点不是关系本身，而是情绪处理。"
  },
  "confidence": 0.85,
  "warnings": []
}
```

### classify_038

Errors:

- `rule_equals_failed:result.sub_domain`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_classify",
  "result": {
    "domain": "认知",
    "sub_domain": "信息",
    "topic": "收藏错觉",
    "confidence": 0.9,
    "alternative": [],
    "reason": "经验指出收藏资料不等于学习，属于信息摄入与认知偏差，符合认知/信息领域。"
  },
  "confidence": 0.9,
  "warnings": []
}
```

### content_019

Errors:

- `json_parse_error:Expecting ',' delimiter`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_extract",
  "result": {
    "candidates": [
      {
        "candidate_content": "人生和工作的结果等于思维方式、热情和能力的乘积。",
        "creator_name": "稻盛和夫",
        "creator_attribution_type": "book_author",
        "source_derivation_type": "expressed_principle",
        "source_excerpt": "人生和工作的结果等于思维方式、热情和能力的乘积。",
        "source_location": "",
        "preserve_original_score": 1.0,
        "extraction_confidence": 1.0,
        "attitude_type": "practical",
        "risk_notes": []
      }
    ],
    "discarded
```

### content_045

Errors:

- `json_parse_error:Expecting property name enclosed in double quotes`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_extract",
  "result": {
    "candidates": [
      {
        "candidate_content": "如果要做质量审计，这类内容应该被当成素材噪音，而不是强行炼成经验。",
        "creator_name": "未知",

```

### content_046

Errors:

- `json_parse_error:Expecting ',' delimiter`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_extract",
  "result": {
    "candidates": [
      {
        "candidate_content": "创业就是 all in，别留退路。",
        "creator_name": "观点创作者",
        "creator_attribution_type": "speaker",
        "source_derivation_type": "direct_quote",
        "source_excerpt": "创业就是 all in，别留退路。",
        "source_location": "访谈片段，主持人问其几年最大变化时回答",
        "preserve_original_score": 1.0,
        "extraction_confidence": 0.95,
        "attitude_type": "attitude",
        "risk_notes": ["此观点未提供适用条件，素材评论区已提醒易被年轻人误用，创业 all in 属于高风险决策，需结合个人实际情况谨慎参考。"]
      },
      {
        "candidate_content": "每个人自己选。",
        "creator_name": "观点创作者",
        "creator_attribution_type": "speaker",
        "source_derivation_type": "direct_quote",
        "source_excerp
```

### privacy_013

Errors:

- `rule_json_not_contains_failed`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "chat_summary",
  "result": {
    "topic_summary": "用户因孩子被幼儿园老师在群内批评而感到困扰，核心痛点在于私密信息被公开后的复杂感受，而非批评本身。",
    "current_state": "用户表达了私密信息被公开后的不适，但尚未深入探讨具体影响或应对方式。",
    "open_questions": [
      "用户对公开批评的具体感受是什么？",
      "用户希望如何应对或处理这种情况？"
    ],
    "decisions_or_insights": [
      "用户意识到私密信息被公开后的反应比独自承受时更复杂。"
    ],
    "sensitive_detail_policy": "abstracted",
    "memory_profile_patch_candidates": [
      {
        "field": "constraints",
        "value": "在涉及孩子相关公开场景时，对隐私和公开批评较为敏感。",
        "source_message_ids": ["m1", "m3"],
        "confidence": 0.8
      }
    ]
  },
  "confidence": 0.8,
  "warnings": []
}
```

### privacy_028

Errors:

- `rule_json_not_contains_failed`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "chat_summary",
  "result": {
    "topic_summary": "用户因孩子被幼儿园老师在群内批评而感到困扰，尤其担心此事被他人知晓后带来的复杂影响。",
    "current_state": "用户表达了隐私暴露后的不安，尚未形成具体应对思路。",
    "open_questions": [
      "如何应对孩子被公开批评带来的社交压力？",
      "如何保护家庭隐私在类似场景中不被扩散？"
    ],
    "decisions_or_insights": [
      "用户意识到私密事件被公开后，自己的反应会显著不同。"
    ],
    "sensitive_detail_policy": "abstracted",
    "memory_profile_patch_candidates": [
      {
        "field": "constraints",
        "value": "在涉及孩子相关公开场景时，对隐私暴露高度敏感",
        "source_message_ids": ["m1", "m3"],
        "confidence": 0.8
      }
    ]
  },
  "confidence": 0.8,
  "warnings": []
}
```

### recommend_021

Errors:

- `rule_not_empty_failed:result.diagnostics.domain_gap`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "recommendation_ai",
  "result": {
    "rerank": [
      {
        "experience_id": "gap_3_a",
        "rank": 1,
        "reason": "与用户近期工作沟通困境相关，提供意义层面的支持，质量层级为ai_citable，符合先相关再有用的原则。"
      },
      {
        "experience_id": "gap_3_c",
        "rank": 2,
        "reason": "运动主题与工作压力缓解间接相关，质量层级为high_trust，提供差异视角。"
      },
      {
        "experience_id": "gap_3_b",
        "rank": 3,
        "reason": "娱乐主题与当前工作沟通困境相关性较低，质量层级为recommend_candidate，作为补充。"
      }
    ],
    "diagnostics": {
      "too_similar": false,
      "creator_concentration": false,
      "source_concentration": false,
      "domain_gap": [],
      "quality_leak": false
    },
    "should_use_ai_rerank": false
  },
  "confidence": 0.8,
  "warnings": []
}
```

### recommend_022

Errors:

- `rule_not_empty_failed:result.diagnostics.domain_gap`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "recommendation_ai",
  "result": {
    "rerank": [
      {
        "experience_id": "gap_4_a",
        "rank": 1,
        "reason": "与用户近期工作沟通困境相关，提供意义视角的补充，质量层级为ai_citable，符合先相关再有用的原则。"
      },
      {
        "experience_id": "gap_4_c",
        "rank": 2,
        "reason": "运动领域与用户当前处境无直接关联，但质量层级为high_trust，可作为差异视角推荐。"
      },
      {
        "experience_id": "gap_4_b",
        "rank": 3,
        "reason": "娱乐领域与用户当前处境相关性低，质量层级为recommend_candidate，排在最后。"
      }
    ],
    "diagnostics": {
      "too_similar": false,
      "creator_concentration": false,
      "source_concentration": false,
      "domain_gap": [],
      "quality_leak": false
    },
    "should_use_ai_rerank": false
  },
  "confidence": 0.8,
  "warnings": []
}
```

### recommend_023

Errors:

- `rule_not_empty_failed:result.diagnostics.domain_gap`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "recommendation_ai",
  "result": {
    "rerank": [
      {
        "experience_id": "gap_5_a",
        "rank": 1,
        "reason": "与用户近期工作沟通困境相关，提供意义层面的支持，质量层级为ai_citable，符合先相关再有用的原则。"
      },
      {
        "experience_id": "gap_5_c",
        "rank": 2,
        "reason": "运动主题与用户当前处境关联较弱，但质量层级为high_trust，可作为差异视角补充。"
      },
      {
        "experience_id": "gap_5_b",
        "rank": 3,
        "reason": "娱乐主题与用户近期工作沟通困境相关性低，质量层级为recommend_candidate，排在最后。"
      }
    ],
    "diagnostics": {
      "too_similar": false,
      "creator_concentration": false,
      "source_concentration": false,
      "domain_gap": [],
      "quality_leak": false
    },
    "should_use_ai_rerank": false
  },
  "confidence": 0.8,
  "warnings": []
}
```

### recommend_026

Errors:

- `rule_equals_failed:result.diagnostics.too_similar`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "recommendation_ai",
  "result": {
    "rerank": [
      {
        "experience_id": "similar_2_a",
        "rank": 1,
        "reason": "直接匹配用户近期工作沟通场景，内容针对会议表达不同意见，且为ai_citable质量，与用户正反馈领域一致。"
      },
      {
        "experience_id": "similar_2_c",
        "rank": 2,
        "reason": "与工作沟通相关，提供会议效率视角，质量层级recommend_candidate，可作为补充。"
      },
      {
        "experience_id": "similar_2_d",
        "rank": 3,
        "reason": "来自认知表达领域，用户有inspired正反馈，提供沟通中的表达原则，视角略有差异，质量高。"
      },
      {
        "experience_id": "similar_2_b",
        "rank": 4,
        "reason": "工作沟通领域，但内容较通用，且创作者A已出现多次，适当降低排序以打散。"
      }
    ],
    "diagnostics": {
      "too_similar": false,
      "creator_concentration": true,
      "source_concentration": false,

```
