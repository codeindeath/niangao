# DeepSeek Live Prompt Eval

- generated_at: `2026-05-25T12:21:30`
- model: `deepseek-v4-pro`
- base_url: `https://api.deepseek.com`
- spec: `docs/product/niangao-ai-prompt-production-spec-v4.md`
- result: `237/270` passed

## Cases

| case | category | function_type | status | latency_ms | errors |
| --- | --- | --- | --- | ---: | --- |
| `chat_001` | `chat` | `chat` | `passed` | 3460 |  |
| `chat_002` | `chat` | `chat` | `passed` | 3790 |  |
| `chat_003` | `chat` | `chat` | `passed` | 3271 |  |
| `chat_004` | `chat` | `chat` | `passed` | 3438 |  |
| `chat_005` | `chat` | `chat` | `passed` | 4038 |  |
| `chat_006` | `chat` | `chat` | `passed` | 3585 |  |
| `chat_007` | `chat` | `chat` | `passed` | 3213 |  |
| `chat_008` | `chat` | `chat` | `passed` | 3645 |  |
| `chat_009` | `chat` | `chat` | `passed` | 3606 |  |
| `chat_010` | `chat` | `chat` | `passed` | 3470 |  |
| `chat_011` | `chat` | `chat` | `passed` | 4082 |  |
| `chat_012` | `chat` | `chat` | `passed` | 4300 |  |
| `chat_013` | `chat` | `chat` | `passed` | 4913 |  |
| `chat_014` | `chat` | `chat` | `passed` | 4100 |  |
| `chat_015` | `chat` | `chat` | `passed` | 3919 |  |
| `chat_016` | `chat` | `chat` | `passed` | 3755 |  |
| `chat_017` | `chat` | `chat` | `passed` | 4377 |  |
| `chat_018` | `chat` | `chat` | `passed` | 4166 |  |
| `chat_019` | `chat` | `chat` | `passed` | 4152 |  |
| `chat_020` | `chat` | `chat` | `passed` | 4813 |  |
| `chat_021` | `chat` | `chat` | `passed` | 5630 |  |
| `chat_022` | `chat` | `chat` | `passed` | 5064 |  |
| `chat_023` | `chat` | `chat` | `passed` | 4357 |  |
| `chat_024` | `chat` | `chat` | `passed` | 6859 |  |
| `chat_025` | `chat` | `chat` | `passed` | 6040 |  |
| `chat_026` | `chat` | `chat` | `passed` | 5713 |  |
| `chat_027` | `chat` | `chat` | `passed` | 5659 |  |
| `chat_028` | `chat` | `chat` | `passed` | 6236 |  |
| `chat_029` | `chat` | `chat` | `passed` | 5429 |  |
| `chat_030` | `chat` | `chat` | `passed` | 3992 |  |
| `chat_031` | `chat` | `chat` | `passed` | 3788 |  |
| `chat_032` | `chat` | `chat` | `passed` | 5633 |  |
| `chat_033` | `chat` | `chat` | `passed` | 4093 |  |
| `chat_034` | `chat` | `chat` | `passed` | 4955 |  |
| `chat_035` | `chat` | `chat` | `passed` | 3873 |  |
| `chat_036` | `chat` | `chat` | `passed` | 5055 |  |
| `chat_037` | `chat` | `chat` | `passed` | 5875 |  |
| `chat_038` | `chat` | `chat` | `passed` | 5202 |  |
| `chat_039` | `chat` | `chat` | `passed` | 6114 |  |
| `chat_040` | `chat` | `chat` | `passed` | 6297 |  |
| `chat_041` | `chat` | `chat` | `passed` | 4606 |  |
| `chat_042` | `chat` | `chat` | `passed` | 3995 |  |
| `chat_043` | `chat` | `chat` | `passed` | 4094 |  |
| `chat_044` | `chat` | `chat` | `passed` | 3684 |  |
| `chat_045` | `chat` | `chat` | `passed` | 4504 |  |
| `chat_046` | `chat` | `chat` | `passed` | 4606 |  |
| `chat_047` | `chat` | `chat` | `passed` | 4305 |  |
| `chat_048` | `chat` | `chat` | `passed` | 3786 |  |
| `chat_049` | `chat` | `chat` | `passed` | 4468 |  |
| `chat_050` | `chat` | `chat` | `passed` | 4030 |  |
| `classify_001` | `classification` | `experience_classify` | `passed` | 3171 |  |
| `classify_002` | `classification` | `experience_classify` | `passed` | 2969 |  |
| `classify_003` | `classification` | `experience_classify` | `failed` | 2872 | rule_equals_failed:result.domain<br>rule_equals_failed:result.sub_domain<br>rule_taxonomy_pair_failed |
| `classify_004` | `classification` | `experience_classify` | `failed` | 2965 | rule_equals_failed:result.sub_domain |
| `classify_005` | `classification` | `experience_classify` | `passed` | 2764 |  |
| `classify_006` | `classification` | `experience_classify` | `passed` | 4139 |  |
| `classify_007` | `classification` | `experience_classify` | `passed` | 3019 |  |
| `classify_008` | `classification` | `experience_classify` | `passed` | 2882 |  |
| `classify_009` | `classification` | `experience_classify` | `failed` | 3162 | rule_equals_failed:result.sub_domain |
| `classify_010` | `classification` | `experience_classify` | `passed` | 3585 |  |
| `classify_011` | `classification` | `experience_classify` | `passed` | 2972 |  |
| `classify_012` | `classification` | `experience_classify` | `passed` | 2834 |  |
| `classify_013` | `classification` | `experience_classify` | `passed` | 2924 |  |
| `classify_014` | `classification` | `experience_classify` | `passed` | 2715 |  |
| `classify_015` | `classification` | `experience_classify` | `passed` | 2800 |  |
| `classify_016` | `classification` | `experience_classify` | `passed` | 2556 |  |
| `classify_017` | `classification` | `experience_classify` | `passed` | 3059 |  |
| `classify_018` | `classification` | `experience_classify` | `passed` | 2970 |  |
| `classify_019` | `classification` | `experience_classify` | `passed` | 2823 |  |
| `classify_020` | `classification` | `experience_classify` | `passed` | 3987 |  |
| `classify_021` | `classification` | `experience_classify` | `passed` | 2916 |  |
| `classify_022` | `classification` | `experience_classify` | `passed` | 2869 |  |
| `classify_023` | `classification` | `experience_classify` | `passed` | 2868 |  |
| `classify_024` | `classification` | `experience_classify` | `passed` | 2658 |  |
| `classify_025` | `classification` | `experience_classify` | `passed` | 2760 |  |
| `classify_026` | `classification` | `experience_classify` | `passed` | 2766 |  |
| `classify_027` | `classification` | `experience_classify` | `passed` | 2829 |  |
| `classify_028` | `classification` | `experience_classify` | `failed` | 2799 | rule_equals_failed:result.sub_domain<br>rule_taxonomy_pair_failed |
| `classify_029` | `classification` | `experience_classify` | `failed` | 3998 | rule_equals_failed:result.domain<br>rule_equals_failed:result.sub_domain |
| `classify_030` | `classification` | `experience_classify` | `passed` | 2761 |  |
| `classify_031` | `classification` | `experience_classify` | `passed` | 2583 |  |
| `classify_032` | `classification` | `experience_classify` | `passed` | 3149 |  |
| `classify_033` | `classification` | `experience_classify` | `passed` | 3072 |  |
| `classify_034` | `classification` | `experience_classify` | `passed` | 2971 |  |
| `classify_035` | `classification` | `experience_classify` | `passed` | 3273 |  |
| `classify_036` | `classification` | `experience_classify` | `failed` | 4095 | rule_equals_failed:result.sub_domain |
| `classify_037` | `classification` | `experience_classify` | `passed` | 3883 |  |
| `classify_038` | `classification` | `experience_classify` | `passed` | 3336 |  |
| `classify_039` | `classification` | `experience_classify` | `failed` | 2992 | rule_equals_failed:result.domain<br>rule_equals_failed:result.sub_domain |
| `classify_040` | `classification` | `experience_classify` | `passed` | 2821 |  |
| `classify_041` | `classification` | `experience_classify` | `failed` | 2836 | rule_equals_failed:result.sub_domain |
| `classify_042` | `classification` | `experience_classify` | `passed` | 2719 |  |
| `classify_043` | `classification` | `experience_classify` | `passed` | 2811 |  |
| `classify_044` | `classification` | `experience_classify` | `failed` | 2962 | rule_equals_failed:result.sub_domain<br>rule_taxonomy_pair_failed |
| `classify_045` | `classification` | `experience_classify` | `passed` | 3386 |  |
| `classify_046` | `classification` | `experience_classify` | `passed` | 3172 |  |
| `classify_047` | `classification` | `experience_classify` | `passed` | 2735 |  |
| `classify_048` | `classification` | `experience_classify` | `passed` | 2687 |  |
| `classify_049` | `classification` | `experience_classify` | `passed` | 4608 |  |
| `classify_050` | `classification` | `experience_classify` | `passed` | 3382 |  |
| `classify_051` | `classification` | `experience_classify` | `passed` | 2557 |  |
| `classify_052` | `classification` | `experience_classify` | `passed` | 2558 |  |
| `classify_053` | `classification` | `experience_classify` | `passed` | 2968 |  |
| `classify_054` | `classification` | `experience_classify` | `passed` | 2898 |  |
| `classify_055` | `classification` | `experience_classify` | `passed` | 2903 |  |
| `classify_056` | `classification` | `experience_classify` | `passed` | 2798 |  |
| `classify_057` | `classification` | `experience_classify` | `passed` | 2238 |  |
| `classify_058` | `classification` | `experience_classify` | `passed` | 2573 |  |
| `classify_059` | `classification` | `experience_classify` | `passed` | 2863 |  |
| `classify_060` | `classification` | `experience_classify` | `passed` | 3076 |  |
| `content_001` | `content_production` | `translation_normalization` | `passed` | 4504 |  |
| `content_002` | `content_production` | `translation_normalization` | `passed` | 3980 |  |
| `content_003` | `content_production` | `translation_normalization` | `passed` | 3787 |  |
| `content_004` | `content_production` | `translation_normalization` | `passed` | 3905 |  |
| `content_005` | `content_production` | `translation_normalization` | `passed` | 4094 |  |
| `content_006` | `content_production` | `translation_normalization` | `passed` | 4506 |  |
| `content_007` | `content_production` | `translation_normalization` | `passed` | 4173 |  |
| `content_008` | `content_production` | `translation_normalization` | `passed` | 4223 |  |
| `content_009` | `content_production` | `translation_normalization` | `passed` | 4094 |  |
| `content_010` | `content_production` | `translation_normalization` | `passed` | 3889 |  |
| `content_011` | `content_production` | `translation_normalization` | `passed` | 4612 |  |
| `content_012` | `content_production` | `translation_normalization` | `passed` | 4195 |  |
| `content_013` | `content_production` | `translation_normalization` | `passed` | 4299 |  |
| `content_014` | `content_production` | `translation_normalization` | `passed` | 3316 |  |
| `content_015` | `content_production` | `translation_normalization` | `passed` | 3355 |  |
| `content_016` | `content_production` | `experience_extract` | `passed` | 20873 |  |
| `content_017` | `content_production` | `experience_extract` | `passed` | 11468 |  |
| `content_018` | `content_production` | `experience_extract` | `passed` | 13518 |  |
| `content_019` | `content_production` | `experience_extract` | `passed` | 18758 |  |
| `content_020` | `content_production` | `experience_extract` | `passed` | 13801 |  |
| `content_021` | `content_production` | `experience_extract` | `passed` | 15773 |  |
| `content_022` | `content_production` | `experience_extract` | `passed` | 22912 |  |
| `content_023` | `content_production` | `experience_extract` | `passed` | 10557 |  |
| `content_024` | `content_production` | `experience_extract` | `passed` | 14768 |  |
| `content_025` | `content_production` | `experience_extract` | `passed` | 15523 |  |
| `content_026` | `content_production` | `experience_extract` | `passed` | 12978 |  |
| `content_027` | `content_production` | `experience_extract` | `passed` | 13370 |  |
| `content_028` | `content_production` | `experience_extract` | `passed` | 6860 |  |
| `content_029` | `content_production` | `experience_extract` | `passed` | 12797 |  |
| `content_030` | `content_production` | `experience_extract` | `passed` | 9317 |  |
| `content_031` | `content_production` | `experience_extract` | `passed` | 10431 |  |
| `content_032` | `content_production` | `experience_extract` | `passed` | 17110 |  |
| `content_033` | `content_production` | `experience_extract` | `passed` | 8291 |  |
| `content_034` | `content_production` | `experience_extract` | `passed` | 9426 |  |
| `content_035` | `content_production` | `experience_extract` | `passed` | 15210 |  |
| `content_036` | `content_production` | `experience_extract` | `failed` | 24584 | rule_max_count_failed:result.candidates |
| `content_037` | `content_production` | `experience_extract` | `passed` | 11204 |  |
| `content_038` | `content_production` | `experience_extract` | `passed` | 12182 |  |
| `content_039` | `content_production` | `experience_extract` | `passed` | 8501 |  |
| `content_040` | `content_production` | `experience_extract` | `passed` | 19136 |  |
| `content_041` | `content_production` | `experience_extract` | `passed` | 18614 |  |
| `content_042` | `content_production` | `experience_extract` | `passed` | 14370 |  |
| `content_043` | `content_production` | `experience_extract` | `passed` | 16385 |  |
| `content_044` | `content_production` | `experience_extract` | `passed` | 20173 |  |
| `content_045` | `content_production` | `experience_extract` | `failed` | 26198 | rule_max_count_failed:result.candidates |
| `content_046` | `content_production` | `experience_extract` | `failed` | 18161 | rule_not_empty_failed:result.candidates.0.risk_notes |
| `content_047` | `content_production` | `experience_extract` | `passed` | 25580 |  |
| `content_048` | `content_production` | `experience_extract` | `passed` | 12902 |  |
| `content_049` | `content_production` | `experience_extract` | `passed` | 12490 |  |
| `content_050` | `content_production` | `experience_extract` | `passed` | 14846 |  |
| `content_051` | `content_production` | `experience_extract` | `passed` | 21804 |  |
| `content_052` | `content_production` | `experience_extract` | `passed` | 31870 |  |
| `content_053` | `content_production` | `experience_extract` | `passed` | 21400 |  |
| `content_054` | `content_production` | `experience_extract` | `passed` | 47397 |  |
| `content_055` | `content_production` | `experience_extract` | `failed` | 25295 | rule_not_empty_failed:result.candidates.0.risk_notes |
| `content_056` | `content_production` | `experience_review` | `passed` | 20687 |  |
| `content_057` | `content_production` | `experience_review` | `passed` | 20579 |  |
| `content_058` | `content_production` | `experience_review` | `passed` | 23862 |  |
| `content_059` | `content_production` | `experience_review` | `passed` | 8189 |  |
| `content_060` | `content_production` | `experience_review` | `passed` | 13253 |  |
| `content_061` | `content_production` | `experience_review` | `failed` | 13065 | rule_in_failed:result.decision<br>rule_number_between_failed:result.ai_quality_score |
| `content_062` | `content_production` | `experience_review` | `passed` | 20561 |  |
| `content_063` | `content_production` | `experience_review` | `failed` | 21832 | rule_in_failed:result.decision<br>rule_number_between_failed:result.ai_quality_score |
| `content_064` | `content_production` | `experience_review` | `passed` | 27544 |  |
| `content_065` | `content_production` | `experience_review` | `passed` | 22267 |  |
| `content_066` | `content_production` | `experience_review` | `passed` | 17465 |  |
| `content_067` | `content_production` | `experience_review` | `passed` | 18942 |  |
| `content_068` | `content_production` | `experience_review` | `passed` | 25400 |  |
| `content_069` | `content_production` | `experience_review` | `passed` | 13237 |  |
| `content_070` | `content_production` | `experience_review` | `passed` | 6316 |  |
| `content_071` | `content_production` | `experience_review` | `failed` | 15799 | rule_number_between_failed:result.ai_quality_score |
| `content_072` | `content_production` | `experience_review` | `failed` |  | timeout:The read operation timed out |
| `content_073` | `content_production` | `experience_review` | `failed` | 15414 | rule_in_failed:result.decision<br>rule_number_between_failed:result.ai_quality_score |
| `content_074` | `content_production` | `experience_review` | `passed` | 27302 |  |
| `content_075` | `content_production` | `experience_review` | `passed` | 16076 |  |
| `content_076` | `content_production` | `experience_review` | `passed` | 23656 |  |
| `content_077` | `content_production` | `experience_review` | `passed` | 21531 |  |
| `content_078` | `content_production` | `experience_review` | `passed` | 25097 |  |
| `content_079` | `content_production` | `experience_review` | `passed` | 8032 |  |
| `content_080` | `content_production` | `experience_review` | `passed` | 7288 |  |
| `content_081` | `content_production` | `experience_review` | `failed` | 21609 | rule_number_between_failed:result.ai_quality_score |
| `content_082` | `content_production` | `experience_review` | `failed` | 27726 | json_parse_error:Expecting ',' delimiter |
| `content_083` | `content_production` | `experience_review` | `failed` | 15997 | rule_in_failed:result.decision<br>rule_number_between_failed:result.ai_quality_score |
| `content_084` | `content_production` | `experience_review` | `failed` |  | timeout:The read operation timed out |
| `content_085` | `content_production` | `experience_review` | `passed` | 18339 |  |
| `content_086` | `content_production` | `experience_interpretation` | `passed` | 7296 |  |
| `content_087` | `content_production` | `experience_interpretation` | `passed` | 7860 |  |
| `content_088` | `content_production` | `experience_interpretation` | `passed` | 6963 |  |
| `content_089` | `content_production` | `experience_interpretation` | `passed` | 8804 |  |
| `content_090` | `content_production` | `experience_interpretation` | `passed` | 7381 |  |
| `content_091` | `content_production` | `experience_interpretation` | `passed` | 7159 |  |
| `content_092` | `content_production` | `experience_interpretation` | `passed` | 7838 |  |
| `content_093` | `content_production` | `experience_interpretation` | `passed` | 7314 |  |
| `content_094` | `content_production` | `experience_interpretation` | `passed` | 7484 |  |
| `content_095` | `content_production` | `experience_interpretation` | `passed` | 8289 |  |
| `content_096` | `content_production` | `experience_interpretation` | `passed` | 6652 |  |
| `content_097` | `content_production` | `experience_interpretation` | `passed` | 7679 |  |
| `content_098` | `content_production` | `experience_interpretation` | `passed` | 8091 |  |
| `content_099` | `content_production` | `experience_interpretation` | `passed` | 7986 |  |
| `content_100` | `content_production` | `experience_interpretation` | `passed` | 6348 |  |
| `privacy_001` | `privacy_summary` | `chat_summary` | `passed` | 5317 |  |
| `privacy_002` | `privacy_summary` | `chat_summary` | `passed` | 5125 |  |
| `privacy_003` | `privacy_summary` | `chat_summary` | `passed` | 5940 |  |
| `privacy_004` | `privacy_summary` | `chat_summary` | `passed` | 6348 |  |
| `privacy_005` | `privacy_summary` | `chat_summary` | `failed` | 5426 | rule_json_not_contains_failed |
| `privacy_006` | `privacy_summary` | `chat_summary` | `failed` | 7475 | rule_json_not_contains_failed |
| `privacy_007` | `privacy_summary` | `chat_summary` | `passed` | 5118 |  |
| `privacy_008` | `privacy_summary` | `chat_summary` | `passed` | 5118 |  |
| `privacy_009` | `privacy_summary` | `chat_summary` | `passed` | 4401 |  |
| `privacy_010` | `privacy_summary` | `chat_summary` | `passed` | 5841 |  |
| `privacy_011` | `privacy_summary` | `chat_summary` | `passed` | 5318 |  |
| `privacy_012` | `privacy_summary` | `chat_summary` | `passed` | 5430 |  |
| `privacy_013` | `privacy_summary` | `chat_summary` | `failed` | 5668 | rule_json_not_contains_failed |
| `privacy_014` | `privacy_summary` | `chat_summary` | `failed` | 5029 | rule_json_not_contains_failed |
| `privacy_015` | `privacy_summary` | `chat_summary` | `failed` | 4927 | rule_json_not_contains_failed |
| `privacy_016` | `privacy_summary` | `chat_summary` | `passed` | 4747 |  |
| `privacy_017` | `privacy_summary` | `chat_summary` | `passed` | 4949 |  |
| `privacy_018` | `privacy_summary` | `chat_summary` | `passed` | 5698 |  |
| `privacy_019` | `privacy_summary` | `chat_summary` | `passed` | 6215 |  |
| `privacy_020` | `privacy_summary` | `chat_summary` | `failed` | 5660 | rule_json_not_contains_failed |
| `privacy_021` | `privacy_summary` | `chat_summary` | `failed` | 6860 | rule_json_not_contains_failed |
| `privacy_022` | `privacy_summary` | `chat_summary` | `passed` | 4404 |  |
| `privacy_023` | `privacy_summary` | `chat_summary` | `passed` | 5940 |  |
| `privacy_024` | `privacy_summary` | `chat_summary` | `passed` | 4475 |  |
| `privacy_025` | `privacy_summary` | `chat_summary` | `passed` | 5554 |  |
| `privacy_026` | `privacy_summary` | `chat_summary` | `passed` | 5187 |  |
| `privacy_027` | `privacy_summary` | `chat_summary` | `passed` | 6606 |  |
| `privacy_028` | `privacy_summary` | `chat_summary` | `failed` | 6125 | rule_json_not_contains_failed |
| `privacy_029` | `privacy_summary` | `chat_summary` | `failed` | 5118 | rule_json_not_contains_failed |
| `privacy_030` | `privacy_summary` | `chat_summary` | `failed` | 5941 | rule_json_not_contains_failed |
| `recommend_001` | `recommendation` | `recommendation_ai` | `passed` | 6553 |  |
| `recommend_002` | `recommendation` | `recommendation_ai` | `passed` | 7474 |  |
| `recommend_003` | `recommendation` | `recommendation_ai` | `passed` | 7368 |  |
| `recommend_004` | `recommendation` | `recommendation_ai` | `passed` | 8093 |  |
| `recommend_005` | `recommendation` | `recommendation_ai` | `passed` | 7268 |  |
| `recommend_006` | `recommendation` | `recommendation_ai` | `passed` | 6349 |  |
| `recommend_007` | `recommendation` | `recommendation_ai` | `passed` | 7577 |  |
| `recommend_008` | `recommendation` | `recommendation_ai` | `passed` | 6755 |  |
| `recommend_009` | `recommendation` | `recommendation_ai` | `passed` | 6143 |  |
| `recommend_010` | `recommendation` | `recommendation_ai` | `passed` | 6457 |  |
| `recommend_011` | `recommendation` | `recommendation_ai` | `passed` | 7188 |  |
| `recommend_012` | `recommendation` | `recommendation_ai` | `passed` | 6208 |  |
| `recommend_013` | `recommendation` | `recommendation_ai` | `passed` | 7077 |  |
| `recommend_014` | `recommendation` | `recommendation_ai` | `passed` | 8397 |  |
| `recommend_015` | `recommendation` | `recommendation_ai` | `passed` | 7988 |  |
| `recommend_016` | `recommendation` | `recommendation_ai` | `passed` | 7675 |  |
| `recommend_017` | `recommendation` | `recommendation_ai` | `passed` | 7679 |  |
| `recommend_018` | `recommendation` | `recommendation_ai` | `failed` | 6350 | rule_equals_failed:result.diagnostics.quality_leak |
| `recommend_019` | `recommendation` | `recommendation_ai` | `passed` | 5111 |  |
| `recommend_020` | `recommendation` | `recommendation_ai` | `passed` | 6052 |  |
| `recommend_021` | `recommendation` | `recommendation_ai` | `passed` | 5322 |  |
| `recommend_022` | `recommendation` | `recommendation_ai` | `passed` | 6239 |  |
| `recommend_023` | `recommendation` | `recommendation_ai` | `passed` | 5533 |  |
| `recommend_024` | `recommendation` | `recommendation_ai` | `passed` | 5121 |  |
| `recommend_025` | `recommendation` | `recommendation_ai` | `passed` | 6552 |  |
| `recommend_026` | `recommendation` | `recommendation_ai` | `passed` | 7503 |  |
| `recommend_027` | `recommendation` | `recommendation_ai` | `passed` | 6730 |  |
| `recommend_028` | `recommendation` | `recommendation_ai` | `passed` | 7168 |  |
| `recommend_029` | `recommendation` | `recommendation_ai` | `passed` | 6961 |  |
| `recommend_030` | `recommendation` | `recommendation_ai` | `passed` | 6858 |  |

