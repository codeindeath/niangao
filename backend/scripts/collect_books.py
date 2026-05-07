#!/usr/bin/env python3
"""
书籍经验采集器 — 从 Goodreads/豆瓣 等来源搜索书籍精华语录，经 AI 提取后入池

用法:
  python3 collect_books.py                    # 采集预置书单
  python3 collect_books.py --search "原子习惯"  # 搜索指定书籍
  python3 collect_books.py --url "https://..."  # 从指定 URL 提取

依赖: pip3 install aiohttp beautifulsoup4
"""

import asyncio
import json
import os
import re
import sys
import time
import urllib.parse
from html.parser import HTMLParser

import aiohttp

# === 配置 ===
AI_SERVICE_URL = os.getenv("AI_SERVICE_URL", "http://localhost:8000")
ADMIN_API_URL = os.getenv("ADMIN_API_URL", "http://localhost:8080")
ADMIN_TOKEN = os.getenv("ADMIN_TOKEN", "")

BATCH_SIZE = 2
MAX_QUOTES_PER_BOOK = 20

# ====================================================================
# 预置书单
# ====================================================================
BOOK_SOURCES = [
    # 中文书籍
    {"title": "原则", "author": "Ray Dalio", "query": "原则 Ray Dalio 经典语录 精华"},
    {"title": "原子习惯", "author": "James Clear", "query": "原子习惯 Atomic Habits 经典语录"},
    {"title": "纳瓦尔宝典", "author": "Naval Ravikant", "query": "纳瓦尔宝典 The Almanack of Naval 经典语录 金句"},
    {"title": "思考快与慢", "author": "Daniel Kahneman", "query": "思考快与慢 Thinking Fast and Slow 经典语录"},
    {"title": "穷查理宝典", "author": "Charlie Munger", "query": "穷查理宝典 Charlie Munger 经典语录 金句"},
    {"title": "非暴力沟通", "author": "Marshall Rosenberg", "query": "非暴力沟通 经典语录 金句"},
    {"title": "高效能人士的七个习惯", "author": "Stephen Covey", "query": "高效能人士的七个习惯 7 Habits 经典语录"},
    {"title": "影响力", "author": "Robert Cialdini", "query": "影响力 Influence Cialdini 经典语录"},
    {"title": "被讨厌的勇气", "author": "岸见一郎", "query": "被讨厌的勇气 经典语录 金句 阿德勒"},
    {"title": "活出生命的意义", "author": "Viktor Frankl", "query": "活出生命的意义 Man's Search for Meaning 经典语录"},
    {"title": "刻意练习", "author": "Anders Ericsson", "query": "刻意练习 Peak Anders Ericsson 经典语录"},
    {"title": "心流", "author": "Mihaly Csikszentmihalyi", "query": "心流 Flow Mihaly 经典语录"},
    {"title": "黑天鹅", "author": "Nassim Taleb", "query": "黑天鹅 The Black Swan Taleb 经典语录 金句"},
    {"title": "反脆弱", "author": "Nassim Taleb", "query": "反脆弱 Antifragile Taleb 经典语录 金句"},
    {"title": "局外人", "author": "Albert Camus", "query": "Albert Camus quotes wisdom philosophy"},
    {"title": "道德经", "author": "老子", "query": "道德经 老子 经典语录 智慧"},
]

# ====================================================================
# HTML 文本提取
# ====================================================================

class TextExtractor(HTMLParser):
    def __init__(self):
        super().__init__()
        self.text = []
        self.skip = False

    def handle_starttag(self, tag, attrs):
        if tag in ('script', 'style', 'nav', 'footer', 'header'):
            self.skip = True

    def handle_endtag(self, tag):
        if tag in ('script', 'style', 'nav', 'footer', 'header'):
            self.skip = False
        if tag in ('p', 'div', 'li', 'br', 'h1', 'h2', 'h3', 'h4'):
            self.text.append('\n')

    def handle_data(self, data):
        if not self.skip and data.strip():
            self.text.append(data.strip())


def extract_text_from_html(html: str) -> str:
    parser = TextExtractor()
    parser.feed(html)
    text = ' '.join(parser.text)
    # Clean up
    text = re.sub(r'\s+', ' ', text)
    text = re.sub(r' \n ', '\n', text)
    return text


def extract_quotes_from_text(text: str) -> list:
    """从文本中提取引号内的内容（引号是强信号）"""
    quotes = []
    # Chinese quotes
    for match in re.finditer(r'[「「]([^」」]{10,100})[」」]', text):
        q = match.group(1).strip()
        if len(q) >= 6:
            quotes.append(q)
    # English quotes
    for match in re.finditer(r'"([^"]{20,200})"', text):
        q = match.group(1).strip()
        if len(q) >= 20 and not q.startswith('http'):
            quotes.append(q)
    return quotes


# ====================================================================
# Web search
# ====================================================================

