-- 015_dynamic_domains: 10×10 动态领域体系
-- 替代旧的硬编码 ENUM 领域系统

-- 领域表：一级 + 二级统一存储
CREATE TABLE domains (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(50) NOT NULL,
    parent_name VARCHAR(50) REFERENCES domains(name) ON DELETE CASCADE,
    sort_order INT DEFAULT 0,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_domains_parent ON domains(parent_name);
CREATE INDEX idx_domains_active ON domains(active);

-- 种子数据：10 一级领域 × 10 二级领域
INSERT INTO domains (name, display_name, parent_name, sort_order) VALUES
-- 1. 职场进阶
('career', '职场进阶', NULL, 1),
('career-planning', '职业规划', 'career', 1),
('managing-up', '向上管理', 'career', 2),
('job-switch', '跳槽转型', 'career', 3),
('workplace-talk', '职场沟通', 'career', 4),
('team-lead', '团队管理', 'career', 5),
('productivity-tools', '效率工具', 'career', 6),
('side-project', '副业探索', 'career', 7),
('interview', '面试技巧', 'career', 8),
('salary-negotiation', '薪资谈判', 'career', 9),
('remote-work', '远程办公', 'career', 10),

-- 2. 创业经营
('business', '创业经营', NULL, 2),
('business-model', '商业模式', 'business', 1),
('product-thinking', '产品思维', 'business', 2),
('marketing', '市场营销', 'business', 3),
('fundraising', '融资策略', 'business', 4),
('hiring', '招聘用人', 'business', 5),
('company-culture', '企业文化', 'business', 6),
('competitive-strategy', '竞争策略', 'business', 7),
('cost-control', '成本控制', 'business', 8),
('growth-hacking', '增长方法', 'business', 9),
('exit-strategy', '退出策略', 'business', 10),

-- 3. 思维模型
('cognition', '思维模型', NULL, 3),
('first-principles', '第一性原理', 'cognition', 1),
('systems-thinking', '系统思维', 'cognition', 2),
('compound-effect', '复利思维', 'cognition', 3),
('inversion', '逆向思维', 'cognition', 4),
('probabilistic-thinking', '概率思维', 'cognition', 5),
('occams-razor', '奥卡姆剃刀', 'cognition', 6),
('opportunity-cost', '机会成本', 'cognition', 7),
('mental-frameworks', '框架思考', 'cognition', 8),
('second-order', '二阶效应', 'cognition', 9),
('antifragile', '反脆弱', 'cognition', 10),

-- 4. 人际沟通
('relationship', '人际沟通', NULL, 4),
('persuasion', '说服技巧', 'relationship', 1),
('conflict-resolution', '冲突处理', 'relationship', 2),
('public-speaking', '公开演讲', 'relationship', 3),
('negotiation', '谈判策略', 'relationship', 4),
('social-networking', '社交网络', 'relationship', 5),
('intimate-relationship', '亲密关系', 'relationship', 6),
('family-bond', '家庭关系', 'relationship', 7),
('empathic-listening', '共情倾听', 'relationship', 8),
('feedback-art', '反馈艺术', 'relationship', 9),
('boundary-setting', '边界设定', 'relationship', 10),

-- 5. 自我管理
('self-mastery', '自我管理', NULL, 5),
('habit-building', '习惯养成', 'self-mastery', 1),
('time-management', '时间管理', 'self-mastery', 2),
('deep-focus', '专注力', 'self-mastery', 3),
('self-motivation', '自驱力', 'self-mastery', 4),
('procrastination', '拖延克服', 'self-mastery', 5),
('goal-setting', '目标设定', 'self-mastery', 6),
('energy-management', '精力管理', 'self-mastery', 7),
('decluttering', '断舍离', 'self-mastery', 8),
('discipline-system', '自律系统', 'self-mastery', 9),
('morning-routine', '晨间仪式', 'self-mastery', 10),

-- 6. 财富自由
('wealth', '财富自由', NULL, 6),
('investing', '投资理财', 'wealth', 1),
('asset-allocation', '资产配置', 'wealth', 2),
('passive-income', '被动收入', 'wealth', 3),
('money-mindset', '消费观念', 'wealth', 4),
('risk-management', '风险控制', 'wealth', 5),
('tax-planning', '税务规划', 'wealth', 6),
('real-estate', '房产策略', 'wealth', 7),
('index-fund', '基金定投', 'wealth', 8),
('financial-independence', '财务独立', 'wealth', 9),
('money-psychology', '金钱心理', 'wealth', 10),

-- 7. 身心健康
('wellness', '身心健康', NULL, 7),
('exercise', '运动健身', 'wellness', 1),
('sleep-optimization', '睡眠优化', 'wellness', 2),
('nutrition', '饮食营养', 'wellness', 3),
('meditation', '冥想正念', 'wellness', 4),
('stress-management', '压力管理', 'wellness', 5),
('emotion-regulation', '情绪调节', 'wellness', 6),
('mental-resilience', '心理韧性', 'wellness', 7),
('energy-optimization', '能量管理', 'wellness', 8),
('body-awareness', '身体信号', 'wellness', 9),
('bad-habits-quit', '戒断坏习惯', 'wellness', 10),

-- 8. 学习成长
('learning', '学习成长', NULL, 8),
('reading-method', '阅读方法', 'learning', 1),
('knowledge-system', '知识体系', 'learning', 2),
('feynman-technique', '费曼技巧', 'learning', 3),
('spaced-repetition', '间隔重复', 'learning', 4),
('cross-discipline', '跨界学习', 'learning', 5),
('practice-feedback', '实践反馈', 'learning', 6),
('mentorship', '导师关系', 'learning', 7),
('deliberate-practice', '刻意练习', 'learning', 8),
('info-filtering', '信息筛选', 'learning', 9),
('metacognition', '元认知', 'learning', 10),

-- 9. 生活哲学
('life-philosophy', '生活哲学', NULL, 9),
('meaning-of-life', '人生意义', 'life-philosophy', 1),
('happiness-essence', '幸福本质', 'life-philosophy', 2),
('minimalism', '极简主义', 'life-philosophy', 3),
('stoicism', '斯多葛', 'life-philosophy', 4),
('acceptance', '接受无常', 'life-philosophy', 5),
('present-moment', '活在当下', 'life-philosophy', 6),
('choice-wisdom', '选择智慧', 'life-philosophy', 7),
('letting-go', '放下执念', 'life-philosophy', 8),
('gratitude', '感恩练习', 'life-philosophy', 9),
('death-awareness', '死亡意识', 'life-philosophy', 10),

-- 10. 创造表达
('creativity', '创造表达', NULL, 10),
('writing-output', '写作输出', 'creativity', 1),
('content-creation', '内容创作', 'creativity', 2),
('product-design', '产品设计', 'creativity', 3),
('art-perception', '艺术感知', 'creativity', 4),
('storytelling', '讲故事', 'creativity', 5),
('brand-building', '品牌建设', 'creativity', 6),
('visual-expression', '视觉表达', 'creativity', 7),
('coding-create', '编程创造', 'creativity', 8),
('music-insight', '音乐感悟', 'creativity', 9),
('innovation-method', '创新方法', 'creativity', 10);