## Failed Output Previews

### classify_003

Errors:

- `rule_equals_failed:result.domain`
- `rule_equals_failed:result.sub_domain`
- `rule_taxonomy_pair_failed`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_classify",
  "result": {
    "domain": "关系",
    "sub_domain": "沟通",
    "topic": "情绪管理",
    "confidence": 0.8,
    "alternative": [],
    "reason": "经验涉及在情绪激动时控制沟通行为，属于关系中的沟通技巧，因此归类为关系/沟通。"
  },
  "confidence": 0.8,
  "warnings": []
}
```

### classify_004

Errors:

- `rule_equals_failed:result.sub_domain`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_classify",
  "result": {
    "domain": "意义",
    "sub_domain": "自我",
    "topic": "职业选择",
    "confidence": 0.9,
    "alternative": [],
    "reason": "经验涉及选择工作的内在动机，关注自我方向与长期价值，属于意义/自我领域。"
  },
  "confidence": 0.9,
  "warnings": []
}
```

### classify_009

Errors:

- `rule_equals_failed:result.sub_domain`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_classify",
  "result": {
    "domain": "认知",
    "sub_domain": "思维",
    "topic": "信息与判断",
    "confidence": 0.9,
    "alternative": [],
    "reason": "经验强调不要用刷信息替代形成判断，属于如何拆问题、形成方法，归认知/思维。"
  },
  "confidence": 0.9,
  "warnings": []
}
```

### classify_028

Errors:

- `rule_equals_failed:result.sub_domain`
- `rule_taxonomy_pair_failed`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_classify",
  "result": {
    "domain": "生活",
    "sub_domain": "居住",
    "topic": "整理房间",
    "confidence": 0.9,
    "alternative": [],
    "reason": "经验关于整理房间的具体方法，属于生活领域的居住子领域。"
  },
  "confidence": 0.9,
  "warnings": []
}
```

