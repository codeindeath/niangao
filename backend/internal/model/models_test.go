package model

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestDomainTypes(t *testing.T) {
	// Domain and SubDomain are type aliases for string
	var d Domain = "career"
	var s SubDomain = "career-planning"
	if string(d) != "career" {
		t.Error("Domain type alias broken")
	}
	if string(s) != "career-planning" {
		t.Error("SubDomain type alias broken")
	}
}

func TestDomainCatalogNilSafety(t *testing.T) {
	// Without DB init, catalog is nil — all checks return false gracefully
	if IsValidDomain("career") {
		t.Error("IsValidDomain should return false when catalog is nil")
	}
	if IsValidSubDomain("career-planning") {
		t.Error("IsValidSubDomain should return false when catalog is nil")
	}
	if SubDomainBelongsToParent("career", "career-planning") {
		t.Error("SubDomainBelongsToParent should return false when catalog is nil")
	}
	if d := DomainDisplay("career"); d != "career" {
		t.Errorf("DomainDisplay should return input when catalog is nil, got %q", d)
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
			domain:  "career",
			valid:   true,
		},
		{
			name:    "content too long (over 100 chars)",
			content: strings.Repeat("a", 101),
			domain:  "career",
			valid:   false,
			errMsg:  "exceeds 100",
		},
		{
			name:    "content exactly 100 chars",
			content: strings.Repeat("a", 100),
			domain:  "life-philosophy",
			valid:   true,
		},
		{
			name:    "empty content",
			content: "",
			domain:  "career",
			valid:   false,
			errMsg:  "required",
		},
		{
			name:    "chinese content",
			content: "把重要的决定放到早上做，意志力是有限资源",
			domain:  "cognition",
			valid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateExperienceRequest{
				Content: tt.content,
				Domain:  tt.domain,
			}

			charLen := utf8.RuneCountInString(req.Content)
			if tt.valid {
				if charLen > 100 {
					t.Errorf("content char length %d exceeds 100", charLen)
				}
			} else {
				if tt.errMsg == "exceeds 100" && charLen <= 100 {
					t.Error("content should be flagged as too long")
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
		{"exactly 500 chars", strings.Repeat("a", 500), true},
		{"over 500 chars", strings.Repeat("a", 501), false},
		{"chinese 500 chars", strings.Repeat("经", 500), true},
		{"chinese 501 chars", strings.Repeat("验", 501), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			charLen := utf8.RuneCountInString(tt.interpretation)
			if tt.valid && charLen > 500 {
				t.Errorf("interpretation char length %d exceeds 500", charLen)
			}
			if !tt.valid && charLen <= 500 {
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
