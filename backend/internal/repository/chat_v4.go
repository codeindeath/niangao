package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/niangao/backend/internal/model"
)

func (r *ConversationRepo) RecentChatTopics(ctx context.Context, userID string, limit int) ([]model.ChatTopic, error) {
	if limit < 1 || limit > 20 {
		limit = 10
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, status, COALESCE(title, ''), COALESCE(domain, ''), COALESCE(sub_domain, ''),
		       COALESCE(topic, ''), clarity_score, COALESCE(summary, ''), last_opened_at, created_at, updated_at
		FROM chat_topics
		WHERE user_id=$1::uuid AND status='active' AND deleted_at IS NULL
		ORDER BY updated_at DESC, id DESC
		LIMIT $2`,
		userID, limit)
	if err != nil {
		return nil, fmt.Errorf("recent chat topics: %w", err)
	}
	defer rows.Close()

	topics, err := scanChatTopics(rows, limit)
	if err != nil {
		return nil, err
	}
	if topics == nil {
		topics = []model.ChatTopic{}
	}
	return topics, nil
}

func (r *ConversationRepo) ChatTopics(ctx context.Context, userID string, limit int, cursor string) (*model.ChatTopicPage, error) {
	if limit < 1 || limit > 50 {
		limit = 20
	}
	offset := parseOffsetCursor(cursor)
	rows, err := r.db.Query(ctx, `
		SELECT id, status, COALESCE(title, ''), COALESCE(domain, ''), COALESCE(sub_domain, ''),
		       COALESCE(topic, ''), clarity_score, COALESCE(summary, ''), last_opened_at, created_at, updated_at
		FROM chat_topics
		WHERE user_id=$1::uuid AND status='active' AND deleted_at IS NULL
		ORDER BY updated_at DESC, id DESC
		LIMIT $2 OFFSET $3`,
		userID, limit+1, offset)
	if err != nil {
		return nil, fmt.Errorf("chat topics: %w", err)
	}
	defer rows.Close()

	topics, err := scanChatTopics(rows, limit+1)
	if err != nil {
		return nil, err
	}
	hasMore := len(topics) > limit
	if hasMore {
		topics = topics[:limit]
	}
	nextCursor := ""
	if hasMore {
		nextCursor = strconv.Itoa(offset + limit)
	}
	return &model.ChatTopicPage{Data: topics, NextCursor: nextCursor, HasMore: hasMore}, nil
}

func (r *ConversationRepo) CreateTempSession(ctx context.Context, userID string, forcedNewTopic bool) (*model.ChatTempSession, error) {
	session := &model.ChatTempSession{ForcedNewTopic: forcedNewTopic}
	err := r.db.QueryRow(ctx, `
		INSERT INTO chat_temp_sessions (user_id, forced_new_topic, created_at, updated_at)
		VALUES ($1::uuid, $2, NOW(), NOW())
		RETURNING id, status, forced_new_topic, promoted_topic_id, created_at, updated_at, discarded_at, purge_after`,
		userID, forcedNewTopic,
	).Scan(
		&session.ID,
		&session.Status,
		&session.ForcedNewTopic,
		&session.PromotedTopicID,
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.DiscardedAt,
		&session.PurgeAfter,
	)
	if err != nil {
		return nil, fmt.Errorf("create temp session: %w", err)
	}
	return session, nil
}

func (r *ConversationRepo) CreateChatTopic(ctx context.Context, userID string, req model.CreateChatTopicRequest) (*model.ChatTopic, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = "新的心事"
	}

	topic := &model.ChatTopic{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO chat_topics (user_id, title, domain, sub_domain, topic, last_opened_at, created_at, updated_at)
		VALUES ($1::uuid, $2, NULLIF($3, ''), NULLIF($4, ''), $5, NOW(), NOW(), NOW())
		RETURNING id, status, COALESCE(title, ''), COALESCE(domain, ''), COALESCE(sub_domain, ''),
		          COALESCE(topic, ''), clarity_score, COALESCE(summary, ''), last_opened_at, created_at, updated_at`,
		userID, title, req.Domain, req.SubDomain, req.Topic,
	).Scan(
		&topic.ID,
		&topic.Status,
		&topic.Title,
		&topic.Domain,
		&topic.SubDomain,
		&topic.Topic,
		&topic.ClarityScore,
		&topic.Summary,
		&topic.LastOpenedAt,
		&topic.CreatedAt,
		&topic.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create chat topic: %w", err)
	}
	return topic, nil
}