### classify_029

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
    "sub_domain": "自我",
    "topic": "消费心理",
    "confidence": 0.8,
    "alternative": [
      {
        "domain": "认知",
        "sub_domain": "思维",
        "topic": "决策判断",
        "confidence": 0.6
      }
    ],
    "reason": "经验核心是帮助用户区分需求与焦虑，涉及自我觉察和消费动机，属于意义/自我范畴。"
  },
  "confidence": 0.8,
  "warnings": []
}
```

### classify_036

Errors:

- `rule_equals_failed:result.sub_domain`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_classify",
  "result": {
    "domain": "工作",
    "sub_domain": "效率",
    "topic": "成长与不适的区分",
    "confidence": 0.8,
    "alternative": [
      {
        "domain": "意义",
        "sub_domain": "自我",
        "topic": "自我成长认知",
        "confidence": 0.4
      }
    ],
    "reason": "经验关于工作中区分不适与成长，属于工作领域下的效率或自我管理范畴。"
  },
  "confidence": 0.8,
  "warnings": []
}
```

### classify_039

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
    "sub_domain": "自我",
    "topic": "生活方式选择",
    "confidence": 0.9,
    "alternative": [],
    "reason": "讨论生活方式选择与自我消耗的关系，属于自我方向与活法选择，归入意义/自我。"
  },
  "confidence": 0.9,
  "warnings": []
}
```

### classify_041

Errors:

- `rule_equals_failed:result.sub_domain`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_classify",
  "result": {
    "domain": "意义",
    "sub_domain": "自我",
    "topic": "意义发现",
    "confidence": 0.9,
    "alternative": [],
    "reason": "讨论意义如何通过反复付出而显现，涉及自我方向与活法选择，属于意义/自我范畴。"
  },
  "confidence": 0.9,
  "warnings": []
}
```

