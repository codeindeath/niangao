import os
import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.api import chat, experience
from app.core.config import settings
from app.services.embedding import EmbeddingService
from app.services.llm import LLMService

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Global services
embedding_service: EmbeddingService = None
llm_service: LLMService = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global embedding_service, llm_service
    logger.info("Starting AI service...")
    embedding_service = EmbeddingService()
    llm_service = LLMService()
    yield
    logger.info("Shutting down AI service...")


app = FastAPI(
    title="年糕 AI Service",
    version="0.1.0",
    lifespan=lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(chat.router, prefix="/api/v1/chat", tags=["chat"])
app.include_router(experience.router, prefix="/api/v1/ai/experience", tags=["experience"])


@app.get("/health")
async def health():
    return {"status": "ok"}


@app.get("/")
async def root():
    return {"service": "年糕 AI", "version": "0.1.0"}
