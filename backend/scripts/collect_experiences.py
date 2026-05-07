#!/usr/bin/env python3
"""
平台经验采集脚本 — 从多源采集经验语句，经 AI 提取+审核后入池

用法:
  python3 collect_experiences.py celebrity  # 名人言论采集
  python3 collect_experiences.py book       # 书籍精华采集
  python3 collect_experiences.py ugc        # 图文 UGC 采集
  python3 collect_experiences.py all        # 全部来源

依赖: pip3 install aiohttp
"""

import asyncio
import json
import os
import sys
import time
import traceback
from datetime import datetime

import aiohttp

# === 配置 ===
AI_SERVICE_URL = os.getenv("AI_SERVICE_URL", "http://localhost:8000")
ADMIN_API_URL = os.getenv("ADMIN_API_URL", "http://localhost:8080")
ADMIN_TOKEN = os.getenv("ADMIN_TOKEN", "")  # 从环境变量取

# 批量配置
BATCH_SIZE = 3   # 并发数（谨慎：AI extract 较慢）
DELAY_BETWEEN_SOURCES = 2  # 秒

# ====================================================================
# 数据源：名人言论
# ====================================================================
CELEBRITY_SOURCES = [
    # 李小龙
    {
        "label": "李小龙",
        "type": "celebrity",
        "texts": [
            "Empty your mind, be formless, shapeless — like water. Now you put water into a cup, it becomes the cup; you put water into a bottle, it becomes the bottle; you put it in a teapot, it becomes the teapot. Now water can flow or it can crash. Be water, my friend.",
            "Adapt what is useful, reject what is useless, and add what is specifically your own.",
            "I fear not the man who has practiced 10,000 kicks once, but I fear the man who has practiced one kick 10,000 times.",
            "The successful warrior is the average man, with laser-like focus.",
            "Mistakes are always forgivable, if one has the courage to admit them.",
            "To hell with circumstances; I create opportunities.",
            "As you think, so shall you become.",
            "A wise man can learn more from a foolish question than a fool can learn from a wise answer.",
            "If you love life, don't waste time, for time is what life is made up of.",
            "The key to immortality is first living a life worth remembering.",
            "Knowing is not enough, we must apply. Willing is not enough, we must do.",
            "Always be yourself, express yourself, have faith in yourself, do not go out and look for a successful personality and duplicate it.",
            "Simplicity is the key to brilliance.",
            "Take no thought of who is right or wrong or who is better than. Be not for or against.",
            "A quick temper will make a fool of you soon enough.",
            "The less effort, the faster and more powerful you will be.",
        ],
    },
    # 乔布斯
    {
        "label": "史蒂夫·乔布斯",
        "type": "celebrity",
        "texts": [
            "Your time is limited, so don't waste it living someone else's life. Don't be trapped by dogma — which is living with the results of other people's thinking. Don't let the noise of others' opinions drown out your own inner voice. And most important, have the courage to follow your heart and intuition.",
            "Stay hungry, stay foolish.",
            "The people who are crazy enough to think they can change the world are the ones who do.",
            "Design is not just what it looks like and feels like. Design is how it works.",
            "Sometimes life hits you in the head with a brick. Don't lose faith.",
            "I'm convinced that about half of what separates the successful entrepreneurs from the non-successful ones is pure perseverance.",
            "You can't connect the dots looking forward; you can only connect them looking backwards.",
            "Remembering that you are going to die is the best way I know to avoid the trap of thinking you have something to lose.",
            "That's been one of my mantras — focus and simplicity. Simple can be harder than complex.",
        ],
    },
    # 张一鸣
    {
        "label": "张一鸣",
        "type": "celebrity",
        "texts": [
            "延迟满足感，是获得更大成就的关键。",
            "做你相信的事情，不要做别人希望你做的事情。",
            "对事情的认知是最关键的。你对事情的理解，就是你在这件事情上的竞争力。",
            "不要被短期利益迷惑，要做难而正确的事情。",
            "很多人一辈子就败在一个'等'字上。",
            "执行力就是不要想太多，先做。做的时候再调整。",
            "选择比努力更重要，但选择本身就是一种需要大量练习和努力的能力。",
            "信息的效率比信息的数量重要得多。",
            "好的判断力来自经验，但经验的获取来自糟糕的判断力。",
            "不要因为竞争去做一件事，要因为你相信它去做。",
            "如果你把一件事情想得很清楚，你不需要很多人。精英团队比大团队有效得多。",
            "一个人可以走得很快，一群人可以走得很远。但如果方向错了，走得越远越糟糕。",
        ],
    },
    # 马斯克
    {
        "label": "埃隆·马斯克",
        "type": "celebrity",
        "texts": [
            "When something is important enough, you do it even if the odds are not in your favor.",
            "The first step is to establish that something is possible; then probability will occur.",
            "I think it's possible for ordinary people to choose to be extraordinary.",
            "Persistence is very important. You should not give up unless you are forced to give up.",
            "If you get up in the morning and think the future is going to be better, it is a bright day. Otherwise, it's not.",
            "People work better when they know what the goal is and why.",
            "It's OK to have your eggs in one basket as long as you control what happens to that basket.",
            "Great companies are built on great products.",
            "Don't just follow the trend. You may have heard me say that it's good to think in terms of the physics approach of first principles.",
            "Brand is just a perception, and perception will match reality over time.",
            "Really, the only thing that makes sense is to strive for greater collective enlightenment.",
        ],
    },
    # 雷军
    {
        "label": "雷军",
        "type": "celebrity",
        "texts": [
            "站在风口上，猪都能飞起来。",
            "不要用战术上的勤奋掩盖战略上的懒惰。",
            "创业就是要做最肥的市场。",
            "专注、极致、口碑、快。",
            "人若无名，便可专心练剑。",
            "人生不要有太多的勉强，做自己喜欢做的事情可能是最佳的选择。",
            "找到一个愿意跟你一起奋斗的人，比找到一个商业模式更重要。",
            "把事情做好，钱自然会来。",
            "你做的每一件事，都在定义你自己。",
            "便宜没好货，好货不便宜。要做感动人心、价格厚道的好产品。",
            "永远相信美好的事情即将发生。",
            "创新决定我们飞多高，品质决定我们飞多远。",
        ],
    },
    # 毛泽东
    {
        "label": "毛泽东",
        "type": "celebrity",
        "texts": [
            "枪杆子里出政权。",
            "星星之火，可以燎原。",
            "一切反动派都是纸老虎。",
            "没有调查就没有发言权。",
            "战略上藐视敌人，战术上重视敌人。",
            "虚心使人进步，骄傲使人落后。",
            "多少事，从来急。天地转，光阴迫。一万年太久，只争朝夕。",
            "世界上怕就怕'认真'二字。",
            "你要知道梨子的滋味，你就得变革梨子，亲口吃一吃。",
            "读书是学习，使用也是学习，而且是更重要的学习。",
            "情况是在不断的变化，要使自己的思想适应新的情况，就得学习。",
        ],
    },
    # 张小龙
    {
        "label": "张小龙",
        "type": "celebrity",
        "texts": [
            "再小的个体，也有自己的品牌。",
            "好的产品是用完即走的。",
            "善良比聪明更重要。",
            "对于产品来说，善良比聪明更重要。因为聪明的产品经理太多了，善良的太少。",
            "用户价值第一。",
            "群体效应是不可预测的，但产品规则必须可预测。",
            "做产品就是要找到用户最本质的需求，而不是表面的需求。",
        ],
    },
    # 李想
    {
        "label": "李想",
        "type": "celebrity",
        "texts": [
            "任何成功都建立在无数次失败之上。",
            "最好的商业模式是你自己会用的产品。",
            "创业不是比谁聪明，而是比谁能坚持到最后。",
            "用户的需求永远是第一位的，而不是你自己的想法。",
        ],
    },
]


