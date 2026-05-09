package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/niangao/backend/internal/model"
)

const defaultAIServiceURL = "http://localhost:8000"

type ReviewRequest struct {
	Content   string `json:"content"`
	Domain    string `json:"domain"`
	SubDomain string `json:"sub_domain"`
}

type ReviewResult struct {
	Approved bool              `json:"approved"`
	Reason   string            `json:"reason"`
	Score    *QualityScoreData `json:"score,omitempty"`
}

type QualityScoreData struct {
	Overall    float64 `json:"overall"`
	Value      float64 `json:"value"`
	Actionable float64 `json:"actionable"`
	Universal  float64 `json:"universal"`
	Original   float64 `json:"original"`
	Clarity    float64 `json:"clarity"`
}

func callAIReview(req ReviewRequest) (*ReviewResult, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	aiURL := os.Getenv("AI_SERVICE_URL")
	if aiURL == "" {
		aiURL = defaultAIServiceURL
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(aiURL+"/api/v1/review", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("call AI review: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI review returned status %d", resp.StatusCode)
	}

	var result ReviewResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode review result: %w", err)
	}

	return &result, nil
}

func qualityScoreToDB(s *QualityScoreData) (score *float64, details *string) {
	if s == nil {
		return nil, nil
	}
	v := s.Overall
	b, err := json.Marshal(s)
	if err != nil {
		return &v, nil
	}
	str := string(b)
	return &v, &str
}

// placeholder types for compilation compatibility
var _ = model.ReviewApproved
var _ = model.ReviewRejected
var _ = model.ReviewPending

// ============================================================
// Classical Chinese translation
// ============================================================

type TranslateRequest struct {
	Content string `json:"content"`
}

type TranslateResult struct {
	IsClassical  bool   `json:"is_classical"`
	OriginalText string `json:"original_text"`
	ModernText   string `json:"modern_text"`
	DetectedLang string `json:"detected_lang"`
}

// callAITranslate detects language and returns translation if needed.
// Handles English→Chinese and Classical→Modern Chinese.
// Returns the result, or nil on error (caller should fall back to original content).
func callAITranslate(content string) *TranslateResult {
	body, err := json.Marshal(TranslateRequest{Content: content})
	if err != nil {
		return nil
	}

	aiURL := os.Getenv("AI_SERVICE_URL")
	if aiURL == "" {
		aiURL = defaultAIServiceURL
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(aiURL+"/api/v1/translate", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var result TranslateResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}

	return &result
}