func (r *ConversationRepo) UpdateChatTopic(ctx context.Context, userID string, topicID string, req model.UpdateChatTopicRequest) (*model.ChatTopic, error) {
	topic := &model.ChatTopic{}
	err := r.db.QueryRow(ctx, `
		UPDATE chat_topics
		SET title=COALESCE(NULLIF($3, ''), title),
		    domain=COALESCE(NULLIF($4, ''), domain),
		    sub_domain=COALESCE(NULLIF($5, ''), sub_domain),
		    topic=COALESCE($6, topic),
		    updated_at=NOW()
		WHERE id=$1::uuid AND user_id=$2::uuid AND status='active' AND deleted_at IS NULL
		RETURNING id, status, COALESCE(title, ''), COALESCE(domain, ''), COALESCE(sub_domain, ''),
		          COALESCE(topic, ''), clarity_score, COALESCE(summary, ''), last_opened_at, created_at, updated_at`,
		topicID, userID, stringPtrValue(req.Title), stringPtrValue(req.Domain), stringPtrValue(req.SubDomain), stringPtrValue(req.Topic),
	).Scan(
		&topic.ID,
		&topic.Status,
		&topic.Title,
		&topic.Domain,
		&topic.SubDomain,
		&topic.Topic,
		&topic.ClarityScore,
		&topic.Summary,
		&topic.LastOpenedAt,
		&topic.CreatedAt,
		&topic.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrExperienceUnavailable
		}
		return nil, fmt.Errorf("update chat topic: %w", err)
	}
	return topic, nil
}

