#!/usr/bin/env python3
"""Unit tests for the DeepSeek live eval runner."""

from __future__ import annotations

import unittest
from unittest.mock import Mock, sentinel, patch

import eval_deepseek_prompts as runner


class EvalRunnerNetworkTest(unittest.TestCase):
    @patch.object(runner.urllib.request, "build_opener")
    @patch.object(runner.urllib.request, "ProxyHandler")
    def test_default_opener_disables_system_proxy(self, mock_proxy_handler, mock_build_opener) -> None:
        mock_proxy_handler.return_value = sentinel.empty_proxy_handler
        mock_build_opener.return_value = sentinel.opener

        opener = runner.build_url_opener(use_system_proxy=False)

        mock_proxy_handler.assert_called_once_with({})
        mock_build_opener.assert_called_once_with(sentinel.empty_proxy_handler)
        self.assertIs(sentinel.opener, opener)

    @patch.object(runner.urllib.request, "build_opener")
    @patch.object(runner.urllib.request, "ProxyHandler")
    def test_system_proxy_mode_uses_default_opener(self, mock_proxy_handler, mock_build_opener) -> None:
        mock_build_opener.return_value = sentinel.opener

        opener = runner.build_url_opener(use_system_proxy=True)

        mock_proxy_handler.assert_not_called()
        mock_build_opener.assert_called_once_with()
        self.assertIs(sentinel.opener, opener)


class EvalRunnerRetryTest(unittest.TestCase):
    def test_empty_content_is_retried_once(self) -> None:
        prompt_packs = {
            "chat": {
                "system": "system",
                "developer": "developer",
                "output_schema": "{}",
                "user_template": "{{payload}}",
            }
        }
        item = {
            "id": "case_empty_then_ok",
            "function_type": "chat",
            "payload": {},
            "validator": lambda parsed: [],
        }
        valid_content = (
            '{"schema_version":"1.1","function_type":"chat",'
            '"result":{},"confidence":0.9,"warnings":[]}'
        )

        with patch.object(
            runner,
            "call_deepseek",
            Mock(
                side_effect=[
                    {
                        "elapsed_ms": 100,
                        "finish_reason": "stop",
                        "reasoning_content_len": 1000,
                        "usage": {},
                        "content": "",
                    },
                    {
                        "elapsed_ms": 120,
                        "finish_reason": "stop",
                        "reasoning_content_len": 0,
                        "usage": {},
                        "content": valid_content,
                    },
                ]
            ),
        ) as mock_call:
            result = runner.run_case(prompt_packs, "https://example.test", "key", "model", item, sentinel.opener, 1, 1)

        self.assertEqual("passed", result["status"])
        self.assertEqual(2, mock_call.call_count)
        self.assertIn("retry_after:empty_content", result["warnings"])


if __name__ == "__main__":
    unittest.main()
