"""
AI Service 单元测试 — Prompt 构建、消息构建、配置验证
"""
import pytest
from app.core.prompts import build_chat_system_prompt, build_chat_messages, CHAT_SYSTEM_PROMPT
from app.core.config import Settings


class TestSystemPrompt:
    """验证人本主义 Prompt 的构建逻辑"""

    def test_prompt_contains_humanistic_principles(self):
        """Prompt 必须包含人本主义核心原则"""
        prompt = build_chat_system_prompt([])
        assert "先理解" in prompt
        assert "引导" in prompt
        assert "简洁温暖" in prompt
        assert "避免奉承" in prompt or "不要强行" in prompt

    def test_prompt_handles_empty_experiences(self):
        """用户没有收藏经验时，不应报错"""
        prompt = build_chat_system_prompt([])
        assert "还没有收藏" in prompt or "没有" in prompt
        assert "{experiences_context}" not in prompt  # 不应该有未填充的占位符

    def test_prompt_formats_experiences(self):
        """经验应该以编号列表形式出现"""
        experiences = [
            {"domain": "career", "content": "接到任务先确认 deadline"},
            {"domain": "life", "content": "每周留 2 小时做复盘"},
        ]
        prompt = build_chat_system_prompt(experiences)
        assert "1. [career] 接到任务先确认 deadline" in prompt
        assert "2. [life] 每周留 2 小时做复盘" in prompt

    def test_prompt_template_is_unchanged(self):
        """确保核心 Prompt 模板基本结构不变"""
        assert "你是「年糕」" in CHAT_SYSTEM_PROMPT
        assert "{experiences_context}" in CHAT_SYSTEM_PROMPT
        assert "收藏" in CHAT_SYSTEM_PROMPT


class TestChatMessages:
    """验证对话消息构建"""

    def test_builds_correct_message_structure(self):
        """消息列表必须以 system 开头，user 结尾"""
        messages = build_chat_messages(
            system_prompt="test system prompt",
            history=[],
            user_message="你好",
        )
        assert len(messages) == 2
        assert messages[0]["role"] == "system"
        assert messages[0]["content"] == "test system prompt"
        assert messages[1]["role"] == "user"
        assert messages[1]["content"] == "你好"

    def test_includes_history(self):
        """历史消息应该包含在 system 和当前 user 之间"""
        history = [
            {"role": "user", "content": "你好"},
            {"role": "assistant", "content": "你好，有什么可以帮你的？"},
            {"role": "user", "content": "我有点累"},
            {"role": "assistant", "content": "能和我说说发生了什么吗？"},
        ]
        messages = build_chat_messages(
            system_prompt="test",
            history=history,
            user_message="工作压力很大",
        )
        # system + 4 history + 1 current user = 6
        assert len(messages) == 6
        assert messages[0]["role"] == "system"
        assert messages[1]["role"] == "user"
        assert messages[1]["content"] == "你好"
        assert messages[5]["role"] == "user"
        assert messages[5]["content"] == "工作压力很大"

    def test_empty_history(self):
        """空历史不会报错"""
        messages = build_chat_messages(
            system_prompt="test",
            history=[],
            user_message="test",
        )
        assert len(messages) == 2


class TestSettings:
    """验证配置默认值"""

    def test_default_values(self):
        settings = Settings()
        assert settings.deepseek_model == "deepseek-chat"
        assert settings.max_daily_chat_rounds == 50
        assert settings.max_context_experiences == 5

    def test_deepseek_base_url_default(self):
        settings = Settings()
        assert "deepseek.com" in settings.deepseek_base_url


class TestPromptEdgeCases:
    """边界情况"""

    def test_experiences_with_missing_fields(self):
        """经验缺少字段时不应崩溃"""
        experiences = [
            {"content": "只有 content"},
            {"domain": "career", "content": ""},
            {},
        ]
        prompt = build_chat_system_prompt(experiences)
        assert prompt is not None
        assert isinstance(prompt, str)

    def test_very_long_experience_content(self):
        """极长的经验内容不破坏 Prompt 格式"""
        experiences = [
            {"domain": "career", "content": "很" * 100},
        ]
        prompt = build_chat_system_prompt(experiences)
        assert "很" * 100 in prompt
        assert len(prompt) < 10000  # Prompt 不应过长

    def test_special_characters_in_experiences(self):
        """特殊字符不破坏 Prompt"""
        experiences = [
            {"domain": "career", "content": 'test <script>alert("xss")</script>'},
            {"domain": "life", "content": "test\nwith\nnewlines"},
        ]
        prompt = build_chat_system_prompt(experiences)
        assert "<script>" in prompt
        assert "\n" in prompt
