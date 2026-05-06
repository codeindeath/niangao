# 年糕管理后台 — 实现与部署方案 v1.0

## 一、技术选型

| 层 | 技术 | 理由 |
|---|------|------|
| **前端** | React 18 + TypeScript + Vite | 与移动端共享 TypeScript 技能栈，Vite 极速 HMR |
| **UI 框架** | 自建组件（CSS Modules） | 10 个模块 ~25 个页面，Ant Design 过重。从设计稿直接写，体积小、定制自由 |
| **图表** | Recharts | React 原生、声明式、中文支持好、轻量 |
| **路由** | React Router v6 | 标准方案 |
| **状态管理** | React Context + useReducer | 管理后台状态不复杂，无需 Redux |
| **HTTP** | fetch + 拦截器（token 注入） | 复用移动端 api.ts 模式 |
| **后端** | Go (Gin) — 复用现有后端 | 新增 `/api/v1/admin/*` 路由组 |
| **数据库** | PostgreSQL — 新增 `admin_logs` 表 | 复用现有连接池 |
| **部署** | Nginx 静态文件 + Go API 同 ECS | 无需新服务器 |

---

## 二、架构图

```
┌──────────────────────────────────────────────────┐
│                    ECS (现有)                      │
│                                                    │
│  ┌──────────┐   ┌──────────┐   ┌──────────────┐  │
│  │ Nginx :80│   │ Go :8080 │   │ Python :8000  │  │
│  │          │   │          │   │              │  │
│  │ /        │──▶│ /api/v1/*│   │ /ai/*        │  │
│  │ /admin/* │   │ /api/v1/ │   │              │  │
│  │          │   │   admin/*│◀──│              │  │
│  └──────────┘   └────┬─────┘   └──────────────┘  │
│                      │                             │
│                 ┌────▼────┐                        │
│                 │   PG 14  │                        │
│                 └─────────┘                        │
└──────────────────────────────────────────────────┘

前端部署：Vite build → dist/ → scp 到 ECS /var/www/admin/
Nginx 配置：location /admin { root /var/www; try_files $uri /admin/index.html; }
```

---

## 三、后端实现计划

### 3.1 数据库变更

```sql
-- 管理员标记
ALTER TABLE users ADD COLUMN is_admin BOOLEAN DEFAULT FALSE;

-- 操作日志表
CREATE TABLE admin_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    admin_id UUID REFERENCES users(id),
    action_type VARCHAR(50) NOT NULL,     -- 'review.approve' | 'content.edit' | 'user.disable' | ...
    target_type VARCHAR(50),              -- 'experience' | 'user' | 'config'
    target_id UUID,
    detail JSONB,                         -- 操作详情
    result VARCHAR(20) DEFAULT 'success', -- 'success' | 'failure'
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_admin_logs_admin ON admin_logs(admin_id);
CREATE INDEX idx_admin_logs_time ON admin_logs(created_at DESC);
CREATE INDEX idx_admin_logs_type ON admin_logs(action_type);
```

### 3.2 新增代码文件

```
backend/
├── cmd/server/main.go              (修改 — 注册 admin 路由)
├── internal/
│   ├── middleware/
│   │   └── admin.go                (新增 — RequireAdmin 中间件)
│   ├── handler/
│   │   ├── admin_dashboard.go      (新增)
│   │   ├── admin_users.go          (新增)
│   │   ├── admin_reviews.go        (新增)
│   │   ├── admin_experiences.go    (新增)
│   │   ├── admin_domains.go        (新增)
│   │   ├── admin_stats.go          (新增)
│   │   ├── admin_config.go         (新增)
│   │   └── admin_logs.go           (新增)
│   ├── repository/
│   │   ├── admin_dashboard.go      (新增)
│   │   ├── admin_users.go          (新增)
│   │   └── admin_logs.go           (新增)
│   └── model/
│       └── admin.go                (新增 — Admin 相关类型)
└── migrations/
    └── 011_admin_system.sql        (新增)
```

### 3.3 Admin API 路由设计

