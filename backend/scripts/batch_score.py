#!/usr/bin/env python3
"""Batch score + interpret all unscored platform experiences."""
import asyncio
import aiohttp
import asyncpg
import os
from datetime import datetime

DB_DSN = os.getenv("DATABASE_URL")
AI_BASE = os.getenv("AI_SERVICE_URL", "http://localhost:8000")
BATCH_SIZE = 5

if not DB_DSN:
    raise RuntimeError("DATABASE_URL is required")

def score_to_reason(score: dict) -> str:
    dims = {k: v for k, v in score.items() if k != 'overall'}
    if not dims:
        return '综合质量好'
    best = max(dims, key=dims.get)
    labels = {
        'value': '内容有干货', 'actionable': '可执行性强',
        'universal': '普适性高', 'original': '观点独特',
        'clarity': '表达清晰有力',
    }
    return labels.get(best, '综合质量好')

async def process_one(row):
    """Score and interpret one experience."""
    eid = str(row['id'])
    content = row['content']
    domain = row['domain']
    sub_domain = row['sub_domain']

    async with aiohttp.ClientSession(timeout=aiohttp.ClientTimeout(total=60)) as session:
        # 1. Review
        try:
            async with session.post(f"{AI_BASE}/api/v1/review", json={
                "content": content, "domain": domain, "sub_domain": sub_domain or "",
            }) as resp:
                if resp.status != 200:
                    return None
                review = await resp.json()
                score = review.get('score', {})
                overall = score.get('overall', 5.0)
                reason = score_to_reason(score)
        except Exception as e:
            print(f"  [{eid[:8]}] review error: {e}")
            return None

        # 1.5 Classical Chinese detection & translation
        modern_content = content
        original_text = None
        try:
            async with session.post(f"{AI_BASE}/api/v1/translate", json={
                "content": content,
            }) as resp:
                if resp.status == 200:
                    trans = await resp.json()
                    if trans.get('is_classical'):
                        modern_content = trans.get('modern_text', content)
                        original_text = trans.get('original_text', content)
                        print(f"  [{eid[:8]}] Classical detected → translated")
        except Exception as e:
            print(f"  [{eid[:8]}] translate error: {e}")

        # 2. Interpretation (use modern_content for better AI understanding)
        try:
            async with session.post(f"{AI_BASE}/api/v1/chat/generate-interpretation", json={
                "content": modern_content, "domain": domain,
            }) as resp:
                if resp.status != 200:
                    interp = ""
                else:
                    data = await resp.json()
                    interp = data.get('interpretation', '')
        except Exception as e:
            print(f"  [{eid[:8]}] interp error: {e}")
            interp = ""

    # Truncate to 500 chars for DB constraint (conservative: target 490 + safety margin)
    if interp:
        max_chars = 490
        if len(interp) > max_chars:
            # Find a safe cut point at a sentence boundary
            cut = interp[:max_chars].rstrip()
            # Try to cut at last period or newline
            for sep in ['。', '\\n\\n', '\\n', '.', '！', '？']:
                idx = cut.rfind(sep)
                if idx > max_chars * 0.7:
                    interp = cut[:idx+1]
                    break
            else:
                interp = cut
        # Final safety: hard truncate if still too long
        if len(interp) > 500:
            interp = interp[:495]

    return (eid, overall, reason, interp, modern_content, original_text)

async def main():
    db = await asyncpg.connect(DB_DSN)
    rows = await db.fetch("""
        SELECT id, content, domain, sub_domain
        FROM experiences
        WHERE is_official = true AND quality_score IS NULL
        ORDER BY created_at
    """)
    total = len(rows)
    print(f"Need to process: {total} experiences")

    sem = asyncio.Semaphore(BATCH_SIZE)

    async def bounded(row):
        async with sem:
            return await process_one(row)

    tasks = [bounded(row) for row in rows]
    done = 0
    ok = 0
    for coro in asyncio.as_completed(tasks):
        result = await coro
        done += 1
        if result:
            eid, overall, reason, interp, modern_content, original_text = result
            try:
                # Update content if classical Chinese was translated
                if original_text:
                    await db.execute("""
                        UPDATE experiences SET content = $1, original_text = $2
                        WHERE id = $3::uuid
                    """, modern_content, original_text, eid)
                # Update score
                await db.execute("""
                    UPDATE experiences SET quality_score = $1, score_reason = $2
                    WHERE id = $3::uuid
                """, overall, reason, eid)
                # Update interpretation (may fail if too long)
                if interp:
                    await db.execute("""
                        UPDATE experiences SET interpretation = $1, interpretation_generated = TRUE
                        WHERE id = $2::uuid
                    """, interp, eid)
                ok += 1
            except Exception as e:
                print(f"  [{eid[:8]}] db error: {e}")
        if done % 10 == 0:
            print(f"Progress: {done}/{total} (ok: {ok})")

    await db.close()
    print(f"Done: {ok}/{total}")

if __name__ == "__main__":
    asyncio.run(main())
