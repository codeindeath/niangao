"""DeepSeek LLM 服务"""
import logging
from typing import AsyncIterator, Dict, List, Optional

from openai import AsyncOpenAI

from app.core.config import settings

logger = logging.getLogger(__name__)

# Module-level singleton (set by main.py lifespan)
llm_service: "LLMService" = None


class LLMService:
    def __init__(self):
        self.client = AsyncOpenAI(
            api_key=settings.deepseek_api_key,
            base_url=settings.deepseek_base_url,
        )

    async def close(self):
        await self.client.close()

    async def chat(
        self,
        messages: List[Dict],
        stream: bool = False,
        temperature: float = 0.7,
        max_tokens: int = 1024,
        response_format: Optional[Dict] = None,
    ):
        if stream:
            return self._chat_stream(messages)
        kwargs = {}
        if response_format is not None:
            kwargs["response_format"] = response_format
        response = await self.client.chat.completions.create(
            model=settings.deepseek_model,
            messages=messages,
            temperature=temperature,
            max_tokens=max_tokens,
            **kwargs,
        )
        return response.choices[0].message.content

    async def _chat_stream(self, messages: List[Dict]) -> AsyncIterator[str]:
        stream = await self.client.chat.completions.create(
            model=settings.deepseek_model,
            messages=messages,
            temperature=0.7,
            max_tokens=1024,
            stream=True,
        )
        async for chunk in stream:
            if chunk.choices[0].delta.content:
                yield chunk.choices[0].delta.content

    async def generate_interpretation(self, content: str, domain: str) -> str:
        prompt = f"""你是一个经验整理助手。请将以下经验扩展成结构化解读。

经验内容：{content}
领域：{domain}

按此格式输出（300字以内）：
- 背景：这条经验为什么重要
- 如何执行：具体操作步骤
- 适用场景：什么时候最有用

直接输出解读，不要前置说明。"""

        response = await self.client.chat.completions.create(
            model=settings.deepseek_model,
            messages=[{"role": "user", "content": prompt}],
            temperature=0.5,
            max_tokens=400,  # ~280-300 Chinese chars + format overhead
        )
        result = response.choices[0].message.content
        # 硬截断：确保不超过 300 字
        runes = list(result)
        if len(runes) > 300:
            result = ''.join(runes[:300])
        return result
