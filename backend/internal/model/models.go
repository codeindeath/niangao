package model

import "time"

type Domain string

const (
	DomainCareer       Domain = "career"
	DomainRelationship Domain = "relationship"
	DomainCognition    Domain = "cognition"
	DomainLife         Domain = "life"
	DomainEmotion      Domain = "emotion"
)

type User struct {
	ID              string  `json:"id"`
	AppleUserID     *string `json:"-"`
	Nickname        string  `json:"nickname"`
	AvatarURL       *string `json:"avatar_url,omitempty"`
	Bio             *string `json:"bio,omitempty"`
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
	IsOfficial              bool      `json:"is_official"`
	SourceLabel             *string   `json:"source_label,omitempty"`
	LikeCount               int       `json:"like_count"`
	BookmarkCount           int       `json:"bookmark_count"`
	InterpretationGenerated bool      `json:"interpretation_generated"`
	Status                  string    `json:"status"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
	// Joined fields
	AuthorName   *string `json:"author_name,omitempty"`
	AuthorAvatar *string `json:"author_avatar,omitempty"`
	IsLiked      bool    `json:"is_liked"`
	IsBookmarked bool    `json:"is_bookmarked"`
}

type CreateExperienceRequest struct {
	Content        string `json:"content" binding:"required,max=100"`
	Interpretation string `json:"interpretation" binding:"max=500"`
	Domain         Domain `json:"domain" binding:"required"`
}

type ExperienceListQuery struct {
	Domain   Domain `form:"domain"`
	Sort     string `form:"sort"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Search   string `form:"search"`
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

var ValidDomains = map[Domain]string{
	DomainCareer:       "职场成长",
	DomainRelationship: "人际关系",
	DomainCognition:    "认知升级",
	DomainLife:         "生活智慧",
	DomainEmotion:      "情感",
}

func IsValidDomain(d Domain) bool {
	_, ok := ValidDomains[d]
	return ok
}
