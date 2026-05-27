from app.services import llm


def test_llm_service_configures_provider_timeout(monkeypatch):
    created_kwargs = {}

    class FakeAsyncOpenAI:
        def __init__(self, **kwargs):
            created_kwargs.update(kwargs)

        async def close(self):
            pass

    monkeypatch.setattr(llm, "AsyncOpenAI", FakeAsyncOpenAI)
    monkeypatch.setattr(llm.settings, "deepseek_api_key", "test-key")
    monkeypatch.setattr(llm.settings, "deepseek_base_url", "https://deepseek.test")
    monkeypatch.setattr(llm.settings, "deepseek_timeout_seconds", 12.5, raising=False)

    llm.LLMService()

    assert created_kwargs["timeout"] == 12.5
