"""
文本标准化：trim + 繁体→简体 + 标点规整
"""

import logging

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field

logger = logging.getLogger(__name__)
router = APIRouter()


class NormalizeRequest(BaseModel):
    content: str = Field(..., max_length=500)


class NormalizeResponse(BaseModel):
    content: str
    changed: bool  # whether any modification was applied


@router.post("/normalize", response_model=NormalizeResponse)
async def normalize_text(req: NormalizeRequest):
    original = req.content
    text = original.strip()

    # 压缩多余空白（连续空格/换行 → 单个）
    import re
    text = re.sub(r"\s+", " ", text)

    # 繁体→简体
    try:
        import zhconv
        simplified = zhconv.convert(text, "zh-cn")
        if simplified != text:
            text = simplified
    except ImportError:
        logger.warning("zhconv not installed, skip traditional→simplified")

    # 规整标点：中文引号统一
    text = text.replace("\u201c", "\u201c")  # left double
    text = text.replace("\u201d", "\u201d")  # right double

    changed = text != original
    return NormalizeResponse(content=text, changed=changed)
