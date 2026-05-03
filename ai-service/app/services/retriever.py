"""经验检索 — jieba 分词 + ILIKE 关键词匹配"""
import logging
from typing import Dict

import asyncpg
import jieba

from app.core.config import settings

logger = logging.getLogger(__name__)

# 常见停用词
STOPWORDS = set(
    "的 了 是 在 我 有 和 就 不 人 都 一 一个 上 也 很 到 说 要 去 你 会 着 没有 看 好 自己 这 那 他 她 它 们 吗 吧 呢 啊 哦 嗯 什么 怎么 为什么 可以 这个 那个 因为 所以 但是 虽然 如果 然后 已经 还是 不过 只是 觉得 应该 可能 比较 非常 真的 太 多 少 很".split()
)


def _extract_keywords(text: str, max_kw: int = 6) -> list[str]:
    """jieba 分词 + 去停用词，提取关键词"""
    words = jieba.cut(text)
    keywords = []
    for w in words:
        w = w.strip()
        if len(w) >= 2 and w not in STOPWORDS:
            keywords.append(w)
    return keywords[:max_kw]


async def search_experiences(
    query_text: str = "",
    user_id: str = "",
    limit: int = 5,
) -> list[Dict]:
    """关键词检索用户经验"""
    conn = None
    try:
        conn = await asyncpg.connect(settings.database_url)

        if query_text:
            keywords = _extract_keywords(query_text)
            logger.info(f"Keywords: {keywords}")
            if keywords:
                # ILIKE 条件：任意关键词匹配 content 或 interpretation
                conditions = []
                params = []
                for i, kw in enumerate(keywords):
                    p = f"${i+1}"
                    conditions.append(
                        f"(e.content ILIKE '%' || {p} || '%' OR e.interpretation ILIKE '%' || {p} || '%')"
                    )
                    params.append(kw)

                where = " OR ".join(conditions)
                query = f"""
                    SELECT e.id, e.content, e.domain, e.like_count,
                           u.nickname as author_name
                    FROM experiences e
                    LEFT JOIN users u ON u.id = e.author_id
                    WHERE e.status = 'published'
                      AND (e.author_id = ${len(keywords) + 1} OR e.author_id = '00000000-0000-0000-0000-000000000000')
                      AND ({where})
                    ORDER BY e.like_count DESC, e.created_at DESC
                    LIMIT ${len(keywords) + 2}
                """
                params.append(user_id)
                params.append(limit)
                rows = await conn.fetch(query, *params)
                return [dict(r) for r in rows]

        # 无关键词时返回用户自己的 + 官方经验
        rows = await conn.fetch(
            """
            SELECT e.id, e.content, e.domain, e.like_count,
                   u.nickname as author_name
            FROM experiences e
            LEFT JOIN users u ON u.id = e.author_id
            WHERE e.status = 'published'
              AND (e.author_id = $1 OR e.author_id = '00000000-0000-0000-0000-000000000000')
            ORDER BY e.like_count DESC, e.created_at DESC
            LIMIT $2
            """,
            user_id,
            limit,
        )
        return [dict(r) for r in rows]
    except Exception as e:
        logger.error(f"Search failed: {e}")
        return []
    finally:
        if conn:
            await conn.close()