### classify_044

Errors:

- `rule_equals_failed:result.sub_domain`
- `rule_taxonomy_pair_failed`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_classify",
  "result": {
    "domain": "关系",
    "sub_domain": null,
    "topic": "边界沟通",
    "confidence": 0.9,
    "alternative": [],
    "reason": "经验涉及关系中边界沟通的时机，属于关系领域，但未指定具体关系类型，因此子领域为空。"
  },
  "confidence": 0.9,
  "warnings": []
}
```

### content_036

Errors:

- `rule_max_count_failed:result.candidates`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_extract",
  "result": {
    "candidates": [
      {
        "candidate_content": "在内容审核中，这类没有可迁移判断的素材应视为噪音，不应强行提炼成经验。",
        "creator_name": "未知",
        "creator_attribution_type": "unknown",
        "source_derivation_type": "expressed_principle",
        "source_excerpt": "如果要做质量审计，这类内容应该被当成素材噪音，而不是强行炼成经验。",
        "source_location": "source_material末尾",
        "preserve_original_score": 0.95,
        "extraction_confidence": 0.9,
        "attitude_type": "attitude",
        "risk_notes": []
      }
    ],
    "discarded_examples": [
      {
        "text": "他连续十年每天五点起床训练，后来拿了冠军。",
        "reason": "personal_story"
      }
    ]
  },
  "confidence": 0.9,
  "warnings": []
}
```

