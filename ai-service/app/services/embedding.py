"""向量检索 — 直接查询 PG 的 pgvector"""
import logging
from typing import List, Dict
import asyncpg

from app.core.config import settings

logger = logging.getLogger(__name__)


async def search_similar_experiences(
    query_embedding: List[float],
    user_id: str,
    limit: int = 5,
) -> List[Dict]:
    """在 PG 中检索与查询最相似的用户经验"""
    conn = None
    try:
        conn = await asyncpg.connect(settings.database_url)
        # pgvector <=> 操作符做余弦相似度排序
        rows = await conn.fetch(
            """
            SELECT e.id, e.content, e.domain, e.like_count,
                   u.nickname as author_name
            FROM experiences e
            LEFT JOIN users u ON u.id = e.author_id
            WHERE e.status = 'published' AND e.author_id = $1
            ORDER BY e.embedding <=> $2::vector
            LIMIT $3
            """,
            user_id,
            str(query_embedding),
            limit,
        )
        return [dict(r) for r in rows]
    except Exception as e:
        logger.error(f"Search failed: {e}")
        return []
    finally:
        if conn:
            await conn.close()


async def index_experience(experience_id: str, embedding: List[float]):
    """将向量写入 PG"""
    conn = None
    try:
        conn = await asyncpg.connect(settings.database_url)
        await conn.execute(
            "UPDATE experiences SET embedding = $1::vector WHERE id = $2",
            str(embedding),
            experience_id,
        )
    except Exception as e:
        logger.error(f"Index failed: {e}")
    finally:
        if conn:
            await conn.close()
