"""
文本标准化：trim + 繁体→简体 + 标点规整
"""

import logging

from fastapi import APIRouter
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

    # ================================================================
    # ⑦ 非必要外壳 — 去壳留核
    # "XX说过/认为/觉得/看来/表示：xxx" → "xxx"
    # "我认为/我觉得/在我看来，xxx" → "xxx"
    # ================================================================
    shell_patterns = [
        # "XX说过/认为/觉得/表示/强调/指出/分享/总结：xxx"
        re.compile(r'^.{1,8}(?:说过|认为|觉得|看来|表示|强调|指出|分享|总结|提到|写道)[：:，,]\s*'),
        # "我/笔者 认为/觉得/看来/发现/总结，xxx"
        re.compile(r'^(?:我|笔者|个人)(?:认为|觉得|看来|以为|发现|总结|的感受?是)[：:，,\s]*'),
        # "在我看来/依我看/就我而言，xxx"
        re.compile(r'^(?:在我看来|依我看|就我而言|对我而言|就我个人而言|个人认为|个人觉得)[，,]\s*'),
        # "有人说过/有人说/常言道/俗话说，xxx"
        re.compile(r'^(?:有人说过?|有人说|常言道|俗话说|老话说|古人云)[：:，,]\s*'),
    ]
    prev = None
    while text != prev:
        prev = text
        for pat in shell_patterns:
            text = pat.sub('', text)

    # ⑧ 格式杂质
    # 引号+署名混排："内容"——作者 或 "内容"-作者
    text = re.sub(r'^\u201c(.+?)\u201d\s*[-——–—]+\s*.{1,15}$', r'\1', text)
    # 纯引号包裹："内容" → 内容
    text = re.sub(r'^\u201c(.+?)\u201d$', r'\1', text)
    text = re.sub(r'^\u2018(.+?)\u2019$', r'\1', text)
    # 半角引号包裹
    text = re.sub(r'^"(.+?)"$', r'\1', text)
    # 署名后缀：内容——作者 / 内容-作者 / 内容—作者
    text = re.sub(r'\s*[-——–—]+\s*.{1,15}$', '', text)
    # 列表符开头（1. 2、 ① • - · ）
    text = re.sub(r'^[\d]+[\.\、\)\)]\s*', '', text)
    text = re.sub(r'^[①②③④⑤⑥⑦⑧⑨⑩]\s*', '', text)
    text = re.sub(r'^[•\-\*\·\►\▸\▪]\s*', '', text)

    # Final trim after all cleaning
    text = text.strip()

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
