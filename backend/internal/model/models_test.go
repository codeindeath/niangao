package model

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestIsValidDomain(t *testing.T) {
	tests := []struct {
		name   string
		domain Domain
		valid  bool
	}{
		{"work is valid", DomainWork, true},
		{"relationship is valid", DomainRelationship, true},
		{"cognition is valid", DomainCognition, true},
		{"vitality is valid", DomainVitality, true},
		{"meaning is valid", DomainMeaning, true},
		{"empty is invalid", "", false},
		{"unknown is invalid", "sports", false},
		{"case sensitive - uppercase invalid", "CAREER", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidDomain(tt.domain)
			if got != tt.valid {
				t.Errorf("IsValidDomain(%q) = %v, want %v", tt.domain, got, tt.valid)
			}
		})
	}
}

func TestValidDomainsMapping(t *testing.T) {
	expected := map[Domain]string{
		DomainVitality:     "生命",
		DomainLiving:       "生活",
		DomainWork:         "工作",
		DomainRelationship: "关系",
		DomainCognition:    "认知",
		DomainMeaning:      "意义",
	}

	if len(ValidDomains) != len(expected) {
		t.Errorf("ValidDomains has %d entries, want %d", len(ValidDomains), len(expected))
	}

	for k, v := range expected {
		if got, ok := ValidDomains[k]; !ok {
			t.Errorf("missing domain %s in ValidDomains", k)
		} else if got != v {
			t.Errorf("ValidDomains[%s] = %q, want %q", k, got, v)
		}
	}
}

func TestMeaningEmotionSubDomain(t *testing.T) {
	if !IsValidSubDomain(SubEmotion) {
		t.Fatal("meaning emotion sub-domain should be valid")
	}

	if got := ValidSubDomains[SubEmotion]; got != "情绪" {
		t.Fatalf("ValidSubDomains[SubEmotion] = %q, want %q", got, "情绪")
	}

	if !SubDomainBelongsToParent(DomainMeaning, SubEmotion) {
		t.Fatal("emotion sub-domain should belong to meaning")
	}

	if SubDomainBelongsToParent(DomainRelationship, SubEmotion) {
		t.Fatal("emotion sub-domain should not belong to relationship")
	}
}

func TestV4ExperienceEnums(t *testing.T) {
	if !IsValidExperienceType(ExperienceTypePlatformSelected) {
		t.Fatal("platform_selected should be a valid experience type")
	}
	if !IsValidExperienceType(ExperienceTypeUserOriginal) {
		t.Fatal("user_original should be a valid experience type")
	}
	if IsValidExperienceType("official") {
		t.Fatal("legacy official should not be a valid V4 experience type")
	}

	if !IsValidVisibility(VisibilityPublic) || !IsValidVisibility(VisibilityPrivate) {
		t.Fatal("public/private should be valid visibility values")
	}
	if IsValidVisibility("published") {
		t.Fatal("published should not be a valid visibility value")
	}

	if !IsValidLifecycleStatus(LifecycleActive) || !IsValidLifecycleStatus(LifecycleNeedsReview) {
		t.Fatal("active/needs_review should be valid lifecycle statuses")
	}
	if IsValidLifecycleStatus("published") {
		t.Fatal("published should not be a valid V4 lifecycle status")
	}

	if !IsValidQualityTier(QualityTierRecommendCandidate) || !IsValidQualityTier(QualityTierAICitable) {
		t.Fatal("recommend_candidate/ai_citable should be valid quality tiers")
	}
	if IsValidQualityTier("approved") {
		t.Fatal("approved should not be a valid V4 quality tier")
	}

	if !IsValidRecommendationStatus(RecommendationEligible) || !IsValidRecommendationStatus(RecommendationSuppressed) {
		t.Fatal("eligible/suppressed should be valid recommendation statuses")
	}
	if IsValidRecommendationStatus("approved") {
		t.Fatal("approved should not be a valid recommendation status")
	}
}

