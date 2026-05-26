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
// 二级领域 v3 — 35 子领域
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
	SubEmotion   SubDomain = "emotion"
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
	SubEmotion:   "情绪",
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
	DomainMeaning:      {SubSelf, SubHappiness, SubEmotion, SubFaith, SubMission, SubBelonging},
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
// V4 experience facts
// ============================================================
type ExperienceType string

const (
	ExperienceTypePlatformSelected ExperienceType = "platform_selected"
	ExperienceTypeUserOriginal     ExperienceType = "user_original"
)

var ValidExperienceTypes = map[ExperienceType]struct{}{
	ExperienceTypePlatformSelected: {},
	ExperienceTypeUserOriginal:     {},
}

func IsValidExperienceType(v ExperienceType) bool {
	_, ok := ValidExperienceTypes[v]
	return ok
}

type Visibility string

const (
	VisibilityPublic  Visibility = "public"
	VisibilityPrivate Visibility = "private"
)

var ValidVisibilities = map[Visibility]struct{}{
	VisibilityPublic:  {},
	VisibilityPrivate: {},
}

func IsValidVisibility(v Visibility) bool {
	_, ok := ValidVisibilities[v]
	return ok
}

type LifecycleStatus string

const (
	LifecycleActive      LifecycleStatus = "active"
	LifecycleHidden      LifecycleStatus = "hidden"
	LifecycleDeleted     LifecycleStatus = "deleted"
	LifecycleNeedsReview LifecycleStatus = "needs_review"
)

var ValidLifecycleStatuses = map[LifecycleStatus]struct{}{
	LifecycleActive:      {},
	LifecycleHidden:      {},
	LifecycleDeleted:     {},
	LifecycleNeedsReview: {},
}

func IsValidLifecycleStatus(v LifecycleStatus) bool {
	_, ok := ValidLifecycleStatuses[v]
	return ok
}

type QualityTier string

const (
	QualityTierUnreviewed         QualityTier = "unreviewed"
	QualityTierPrivateOnly        QualityTier = "private_only"
	QualityTierPublicVisible      QualityTier = "public_visible"
	QualityTierRecommendCandidate QualityTier = "recommend_candidate"
	QualityTierAICitable          QualityTier = "ai_citable"
	QualityTierHighTrust          QualityTier = "high_trust"
)

var ValidQualityTiers = map[QualityTier]struct{}{
	QualityTierUnreviewed:         {},
	QualityTierPrivateOnly:        {},
	QualityTierPublicVisible:      {},
	QualityTierRecommendCandidate: {},
	QualityTierAICitable:          {},
	QualityTierHighTrust:          {},
}

var QualityTierRank = map[QualityTier]int{
	QualityTierUnreviewed:         0,
	QualityTierPrivateOnly:        1,
	QualityTierPublicVisible:      2,
	QualityTierRecommendCandidate: 3,
	QualityTierAICitable:          4,
	QualityTierHighTrust:          5,
}

func IsValidQualityTier(v QualityTier) bool {
	_, ok := ValidQualityTiers[v]
	return ok
}

func QualityTierAtLeast(v QualityTier, min QualityTier) bool {
	rank, ok := QualityTierRank[v]
	if !ok {
		return false
	}

	minRank, ok := QualityTierRank[min]
	if !ok {
		return false
	}

	return rank >= minRank
}

type RecommendationStatus string

const (
	RecommendationEligible   RecommendationStatus = "eligible"
	RecommendationIneligible RecommendationStatus = "ineligible"
	RecommendationSuppressed RecommendationStatus = "suppressed"
)

var ValidRecommendationStatuses = map[RecommendationStatus]struct{}{
	RecommendationEligible:   {},
	RecommendationIneligible: {},
	RecommendationSuppressed: {},
}

func IsValidRecommendationStatus(v RecommendationStatus) bool {
	_, ok := ValidRecommendationStatuses[v]
	return ok
}

type InterpretationStatus string

