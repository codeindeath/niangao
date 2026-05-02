import os
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    # DeepSeek
    deepseek_api_key: str = os.getenv("DEEPSEEK_API_KEY", "")
    deepseek_base_url: str = os.getenv("DEEPSEEK_BASE_URL", "https://api.deepseek.com")
    deepseek_model: str = os.getenv("DEEPSEEK_MODEL", "deepseek-chat")

    # Supabase
    supabase_url: str = os.getenv("SUPABASE_URL", "")
    supabase_service_key: str = os.getenv("SUPABASE_SERVICE_KEY", "")

    # Embedding
    embedding_model: str = os.getenv("EMBEDDING_MODEL", "text-embedding-3-small")
    embedding_dim: int = 1536

    # Limits
    max_daily_chat_rounds: int = int(os.getenv("MAX_DAILY_CHAT_ROUNDS", "50"))
    max_context_experiences: int = int(os.getenv("MAX_CONTEXT_EXPERIENCES", "5"))

    class Config:
        env_file = ".env"


settings = Settings()
