"""经验 AI 辅助 API"""
import logging

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

from app.main import llm_service

logger = logging.getLogger(__name__)

router = APIRouter()


class GenerateEmbeddingRequest(BaseModel):
    experience_id: str
    content: str


@router.post("/generate-embedding")
async def generate_embedding(req: GenerateEmbeddingRequest):
    """为经验生成向量并索引"""
    try:
        embedding = await llm_service.get_embedding(req.content)
        from app.main import embedding_service
        await embedding_service.index_experience(req.experience_id, embedding)
        return {"status": "ok", "dimensions": len(embedding)}
    except Exception as e:
        logger.error(f"Generate embedding error: {e}")
        raise HTTPException(status_code=500, detail="向量生成失败")
