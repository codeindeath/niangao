"""领域标签 — 10×10 动态体系，与 DB domains 表同步"""
from typing import Dict, List

# 一二级领域完整定义
DOMAIN_TREE: Dict[str, Dict] = {
    "career": {
        "label": "职场进阶",
        "subs": {
            "career-planning": "职业规划", "managing-up": "向上管理",
            "job-switch": "跳槽转型", "workplace-talk": "职场沟通",
            "team-lead": "团队管理", "productivity-tools": "效率工具",
            "side-project": "副业探索", "interview": "面试技巧",
            "salary-negotiation": "薪资谈判", "remote-work": "远程办公",
        },
    },
    "business": {
        "label": "创业经营",
        "subs": {
            "business-model": "商业模式", "product-thinking": "产品思维",
            "marketing": "市场营销", "fundraising": "融资策略",
            "hiring": "招聘用人", "company-culture": "企业文化",
            "competitive-strategy": "竞争策略", "cost-control": "成本控制",
            "growth-hacking": "增长方法", "exit-strategy": "退出策略",
        },
    },
    "cognition": {
        "label": "思维模型",
        "subs": {
            "first-principles": "第一性原理", "systems-thinking": "系统思维",
            "compound-effect": "复利思维", "inversion": "逆向思维",
            "probabilistic-thinking": "概率思维", "occams-razor": "奥卡姆剃刀",
            "opportunity-cost": "机会成本", "mental-frameworks": "框架思考",
            "second-order": "二阶效应", "antifragile": "反脆弱",
        },
    },
    "relationship": {
        "label": "人际沟通",
        "subs": {
            "persuasion": "说服技巧", "conflict-resolution": "冲突处理",
            "public-speaking": "公开演讲", "negotiation": "谈判策略",
            "social-networking": "社交网络", "intimate-relationship": "亲密关系",
            "family-bond": "家庭关系", "empathic-listening": "共情倾听",
            "feedback-art": "反馈艺术", "boundary-setting": "边界设定",
        },
    },
    "self-mastery": {
        "label": "自我管理",
        "subs": {
            "habit-building": "习惯养成", "time-management": "时间管理",
            "deep-focus": "专注力", "self-motivation": "自驱力",
            "procrastination": "拖延克服", "goal-setting": "目标设定",
            "energy-management": "精力管理", "decluttering": "断舍离",
            "discipline-system": "自律系统", "morning-routine": "晨间仪式",
        },
    },
    "wealth": {
        "label": "财富自由",
        "subs": {
            "investing": "投资理财", "asset-allocation": "资产配置",
            "passive-income": "被动收入", "money-mindset": "消费观念",
            "risk-management": "风险控制", "tax-planning": "税务规划",
            "real-estate": "房产策略", "index-fund": "基金定投",
            "financial-independence": "财务独立", "money-psychology": "金钱心理",
        },
    },
    "wellness": {
        "label": "身心健康",
        "subs": {
            "exercise": "运动健身", "sleep-optimization": "睡眠优化",
            "nutrition": "饮食营养", "meditation": "冥想正念",
            "stress-management": "压力管理", "emotion-regulation": "情绪调节",
            "mental-resilience": "心理韧性", "energy-optimization": "能量管理",
            "body-awareness": "身体信号", "bad-habits-quit": "戒断坏习惯",
        },
    },
    "learning": {
        "label": "学习成长",
        "subs": {
            "reading-method": "阅读方法", "knowledge-system": "知识体系",
            "feynman-technique": "费曼技巧", "spaced-repetition": "间隔重复",
            "cross-discipline": "跨界学习", "practice-feedback": "实践反馈",
            "mentorship": "导师关系", "deliberate-practice": "刻意练习",
            "info-filtering": "信息筛选", "metacognition": "元认知",
        },
    },
    "life-philosophy": {
        "label": "生活哲学",
        "subs": {
            "meaning-of-life": "人生意义", "happiness-essence": "幸福本质",
            "minimalism": "极简主义", "stoicism": "斯多葛",
            "acceptance": "接受无常", "present-moment": "活在当下",
            "choice-wisdom": "选择智慧", "letting-go": "放下执念",
            "gratitude": "感恩练习", "death-awareness": "死亡意识",
        },
    },
    "creativity": {
        "label": "创造表达",
        "subs": {
            "writing-output": "写作输出", "content-creation": "内容创作",
            "product-design": "产品设计", "art-perception": "艺术感知",
            "storytelling": "讲故事", "brand-building": "品牌建设",
            "visual-expression": "视觉表达", "coding-create": "编程创造",
            "music-insight": "音乐感悟", "innovation-method": "创新方法",
        },
    },
}


def get_domain_label(name: str) -> str:
    if name in DOMAIN_TREE:
        return DOMAIN_TREE[name]["label"]
    return name


def get_sub_label(domain: str, sub: str) -> str:
    if domain in DOMAIN_TREE:
        return DOMAIN_TREE[domain]["subs"].get(sub, sub)
    return sub


def get_domain_list_for_prompt() -> str:
    """生成注入 AI prompt 的领域列表"""
    lines = []
    for dk, dv in DOMAIN_TREE.items():
        lines.append(f"  {dk}（{dv['label']}）→ {', '.join(dv['subs'].keys())}")
    return "\n".join(lines)


def get_all_subdomains() -> List[str]:
    """返回全部子领域名列表"""
    subs = []
    for dv in DOMAIN_TREE.values():
        subs.extend(dv["subs"].keys())
    return subs
