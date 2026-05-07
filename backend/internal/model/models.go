package model

import "time"

// ============================================================
// Review status
// ============================================================
type ReviewStatus string

const (
	ReviewPending  ReviewStatus = "pending"
	ReviewApproved ReviewStatus = "approved"
	ReviewRejected ReviewStatus = "rejected"
	ReviewPrivate  ReviewStatus = "private"
)

// ============================================================
// Structs
// ============================================================

type User struct {
	ID              string  `json:"id"`
	AppleUserID     *string `json:"-"`
	Nickname        string  `json:"nickname"`
	AvatarURL       *string `json:"avatar_url,omitempty"`
	Bio             *string `json:"bio,omitempty"`
	Title           *string `json:"title,omitempty"`
	ExperienceCount int     `json:"experience_count"`
	BookmarkCount   int     `json:"bookmark_count"`
	PracticedCount  int     `json:"practiced_count"`
}

type Experience struct {
	ID                      string    `json:"id"`
	AuthorID                string    `json:"author_id"`
	Content                 string    `json:"content"`
	Interpretation          *string   `json:"interpretation,omitempty"`
	Domain                  Domain    `json:"domain"`
	SubDomain               string    `json:"sub_domain,omitempty"`
	IsOfficial              bool      `json:"is_official"`
	IsPrivate               bool      `json:"is_private"`
	SourceLabel             *string   `json:"source_label,omitempty"`
	LikeCount               int       `json:"like_count"`
	BookmarkCount           int       `json:"bookmark_count"`
	InterpretationGenerated bool      `json:"interpretation_generated"`
	Status                  string    `json:"status"`
	ReviewStatus            string    `json:"review_status"`
	ReviewReason            *string   `json:"review_reason,omitempty"`
	QualityScore            *float64  `json:"quality_score,omitempty"`
	ScoreDetails            *string   `json:"score_details,omitempty"`
	CreatorName             *string   `json:"creator_name,omitempty"`
	SourceType              string    `json:"source_type"`
	ScoreReason             *string   `json:"score_reason,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
	DeletedAt               *time.Time `json:"deleted_at,omitempty"`
	// Joined fields
	AuthorName   *string `json:"author_name,omitempty"`
	AuthorAvatar *string `json:"author_avatar,omitempty"`
	AuthorTitle  *string `json:"author_title,omitempty"`
	IsLiked      bool    `json:"is_liked"`
	IsBookmarked bool    `json:"is_bookmarked"`
}

type CreateExperienceRequest struct {
	Content        string    `json:"content" binding:"required"`
	Domain         Domain    `json:"domain" binding:"required"`
	SubDomain      SubDomain `json:"sub_domain" binding:"required"`
	Interpretation string    `json:"interpretation"`
	IsPrivate      bool      `json:"is_private"`
}

type ExperienceListQuery struct {
	Domain    Domain `form:"domain"`
	SubDomain string `form:"sub_domain"`
	Sort      string `form:"sort"`
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	Search    string `form:"search"`
}

type Conversation struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     *string   `json:"title,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Message struct {
	ID                      string    `json:"id"`
	ConversationID          string    `json:"conversation_id"`
	Role                    string    `json:"role"`
	Content                 string    `json:"content"`
	ReferencedExperienceIDs []string  `json:"referenced_experience_ids,omitempty"`
	CreatedAt               time.Time `json:"created_at"`
}

type ChatRequest struct {
	ConversationID string `json:"conversation_id,omitempty"`
	Message        string `json:"message" binding:"required"`
}
