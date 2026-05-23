package model

import "time"

// ============================================================
// 一级领域 v3 — 6 领域体系
// ============================================================
type Domain string

const (
	DomainVitality     Domain = "vitality"     // 生命
	DomainLiving       Domain = "living"       // 生活
	DomainWork         Domain = "work"         // 工作
	DomainRelationship Domain = "relationship" // 关系
	DomainCognition    Domain = "cognition"    // 认知
	DomainMeaning      Domain = "meaning"      // 意义
)

var ValidDomains = map[Domain]string{
	DomainVitality:     "生命",
	DomainLiving:       "生活",
	DomainWork:         "工作",
	DomainRelationship: "关系",
	DomainCognition:    "认知",
	DomainMeaning:      "意义",
}

func IsValidDomain(d Domain) bool {
	_, ok := ValidDomains[d]
	return ok
}

// ============================================================
// 二级领域 v3 — 34 子领域
// ============================================================
type SubDomain string

// vitality 生命
const (
	SubHealth   SubDomain = "health"
	SubHousing  SubDomain = "housing"
	SubTransit  SubDomain = "transit"
	SubDiet     SubDomain = "diet"
	SubExercise SubDomain = "exercise"
)

// living 生活
const (
	SubPets     SubDomain = "pets"
	SubTravel   SubDomain = "travel"
	SubFashion  SubDomain = "fashion"
	SubSelfcare SubDomain = "selfcare"
	SubShopping SubDomain = "shopping"
	SubFun      SubDomain = "fun"
)

// work 工作
const (
	SubJobhunt      SubDomain = "jobhunt"
	SubPromotion    SubDomain = "promotion"
	SubStartup      SubDomain = "startup"
	SubWorkComm     SubDomain = "work-comm"
	SubManagement   SubDomain = "management"
	SubProductivity SubDomain = "productivity"
)

// relationship 关系
const (
	SubMarriage   SubDomain = "marriage"
	SubRomance    SubDomain = "romance"
	SubFriendship SubDomain = "friendship"
	SubParenting  SubDomain = "parenting"
	SubParents    SubDomain = "parents"
	SubSiblings   SubDomain = "siblings"
)

// cognition 认知
const (
	SubCognitiveLearning SubDomain = "cog-learning"
	SubThinking          SubDomain = "thinking"
	SubInfo              SubDomain = "info"
	SubTools             SubDomain = "tools"
	SubCreativity        SubDomain = "creativity"
	SubExpression        SubDomain = "expression"
)

// meaning 意义
const (
	SubSelf      SubDomain = "self"
	SubHappiness SubDomain = "happiness"
	SubFaith     SubDomain = "faith"
	SubMission   SubDomain = "mission"
	SubBelonging SubDomain = "belonging"
)

var ValidSubDomains = map[SubDomain]string{
	// 生命
	SubHealth:   "健康",
	SubHousing:  "居住",
	SubTransit:  "出行",
	SubDiet:     "饮食",
	SubExercise: "运动",
	// 生活
	SubPets:     "宠物",
	SubTravel:   "旅行",
	SubFashion:  "衣着",
	SubSelfcare: "养护",
	SubShopping: "购物",
	SubFun:      "娱乐",
	// 工作
	SubJobhunt:      "求职",
	SubPromotion:    "升职",
	SubStartup:      "创业",
	SubWorkComm:     "沟通",
	SubManagement:   "管理",
	SubProductivity: "效率",
	// 关系
	SubMarriage:   "夫妻",
	SubRomance:    "恋人",
	SubFriendship: "朋友",
	SubParenting:  "亲子",
	SubParents:    "父母",
	SubSiblings:   "兄妹",
	// 认知
	SubCognitiveLearning: "学习",
	SubThinking:          "思维",
	SubInfo:              "信息",
	SubTools:             "工具",
	SubCreativity:        "创造",
	SubExpression:        "表达",
	// 意义
	SubSelf:      "自我",
	SubHappiness: "幸福",
	SubFaith:     "信仰",
	SubMission:   "使命",
	SubBelonging: "归属",
}

// SubDomainsByParent maps parent domain to its child sub-domains
var SubDomainsByParent = map[Domain][]SubDomain{
	DomainVitality:     {SubHealth, SubHousing, SubTransit, SubDiet, SubExercise},
	DomainLiving:       {SubPets, SubTravel, SubFashion, SubSelfcare, SubShopping, SubFun},
	DomainWork:         {SubJobhunt, SubPromotion, SubStartup, SubWorkComm, SubManagement, SubProductivity},
	DomainRelationship: {SubMarriage, SubRomance, SubFriendship, SubParenting, SubParents, SubSiblings},
	DomainCognition:    {SubCognitiveLearning, SubThinking, SubInfo, SubTools, SubCreativity, SubExpression},
	DomainMeaning:      {SubSelf, SubHappiness, SubFaith, SubMission, SubBelonging},
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
	ID                      string     `json:"id"`
	AuthorID                string     `json:"author_id"`
	Content                 string     `json:"content"`
	Interpretation          *string    `json:"interpretation,omitempty"`
	Domain                  Domain     `json:"domain"`
	SubDomain               *string    `json:"sub_domain,omitempty"`
	Topics                  string     `json:"topics"`
	IsOfficial              bool       `json:"is_official"`
	IsPrivate               bool       `json:"is_private"`
	SourceLabel             *string    `json:"source_label,omitempty"`
	LikeCount               int        `json:"like_count"`
	BookmarkCount           int        `json:"bookmark_count"`
	InterpretationGenerated bool       `json:"interpretation_generated"`
	Status                  string     `json:"status"`
	ReviewStatus            string     `json:"review_status"`
	ReviewReason            *string    `json:"review_reason,omitempty"`
	QualityScore            *float64   `json:"quality_score,omitempty"`
	ScoreDetails            *string    `json:"score_details,omitempty"`
	CreatorName             *string    `json:"creator_name,omitempty"`
	SourceType              string     `json:"source_type"`
	ScoreReason             *string    `json:"score_reason,omitempty"`
	OriginalText            *string    `json:"original_text,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
	DeletedAt               *time.Time `json:"deleted_at,omitempty"`
	RandomSort              float64    `json:"-"`
	// Joined fields
	AuthorName   *string `json:"author_name,omitempty"`
	AuthorAvatar *string `json:"author_avatar,omitempty"`
	AuthorTitle  *string `json:"author_title,omitempty"`
	IsLiked      bool    `json:"is_liked"`
	IsBookmarked bool    `json:"is_bookmarked"`
}

type CreateExperienceRequest struct {
	Content        string    `json:"content" binding:"required"`
	Domain         Domain    `json:"domain"`
	SubDomain      SubDomain `json:"sub_domain"`
	Interpretation string    `json:"interpretation"`
	Topics         string    `json:"topics"`
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
