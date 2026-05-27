import os
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", extra="ignore")

    # DeepSeek
    deepseek_api_key: str = os.getenv("DEEPSEEK_API_KEY", "")
    deepseek_base_url: str = os.getenv("DEEPSEEK_BASE_URL", "https://api.deepseek.com")
    deepseek_model: str = os.getenv("DEEPSEEK_MODEL", "deepseek-chat")
    deepseek_timeout_seconds: float = float(os.getenv("DEEPSEEK_TIMEOUT_SECONDS", "60"))

    # PostgreSQL
    database_url: str = os.getenv("DATABASE_URL", "")

    max_daily_chat_rounds: int = int(os.getenv("MAX_DAILY_CHAT_ROUNDS", "50"))
    max_context_experiences: int = int(os.getenv("MAX_CONTEXT_EXPERIENCES", "50"))
    enable_legacy_ai_routes: bool = os.getenv("ENABLE_LEGACY_AI_ROUTES", "").lower() in {"1", "true", "yes", "on"}


settings = Settings()
