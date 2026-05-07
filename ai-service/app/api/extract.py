"""经验提取 API — 从原文中摘取现成经验语句"""
import json
import logging
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field

from app.core.domain_labels import (
    DOMAIN_TREE, get_all_subdomains, get_domain_list_for_prompt,
)

logger = logging.getLogger(__name__)
router = APIRouter()


class ExtractRequest(BaseModel):
    source_text: str = Field(..., min_length=1, max_length=3000)
    source_label: str = ""   # 书名人名等
    source_type: str = "book"  # book/celebrity/ugc


class ExtractedItem(BaseModel):
    content: str
    domain: str
    sub_domain: str = ""
    exp_type: str = "method"  # method/perspective/principle
    needs_new_subdomain: bool = False
    suggested_sub_name: str = ""
    suggested_sub_label: str = ""


EXTRACT_PROMPT = """你是一个经验采集助手。你的任务是从一段原文中找出所有可以独立成为"经验"的语句，原封不动地摘出来。

## 什么是"经验"
经验不是知识点，不是事实陈述，不是作者在叙事。经验是一个人通过经历提炼出来的、对其他人有启发或指导价值的内容。它有三种常见形态：

1. 方法型 — 可以照做的具体做法
2. 视角型 — 换一个角度看问题的一句话
3. 原则型 — 指导行为的底层信念

共同特征：读完之后，你对某件事的感觉变了，或者你知道该怎么想了——而不是仅仅"知道了一个事实"。

## 不属于经验的内容
- 纯知识/事实
- 作者在讲自己的故事但没提炼出可迁移的东西
- 泛泛的鸡汤
- 太依赖上下文才能理解的片段
- 超过 100 字的段落

## 可选的领域标签

{domain_list}

## 任务
从以下原文中找出所有符合上述定义的经验语句。每条经验：
- 照抄原文，一个字不改，可以截断到一句完整的话
- 归入最匹配的一级领域和子领域（从上面的列表里选）
- 标注类型（method/perspective/principle）
- 如果确实找不到匹配的子领域，设 needs_new_subdomain=true，并给出建议的新子领域名（suggested_sub_name 用英文kebab-case，suggested_sub_label 用中文2-5字）

如果原文中没有可提取的经验，返回空列表。不要为了凑数而把不是经验的内容塞进来。

原文：
{source_text}

返回纯 JSON 数组（不要 markdown 代码块）：
[
  {{"content": "经验原文", "domain": "career", "sub_domain": "career-planning", "exp_type": "method"}},
  ...
]"""


@router.post("/extract")
async def extract_experiences(req: ExtractRequest):
    import app.services.llm as llm_module

    if not llm_module.llm_service:
        raise HTTPException(status_code=503, detail="AI 服务未就绪")

    domain_list = get_domain_list_for_prompt()
    prompt = EXTRACT_PROMPT.format(
        domain_list=domain_list,
        source_text=req.source_text,
    )

    try:
        response = await llm_module.llm_service.chat(
            messages=[{"role": "user", "content": prompt}],
            temperature=0.3,
        )
        content = response.strip()
        if content.startswith("```"):
            content = content.split("\n", 1)[1]
            if content.endswith("```"):
                content = content[:-3]
            content = content.strip()
            if content.startswith("json"):
                content = content[4:].strip()

        items = json.loads(content)
        if not isinstance(items, list):
            items = []

        # Validate domains
        all_subs = set(get_all_subdomains())
        valid_parents = set(DOMAIN_TREE.keys())
        for item in items:
            if item.get("domain") not in valid_parents:
                item["domain"] = "cognition"  # fallback
            if item.get("sub_domain") not in all_subs:
                item["needs_new_subdomain"] = True
                item["sub_domain"] = ""

        logger.info(f"Extracted {len(items)} experiences from {len(req.source_text)} chars")
        return {"items": items, "count": len(items)}

    except json.JSONDecodeError as e:
        logger.error(f"Failed to parse extract response: {content[:200]}")
        raise HTTPException(status_code=500, detail=f"AI 返回格式异常: {str(e)}")
    except Exception as e:
        logger.error(f"Extract failed: {e}")
        raise HTTPException(status_code=500, detail=f"提取失败: {str(e)}")
