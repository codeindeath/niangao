"""
对话 API — 年糕 AI 聊天
"""
import logging
from typing import List, Dict, Optional

from fastapi import APIRouter, Depends, HTTPException
from pydantic import BaseModel
from sse_starlette.sse import EventSourceResponse

from app.core.prompts import build_system_prompt, build_chat_messages
from app.main import llm_service, embedding_service

logger = logging.getLogger(__name__)

router = APIRouter()


class ChatRequest(BaseModel):
    message: str
    conversation_id: Optional[str] = None
    history: List[Dict] = []
    user_id: str


class InterpretationRequest(BaseModel):
    content: str
    domain: str


@router.post("/send")
async def send_message(req: ChatRequest):
    """非流式对话"""
    try:
        # 1. 获取用户消息的向量
        query_embedding = await llm_service.get_embedding(req.message)

        # 2. 检索相关经验
        experiences = await embedding_service.search_similar(
            query_embedding,
            req.user_id,
            limit=5,
        )

        # 3. 构建 Prompt
        system_prompt = build_system_prompt(experiences)
        messages = build_chat_messages(system_prompt, req.history, req.message)

        # 4. 调用 LLM
        response = await llm_service.chat(messages, stream=False)

        # 5. 提取引用的经验 ID
        referenced_ids = [e.get("id") for e in experiences if e.get("content", "") in response]

        return {
            "reply": response,
            "referenced_experience_ids": referenced_ids,
        }
    except Exception as e:
        logger.error(f"Chat error: {e}")
        raise HTTPException(status_code=500, detail="对话服务暂时不可用")


@router.post("/stream")
async def stream_message(req: ChatRequest):
    """流式对话 (SSE)"""
    async def event_generator():
        try:
            # 检索经验
            query_embedding = await llm_service.get_embedding(req.message)
            experiences = await embedding_service.search_similar(
                query_embedding, req.user_id, limit=5,
            )

            system_prompt = build_system_prompt(experiences)
            messages = build_chat_messages(system_prompt, req.history, req.message)

            full_response = ""
            async for token in llm_service.chat(messages, stream=True):
                full_response += token
                yield {"data": token}

            # 最后发送引用的经验 ID
            yield {"event": "references", "data": str(
                [e.get("id") for e in experiences]
            )}

        except Exception as e:
            logger.error(f"Stream error: {e}")
            yield {"event": "error", "data": str(e)}

    return EventSourceResponse(event_generator())


@router.post("/generate-interpretation")
async def generate_interpretation(req: InterpretationRequest):
    """AI 生成经验解读"""
    try:
        interpretation = await llm_service.generate_interpretation(
            req.content, req.domain
        )
        return {"interpretation": interpretation}
    except Exception as e:
        logger.error(f"Generate interpretation error: {e}")
        raise HTTPException(status_code=500, detail="生成解读失败")
