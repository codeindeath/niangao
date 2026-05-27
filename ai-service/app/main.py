import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.api import gateway
from app.core.config import settings
from app.middleware.request_id import RequestIDMiddleware
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

app.add_middleware(RequestIDMiddleware)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

def register_api_routers(fastapi_app: FastAPI, enable_legacy_ai_routes: bool = False) -> None:
    fastapi_app.include_router(gateway.router, prefix="/api/v1/ai-gateway", tags=["ai-gateway"])
    if not enable_legacy_ai_routes:
        return

    from app.api import chat, normalize, review, translate

    fastapi_app.include_router(chat.router, prefix="/api/v1/chat", tags=["chat"])
    fastapi_app.include_router(review.router, prefix="/api/v1", tags=["review"])
    fastapi_app.include_router(translate.router, prefix="/api/v1", tags=["translate"])
    fastapi_app.include_router(normalize.router, prefix="/api/v1", tags=["normalize"])


register_api_routers(app, settings.enable_legacy_ai_routes)


@app.get("/health")
async def health():
    return {"status": "ok"}
