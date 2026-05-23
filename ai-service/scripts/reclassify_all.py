"""
重新推断全部经验的领域和子领域（v3 新体系）。
"""

import json
import os

import psycopg2
from openai import OpenAI

DB_URL = os.getenv("DATABASE_URL")
DEEPSEEK_API_KEY = os.getenv("DEEPSEEK_API_KEY")
DEEPSEEK_BASE_URL = os.getenv("DEEPSEEK_BASE_URL", "https://api.deepseek.com")

if not DB_URL:
    raise RuntimeError("DATABASE_URL is required")
if not DEEPSEEK_API_KEY:
    raise RuntimeError("DEEPSEEK_API_KEY is required")

client = OpenAI(
    api_key=DEEPSEEK_API_KEY,
    base_url=DEEPSEEK_BASE_URL,
)

# 中文→英文 key 映射（兜底）
CN2EN = {
    "生命": "vitality", "生活": "living", "工作": "work",
    "关系": "relationship", "认知": "cognition", "意义": "meaning",
    "健康": "health", "居住": "housing", "出行": "transit",
    "饮食": "diet", "运动": "exercise",
    "宠物": "pets", "旅行": "travel", "衣着": "fashion",
    "养护": "selfcare", "购物": "shopping", "娱乐": "fun",
    "求职": "jobhunt", "升职": "promotion", "创业": "startup",
    "沟通": "work-comm", "管理": "management", "效率": "productivity",
    "夫妻": "marriage", "恋人": "romance", "朋友": "friendship",
    "亲子": "parenting", "父母": "parents", "兄妹": "siblings",
    "学习": "cog-learning", "思维": "thinking", "信息": "info",
    "工具": "tools", "创造": "creativity", "表达": "expression",
    "自我": "self", "幸福": "happiness", "情绪": "emotion", "信仰": "faith",
    "使命": "mission", "归属": "belonging",
}

SYSTEM = "你是一个精准的领域分类器。只返回 JSON，使用英文 key。格式: {\"domain\":\"vitality\",\"sub_domain\":\"health\"}。可选值：vitality/living/work/relationship/cognition/meaning。子领域见用户消息。"


def classify(content):
    resp = client.chat.completions.create(
        model="deepseek-chat",
        messages=[
            {"role": "system", "content": SYSTEM},
            {"role": "user", "content": f"""领域和子领域选项：
vitality(生命): health(健康)/housing(居住)/transit(出行)/diet(饮食)/exercise(运动)
living(生活): pets(宠物)/travel(旅行)/fashion(衣着)/selfcare(养护)/shopping(购物)/fun(娱乐)
work(工作): jobhunt(求职)/promotion(升职)/startup(创业)/work-comm(沟通)/management(管理)/productivity(效率)
relationship(关系): marriage(夫妻)/romance(恋人)/friendship(朋友)/parenting(亲子)/parents(父母)/siblings(兄妹)
cognition(认知): cog-learning(学习)/thinking(思维)/info(信息)/tools(工具)/creativity(创造)/expression(表达)
meaning(意义): self(自我)/happiness(幸福)/emotion(情绪)/faith(信仰)/mission(使命)/belonging(归属)

经验: {content}"""},
        ],
        temperature=0.1,
        max_tokens=80,
    )
    text = resp.choices[0].message.content.strip()
    if text.startswith("```"):
        text = text.split("\n", 1)[1].rsplit("\n", 1)[0]
    r = json.loads(text)
    # 兜底：中文→英文
    d = r.get("domain", "")
    s = r.get("sub_domain", "")
    if d in CN2EN:
        d = CN2EN[d]
    if s in CN2EN:
        s = CN2EN[s]
    return d, s


def main():
    conn = psycopg2.connect(DB_URL)
    cur = conn.cursor()
    cur.execute("SELECT id, content FROM experiences ORDER BY created_at")
    rows = cur.fetchall()
    total = len(rows)
    print(f"共 {total} 条")

    ok = err = 0
    for i, (eid, content) in enumerate(rows):
        try:
            d, s = classify(content)
            cur.execute("UPDATE experiences SET domain=%s, sub_domain=%s WHERE id=%s", (d, s, eid))
            conn.commit()
            ok += 1
        except Exception as e:
            conn.rollback()
            err += 1
            if err <= 3:
                print(f"  err [{eid[:8]}]: {e}")
            if err > 10:
                print("错误过多，停止")
                break
        if (i + 1) % 30 == 0:
            print(f"  {i+1}/{total} ok={ok} err={err}")

    cur.close()
    conn.close()
    print(f"\n完成: ok={ok} err={err} total={total}")

    # 显示新分布
    conn2 = psycopg2.connect(DB_URL)
    cur2 = conn2.cursor()
    cur2.execute("SELECT domain, count(*) FROM experiences GROUP BY domain ORDER BY count(*) DESC")
    print("\n新分布：")
    for d, c in cur2.fetchall():
        print(f"  {d}: {c}")
    cur2.close()
    conn2.close()


if __name__ == "__main__":
    main()
