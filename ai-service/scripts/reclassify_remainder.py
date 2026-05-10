"""快速重分类残留的旧领域经验（career/life/emotion → v3）"""

import json, time
import psycopg2
from openai import OpenAI

DB_URL = "postgres://niangao:NiangaoDB2026!@localhost:5432/niangao?sslmode=disable"
client = OpenAI(api_key="sk-fcaca52db0e3400da406a12f2005ca7a", base_url="https://api.deepseek.com")

CN2EN = {"生命":"vitality","生活":"living","工作":"work","关系":"relationship","认知":"cognition","意义":"meaning",
         "健康":"health","居住":"housing","出行":"transit","饮食":"diet","运动":"exercise","宠物":"pets","旅行":"travel",
         "衣着":"fashion","养护":"selfcare","购物":"shopping","娱乐":"fun","求职":"jobhunt","升职":"promotion",
         "创业":"startup","沟通":"work-comm","管理":"management","效率":"productivity","夫妻":"marriage","恋人":"romance",
         "朋友":"friendship","亲子":"parenting","父母":"parents","兄妹":"siblings","学习":"cog-learning","思维":"thinking",
         "信息":"info","工具":"tools","创造":"creativity","表达":"expression","自我":"self","幸福":"happiness",
         "信仰":"faith","使命":"mission","归属":"belonging"}

SYSTEM = "只返回JSON: {\"domain\":\"english_key\",\"sub_domain\":\"english_key\"}"
USER_PREFIX = """领域:vitality(生命) health/housing/transit/diet/exercise
living(生活) pets/travel/fashion/selfcare/shopping/fun
work(工作) jobhunt/promotion/startup/work-comm/management/productivity
relationship(关系) marriage/romance/friendship/parenting/parents/siblings
cognition(认知) cog-learning/thinking/info/tools/creativity/expression
meaning(意义) self/happiness/faith/mission/belonging

经验:"""

def classify(content):
    for attempt in range(3):
        try:
            resp = client.chat.completions.create(
                model="deepseek-chat",
                messages=[{"role":"system","content":SYSTEM},{"role":"user","content":f"{USER_PREFIX}{content}"}],
                temperature=0.1, max_tokens=60, timeout=15,
            )
            text = resp.choices[0].message.content.strip()
            if text.startswith("```"): text = text.split("\n",1)[1].rsplit("\n",1)[0]
            r = json.loads(text)
            d, s = r.get("domain",""), r.get("sub_domain","")
            if d in CN2EN: d = CN2EN[d]
            if s in CN2EN: s = CN2EN[s]
            return d, s
        except Exception as e:
            if attempt < 2: time.sleep(1)
            else: raise

OLD = ("career", "life", "emotion")
conn = psycopg2.connect(DB_URL)
cur = conn.cursor()
cur.execute(f"SELECT id, content FROM experiences WHERE domain IN {OLD} ORDER BY created_at")
rows = cur.fetchall()
print(f"残留 {len(rows)} 条")

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
        print(f"  err [{eid[:8]}]: {e}")
        if err > 5: break
    if (i+1) % 10 == 0: print(f"  {i+1}/{len(rows)} ok={ok} err={err}")

cur2 = conn.cursor()
cur2.execute("SELECT domain, count(*) FROM experiences GROUP BY domain ORDER BY count(*) DESC")
print("\n最终分布:")
for d, c in cur2.fetchall(): print(f"  {d}: {c}")
cur.close(); cur2.close(); conn.close()