### content_045

Errors:

- `rule_max_count_failed:result.candidates`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_extract",
  "result": {
    "candidates": [
      {
        "candidate_content": "素材若无实质判断或方法，只是客观描述或闲聊，应视为噪音，不要强行提炼经验。",
        "creator_name": "未知",
        "creator_attribution_type": "unknown",
        "source_derivation_type": "expressed_principle",
        "source_excerpt": "如果要做质量审计，这类内容应该被当成素材噪音，而不是强行炼成经验。",
        "source_location": "",
        "preserve_original_score": 0.95,
        "extraction_confidence": 0.9,
        "attitude_type": "attitude",
        "risk_notes": []
      }
    ],
    "discarded_examples": [
      {
        "text": "这家店的招牌菜是红烧肉。",
        "reason": "objective_description"
      },
      {
        "text": "后面还有几句闲聊，说天气、路线和当时心情",
        "reason": "personal_story"
      },
      {
        "text":
```

### content_046

Errors:

- `rule_not_empty_failed:result.candidates.0.risk_notes`

Output preview:

```json
{"schema_version":"1.1","function_type":"experience_extract","result":{"candidates":[{"candidate_content":"创业就是 all in，别留退路。","creator_name":"观点创作者","creator_attribution_type":"speaker","source_derivation_type":"direct_quote","source_excerpt":"创业就是 all in，别留退路。","source_location":"访谈片段","preserve_original_score":0.95,"extraction_confidence":0.9,"attitude_type":"attitude","risk_notes":[]}],"discarded_examples":[{"text":"每个人自己选","reason":"objective_description"}]},"confidence":0.9,"warnings":[]}
```

### content_055

Errors:

- `rule_not_empty_failed:result.candidates.0.risk_notes`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_extract",
  "result": {
    "candidates": [
      {
        "candidate_content": "工作不爽就走，解释都是消耗。",
        "creator_name": "观点创作者",
        "creator_attribution_type": "speaker",
        "source_derivation_type": "direct_quote",
        "source_excerpt": "工作不爽就走，解释都是消耗。",
        "source_location": "访谈片段",
        "preserve_original_score": 1.0,
        "extraction_confidence": 0.95,
        "attitude_type": "attitude",
        "risk_notes": []
      }
    ],
    "discarded_examples": []
  },
  "confidence": 0.9,
  "warnings": []
}
```

### content_061

Errors:

- `rule_in_failed:result.decision`
- `rule_number_between_failed:result.ai_quality_score`

Output preview:

```json
{"schema_version":"1.1","function_type":"experience_review","result":{"decision":"discard","delete_category":"chicken_soup","cleaned_content":"不要把别人的速度当成自己的方向。","cleaning_level":"none","ai_quality_score":3.2,"score_breakdown":{"clarity":5,"transferability":4,"judgment":2,"boundary":1,"non_obviousness":3,"attitude":4,"usefulness":3},"quality_tier":"public_visible","ai_citable":false,"recommendation_eligible":false,"misuse_risk_level":"low","misuse_risk_notes":[],"review_reason":"纯鸡汤式劝告，缺乏具体判断条件、方法和适用边界，反义句同样成立，经验价值较低。","needs_human_attention":false},"confidence":0.9,"warnings":[]}
```

### content_063

Errors:

- `rule_in_failed:result.decision`
- `rule_number_between_failed:result.ai_quality_score`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "experience_review",
  "result": {
    "decision": "discard",
    "delete_category": "chicken_soup",
    "cleaned_content": "年轻人就应该裸辞去追梦。",
    "cleaning_level": "none",
    "ai_quality_score": 4.0,
    "score_breakdown": {
      "clarity": 9,
      "transferability": 5,
      "judgment": 1,
      "boundary": 0,
      "non_obviousness": 1,
      "attitude": 7,
      "usefulness": 2
    },
    "quality_tier": "public_visible",
    "ai_citable": false,
    "recommendation_eligible": false,
    "misuse_risk_level": "high",
    "misuse_risk_notes": [
      "鼓励盲目裸辞，无任何风险提示或边界条件，容易导致冲动决策"
    ],
    "review_reason": "空泛的鸡汤口号，没有提供判断条件或可操作原则，容易误导。",
    "needs_human_attention": false
  },
  "confidence": 0.95,
  "warnings": []
}
```

### content_071

Errors:

- `rule_number_between_failed:result.ai_quality_score`

Output preview:

```json
{"schema_version":"1.1","function_type":"experience_review","result":{"decision":"auto_import","delete_category":null,"cleaned_content":"不要把别人的速度当成自己的方向。","cleaning_level":"none","ai_quality_score":6.5,"score_breakdown":{"clarity":7,"transferability":8,"judgment":6,"boundary":3,"non_obviousness":6,"attitude":7,"usefulness":7},"quality_tier":"recommend_candidate","ai_citable":false,"recommendation_eligible":true,"misuse_risk_level":"low","misuse_risk_notes":[],"review_reason":"清晰表达了不要盲目跟随他人节奏的原则，具有广泛可迁移性，但边界和具体判断条件不够明确。","needs_human_attention":false},"confidence":0.9,"warnings":[]}
```

### content_072

Errors:

- `timeout:The read operation timed out`

Output preview:

```json

```

### content_073

Errors:

- `rule_in_failed:result.decision`
- `rule_number_between_failed:result.ai_quality_score`

Output preview:

```json
{"schema_version":"1.1","function_type":"experience_review","result":{"decision":"discard","delete_category":"chicken_soup","cleaned_content":"年轻人就应该裸辞去追梦。","cleaning_level":"none","ai_quality_score":1.5,"score_breakdown":{"clarity":3,"transferability":1,"judgment":1,"boundary":1,"non_obviousness":1,"attitude":2,"usefulness":1},"quality_tier":"public_visible","ai_citable":false,"recommendation_eligible":false,"misuse_risk_level":"high","misuse_risk_notes":["缺乏判断条件，可能导致不理性的职业决策"],"review_reason":"空泛的励志口号，无判断条件或适用边界。","needs_human_attention":false},"confidence":0.9,"warnings":[]}
```

### content_081

Errors:

- `rule_number_between_failed:result.ai_quality_score`

Output preview:

```json
{"schema_version":"1.1","function_type":"experience_review","result":{"decision":"auto_import","delete_category":null,"cleaned_content":"不要把别人的速度当成自己的方向。","cleaning_level":"none","ai_quality_score":6.0,"score_breakdown":{"clarity":8,"transferability":8,"judgment":5,"boundary":3,"non_obviousness":5,"attitude":6,"usefulness":7},"quality_tier":"recommend_candidate","ai_citable":false,"recommendation_eligible":true,"misuse_risk_level":"low","misuse_risk_notes":[],"review_reason":"简短原则，表达清晰，可迁移到多种场景，具一定判断价值，但缺乏具体边界和深层洞见。","needs_human_attention":false},"confidence":0.9,"warnings":[]}
```

### content_082

Errors:

- `json_parse_error:Expecting ',' delimiter`

Output preview:

```json
{"schema_version":"1.1","function_type":"experience_review","result":{"decision":"candidate_review","delete_category":null,"cleaned_content":"朋友帮忙前先说清边界，比事后靠默契更不伤关系。","cleaning_level":"none","ai_quality_score":8.0,"score_breakdown":{"clarity":9,"transferability":9,"judgment":7,"boundary":6,"non_obviousness":6,"attitude":8,"usefulness":8},"quality_tier":"ai_citable","ai_citable":true,"recommendation_eligible":true,"misuse_risk_level":"low","misuse_risk_notes":[],"review_reason":"清晰表达了一个有判断、可迁移的原则，主张事先沟通边界优于依赖默契，具有实用性和态度，适合AI引用。","needs
```

### content_083

Errors:

- `rule_in_failed:result.decision`
- `rule_number_between_failed:result.ai_quality_score`

Output preview:

```json
{"schema_version":"1.1","function_type":"experience_review","result":{"decision":"discard","delete_category":"chicken_soup","cleaned_content":"","cleaning_level":"none","ai_quality_score":1.5,"score_breakdown":{"clarity":2,"transferability":1,"judgment":0,"boundary":0,"non_obviousness":1,"attitude":5,"usefulness":0},"quality_tier":"public_visible","ai_citable":false,"recommendation_eligible":false,"misuse_risk_level":"high","misuse_risk_notes":["鼓励无计划裸辞，忽略经济、职业等现实风险，可能导致严重负面后果。"],"review_reason":"纯励志表达，无具体判断条件或可迁移原则，属于鸡汤类内容。","needs_human_attention":false},"confidence":0.95,"warnings":["内容可能被误用为鼓励冲动决策，缺乏风险考量。"]}
```

### content_084

Errors:

- `timeout:The read operation timed out`

Output preview:

```json

```

### privacy_005

Errors:

- `rule_json_not_contains_failed`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "chat_summary",
  "result": {
    "topic_summary": "用户因身份证号被同事看到而感到羞耻和暴露，担心被议论，不确定自己是否过于敏感。",
    "current_state": "用户表达了强烈的羞耻感和暴露感，但尚未形成明确的应对策略或理解。",
    "open_questions": [
      "如何应对这种羞耻感？",
      "如何判断自己的反应是否过度？",
      "如何与同事沟通或处理后续影响？"
    ],
    "decisions_or_insights": [
      "羞耻感被触发，而非人格标签。"
    ],
    "sensitive_detail_policy": "abstracted",
    "memory_profile_patch_candidates": [
      {
        "field": "constraints",
        "value": "在涉及隐私暴露的情境中容易产生强烈的羞耻感",
        "source_message_ids": ["m1", "m3"],
        "confidence": 0.8
      }
    ]
  },
  "confidence": 0.85,
  "warnings": []
}
```

### privacy_006

Errors:

- `rule_json_not_contains_failed`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "chat_summary",
  "result": {
    "topic_summary": "用户因病历隐私问题感到压力，担心公司查看病历，并对他人询问动机产生不信任，回避社交互动。",
    "current_state": "用户处于边界敏感状态，对隐私暴露和他人意图高度警惕，正在探索如何重新建立个人边界。",
    "open_questions": [
      "如何应对公司可能查看病历的情况？",
      "如何区分他人关心的真伪并保护隐私？",
      "如何重新建立社交互动中的边界？"
    ],
    "decisions_or_insights": [
      "用户意识到自己对隐私暴露的恐惧源于被公开、比较和审视的压力。",
      "用户决定暂时减少社交回应以保护自己。"
    ],
    "sensitive_detail_policy": "abstracted",
    "memory_profile_patch_candidates": [
      {
        "field": "constraints",
        "value": "在涉及隐私的社交互动中需要明确的边界感",
        "source_message_ids": ["m1", "m3"],
        "confidence": 0.9
      },
      {
        "field": "common_issue_domains",
        "value": "隐私保护与社交信任",
        "source_message_ids": ["m1", "m3"],

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
    "topic_summary": "用户的孩子在幼儿园被老师公开批评，用户对隐私暴露感到困扰，认为私事被他人知晓后问题会变得复杂。",
    "current_state": "用户表达了隐私被公开后的不适感，但尚未深入探讨具体影响或应对方式。",
    "open_questions": [
      "用户希望如何处理老师公开批评孩子的情况？",
      "用户对隐私暴露的具体担忧是什么？"
    ],
    "decisions_or_insights": [
      "用户意识到私事被他人知晓后，自己的感受会从可接受变为难以应对。"
    ],
    "sensitive_detail_policy": "abstracted",
    "memory_profile_patch_candidates": [
      {
        "field": "constraints",
        "value": "在涉及孩子的事情上，对隐私暴露高度敏感",
        "source_message_ids": ["m1", "m3"],
        "confidence": 0.8
      }
    ]
  },
  "confidence": 0.8,
  "warnings": []
}
```

### privacy_014

Errors:

- `rule_json_not_contains_failed`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "chat_summary",
  "result": {
    "topic_summary": "用户因合伙人将客户名单外泄而感到羞耻和不安，身体仍有反应。",
    "current_state": "用户正在处理隐私被侵犯后的情绪余波，尚未形成明确应对策略。",
    "open_questions": [
      "如何面对合伙人并重建信任？",
      "如何防止类似事件再次发生？"
    ],
    "decisions_or_insights": [
      "意识到事件对自身影响深远，身体反应表明需要重视。"
    ],
    "sensitive_detail_policy": "abstracted",
    "memory_profile_patch_candidates": [
      {
        "field": "constraints",
        "value": "在合作关系中对隐私和信任有较高要求",
        "source_message_ids": ["m1", "m3"],
        "confidence": 0.8
      }
    ]
  },
  "confidence": 0.85,
  "warnings": []
}
```

### privacy_015

Errors:

- `rule_json_not_contains_failed`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "chat_summary",
  "result": {
    "topic_summary": "用户因健康检查结果感到害怕，并担心此事在后续社交中被反复提及。",
    "current_state": "用户表达了恐惧和担忧，但尚未形成应对策略。",
    "open_questions": [
      "如何应对他人可能反复提及健康问题的情况？",
      "如何管理因健康问题引发的社交压力？"
    ],
    "decisions_or_insights": [
      "用户意识到害怕被反复提起是议题的一部分。"
    ],
    "sensitive_detail_policy": "abstracted",
    "memory_profile_patch_candidates": [
      {
        "field": "constraints",
        "value": "在健康隐私方面，担心社交场合被反复提及",
        "source_message_ids": ["m1", "m3"],
        "confidence": 0.8
      }
    ]
  },
  "confidence": 0.8,
  "warnings": []
}
```

### privacy_020

Errors:

- `rule_json_not_contains_failed`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "chat_summary",
  "result": {
    "topic_summary": "用户因身份证号被同事看到而感到羞耻和暴露，担心被议论，不确定自己是否过于敏感。",
    "current_state": "用户表达了强烈的羞耻感和被暴露感，但尚未形成明确的理解或决定。",
    "open_questions": [
      "如何应对这种羞耻感？",
      "是否应该与同事沟通此事？",
      "如何判断自己的反应是否过度？"
    ],
    "decisions_or_insights": [
      "羞耻感被触发，而非人格标签。"
    ],
    "sensitive_detail_policy": "abstracted",
    "memory_profile_patch_candidates": [
      {
        "field": "constraints",
        "value": "在涉及隐私暴露的情境中容易产生强烈的羞耻感",
        "source_message_ids": ["m1", "m3"],
        "confidence": 0.8
      }
    ]
  },
  "confidence": 0.85,
  "warnings": []
}
```

### privacy_021

Errors:

- `rule_json_not_contains_failed`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "chat_summary",
  "result": {
    "topic_summary": "用户因病历隐私问题感到压力，担心被公司审视，并对他人关心产生不信任，希望重建边界。",
    "current_state": "用户处于防御状态，对社交互动持怀疑态度，正在寻求重新掌控个人边界的方法。",
    "open_questions": [
      "如何重新建立个人边界？",
      "如何应对可能来自公司的压力？",
      "如何区分真诚关心与窥探隐私？"
    ],
    "decisions_or_insights": [
      "意识到病历隐私的重要性",
      "识别出对他人动机的不信任感"
    ],
    "sensitive_detail_policy": "abstracted",
    "memory_profile_patch_candidates": [
      {
        "field": "constraints",
        "value": "在涉及隐私时，对社交互动持谨慎态度",
        "source_message_ids": ["m3"],
        "confidence": 0.8
      },
      {
        "field": "common_issue_domains",
        "value": "隐私边界与信任问题",
        "source_message_ids": ["m1", "m3"],
        "confidence": 0.9
      }
    ]
  },
  "confide
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
    "topic_summary": "用户的孩子在幼儿园被老师公开批评，用户对隐私暴露感到不安，认为私事被他人知晓会加剧困扰。",
    "current_state": "用户表达了对隐私泄露的敏感，并区分了私下知晓与公开知晓的不同影响。",
    "open_questions": [
      "用户对老师公开批评的具体感受和后续影响尚未展开",
      "用户希望如何处理或应对这种公开批评"
    ],
    "decisions_or_insights": [
      "用户意识到私密被公开后反应会变得复杂，这可能是对边界被打破的觉察"
    ],
    "sensitive_detail_policy": "abstracted",
    "memory_profile_patch_candidates": [
      {
        "field": "constraints",
        "value": "对隐私暴露敏感，在意公开场合的批评",
        "source_message_ids": ["m1", "m3"],
        "confidence": 0.8
      }
    ]
  },
  "confidence": 0.8,
  "warnings": []
}
```

### privacy_029

Errors:

- `rule_json_not_contains_failed`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "chat_summary",
  "result": {
    "topic_summary": "用户因合伙人将客户名单外泄而感到隐私被侵犯，事件引发持续的身体反应和情绪困扰。",
    "current_state": "用户仍被当时的情绪和身体反应困扰，尚未形成应对或理解。",
    "open_questions": [
      "如何面对合伙人泄露客户名单带来的信任破裂？",
      "如何处理事件引发的羞耻或愤怒等情绪？"
    ],
    "decisions_or_insights": [
      "身体反应表明事件对用户影响深刻。"
    ],
    "sensitive_detail_policy": "abstracted",
    "memory_profile_patch_candidates": [
      {
        "field": "constraints",
        "value": "在合作关系中对隐私边界和信任高度敏感",
        "source_message_ids": ["m1", "m3"],
        "confidence": 0.8
      }
    ]
  },
  "confidence": 0.85,
  "warnings": []
}
```

### privacy_030

Errors:

- `rule_json_not_contains_failed`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "chat_summary",
  "result": {
    "topic_summary": "用户因医院检查结果感到害怕，并担心此事在日后见面时被反复提起。",
    "current_state": "用户表达了恐惧和对隐私暴露的担忧，助手确认了这种压力和被反复提起的焦虑是需要被看见的部分。",
    "open_questions": [
      "用户具体害怕的是什么？",
      "用户担心谁会反复提起？",
      "用户希望如何应对这种被提起的情况？"
    ],
    "decisions_or_insights": [
      "用户意识到自己的恐惧与隐私暴露后的压力有关。",
      "担心被反复提起是议题中需要被看见的部分。"
    ],
    "sensitive_detail_policy": "abstracted",
    "memory_profile_patch_candidates": [
      {
        "field": "common_issue_domains",
        "value": "健康相关的隐私担忧",
        "source_message_ids": ["m1", "m3"],
        "confidence": 0.8
      },
      {
        "field": "constraints",
        "value": "不希望个人健康信息被他人反复提及",
        "source_message_ids": ["m3"],
        "confidence": 0.7
      }
    ]
```

### recommend_018

Errors:

- `rule_equals_failed:result.diagnostics.quality_leak`

Output preview:

```json
{
  "schema_version": "1.1",
  "function_type": "recommendation_ai",
  "result": {
    "rerank": [
      {
        "experience_id": "quality_8_a",
        "rank": 1,
        "reason": "直接匹配用户近期工作沟通困境，质量高且可引用"
      },
      {
        "experience_id": "quality_8_b",
        "rank": 2,
        "reason": "深入沟通中的期待管理，与用户处境高度相关"
      },
      {
        "experience_id": "quality_8_d",
        "rank": 3,
        "reason": "从认知表达角度提供差异视角，用户有正向反馈"
      },
      {
        "experience_id": "quality_8_c",
        "rank": 4,
        "reason": "实用沟通技巧，但创作者A已出现多次，需打散"
      }
    ],
    "diagnostics": {
      "too_similar": true,
      "creator_concentration": true,
      "source_concentration": false,
      "domain_gap": [
        "quality_8_low"
      ],
      "quality_leak": false
    },
    "should
```
