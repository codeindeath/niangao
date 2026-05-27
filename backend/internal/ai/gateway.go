package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/niangao/backend/internal/middleware"
	"github.com/niangao/backend/internal/model"
)

type Gateway struct {
	baseURL    string
	httpClient *http.Client
}

func NewGateway(baseURL string) *Gateway {
	return NewGatewayWithTimeout(baseURL, 65*time.Second)
}

func NewGatewayWithTimeout(baseURL string, timeout time.Duration) *Gateway {
	if timeout <= 0 {
		timeout = 65 * time.Second
	}
	return &Gateway{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: timeout},
	}
}

type gatewayCallRequest struct {
	FunctionType  string `json:"function_type"`
	Payload       any    `json:"payload"`
	UserID        string `json:"user_id"`
	ChatTopicID   string `json:"chat_topic_id,omitempty"`
	ChatMessageID string `json:"chat_message_id,omitempty"`
}

type gatewayCallResponse struct {
	Result struct {
		ReplyText      string                       `json:"reply_text"`
		Citations      []model.ChatCitationDecision `json:"citations"`
		NoteSuggestion *model.ChatNoteSuggestion    `json:"note_suggestion"`
		EmotionLevel   string                       `json:"emotion_level"`
		RiskLevel      string                       `json:"risk_level"`
		ReplyMode      string                       `json:"reply_mode"`
	} `json:"result"`
	Confidence float64  `json:"confidence"`
	Warnings   []string `json:"warnings"`
}

type topicClassifyGatewayCallResponse struct {
	Result struct {
		ClarityScore             float64 `json:"clarity_score"`
		ShouldCreateTopic        bool    `json:"should_create_topic"`
		Title                    string  `json:"title"`
		Domain                   string  `json:"domain"`
		SubDomain                string  `json:"sub_domain"`
		TopicKeyword             string  `json:"topic_keyword"`
		CandidateExistingTopicID *string `json:"candidate_existing_topic_id"`
		ShouldBindExistingTopic  bool    `json:"should_bind_existing_topic"`
		DiscardIfUserLeaves      bool    `json:"discard_if_user_leaves"`
		Reason                   string  `json:"reason"`
	} `json:"result"`
	Confidence float64  `json:"confidence"`
	Warnings   []string `json:"warnings"`
}