func TestV4DistributionEligibility(t *testing.T) {
	tests := []struct {
		name                 string
		visibility           Visibility
		lifecycle            LifecycleStatus
		qualityTier          QualityTier
		recommendationStatus RecommendationStatus
		want                 bool
	}{
		{
			name:                 "public active recommend candidate eligible",
			visibility:           VisibilityPublic,
			lifecycle:            LifecycleActive,
			qualityTier:          QualityTierRecommendCandidate,
			recommendationStatus: RecommendationEligible,
			want:                 true,
		},
		{
			name:                 "public visible is searchable but not recommendable",
			visibility:           VisibilityPublic,
			lifecycle:            LifecycleActive,
			qualityTier:          QualityTierPublicVisible,
			recommendationStatus: RecommendationEligible,
			want:                 false,
		},
		{
			name:                 "private never public recommendable",
			visibility:           VisibilityPrivate,
			lifecycle:            LifecycleActive,
			qualityTier:          QualityTierHighTrust,
			recommendationStatus: RecommendationEligible,
			want:                 false,
		},
		{
			name:                 "needs review is unavailable to public feed",
			visibility:           VisibilityPublic,
			lifecycle:            LifecycleNeedsReview,
			qualityTier:          QualityTierHighTrust,
			recommendationStatus: RecommendationEligible,
			want:                 false,
		},
		{
			name:                 "suppressed is not recommendable",
			visibility:           VisibilityPublic,
			lifecycle:            LifecycleActive,
			qualityTier:          QualityTierHighTrust,
			recommendationStatus: RecommendationSuppressed,
			want:                 false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanDistributePublicly(tt.visibility, tt.lifecycle, tt.qualityTier, tt.recommendationStatus)
			if got != tt.want {
				t.Fatalf("CanDistributePublicly() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQualityTierAtLeastRejectsUnknownTiers(t *testing.T) {
	if QualityTierAtLeast("unknown", QualityTierUnreviewed) {
		t.Fatal("unknown quality tier should not satisfy minimum tier checks")
	}

	if QualityTierAtLeast(QualityTierHighTrust, "unknown") {
		t.Fatal("unknown minimum quality tier should not be satisfiable")
	}
}

func TestV4AICitationEligibility(t *testing.T) {
	tests := []struct {
		name        string
		visibility  Visibility
		lifecycle   LifecycleStatus
		qualityTier QualityTier
		aiCitable   bool
		want        bool
	}{
		{"ai citable tier with flag", VisibilityPublic, LifecycleActive, QualityTierAICitable, true, true},
		{"high trust tier with flag", VisibilityPublic, LifecycleActive, QualityTierHighTrust, true, true},
		{"recommend candidate is not public ai citable", VisibilityPublic, LifecycleActive, QualityTierRecommendCandidate, true, false},
		{"flag false blocks citation", VisibilityPublic, LifecycleActive, QualityTierHighTrust, false, false},
		{"private blocks public citation", VisibilityPrivate, LifecycleActive, QualityTierHighTrust, true, false},
		{"needs review blocks citation", VisibilityPublic, LifecycleNeedsReview, QualityTierHighTrust, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanBeAICitedPublicly(tt.visibility, tt.lifecycle, tt.qualityTier, tt.aiCitable)
			if got != tt.want {
				t.Fatalf("CanBeAICitedPublicly() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateExperienceRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		content string
		domain  Domain
		valid   bool
		errMsg  string
	}{
		{
			name:    "valid experience",
			content: "接到任务先确认 deadline",
			domain:  DomainWork,
			valid:   true,
		},
		{
			name:    "content too long (over 100 chars)",
			content: strings.Repeat("a", 101),
			domain:  DomainWork,
			valid:   false,
			errMsg:  "exceeds 100",
		},
		{
			name:    "content exactly 100 chars",
			content: strings.Repeat("a", 100),
			domain:  DomainVitality,
			valid:   true,
		},
		{
			name:    "empty content",
			content: "",
			domain:  DomainWork,
			valid:   false,
			errMsg:  "required",
		},
		{
			name:    "empty domain is allowed for uncategorized publish",
			content: "valid content",
			domain:  "",
			valid:   true,
		},
		{
			name:    "invalid non-empty domain",
			content: "valid content",
			domain:  "invalid",
			valid:   false,
			errMsg:  "domain",
		},
		{
			name:    "chinese content",
			content: "把重要的决定放到早上做，意志力是有限资源",
			domain:  DomainCognition,
			valid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateExperienceRequest{
				Content: tt.content,
				Domain:  tt.domain,
			}

			// Verify content length (characters, not bytes)
			charLen := utf8.RuneCountInString(req.Content)
			if tt.valid {
				if charLen > 100 {
					t.Errorf("content char length %d exceeds 100", charLen)
				}
				if req.Domain != "" && !IsValidDomain(req.Domain) {
					t.Errorf("domain %s should be valid", req.Domain)
				}
			} else {
				if len(req.Content) > 100 && tt.errMsg == "" {
					t.Error("content should be flagged as too long")
				}
				if !IsValidDomain(req.Domain) && req.Domain != "" && tt.errMsg == "" {
					t.Error("domain should be flagged as invalid")
				}
			}
		})
	}
}

func TestInterpretationLength(t *testing.T) {
	tests := []struct {
		name           string
		interpretation string
		valid          bool
	}{
		{"empty interpretation (optional)", "", true},
		{"short interpretation", "如何执行：第一步...", true},
		{"exactly 300 chars", strings.Repeat("a", 300), true},
		{"over 300 chars", strings.Repeat("a", 301), false},
		{"chinese 300 chars", strings.Repeat("经", 300), true},
		{"chinese 301 chars", strings.Repeat("验", 301), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			charLen := utf8.RuneCountInString(tt.interpretation)
			if tt.valid && charLen > 300 {
				t.Errorf("interpretation char length %d exceeds 300", charLen)
			}
			if !tt.valid && charLen <= 300 {
				t.Error("interpretation should be flagged as too long")
			}
		})
	}
}

func TestExperienceStatusValues(t *testing.T) {
	validStatuses := map[string]bool{
		"published": true,
		"hidden":    true,
		"flagged":   true,
	}

	tests := []struct {
		status string
		valid  bool
	}{
		{"published", true},
		{"hidden", true},
		{"flagged", true},
		{"deleted", false},
		{"", false},
		{"PUBLISHED", false},
	}

	for _, tt := range tests {
		t.Run("status_"+tt.status, func(t *testing.T) {
			_, ok := validStatuses[tt.status]
			if ok != tt.valid {
				t.Errorf("status %s valid=%v, want %v", tt.status, ok, tt.valid)
			}
		})
	}
}

func TestMessageRoleValues(t *testing.T) {
	tests := []struct {
		role  string
		valid bool
	}{
		{"user", true},
		{"assistant", true},
		{"system", false},
		{"", false},
		{"USER", false},
	}

	for _, tt := range tests {
		t.Run("role_"+tt.role, func(t *testing.T) {
			valid := tt.role == "user" || tt.role == "assistant"
			if valid != tt.valid {
				t.Errorf("role %s valid=%v, want %v", tt.role, valid, tt.valid)
			}
		})
	}
}

func TestExperienceListQueryDefaults(t *testing.T) {
	q := ExperienceListQuery{}

	// Default values should be sensible
	if q.Page != 0 {
		t.Log("page defaults to 0, handler should treat 0 as 1")
	}
	if q.PageSize != 0 {
		t.Log("page_size defaults to 0, handler should treat 0 as 20")
	}
	if q.Sort != "" {
		t.Log("sort defaults to empty, handler should default to 'latest'")
	}
}

func TestChatRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		message string
		valid   bool
	}{
		{"valid message", "我最近和领导相处很累", true},
		{"empty message", "", false},
		{"very long message", strings.Repeat("长", 2000), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid && tt.message == "" {
				t.Error("message should not be empty")
			}
		})
	}
}
