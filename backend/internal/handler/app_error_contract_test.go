package handler

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestAppFacingV4HandlersUseStructuredErrorResponses(t *testing.T) {
	files := []string{
		"auth.go",
		"chat_v4.go",
		"experience.go",
		"experience_actions_v4.go",
		"feed_v4.go",
		"me_account_v4.go",
		"me_feedback_v4.go",
		"me_profile_v4.go",
		"me_stats_v4.go",
		"search_v4.go",
	}
	forbidden := regexp.MustCompile(`gin\.H\{\s*"error"\s*:\s*("[^"]*"|err\.Error\(\))`)

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			src, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("read %s: %v", file, err)
			}
			if match := forbidden.Find(src); match != nil {
				t.Fatalf("%s returns an unstructured App-facing error near %q", file, string(match))
			}
		})
	}
}

func TestAppFacingAuthErrorsAvoidTechnicalTokenCopy(t *testing.T) {
	sourceBytes, err := os.ReadFile("auth.go")
	if err != nil {
		t.Fatalf("read auth.go: %v", err)
	}
	source := string(sourceBytes)
	respondErrorCall := regexp.MustCompile(`respondError\([^,\n]+,\s*http\.Status[A-Za-z]+,\s*"[^"]+",\s*"([^"]+)"\)`)
	technicalTerms := []string{
		"identity_token",
		"refresh_token",
		"refresh token",
		"token",
	}

	for _, match := range respondErrorCall.FindAllStringSubmatch(source, -1) {
		message := match[1]
		lowered := strings.ToLower(message)
		for _, term := range technicalTerms {
			if strings.Contains(lowered, term) {
				t.Fatalf("auth error message %q exposes technical term %q", message, term)
			}
		}
	}
}

func TestRefreshTokenMissingUserUsesExpiredAuthCopy(t *testing.T) {
	sourceBytes, err := os.ReadFile("auth.go")
	if err != nil {
		t.Fatalf("read auth.go: %v", err)
	}
	source := string(sourceBytes)

	if strings.Contains(source, `"user_not_found", "用户不存在"`) {
		t.Fatal("refresh-token user lookup failures should not expose 用户不存在 to the App")
	}
	if strings.Contains(source, `http.StatusInternalServerError, "user_not_found"`) {
		t.Fatal("refresh-token user lookup failures should behave like expired auth, not a 500")
	}
}