const (
	InterpretationNone    InterpretationStatus = "none"
	InterpretationPending InterpretationStatus = "pending"
	InterpretationReady   InterpretationStatus = "ready"
	InterpretationStale   InterpretationStatus = "stale"
	InterpretationFailed  InterpretationStatus = "failed"
)

type SourceScene string

const (
	SourceSceneNote SourceScene = "note"
	SourceSceneChat SourceScene = "chat"
)

func CanDistributePublicly(visibility Visibility, lifecycle LifecycleStatus, qualityTier QualityTier, recommendationStatus RecommendationStatus) bool {
	return visibility == VisibilityPublic &&
		lifecycle == LifecycleActive &&
		recommendationStatus == RecommendationEligible &&
		QualityTierAtLeast(qualityTier, QualityTierRecommendCandidate)
}

func CanBeAICitedPublicly(visibility Visibility, lifecycle LifecycleStatus, qualityTier QualityTier, aiCitable bool) bool {
	return visibility == VisibilityPublic &&
		lifecycle == LifecycleActive &&
		aiCitable &&
		QualityTierAtLeast(qualityTier, QualityTierAICitable)
}

// ============================================================
// Structs
// ============================================================

type User struct {
	ID              string  `json:"id"`
	AppleUserID     *string `json:"-"`
	Nickname        string  `json:"nickname"`
	DisplayName     *string `json:"display_name,omitempty"`
	AvatarURL       *string `json:"avatar_url,omitempty"`
	Bio             *string `json:"bio,omitempty"`
	Title           *string `json:"title,omitempty"`
	ExperienceCount int     `json:"experience_count"`
	BookmarkCount   int     `json:"bookmark_count"`
	PracticedCount  int     `json:"practiced_count"`
}

type Experience struct {
	ID                        string     `json:"id"`
	AuthorID                  string     `json:"author_id"`
	OwnerUserID               *string    `json:"owner_user_id,omitempty"`
	Content                   string     `json:"content"`
	Interpretation            *string    `json:"interpretation,omitempty"`
	ExperienceType            string     `json:"experience_type,omitempty"`
	Visibility                string     `json:"visibility,omitempty"`
	LifecycleStatus           string     `json:"lifecycle_status,omitempty"`
	Domain                    Domain     `json:"domain"`
	SubDomain                 *string    `json:"sub_domain,omitempty"`
	Topics                    string     `json:"topics"`
	Topic                     string     `json:"topic,omitempty"`
	IsOfficial                bool       `json:"is_official"`
	IsPrivate                 bool       `json:"is_private"`
	SourceLabel               *string    `json:"source_label,omitempty"`
	LikeCount                 int        `json:"like_count"`
	BookmarkCount             int        `json:"bookmark_count"`
	InspirationCount          int        `json:"inspiration_count,omitempty"`
	CollectionCount           int        `json:"collection_count,omitempty"`
	InterpretationGenerated   bool       `json:"interpretation_generated"`
	InterpretationStatus      string     `json:"interpretation_status,omitempty"`
	Status                    string     `json:"status"`
	ReviewStatus              string     `json:"review_status"`
	ReviewReason              *string    `json:"review_reason,omitempty"`
	QualityScore              *float64   `json:"quality_score,omitempty"`
	QualityTier               string     `json:"quality_tier,omitempty"`
	ScoreDetails              *string    `json:"score_details,omitempty"`
	CreatorName               *string    `json:"creator_name,omitempty"`
	CreatorDisplayName        *string    `json:"creator_display_name,omitempty"`
	SourceType                string     `json:"source_type"`
	SourceScene               string     `json:"source_scene,omitempty"`
	SourceChatTopicID         *string    `json:"source_chat_topic_id,omitempty"`
	SourceChatMessageID       *string    `json:"source_chat_message_id,omitempty"`
	SourceChatMessageSnapshot *string    `json:"source_chat_message_snapshot,omitempty"`
	SourceReliability         string     `json:"source_reliability,omitempty"`
	RecommendationStatus      string     `json:"recommendation_status,omitempty"`
	AICitable                 bool       `json:"ai_citable,omitempty"`
	StarRating                int        `json:"star_rating,omitempty"`
	ScoreReason               *string    `json:"score_reason,omitempty"`
	OriginalText              *string    `json:"original_text,omitempty"`
	CreatedAt                 time.Time  `json:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at"`
	DeletedAt                 *time.Time `json:"deleted_at,omitempty"`
	RandomSort                float64    `json:"-"`
	// Joined fields
	AuthorName   *string `json:"author_name,omitempty"`
	AuthorAvatar *string `json:"author_avatar,omitempty"`
	AuthorTitle  *string `json:"author_title,omitempty"`
	IsLiked      bool    `json:"is_liked"`
	IsBookmarked bool    `json:"is_bookmarked"`
}