func (r *ConversationRepo) DeleteChatTopic(ctx context.Context, userID string, topicID string) error {
	result, err := r.db.Exec(ctx, `
		UPDATE chat_topics
		SET status='deleted', deleted_at=NOW(), updated_at=NOW()
		WHERE id=$1::uuid AND user_id=$2::uuid AND status='active' AND deleted_at IS NULL`,
		topicID, userID)
	if err != nil {
		return fmt.Errorf("delete chat topic: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrExperienceUnavailable
	}
	return nil
}

func scanChatTopics(rows pgx.Rows, capacity int) ([]model.ChatTopic, error) {
	topics := make([]model.ChatTopic, 0, capacity)
	for rows.Next() {
		var topic model.ChatTopic
		if err := rows.Scan(
			&topic.ID,
			&topic.Status,
			&topic.Title,
			&topic.Domain,
			&topic.SubDomain,
			&topic.Topic,
			&topic.ClarityScore,
			&topic.Summary,
			&topic.LastOpenedAt,
			&topic.CreatedAt,
			&topic.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan chat topic: %w", err)
		}
		topics = append(topics, topic)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chat topics: %w", err)
	}
	return topics, nil
}

func stringPtrValue(v *string) *string {
	if v == nil {
		return nil
	}
	s := strings.TrimSpace(*v)
	return &s
}

func (r *ConversationRepo) ChatMessages(ctx context.Context, userID string, scope model.ChatMessageScope, limit int, cursor string) (*model.ChatMessagePage, error) {
	if limit < 1 || limit > 50 {
		limit = 30
	}
	if _, err := r.VerifyChatScope(ctx, userID, scope); err != nil {
		return nil, err
	}
	offset := parseOffsetCursor(cursor)
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, topic_id, temp_session_id, role, content, status, risk_level,
		       client_message_id, referenced_experience_ids, created_at
		FROM chat_messages
		WHERE user_id=$1::uuid
		  AND (
		    ($2='topic' AND topic_id=$3::uuid) OR
		    ($2='temp_session' AND temp_session_id=$3::uuid)
		  )
		  AND status <> 'deleted'
		ORDER BY created_at ASC, id ASC
		LIMIT $4 OFFSET $5`,
		userID, string(scope.Kind), scope.ID, limit+1, offset)
	if err != nil {
		return nil, fmt.Errorf("chat messages: %w", err)
	}
	defer rows.Close()

	messages, err := scanChatMessages(rows, limit+1)
	if err != nil {
		return nil, err
	}
	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}
	if err := r.attachChatReferenceCards(ctx, userID, messages); err != nil {
		return nil, err
	}
	nextCursor := ""
	if hasMore {
		nextCursor = strconv.Itoa(offset + limit)
	}
	return &model.ChatMessagePage{Data: messages, NextCursor: nextCursor, HasMore: hasMore}, nil
}

func (r *ConversationRepo) attachChatReferenceCards(ctx context.Context, userID string, messages []model.ChatMessage) error {
	messageIndex := make(map[string]int)
	messageIDs := make([]string, 0, len(messages))
	for i := range messages {
		if messages[i].Role != "assistant" || messages[i].ID == "" {
			continue
		}
		messageIndex[messages[i].ID] = i
		messageIDs = append(messageIDs, messages[i].ID)
	}
	if len(messageIDs) == 0 {
		return nil
	}

	rows, err := r.db.Query(ctx, `
		WITH cited AS (
		  SELECT
		    cc.id AS citation_id,
		    cc.message_id,
		    cc.experience_id,
		    cc.citation_type,
		    cc.shown_at,
		    e.content,
		    e.deleted_at,
		    e.visibility AS visibility,
		    e.lifecycle_status AS lifecycle_status,
		    COALESCE(e.owner_user_id, e.author_id) AS owner_user_id,
		    (
		      e.deleted_at IS NULL
		      AND (
		        (
		          e.visibility = 'public'
		          AND e.lifecycle_status = 'active'
		        )
		        OR (
		          COALESCE(e.owner_user_id, e.author_id) = $1::uuid
		          AND e.lifecycle_status <> 'deleted'
		        )
		      )
		    ) AS visible_to_viewer
		  FROM chat_citations cc
		  JOIN experiences e ON e.id = cc.experience_id
		  WHERE cc.message_id = ANY($2::uuid[])
		)
		SELECT
		  c.message_id,
		  c.experience_id,
		  CASE WHEN c.visible_to_viewer THEN c.content ELSE '' END AS content,
		  EXISTS(
		    SELECT 1 FROM experience_collections ec
		    WHERE ec.user_id = $1::uuid
		      AND ec.experience_id = c.experience_id
		      AND ec.status = 'active'
		  ) AS is_collected,
		  c.citation_type,
		  CASE WHEN c.visible_to_viewer THEN '' ELSE 'experience_unavailable' END AS unavailable_reason
		FROM cited c
		ORDER BY c.message_id, c.shown_at NULLS LAST, c.citation_id`,
		userID, messageIDs)
	if err != nil {
		return fmt.Errorf("chat reference cards: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var messageID string
		var card model.ChatReferenceCard
		if err := rows.Scan(
			&messageID,
			&card.ExperienceID,
			&card.Content,
			&card.IsCollected,
			&card.CitationType,
			&card.UnavailableReason,
		); err != nil {
			return fmt.Errorf("scan chat reference card: %w", err)
		}
		if idx, ok := messageIndex[messageID]; ok {
			messages[idx].ReferenceCards = append(messages[idx].ReferenceCards, card)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate chat reference cards: %w", err)
	}
	return nil
}

func (r *ConversationRepo) VerifyChatScope(ctx context.Context, userID string, scope model.ChatMessageScope) (*model.ChatScopeContext, error) {
	switch scope.Kind {
	case model.ChatScopeTopic:
		topic := &model.ChatTopic{}
		err := r.db.QueryRow(ctx, `
			SELECT id, status, COALESCE(title, ''), COALESCE(domain, ''), COALESCE(sub_domain, ''),
			       COALESCE(topic, ''), clarity_score, COALESCE(summary, ''), last_opened_at, created_at, updated_at
			FROM chat_topics
			WHERE id=$1::uuid AND user_id=$2::uuid AND status='active' AND deleted_at IS NULL`,
			scope.ID, userID,
		).Scan(
			&topic.ID,
			&topic.Status,
			&topic.Title,
			&topic.Domain,
			&topic.SubDomain,
			&topic.Topic,
			&topic.ClarityScore,
			&topic.Summary,
			&topic.LastOpenedAt,
			&topic.CreatedAt,
			&topic.UpdatedAt,
		)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, ErrExperienceUnavailable
			}
			return nil, fmt.Errorf("verify chat topic: %w", err)
		}
		return &model.ChatScopeContext{Scope: scope, SessionState: "stable_topic", Topic: topic}, nil
	case model.ChatScopeTempSession:
		session := &model.ChatTempSession{}
		err := r.db.QueryRow(ctx, `
			SELECT id, status, forced_new_topic, promoted_topic_id, created_at, updated_at, discarded_at, purge_after
			FROM chat_temp_sessions
			WHERE id=$1::uuid AND user_id=$2::uuid AND status='active'`,
			scope.ID, userID,
		).Scan(
			&session.ID,
			&session.Status,
			&session.ForcedNewTopic,
			&session.PromotedTopicID,
			&session.CreatedAt,
			&session.UpdatedAt,
			&session.DiscardedAt,
			&session.PurgeAfter,
		)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, ErrExperienceUnavailable
			}
			return nil, fmt.Errorf("verify temp session: %w", err)
		}
		return &model.ChatScopeContext{Scope: scope, SessionState: "temp_session", TempSession: session}, nil
	default:
		return nil, ErrExperienceUnavailable
	}
}

func (r *ConversationRepo) AddChatMessage(ctx context.Context, userID string, req model.SaveChatMessageRequest) (*model.ChatMessage, error) {
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "sent"
	}
	riskLevel := strings.TrimSpace(req.RiskLevel)
	if riskLevel == "" {
		riskLevel = "normal"
	}
	metadata := req.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("marshal chat metadata: %w", err)
	}
	referencedExperienceIDs := req.ReferencedExperienceIDs
	if referencedExperienceIDs == nil {
		referencedExperienceIDs = []string{}
	}

	var topicID *string
	var tempSessionID *string
	if req.Scope.Kind == model.ChatScopeTopic {
		topicID = &req.Scope.ID
	} else if req.Scope.Kind == model.ChatScopeTempSession {
		tempSessionID = &req.Scope.ID
	} else {
		return nil, ErrExperienceUnavailable
	}

	message := &model.ChatMessage{}
	err = r.db.QueryRow(ctx, `
		INSERT INTO chat_messages (
		  user_id, topic_id, temp_session_id, role, content, status, risk_level,
		  client_message_id, referenced_experience_ids, metadata, created_at
		)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6, $7, NULLIF($8, ''), $9::uuid[], $10::jsonb, NOW())
		ON CONFLICT (user_id, client_message_id) WHERE client_message_id IS NOT NULL DO UPDATE
		  SET metadata = chat_messages.metadata
		RETURNING id, user_id, topic_id, temp_session_id, role, content, status, risk_level,
		          client_message_id, referenced_experience_ids, created_at`,
		userID,
		topicID,
		tempSessionID,
		req.Role,
		req.Content,
		status,
		riskLevel,
		strings.TrimSpace(req.ClientMessageID),
		referencedExperienceIDs,
		string(metadataBytes),
	).Scan(
		&message.ID,
		&message.UserID,
		&message.TopicID,
		&message.TempSessionID,
		&message.Role,
		&message.Content,
		&message.Status,
		&message.RiskLevel,
		&message.ClientMessageID,
		&message.ReferencedExperienceIDs,
		&message.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("add chat message: %w", err)
	}

	if req.Scope.Kind == model.ChatScopeTopic {
		_, _ = r.db.Exec(ctx, `UPDATE chat_topics SET last_opened_at=NOW(), updated_at=NOW() WHERE id=$1::uuid`, req.Scope.ID)
	} else {
		_, _ = r.db.Exec(ctx, `UPDATE chat_temp_sessions SET updated_at=NOW() WHERE id=$1::uuid`, req.Scope.ID)
	}
	return message, nil
}

func (r *ConversationRepo) RecentChatMessages(ctx context.Context, userID string, scope model.ChatMessageScope, limit int) ([]model.ChatMessage, error) {
	if limit < 1 || limit > 30 {
		limit = 12
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, topic_id, temp_session_id, role, content, status, risk_level,
		       client_message_id, referenced_experience_ids, created_at
		FROM chat_messages
		WHERE user_id=$1::uuid
		  AND (
		    ($2='topic' AND topic_id=$3::uuid) OR
		    ($2='temp_session' AND temp_session_id=$3::uuid)
		  )
		  AND status <> 'deleted'
		ORDER BY created_at DESC, id DESC
		LIMIT $4`,
		userID, string(scope.Kind), scope.ID, limit)
	if err != nil {
		return nil, fmt.Errorf("recent chat messages: %w", err)
	}
	defer rows.Close()

	descMessages, err := scanChatMessages(rows, limit)
	if err != nil {
		return nil, err
	}
	for i, j := 0, len(descMessages)-1; i < j; i, j = i+1, j-1 {
		descMessages[i], descMessages[j] = descMessages[j], descMessages[i]
	}
	if descMessages == nil {
		descMessages = []model.ChatMessage{}
	}
	return descMessages, nil
}

