"""向量化和检索服务"""
import logging
from typing import List, Dict

from supabase import create_client

from app.core.config import settings

logger = logging.getLogger(__name__)


class EmbeddingService:
    """管理经验向量的存储和检索"""

    def __init__(self):
        self.supabase = create_client(
            settings.supabase_url,
            settings.supabase_service_key,
        )

    async def index_experience(self, experience_id: str, embedding: List[float]):
        """将经验的向量存入数据库"""
        try:
            self.supabase.table("experiences").update({
                "embedding": embedding,
            }).eq("id", experience_id).execute()
        except Exception as e:
            logger.error(f"Failed to index experience {experience_id}: {e}")

    async def search_similar(
        self,
        query_embedding: List[float],
        user_id: str,
        limit: int = 5,
    ) -> List[Dict]:
        """检索与查询最相似的用户经验"""
        try:
            result = self.supabase.rpc(
                "search_user_experiences",
                {
                    "query_embedding": query_embedding,
                    "user_id": user_id,
                    "match_limit": limit,
                },
            ).execute()
            return result.data or []
        except Exception as e:
            logger.error(f"Failed to search experiences: {e}")
            return []