type ExperienceCard struct {
	ID                             string `json:"id"`
	OwnerUserID                    string `json:"owner_user_id,omitempty"`
	Content                        string `json:"content,omitempty"`
	ExperienceType                 string `json:"experience_type,omitempty"`
	Visibility                     string `json:"visibility,omitempty"`
	LifecycleStatus                string `json:"lifecycle_status,omitempty"`
	Domain                         string `json:"domain,omitempty"`
	SubDomain                      string `json:"sub_domain,omitempty"`
	Topic                          string `json:"topic,omitempty"`
	CreatorDisplayName             string `json:"creator_display_name,omitempty"`
	InterpretationStatus           string `json:"interpretation_status,omitempty"`
	InterpretationSummaryAvailable bool   `json:"interpretation_summary_available"`
	QualityTier                    string `json:"quality_tier,omitempty"`
	StarRating                     int    `json:"star_rating,omitempty"`
	InspirationCount               int    `json:"inspiration_count,omitempty"`
	CollectionCount                int    `json:"collection_count,omitempty"`
	IsCollected                    bool   `json:"is_collected"`
	IsInspired                     bool   `json:"is_inspired"`
	UnavailableReason              string `json:"unavailable_reason,omitempty"`
}

type ExperienceEventRequest struct {
	EventType     string         `json:"event_type"`
	SourceContext string         `json:"source_context,omitempty"`
	ContextID     string         `json:"context_id,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

type FeedPage struct {
	Data       []ExperienceCard `json:"data"`
	NextCursor string           `json:"next_cursor"`
	SessionID  string           `json:"session_id"`
	HasMore    bool             `json:"has_more"`
}

type AssetStats struct {
	MyExperiences      int `json:"my_experiences"`
	Collections        int `json:"collections"`
	MonthAdded         int `json:"month_added"`
	PublicExperiences  int `json:"public_experiences"`
	PrivateExperiences int `json:"private_experiences"`
	FromNote           int `json:"from_note"`
	FromChat           int `json:"from_chat"`
}

type ContributionStats struct {
	InspiredUsers      int `json:"inspired_users"`
	CollectedCount     int `json:"collected_count"`
	MonthInspiredUsers int `json:"month_inspired_users"`
	MonthCollected     int `json:"month_collected"`
}

type ChangeStats struct {
	ChatTopics           int `json:"chat_topics"`
	ClearerCount         int `json:"clearer_count"`
	MonthChatExperiences int `json:"month_chat_experiences"`
}

type RecentHarvestStats struct {
	Range           string `json:"range"`
	NoteAdded       int    `json:"note_added"`
	ChatExperiences int    `json:"chat_experiences"`
	InspiredUsers   int    `json:"inspired_users"`
	CollectedCount  int    `json:"collected_count"`
}

type RespondedExperienceCard struct {
	ID               string    `json:"id"`
	Content          string    `json:"content"`
	Domain           string    `json:"domain"`
	SubDomain        string    `json:"sub_domain,omitempty"`
	StarRating       int       `json:"star_rating"`
	InspirationCount int       `json:"inspiration_count"`
	CollectionCount  int       `json:"collection_count"`
	LastRespondedAt  time.Time `json:"last_responded_at"`
}

type MeProfile struct {
	DisplayName        string   `json:"display_name"`
	CareerStage        string   `json:"career_stage,omitempty"`
	RelationshipStatus string   `json:"relationship_status,omitempty"`
	IsParent           *bool    `json:"is_parent,omitempty"`
	CommonIssues       []string `json:"common_issues"`
	FreeDescription    string   `json:"free_description,omitempty"`
	ProfileVersion     int      `json:"profile_version"`
}

type MeProfilePatch struct {
	DisplayName        *string   `json:"display_name"`
	CareerStage        *string   `json:"career_stage"`
	RelationshipStatus *string   `json:"relationship_status"`
	IsParent           *bool     `json:"is_parent"`
	CommonIssues       *[]string `json:"common_issues"`
	FreeDescription    *string   `json:"free_description"`
}

type ChatTopic struct {
	ID           string     `json:"id"`
	Status       string     `json:"status"`
	Title        string     `json:"title"`
	Domain       string     `json:"domain,omitempty"`
	SubDomain    string     `json:"sub_domain,omitempty"`
	Topic        string     `json:"topic,omitempty"`
	ClarityScore *float64   `json:"clarity_score,omitempty"`
	Summary      string     `json:"summary,omitempty"`
	LastOpenedAt *time.Time `json:"last_opened_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type ChatTopicPage struct {
	Data       []ChatTopic `json:"data"`
	NextCursor string      `json:"next_cursor"`
	HasMore    bool        `json:"has_more"`
}