```
v1.Group("/admin", middleware.RequireAdmin())
{
    // 仪表盘
    admin.GET("/dashboard", h.Dashboard)

    // 用户管理
    admin.GET("/users", h.ListUsers)
    admin.GET("/users/:id", h.GetUser)
    admin.GET("/users/:id/experiences", h.GetUserExperiences)
    admin.PUT("/users/:id/status", h.UpdateUserStatus)       // 禁用/启用

    // 内容审核
    admin.GET("/reviews", h.ListReviews)                      // 审核队列
    admin.GET("/reviews/:id", h.GetReviewDetail)              // 含AI打分明细
    admin.POST("/reviews/:id/approve", h.ApproveReview)       // 通过
    admin.POST("/reviews/:id/reject", h.RejectReview)         // 拒绝
    admin.POST("/reviews/:id/retry", h.RetryAIReview)         // 退回重审
    admin.POST("/reviews/batch", h.BatchReview)               // 批量审核

    // 内容管理
    admin.GET("/experiences", h.ListAllExperiences)           // 全量列表
    admin.GET("/experiences/:id", h.GetExperienceDetail)
    admin.PUT("/experiences/:id", h.UpdateExperience)         // 编辑
    admin.DELETE("/experiences/:id", h.DeleteExperience)      // 软删
    admin.POST("/experiences/:id/restore", h.RestoreExperience)
    admin.POST("/experiences/batch", h.BatchUpdateExperiences)

    // 平台内容
    admin.GET("/platform-experiences", h.ListPlatform)
    admin.POST("/platform-experiences", h.CreatePlatform)
    admin.PUT("/platform-experiences/:id", h.UpdatePlatform)
    admin.POST("/platform-experiences/import", h.ImportPlatform)
    admin.POST("/platform-experiences/batch-ai", h.BatchAIScore)

    // 领域管理
    admin.GET("/domains", h.ListDomains)
    admin.POST("/domains", h.CreateDomain)
    admin.PUT("/domains/:id", h.UpdateDomain)
    admin.PUT("/domains/reorder", h.ReorderDomains)

    // AI 服务
    admin.GET("/ai/status", h.GetAIStatus)
    admin.GET("/ai/stats", h.GetAIStats)

    // 数据统计
    admin.GET("/stats/users", h.StatsUsers)
    admin.GET("/stats/experiences", h.StatsExperiences)
    admin.GET("/stats/interactions", h.StatsInteractions)
    admin.GET("/stats/reviews", h.StatsReviews)
    admin.GET("/stats/retention", h.StatsRetention)

    // 系统配置
    admin.GET("/config", h.GetConfig)
    admin.PUT("/config", h.UpdateConfig)
    admin.GET("/config/sensitive-words", h.ListSensitiveWords)
    admin.POST("/config/sensitive-words", h.AddSensitiveWord)
    admin.DELETE("/config/sensitive-words/:id", h.RemoveSensitiveWord)

    // 操作日志
    admin.GET("/logs", h.ListLogs)
}
```

### 3.4 分钟级中间件

```go
func RequireAdmin() gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        var isAdmin bool
        err := db.QueryRow(ctx, "SELECT is_admin FROM users WHERE id=$1", userID).Scan(&isAdmin)
        if err != nil || !isAdmin {
            c.AbortWithStatusJSON(403, gin.H{"error": "需要管理员权限"})
            return
        }
        c.Next()
    }
}
```

---

## 四、前端实现计划

### 4.1 项目结构

```
admin/
├── package.json
├── vite.config.ts
├── tsconfig.json
├── index.html
├── src/
│   ├── main.tsx
│   ├── App.tsx                    (路由 + 布局)
│   ├── api/
│   │   ├── client.ts              (fetch 封装 + token 注入)
│   │   └── endpoints.ts           (API 方法)
│   ├── components/
│   │   ├── Layout.tsx             (侧边栏 + 顶栏 + 内容区)
│   │   ├── Sidebar.tsx            (导航菜单)
│   │   ├── StatCard.tsx           (仪表盘指标卡片)
│   │   ├── DataTable.tsx          (通用表格 + 分页)
│   │   ├── FilterBar.tsx          (搜索 + 筛选栏)
│   │   ├── Modal.tsx              (通用弹窗)
│   │   ├── Badge.tsx              (状态标签)
│   │   ├── Chart.tsx              (图表封装 based on Recharts)
│   │   └── ConfirmDialog.tsx      (确认对话框)
│   ├── pages/
│   │   ├── Dashboard.tsx          (总览)
│   │   ├── reviews/
│   │   │   ├── ReviewQueue.tsx    (审核队列)
│   │   │   └── ReviewDetail.tsx   (审核详情弹窗)
│   │   ├── experiences/
│   │   │   ├── ExperienceList.tsx (内容列表)
│   │   │   └── ExperienceEdit.tsx (编辑弹窗)
│   │   ├── platform/
│   │   │   ├── PlatformList.tsx   (平台内容)
│   │   │   ├── PlatformCreate.tsx (新建)
│   │   │   └── PlatformImport.tsx (批量导入)
│   │   ├── users/
│   │   │   ├── UserList.tsx       (用户列表)
│   │   │   └── UserDetail.tsx     (用户详情)
│   │   ├── domains/
│   │   │   └── DomainManage.tsx   (领域管理)
│   │   ├── ai/
│   │   │   └── AIService.tsx      (AI 服务)
│   │   ├── stats/
│   │   │   └── StatsPage.tsx      (数据统计)
│   │   ├── config/
│   │   │   └── SettingsPage.tsx   (系统配置)
│   │   └── logs/
│   │       └── LogList.tsx        (操作日志)
│   ├── hooks/
│   │   ├── useAuth.ts             (认证 hook)
│   │   └── usePagination.ts       (分页 hook)
│   └── styles/
│       └── global.css             (全局样式 + CSS 变量)
```

