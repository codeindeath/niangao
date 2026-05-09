import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.api import chat, review, translate
from app.core.config import settings
from app.services.llm import LLMService
import app.services.llm as llm_module

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info("年糕 AI 服务启动中...")
    llm_module.llm_service = LLMService()
    logger.info("年糕 AI 服务就绪")
    yield
    logger.info("年糕 AI 服务关闭")
    if llm_module.llm_service:
        await llm_module.llm_service.close()


app = FastAPI(title="年糕 AI Service", version="0.1.0", lifespan=lifespan)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(chat.router, prefix="/api/v1/chat", tags=["chat"])
app.include_router(review.router, prefix="/api/v1", tags=["review"])
app.include_router(translate.router, prefix="/api/v1", tags=["translate"])


@app.get("/health")
async def health():
    return {"status": "ok"}
