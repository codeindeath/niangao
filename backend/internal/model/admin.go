package model

import "time"

// ============================================================
// Admin Dashboard
// ============================================================

type AdminDashboard struct {
	TotalUsers      int64 `json:"total_users"`
	TotalExperiences int64 `json:"total_experiences"`
	TodayNewUsers   int64 `json:"today_new_users"`
	TodayNewExps    int64 `json:"today_new_exps"`
	TodayActiveUsers int64 `json:"today_active_users"`
	TodayAIChats    int64 `json:"today_ai_chats"`
	PendingReviews  int64 `json:"pending_reviews"`
	TodayApproved   int64 `json:"today_approved"`
	TodayRejected   int64 `json:"today_rejected"`
}

type Trend struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type AIStatus struct {
	Model    string  `json:"model"`
	Healthy  bool    `json:"healthy"`
	LatencyMs float64 `json:"latency_ms"`
	SuccessRate float64 `json:"success_rate"`
	LastHourCalls int64 `json:"last_hour_calls"`
}

// ============================================================
// Admin Review
// ============================================================

type ReviewItem struct {
	ID           string  `json:"id"`
	Content      string  `json:"content"`
	Domain       string  `json:"domain"`
	SubDomain    *string `json:"sub_domain"`
	SourceType   string  `json:"source_type"`
	ReviewStatus string  `json:"review_status"`
	AIVerdict    *string `json:"ai_verdict"`
	AIScore      *float64 `json:"ai_score"`
	AIScoreDetail *string `json:"ai_score_detail"`
	AIInterpretation *string `json:"ai_interpretation"`
	HardPolicyResult *string `json:"hard_policy_result"`
	AuthorName   string  `json:"author_name"`
	SubmittedAt  time.Time `json:"submitted_at"`
}

type ReviewListQuery struct {
	Domain     string `form:"domain"`
	SourceType string `form:"source_type"`
	AIVerdict  string `form:"ai_verdict"`
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
}

type ReviewAction struct {
	Action string  `json:"action"` // "approve" | "reject"
	Reason *string `json:"reason"`
}

type BatchReview struct {
	IDs    []string `json:"ids"`
	Action string   `json:"action"`
	Reason *string  `json:"reason"`
}

// ============================================================
// Admin User
// ============================================================

type AdminUserItem struct {
	ID          string     `json:"id"`
	Nickname    string     `json:"nickname"`
	AvatarURL   *string    `json:"avatar_url"`
	Title       *string    `json:"title"`
	AuthProvider string    `json:"auth_provider"`
	IsAdmin     bool       `json:"is_admin"`
	IsActive    bool       `json:"is_active"`
	ExpCount    int        `json:"exp_count"`
	CreatedAt   time.Time  `json:"created_at"`
	LastLoginAt *time.Time `json:"last_login_at"`
}

type AdminUserDetail struct {
	AdminUserItem
	Bio            *string `json:"bio"`
	LikeReceived   int     `json:"like_received"`
	BookmarkReceived int   `json:"bookmark_received"`
	ViewedCount    int     `json:"viewed_count"`
	LikedCount     int     `json:"liked_count"`
	BookmarkedCount int    `json:"bookmarked_count"`
	ChatCount      int     `json:"chat_count"`
	MsgCount       int     `json:"msg_count"`
}

type UserStatusUpdate struct {
	Active bool   `json:"active"`
	Reason *string `json:"reason"`
}

// ============================================================
// Admin Domain
// ============================================================

// DomainSortOrder defines the display order of top-level domains (loaded from DB).
var DomainSortOrder = []Domain{} // Populated at runtime from DomainCatalog

type DomainItem struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Icon        string `json:"icon"`
	ParentName  *string `json:"parent_name"`
	ExpCount    int    `json:"exp_count"`
	SortOrder   int    `json:"sort_order"`
	Active      bool   `json:"active"`
}

type DomainCreate struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Icon        string `json:"icon"`
	ParentName  *string `json:"parent_name"`
}

type DomainReorder struct {
	Names     []string `json:"names"`
	ParentName *string `json:"parent_name"`
}

// ============================================================
// Admin Config
// ============================================================

type ConfigUpdate struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// ============================================================
// Admin Log
// ============================================================

type AdminLogItem struct {
	ID         string    `json:"id"`
	AdminID    string    `json:"admin_id"`
	AdminName  string    `json:"admin_name"`
	ActionType string    `json:"action_type"`
	TargetType *string   `json:"target_type"`
	TargetID   *string   `json:"target_id"`
	Detail     *string   `json:"detail"`
	Result     string    `json:"result"`
	CreatedAt  time.Time `json:"created_at"`
}

type AdminLogQuery struct {
	AdminID    string `form:"admin_id"`
	ActionType string `form:"action_type"`
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
}

// ============================================================
// Common
// ============================================================

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
}
