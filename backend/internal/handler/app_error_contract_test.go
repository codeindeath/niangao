package handler

import (
	"os"
	"regexp"
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
