import json

from fastapi import FastAPI
from fastapi.testclient import TestClient

from app.api import gateway
import app.services.llm as llm_module


class FakeLLM:
    def __init__(self, response):
        self.responses = response if isinstance(response, list) else [response]
        self.messages = None
        self.calls = 0

    async def chat(self, messages, stream=False, **kwargs):
        self.messages = messages
        self.kwargs = kwargs
        index = min(self.calls, len(self.responses) - 1)
        self.calls += 1
        return self.responses[index]


def make_client(fake_llm):
    llm_module.llm_service = fake_llm
    app = FastAPI()
    app.include_router(gateway.router, prefix="/api/v1/ai-gateway")
    return TestClient(app), fake_llm


def chat_payload():
    return {
        "function_type": "chat",
        "user_id": "user-1",
        "chat_message_id": "msg-1",
        "payload": {
            "user_id": "user-1",
            "message_id": "msg-1",
            "user_message": "我最近很想辞职，但又怕后悔",
            "session_state": "stable_topic",
            "scope": {"kind": "topic", "id": "topic-1"},
            "topic": {
                "id": "topic-1",
                "status": "active",
                "title": "工作里的不甘心",
                "domain": "work",
                "sub_domain": "communication",
                "topic": "辞职犹豫",
            },
            "recent_messages": [
                {"role": "user", "content": "我觉得每天都在硬撑"},
            ],
            "pre_classification": {
                "emotion_level": "medium",
                "user_intent": "decide",
                "risk_level": "high_decision",
                "risk_reasons": ["high_impact_decision"],
                "should_avoid_citation": False,
            },
            "candidate_experiences": [
                {
                    "experience_id": "exp-1",
                    "content": "先把触发你想离开的具体点写下来，再决定是不是离开。",
                    "creator_name": "某个认真生活的人",
                    "source_relation": "collected",
                    "visibility": "public",
                    "quality_tier": "ai_citable",
                    "source_reliability": "high",
                    "citation_policy": "card_allowed",
                    "relevance_reason": "与当前议题领域接近",
                }
            ],
            "context_flags": ["high_risk_decision"],
            "limits": {"max_reply_chars_soft": 500, "max_citation_cards": 1},
        },
    }


def test_gateway_chat_returns_structured_result_and_uses_prompt_contract():
    fake_response = json.dumps(
        {
            "schema_version": "1.1",
            "function_type": "chat",
            "result": {
                "reply_text": "先别急着把辞职变成唯一出口。可以先写下真正让你想离开的触发点。",
                "citations": [
                    {
                        "experience_id": "exp-1",
                        "usage_type": "card",
                        "show_card": True,
                        "citation_sentence": "有人在类似处境里会先看触发点。",
                        "reason_code": "high_relevance",
                        "strength": "weak",
                    }
                ],
                "note_suggestion": {"should_show": False, "suggested_text": None, "source_message_ids": []},
                "emotion_level": "medium",
                "risk_level": "high_decision",
                "reply_mode": "compare",
                "followup_question_count": 0,
                "internal_flags": [],
            },
            "confidence": 0.82,
            "warnings": [],
        },
        ensure_ascii=False,
    )
    client, fake_llm = make_client(FakeLLM(fake_response))

    response = client.post("/api/v1/ai-gateway/call", json=chat_payload())

    assert response.status_code == 200, response.text
    body = response.json()
    assert body["result"]["reply_text"].startswith("先别急")
    assert body["result"]["citations"][0]["experience_id"] == "exp-1"
    assert fake_llm.messages[0]["role"] == "system"
    assert "你是「年糕」" in fake_llm.messages[0]["content"]
    assert "输出必须是合法 JSON" in fake_llm.messages[0]["content"]
    assert fake_llm.messages[-1]["role"] == "user"
    assert "candidate_experiences" in fake_llm.messages[-1]["content"]
    assert fake_llm.kwargs["response_format"] == {"type": "json_object"}
    assert fake_llm.kwargs["temperature"] == 0.45


def test_gateway_chat_strips_markdown_json_fence():
    fake_response = """```json
{"schema_version":"1.1","function_type":"chat","result":{"reply_text":"我们先把这事放慢一点看。","citations":[],"note_suggestion":{"should_show":false,"suggested_text":null,"source_message_ids":[]},"emotion_level":"medium","risk_level":"normal","reply_mode":"hold","followup_question_count":0,"internal_flags":[]},"confidence":0.7,"warnings":[]}
```"""
    client, _ = make_client(FakeLLM(fake_response))

    response = client.post("/api/v1/ai-gateway/call", json=chat_payload())

    assert response.status_code == 200, response.text
    assert response.json()["result"]["reply_text"] == "我们先把这事放慢一点看。"