func (g *Gateway) GenerateChatReply(ctx context.Context, req model.ChatGatewayRequest) (*model.ChatGatewayResponse, error) {
	if g == nil || g.baseURL == "" {
		return nil, fmt.Errorf("ai gateway base url is empty")
	}
	chatTopicID := ""
	if req.Topic != nil {
		chatTopicID = req.Topic.ID
	}
	body, err := json.Marshal(gatewayCallRequest{
		FunctionType:  "chat",
		Payload:       req,
		UserID:        req.UserID,
		ChatTopicID:   chatTopicID,
		ChatMessageID: req.UserMessageID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal ai gateway request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/api/v1/ai-gateway/call", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create ai gateway request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	setRequestIDHeader(ctx, httpReq)

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("call ai gateway: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read ai gateway response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ai gateway returned %d: %s", resp.StatusCode, string(respBytes))
	}

	var gatewayResp gatewayCallResponse
	if err := json.Unmarshal(respBytes, &gatewayResp); err != nil {
		return nil, fmt.Errorf("parse ai gateway response: %w", err)
	}
	reply := strings.TrimSpace(gatewayResp.Result.ReplyText)
	if reply == "" {
		return nil, fmt.Errorf("ai gateway returned empty reply")
	}
	return &model.ChatGatewayResponse{
		ReplyText:      reply,
		Citations:      gatewayResp.Result.Citations,
		NoteSuggestion: gatewayResp.Result.NoteSuggestion,
		EmotionLevel:   gatewayResp.Result.EmotionLevel,
		RiskLevel:      gatewayResp.Result.RiskLevel,
		ReplyMode:      gatewayResp.Result.ReplyMode,
		Confidence:     gatewayResp.Confidence,
		Warnings:       gatewayResp.Warnings,
	}, nil
}

func (g *Gateway) ClassifyChatTopic(ctx context.Context, req model.ChatTopicClassificationRequest) (*model.ChatTopicClassificationResponse, error) {
	if g == nil || g.baseURL == "" {
		return nil, fmt.Errorf("ai gateway base url is empty")
	}
	body, err := json.Marshal(gatewayCallRequest{
		FunctionType: "chat_topic_classify",
		Payload:      req,
		UserID:       req.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal topic classify gateway request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/api/v1/ai-gateway/call", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create topic classify gateway request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	setRequestIDHeader(ctx, httpReq)

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("call topic classify gateway: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read topic classify gateway response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("topic classify gateway returned %d: %s", resp.StatusCode, string(respBytes))
	}

	var gatewayResp topicClassifyGatewayCallResponse
	if err := json.Unmarshal(respBytes, &gatewayResp); err != nil {
		return nil, fmt.Errorf("parse topic classify gateway response: %w", err)
	}
	return &model.ChatTopicClassificationResponse{
		ClarityScore:             gatewayResp.Result.ClarityScore,
		ShouldCreateTopic:        gatewayResp.Result.ShouldCreateTopic,
		Title:                    strings.TrimSpace(gatewayResp.Result.Title),
		Domain:                   gatewayResp.Result.Domain,
		SubDomain:                gatewayResp.Result.SubDomain,
		TopicKeyword:             gatewayResp.Result.TopicKeyword,
		CandidateExistingTopicID: topicClassifyStringValue(gatewayResp.Result.CandidateExistingTopicID),
		ShouldBindExistingTopic:  gatewayResp.Result.ShouldBindExistingTopic,
		DiscardIfUserLeaves:      gatewayResp.Result.DiscardIfUserLeaves,
		Reason:                   gatewayResp.Result.Reason,
		Confidence:               gatewayResp.Confidence,
		Warnings:                 gatewayResp.Warnings,
	}, nil
}

func topicClassifyStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

type rewriteGatewayCallResponse struct {
	Result struct {
		CanRewrite         bool   `json:"can_rewrite"`
		Content            string `json:"content"`
		Domain             string `json:"domain"`
		SubDomain          string `json:"sub_domain"`
		Topic              string `json:"topic"`
		RewriteLevel       string `json:"rewrite_level"`
		SourcePreservation string `json:"source_preservation"`
		NeedsUserEdit      bool   `json:"needs_user_edit"`
		Reason             string `json:"reason"`
	} `json:"result"`
	Confidence float64  `json:"confidence"`
	Warnings   []string `json:"warnings"`
}

func (g *Gateway) RewriteExperience(ctx context.Context, req model.ExperienceRewriteGatewayRequest) (*model.ExperienceRewriteGatewayResponse, error) {
	if g == nil || g.baseURL == "" {
		return nil, fmt.Errorf("ai gateway base url is empty")
	}
	body, err := json.Marshal(gatewayCallRequest{
		FunctionType: "experience_rewrite",
		Payload:      req,
		UserID:       req.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal rewrite gateway request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/api/v1/ai-gateway/call", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create rewrite gateway request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	setRequestIDHeader(ctx, httpReq)

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("call rewrite gateway: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read rewrite gateway response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rewrite gateway returned %d: %s", resp.StatusCode, string(respBytes))
	}

	var gatewayResp rewriteGatewayCallResponse
	if err := json.Unmarshal(respBytes, &gatewayResp); err != nil {
		return nil, fmt.Errorf("parse rewrite gateway response: %w", err)
	}
	return &model.ExperienceRewriteGatewayResponse{
		CanRewrite:         gatewayResp.Result.CanRewrite,
		RewrittenContent:   strings.TrimSpace(gatewayResp.Result.Content),
		Domain:             gatewayResp.Result.Domain,
		SubDomain:          gatewayResp.Result.SubDomain,
		Topic:              gatewayResp.Result.Topic,
		RewriteLevel:       gatewayResp.Result.RewriteLevel,
		SourcePreservation: gatewayResp.Result.SourcePreservation,
		NeedsUserEdit:      gatewayResp.Result.NeedsUserEdit,
		Reason:             gatewayResp.Result.Reason,
		Confidence:         gatewayResp.Confidence,
		Warnings:           gatewayResp.Warnings,
	}, nil
}

func setRequestIDHeader(ctx context.Context, req *http.Request) {
	if requestID := middleware.RequestIDFromContext(ctx); requestID != "" {
		req.Header.Set(middleware.RequestIDHeader, requestID)
	}
}
