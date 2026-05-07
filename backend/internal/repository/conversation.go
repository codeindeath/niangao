package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niangao/backend/internal/model"
)

type ConversationRepo struct {
	db *pgxpool.Pool
}

func NewConversationRepo(db *pgxpool.Pool) *ConversationRepo {
	return &ConversationRepo{db: db}
}

func (r *ConversationRepo) Create(ctx context.Context, userID string) (*model.Conversation, error) {
	c := &model.Conversation{
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := r.db.QueryRow(ctx,
		`INSERT INTO conversations (user_id, created_at, updated_at)
		 VALUES ($1,$2,$3) RETURNING id`,
		c.UserID, c.CreatedAt, c.UpdatedAt,
	).Scan(&c.ID)
	if err != nil {
		return nil, fmt.Errorf("create conversation: %w", err)
	}

	return c, nil
}

func (r *ConversationRepo) AddMessage(ctx context.Context, convID, role, content string, refExpIDs []string) (*model.Message, error) {
	m := &model.Message{
		ConversationID:          convID,
		Role:                    role,
		Content:                 content,
		ReferencedExperienceIDs: refExpIDs,
		CreatedAt:               time.Now(),
	}

	err := r.db.QueryRow(ctx,
		`INSERT INTO messages (conversation_id, role, content, referenced_experience_ids, created_at)
		 VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		m.ConversationID, m.Role, m.Content, m.ReferencedExperienceIDs, m.CreatedAt,
	).Scan(&m.ID)
	if err != nil {
		return nil, fmt.Errorf("add message: %w", err)
	}

	// Update conversation timestamp
	_, _ = r.db.Exec(ctx,
		`UPDATE conversations SET updated_at=$1 WHERE id=$2`,
		time.Now(), convID,
	)

	return m, nil
}

func (r *ConversationRepo) GetMessages(ctx context.Context, convID string, limit int) ([]model.Message, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, conversation_id, role, content, referenced_experience_ids, created_at
		 FROM messages WHERE conversation_id=$1
		 ORDER BY created_at ASC LIMIT $2`,
		convID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var m model.Message
		err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content,
			&m.ReferencedExperienceIDs, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}

	return messages, nil
}

func (r *ConversationRepo) GetByID(ctx context.Context, id string) (*model.Conversation, error) {
	c := &model.Conversation{}
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, title, created_at, updated_at
		 FROM conversations WHERE id=$1`,
		id,
	).Scan(&c.ID, &c.UserID, &c.Title, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *ConversationRepo) ListByUser(ctx context.Context, userID string) ([]model.Conversation, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, title, created_at, updated_at
		 FROM conversations WHERE user_id=$1
		 ORDER BY updated_at DESC LIMIT 50`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var convs []model.Conversation
	for rows.Next() {
		var c model.Conversation
		err := rows.Scan(&c.ID, &c.UserID, &c.Title, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		convs = append(convs, c)
	}

	return convs, nil
}

// GetOrCreateByUser returns the user's most recent conversation, creating one if none exists.
func (r *ConversationRepo) GetOrCreateByUser(ctx context.Context, userID string) (*model.Conversation, error) {
	// Try to find existing
	c := &model.Conversation{}
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, title, created_at, updated_at
		 FROM conversations WHERE user_id=$1
		 ORDER BY updated_at DESC LIMIT 1`,
		userID,
	).Scan(&c.ID, &c.UserID, &c.Title, &c.CreatedAt, &c.UpdatedAt)

	if err == nil {
		return c, nil
	}
	if err != pgx.ErrNoRows {
		return nil, fmt.Errorf("get conversation: %w", err)
	}

	// Create new
	return r.Create(ctx, userID)
}

// GetMessagesSince returns messages created after the given time.
func (r *ConversationRepo) GetMessagesSince(ctx context.Context, convID string, since time.Duration) ([]model.Message, error) {
	cutoff := time.Now().Add(-since)
	rows, err := r.db.Query(ctx,
		`SELECT id, conversation_id, role, content, referenced_experience_ids, created_at
		 FROM messages WHERE conversation_id=$1 AND created_at >= $2
		 ORDER BY created_at ASC`,
		convID, cutoff,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var m model.Message
		err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content,
			&m.ReferencedExperienceIDs, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	if messages == nil {
		messages = make([]model.Message, 0)
	}
	return messages, nil
}

// CountTodayMessages counts user+assistant message pairs today for rate limiting.
func (r *ConversationRepo) CountTodayMessages(ctx context.Context, convID string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM messages
		 WHERE conversation_id=$1
		   AND role='user'
		   AND created_at::date = CURRENT_DATE`,
		convID,
	).Scan(&count)
	return count, err
}
