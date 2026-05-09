"""
语言检测 + 翻译 API

自动检测输入文本的语言类型，执行对应翻译：
- 英文 → 优雅中文
- 古文/文言文 → 现代中文
- 现代中文 → 跳过
"""

import json
import logging

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field

logger = logging.getLogger(__name__)
router = APIRouter()


class TranslateRequest(BaseModel):
    content: str = Field(..., min_length=1, max_length=500)


class TranslateResponse(BaseModel):
    is_classical: bool     # 兼容旧字段名，语义：是否需要翻译
    original_text: str
    modern_text: str
    detected_lang: str     # "en" | "classical-zh" | "modern-zh"


TRANSLATE_PROMPT = """你是语言专家。分析以下文本的语言类型，执行对应翻译。

判定规则：
1. 如果是**英文** → 翻译成优雅简洁的中文。保持原文精炼，不要啰嗦，不要添加解释。
2. 如果是**文言文/古文**（之乎者也、单音节词为主、省略主语、典故引用）→ 翻译成优雅简洁的现代中文。
3. 如果是**现代中文白话文** → 不需要翻译，modern_text 直接返回原文一字不改。

只返回纯 JSON（不要 markdown 代码块）：
{{"lang": "en" | "classical-zh" | "modern-zh", "modern_text": "翻译结果或原文"}}

待分析文本：
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
        lang = str(result.get("lang", "modern-zh"))
        modern_text = str(result.get("modern_text", req.content))

        needs_translation = lang in ("en", "classical-zh")

        return TranslateResponse(
            is_classical=needs_translation,
            original_text=req.content,
            modern_text=modern_text if needs_translation else req.content,
            detected_lang=lang,
        )

    except (json.JSONDecodeError, KeyError) as e:
        logger.error(f"Failed to parse translate response: {content[:200]}")
        return TranslateResponse(
            is_classical=False,
            original_text=req.content,
            modern_text=req.content,
            detected_lang="modern-zh",
        )
    except Exception as e:
        logger.error(f"Translate failed: {e}")
        raise HTTPException(status_code=500, detail=f"翻译失败: {str(e)}")