### 4.2 页面优先级（分阶段交付）

| 阶段 | 页面 | 预计 |
|------|------|------|
| **P0** | 登录页 + 总览仪表盘 + 内容审核（队列+详情+操作） | 核心运营闭环 |
| **P1** | 内容管理（列表+编辑+状态） + 平台内容管理 | 内容维护 |
| **P2** | 用户管理 + 领域管理 | 基础管理 |
| **P3** | 数据统计 + 系统配置 + 操作日志 + AI 服务 | 完整后台 |

---

## 五、部署方案

### 5.1 部署配置

```nginx
# /etc/nginx/sites-available/niangao (追加)
server {
    listen 80;
    server_name niangao.app;  # 或 IP

    # 管理后台前端
    location /admin {
        alias /var/www/admin/;
        try_files $uri $uri/ /admin/index.html;
    }

    # 管理后台 API (复用现有)
    location /api/v1/admin/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # 现有路由保持不变
    location /api/ { proxy_pass http://127.0.0.1:8080; }
    location /ai/  { proxy_pass http://127.0.0.1:8000; }
}
```

### 5.2 部署命令

```bash
# 前端构建
cd admin && npm run build   # → dist/

# 部署到 ECS
scp -r dist/* root@115.190.177.146:/var/www/admin/

# 后端部署（同现有流程）
cd backend && go build -o server ./cmd/server/
scp server root@115.190.177.146:/root/niangao/backend/
ssh root@115.190.177.146 'systemctl restart niangao-backend'
```

### 5.3 认证方案

- 管理后台独立登录页（`/admin/login`）
- 使用现有 `POST /api/v1/auth/admin/login` 端点（新增）
- 登录成功后在 localStorage 存 JWT token
- 所有 admin API 请求携带 `Authorization: Bearer <token>`
- Go 中间件验证 token + `is_admin` 字段

---

## 六、开发估时

| 阶段 | 内容 | 估时 |
|------|------|------|
| 后端基础 | DB迁移 + admin中间件 + 仪表盘API | 0.5 天 |
| 后端 P0 | 审核API + 内容管理API | 1 天 |
| 后端 P1-P3 | 用户/平台/领域/统计/配置/日志 API | 1.5 天 |
| 前端搭建 | Vite项目 + 布局 + 路由 + 通用组件 | 0.5 天 |
| 前端 P0 | 登录 + 仪表盘 + 审核页 | 1 天 |
| 前端 P1 | 内容管理 + 平台内容 | 1 天 |
| 前端 P2-P3 | 用户/领域/统计/配置/日志 | 1.5 天 |
| 联调 + 部署 | 前后端联调 + ECS 部署 + 验证 | 0.5 天 |
| **合计** | | **7.5 天** |

---

## 七、与现有系统的关系

| 维度 | 现有系统 | 管理后台 |
|------|---------|---------|
| 代码仓库 | `niangao/backend/` | 同仓库，新增文件 |
| 数据库 | PostgreSQL | 同库，新增 admin_logs 表 |
| 认证 | JWT (users 表) | 同 JWT + is_admin 字段 |
| 部署 | ECS 单机 | 同 ECS，Nginx 新增路由 |
| 前端仓库 | `niangao/mobile/` | `niangao/admin/`（新建目录） |
| 技术栈 | React Native | React Web（共享 TypeScript 技能） |
