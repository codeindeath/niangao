package model

import "time"

// ============================================================
// 一级领域
// ============================================================
type Domain string

const (
	DomainCareer       Domain = "career"
	DomainRelationship Domain = "relationship"
	DomainCognition    Domain = "cognition"
	DomainLife         Domain = "life"
	DomainEmotion      Domain = "emotion"
)

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

// ============================================================
// 二级领域
// ============================================================
type SubDomain string

// career 职场成长
const (
	SubCareerPlanning   SubDomain = "career-planning"
	SubSkillBuilding    SubDomain = "skill-building"
	SubSideHustle       SubDomain = "side-hustle"
	SubWorkplaceComm    SubDomain = "workplace-comm"
)

// relationship 人际关系
const (
	SubIntimate      SubDomain = "intimate"
	SubFamily        SubDomain = "family"
	SubSocialSkill   SubDomain = "social-skill"
	SubCommunication SubDomain = "communication"
)

// cognition 认知升级
const (
	SubMentalModel SubDomain = "mental-model"
	SubLearning    SubDomain = "learning"
	SubDecision    SubDomain = "decision"
	SubPsychology  SubDomain = "psychology"
)

// life 生活智慧
const (
	SubFinance    SubDomain = "finance"
	SubHealth     SubDomain = "health"
	SubTimeMgmt   SubDomain = "time-mgmt"
	SubHabits     SubDomain = "habits"
	SubDigitalLife SubDomain = "digital-life"
)

// emotion 情绪情感
const (
	SubRegulation  SubDomain = "regulation"
	SubSelfGrowth  SubDomain = "self-growth"
	SubHappiness   SubDomain = "happiness"
	SubStressMgmt  SubDomain = "stress-mgmt"
)

var ValidSubDomains = map[SubDomain]string{
	SubCareerPlanning: "职业规划",
	SubSkillBuilding:  "技能提升",
	SubSideHustle:     "副业创业",
	SubWorkplaceComm:  "职场沟通",

	SubIntimate:      "亲密关系",
	SubFamily:        "家庭关系",
	SubSocialSkill:   "社交技巧",
	SubCommunication: "沟通表达",

	SubMentalModel: "思维模型",
	SubLearning:    "学习方法",
	SubDecision:    "决策判断",
	SubPsychology:  "心理认知",

	SubFinance:    "理财规划",
	SubHealth:     "健康养生",
	SubTimeMgmt:   "时间管理",
	SubHabits:     "习惯养成",
	SubDigitalLife: "数字生活",

	SubRegulation: "情绪调节",
	SubSelfGrowth: "自我成长",
	SubHappiness:  "幸福感",
	SubStressMgmt: "压力管理",
}

// SubDomainsByParent maps parent domain to its child sub-domains
var SubDomainsByParent = map[Domain][]SubDomain{
	DomainCareer:       {SubCareerPlanning, SubSkillBuilding, SubSideHustle, SubWorkplaceComm},
	DomainRelationship: {SubIntimate, SubFamily, SubSocialSkill, SubCommunication},
	DomainCognition:    {SubMentalModel, SubLearning, SubDecision, SubPsychology},
	DomainLife:         {SubFinance, SubHealth, SubTimeMgmt, SubHabits, SubDigitalLife},
	DomainEmotion:      {SubRegulation, SubSelfGrowth, SubHappiness, SubStressMgmt},
}

func IsValidSubDomain(d SubDomain) bool {
	_, ok := ValidSubDomains[d]
	return ok
}

func SubDomainBelongsToParent(parent Domain, child SubDomain) bool {
	children, ok := SubDomainsByParent[parent]
	if !ok {
		return false
	}
	for _, c := range children {
		if c == child {
			return true
		}
	}
	return false
}

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
	SubDomain               *string   `json:"sub_domain,omitempty"`
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
	OriginalText            *string   `json:"original_text,omitempty"`
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