func (r *ConversationRepo) PromoteTempSession(ctx context.Context, userID string, tempSessionID string, req model.PromoteChatTempSessionRequest) (*model.ChatTopic, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = "新的心事"
	}
	topicKeyword := strings.TrimSpace(req.Topic)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin promote temp session: %w", err)
	}
	defer tx.Rollback(ctx)

	var lockedID string
	err = tx.QueryRow(ctx, `
		SELECT id
		FROM chat_temp_sessions
		WHERE id=$1::uuid AND user_id=$2::uuid AND status='active'
		FOR UPDATE`,
		tempSessionID, userID,
	).Scan(&lockedID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrExperienceUnavailable
		}
		return nil, fmt.Errorf("lock temp session for promotion: %w", err)
	}

	topic := &model.ChatTopic{}
	err = tx.QueryRow(ctx, `
		INSERT INTO chat_topics (
		  user_id, title, domain, sub_domain, topic, clarity_score,
		  last_opened_at, created_at, updated_at
		)
		VALUES ($1::uuid, $2, NULLIF($3, ''), NULLIF($4, ''), $5, $6, NOW(), NOW(), NOW())
		RETURNING id, status, COALESCE(title, ''), COALESCE(domain, ''), COALESCE(sub_domain, ''),
		          COALESCE(topic, ''), clarity_score, COALESCE(summary, ''), last_opened_at, created_at, updated_at`,
		userID, title, strings.TrimSpace(req.Domain), strings.TrimSpace(req.SubDomain), topicKeyword, req.ClarityScore,
	).Scan(
		&topic.ID,
		&topic.Status,
		&topic.Title,
		&topic.Domain,
		&topic.SubDomain,
		&topic.Topic,
		&topic.ClarityScore,
		&topic.Summary,
		&topic.LastOpenedAt,
		&topic.CreatedAt,
		&topic.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create promoted chat topic: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE chat_messages
		SET topic_id=$3::uuid, temp_session_id=NULL
		WHERE user_id=$1::uuid AND temp_session_id=$2::uuid`,
		userID, tempSessionID, topic.ID); err != nil {
		return nil, fmt.Errorf("move temp messages to topic: %w", err)
	}

	result, err := tx.Exec(ctx, `
		UPDATE chat_temp_sessions
		SET status='promoted', promoted_topic_id=$3::uuid, updated_at=NOW()
		WHERE id=$1::uuid AND user_id=$2::uuid AND status='active'`,
		tempSessionID, userID, topic.ID)
	if err != nil {
		return nil, fmt.Errorf("mark temp session promoted: %w", err)
	}
	if result.RowsAffected() == 0 {
		return nil, ErrExperienceUnavailable
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit promote temp session: %w", err)
	}
	return topic, nil
}

func (r *ConversationRepo) CandidateExperiencesForChat(ctx context.Context, userID string, scope model.ChatScopeContext, userMessage string, riskLevel string, limit int) ([]model.ChatCandidateExperience, error) {
	if limit < 1 || limit > 10 {
		limit = 5
	}
	domain, subDomain, topic := "", "", ""
	if scope.Topic != nil {
		domain = scope.Topic.Domain
		subDomain = scope.Topic.SubDomain
		topic = scope.Topic.Topic
	}
	allowLowReliability := riskLevel != "high_decision" && riskLevel != "professional_sensitive" && riskLevel != "safety_sensitive"

	rows, err := r.db.Query(ctx, `
		WITH collected_recent AS (
		  SELECT e.id, e.content, COALESCE(e.creator_display_name, u.display_name, u.nickname, '一个年糕用户') AS creator_name,
		         'collected'::text AS source_relation, e.visibility AS visibility,
		         COALESCE(e.quality_tier, 'public_visible') AS quality_tier,
		         COALESCE(e.source_reliability, '') AS source_reliability,
		         '' AS source_derivation_type,
		         TRUE AS is_collected,
		         c.created_at AS relation_time,
		         2 AS source_priority,
		         CASE WHEN COALESCE(e.domain::text, '')=$2 THEN 2 ELSE 0 END +
		         CASE WHEN COALESCE(e.sub_domain, '')=$3 AND $3<>'' THEN 1 ELSE 0 END +
		         CASE WHEN COALESCE(e.topic, '')=$4 AND $4<>'' THEN 1 ELSE 0 END AS relevance_score
		  FROM experience_collections c
		  JOIN experiences e ON e.id=c.experience_id
		  LEFT JOIN users u ON u.id=e.owner_user_id
		  WHERE c.user_id=$1::uuid AND c.status='active'
		    AND e.lifecycle_status='active'
		    AND e.deleted_at IS NULL
		  ORDER BY c.created_at DESC
		  LIMIT 50
		),
		own_recent AS (
		  SELECT e.id, e.content, COALESCE(e.creator_display_name, u.display_name, u.nickname, '你') AS creator_name,
		         'own'::text AS source_relation, e.visibility AS visibility,
		         COALESCE(e.quality_tier, 'private_only') AS quality_tier,
		         COALESCE(e.source_reliability, '') AS source_reliability,
		         'user_original' AS source_derivation_type,
		         EXISTS (
		           SELECT 1 FROM experience_collections c
		           WHERE c.user_id=$1::uuid AND c.experience_id=e.id AND c.status='active'
		         ) AS is_collected,
		         e.updated_at AS relation_time,
		         1 AS source_priority,
		         CASE WHEN COALESCE(e.domain::text, '')=$2 THEN 2 ELSE 0 END +
		         CASE WHEN COALESCE(e.sub_domain, '')=$3 AND $3<>'' THEN 1 ELSE 0 END +
		         CASE WHEN COALESCE(e.topic, '')=$4 AND $4<>'' THEN 1 ELSE 0 END AS relevance_score
		  FROM experiences e
		  LEFT JOIN users u ON u.id=e.owner_user_id
		  WHERE e.owner_user_id=$1::uuid
		    AND e.lifecycle_status='active'
		    AND e.deleted_at IS NULL
		  ORDER BY e.updated_at DESC
		  LIMIT 50
		),
		public_pool AS (
		  SELECT e.id, e.content, COALESCE(e.creator_display_name, '精选') AS creator_name,
		         CASE WHEN COALESCE(e.experience_type, 'platform_selected')='user_original' THEN 'public_original' ELSE 'public' END AS source_relation,
		         e.visibility AS visibility,
		         COALESCE(e.quality_tier, 'ai_citable') AS quality_tier,
		         COALESCE(e.source_reliability, '') AS source_reliability,
		         '' AS source_derivation_type,
		         FALSE AS is_collected,
		         e.updated_at AS relation_time,
		         3 AS source_priority,
		         CASE WHEN COALESCE(e.domain::text, '')=$2 THEN 2 ELSE 0 END +
		         CASE WHEN COALESCE(e.sub_domain, '')=$3 AND $3<>'' THEN 1 ELSE 0 END +
		         CASE WHEN COALESCE(e.topic, '')=$4 AND $4<>'' THEN 1 ELSE 0 END AS relevance_score
		  FROM experiences e
		  WHERE e.visibility='public'
		    AND e.lifecycle_status='active'
		    AND e.deleted_at IS NULL
		    AND e.ai_citable=TRUE
		    AND e.quality_tier IN ('ai_citable', 'high_trust')
		    AND ($5::boolean OR COALESCE(e.source_reliability, '') <> 'low')
		  ORDER BY e.updated_at DESC
		  LIMIT 80
		)
		SELECT id, content, creator_name, source_relation, visibility, quality_tier,
		       source_reliability, source_derivation_type, is_collected,
		       CASE
		         WHEN source_relation='own' THEN 'strong'
		         WHEN source_relation='collected' THEN 'card_allowed'
		         ELSE 'weak_context'
		       END AS citation_policy,
		       CASE
		         WHEN relevance_score > 0 THEN '与当前议题领域接近'
		         ELSE '近期高质量经验'
		       END AS relevance_reason
		FROM (
		  SELECT * FROM own_recent
		  UNION ALL
		  SELECT * FROM collected_recent
		  UNION ALL
		  SELECT * FROM public_pool
		) candidates
		ORDER BY relevance_score DESC, source_priority ASC, relation_time DESC
		LIMIT $6`,
		userID, domain, subDomain, topic, allowLowReliability, limit)
	if err != nil {
		return nil, fmt.Errorf("candidate experiences for chat: %w", err)
	}
	defer rows.Close()

	candidates := make([]model.ChatCandidateExperience, 0, limit)
	seen := make(map[string]struct{})
	for rows.Next() {
		var candidate model.ChatCandidateExperience
		if err := rows.Scan(
			&candidate.ExperienceID,
			&candidate.Content,
			&candidate.CreatorName,
			&candidate.SourceRelation,
			&candidate.Visibility,
			&candidate.QualityTier,
			&candidate.SourceReliability,
			&candidate.SourceDerivationType,
			&candidate.IsCollected,
			&candidate.CitationPolicy,
			&candidate.RelevanceReason,
		); err != nil {
			return nil, fmt.Errorf("scan chat candidate experience: %w", err)
		}
		if _, ok := seen[candidate.ExperienceID]; ok {
			continue
		}
		seen[candidate.ExperienceID] = struct{}{}
		candidates = append(candidates, candidate)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chat candidate experiences: %w", err)
	}
	if candidates == nil {
		candidates = []model.ChatCandidateExperience{}
	}
	return candidates, nil
}

func (r *ConversationRepo) SaveChatCitations(ctx context.Context, assistantMessageID string, cards []model.ChatReferenceCard) error {
	if len(cards) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, card := range cards {
		citationType := card.CitationType
		if citationType == "" {
			citationType = "public_featured"
		}
		batch.Queue(`
			INSERT INTO chat_citations (message_id, experience_id, citation_type, shown_at, created_at)
			VALUES ($1::uuid, $2::uuid, $3, NOW(), NOW())
			ON CONFLICT DO NOTHING`,
			assistantMessageID, card.ExperienceID, citationType,
		)
	}
	results := r.db.SendBatch(ctx, batch)
	defer results.Close()
	for range cards {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("save chat citation: %w", err)
		}
	}
	return nil
}

func scanChatMessages(rows pgx.Rows, capacity int) ([]model.ChatMessage, error) {
	messages := make([]model.ChatMessage, 0, capacity)
	for rows.Next() {
		var message model.ChatMessage
		if err := rows.Scan(
			&message.ID,
			&message.UserID,
			&message.TopicID,
			&message.TempSessionID,
			&message.Role,
			&message.Content,
			&message.Status,
			&message.RiskLevel,
			&message.ClientMessageID,
			&message.ReferencedExperienceIDs,
			&message.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan chat message: %w", err)
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chat messages: %w", err)
	}
	return messages, nil
}
