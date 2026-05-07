"""经验审核 + 打分 API — 优化版"""
import json
import logging
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field

from app.core.domain_labels import get_domain_label, get_sub_label

logger = logging.getLogger(__name__)
router = APIRouter()


class ReviewRequest(BaseModel):
    content: str = Field(..., min_length=1, max_length=100)
    domain: str
    sub_domain: str = ""


class ScoreDetail(BaseModel):
    overall: float = Field(..., ge=0, le=10)
    actionable: float = Field(..., ge=0, le=10)
    insightful: float = Field(..., ge=0, le=10)
    fresh: float = Field(..., ge=0, le=10)
    universal: float = Field(..., ge=0, le=10)
    clarity: float = Field(..., ge=0, le=10)


class ReviewResponse(BaseModel):
    approved: bool
    reason: str
    score: ScoreDetail | None = None


REVIEW_PROMPT = """你是一个经验内容审核专家。请审核以下经验，判断是否适合收录到经验分享平台。

## 这条经验
"{content}"
领域：{domain_label}

## 审核标准

### 第一步：是否应该收录？

你在审核的是"平台生产经验"——来自名人、书籍、UGC 中已有的经验语句。这些经验通常短小精悍（10-50字），不是用户自己写的长篇经验。

**直接拒绝**（approved=false）：
- 违反中国法律法规或含不良引导
- 纯知识点/事实陈述（"地球绕太阳转"），不是经验
- 不是完整的一句话（"努力""坚持" 等碎片）

**以下情况应通过**（即使 actionable 低）：
- 视角型：换一个角度看问题的短句（"Stay hungry, stay foolish"）
- 原则型：指导行为的底层信念（"延迟满足感，是获得更大成就的关键"）
- 名言警句：流传广泛的智慧短句（"星星之火，可以燎原"）

判断标准：读完这句话，你对某件事的感觉有没有变？有没有觉得"嗯，可以换个思路"？有就过。不要求"看了就能做"。

### 第二步：如果收录，打多少分？

五个维度，各 0-10 分：

| 维度 | 衡量什么 | 高分特征 |
|------|---------|---------|
| actionable | 可执行度 | 看了就能用，具体到动作 |
| insightful | 洞察深度 | 让人停下来想一下，换个角度看问题 |
| fresh | 新鲜感/趣味性 | 不是老生常谈，有点意外或者眼前一亮 |
| universal | 普适性 | 大多数人能用上 |
| clarity | 表达清晰 | 简洁有力，不绕 |

overall = 上述五项的综合直觉

评分分布参考：
- 7-8 分：不错的经验，收录
- 8-9 分：很好的经验，值得推荐
- 9-10 分：极少数精品
- 5-6 分：勉强可用但不够好 → 拒绝
- <5 分：明显不够格 → 拒绝

注意：
- 视角型和原则型经验可能 actionable 低但 insightful/fresh 高——正常
- 不要因为 actionable 低就直接拒掉视角型经验

返回纯 JSON（不要 markdown 代码块）：
{{"approved": true/false, "reason": "一句简短理由（≤20字）",
  "score": {{"overall": 0.0, "actionable": 0.0, "insightful": 0.0, "fresh": 0.0, "universal": 0.0, "clarity": 0.0}}}}"""


@router.post("/review", response_model=ReviewResponse)
async def review_experience(req: ReviewRequest):
    import app.services.llm as llm_module

    if not llm_module.llm_service:
        raise HTTPException(status_code=503, detail="AI 服务未就绪")

    domain_label = get_domain_label(req.domain)
    sub_label = get_sub_label(req.domain, req.sub_domain) if req.sub_domain else ""
    full_label = f"{domain_label} · {sub_label}" if sub_label else domain_label

    prompt = REVIEW_PROMPT.format(content=req.content, domain_label=full_label)

    try:
        response = await llm_module.llm_service.chat(
            messages=[{"role": "user", "content": prompt}],
        )
        content = response.strip()
        if content.startswith("```"):
            content = content.split("\n", 1)[1]
            if content.endswith("```"):
                content = content[:-3]
            content = content.strip()
            if content.startswith("json"):
                content = content[4:].strip()

        result = json.loads(content)

        approved = bool(result.get("approved", False))
        reason = str(result.get("reason", ""))
        score_data = result.get("score")

        if score_data and approved:
            score = ScoreDetail(
                overall=float(score_data.get("overall", 5)),
                actionable=float(score_data.get("actionable", 5)),
                insightful=float(score_data.get("insightful", 5)),
                fresh=float(score_data.get("fresh", 5)),
                universal=float(score_data.get("universal", 5)),
                clarity=float(score_data.get("clarity", 5)),
            )
        else:
            score = None

        logger.info(f"Review: approved={approved}, overall={score.overall if score else 'N/A'}")
        return ReviewResponse(approved=approved, reason=reason, score=score)

    except json.JSONDecodeError as e:
        logger.error(f"Failed to parse review response: {content[:200]}")
        raise HTTPException(status_code=500, detail=f"AI 返回格式异常: {str(e)}")
    except Exception as e:
        logger.error(f"Review failed: {e}")
        raise HTTPException(status_code=500, detail=f"审核失败: {str(e)}")