type ChatTempSession struct {
	ID              string     `json:"id"`
	Status          string     `json:"status"`
	ForcedNewTopic  bool       `json:"forced_new_topic"`
	PromotedTopicID *string    `json:"promoted_topic_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DiscardedAt     *time.Time `json:"discarded_at,omitempty"`
	PurgeAfter      *time.Time `json:"purge_after,omitempty"`
}

type CreateChatTopicRequest struct {
	Title     string `json:"title"`
	Domain    string `json:"domain"`
	SubDomain string `json:"sub_domain"`
	Topic     string `json:"topic"`
}

type UpdateChatTopicRequest struct {
	Title     *string `json:"title"`
	Domain    *string `json:"domain"`
	SubDomain *string `json:"sub_domain"`
	Topic     *string `json:"topic"`
}

type ChatScopeKind string

const (
	ChatScopeTopic       ChatScopeKind = "topic"
	ChatScopeTempSession ChatScopeKind = "temp_session"
)

type ChatMessageScope struct {
	Kind ChatScopeKind `json:"kind"`
	ID   string        `json:"id"`
}

type ChatScopeContext struct {
	Scope        ChatMessageScope `json:"scope"`
	SessionState string           `json:"session_state"`
	Topic        *ChatTopic       `json:"topic,omitempty"`
	TempSession  *ChatTempSession `json:"temp_session,omitempty"`
}

type ChatMessage struct {
	ID                      string              `json:"id"`
	UserID                  string              `json:"user_id,omitempty"`
	TopicID                 *string             `json:"topic_id,omitempty"`
	TempSessionID           *string             `json:"temp_session_id,omitempty"`
	Role                    string              `json:"role"`
	Content                 string              `json:"content"`
	Status                  string              `json:"status"`
	RiskLevel               string              `json:"risk_level"`
	ClientMessageID         *string             `json:"client_message_id,omitempty"`
	ReferencedExperienceIDs []string            `json:"referenced_experience_ids"`
	ReferenceCards          []ChatReferenceCard `json:"reference_cards,omitempty"`
	CreatedAt               time.Time           `json:"created_at"`
}

type ChatMessagePage struct {
	Data       []ChatMessage `json:"data"`
	NextCursor string        `json:"next_cursor"`
	HasMore    bool          `json:"has_more"`
}

type SaveChatMessageRequest struct {
	Scope                   ChatMessageScope `json:"scope"`
	Role                    string           `json:"role"`
	Content                 string           `json:"content"`
	Status                  string           `json:"status"`
	RiskLevel               string           `json:"risk_level"`
	ClientMessageID         string           `json:"client_message_id,omitempty"`
	ReferencedExperienceIDs []string         `json:"referenced_experience_ids,omitempty"`
	Metadata                map[string]any   `json:"metadata,omitempty"`
}

type SendChatMessageRequest struct {
	Content         string `json:"content"`
	ClientMessageID string `json:"client_message_id"`
}

type ChatPreClassification struct {
	EmotionLevel        string   `json:"emotion_level"`
	UserIntent          string   `json:"user_intent"`
	RiskLevel           string   `json:"risk_level"`
	RiskReasons         []string `json:"risk_reasons"`
	ShouldAvoidCitation bool     `json:"should_avoid_citation"`
}

type ChatCandidateExperience struct {
	ExperienceID         string `json:"experience_id"`
	Content              string `json:"content"`
	CreatorName          string `json:"creator_name"`
	SourceRelation       string `json:"source_relation"`
	Visibility           string `json:"visibility"`
	QualityTier          string `json:"quality_tier"`
	SourceReliability    string `json:"source_reliability,omitempty"`
	SourceDerivationType string `json:"source_derivation_type,omitempty"`
	CitationPolicy       string `json:"citation_policy"`
	RelevanceReason      string `json:"relevance_reason"`
	IsCollected          bool   `json:"is_collected"`
}

type ChatCitationDecision struct {
	ExperienceID     string `json:"experience_id"`
	UsageType        string `json:"usage_type"`
	ShowCard         bool   `json:"show_card"`
	CitationSentence string `json:"citation_sentence"`
	ReasonCode       string `json:"reason_code"`
	Strength         string `json:"strength"`
}

type ChatReferenceCard struct {
	ExperienceID      string `json:"experience_id"`
	Content           string `json:"content"`
	IsCollected       bool   `json:"is_collected"`
	CitationType      string `json:"citation_type"`
	CitationSentence  string `json:"citation_sentence,omitempty"`
	ReasonCode        string `json:"reason_code,omitempty"`
	UnavailableReason string `json:"unavailable_reason,omitempty"`
}

type ChatNoteSuggestion struct {
	ShouldShow       bool     `json:"should_show"`
	SuggestedText    *string  `json:"suggested_text,omitempty"`
	SourceMessageIDs []string `json:"source_message_ids"`
}

type ChatGatewayRequest struct {
	UserID               string                    `json:"user_id"`
	UserMessageID        string                    `json:"message_id"`
	UserMessage          string                    `json:"user_message"`
	SessionState         string                    `json:"session_state"`
	Scope                ChatMessageScope          `json:"scope"`
	Topic                *ChatTopic                `json:"topic,omitempty"`
	RecentMessages       []ChatMessage             `json:"recent_messages"`
	PreClassification    ChatPreClassification     `json:"pre_classification"`
	CandidateExperiences []ChatCandidateExperience `json:"candidate_experiences"`
	ContextFlags         []string                  `json:"context_flags"`
	Limits               map[string]int            `json:"limits"`
}

type ChatGatewayResponse struct {
	ReplyText      string                 `json:"reply_text"`
	Citations      []ChatCitationDecision `json:"citations"`
	NoteSuggestion *ChatNoteSuggestion    `json:"note_suggestion,omitempty"`
	EmotionLevel   string                 `json:"emotion_level"`
	RiskLevel      string                 `json:"risk_level"`
	ReplyMode      string                 `json:"reply_mode,omitempty"`
	Confidence     float64                `json:"confidence,omitempty"`
	Warnings       []string               `json:"warnings,omitempty"`
}

type ChatTopicClassificationRequest struct {
	UserID              string              `json:"user_id"`
	TempSessionID       string              `json:"temp_session_id"`
	Messages            []ChatMessage       `json:"messages"`
	RecentTopics        []ChatTopic         `json:"recent_topics"`
	UserClickedNewTopic bool                `json:"user_clicked_new_topic"`
	DomainTaxonomy      map[string][]string `json:"domain_taxonomy"`
}

type ChatTopicClassificationResponse struct {
	ClarityScore             float64  `json:"clarity_score"`
	ShouldCreateTopic        bool     `json:"should_create_topic"`
	Title                    string   `json:"title"`
	Domain                   string   `json:"domain,omitempty"`
	SubDomain                string   `json:"sub_domain,omitempty"`
	TopicKeyword             string   `json:"topic_keyword,omitempty"`
	CandidateExistingTopicID string   `json:"candidate_existing_topic_id,omitempty"`
	ShouldBindExistingTopic  bool     `json:"should_bind_existing_topic"`
	DiscardIfUserLeaves      bool     `json:"discard_if_user_leaves"`
	Reason                   string   `json:"reason,omitempty"`
	Confidence               float64  `json:"confidence,omitempty"`
	Warnings                 []string `json:"warnings,omitempty"`
}

type PromoteChatTempSessionRequest struct {
	Title                string  `json:"title"`
	Domain               string  `json:"domain,omitempty"`
	SubDomain            string  `json:"sub_domain,omitempty"`
	Topic                string  `json:"topic,omitempty"`
	ClarityScore         float64 `json:"clarity_score"`
	ClassificationReason string  `json:"classification_reason,omitempty"`
}

type SendChatMessageResponse struct {
	UserMessage    ChatMessage         `json:"user_message"`
	Message        ChatMessage         `json:"message"`
	ReferenceCards []ChatReferenceCard `json:"reference_cards"`
	NoteSuggestion ChatNoteSuggestion  `json:"note_suggestion"`
	SessionState   string              `json:"session_state"`
	PromotedTopic  *ChatTopic          `json:"promoted_topic,omitempty"`
	Retryable      bool                `json:"retryable,omitempty"`
}

type ExperienceRewriteRequest struct {
	Content               string     `json:"content"`
	Source                string     `json:"source"`
	SourceMessageIDs      []string   `json:"source_message_ids"`
	DefaultVisibility     Visibility `json:"default_visibility"`
	UserSelectedDomain    Domain     `json:"user_selected_domain"`
	UserSelectedSubDomain SubDomain  `json:"user_selected_sub_domain"`
	TopicContext          string     `json:"topic_context"`
}

type ExperienceRewriteGatewayRequest struct {
	UserID                string     `json:"user_id"`
	Source                string     `json:"source"`
	RawText               string     `json:"raw_text"`
	SourceMessageIDs      []string   `json:"source_message_ids"`
	DefaultVisibility     Visibility `json:"default_visibility"`
	UserSelectedDomain    Domain     `json:"user_selected_domain,omitempty"`
	UserSelectedSubDomain SubDomain  `json:"user_selected_sub_domain,omitempty"`
	TopicContext          string     `json:"topic_context,omitempty"`
}

type ExperienceRewriteGatewayResponse struct {
	CanRewrite         bool     `json:"can_rewrite"`
	RewrittenContent   string   `json:"rewritten_content"`
	Domain             string   `json:"domain,omitempty"`
	SubDomain          string   `json:"sub_domain,omitempty"`
	Topic              string   `json:"topic,omitempty"`
	RewriteLevel       string   `json:"rewrite_level,omitempty"`
	SourcePreservation string   `json:"source_preservation,omitempty"`
	NeedsUserEdit      bool     `json:"needs_user_edit"`
	Reason             string   `json:"reason,omitempty"`
	Confidence         float64  `json:"confidence,omitempty"`
	Warnings           []string `json:"warnings"`
}

type CreateExperienceRequest struct {
	Content                   string     `json:"content" binding:"required"`
	Visibility                Visibility `json:"visibility"`
	Domain                    Domain     `json:"domain"`
	SubDomain                 SubDomain  `json:"sub_domain"`
	Interpretation            string     `json:"interpretation"`
	Topics                    string     `json:"topics"`
	Topic                     string     `json:"topic"`
	SourceScene               string     `json:"source_scene"`
	SourceChatTopicID         string     `json:"source_chat_topic_id"`
	SourceChatMessageID       string     `json:"source_chat_message_id"`
	SourceChatMessageSnapshot string     `json:"source_chat_message_snapshot"`
	SourceMessageIDs          []string   `json:"source_message_ids"`
	IsPrivate                 bool       `json:"is_private"`
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