async def search_duckduckgo(session: aiohttp.ClientSession, query: str, max_results: int = 5):
    """搜索 DuckDuckGo 获取结果 URL"""
    url = f"https://html.duckduckgo.com/html/?q={urllib.parse.quote(query)}"
    headers = {
        "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
    }
    try:
        async with session.get(url, headers=headers, timeout=aiohttp.ClientTimeout(total=15)) as resp:
            html = await resp.text()
            # Extract result links
            urls = re.findall(r'class="result__a"[^>]*href="([^"]+)"', html)
            # Clean up redirect URLs
            cleaned = []
            for u in urls:
                # DuckDuckGo wraps URLs in redirect
                parsed = urllib.parse.urlparse(u)
                qs = urllib.parse.parse_qs(parsed.query)
                if 'uddg' in qs:
                    cleaned.append(urllib.parse.unquote(qs['uddg'][0]))
                elif u.startswith('http'):
                    cleaned.append(u)
            return cleaned[:max_results]
    except Exception as e:
        print(f"  Search error: {e}")
        return []


async def fetch_page(session: aiohttp.ClientSession, url: str):
    """抓取页面文本"""
    headers = {
        "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
    }
    try:
        async with session.get(url, headers=headers, timeout=aiohttp.ClientTimeout(total=20)) as resp:
            if resp.status != 200:
                return ""
            html = await resp.text()
            return extract_text_from_html(html)
    except Exception as e:
        print(f"  Fetch error for {url[:60]}: {e}")
        return ""


# ====================================================================
# AI Pipeline
# ====================================================================

async def call_pipeline(session: aiohttp.ClientSession, text: str, label: str, source_type: str = "book"):
    """调用 Go pipeline/ingest 端点"""
    headers = {"Content-Type": "application/json"}
    if ADMIN_TOKEN:
        headers["Authorization"] = f"Bearer {ADMIN_TOKEN}"

    try:
        async with session.post(
            f"{ADMIN_API_URL}/api/v1/admin/platform/pipeline/ingest",
            json={"source_text": text, "source_label": label, "source_type": source_type},
            headers=headers,
            timeout=aiohttp.ClientTimeout(total=180),
        ) as resp:
            if resp.status != 200:
                return None
            return await resp.json()
    except Exception as e:
        print(f"  Pipeline error: {e}")
        return None


# ====================================================================
# Book processor
# ====================================================================

async def process_book(session: aiohttp.ClientSession, book: dict, sem: asyncio.Semaphore):
    """处理一本书：搜索 → 抓取 → 提取 → 入库"""
    title = book["title"]
    author = book["author"]
    print(f"\n📚 {title} ({author})")

    # Search for quote pages
    urls = await search_duckduckgo(session, book["query"])
    if not urls:
        print(f"  No search results")
        return {"title": title, "saved": 0}

    all_text = []
    for url in urls[:3]:  # Top 3 results
        print(f"  Fetching: {url[:80]}...")
        text = await fetch_page(session, url)
        if text:
            # Extract quoted content as a strong signal
            quotes = extract_quotes_from_text(text)
            if quotes:
                all_text.extend(quotes)
            else:
                # Fall back to raw text (first 3000 chars)
                all_text.append(text[:3000])

    if not all_text:
        print(f"  No content extracted")
        return {"title": title, "saved": 0}

    # Combine and deduplicate (by first 30 chars)
    seen = set()
    unique = []
    for q in all_text:
        key = q[:30]
        if key not in seen:
            seen.add(key)
            unique.append(q)

    print(f"  Found {len(unique)} unique text segments")

    # Feed to pipeline in chunks
    total_saved = 0
    label = f"{author} - {title}"

    async with sem:
        for i in range(0, min(len(unique), MAX_QUOTES_PER_BOOK), 3):
            chunk = "\n".join(unique[i:i+3])
            if len(chunk) < 20:
                continue
            result = await call_pipeline(session, chunk, label, "book")
            if result:
                saved = result.get("saved", 0)
                total_saved += saved
                print(f"  Chunk {i//3+1}: extracted {result.get('extracted',0)}, saved {saved}")
            await asyncio.sleep(1)

    print(f"  ✅ {title}: saved {total_saved}")
    return {"title": title, "saved": total_saved}


# ====================================================================
# Main
# ====================================================================

async def main():
    if "--search" in sys.argv:
        query = sys.argv[sys.argv.index("--search") + 1]
        books = [{"title": query, "author": "", "query": f"{query} 经典语录 金句 wisdom quotes"}]
    elif "--url" in sys.argv:
        url = sys.argv[sys.argv.index("--url") + 1]
        async with aiohttp.ClientSession() as session:
            text = await fetch_page(session, url)
            if text:
                result = await call_pipeline(session, text[:3000], "webpage", "ugc")
                print(json.dumps(result, indent=2, ensure_ascii=False))
        return
    else:
        books = BOOK_SOURCES

    print(f"\n{'='*60}")
    print(f"  平台经验采集 — 书籍精华")
    print(f"  AI: {AI_SERVICE_URL}  |  API: {ADMIN_API_URL}")
    print(f"  书单: {len(books)} 本")
    print(f"  Token: {'已设置' if ADMIN_TOKEN else '⚠️ 未设置'}")
    print(f"{'='*60}")

    sem = asyncio.Semaphore(BATCH_SIZE)
    total_saved = 0

    async with aiohttp.ClientSession() as session:
        for book in books:
            result = await process_book(session, book, sem)
            if result:
                total_saved += result["saved"]
            await asyncio.sleep(2)  # Rate limit between books

    print(f"\n{'='*60}")
    print(f"  总计入库: {total_saved} 条")
    print(f"{'='*60}\n")


if __name__ == "__main__":
    asyncio.run(main())
