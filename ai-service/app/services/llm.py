"""DeepSeek LLM 服务"""
import logging
from typing import AsyncIterator, Dict, List

from openai import AsyncOpenAI

from app.core.config import settings

logger = logging.getLogger(__name__)


class LLMService:
    def __init__(self):
        self.client = AsyncOpenAI(
            api_key=settings.deepseek_api_key,
            base_url=settings.deepseek_base_url,
        )
        self.model = settings.deepseek_model

    async def chat(
        self,
        messages: List[Dict],
        stream: bool = False,
    ):
        """调用 DeepSeek 对话 API"""
        if stream:
            return self._chat_stream(messages)
        else:
            response = await self.client.chat.completions.create(
                model=self.model,
                messages=messages,
                temperature=0.7,
                max_tokens=1024,
            )
            return response.choices[0].message.content

    async def _chat_stream(self, messages: List[Dict]) -> AsyncIterator[str]:
        """流式对话"""
        stream = await self.client.chat.completions.create(
            model=self.model,
            messages=messages,
            temperature=0.7,
            max_tokens=1024,
            stream=True,
        )
        async for chunk in stream:
            if chunk.choices[0].delta.content:
                yield chunk.choices[0].delta.content

    async def generate_interpretation(self, content: str, domain: str) -> str:
        """AI 生成经验解读"""
        prompt = f"""你是一个经验整理助手。用户写了一条经验，请帮他把这条经验扩展成结构化的解读。

经验内容：{content}
领域：{domain}

请按以下格式输出（500字以内）：
- 背景：这条经验为什么重要，从何而来
- 如何执行：具体的操作步骤
- 适用场景：什么情况下这条经验最有用

直接输出解读内容，不要前置说明。"""

        response = await self.client.chat.completions.create(
            model=self.model,
            messages=[{"role": "user", "content": prompt}],
            temperature=0.5,
            max_tokens=600,
        )
        return response.choices[0].message.content

    async def get_embedding(self, text: str) -> List[float]:
        """获取文本向量"""
        response = await self.client.embeddings.create(
            model=settings.embedding_model,
            input=text,
        )
        return response.data[0].embedding
