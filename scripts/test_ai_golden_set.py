#!/usr/bin/env python3
"""Structural tests for Niangao AI golden eval set."""

from __future__ import annotations

import json
from collections import Counter
import re
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
GOLDEN_DIR = ROOT / "docs/product/ai-prompt-eval/golden-set"

EXPECTED_COUNTS = {
    "chat.jsonl": 50,
    "content-production.jsonl": 100,
    "classification.jsonl": 60,
    "privacy-summary.jsonl": 30,
    "recommendation.jsonl": 30,
}

ALLOWED_FUNCTION_TYPES = {
    "chat",
    "chat_summary",
    "translation_normalization",
    "experience_extract",
    "experience_review",
    "experience_classify",
    "experience_interpretation",
    "recommendation_ai",
}

ENUM_LIKE_TEXTS = {
    "expressed_principle",
    "recommend_candidate",
    "public_visible",
    "high_trust",
    "chat_citation_positive",
    "search_click",
    "high_decision",
    "record_insight",
    "card_allowed",
}


class GoldenSetStructureTest(unittest.TestCase):
    def load_jsonl(self, filename: str) -> list[dict]:
        path = GOLDEN_DIR / filename
        self.assertTrue(path.exists(), f"missing golden set file: {path}")
        rows = []
        for lineno, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
            self.assertTrue(line.strip(), f"{filename}:{lineno} blank line")
            rows.append(json.loads(line))
        return rows

    def test_expected_files_and_counts(self) -> None:
        for filename, expected_count in EXPECTED_COUNTS.items():
            with self.subTest(filename=filename):
                rows = self.load_jsonl(filename)
                self.assertEqual(expected_count, len(rows))

    def test_case_ids_are_unique_and_prefixed(self) -> None:
        seen: set[str] = set()
        for filename in EXPECTED_COUNTS:
            for row in self.load_jsonl(filename):
                case_id = row.get("case_id")
                self.assertIsInstance(case_id, str)
                self.assertRegex(case_id, r"^[a-z][a-z0-9_]+_[0-9]{3}$")
                self.assertNotIn(case_id, seen)
                seen.add(case_id)

    def test_cases_have_payload_and_expectations(self) -> None:
        for filename in EXPECTED_COUNTS:
            for row in self.load_jsonl(filename):
                with self.subTest(case_id=row.get("case_id")):
                    self.assertIn(row.get("function_type"), ALLOWED_FUNCTION_TYPES)
                    self.assertIsInstance(row.get("category"), str)
                    self.assertIsInstance(row.get("tags"), list)
                    self.assertIsInstance(row.get("payload"), dict)
                    self.assertIsInstance(row.get("rules"), list)
                    self.assertGreaterEqual(len(row["rules"]), 2)
                    for rule in row["rules"]:
                        self.assertIsInstance(rule.get("type"), str)

    def test_chat_cases_are_contextual_not_template_snippets(self) -> None:
        rows = self.load_jsonl("chat.jsonl")
        contextual = [row for row in rows if row["payload"].get("recent_messages")]
        colloquial_markers = ("其实", "就是", "有点", "唉", "吧", "老是", "昨晚", "今天", "不知道", "然后", "但是", "可能")
        colloquial = [
            row
            for row in rows
            if any(marker in row["payload"].get("user_message", "") for marker in colloquial_markers)
        ]
        long_enough = [row for row in rows if len(row["payload"].get("user_message", "")) >= 24]
        self.assertGreaterEqual(len(contextual), 30)
        self.assertGreaterEqual(len(colloquial), 35)
        self.assertGreaterEqual(len(long_enough), 40)

    def test_content_cases_include_messy_real_source_material(self) -> None:
        rows = self.load_jsonl("content-production.jsonl")
        source_rows = [
            row
            for row in rows
            if row["function_type"] in {"translation_normalization", "experience_extract"}
        ]
        long_sources = [
            row
            for row in source_rows
            if len(row["payload"].get("source_material", "")) >= 120
        ]
        noisy_sources = [
            row
            for row in source_rows
            if any(token in row["payload"].get("source_material", "") for token in ("00:", "嗯", "……", "评论区", "访谈", "书里", "弹幕", "Q:"))
        ]
        self.assertGreaterEqual(len(long_sources), 35)
        self.assertGreaterEqual(len(noisy_sources), 25)

    def test_privacy_summary_cases_have_conversation_shape(self) -> None:
        rows = self.load_jsonl("privacy-summary.jsonl")
        multi_user_turns = [
            row
            for row in rows
            if sum(1 for message in row["payload"].get("messages", []) if message.get("role") == "user") >= 2
        ]
        self.assertGreaterEqual(len(multi_user_turns), 24)

    def test_recommendation_cases_include_behavior_signals(self) -> None:
        rows = self.load_jsonl("recommendation.jsonl")
        with_signals = [
            row
            for row in rows
            if row["payload"].get("user_context", {}).get("behavior_signals")
            or row["payload"].get("user_context", {}).get("recent_events")
        ]
        self.assertGreaterEqual(len(with_signals), 24)

    def test_natural_language_payloads_are_not_over_reused(self) -> None:
        counter: Counter[str] = Counter()

        def walk(value):
            if isinstance(value, str):
                text = value.strip()
                if len(text) >= 16 and text not in ENUM_LIKE_TEXTS:
                    counter[text] += 1
            elif isinstance(value, dict):
                for child in value.values():
                    walk(child)
            elif isinstance(value, list):
                for child in value:
                    walk(child)

        for filename in EXPECTED_COUNTS:
            for row in self.load_jsonl(filename):
                walk(row["payload"])

        overused = {text: count for text, count in counter.items() if count > 8}
        self.assertEqual({}, overused)

    def test_long_chinese_fragments_are_not_over_reused(self) -> None:
        fragments: Counter[str] = Counter()

        def walk(value):
            if isinstance(value, str):
                for size in (14, 18, 22):
                    for index in range(0, max(0, len(value) - size + 1)):
                        fragment = value[index : index + size]
                        if re.search(r"[\u4e00-\u9fff]", fragment):
                            fragments[fragment] += 1
            elif isinstance(value, dict):
                for child in value.values():
                    walk(child)
            elif isinstance(value, list):
                for child in value:
                    walk(child)

        for filename in EXPECTED_COUNTS:
            for row in self.load_jsonl(filename):
                walk(row["payload"])

        overused = {fragment: count for fragment, count in fragments.items() if count > 16}
        self.assertEqual({}, overused)


if __name__ == "__main__":
    unittest.main()
