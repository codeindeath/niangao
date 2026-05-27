package ai

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/niangao/backend/internal/model"
)

type CallLogEntry struct {
	FunctionType           string
	CallSource             string
	UserID                 string
	ExperienceID           string
	ChatTopicID            string
	ChatMessageID          string
	Status                 string
	ErrorCode              string
	LatencyMS              int
	SanitizedInputSummary  string
	SanitizedOutputSummary string
	StartedAt              time.Time
	FinishedAt             time.Time
}

type CallLogger interface {
	RecordAICall(ctx context.Context, entry CallLogEntry) error
}

func (g *Gateway) recordAICall(ctx context.Context, entry CallLogEntry) {
	if g == nil || g.logger == nil || entry.FunctionType == "" {
		return
	}
	if entry.StartedAt.IsZero() {
		entry.StartedAt = time.Now().UTC()
	}
	if entry.FinishedAt.IsZero() {
		entry.FinishedAt = time.Now().UTC()
	}
	if entry.LatencyMS < 0 {
		entry.LatencyMS = 0
	}
	logCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	defer cancel()
	if err := g.logger.RecordAICall(logCtx, entry); err != nil {
		log.Printf("record ai call log failed function=%s status=%s: %v", entry.FunctionType, entry.Status, err)
	}
}

func aiCallFailureStatus(err error) (string, string) {
	if err == nil {
		return "success", ""
	}
	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "context deadline exceeded") {
		return "timeout", "timeout"
	}
	if strings.Contains(err.Error(), "Client.Timeout") {
		return "timeout", "timeout"
	}
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "timeout", "timeout"
	}
	return "failed", "request_failed"
}

func chatInputSummary(req model.ChatGatewayRequest) string {
	return fmt.Sprintf(
		"user_message_chars=%d recent_messages=%d candidate_experiences=%d context_flags=%d",
		len([]rune(req.UserMessage)),
		len(req.RecentMessages),
		len(req.CandidateExperiences),
		len(req.ContextFlags),
	)
}

func chatOutputSummary(resp *model.ChatGatewayResponse) string {
	if resp == nil {
		return ""
	}
	citations := len(resp.Citations)
	if resp.NoteSuggestion != nil && resp.NoteSuggestion.ShouldShow {
		return fmt.Sprintf("reply_chars=%d citations=%d note_suggestion=true", len([]rune(resp.ReplyText)), citations)
	}
	return fmt.Sprintf("reply_chars=%d citations=%d note_suggestion=false", len([]rune(resp.ReplyText)), citations)
}

func topicClassifyInputSummary(req model.ChatTopicClassificationRequest) string {
	return fmt.Sprintf(
		"messages=%d recent_topics=%d user_clicked_new_topic=%t taxonomy_domains=%d",
		len(req.Messages),
		len(req.RecentTopics),
		req.UserClickedNewTopic,
		len(req.DomainTaxonomy),
	)
}

func topicClassifyOutputSummary(resp *model.ChatTopicClassificationResponse) string {
	if resp == nil {
		return ""
	}
	return fmt.Sprintf("should_create_topic=%t clarity_score=%.2f has_domain=%t", resp.ShouldCreateTopic, resp.ClarityScore, resp.Domain != "")
}

func rewriteInputSummary(req model.ExperienceRewriteGatewayRequest) string {
	return fmt.Sprintf(
		"raw_text_chars=%d source=%s source_message_ids=%d has_user_domain=%t",
		len([]rune(req.RawText)),
		req.Source,
		len(req.SourceMessageIDs),
		req.UserSelectedDomain != "",
	)
}

func rewriteOutputSummary(resp *model.ExperienceRewriteGatewayResponse) string {
	if resp == nil {
		return ""
	}
	return fmt.Sprintf(
		"can_rewrite=%t rewritten_chars=%d has_domain=%t needs_user_edit=%t",
		resp.CanRewrite,
		len([]rune(resp.RewrittenContent)),
		resp.Domain != "",
		resp.NeedsUserEdit,
	)
}
