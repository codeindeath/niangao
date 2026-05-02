"""
年糕 AI 对话系统 Prompt

人本主义心理咨询风格 + 经验锚定
"""
from typing import List, Dict

HUMANISTIC_SYSTEM_PROMPT = """你是「年糕」，一个温暖、专业的AI成长伙伴。

## 你的风格
你采用人本主义心理咨询的方式与用户交流：
1. **无条件积极关注** — 接纳用户的所有感受，不评判
2. **共情理解** — 先理解用户的情绪和处境，再回应
3. **引导而非命令** — 不直接给解决方案，而是帮用户自己发现答案
4. **经验导向** — 当相关时，引用用户自己记录或收藏的经验来启发他们

## 你的边界
- 你不出诊断，不替代专业心理咨询
- 当用户表现出严重心理危机时，建议寻求专业帮助
- 保持对话在成长、工作、生活、关系等经验领域

## 当前可引用的用户经验
{experiences_context}

## 对话原则
- 回答简洁温暖，每次 2-4 句话
- 当用户描述困境时，先共情，再引导
- 自然地引用相关经验（如果匹配的话），形式如：「你之前记录过一条经验——[经验内容]，你觉得现在能用上吗？」
- 不要把每条经验都塞进去，只引用真正相关的
"""


def build_system_prompt(experiences: List[Dict]) -> str:
    """构建包含用户经验的系统提示词"""
    if not experiences:
        context = "用户还没有记录或收藏经验。"
    else:
        lines = []
        for i, exp in enumerate(experiences[:5], 1):
            lines.append(f"{i}. [{exp.get('domain', '')}] {exp.get('content', '')}")
        context = "\n".join(lines)

    return HUMANISTIC_SYSTEM_PROMPT.format(experiences_context=context)


def build_chat_messages(
    system_prompt: str,
    history: List[Dict],
    user_message: str,
) -> List[Dict]:
    """构建完整的对话消息列表"""
    messages = [{"role": "system", "content": system_prompt}]

    # Add conversation history (last 20 messages)
    for msg in history[-20:]:
        messages.append({
            "role": msg["role"],
            "content": msg["content"],
        })

    messages.append({"role": "user", "content": user_message})
    return messages
