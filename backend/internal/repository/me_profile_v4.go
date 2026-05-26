package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/niangao/backend/internal/model"
)

func (r *ExperienceRepo) MeProfile(ctx context.Context, userID string) (*model.MeProfile, error) {
	var displayName string
	var profileJSON string
	if err := r.db.QueryRow(ctx, `
		SELECT
			COALESCE(NULLIF(TRIM(display_name), ''), NULLIF(TRIM(nickname), ''), ''),
			COALESCE(user_settings->'profile', '{}'::jsonb)::text
		FROM users
		WHERE id=$1`,
		userID,
	).Scan(&displayName, &profileJSON); err != nil {
		return nil, fmt.Errorf("get me profile: %w", err)
	}

	profile := &model.MeProfile{CommonIssues: []string{}}
	if err := json.Unmarshal([]byte(profileJSON), profile); err != nil {
		return nil, fmt.Errorf("decode me profile: %w", err)
	}
	profile.DisplayName = displayName
	if profile.CommonIssues == nil {
		profile.CommonIssues = []string{}
	}
	return profile, nil
}

func (r *ExperienceRepo) UpdateMeProfile(ctx context.Context, userID string, patch model.MeProfilePatch) (*model.MeProfile, error) {
	profile, err := r.MeProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	if patch.DisplayName != nil {
		profile.DisplayName = strings.TrimSpace(*patch.DisplayName)
	}
	if patch.CareerStage != nil {
		profile.CareerStage = strings.TrimSpace(*patch.CareerStage)
	}
	if patch.RelationshipStatus != nil {
		profile.RelationshipStatus = strings.TrimSpace(*patch.RelationshipStatus)
	}
	if patch.IsParent != nil {
		profile.IsParent = patch.IsParent
	}
	if patch.CommonIssues != nil {
		profile.CommonIssues = *patch.CommonIssues
	}
	if patch.FreeDescription != nil {
		profile.FreeDescription = strings.TrimSpace(*patch.FreeDescription)
	}
	profile.ProfileVersion++

	profileJSON, err := json.Marshal(profile)
	if err != nil {
		return nil, fmt.Errorf("encode me profile: %w", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin update profile: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		UPDATE users
		SET display_name=$2,
		    user_settings=jsonb_set(COALESCE(user_settings, '{}'::jsonb), '{profile}', $3::jsonb, true),
		    updated_at=NOW()
		WHERE id=$1`,
		userID, profile.DisplayName, string(profileJSON)); err != nil {
		return nil, fmt.Errorf("update user profile: %w", err)
	}

	if patch.DisplayName != nil {
		if _, err := tx.Exec(ctx, `
			UPDATE experiences
			SET creator_display_name=$2, updated_at=NOW()
			WHERE COALESCE(owner_user_id, author_id)=$1::uuid
			  AND COALESCE(experience_type, 'user_original')='user_original'
			  AND deleted_at IS NULL`,
			userID, profile.DisplayName); err != nil {
			return nil, fmt.Errorf("sync experience display name: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit update profile: %w", err)
	}
	return profile, nil
}

func (r *ExperienceRepo) CreateMeFeedback(ctx context.Context, userID, feedbackType, content, appVersion, device, osVersion string) error {
	if _, err := r.db.Exec(ctx, `
		INSERT INTO feedback (user_id, feedback_type, content, app_version, device, os_version, status, created_at, updated_at)
		VALUES ($1::uuid, $2, $3, NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), 'new', NOW(), NOW())`,
		userID, feedbackType, content, appVersion, device, osVersion); err != nil {
		return fmt.Errorf("create me feedback: %w", err)
	}
	return nil
}