# ====================================================================
# 核心流程：extract → review → save
# ====================================================================

async def process_source(session: aiohttp.ClientSession, source: dict, sem: asyncio.Semaphore):
    """对单个来源文本执行完整的 extract→review→save 流水线"""
    label = source.get("label", "未知来源")
    source_type = source.get("type", "celebrity")
    texts = source.get("texts", [])

    total_extracted = 0
    total_saved = 0

    for text in texts:
        if not text.strip():
            continue

        # Step 1: AI 提取
        items = await extract_experiences(session, text, label, source_type)
        if not items:
            continue

        total_extracted += len(items)
        print(f"  [{label}] 提取 {len(items)} 条经验")

        # Step 2: AI 审核 + Step 3: 入库
        reviewed = []
        for item in items:
            if not item.get("content"):
                continue

            async with sem:
                result = await review_experience(session, item)
                if result:
                    reviewed.append(result)

        if reviewed:
            saved = await bulk_save(session, reviewed)
            total_saved += saved
            print(f"  [{label}] 入库 {saved}/{len(reviewed)} 条")

        await asyncio.sleep(DELAY_BETWEEN_SOURCES)

    return {"label": label, "extracted": total_extracted, "saved": total_saved}


async def extract_experiences(session: aiohttp.ClientSession, text: str, label: str, source_type: str):
    """调用 AI extract 端点"""
    try:
        async with session.post(
            f"{AI_SERVICE_URL}/api/v1/extract",
            json={"source_text": text, "source_label": label, "source_type": source_type},
            timeout=aiohttp.ClientTimeout(total=90),
        ) as resp:
            if resp.status != 200:
                print(f"  Extract HTTP {resp.status}: {await resp.text()[:200]}")
                return []
            data = await resp.json()
            return data.get("items", [])
    except Exception as e:
        print(f"  Extract error: {e}")
        return []


