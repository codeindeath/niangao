package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StatsRepo struct {
	db *pgxpool.Pool
}

func NewStatsRepo(db *pgxpool.Pool) *StatsRepo {
	return &StatsRepo{db: db}
}

type DomainCount struct {
	Domain string `json:"domain"`
	Count  int    `json:"count"`
}

type StatBlock struct {
	Count  int           `json:"count"`
	Total1 int           `json:"total1,omitempty"` // liked_by_others 或 liked
	Total2 int           `json:"total2,omitempty"` // bookmarked_by_others 或 bookmarked
	Dist   []DomainCount `json:"domain_distribution"`
}

type UserStats struct {
	Published struct {
		Count              int `json:"count"`
		LikedByOthers      int `json:"liked_by_others"`
		BookmarkedByOthers int `json:"bookmarked_by_others"`
	} `json:"published"`
	PublishedDist struct {
		Published          []DomainCount `json:"published"`
		LikedByOthers      []DomainCount `json:"liked_by_others"`
		BookmarkedByOthers []DomainCount `json:"bookmarked_by_others"`
	} `json:"published_dist"`
	Interactions struct {
		Viewed     int `json:"viewed"`
		Liked      int `json:"liked"`
		Bookmarked int `json:"bookmarked"`
	} `json:"interactions"`
	InteractionsDist struct {
		Viewed     []DomainCount `json:"viewed"`
		Liked      []DomainCount `json:"liked"`
		Bookmarked []DomainCount `json:"bookmarked"`
	} `json:"interactions_dist"`
	Chat struct {
		Conversations int `json:"conversations"`
		Messages      int `json:"messages"`
	} `json:"chat"`
}

func (r *StatsRepo) GetStats(ctx context.Context, userID string) (*UserStats, error) {
	s := &UserStats{}

	// ── Published counts ──
	r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM experiences WHERE author_id=$1 AND review_status='approved' AND deleted_at IS NULL`,
		userID,
	).Scan(&s.Published.Count)

	r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM likes WHERE experience_id IN (SELECT id FROM experiences WHERE author_id=$1)`,
		userID,
	).Scan(&s.Published.LikedByOthers)

	r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM bookmarks WHERE experience_id IN (SELECT id FROM experiences WHERE author_id=$1)`,
		userID,
	).Scan(&s.Published.BookmarkedByOthers)

	// ── Interaction counts ──
	r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM user_views WHERE user_id=$1`, userID,
	).Scan(&s.Interactions.Viewed)

	r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM likes WHERE user_id=$1`, userID,
	).Scan(&s.Interactions.Liked)

	r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM bookmarks WHERE user_id=$1`, userID,
	).Scan(&s.Interactions.Bookmarked)

	// ── Chat counts ──
	r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM conversations WHERE user_id=$1`, userID,
	).Scan(&s.Chat.Conversations)

	r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM messages WHERE conversation_id IN (SELECT id FROM conversations WHERE user_id=$1)`,
		userID,
	).Scan(&s.Chat.Messages)

	// ── Domain distributions ──
	s.PublishedDist.Published = r.domainDist(ctx,
		`SELECT COALESCE(e.sub_domain, ''), COUNT(*) FROM experiences e
		 WHERE e.author_id=$1 AND e.review_status='approved' AND e.deleted_at IS NULL
		 GROUP BY e.sub_domain ORDER BY COUNT(*) DESC`, userID)

	s.PublishedDist.LikedByOthers = r.domainDist(ctx,
		`SELECT COALESCE(e.sub_domain, ''), COUNT(*) FROM likes l
		 JOIN experiences e ON l.experience_id=e.id
		 WHERE e.author_id=$1 AND e.review_status='approved'
		 GROUP BY e.sub_domain ORDER BY COUNT(*) DESC`, userID)

	s.PublishedDist.BookmarkedByOthers = r.domainDist(ctx,
		`SELECT COALESCE(e.sub_domain, ''), COUNT(*) FROM bookmarks b
		 JOIN experiences e ON b.experience_id=e.id
		 WHERE e.author_id=$1
		 GROUP BY e.sub_domain ORDER BY COUNT(*) DESC`, userID)

	s.InteractionsDist.Viewed = r.domainDist(ctx,
		`SELECT COALESCE(e.sub_domain, ''), COUNT(*) FROM user_views v
		 JOIN experiences e ON v.experience_id=e.id
		 WHERE v.user_id=$1
		 GROUP BY e.sub_domain ORDER BY COUNT(*) DESC`, userID)

	s.InteractionsDist.Liked = r.domainDist(ctx,
		`SELECT COALESCE(e.sub_domain, ''), COUNT(*) FROM likes l
		 JOIN experiences e ON l.experience_id=e.id
		 WHERE l.user_id=$1
		 GROUP BY e.sub_domain ORDER BY COUNT(*) DESC`, userID)

	s.InteractionsDist.Bookmarked = r.domainDist(ctx,
		`SELECT COALESCE(e.sub_domain, ''), COUNT(*) FROM bookmarks b
		 JOIN experiences e ON b.experience_id=e.id
		 WHERE b.user_id=$1
		 GROUP BY e.sub_domain ORDER BY COUNT(*) DESC`, userID)

	return s, nil
}

func (r *StatsRepo) RecordView(ctx context.Context, userID, experienceID string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO user_views (user_id, experience_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, experienceID,
	)
	return err
}

func (r *StatsRepo) domainDist(ctx context.Context, query string, userID string) []DomainCount {
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var result []DomainCount
	for rows.Next() {
		var d DomainCount
		if err := rows.Scan(&d.Domain, &d.Count); err != nil {
			continue
		}
		result = append(result, d)
	}
	// Front-end will apply top-5 logic; return all sorted.
	return result
}
