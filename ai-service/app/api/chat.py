"""对话 API — 由 Go 后端编排调用"""
import logging
from typing import Dict, List, Optional

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

from app.core.config import settings
from app.core.prompts import build_chat_system_prompt, build_chat_messages
import app.services.llm as llm_module

logger = logging.getLogger(__name__)
router = APIRouter()


class BookmarkedExp(BaseModel):
    id: str
    content: str
    domain: str


class ChatRequest(BaseModel):
    message: str
    user_id: str
    history: List[Dict] = []
    bookmarked_experiences: List[BookmarkedExp] = []


class InterpretationRequest(BaseModel):
    content: str
    domain: str


@router.post("/send")
async def send_message(req: ChatRequest):
    """由 Go 后端调用。接收完整上下文（历史 + 收藏经验），返回 AI 回复。"""
    try:
        # 构建系统提示词（收藏经验 + 对话历史）
        bookmarks_dict = [b.model_dump() for b in req.bookmarked_experiences]
        system_prompt = build_chat_system_prompt(bookmarks_dict, req.history)

        # 组装消息
        messages = build_chat_messages(system_prompt, req.history, req.message)
        response = await llm_module.llm_service.chat(messages, stream=False)

        # 引用哪些经验由 LLM 决定——我们不在代码层硬匹配
        return {
            "reply": response,
            "referenced_experience_ids": [
                b.id for b in req.bookmarked_experiences
            ],
        }
    except Exception as e:
        logger.error(f"Chat error: {e}")
        raise HTTPException(status_code=500, detail="对话服务暂时不可用")


@router.post("/generate-interpretation")
async def generate_interpretation(req: InterpretationRequest):
    try:
        interpretation = await llm_module.llm_service.generate_interpretation(
            req.content, req.domain
        )
        return {"interpretation": interpretation}
    except Exception as e:
        raise HTTPException(status_code=500, detail="生成解读失败")
