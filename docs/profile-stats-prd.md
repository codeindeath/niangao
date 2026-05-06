# 【我的】页统计需求 PRD

## 页面结构（自上而下）

### 1. 个人信息栏
```
[头像]  昵称          ›
        称号
```
- 点击整行 → 进入个人信息编辑页（已有 ProfileEditScreen）
- 昵称和称号来自 UserProfile（已有接口）

### 2. 第一块统计：我的内容表现

```
我发布的 12  |  获点赞 47  |  获收藏 23    ← V5-A1 大数字卡片风格

领域分布   [发布] [获赞] [被收藏]           ← 三按钮切换
███████████████████████████████ 思维模型  6
████████████████████ 决策判断            4
████████████████     情绪调节            3
████████████         学习方法            3
██████████           亲密关系            2
```

| 切换按钮 | 数据口径 | 当前接口 |
|---------|---------|---------|
| 发布 | 我创建的经验总数 + 领域分布 | `fetchMyExperiences` 可算 |
| 获赞 | 我创建的经验被他人点赞总数 + 领域分布 | 需要新接口 |
| 被收藏 | 我创建的经验被他人收藏总数 + 领域分布 | 需要新接口 |

### 3. 第二块统计：我的互动足迹

```
我看过的 89  |  我点赞 47  |  我收藏 23

领域分布   [看过] [点赞] [收藏]
███████████████████████████████ 思维模型  18
█████████████████████████       决策判断  15
█████████████████████           情绪调节  12
███████████████████             学习方法  10
█████████████████               亲密关系   8
```

| 切换按钮 | 数据口径 | 当前接口 |
|---------|---------|---------|
| 看过 | 我浏览过的经验总数 + 领域分布 | 需要新接口 |
| 点赞 | 我点赞过的经验总数 + 领域分布 | 需要新接口 |
| 收藏 | 我收藏过的经验总数 + 领域分布 | `fetchMyBookmarks` 可算总数，分布需新接口 |

### 4. 对话总数

```
        ┌──────────────┐  ┌──────────────┐
        │      8       │  │     142      │
        │   次对话      │  │   条消息      │
        └──────────────┘  └──────────────┘
```

| 指标 | 数据口径 | 当前接口 |
|------|---------|---------|
| 次对话 | 对话轮次数 | 需要新接口 |
| 条消息 | 总消息数 | 需要新接口 |

### 5. 服务列表（已有，不改）
经验包 → 对话人格 → 设置

---

## 领域分布计算规则

1. 按**二级领域**（sub_domain）分组统计
2. 按数量降序排列
3. 二级领域数量 > 5 → 只展示前 5 个
4. 二级领域数量 ≤ 5 → 展示全部
5. 领域标签使用中文名（SUB_LABELS 映射）

---

## 需要新增的后端接口

### 统计接口建议合并为一个，避免 6 次请求：

**`GET /api/v1/user/stats`** （需要登录）

响应结构：
```json
{
  "published": {
    "count": 12,
    "liked_by_others": 47,
    "bookmarked_by_others": 23,
    "domain_distribution": {
      "published": [
        {"domain": "mental-model", "count": 5},
        {"domain": "decision", "count": 3}
      ],
      "liked_by_others": [...],
      "bookmarked_by_others": [...]
    }
  },
  "interactions": {
    "viewed": 89,
    "liked": 47,
    "bookmarked": 23,
    "domain_distribution": {
      "viewed": [
        {"domain": "mental-model", "count": 18}
      ],
      "liked": [...],
      "bookmarked": [...]
    }
  },
  "chat": {
    "conversations": 8,
    "messages": 142
  }
}
```

### 各字段的 SQL 查询逻辑

**published.count** — `SELECT COUNT(*) FROM experiences WHERE author_id=$1 AND review_status='approved' AND deleted_at IS NULL`

**published.liked_by_others** — `SELECT COUNT(*) FROM likes WHERE experience_id IN (SELECT id FROM experiences WHERE author_id=$1)`

**published.bookmarked_by_others** — `SELECT COUNT(*) FROM bookmarks WHERE experience_id IN (SELECT id FROM experiences WHERE author_id=$1)`

**published.domain_distribution.published** — `SELECT sub_domain, COUNT(*) as cnt FROM experiences WHERE author_id=$1 AND review_status='approved' GROUP BY sub_domain ORDER BY cnt DESC`

**interactions.viewed** — 需要用户浏览记录表（目前没有，需要新表 `user_views`）

**interactions.liked** — `SELECT COUNT(*) FROM likes WHERE user_id=$1`

**interactions.bookmarked** — `SELECT COUNT(*) FROM bookmarks WHERE user_id=$1`

**interactions 的 domain_distribution** — 需要 JOIN experiences：
```sql
SELECT e.sub_domain, COUNT(*) FROM likes l
JOIN experiences e ON l.experience_id = e.id
WHERE l.user_id = $1
GROUP BY e.sub_domain ORDER BY COUNT(*) DESC
```

**chat.conversations / chat.messages** — 需要对话/消息表（目前不确定有没有）

---

## 待确认问题

1. **"获点赞" / "获收藏" 的口径**：是"我的经验被他人点赞/收藏"，还是"我点赞/收藏了别人的经验"？
   - 按原文"获点赞""获收藏"，理解是被动：我的内容被他人互动。确认？
   
2. **"我看过的"**：目前没有浏览记录表，有两个方案：
   - A：新建 `user_views` 表，记录每次浏览（完整但成本高）
   - B：先用"我看过的 = 首页推荐流中看过的" 或暂不实现，只做点赞+收藏
   
3. **对话统计**：需要确认数据库里对话和消息的表结构。如果暂无，是否先用占位数据（显示 0），后续对话功能上线再接入？

4. **前端请求时机**：进入我的页时一次请求 `GET /api/v1/user/stats` 获取全部数据，切换 tab 时仅前端过滤，不再请求。
