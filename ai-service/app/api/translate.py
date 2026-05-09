"""
古文检测 + 现代文翻译 API

判定输入文本是否为文言文/古文。
若是，翻译成优雅简洁的现代汉语；若不是，直接返回原文。
"""

import json
import logging

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field

logger = logging.getLogger(__name__)
router = APIRouter()


class TranslateRequest(BaseModel):
    content: str = Field(..., min_length=1, max_length=200)


class TranslateResponse(BaseModel):
    is_classical: bool
    original_text: str
    modern_text: str


TRANSLATE_PROMPT = """你是古汉语专家。判断以下文本是否为文言文/古文（非现代白话文）。

判定标准：
- 文言文特征：之乎者也、单音节词为主、省略主语、倒装句式、典故引用
- 现代白话文特征：多音节词、口语化、语法完整

如果是古文，请翻译成优雅简洁的现代汉语。要求：
- 保持原文的精炼气质，不要啰嗦
- 不要自行添加原文没有的内容或解释
- 用词优雅但不拗口

如果不是古文，modern_text 直接返回原文（一字不改）。

只返回纯 JSON（不要 markdown 代码块）：
{{"is_classical": true/false, "modern_text": "现代文翻译或原文"}}

待判定文本：
{content}"""


@router.post("/translate", response_model=TranslateResponse)
async def translate_content(req: TranslateRequest):
    import app.services.llm as llm_module

    if not llm_module.llm_service:
        raise HTTPException(status_code=503, detail="AI 服务未就绪")

    prompt = TRANSLATE_PROMPT.format(content=req.content)

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
        is_classical = bool(result.get("is_classical", False))
        modern_text = str(result.get("modern_text", req.content))

        return TranslateResponse(
            is_classical=is_classical,
            original_text=req.content,
            modern_text=modern_text if is_classical else req.content,
        )

    except json.JSONDecodeError as e:
        logger.error(f"Failed to parse translate response: {content[:200]}")
        # Fallback: 返回原文
        return TranslateResponse(
            is_classical=False,
            original_text=req.content,
            modern_text=req.content,
        )
    except Exception as e:
        logger.error(f"Translate failed: {e}")
        raise HTTPException(status_code=500, detail=f"翻译失败: {str(e)}")