async def review_experience(session: aiohttp.ClientSession, item: dict):
    """调用 AI review 端点"""
    try:
        async with session.post(
            f"{AI_SERVICE_URL}/api/v1/review",
            json={
                "content": item.get("content", ""),
                "domain": item.get("domain", "cognition"),
                "sub_domain": item.get("sub_domain", ""),
            },
            timeout=aiohttp.ClientTimeout(total=60),
        ) as resp:
            if resp.status != 200:
                return None
            data = await resp.json()

            approved = data.get("approved", False)
            score = data.get("score", {})
            overall = score.get("overall", 5.0) if score else 5.0
            reason = data.get("reason", "")

            return {
                "content": item.get("content", ""),
                "domain": item.get("domain", "cognition"),
                "sub_domain": item.get("sub_domain", ""),
                "creator_name": item.get("source_label", ""),
                "source_label": item.get("source_label", ""),
                "quality_score": overall,
                "score_reason": reason[:100] if reason else "",
                "approved": approved,
            }
    except Exception as e:
        print(f"  Review error: {e}")
        return None


async def bulk_save(session: aiohttp.ClientSession, items: list):
    """调用 Go bulk-save 端点"""
    if not items:
        return 0

    headers = {"Content-Type": "application/json"}
    if ADMIN_TOKEN:
        headers["Authorization"] = f"Bearer {ADMIN_TOKEN}"

    try:
        async with session.post(
            f"{ADMIN_API_URL}/api/v1/admin/platform/pipeline/bulk-save",
            json={"items": items},
            headers=headers,
            timeout=aiohttp.ClientTimeout(total=30),
        ) as resp:
            if resp.status == 401:
                print(f"  Bulk-save auth failed (need ADMIN_TOKEN)")
                return 0
            if resp.status != 200:
                print(f"  Bulk-save HTTP {resp.status}: {await resp.text()[:200]}")
                return 0
            data = await resp.json()
            return data.get("saved", 0)
    except Exception as e:
        print(f"  Bulk-save error: {e}")
        return 0


