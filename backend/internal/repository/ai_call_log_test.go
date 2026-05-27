package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/niangao/backend/internal/ai"
)

type fakeAICallLogDB struct {
	query string
	args  []any
}

func (f *fakeAICallLogDB) Exec(_ context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	f.query = query
	f.args = args
	return pgconn.NewCommandTag("INSERT 0 1"), nil
}

func TestAICallLogRepoUsesFunctionConfigForRequiredLogFields(t *testing.T) {
	db := &fakeAICallLogDB{}
	repo := NewAICallLogRepo(db)
	started := time.Date(2026, 5, 27, 22, 0, 0, 0, time.UTC)
	finished := started.Add(180 * time.Millisecond)

	if err := repo.RecordAICall(context.Background(), ai.CallLogEntry{
		FunctionType:           "chat",
		CallSource:             "app_chat",
		UserID:                 "11111111-1111-4111-8111-111111111111",
		ChatTopicID:            "33333333-3333-4333-8333-333333333333",
		ChatMessageID:          "22222222-2222-4222-8222-222222222222",
		Status:                 "success",
		LatencyMS:              180,
		SanitizedInputSummary:  "messages=1",
		SanitizedOutputSummary: "reply_chars=12",
		StartedAt:              started,
		FinishedAt:             finished,
	}); err != nil {
		t.Fatalf("RecordAICall returned error: %v", err)
	}

	if !strings.Contains(db.query, "INSERT INTO ai_call_logs") {
		t.Fatalf("query does not insert ai_call_logs: %s", db.query)
	}
	for _, fragment := range []string{
		"FROM ai_function_configs",
		"cfg.key_alias",
		"cfg.model",
		"cfg.prompt_version",
		"cfg.schema_version",
		"cfg.queue_name",
	} {
		if !strings.Contains(db.query, fragment) {
			t.Fatalf("query missing %q: %s", fragment, db.query)
		}
	}
	if len(db.args) != 13 {
		t.Fatalf("arg count = %d, want 13: %#v", len(db.args), db.args)
	}
	if db.args[0] != "chat" || db.args[1] != "app_chat" {
		t.Fatalf("function/source args = %#v", db.args[:2])
	}
	if db.args[7] != "success" || db.args[8] != "" {
		t.Fatalf("status/error args = %#v", db.args[7:9])
	}
}
