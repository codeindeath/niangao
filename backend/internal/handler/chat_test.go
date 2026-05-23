package handler

import (
	"testing"
	"time"

	"github.com/niangao/backend/internal/model"
)

func TestPickWelcomeGreeting(t *testing.T) {
	// Run many times to ensure no panic and reasonable distribution
	seen := make(map[string]int)
	for i := 0; i < 500; i++ {
		g := pickWelcomeGreeting()
		if g == "" {
			t.Fatal("empty greeting")
		}
		seen[g]++
	}
	// Should have at least 3 unique greetings
	if len(seen) < 3 {
		t.Errorf("expected at least 3 unique greetings from pool, got %d", len(seen))
	}
}

func TestPickGreetingByGap(t *testing.T) {
	tests := []struct {
		name string
		gap  time.Duration
	}{
		{"2-12h", 3 * time.Hour},
		{"12-24h", 18 * time.Hour},
		{"1-3d", 48 * time.Hour},
		{"3-7d", 120 * time.Hour},
		{"7-30d", 14 * 24 * time.Hour},
		{"over 30d (falls back to last)", 60 * 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seen := make(map[string]int)
			for i := 0; i < 50; i++ {
				g := pickGreetingByGap(tt.gap)
				if g == "" {
					t.Fatal("empty greeting")
				}
				seen[g]++
			}
			if len(seen) < 2 {
				t.Errorf("gap=%v: expected at least 2 unique greetings, got %d", tt.gap, len(seen))
			}
		})
	}
}

func TestPickGreeting(t *testing.T) {
	// 3 hours ago → should be in 2-12h category
	g := pickGreeting(time.Now().Add(-3 * time.Hour))
	if g == "" {
		t.Fatal("empty greeting")
	}

	// 2 days ago → should be in 1-3d category
	g = pickGreeting(time.Now().Add(-48 * time.Hour))
	if g == "" {
		t.Fatal("empty greeting")
	}
}

func TestGreetingPoolSize(t *testing.T) {
	total := 0
	for _, cat := range greetingCategories {
		total += len(cat.messages)
	}
	if total != 50 {
		t.Errorf("expected 50 greeting templates total, got %d", total)
	}
}

func TestInferChatDomainFromCurrentMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{"work", "最近和老板沟通项目压力很大", string(model.DomainWork)},
		{"relationship", "我和女朋友吵架之后一直很难受", string(model.DomainRelationship)},
		{"meaning", "最近很迷茫，不知道人生方向在哪里", string(model.DomainMeaning)},
		{"meaning emotion", "最近情绪起伏很大，不知道怎么和自己相处", string(model.DomainMeaning)},
		{"none", "今天只是想随便聊聊", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inferChatDomain(tt.message, nil); got != tt.want {
				t.Fatalf("inferChatDomain() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInferChatDomainUsesRecentHistory(t *testing.T) {
	history := []model.Message{
		{Role: "user", Content: "最近面试和求职都不顺"},
		{Role: "assistant", Content: "听起来求职过程消耗了你很多力气"},
	}

	if got := inferChatDomain("嗯，就是这样", history); got != string(model.DomainWork) {
		t.Fatalf("inferChatDomain() = %q, want %q", got, model.DomainWork)
	}
}