# ====================================================================
# 手动输入模式
# ====================================================================

async def manual_ingest(session: aiohttp.ClientSession, source_text: str, label: str, source_type: str = "celebrity"):
    """通过 Go pipeline/ingest 端点一站式处理"""
    headers = {"Content-Type": "application/json"}
    if ADMIN_TOKEN:
        headers["Authorization"] = f"Bearer {ADMIN_TOKEN}"

    try:
        async with session.post(
            f"{ADMIN_API_URL}/api/v1/admin/platform/pipeline/ingest",
            json={"source_text": source_text, "source_label": label, "source_type": source_type},
            headers=headers,
            timeout=aiohttp.ClientTimeout(total=300),
        ) as resp:
            if resp.status != 200:
                print(f"  Ingest HTTP {resp.status}: {await resp.text()[:500]}")
                return
            data = await resp.json()
            print(f"\n  结果: 提取 {data.get('extracted',0)} | 通过 {data.get('reviewed',0)} | "
                  f"入库 {data.get('saved',0)} | 被拒 {data.get('rejected',0)} | "
                  f"错误 {data.get('errors',0)} | 新领域 {data.get('new_domains',0)}")
            return data
    except Exception as e:
        print(f"  Ingest error: {e}")
        return None


# ====================================================================
# 主入口
# ====================================================================

async def run_celebrity_collection():
    """采集所有名人言论"""
    print(f"\n{'='*60}")
    print(f"  平台经验采集 — 名人言论")
    print(f"  AI: {AI_SERVICE_URL}")
    print(f"  API: {ADMIN_API_URL}")
    print(f"  来源数: {len(CELEBRITY_SOURCES)}")
    print(f"  Token: {'已设置' if ADMIN_TOKEN else '⚠️ 未设置（需管理后台登录权限）'}")
    print(f"{'='*60}\n")

    sem = asyncio.Semaphore(BATCH_SIZE)

    async with aiohttp.ClientSession() as session:
        # 并行处理所有来源
        tasks = [process_source(session, src, sem) for src in CELEBRITY_SOURCES]
        results = await asyncio.gather(*tasks, return_exceptions=True)

    # 汇总
    total_extracted = 0
    total_saved = 0
    print(f"\n{'='*60}")
    print(f"  采集汇总")
    print(f"{'='*60}")
    for r in results:
        if isinstance(r, Exception):
            print(f"  ❌ 异常: {r}")
        elif r:
            print(f"  {r['label']}: 提取 {r['extracted']} → 入库 {r['saved']}")
            total_extracted += r['extracted']
            total_saved += r['saved']

    print(f"\n  总计: 提取 {total_extracted} 条, 入库 {total_saved} 条")
    print(f"{'='*60}\n")


async def run_manual_mode():
    """交互式手动输入"""
    print("\n平台经验采集 — 手动模式")
    print("输入来源文本（Ctrl+D 结束）:\n")

    lines = []
    try:
        while True:
            line = input()
            lines.append(line)
    except EOFError:
        pass

    text = "\n".join(lines)
    if not text.strip():
        print("未输入任何文本")
        return

    label = input("\n来源标签（书名/人名）: ").strip()
    source_type = input("类型 (book/celebrity/ugc) [celebrity]: ").strip() or "celebrity"

    async with aiohttp.ClientSession() as session:
        await manual_ingest(session, text, label, source_type)


async def main():
    if len(sys.argv) < 2:
        print("用法: python3 collect_experiences.py [celebrity|book|ugc|all|manual]")
        sys.exit(1)

    mode = sys.argv[1]

    if mode == "manual":
        await run_manual_mode()
    elif mode in ("celebrity", "all"):
        await run_celebrity_collection()
    else:
        print(f"模式 '{mode}' 数据源待补充（当前仅支持 celebrity/all/manual）")
        print("名人言论已就绪，book/ugc 需接入外部搜索引擎采集文本")


if __name__ == "__main__":
    asyncio.run(main())