def test_gateway_rejects_unsupported_function_type():
    client, _ = make_client(FakeLLM("{}"))
    payload = chat_payload()
    payload["function_type"] = "experience_extract"

    response = client.post("/api/v1/ai-gateway/call", json=payload)

    assert response.status_code == 400


def test_gateway_rejects_empty_model_reply():
    fake_response = json.dumps(
        {
            "schema_version": "1.1",
            "function_type": "chat",
            "result": {
                "reply_text": " ",
                "citations": [],
                "note_suggestion": {"should_show": False, "suggested_text": None, "source_message_ids": []},
                "emotion_level": "low",
                "risk_level": "normal",
                "reply_mode": "hold",
                "followup_question_count": 0,
                "internal_flags": [],
            },
            "confidence": 0.3,
            "warnings": [],
        }
    )
    client, _ = make_client(FakeLLM(fake_response))

    response = client.post("/api/v1/ai-gateway/call", json=chat_payload())

    assert response.status_code == 502


def rewrite_payload():
    return {
        "function_type": "experience_rewrite",
        "user_id": "user-1",
        "payload": {
            "user_id": "user-1",
            "source": "manual_note",
            "raw_text": "我发现我不是怕换工作，是怕再次证明自己选错了",
            "source_message_ids": [],
            "default_visibility": "public",
            "user_selected_domain": "meaning",
            "user_selected_sub_domain": "self",
            "topic_context": "换工作的犹豫",
        },
    }


def test_gateway_rewrite_returns_structured_result_and_prompt_contract():
    fake_response = json.dumps(
        {
            "schema_version": "1.0",
            "function_type": "experience_rewrite",
            "result": {
                "can_rewrite": True,
                "content": "我不是怕换工作，是怕又一次证明自己选错了。",
                "domain": "meaning",
                "sub_domain": "self",
                "topic": "选择焦虑",
                "rewrite_level": "light",
                "source_preservation": "high",
                "needs_user_edit": False,
                "reason": "保留原判断，去掉重复表达",
            },
            "confidence": 0.88,
            "warnings": [],
        },
        ensure_ascii=False,
    )
    client, fake_llm = make_client(FakeLLM(fake_response))

    response = client.post("/api/v1/ai-gateway/call", json=rewrite_payload())

    assert response.status_code == 200, response.text
    body = response.json()
    assert body["result"]["content"] == "我不是怕换工作，是怕又一次证明自己选错了。"
    assert body["result"]["domain"] == "meaning"
    assert "不替用户发明没有表达过的结论" in fake_llm.messages[0]["content"]
    assert fake_llm.kwargs["response_format"] == {"type": "json_object"}
    assert fake_llm.kwargs["temperature"] == 0.2


def test_gateway_rewrite_retries_once_when_content_over_100_chars():
    overlong = json.dumps(
        {
            "schema_version": "1.0",
            "function_type": "experience_rewrite",
            "result": {
                "can_rewrite": True,
                "content": "年" * 101,
                "domain": "meaning",
                "sub_domain": "self",
                "topic": "选择焦虑",
                "rewrite_level": "medium",
                "source_preservation": "medium",
                "needs_user_edit": True,
                "reason": "太长",
            },
            "confidence": 0.5,
            "warnings": [],
        },
        ensure_ascii=False,
    )
    valid = json.dumps(
        {
            "schema_version": "1.0",
            "function_type": "experience_rewrite",
            "result": {
                "can_rewrite": True,
                "content": "先承认自己怕选错，再决定要不要换工作。",
                "domain": "meaning",
                "sub_domain": "self",
                "topic": "选择焦虑",
                "rewrite_level": "light",
                "source_preservation": "high",
                "needs_user_edit": False,
                "reason": "压缩到 100 字以内",
            },
            "confidence": 0.8,
            "warnings": [],
        },
        ensure_ascii=False,
    )
    client, fake_llm = make_client(FakeLLM([overlong, valid]))

    response = client.post("/api/v1/ai-gateway/call", json=rewrite_payload())

    assert response.status_code == 200, response.text
    assert response.json()["result"]["content"] == "先承认自己怕选错，再决定要不要换工作。"
    assert fake_llm.calls == 2


def test_gateway_rewrite_accepts_cannot_rewrite():
    fake_response = json.dumps(
        {
            "schema_version": "1.0",
            "function_type": "experience_rewrite",
            "result": {
                "can_rewrite": False,
                "content": "",
                "domain": "",
                "sub_domain": "",
                "topic": "",
                "rewrite_level": "reject",
                "source_preservation": "high",
                "needs_user_edit": False,
                "reason": "原文只有情绪，没有可复用经验",
            },
            "confidence": 0.73,
            "warnings": [],
        },
        ensure_ascii=False,
    )
    client, _ = make_client(FakeLLM(fake_response))

    response = client.post("/api/v1/ai-gateway/call", json=rewrite_payload())

    assert response.status_code == 200, response.text
    assert response.json()["result"]["can_rewrite"] is False
