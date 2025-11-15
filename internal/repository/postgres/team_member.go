package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"pr-review/internal/models"
)

const (
	insertTeamMemberQuery = `
		INSERT INTO pr_review.team_members (team_id, user_id)
		VALUES ($1, $2)`

	selectTeamMembersQuery = `
		SELECT u.id, u.name, u.team_id, u.is_active
		FROM pr_review.team_members tm
		JOIN pr_review.users u ON tm.user_id = u.id
		WHERE tm.team_id = $1`

	deleteTeamMemberQuery = `
		DELETE FROM pr_review.team_members
		WHERE team_id = $1 AND user_id = $2`
)

type TeamMemberRepository struct {
	db *sqlx.DB
}

func NewTeamMemberRepository(db *sqlx.DB) *TeamMemberRepository {
	return &TeamMemberRepository{db: db}
}

func (r *TeamMemberRepository) AddMembers(ctx context.Context, teamID int64, userIDs []int64) error {
	if teamID == 0 {
		return fmt.Errorf("team id is required")
	}
	if len(userIDs) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer rollbackTransaction(tx)

	for _, userID := range userIDs {
		if _, err := tx.ExecContext(ctx, insertTeamMemberQuery, teamID, userID); err != nil {
			return fmt.Errorf("insert team member (team_id=%d, user_id=%d): %w", teamID, userID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (r *TeamMemberRepository) GetMembersByTeamID(ctx context.Context, teamID int64) ([]*models.User, error) {
	if teamID == 0 {
		return []*models.User{}, nil
	}

	var users []*models.User
	if err := r.db.SelectContext(ctx, &users, selectTeamMembersQuery, teamID); err != nil {
		return nil, fmt.Errorf("get team members: %w", err)
	}

	if users == nil {
		users = []*models.User{}
	}

	return users, nil
}

func (r *TeamMemberRepository) RemoveMember(ctx context.Context, teamID, userID int64) error {
	if teamID == 0 || userID == 0 {
		return fmt.Errorf("team id and user id are required")
	}

	if _, err := r.db.ExecContext(ctx, deleteTeamMemberQuery, teamID, userID); err != nil {
		return fmt.Errorf("delete team member: %w", err)
	}

	return nil
}
