"""经验审核 + 打分 API"""
import json
import logging
from typing import Optional

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field

logger = logging.getLogger(__name__)
router = APIRouter()


class ReviewRequest(BaseModel):
    content: str = Field(..., min_length=1, max_length=100)
    domain: str
    sub_domain: str


class ScoreDetail(BaseModel):
    overall: float = Field(..., ge=0, le=10)
    value: float = Field(..., ge=0, le=10)
    actionable: float = Field(..., ge=0, le=10)
    universal: float = Field(..., ge=0, le=10)
    original: float = Field(..., ge=0, le=10)
    clarity: float = Field(..., ge=0, le=10)


class ReviewResponse(BaseModel):
    approved: bool
    reason: str
    score: Optional[ScoreDetail] = None


REVIEW_PROMPT = """你是一个经验内容审核专家。请审核以下用户提交的经验：

【经验内容】{content}
【一级领域】{domain_label}
【二级领域】{sub_domain_label}

请从以下维度审核并返回 JSON：

1. 合规性检查：是否符合中华人民共和国法律法规？是否安全、正向、无不良引导？

2. 内容质量检查 — 以下 5 种类型必须拒绝（approved=false）：
   ① 个人经历：仅描述个人具体经历（"我昨天…""我在某公司时…"），未提炼出普适性经验教训
   ② 客观描述：仅陈述事实/定义/流程/数据，无个人洞见或可迁移启发
   ③ 纯鸡汤：空洞的励志口号或情绪抚慰（"相信自己""一切都是最好的安排"），无可操作方法
   ④ 段子/笑话：以娱乐为目的的段子、梗、幽默短句，无经验价值
   ⑤ 空泛无物：内容过短（≤8字）且无任何实质经验见解

3. 内容价值：是否是可以指导行动的经验（而非单纯的知识点如数学公式）？

打分维度（0-10，保留一位小数）：
- overall: 综合价值分
- value: 内容价值度（是否能带来新的认知或启发）
- actionable: 实操可执行度（读者能否直接应用）
- universal: 普适性（是否适用于大多数人）
- original: 原创性（是否是独到见解而非老生常谈）
- clarity: 清晰度（表达是否简洁明了）

返回纯 JSON（不要 markdown 代码块）：
{{"approved": true/false, "reason": "审核理由或拒绝原因（30字以内，拒绝时注明类型如【个人经历】【纯鸡汤】）",
  "score": {{"overall": 0.0, "value": 0.0, "actionable": 0.0, "universal": 0.0, "original": 0.0, "clarity": 0.0}}}}

注意：
- 如果内容违反法律法规、涉黄涉暴涉政，approved=false
- 如果内容属于上述 5 种拒绝类型之一，approved=false
- 打分要实事求是，大多数经验在 5-8 分区间
- reason 要简洁明确，拒绝时标注拒绝类型"""
DOMAIN_LABELS = {
    "vitality": "生命",
    "living": "生活",
    "work": "工作",
    "relationship": "关系",
    "cognition": "认知",
    "meaning": "意义",
}

SUB_DOMAIN_LABELS = {
    "health": "健康", "housing": "居住", "transit": "出行",
    "diet": "饮食", "exercise": "运动",
    "pets": "宠物", "travel": "旅行", "fashion": "衣着",
    "selfcare": "养护", "shopping": "购物", "fun": "娱乐",
    "jobhunt": "求职", "promotion": "升职", "startup": "创业",
    "work-comm": "沟通", "management": "管理", "productivity": "效率",
    "marriage": "夫妻", "romance": "恋人", "friendship": "朋友",
    "parenting": "亲子", "parents": "父母", "siblings": "兄妹",
    "cog-learning": "学习", "thinking": "思维", "info": "信息",
    "tools": "工具", "creativity": "创造", "expression": "表达",
    "self": "自我", "happiness": "幸福", "emotion": "情绪", "faith": "信仰",
    "mission": "使命", "belonging": "归属",
}


@router.post("/review", response_model=ReviewResponse)
async def review_experience(req: ReviewRequest):
    import app.services.llm as llm_module

    if not llm_module.llm_service:
        raise HTTPException(status_code=503, detail="AI 服务未就绪")

    domain_label = DOMAIN_LABELS.get(req.domain, req.domain)
    sub_label = SUB_DOMAIN_LABELS.get(req.sub_domain, req.sub_domain)

    prompt = REVIEW_PROMPT.format(
        content=req.content,
        domain_label=domain_label,
        sub_domain_label=sub_label,
    )

    try:
        response = await llm_module.llm_service.chat(
            messages=[{"role": "user", "content": prompt}],
        )
        content = response.strip()
        # 清理可能的 markdown 代码块
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
                value=float(score_data.get("value", 5)),
                actionable=float(score_data.get("actionable", 5)),
                universal=float(score_data.get("universal", 5)),
                original=float(score_data.get("original", 5)),
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
