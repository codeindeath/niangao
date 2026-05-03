"""对话 API"""
import logging
from typing import Dict, List, Optional

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from sse_starlette.sse import EventSourceResponse

from app.core.config import settings
from app.core.prompts import build_system_prompt, build_chat_messages
import app.services.llm as llm_module
import app.services.embedding as emb_module

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
    try:
        # 关键词检索用户经验
        experiences = await emb_module.search_similar_experiences(
            query_text=req.message,
            user_id=req.user_id,
            limit=settings.max_context_experiences,
        )

        system_prompt = build_system_prompt(experiences)
        messages = build_chat_messages(system_prompt, req.history, req.message)
        response = await llm_module.llm_service.chat(messages, stream=False)

        return {
            "reply": response,
            "referenced_experience_ids": [e.get("id") for e in experiences],
        }
    except Exception as e:
        logger.error(f"Chat error: {e}")
        raise HTTPException(status_code=500, detail="对话服务暂时不可用")


@router.post("/stream")
async def stream_message(req: ChatRequest):
    async def event_generator():
        try:
            experiences = await emb_module.search_similar_experiences(
                query_text=req.message,
                user_id=req.user_id,
                limit=settings.max_context_experiences,
            )

            system_prompt = build_system_prompt(experiences)
            messages = build_chat_messages(system_prompt, req.history, req.message)

            full = ""
            async for token in llm_module.llm_service.chat(messages, stream=True):
                full += token
                yield {"data": token}
            yield {"event": "references", "data": str([e.get("id") for e in experiences])}
        except Exception as e:
            yield {"event": "error", "data": str(e)}

    return EventSourceResponse(event_generator())


@router.post("/generate-interpretation")
async def generate_interpretation(req: InterpretationRequest):
    try:
        interpretation = await llm_module.llm_service.generate_interpretation(
            req.content, req.domain
        )
        return {"interpretation": interpretation}
    except Exception as e:
        raise HTTPException(status_code=500, detail="生成解读失败")
